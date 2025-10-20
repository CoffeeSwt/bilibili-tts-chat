package tts_api

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/CoffeeSwt/bilibili-tts-chat/config"

	"github.com/golang/glog"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// WebSocketTTSClient WebSocket TTS客户端
type WebSocketTTSClient struct {
	conn                 *websocket.Conn
	appID                string
	accessToken          string
	endpoint             string
	mu                   sync.RWMutex
	connected            bool
	ctx                  context.Context
	cancel               context.CancelFunc
	lastPingTime         time.Time
	reconnectAttempts    int
	maxReconnectAttempts int
	heartbeatTicker      *time.Ticker
	heartbeatDone        chan bool

	// 网络状态监控
	networkStats struct {
		totalRequests    int64
		failedRequests   int64
		timeoutRequests  int64
		lastFailureTime  time.Time
		avgResponseTime  time.Duration
		lastResponseTime time.Time
	}
}

// TTSRequest TTS请求结构
type TTSRequest struct {
	Text     string `json:"text"`
	Voice    string `json:"voice,omitempty"`
	Encoding string `json:"encoding,omitempty"`
	Cluster  string `json:"cluster,omitempty"`
}

// TTSResponse TTS响应结构
type TTSResponse struct {
	AudioData []byte `json:"audio_data"`
	Error     string `json:"error,omitempty"`
	Finished  bool   `json:"finished"`
}

// NewWebSocketTTSClient 创建新的WebSocket TTS客户端
func NewWebSocketTTSClient(appID, accessToken string) *WebSocketTTSClient {
	ctx, cancel := context.WithCancel(context.Background())
	return &WebSocketTTSClient{
		appID:                appID,
		accessToken:          accessToken,
		endpoint:             config.WSEndpoint,
		ctx:                  ctx,
		cancel:               cancel,
		maxReconnectAttempts: 3,
		heartbeatDone:        make(chan bool),
	}
}

// Connect 建立WebSocket连接
func (c *WebSocketTTSClient) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.connectUnsafe()
}

// connectUnsafe 内部连接方法（不使用锁）
func (c *WebSocketTTSClient) connectUnsafe() error {
	if c.connected && c.isConnectionHealthy() {
		return nil
	}

	// 如果连接存在但不健康，先断开
	if c.conn != nil {
		c.conn.Close()
		c.connected = false
	}

	header := http.Header{}
	header.Set("Authorization", fmt.Sprintf("Bearer;%s", c.accessToken))

	// 创建自定义Dialer，设置更大的读写缓冲区以处理大型TTS响应
	dialer := &websocket.Dialer{
		ReadBufferSize:   64 * 1024,        // 64KB 读取缓冲区
		WriteBufferSize:  64 * 1024,        // 64KB 写入缓冲区
		HandshakeTimeout: 30 * time.Second, // 握手超时
		NetDial: func(network, addr string) (net.Conn, error) {
			// 创建带超时的网络连接
			conn, err := net.DialTimeout(network, addr, 15*time.Second)
			if err != nil {
				return nil, err
			}

			// 设置TCP连接的KeepAlive和超时参数
			if tcpConn, ok := conn.(*net.TCPConn); ok {
				tcpConn.SetKeepAlive(true)
				tcpConn.SetKeepAlivePeriod(30 * time.Second)
				tcpConn.SetNoDelay(true) // 禁用Nagle算法，减少延迟
			}

			return conn, nil
		},
	}

	glog.Infof("尝试连接到TTS WebSocket服务: %s", c.endpoint)
	glog.Infof("连接参数: AppID=%s, Token=%s...", c.appID, func() string {
		if len(c.accessToken) > 10 {
			return c.accessToken[:10]
		}
		return c.accessToken
	}())

	conn, resp, err := dialer.DialContext(c.ctx, c.endpoint, header)
	if err != nil {
		c.reconnectAttempts++

		// 提供更详细的错误信息
		var errorDetail string
		if resp != nil {
			errorDetail = fmt.Sprintf("HTTP状态: %d, 响应头: %v", resp.StatusCode, resp.Header)
			if resp.StatusCode == 401 {
				errorDetail += " (可能是认证失败，请检查AppID和AccessToken)"
			} else if resp.StatusCode == 403 {
				errorDetail += " (可能是权限不足或配额用尽)"
			} else if resp.StatusCode >= 500 {
				errorDetail += " (服务器内部错误)"
			}
		} else {
			errorDetail = "网络连接失败，请检查网络连接和服务地址"
		}

		glog.Errorf("WebSocket连接失败 (第%d次尝试): %v, %s", c.reconnectAttempts, err, errorDetail)
		return fmt.Errorf("failed to connect to WebSocket (attempt %d/%d): %v, %s",
			c.reconnectAttempts, c.maxReconnectAttempts, err, errorDetail)
	}

	c.conn = conn
	c.connected = true
	c.lastPingTime = time.Now()
	c.reconnectAttempts = 0 // 重置重连计数

	// 设置连接参数
	c.conn.SetPongHandler(func(string) error {
		c.lastPingTime = time.Now()
		return nil
	})

	// 启动心跳机制
	c.startHeartbeat()

	glog.Infof("WebSocket connection established, Logid: %s", resp.Header.Get("x-tt-logid"))
	return nil
}

