package net

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/CoffeeSwt/bilibili-tts-chat/handler"

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
	log.Println("Initiating WebSocket client shutdown")
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
		log.Println("All goroutines finished gracefully")
	case <-time.After(10 * time.Second):
		log.Println("Timeout waiting for goroutines to finish")
	}

	// Wait for connection manager to finish
	select {
	case <-wc.doneChan:
		log.Println("Connection manager finished")
	case <-time.After(5 * time.Second):
		log.Println("Timeout waiting for connection manager to finish")
	}

	wc.setState(StateDisconnected)
	log.Println("WebSocket client shutdown complete")
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
			log.Println("WebSocket connection manager shutting down")
			return
		case <-wc.shutdownChan:
			log.Println("WebSocket connection manager received shutdown signal")
			return
		default:
			// 检查是否应该停止重连
			currentState := wc.getState()
			if currentState == StateShuttingDown {
				log.Println("Connection manager: Shutting down state detected, stopping")
				return
			}

			if err := wc.connect(); err != nil {
				log.Printf("Failed to connect: %v", err)
				if wc.reconnectCount >= wc.maxReconnects {
					log.Printf("Max reconnection attempts (%d) reached, giving up", wc.maxReconnects)
					wc.setState(StateShuttingDown)
					return
				}
				wc.waitForReconnect()
				continue
			}

			// Connection successful, reset reconnect count
			wc.reconnectCount = 0
			log.Printf("Connection established successfully after %d attempts", wc.reconnectCount)

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
				log.Println("Message processing goroutines finished")
				// 检查是否应该重连
				if wc.getState() == StateShuttingDown {
					log.Println("Shutting down, not attempting reconnection")
					return
				}
				log.Println("Preparing for reconnection...")
			case <-wc.ctx.Done():
				log.Println("Context cancelled, shutting down connection manager")
				return
			case <-wc.shutdownChan:
				log.Println("Shutdown signal received, stopping connection manager")
				return
			}
		}
	}
}

// connect establishes a WebSocket connection
func (wc *WebsocketClient) connect() error {
	wc.setState(StateConnecting)
	log.Printf("Attempting to connect to %s (attempt %d/%d)", wc.wsAddr, wc.reconnectCount+1, wc.maxReconnects)

	// Set connection timeout
	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = 10 * time.Second

	conn, _, err := dialer.Dial(wc.wsAddr, nil)
	if err != nil {
		wc.setState(StateDisconnected)
		wc.reconnectCount++
		return fmt.Errorf("failed to dial WebSocket: %w", err)
	}

	wc.conn = conn
	wc.setState(StateConnected)
	wc.lastHeartbeat = time.Now()
	wc.lastPong = time.Now()

	// 设置pong处理器
	wc.conn.SetPongHandler(func(appData string) error {
		log.Println("Received pong from server")
		wc.lastPong = time.Now()
		return nil
	})

	// 设置ping处理器（自动回复pong）
	wc.conn.SetPingHandler(func(appData string) error {
		log.Println("Received ping from server, sending pong")
		wc.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		return wc.conn.WriteMessage(websocket.PongMessage, []byte(appData))
	})

	log.Printf("WebSocket connected successfully to %s", wc.wsAddr)

	// Send authentication
	if err := wc.sendAuth(wc.authBody); err != nil {
		wc.disconnect()
		return fmt.Errorf("failed to authenticate: %w", err)
	}

	return nil
}

// disconnect closes the WebSocket connection
func (wc *WebsocketClient) disconnect() {
	wc.setState(StateDisconnected)
	if wc.conn != nil {
		wc.conn.Close()
		wc.conn = nil
	}
	wc.authed = false
}

// waitForReconnect implements exponential backoff for reconnection
func (wc *WebsocketClient) waitForReconnect() {
	wc.setState(StateReconnecting)

	// Calculate delay with exponential backoff
	delay := time.Duration(float64(wc.baseDelay) * math.Pow(2, float64(wc.reconnectCount)))
	if delay > wc.maxDelay {
		delay = wc.maxDelay
	}

	// 添加一些随机性以避免雷群效应
	jitter := time.Duration(float64(delay) * 0.1 * (2*rand.Float64() - 1))
	delay += jitter

	log.Printf("Waiting %v before reconnection attempt %d/%d (base: %v, jitter: %v)", 
		delay, wc.reconnectCount+1, wc.maxReconnects, delay-jitter, jitter)

	// 创建一个ticker来显示等待进度
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	start := time.Now()
	
	for {
		select {
		case <-time.After(delay):
			log.Printf("Reconnection wait completed after %v", time.Since(start))
			return
		case <-ticker.C:
			remaining := delay - time.Since(start)
			if remaining > 0 {
				log.Printf("Reconnection in %v...", remaining.Round(time.Second))
			}
		case <-wc.ctx.Done():
			log.Println("Context cancelled during reconnection wait")
			return
		case <-wc.shutdownChan:
			log.Println("Shutdown requested during reconnection wait")
			return
		}
	}
}

