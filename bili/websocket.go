package bili

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/CoffeeSwt/bilibili-tts-chat/handler"
	"github.com/CoffeeSwt/bilibili-tts-chat/logger"

	"github.com/gorilla/websocket"
)

const (
	MaxBodySize     = int32(1 << 11)
	CmdSize         = 4
	PackSize        = 4
	HeaderSize      = 2
	VerSize         = 2
	OperationSize   = 4
	SeqIdSize       = 4
	HeartbeatSize   = 4
	RawHeaderSize   = PackSize + HeaderSize + VerSize + OperationSize + SeqIdSize
	MaxPackSize     = MaxBodySize + int32(RawHeaderSize)
	PackOffset      = 0
	HeaderOffset    = PackOffset + PackSize
	VerOffset       = HeaderOffset + HeaderSize
	OperationOffset = VerOffset + VerSize
	SeqIdOffset     = OperationOffset + OperationSize
	HeartbeatOffset = SeqIdOffset + SeqIdSize
)

const (
	OP_HEARTBEAT       = int32(2)
	OP_HEARTBEAT_REPLY = int32(3)
	OP_SEND_SMS_REPLY  = int32(5)
	OP_AUTH            = int32(7)
	OP_AUTH_REPLY      = int32(8)
)

// ConnectionState represents the current state of the WebSocket connection
type ConnectionState int

const (
	StateDisconnected ConnectionState = iota
	StateConnecting
	StateConnected
	StateReconnecting
	StateShuttingDown
)

func (s ConnectionState) String() string {
	switch s {
	case StateDisconnected:
		return "Disconnected"
	case StateConnecting:
		return "Connecting"
	case StateConnected:
		return "Connected"
	case StateReconnecting:
		return "Reconnecting"
	case StateShuttingDown:
		return "ShuttingDown"
	default:
		return "Unknown"
	}
}

// ConnectionQuality represents connection quality metrics
type ConnectionQuality struct {
	TotalConnections   int64
	SuccessfulConnects int64
	FailedConnects     int64
	AbnormalClosures   int64
	TimeoutErrors      int64
	NetworkErrors      int64
	LastConnectTime    time.Time
	LastDisconnectTime time.Time
	AverageLatency     time.Duration
	ConnectionUptime   time.Duration
	ReconnectAttempts  int64
}

type WebsocketClient struct {
	// Connection management
	conn       *websocket.Conn
	wsAddr     string
	authBody   string
	state      ConnectionState
	stateMutex sync.RWMutex

	// Message handling
	msgBuf         chan *Proto
	sequenceId     int32
	dispather      map[int32]protoLogic
	authed         bool
	messageHandler *handler.MessageHandler

	// Reconnection management
	ctx            context.Context
	cancel         context.CancelFunc
	reconnectCount int
	maxReconnects  int
	baseDelay      time.Duration
	maxDelay       time.Duration

	// Health monitoring
	lastHeartbeat time.Time
	lastPong      time.Time
	healthTicker  *time.Ticker

	// Graceful shutdown
	shutdownChan chan struct{}
	doneChan     chan struct{}
	wg           sync.WaitGroup

	// Connection quality metrics
	quality              ConnectionQuality
	qualityMutex         sync.RWMutex
	connectionStart      time.Time
	pingLatency          time.Duration
	consecutiveErrors    int64
	maxConsecutiveErrors int64
}

type protoLogic func(p *Proto) (err error)

type Proto struct {
	PacketLength int32
	HeaderLength int16
	Version      int16
	Operation    int32
	SequenceId   int32
	Body         []byte
	BodyMuti     [][]byte
}

type AuthRespParam struct {
	Code int64 `json:"code,omitempty"`
}

// StartWebsocket 启动长连
func StartWebsocket(wsAddr, authBody string) (err error) {
	ctx, cancel := context.WithCancel(context.Background())

	wc := &WebsocketClient{
		wsAddr:         wsAddr,
		authBody:       authBody,
		state:          StateDisconnected,
		msgBuf:         make(chan *Proto, 1024),
		dispather:      make(map[int32]protoLogic),
		messageHandler: handler.NewMessageHandler(),
		ctx:            ctx,
		cancel:         cancel,
		maxReconnects:  10,
		baseDelay:      time.Second,
		maxDelay:       30 * time.Second,
		shutdownChan:   make(chan struct{}),
		doneChan:       make(chan struct{}),
	}

	// 注册分发处理函数
	wc.dispather[OP_AUTH_REPLY] = wc.authResp
	wc.dispather[OP_HEARTBEAT_REPLY] = wc.heartBeatResp
	wc.dispather[OP_SEND_SMS_REPLY] = wc.msgResp

	// 启动连接管理
	go wc.connectionManager()

	// 启动健康监控
	go wc.healthMonitor()

	return nil
}