// Disconnect 断开WebSocket连接
func (c *WebSocketTTSClient) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected || c.conn == nil {
		return nil
	}

	c.cancel()

	// 停止心跳
	c.stopHeartbeat()

	err := c.conn.Close()
	c.connected = false
	c.conn = nil
	return err
}

// IsConnected 检查连接状态
func (c *WebSocketTTSClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected && c.isConnectionHealthy()
}

// isConnectionHealthy 检查连接是否健康（内部方法，不加锁）
func (c *WebSocketTTSClient) isConnectionHealthy() bool {
	if c.conn == nil {
		return false
	}

	// 检查连接是否超时（超过60秒没有活动）
	if time.Since(c.lastPingTime) > 60*time.Second {
		glog.Warningf("连接超时，上次活动时间: %v", c.lastPingTime)
		return false
	}

	// 尝试发送ping来检查连接
	if err := c.conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(5*time.Second)); err != nil {
		glog.Warningf("连接健康检查失败: %v", err)
		return false
	}

	return true
}

// SendTTSRequest 发送TTS请求并接收音频数据
func (c *WebSocketTTSClient) SendTTSRequest(req TTSRequest) (*TTSResponse, error) {
	// 使用互斥锁保护整个请求-响应过程，防止并发访问
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.sendTTSRequestWithRetry(req, 0)
}