// setState safely updates the connection state
func (wc *WebsocketClient) setState(state ConnectionState) {
	wc.stateMutex.Lock()
	defer wc.stateMutex.Unlock()

	if wc.state != state {
		log.Printf("WebSocket state changed: %s -> %s", wc.state, state)
		wc.state = state
	}
}

// getState safely gets the current connection state
func (wc *WebsocketClient) getState() ConnectionState {
	wc.stateMutex.RLock()
	defer wc.stateMutex.RUnlock()
	return wc.state
}

// ReadMsg 读取长连信息
func (wc *WebsocketClient) ReadMsg() {
	defer wc.wg.Done()
	defer log.Println("ReadMsg goroutine exiting")

	// 连续超时计数器
	timeoutCount := 0
	maxTimeouts := 3 // 允许连续3次超时

	for {
		select {
		case <-wc.ctx.Done():
			log.Println("ReadMsg: Context cancelled, stopping message reading")
			return
		default:
			if wc.conn == nil {
				log.Println("ReadMsg: Connection is nil, stopping message reading")
				return
			}

			// 设置更长的读取超时时间 (120秒)
			wc.conn.SetReadDeadline(time.Now().Add(120 * time.Second))

			_, data, err := wc.conn.ReadMessage()
			if err != nil {
				// 检查是否为超时错误
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					timeoutCount++
					log.Printf("WebSocket read timeout (%d/%d), attempting to recover", timeoutCount, maxTimeouts)
					
					// 如果连续超时次数未达到上限，尝试发送ping测试连接
					if timeoutCount < maxTimeouts {
						if pingErr := wc.sendPing(); pingErr != nil {
							log.Printf("Failed to send ping during timeout recovery: %v", pingErr)
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
					log.Println("ReadMsg: Context cancelled while sending message to buffer")
					return
				default:
					log.Println("ReadMsg: Message buffer full, dropping message")
				}
			}
		}
	}
}

// unpack unpacks the raw message data into a Proto struct
func (wc *WebsocketClient) unpack(buf []byte) *Proto {
	if len(buf) < RawHeaderSize {
		log.Printf("Message too short: %d bytes", len(buf))
		return nil
	}

	retProto := &Proto{}
	retProto.PacketLength = int32(binary.BigEndian.Uint32(buf[PackOffset:HeaderOffset]))
	retProto.HeaderLength = int16(binary.BigEndian.Uint16(buf[HeaderOffset:VerOffset]))
	retProto.Version = int16(binary.BigEndian.Uint16(buf[VerOffset:OperationOffset]))
	retProto.Operation = int32(binary.BigEndian.Uint32(buf[OperationOffset:SeqIdOffset]))
	retProto.SequenceId = int32(binary.BigEndian.Uint32(buf[SeqIdOffset:]))

	if retProto.PacketLength < 0 || retProto.PacketLength > MaxPackSize {
		log.Printf("Invalid packet length: %d", retProto.PacketLength)
		return nil
	}
	if retProto.HeaderLength != RawHeaderSize {
		log.Printf("Invalid header length: %d", retProto.HeaderLength)
		return nil
	}

	if bodyLen := int(retProto.PacketLength - int32(retProto.HeaderLength)); bodyLen > 0 {
		if len(buf) < int(retProto.PacketLength) {
			log.Printf("Buffer too short for packet: got %d, need %d", len(buf), retProto.PacketLength)
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

// handleReadError processes WebSocket read errors and determines appropriate action
func (wc *WebsocketClient) handleReadError(err error) {
	log.Printf("WebSocket read error: %v", err)

	// 分析错误类型并决定处理策略
	shouldReconnect := true
	errorType := "unknown"

	// Check if it's a close error
	if closeErr, ok := err.(*websocket.CloseError); ok {
		errorType = "close_error"
		switch closeErr.Code {
		case websocket.CloseNormalClosure:
			log.Println("WebSocket closed normally")
			shouldReconnect = true
		case websocket.CloseGoingAway:
			log.Println("WebSocket server going away")
			shouldReconnect = true
		case websocket.CloseAbnormalClosure:
			log.Println("WebSocket abnormal closure detected (1006)")
			shouldReconnect = true
		case websocket.CloseNoStatusReceived:
			log.Println("WebSocket closed without status")
			shouldReconnect = true
		case websocket.ClosePolicyViolation:
			log.Println("WebSocket policy violation - may not reconnect")
			shouldReconnect = false
		case websocket.CloseUnsupportedData:
			log.Println("WebSocket unsupported data - may not reconnect")
			shouldReconnect = false
		default:
			log.Printf("WebSocket closed with code %d: %s", closeErr.Code, closeErr.Text)
			shouldReconnect = true
		}
	} else if netErr, ok := err.(net.Error); ok {
		if netErr.Timeout() {
			errorType = "timeout"
			log.Println("WebSocket read timeout - will attempt reconnection")
			shouldReconnect = true
		} else {
			errorType = "network_error"
			log.Printf("Network error: %v - will attempt reconnection", netErr)
			shouldReconnect = true
		}
	} else {
		errorType = "unknown"
		log.Printf("Unknown WebSocket error: %v - will attempt reconnection", err)
		shouldReconnect = true
	}

	// 记录错误统计信息
	log.Printf("Error analysis: type=%s, should_reconnect=%v", errorType, shouldReconnect)

	// 断开连接
	wc.disconnect()

	// 如果不应该重连，设置状态为永久断开
	if !shouldReconnect {
		wc.setState(StateShuttingDown)
		log.Println("Connection marked as permanently disconnected due to error type")
	}
}

// DoEvent 处理信息
func (wc *WebsocketClient) DoEvent() {
	defer wc.wg.Done()
	defer log.Println("DoEvent goroutine exiting")

	for {
		select {
		case <-wc.ctx.Done():
			log.Println("DoEvent: Context cancelled, stopping event processing")
			return
		case proto := <-wc.msgBuf:
			if proto == nil {
				continue
			}

			if logic, ok := wc.dispather[proto.Operation]; ok {
				func() {
					defer func() {
						if r := recover(); r != nil {
							log.Printf("Panic in event handler for operation %d: %v", proto.Operation, r)
						}
					}()
					logic(proto)
				}()
			} else {
				log.Printf("No handler for operation: %d", proto.Operation)
			}
		}
	}
}

// healthMonitor monitors connection health and sends heartbeats
func (wc *WebsocketClient) healthMonitor() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-wc.ctx.Done():
			log.Println("Health monitor shutting down")
			return
		case <-wc.shutdownChan:
			log.Println("Health monitor received shutdown signal")
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
		log.Printf("Connection appears dead (no activity for %v), disconnecting", timeSinceLastPong)
		wc.disconnect()
		return
	}

	// 发送WebSocket ping（更轻量级的连接测试）
	timeSinceLastHeartbeat := now.Sub(wc.lastHeartbeat)
	if timeSinceLastHeartbeat > 30*time.Second {
		// 先尝试发送WebSocket ping
		if err := wc.sendPing(); err != nil {
			log.Printf("Failed to send ping: %v", err)
			// 如果ping失败，尝试发送应用层心跳
			if err := wc.sendHeartBeat(); err != nil {
				log.Printf("Failed to send heartbeat: %v", err)
				wc.disconnect()
				return
			}
		}
		wc.lastHeartbeat = now
		log.Printf("Health check completed (last_pong: %v ago)", timeSinceLastPong)
	}

	// 如果太久没有收到pong，发送额外的ping
	if timeSinceLastPong > 90*time.Second {
		log.Printf("No pong received for %v, sending additional ping", timeSinceLastPong)
		if err := wc.sendPing(); err != nil {
			log.Printf("Failed to send additional ping: %v", err)
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
	log.Println("[WebsocketClient | sendHeartBeat] seq:", msg.SequenceId)
	return nil
}

// sendPing 发送WebSocket ping帧来测试连接
func (wc *WebsocketClient) sendPing() error {
	if wc.conn == nil {
		return fmt.Errorf("connection is nil")
	}

	// Set write deadline for ping
	wc.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	// Send WebSocket ping frame
	if err := wc.conn.WriteMessage(websocket.PingMessage, []byte("ping")); err != nil {
		log.Printf("Failed to send ping: %v", err)
		return err
	}

	log.Println("Ping sent to test connection")
	return nil
}

// sendMsg 发送信息
func (wc *WebsocketClient) sendMsg(msg *Proto) (err error) {
	if wc.conn == nil {
		return fmt.Errorf("connection is nil")
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
		log.Printf("Failed to send message (seq: %d, op: %d): %v", msg.SequenceId, msg.Operation, err)
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
	log.Println("[WebsocketClient | authResp] auth success")
	return
}

// heartBeatResp  心跳结果
func (wc *WebsocketClient) heartBeatResp(msg *Proto) (err error) {
	log.Println("[WebsocketClient | heartBeatResp] get HeartBeat resp", msg.Body)
	return
}

// msgResp 消息回调处理函数
func (wc *WebsocketClient) msgResp(msg *Proto) (err error) {
	for _, cmd := range msg.BodyMuti {
		// 记录原始消息（可选，用于调试）
		log.Printf("[WebsocketClient | msgResp] 收到原始消息: %s", string(cmd))

		// 使用消息处理器处理消息
		if wc.messageHandler != nil {
			if handleErr := wc.messageHandler.HandleMessage(cmd); handleErr != nil {
				log.Printf("[WebsocketClient | msgResp] 消息处理失败: %v", handleErr)
				// 这里可以选择是否继续处理其他消息，还是返回错误
				// 目前选择记录错误但继续处理其他消息
			}
		} else {
			log.Printf("[WebsocketClient | msgResp] 消息处理器未初始化")
		}
	}
	return
}