// Shutdown gracefully shuts down the WebSocket client
func (wc *WebsocketClient) Shutdown() error {
	logger.Info("正在关闭WebSocket客户端...")
	wc.setState(StateShuttingDown)

	// Signal shutdown to all goroutines
	close(wc.shutdownChan)

	// Cancel context to stop all operations
	if wc.cancel != nil {
		wc.cancel()
	}

	// Stop health monitoring
	if wc.healthTicker != nil {
		wc.healthTicker.Stop()
	}

	// Close connection gracefully
	if wc.conn != nil {
		// Send close frame
		wc.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))

		// Set close deadline
		wc.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))

		// Close the connection
		wc.conn.Close()
		wc.conn = nil
	}

	// Wait for goroutines to finish with timeout
	done := make(chan struct{})
	go func() {
		wc.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.Info("所有goroutine已正常完成")
	case <-time.After(10 * time.Second):
		logger.Error("超时等待goroutine完成")
	}

	// Wait for connection manager to finish
	select {
	case <-wc.doneChan:
		logger.Info("连接管理器已正常完成")
	case <-time.After(5 * time.Second):
		logger.Error("超时等待连接管理器完成")
	}

	wc.setState(StateDisconnected)
	logger.Info("WebSocket客户端关闭完成")
	return nil
}

// IsConnected returns true if the WebSocket is currently connected
func (wc *WebsocketClient) IsConnected() bool {
	return wc.getState() == StateConnected
}

// GetConnectionState returns the current connection state
func (wc *WebsocketClient) GetConnectionState() ConnectionState {
	return wc.getState()
}

// connectionManager manages the WebSocket connection lifecycle
func (wc *WebsocketClient) connectionManager() {
	defer close(wc.doneChan)

	for {
		select {
		case <-wc.ctx.Done():
			logger.Info("WebSocket连接管理器正在关闭")
			return
		case <-wc.shutdownChan:
			logger.Info("WebSocket连接管理器收到关闭信号")
			return
		default:
			// 检查是否应该停止重连
			currentState := wc.getState()
			if currentState == StateShuttingDown {
				logger.Info("连接管理器检测到关闭状态，正在停止")
				return
			}

			if err := wc.connect(); err != nil {
				logger.Error("连接失败，正在重试...", err)
				if wc.reconnectCount >= wc.maxReconnects {
					logger.Error(fmt.Sprintf("最大重试次数(%d)已达，停止重试", wc.maxReconnects))
					wc.setState(StateShuttingDown)
					return
				}
				wc.waitForReconnect()
				continue
			}

			// Connection successful, reset reconnect count
			wc.reconnectCount = 0
			logger.Info(fmt.Sprintf("连接成功，已重试次数 %d 次", wc.reconnectCount))

			// Start message processing
			wc.wg.Add(2)
			go wc.ReadMsg()
			go wc.DoEvent()

			// Wait for connection to close or shutdown signal
			done := make(chan struct{})
			go func() {
				wc.wg.Wait()
				close(done)
			}()

			select {
			case <-done:
				logger.Info("消息处理goroutine已正常完成")
				// 检查是否应该重连
				if wc.getState() == StateShuttingDown {
					logger.Info("关闭状态检测到，不再尝试重连")
					return
				}
				logger.Info("准备进行重连...")
			case <-wc.ctx.Done():
				logger.Info("上下文已取消，正在关闭连接管理器")
				return
			case <-wc.shutdownChan:
				logger.Info("收到关闭信号，正在停止连接管理器")
				return
			}
		}
	}
}