// sendTTSRequestWithRetry 带重试的TTS请求发送
func (c *WebSocketTTSClient) sendTTSRequestWithRetry(req TTSRequest, retryCount int) (*TTSResponse, error) {
	maxRetries := 3
	startTime := time.Now()

	// 检查网络健康状态
	healthy, reason := c.getNetworkHealth()
	if !healthy && retryCount == 0 {
		glog.Warningf("网络状态不健康: %s，将使用更保守的重试策略", reason)
	}

	// 检查并建立连接
	if !c.connected || !c.isConnectionHealthy() {
		glog.Infof("连接不健康，尝试重新连接...")
		if err := c.connectUnsafe(); err != nil {
			if retryCount < maxRetries && c.reconnectAttempts <= c.maxReconnectAttempts {
				// 智能指数退避策略：基础延迟 + 随机抖动
				baseDelay := time.Duration(1<<uint(retryCount)) * time.Second // 1s, 2s, 4s, 8s...
				jitter := time.Duration(retryCount*500) * time.Millisecond    // 添加抖动避免雷群效应
				totalDelay := baseDelay + jitter

				glog.Warningf("连接失败，等待%v后重试 (第%d次重试): %v", totalDelay, retryCount+1, err)
				time.Sleep(totalDelay)
				return c.sendTTSRequestWithRetry(req, retryCount+1)
			}
			return nil, fmt.Errorf("连接失败，已达到最大重试次数: %v", err)
		}
	}

	// 设置默认值
	if req.Voice == "" {
		req.Voice = config.DefaultVoice
	}
	if req.Encoding == "" {
		req.Encoding = config.DefaultEncoding
	}
	if req.Cluster == "" {
		req.Cluster = VoiceToCluster(req.Voice)
	}

	// 记录音色和集群映射关系用于调试
	glog.Infof("音色和集群映射详情:")
	glog.Infof("  - 音色ID: %s", req.Voice)
	glog.Infof("  - 映射集群: %s", req.Cluster)

	// 分析音色类型
	var voiceType string
	if strings.HasPrefix(req.Voice, "ICL_") {
		voiceType = "ICL系列音色"
	} else if strings.HasPrefix(req.Voice, "saturn_") {
		voiceType = "豆包语音合成模型2.0音色（saturn前缀）"
	} else if strings.Contains(req.Voice, "_saturn_") {
		voiceType = "豆包语音合成模型2.0音色（saturn后缀）"
	} else if strings.Contains(req.Voice, "_jupiter_") {
		voiceType = "端到端实时语音大模型-O版本服务端音色"
	} else if strings.Contains(req.Voice, "_mars_") {
		voiceType = "豆包语音合成模型1.0音色"
	} else if strings.Contains(req.Voice, "_uranus_") {
		voiceType = "uranus系列音色"
	} else if strings.Contains(req.Voice, "_moon_") {
		voiceType = "moon系列音色"
	} else {
		voiceType = "其他音色"
	}
	glog.Infof("  - 音色类型: %s", voiceType)

	// 验证请求参数
	if err := c.validateTTSRequest(req); err != nil {
		glog.Errorf("请求参数验证失败: %v", err)
		return nil, fmt.Errorf("invalid request parameters: %v", err)
	}

	// 构建请求
	reqID := uuid.New().String()
	userID := uuid.New().String()

	request := map[string]interface{}{
		"app": map[string]interface{}{
			"appid":   c.appID,
			"token":   c.accessToken,
			"cluster": req.Cluster,
		},
		"user": map[string]interface{}{
			"uid": userID,
		},
		"audio": map[string]interface{}{
			"voice_type": req.Voice,
			"encoding":   req.Encoding,
		},
		"request": map[string]interface{}{
			"reqid":          reqID,
			"text":           req.Text,
			"operation":      "submit",
			"with_timestamp": "1",
			"extra_param": func() string {
				str, _ := json.Marshal(map[string]interface{}{
					"disable_markdown_filter": false,
				})
				return string(str)
			}(),
		},
	}

	// 记录详细的请求参数用于调试
	glog.Infof("TTS请求参数详情:")
	glog.Infof("  - AppID: %s", c.appID)
	glog.Infof("  - Token: %s...", func() string {
		if len(c.accessToken) > 10 {
			return c.accessToken[:10]
		}
		return c.accessToken
	}())
	glog.Infof("  - Cluster: %s", req.Cluster)
	glog.Infof("  - Voice: %s", req.Voice)
	glog.Infof("  - Encoding: %s", req.Encoding)
	glog.Infof("  - RequestID: %s", reqID)
	glog.Infof("  - UserID: %s", userID)
	glog.Infof("  - Text: %s", func() string {
		if len(req.Text) > 50 {
			return req.Text[:50] + "..."
		}
		return req.Text
	}())

	payload, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	// 发送请求
	glog.Infof("发送TTS请求到服务器...")
	if err = FullClientRequest(c.conn, payload); err != nil {
		glog.Errorf("发送TTS请求失败: %v", err)

		if isConnectionError(err) && retryCount < maxRetries {
			// 智能指数退避策略：基础延迟 + 随机抖动
			baseDelay := time.Duration(1<<uint(retryCount)) * time.Second // 1s, 2s, 4s, 8s...
			jitter := time.Duration(retryCount*500) * time.Millisecond    // 添加抖动避免雷群效应
			totalDelay := baseDelay + jitter

			glog.Warningf("发送请求失败，标记连接为不健康并等待%v后重试 (第%d次重试): %v", totalDelay, retryCount+1, err)
			c.connected = false // 标记连接为不健康
			time.Sleep(totalDelay)
			return c.sendTTSRequestWithRetry(req, retryCount+1)
		}

		// 提供更详细的错误信息
		var errorDetail string
		if strings.Contains(err.Error(), "connection") {
			errorDetail = "连接已断开，请检查网络连接"
		} else if strings.Contains(err.Error(), "timeout") {
			errorDetail = "请求超时，可能是网络延迟过高"
		} else if strings.Contains(err.Error(), "closed") {
			errorDetail = "连接已关闭，可能是服务器主动断开连接"
		} else {
			errorDetail = "未知的发送错误"
		}

		return nil, fmt.Errorf("failed to send request: %v (%s)", err, errorDetail)
	}
	glog.Infof("TTS请求发送成功，等待服务器响应...")

	// 接收响应
	response, err := c.receiveResponseUnsafe()
	responseTime := time.Since(startTime)

	if err != nil {
		// 检查是否为超时错误
		isTimeout := isTimeoutError(err)
		c.updateNetworkStats(false, isTimeout, responseTime)

		glog.Errorf("接收TTS响应失败: %v", err)

		if isConnectionError(err) && retryCount < maxRetries {
			// 标记连接为不健康
			c.connected = false
			
			var totalDelay time.Duration
			
			// 针对1006异常关闭错误和其他严重连接错误，使用快速重试策略
			if shouldFastRetry(err) {
				// 快速重试：短延迟，适用于异常关闭等可能快速恢复的错误
				if retryCount == 0 {
					totalDelay = 100 * time.Millisecond // 第一次重试几乎立即
				} else if retryCount == 1 {
					totalDelay = 500 * time.Millisecond // 第二次重试稍微等待
				} else {
					// 后续重试使用指数退避，但起始延迟较小
					baseDelay := time.Duration(1<<uint(retryCount-2)) * time.Second // 1s, 2s, 4s...
					jitter := time.Duration(retryCount*200) * time.Millisecond     // 减少抖动
					totalDelay = baseDelay + jitter
				}
				
				glog.Warningf("检测到异常关闭错误(1006)，快速重试 - 等待%v后重试 (第%d次重试): %v", totalDelay, retryCount+1, err)
			} else {
				// 普通连接错误使用标准指数退避策略
				baseDelay := time.Duration(1<<uint(retryCount)) * time.Second // 1s, 2s, 4s, 8s...
				jitter := time.Duration(retryCount*500) * time.Millisecond    // 添加抖动避免雷群效应
				totalDelay = baseDelay + jitter
				
				glog.Warningf("接收响应失败，标记连接为不健康并等待%v后重试 (第%d次重试): %v", totalDelay, retryCount+1, err)
			}
			
			time.Sleep(totalDelay)
			return c.sendTTSRequestWithRetry(req, retryCount+1)
		}

		// 提供更详细的错误信息
		var errorDetail string
		if isTimeout {
			errorDetail = "响应超时，可能是服务器处理时间过长或网络延迟"
		} else if isAbnormalCloseError(err) {
			errorDetail = "WebSocket异常关闭(1006)，通常由网络中断或服务器异常断开导致，已尝试快速重连"
		} else if strings.Contains(err.Error(), "Init Engine Instance failed") {
			errorDetail = "TTS引擎初始化失败，可能是认证问题或服务配置错误"
		} else if strings.Contains(err.Error(), "unexpected EOF") {
			errorDetail = "连接意外中断，可能是网络不稳定或服务器重启"
		} else if strings.Contains(err.Error(), "connection") {
			errorDetail = "连接中断，可能是网络不稳定"
		} else if strings.Contains(err.Error(), "closed") {
			errorDetail = "连接被关闭，可能是服务器主动断开"
		} else {
			errorDetail = "未知的响应错误"
		}

		return nil, fmt.Errorf("failed to receive response: %v (%s)", err, errorDetail)
	}

	// 成功接收响应，更新统计信息
	c.updateNetworkStats(true, false, responseTime)
	glog.V(2).Infof("TTS请求成功完成，响应时间: %v", responseTime)

	return response, nil
}