// updateConnectionQuality updates connection quality metrics
func (wc *WebsocketClient) updateConnectionQuality(eventType string, err error) {
	wc.qualityMutex.Lock()
	defer wc.qualityMutex.Unlock()

	if err != nil {
		logger.Error("更新连接质量失败", err)
	}

	switch eventType {
	case "connect_attempt":
		atomic.AddInt64(&wc.quality.TotalConnections, 1)
	case "connect_success":
		atomic.AddInt64(&wc.quality.SuccessfulConnects, 1)
		wc.quality.LastConnectTime = time.Now()
		wc.connectionStart = time.Now()
		atomic.StoreInt64(&wc.consecutiveErrors, 0)
	case "connect_failed":
		atomic.AddInt64(&wc.quality.FailedConnects, 1)
		atomic.AddInt64(&wc.consecutiveErrors, 1)
	case "disconnect":
		wc.quality.LastDisconnectTime = time.Now()
		if !wc.connectionStart.IsZero() {
			wc.quality.ConnectionUptime = time.Since(wc.connectionStart)
		}
	case "abnormal_closure":
		atomic.AddInt64(&wc.quality.AbnormalClosures, 1)
		atomic.AddInt64(&wc.consecutiveErrors, 1)
	case "timeout_error":
		atomic.AddInt64(&wc.quality.TimeoutErrors, 1)
		atomic.AddInt64(&wc.consecutiveErrors, 1)
	case "network_error":
		atomic.AddInt64(&wc.quality.NetworkErrors, 1)
		atomic.AddInt64(&wc.consecutiveErrors, 1)
	case "reconnect_attempt":
		atomic.AddInt64(&wc.quality.ReconnectAttempts, 1)
	}

	// Update max consecutive errors
	current := atomic.LoadInt64(&wc.consecutiveErrors)
	if current > wc.maxConsecutiveErrors {
		wc.maxConsecutiveErrors = current
	}
}

// GetConnectionQuality returns current connection quality metrics
func (wc *WebsocketClient) GetConnectionQuality() ConnectionQuality {
	wc.qualityMutex.RLock()
	defer wc.qualityMutex.RUnlock()

	quality := wc.quality
	quality.AverageLatency = wc.pingLatency
	if !wc.connectionStart.IsZero() && wc.getState() == StateConnected {
		quality.ConnectionUptime = time.Since(wc.connectionStart)
	}
	return quality
}

// LogConnectionQuality logs current connection quality metrics
func (wc *WebsocketClient) LogConnectionQuality() {
	quality := wc.GetConnectionQuality()
	logger.Info("连接质量指标:")
	logger.Info(fmt.Sprintf("  总连接次数: %d", quality.TotalConnections))
	logger.Info(fmt.Sprintf("  成功连接次数: %d", quality.SuccessfulConnects))
	logger.Info(fmt.Sprintf("  异常关闭次数: %d", quality.AbnormalClosures))
	logger.Info(fmt.Sprintf("  超时错误次数: %d", quality.TimeoutErrors))
	logger.Info(fmt.Sprintf("  网络错误次数: %d", quality.NetworkErrors))
	logger.Info(fmt.Sprintf("  重连尝试次数: %d", quality.ReconnectAttempts))
	logger.Info(fmt.Sprintf("  当前连接时长: %v", quality.ConnectionUptime))
	logger.Info(fmt.Sprintf("  平均延迟: %v", quality.AverageLatency))
	logger.Info(fmt.Sprintf("  连续错误次数: %d (最大: %d)", atomic.LoadInt64(&wc.consecutiveErrors), wc.maxConsecutiveErrors))
}

// isEOFError checks if the error is an EOF or connection closed error
func (wc *WebsocketClient) isEOFError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())

	// 检查直接的EOF错误
	if err == io.EOF {
		return true
	}

	// 检查各种网络连接错误模式
	eofPatterns := []string{
		"unexpected eof",
		"connection reset by peer",
		"broken pipe",
		"use of closed network connection",
		"connection refused",
		"network is unreachable",
		"no route to host",
		"connection timed out",
		"connection aborted",
		"wsaconnreset",    // Windows specific
		"wsaconnaborted",  // Windows specific
		"wsaeconnreset",   // Windows specific
		"wsaeconnaborted", // Windows specific
	}

	for _, pattern := range eofPatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	// 检查是否为网络错误类型
	if netErr, ok := err.(net.Error); ok {
		// 临时网络错误通常表示网络问题
		return netErr.Temporary()
	}

	return false
}

// shouldAttemptReconnect determines if reconnection should be attempted based on error type and connection quality
func (wc *WebsocketClient) shouldAttemptReconnect(err error) bool {
	// Check consecutive error threshold
	consecutiveErrors := atomic.LoadInt64(&wc.consecutiveErrors)
	if consecutiveErrors > 10 {
		logger.Warn("连续错误次数超过10次，临时退避", consecutiveErrors)
		return false
	}

	// Check if it's a close error
	if closeErr, ok := err.(*websocket.CloseError); ok {
		switch closeErr.Code {
		case websocket.ClosePolicyViolation, websocket.CloseUnsupportedData:
			return false // Don't reconnect for these errors
		case websocket.CloseAbnormalClosure:
			// For 1006 errors, check if it's likely a network issue
			return wc.isEOFError(err) || strings.Contains(closeErr.Text, "unexpected EOF")
		default:
			return true
		}
	}

	// Check for network errors
	if netErr, ok := err.(net.Error); ok {
		return netErr.Temporary() || netErr.Timeout()
	}

	// Check for EOF and connection errors
	if wc.isEOFError(err) {
		return true
	}

	return true // Default to attempting reconnection
}