// updateNetworkStats 更新网络统计信息
func (c *WebSocketTTSClient) updateNetworkStats(success bool, isTimeout bool, responseTime time.Duration) {
	c.networkStats.totalRequests++

	if !success {
		c.networkStats.failedRequests++
		c.networkStats.lastFailureTime = time.Now()

		if isTimeout {
			c.networkStats.timeoutRequests++
		}
	} else {
		// 更新平均响应时间（简单移动平均）
		if c.networkStats.avgResponseTime == 0 {
			c.networkStats.avgResponseTime = responseTime
		} else {
			c.networkStats.avgResponseTime = (c.networkStats.avgResponseTime + responseTime) / 2
		}
		c.networkStats.lastResponseTime = time.Now()
	}
}

// getNetworkHealth 获取网络健康状态
func (c *WebSocketTTSClient) getNetworkHealth() (healthy bool, reason string) {
	if c.networkStats.totalRequests == 0 {
		return true, "no requests yet"
	}

	// 计算失败率
	failureRate := float64(c.networkStats.failedRequests) / float64(c.networkStats.totalRequests)
	timeoutRate := float64(c.networkStats.timeoutRequests) / float64(c.networkStats.totalRequests)

	// 检查最近是否有失败
	timeSinceLastFailure := time.Since(c.networkStats.lastFailureTime)

	// 网络健康判断标准
	if failureRate > 0.5 { // 失败率超过50%
		return false, fmt.Sprintf("high failure rate: %.2f%%", failureRate*100)
	}

	if timeoutRate > 0.3 { // 超时率超过30%
		return false, fmt.Sprintf("high timeout rate: %.2f%%", timeoutRate*100)
	}

	if timeSinceLastFailure < 30*time.Second && c.networkStats.failedRequests > 3 {
		return false, "recent failures detected"
	}

	return true, "network healthy"
}

// isTimeoutError 检查错误是否为超时错误
func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "deadline exceeded") ||
		strings.Contains(errStr, "i/o timeout")
}

// isAbnormalCloseError 检查是否为WebSocket异常关闭错误（1006）
func isAbnormalCloseError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "close 1006") || strings.Contains(errStr, "abnormal closure")
}

// shouldFastRetry 判断是否应该快速重试（针对1006等异常关闭错误）
func shouldFastRetry(err error) bool {
	return isAbnormalCloseError(err) || 
		   strings.Contains(strings.ToLower(err.Error()), "unexpected eof") ||
		   strings.Contains(strings.ToLower(err.Error()), "broken pipe")
}

// startHeartbeat 启动心跳机制
func (c *WebSocketTTSClient) startHeartbeat() {
	// 停止之前的心跳（如果存在）
	c.stopHeartbeat()

	// 创建新的心跳定时器，每20秒发送一次ping（更频繁的心跳）
	c.heartbeatTicker = time.NewTicker(20 * time.Second)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				glog.Errorf("心跳goroutine发生panic: %v", r)
			}
		}()

		consecutiveFailures := 0
		maxConsecutiveFailures := 3 // 连续失败3次后触发重连

		for {
			select {
			case <-c.heartbeatTicker.C:
				// 发送ping消息
				if err := c.sendPing(); err != nil {
					consecutiveFailures++
					glog.Warningf("发送心跳ping失败 (连续失败%d次): %v", consecutiveFailures, err)
					
					// 检查是否为连接错误
					if isConnectionError(err) || consecutiveFailures >= maxConsecutiveFailures {
						glog.Errorf("心跳检测到连接异常，标记连接为不健康")
						c.mu.Lock()
						c.connected = false
						c.mu.Unlock()
						
						// 如果是严重的连接错误，尝试主动重连
						if isAbnormalCloseError(err) || consecutiveFailures >= maxConsecutiveFailures {
							glog.Infof("尝试主动重连...")
							go c.attemptReconnect()
						}
					}
				} else {
					// 心跳成功，重置失败计数
					if consecutiveFailures > 0 {
						glog.V(2).Infof("心跳恢复正常，重置失败计数")
						consecutiveFailures = 0
					}
					c.lastPingTime = time.Now()
				}
			case <-c.heartbeatDone:
				glog.V(3).Infof("心跳机制已停止")
				return
			case <-c.ctx.Done():
				glog.V(3).Infof("上下文取消，停止心跳")
				return
			}
		}
	}()

	glog.V(2).Infof("心跳机制已启动（间隔20秒）")
}