// handleReadError processes WebSocket read errors and determines appropriate action
func (wc *WebsocketClient) handleReadError(err error) {
	logger.Error("WebSocket 读取错误", err)

	// Analyze error type and update metrics
	shouldReconnect := wc.shouldAttemptReconnect(err)
	errorType := "unknown"

	// Check if it's a close error
	if closeErr, ok := err.(*websocket.CloseError); ok {
		errorType = "close_error"
		switch closeErr.Code {
		case websocket.CloseNormalClosure:
			logger.Info("WebSocket 正常关闭")
			wc.updateConnectionQuality("disconnect", err)
		case websocket.CloseGoingAway:
			logger.Info("WebSocket 服务器正在关闭")
			wc.updateConnectionQuality("disconnect", err)
		case websocket.CloseAbnormalClosure:
			logger.Warn(fmt.Sprintf("WebSocket 异常关闭 (1006): %s", closeErr.Text))
			wc.updateConnectionQuality("abnormal_closure", err)

			// Enhanced logging and analysis for 1006 errors
			if wc.isEOFError(err) {
				logger.Info("1006 error appears to be network-related (EOF detected)")
				// 网络相关的1006错误，使用较短的重连延迟
				wc.baseDelay = 2 * time.Second
			} else {
				logger.Info("1006 error may be server-side issue")
				// 服务器端问题，使用较长的重连延迟
				wc.baseDelay = 5 * time.Second
			}

			// 检查1006错误频率，如果过于频繁则增加延迟
			quality := wc.GetConnectionQuality()
			if quality.AbnormalClosures > 5 && quality.TotalConnections > 0 {
				abnormalRate := float64(quality.AbnormalClosures) / float64(quality.TotalConnections)
				if abnormalRate > 0.5 { // 超过50%的连接异常关闭
					logger.Warn("高异常关闭率检测到 (%.2f%%)，增加重连延迟", abnormalRate*100)
					wc.baseDelay = 10 * time.Second
				}
			}
		case websocket.CloseNoStatusReceived:
			logger.Info("WebSocket 未收到状态码，可能是服务器异常关闭")
			wc.updateConnectionQuality("abnormal_closure", err)
		case websocket.ClosePolicyViolation:
			logger.Info("WebSocket 策略违规 - 不会重新连接")
			shouldReconnect = false
			wc.updateConnectionQuality("disconnect", err)
		case websocket.CloseUnsupportedData:
			logger.Info("WebSocket 不支持的数据 - 不会重新连接")
			shouldReconnect = false
			wc.updateConnectionQuality("disconnect", err)
		default:
			logger.Info(fmt.Sprintf("WebSocket 关闭代码 %d: %s", closeErr.Code, closeErr.Text))
			wc.updateConnectionQuality("abnormal_closure", err)
		}
	} else if netErr, ok := err.(net.Error); ok {
		if netErr.Timeout() {
			errorType = "timeout"
			logger.Warn(fmt.Sprintf("WebSocket 读取超时: %v - 会尝试重新连接", netErr))
			wc.updateConnectionQuality("timeout_error", err)
		} else {
			errorType = "network_error"
			logger.Warn(fmt.Sprintf("网络错误: %v - 会尝试重新连接", netErr))
			wc.updateConnectionQuality("network_error", err)
		}
	} else if wc.isEOFError(err) {
		errorType = "eof_error"
		logger.Warn(fmt.Sprintf("EOF 或连接关闭错误: %v - 会尝试重新连接", err))
		wc.updateConnectionQuality("abnormal_closure", err)
	} else {
		errorType = "unknown"
		logger.Warn(fmt.Sprintf("未知 WebSocket 错误: %v - 会尝试重新连接", err))
		wc.updateConnectionQuality("network_error", err)
	}

	// Enhanced error analysis logging
	quality := wc.GetConnectionQuality()
	logger.Info(fmt.Sprintf("错误分析: 类型=%s, 是否重新连接=%v, 连续错误次数=%d, 总异常关闭次数=%d",
		errorType, shouldReconnect, atomic.LoadInt64(&wc.consecutiveErrors), quality.AbnormalClosures))

	// Disconnect the connection
	wc.disconnect()

	// If we shouldn't reconnect, set state to shutting down
	if !shouldReconnect {
		wc.setState(StateShuttingDown)
		logger.Info(fmt.Sprintf("连接已标记为永久断开，因为错误类型为 %s", errorType))
		wc.LogConnectionQuality() // Log final quality metrics
	}
}