// stopHeartbeat 停止心跳机制
func (c *WebSocketTTSClient) stopHeartbeat() {
	if c.heartbeatTicker != nil {
		c.heartbeatTicker.Stop()
		c.heartbeatTicker = nil
	}

	// 非阻塞发送停止信号
	select {
	case c.heartbeatDone <- true:
	default:
	}
}

// sendPing 发送ping消息
func (c *WebSocketTTSClient) sendPing() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected || c.conn == nil {
		return fmt.Errorf("连接未建立")
	}

	// 设置写入超时
	c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	if err := c.conn.WriteMessage(websocket.PingMessage, []byte("heartbeat")); err != nil {
		return fmt.Errorf("发送ping消息失败: %v", err)
	}

	glog.V(3).Infof("发送心跳ping消息")
	return nil
}

// attemptReconnect 尝试重新连接
func (c *WebSocketTTSClient) attemptReconnect() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// 避免重复重连
	if c.connected {
		return
	}
	
	glog.Infof("开始尝试重新连接...")
	
	// 先断开现有连接
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	
	// 尝试重连，最多重试3次
	maxReconnectAttempts := 3
	for attempt := 1; attempt <= maxReconnectAttempts; attempt++ {
		glog.Infof("重连尝试 %d/%d", attempt, maxReconnectAttempts)
		
		if err := c.connectUnsafe(); err != nil {
			glog.Errorf("重连失败 (尝试%d/%d): %v", attempt, maxReconnectAttempts, err)
			
			if attempt < maxReconnectAttempts {
				// 等待一段时间后重试
				delay := time.Duration(attempt) * time.Second
				glog.Infof("等待%v后重试...", delay)
				time.Sleep(delay)
			}
		} else {
			glog.Infof("重连成功！")
			c.connected = true
			c.reconnectAttempts = 0
			return
		}
	}
	
	glog.Errorf("重连失败，已达到最大重试次数")
}

// receiveResponseUnsafe 接收WebSocket响应（不加锁版本）
func (c *WebSocketTTSClient) receiveResponseUnsafe() (*TTSResponse, error) {
	var audioData []byte
	var response TTSResponse
	maxAudioSize := 50 * 1024 * 1024 // 50MB
	messageCount := 0
	startTime := time.Now()

	glog.V(2).Infof("开始接收TTS响应...")

	for {
		select {
		case <-c.ctx.Done():
			return nil, fmt.Errorf("context cancelled")
		default:
			// 设置读取超时
			c.conn.SetReadDeadline(time.Now().Add(30 * time.Second))

			msg, err := ReceiveMessage(c.conn)
			if err != nil {
				// 详细的错误日志记录
				// elapsed := time.Since(startTime)
				// glog.Errorf("接收WebSocket消息失败 (已接收%d条消息，耗时%v): %v", messageCount, elapsed, err)

				// 检查是否为连接错误
				if isConnectionError(err) {
					glog.Warningf("检测到连接错误，标记连接为不健康: %v", err)
					c.connected = false
				}

				return nil, fmt.Errorf("failed to receive message: %v", err)
			}

			messageCount++
			glog.V(3).Infof("接收到第%d条消息，类型: %s", messageCount, msg.MsgType)

			switch msg.MsgType {
			case MsgTypeFrontEndResultServer:
				// 处理前端结果
				glog.V(2).Infof("收到前端结果 (第%d条消息): %s", messageCount, string(msg.Payload))

			case MsgTypeAudioOnlyServer:
				// 处理音频数据
				audioChunkSize := len(msg.Payload)
				audioData = append(audioData, msg.Payload...)
				totalAudioSize := len(audioData)

				glog.V(3).Infof("收到音频数据块 (第%d条消息): %d字节，总计: %d字节", messageCount, audioChunkSize, totalAudioSize)

				if totalAudioSize > maxAudioSize {
					glog.Errorf("音频数据过大: %d字节 (最大允许: %d字节)", totalAudioSize, maxAudioSize)
					return nil, fmt.Errorf("audio data too large: %d bytes", totalAudioSize)
				}

				// 检查是否为最后一个包
				if msg.Sequence < 0 {
					elapsed := time.Since(startTime)
					glog.Infof("TTS响应接收完成: 共%d条消息，音频数据%d字节，耗时%v", messageCount, totalAudioSize, elapsed)

					response.AudioData = audioData
					response.Finished = true
					return &response, nil
				}

			case MsgTypeError:
				// 处理错误
				errorMsg := string(msg.Payload)
				glog.Errorf("服务器返回错误 (第%d条消息): %s", messageCount, errorMsg)
				return nil, fmt.Errorf("server error: %s", errorMsg)

			default:
				glog.Warningf("收到未知消息类型 (第%d条消息): %s，数据长度: %d字节", messageCount, msg.MsgType, len(msg.Payload))
			}
		}
	}
}