// DoEvent 处理信息
func (wc *WebsocketClient) DoEvent() {
	defer wc.wg.Done()
	defer logger.Info("DoEvent  goroutine 退出")

	for {
		select {
		case <-wc.ctx.Done():
			logger.Info("DoEvent: Context cancelled, stopping event processing")
			return
		case proto := <-wc.msgBuf:
			if proto == nil {
				continue
			}

			if logic, ok := wc.dispather[proto.Operation]; ok {
				func() {
					defer func() {
						if r := recover(); r != nil {
							logger.Error(fmt.Sprintf("Panic in event handler for operation %d: %v", proto.Operation, r))
						}
					}()
					logic(proto)
				}()
			} else {
				logger.Info(fmt.Sprintf("No handler for operation: %d", proto.Operation))
			}
		}
	}
}

// healthMonitor monitors connection health and sends heartbeats
func (wc *WebsocketClient) healthMonitor() {
	// 根据官方协议要求，心跳频率为20秒
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-wc.ctx.Done():
			logger.Info("Health monitor shutting down")
			return
		case <-wc.shutdownChan:
			logger.Info("Health monitor received shutdown signal")
			return
		case <-ticker.C:
			wc.performHealthCheck()
		}
	}
}

// performHealthCheck checks connection health and sends heartbeat if needed
func (wc *WebsocketClient) performHealthCheck() {
	state := wc.getState()
	if state != StateConnected {
		return
	}

	now := time.Now()

	// 检查是否太久没有收到任何响应（包括pong和消息）
	timeSinceLastPong := now.Sub(wc.lastPong)
	if timeSinceLastPong > 150*time.Second { // 增加到150秒
		logger.Warn(fmt.Sprintf("连接已断开 (无活动 %v)，断开连接", timeSinceLastPong))
		wc.disconnect()
		return
	}

	// 发送WebSocket ping和应用层心跳（根据官方协议20秒频率）
	timeSinceLastHeartbeat := now.Sub(wc.lastHeartbeat)
	if timeSinceLastHeartbeat > 20*time.Second {
		// 先尝试发送WebSocket ping
		if err := wc.sendPing(); err != nil {
			logger.Error(fmt.Sprintf("发送 WebSocket ping 失败: %v", err))
			// 如果ping失败，尝试发送应用层心跳
			if err := wc.sendHeartBeat(); err != nil {
				logger.Error(fmt.Sprintf("发送应用层心跳失败: %v", err))
				wc.disconnect()
				return
			}
		} else {
			// ping成功后，也发送应用层心跳（官方协议要求）
			if err := wc.sendHeartBeat(); err != nil {
				logger.Error(fmt.Sprintf("发送应用层心跳失败: %v", err))
				// 应用层心跳失败不立即断开连接，因为WebSocket ping成功
			}
		}
		wc.lastHeartbeat = now
		logger.Info(fmt.Sprintf("健康检查完成 (上次pong %v 前)", timeSinceLastPong))
	}

	// 如果太久没有收到pong，发送额外的ping
	if timeSinceLastPong > 90*time.Second {
		logger.Warn(fmt.Sprintf("未收到pong %v，发送额外ping", timeSinceLastPong))
		if err := wc.sendPing(); err != nil {
			logger.Error(fmt.Sprintf("发送额外ping失败: %v", err))
			wc.disconnect()
		}
	}
}

// sendAuth 发送鉴权
func (wc *WebsocketClient) sendAuth(authBody string) (err error) {
	p := &Proto{
		Operation: OP_AUTH,
		Body:      []byte(authBody),
	}
	return wc.sendMsg(p)
}

// sendHeartBeat 发送心跳
func (wc *WebsocketClient) sendHeartBeat() error {
	if !wc.authed {
		return fmt.Errorf("not authenticated")
	}
	if wc.conn == nil {
		return fmt.Errorf("connection is nil")
	}

	msg := &Proto{}
	msg.Operation = OP_HEARTBEAT
	msg.SequenceId = wc.sequenceId
	wc.sequenceId++
	err := wc.sendMsg(msg)
	if err != nil {
		return err
	}
	logger.Info(fmt.Sprintf("发送心跳 seq: %d", msg.SequenceId))
	return nil
}

// sendPing 发送WebSocket ping帧来测试连接
func (wc *WebsocketClient) sendPing() error {
	if wc.conn == nil {
		return fmt.Errorf("connection is nil")
	}

	// Set write deadline for ping
	wc.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	// Send WebSocket ping frame with timestamp for latency measurement
	timestamp := time.Now().Format(time.RFC3339Nano)
	pingData := "ping_" + timestamp

	if err := wc.conn.WriteMessage(websocket.PingMessage, []byte(pingData)); err != nil {
		logger.Error(fmt.Sprintf("发送 ping 帧失败: %v", err))
		return err
	}

	logger.Info(fmt.Sprintf("发送 ping 帧: %s", pingData))
	return nil
}

// waitForReconnect implements enhanced exponential backoff for reconnection
func (wc *WebsocketClient) waitForReconnect() {
	wc.setState(StateReconnecting)
	wc.updateConnectionQuality("reconnect_attempt", nil)

	// Enhanced backoff calculation considering connection quality
	quality := wc.GetConnectionQuality()
	consecutiveErrors := atomic.LoadInt64(&wc.consecutiveErrors)

	// Base delay with exponential backoff
	delay := time.Duration(float64(wc.baseDelay) * math.Pow(2, float64(wc.reconnectCount)))
	if delay > wc.maxDelay {
		delay = wc.maxDelay
	}

	// Additional delay for consecutive errors (circuit breaker pattern)
	if consecutiveErrors > 5 {
		additionalDelay := time.Duration(consecutiveErrors-5) * 5 * time.Second
		delay += additionalDelay
		logger.Warn(fmt.Sprintf("添加断路器延迟 %v 因为 %d 个连续错误", additionalDelay, consecutiveErrors))
	}

	// Additional delay for high abnormal closure rate
	if quality.TotalConnections > 10 {
		abnormalRate := float64(quality.AbnormalClosures) / float64(quality.TotalConnections)
		if abnormalRate > 0.5 { // More than 50% abnormal closures
			delay = delay * 2
			logger.Warn(fmt.Sprintf("增加重连延迟 %v 因为异常关闭率 %.2f%%", delay, abnormalRate*100))
		}
	}

	// Add jitter to avoid thundering herd
	jitter := time.Duration(float64(delay) * 0.1 * (2*rand.Float64() - 1))
	delay += jitter

	logger.Info(fmt.Sprintf("等待 %v 前重连尝试 %d/%d (基础: %v, 抖动: %v, 连续错误: %d)",
		delay, wc.reconnectCount+1, wc.maxReconnects, wc.baseDelay, jitter, consecutiveErrors))

	// Log connection quality every few attempts
	if wc.reconnectCount%3 == 0 {
		wc.LogConnectionQuality()
	}

	// Create a ticker to show waiting progress
	ticker := time.NewTicker(10 * time.Second) // Increased interval
	defer ticker.Stop()

	start := time.Now()

	for {
		select {
		case <-time.After(delay):
			logger.Info(fmt.Sprintf("重连等待完成 %v", time.Since(start)))
			return
		case <-ticker.C:
			remaining := delay - time.Since(start)
			if remaining > 0 {
				logger.Info(fmt.Sprintf("重连等待中 %v... (连续错误: %d)", remaining.Round(time.Second), consecutiveErrors))
			}
		case <-wc.ctx.Done():
			logger.Error("上下文取消重连等待")
			return
		case <-wc.shutdownChan:
			logger.Info("收到关闭信号重连等待")
			return
		}
	}
}

// setState safely updates the connection state
func (wc *WebsocketClient) setState(state ConnectionState) {
	wc.stateMutex.Lock()
	defer wc.stateMutex.Unlock()

	if wc.state != state {
		logger.Info(fmt.Sprintf("WebSocket 状态变更: %s -> %s", wc.state, state))
		wc.state = state
	}
}

// getState safely gets the current connection state
func (wc *WebsocketClient) getState() ConnectionState {
	wc.stateMutex.RLock()
	defer wc.stateMutex.RUnlock()
	return wc.state
}