// VoiceToCluster 根据音色确定集群
// validateTTSRequest 验证TTS请求参数
func (c *WebSocketTTSClient) validateTTSRequest(req TTSRequest) error {
	// 验证文本内容
	if strings.TrimSpace(req.Text) == "" {
		return fmt.Errorf("text cannot be empty")
	}

	// 验证文本长度（假设最大长度为5000字符）
	if len(req.Text) > 5000 {
		return fmt.Errorf("text too long: %d characters (max: 5000)", len(req.Text))
	}

	// 验证音色参数
	if req.Voice == "" {
		return fmt.Errorf("voice cannot be empty")
	}

	// 验证编码格式
	validEncodings := []string{"mp3", "wav", "pcm"}
	isValidEncoding := false
	for _, encoding := range validEncodings {
		if req.Encoding == encoding {
			isValidEncoding = true
			break
		}
	}
	if !isValidEncoding {
		return fmt.Errorf("invalid encoding: %s (supported: %v)", req.Encoding, validEncodings)
	}

	// 验证集群参数
	validClusters := []string{"volcano_tts", "volcano_icl"}
	isValidCluster := false
	for _, cluster := range validClusters {
		if req.Cluster == cluster {
			isValidCluster = true
			break
		}
	}
	if !isValidCluster {
		return fmt.Errorf("invalid cluster: %s (supported: %v)", req.Cluster, validClusters)
	}

	// 验证客户端配置
	if c.appID == "" {
		return fmt.Errorf("appID is not configured")
	}
	if c.accessToken == "" {
		return fmt.Errorf("access token is not configured")
	}

	return nil
}

func VoiceToCluster(voice string) string {
	// 根据火山引擎文档的音色分类来确定集群

	// ICL系列音色使用 volcano_icl 集群
	if strings.HasPrefix(voice, "ICL_") {
		return "volcano_icl"
	}

	// 豆包语音合成模型2.0音色（saturn系列）使用 volcano_tts 集群
	if strings.HasPrefix(voice, "saturn_") {
		return "volcano_tts"
	}

	// 端到端实时语音大模型-O版本服务端音色（jupiter系列）使用 volcano_tts 集群
	if strings.Contains(voice, "_jupiter_") {
		return "volcano_tts"
	}

	// 豆包语音合成模型1.0音色（mars系列）使用 volcano_tts 集群
	if strings.Contains(voice, "_mars_") {
		return "volcano_tts"
	}

	// 其他saturn系列音色（如saturn_bigtts后缀）使用 volcano_tts 集群
	if strings.Contains(voice, "_saturn_") {
		return "volcano_tts"
	}

	// uranus系列音色使用 volcano_tts 集群
	if strings.Contains(voice, "_uranus_") {
		return "volcano_tts"
	}

	// moon系列音色使用 volcano_tts 集群
	if strings.Contains(voice, "_moon_") {
		return "volcano_tts"
	}

	// 兼容旧的S_前缀音色（映射到ICL）
	if len(voice) > 2 && voice[:2] == "S_" {
		return "volcano_icl"
	}

	// 默认使用 volcano_tts 集群
	return "volcano_tts"
}