// connect establishes a WebSocket connection
func (wc *WebsocketClient) connect() error {
	wc.setState(StateConnecting)
	wc.updateConnectionQuality("connect_attempt", nil)

	logger.Info(fmt.Sprintf("尝试连接 %s (第 %d/%d 次)", wc.wsAddr, wc.reconnectCount+1, wc.maxReconnects))

	// Set connection timeout with progressive increase for consecutive failures
	timeoutMultiplier := 1 + (wc.reconnectCount / 3) // Increase timeout every 3 attempts
	if timeoutMultiplier > 5 {
		timeoutMultiplier = 5 // Cap at 5x
	}

	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = time.Duration(timeoutMultiplier) * 15 * time.Second

	conn, _, err := dialer.Dial(wc.wsAddr, nil)
	if err != nil {
		wc.setState(StateDisconnected)
		wc.reconnectCount++
		wc.updateConnectionQuality("connect_failed", err)

		// Enhanced error logging for connection failures
		if wc.isEOFError(err) {
			logger.Error(fmt.Sprintf("连接失败，网络问题 (EOF): %v", err))
		} else if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			logger.Error(fmt.Sprintf("连接失败，超时: %v", err))
		} else {
			logger.Error(fmt.Sprintf("连接失败: %v", err))
		}

		return fmt.Errorf("连接失败: %w", err)
	}

	wc.conn = conn
	wc.setState(StateConnected)
	wc.updateConnectionQuality("connect_success", nil)
	wc.lastHeartbeat = time.Now()
	wc.lastPong = time.Now()

	// Enhanced ping/pong handlers with latency measurement
	wc.conn.SetPongHandler(func(appData string) error {
		now := time.Now()
		wc.lastPong = now

		// Calculate ping latency if we have timing data
		if strings.HasPrefix(string(appData), "ping_") {
			if timestamp, err := time.Parse(time.RFC3339Nano, string(appData)[5:]); err == nil {
				wc.pingLatency = now.Sub(timestamp)
				logger.Info(fmt.Sprintf("Ping 延迟: %v", wc.pingLatency))
			}
		} else {
			logger.Info("收到服务器 Pong")
		}
		return nil
	})

	// Enhanced ping handler
	wc.conn.SetPingHandler(func(appData string) error {
		logger.Info("收到服务器 Ping，发送 Pong")
		wc.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		return wc.conn.WriteMessage(websocket.PongMessage, []byte(appData))
	})

	logger.Info(fmt.Sprintf("成功连接到 %s", wc.wsAddr))

	// Send authentication
	if err := wc.sendAuth(wc.authBody); err != nil {
		wc.disconnect()
		wc.updateConnectionQuality("connect_failed", err)
		return fmt.Errorf("认证失败: %w", err)
	}

	return nil
}

// disconnect closes the WebSocket connection
func (wc *WebsocketClient) disconnect() {
	wc.setState(StateDisconnected)
	wc.updateConnectionQuality("disconnect", nil)
	if wc.conn != nil {
		wc.conn.Close()
		wc.conn = nil
	}
	wc.authed = false
}

// ReadMsg 读取长连信息
func (wc *WebsocketClient) ReadMsg() {
	defer wc.wg.Done()
	defer logger.Info("ReadMsg  goroutine 退出")

	// 连续超时计数器
	timeoutCount := 0
	maxTimeouts := 3 // 允许连续3次超时

	for {
		select {
		case <-wc.ctx.Done():
			logger.Info("ReadMsg: 上下文取消，停止消息读取")
			return
		default:
			if wc.conn == nil {
				logger.Info("ReadMsg: 连接为空，停止消息读取")
				return
			}

			// 设置更长的读取超时时间 (120秒)
			// wc.conn.SetReadDeadline(time.Now().Add(10 * time.Second))

			_, data, err := wc.conn.ReadMessage()
			if err != nil {
				// 检查是否为超时错误
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					timeoutCount++
					logger.Info(fmt.Sprintf("WebSocket 读取超时 (%d/%d), 尝试恢复连接", timeoutCount, maxTimeouts))

					// 如果连续超时次数未达到上限，尝试发送ping测试连接
					if timeoutCount < maxTimeouts {
						if pingErr := wc.sendPing(); pingErr != nil {
							logger.Error(fmt.Sprintf("发送 Ping 测试连接失败: %v", pingErr))
							wc.handleReadError(err)
							return
						}
						// 继续尝试读取
						continue
					}
				}

				// 重置超时计数器（非超时错误或超时次数过多）
				timeoutCount = 0
				wc.handleReadError(err)
				return
			}

			// 成功读取消息，重置超时计数器
			timeoutCount = 0

			// Update last activity time
			wc.lastPong = time.Now()

			// Process the message
			proto := wc.unpack(data)
			if proto != nil {
				select {
				case wc.msgBuf <- proto:
					// Message sent successfully
				case <-wc.ctx.Done():
					logger.Info("ReadMsg: 上下文取消，停止发送消息到缓冲区")
					return
				default:
					logger.Info("ReadMsg: 消息缓冲区已满，丢弃消息")
				}
			}
		}
	}
}