// TTSWebSocket WebSocket版本的TTS函数
func TTSWebSocket(text string) ([]byte, error) {
	client := NewWebSocketTTSClient(config.GetTTSAppID(), config.GetTTSAccessKey())
	defer client.Disconnect()

	req := TTSRequest{
		Text:     text,
		Voice:    config.DefaultVoice,
		Encoding: config.DefaultEncoding,
	}

	resp, err := client.SendTTSRequest(req)
	if err != nil {
		return nil, err
	}

	if resp.Error != "" {
		return nil, fmt.Errorf("TTS error: %s", resp.Error)
	}

	return resp.AudioData, nil
}

// TTSWebSocketWithOptions WebSocket版本的TTS函数（带选项）
func TTSWebSocketWithOptions(text, voice, encoding string) ([]byte, error) {
	client := NewWebSocketTTSClient(config.GetTTSAppID(), config.GetTTSAccessKey())
	defer client.Disconnect()

	req := TTSRequest{
		Text:     text,
		Voice:    voice,
		Encoding: encoding,
	}

	resp, err := client.SendTTSRequest(req)
	if err != nil {
		return nil, err
	}

	if resp.Error != "" {
		return nil, fmt.Errorf("TTS error: %s", resp.Error)
	}

	return resp.AudioData, nil
}

// StreamingTTSClient 流式TTS客户端，支持持续连接
type StreamingTTSClient struct {
	*WebSocketTTSClient
	requestChan  chan TTSRequest
	responseChan chan *TTSResponse
	errorChan    chan error
	stopChan     chan struct{}
	wg           sync.WaitGroup
}

// NewStreamingTTSClient 创建流式TTS客户端
func NewStreamingTTSClient(appID, accessToken string) *StreamingTTSClient {
	return &StreamingTTSClient{
		WebSocketTTSClient: NewWebSocketTTSClient(appID, accessToken),
		requestChan:        make(chan TTSRequest, 10),
		responseChan:       make(chan *TTSResponse, 10),
		errorChan:          make(chan error, 10),
		stopChan:           make(chan struct{}),
	}
}

// Start 启动流式客户端
func (s *StreamingTTSClient) Start() error {
	if err := s.Connect(); err != nil {
		return err
	}

	s.wg.Add(1)
	go s.processRequests()

	return nil
}

// Stop 停止流式客户端
func (s *StreamingTTSClient) Stop() {
	close(s.stopChan)
	s.wg.Wait()
	s.Disconnect()
}

// SendRequest 发送TTS请求
func (s *StreamingTTSClient) SendRequest(req TTSRequest) {
	select {
	case s.requestChan <- req:
	case <-s.stopChan:
	}
}

// GetResponse 获取TTS响应
func (s *StreamingTTSClient) GetResponse() *TTSResponse {
	select {
	case resp := <-s.responseChan:
		return resp
	case <-s.stopChan:
		return nil
	}
}

// GetError 获取错误
func (s *StreamingTTSClient) GetError() error {
	select {
	case err := <-s.errorChan:
		return err
	case <-s.stopChan:
		return nil
	}
}

// processRequests 处理请求
func (s *StreamingTTSClient) processRequests() {
	defer s.wg.Done()

	for {
		select {
		case req := <-s.requestChan:
			resp, err := s.SendTTSRequest(req)
			if err != nil {
				select {
				case s.errorChan <- err:
				case <-s.stopChan:
					return
				}
			} else {
				select {
				case s.responseChan <- resp:
				case <-s.stopChan:
					return
				}
			}
		case <-s.stopChan:
			return
		}
	}
}