// unpack unpacks the raw message data into a Proto struct
func (wc *WebsocketClient) unpack(buf []byte) *Proto {
	if len(buf) < RawHeaderSize {
		logger.Info(fmt.Sprintf("消息太短: %d 字节", len(buf)))
		return nil
	}

	retProto := &Proto{}
	retProto.PacketLength = int32(binary.BigEndian.Uint32(buf[PackOffset:HeaderOffset]))
	retProto.HeaderLength = int16(binary.BigEndian.Uint16(buf[HeaderOffset:VerOffset]))
	retProto.Version = int16(binary.BigEndian.Uint16(buf[VerOffset:OperationOffset]))
	retProto.Operation = int32(binary.BigEndian.Uint32(buf[OperationOffset:SeqIdOffset]))
	retProto.SequenceId = int32(binary.BigEndian.Uint32(buf[SeqIdOffset:]))

	if retProto.PacketLength < 0 || retProto.PacketLength > MaxPackSize {
		logger.Info(fmt.Sprintf("无效的包长度: %d", retProto.PacketLength))
		return nil
	}
	if retProto.HeaderLength != RawHeaderSize {
		logger.Info(fmt.Sprintf("无效的头长度: %d", retProto.HeaderLength))
		return nil
	}

	if bodyLen := int(retProto.PacketLength - int32(retProto.HeaderLength)); bodyLen > 0 {
		if len(buf) < int(retProto.PacketLength) {
			logger.Info(fmt.Sprintf("缓冲区太短，无法容纳完整包:  got %d, need %d", len(buf), retProto.PacketLength))
			return nil
		}
		retProto.Body = buf[retProto.HeaderLength:retProto.PacketLength]
	} else {
		retProto.Body = []byte{}
	}

	retProto.BodyMuti = [][]byte{retProto.Body}
	if len(retProto.BodyMuti) > 0 {
		retProto.Body = retProto.BodyMuti[0]
	}

	return retProto
}

// sendMsg 发送信息
func (wc *WebsocketClient) sendMsg(msg *Proto) (err error) {
	if wc.conn == nil {
		return fmt.Errorf("连接为空")
	}

	wc.sequenceId++
	msg.SequenceId = wc.sequenceId

	dataBuff := &bytes.Buffer{}
	packLen := int32(RawHeaderSize + len(msg.Body))
	msg.HeaderLength = RawHeaderSize
	binary.Write(dataBuff, binary.BigEndian, packLen)
	binary.Write(dataBuff, binary.BigEndian, int16(RawHeaderSize))
	binary.Write(dataBuff, binary.BigEndian, msg.Version)
	binary.Write(dataBuff, binary.BigEndian, msg.Operation)
	binary.Write(dataBuff, binary.BigEndian, msg.SequenceId)
	binary.Write(dataBuff, binary.BigEndian, msg.Body)

	// Set write deadline
	wc.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	err = wc.conn.WriteMessage(websocket.BinaryMessage, dataBuff.Bytes())
	if err != nil {
		logger.Error(fmt.Sprintf("发送消息失败 (seq: %d, op: %d): %v", msg.SequenceId, msg.Operation, err))
		return
	}
	return
}

// authResp 鉴权处理函数
func (wc *WebsocketClient) authResp(msg *Proto) (err error) {
	resp := &AuthRespParam{}
	err = json.Unmarshal(msg.Body, resp)
	if err != nil {
		return
	}
	if resp.Code != 0 {
		return
	}
	wc.authed = true
	logger.Info("鉴权成功")
	return
}

// heartBeatResp  心跳结果
func (wc *WebsocketClient) heartBeatResp(msg *Proto) (err error) {
	logger.Info(fmt.Sprintf("收到心跳响应: %s", string(msg.Body)))
	return
}

// msgResp 消息回调处理函数
func (wc *WebsocketClient) msgResp(msg *Proto) (err error) {
	for _, cmd := range msg.BodyMuti {
		// 记录原始消息（可选，用于调试）
		logger.Info("收到原始消息", "长度", len(cmd))

		// 使用消息处理器处理消息
		if wc.messageHandler != nil {
			if handleErr := wc.messageHandler.HandleMessage(cmd); handleErr != nil {
				logger.Error(fmt.Sprintf("消息处理失败: %v", handleErr))
				// 这里可以选择是否继续处理其他消息，还是返回错误
				// 目前选择记录错误但继续处理其他消息
			}
		} else {
			logger.Info("消息处理器未初始化")
		}
	}
	return
}
