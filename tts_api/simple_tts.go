package tts_api

import (
	"encoding/base64"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/CoffeeSwt/bilibili-tts-chat/config"
)

var (
	connectionPool *TTSConnectionPool
	poolMutex      sync.Mutex
	isInitialized  bool
)

// TTSConnectionPool TTS连接池
type TTSConnectionPool struct {
	clients     []*WebSocketTTSClient
	available   chan *WebSocketTTSClient
	maxSize     int
	currentSize int
	mu          sync.Mutex
	requestChan chan *TTSPoolRequest
	stopChan    chan struct{}
}

// TTSPoolRequest 连接池请求
type TTSPoolRequest struct {
	Request    TTSRequest
	ResponseCh chan *TTSPoolResponse
}

// TTSPoolResponse 连接池响应
type TTSPoolResponse struct {
	Response *TTSResponse
	Error    error
}

// NewTTSConnectionPool 创建新的TTS连接池
func NewTTSConnectionPool(maxSize int) *TTSConnectionPool {
	pool := &TTSConnectionPool{
		clients:     make([]*WebSocketTTSClient, 0, maxSize),
		available:   make(chan *WebSocketTTSClient, maxSize),
		maxSize:     maxSize,
		requestChan: make(chan *TTSPoolRequest, 100), // 缓冲100个请求
		stopChan:    make(chan struct{}),
	}

	// 启动请求处理器
	go pool.requestProcessor()

	return pool
}

// requestProcessor 处理TTS请求的工作协程
func (p *TTSConnectionPool) requestProcessor() {
	for {
		select {
		case req := <-p.requestChan:
			go p.handleRequest(req)
		case <-p.stopChan:
			return
		}
	}
}

// handleRequest 处理单个TTS请求
func (p *TTSConnectionPool) handleRequest(req *TTSPoolRequest) {
	client, err := p.getClient()
	if err != nil {
		req.ResponseCh <- &TTSPoolResponse{
			Error: fmt.Errorf("failed to get client: %v", err),
		}
		return
	}

	// 检查连接状态，如果连接断开则重连
	if !client.IsConnected() {
		if connectErr := client.Connect(); connectErr != nil {
			// 连接失败，创建新客户端
			newClient, createErr := p.createNewClient()
			if createErr != nil {
				req.ResponseCh <- &TTSPoolResponse{
					Error: fmt.Errorf("failed to create new client after connection failure: %v", createErr),
				}
				return
			}
			client = newClient
		}
	}

	// 发送TTS请求
	response, err := client.SendTTSRequest(req.Request)

	// 如果请求失败且是连接相关错误，尝试重连一次
	if err != nil && isConnectionError(err) {
		if reconnectErr := client.Connect(); reconnectErr == nil {
			response, err = client.SendTTSRequest(req.Request)
		}
	}

	// 将客户端返回到池中（只有在连接正常时才返回）
	if client.IsConnected() {
		p.returnClient(client)
	} else {
		// 连接异常，丢弃这个客户端
		client.Disconnect()
	}

	// 发送响应
	req.ResponseCh <- &TTSPoolResponse{
		Response: response,
		Error:    err,
	}
}

// isConnectionError 检查是否为连接相关错误
func isConnectionError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "connection") ||
		strings.Contains(errStr, "websocket") ||
		strings.Contains(errStr, "bad MASK") ||
		strings.Contains(errStr, "closed network")
}

// getClient 从连接池获取客户端
func (p *TTSConnectionPool) getClient() (*WebSocketTTSClient, error) {
	// 首先尝试从可用连接中获取
	select {
	case client := <-p.available:
		if client.IsConnected() {
			return client, nil
		}
		// 连接已断开，尝试重连
		if err := client.Connect(); err != nil {
			// 重连失败，创建新连接
			return p.createNewClient()
		}
		return client, nil
	default:
		// 没有可用连接，尝试创建新连接
		return p.createNewClient()
	}
}

// createNewClient 创建新的客户端连接
func (p *TTSConnectionPool) createNewClient() (*WebSocketTTSClient, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.currentSize >= p.maxSize {
		// 连接池已满，等待可用连接
		select {
		case client := <-p.available:
			if !client.IsConnected() {
				if err := client.Connect(); err != nil {
					return nil, fmt.Errorf("failed to reconnect: %v", err)
				}
			}
			return client, nil
		case <-time.After(10 * time.Second):
			return nil, fmt.Errorf("timeout waiting for available connection")
		}
	}

	client := NewWebSocketTTSClient(config.GetTTSAppID(), config.GetTTSAccessKey())
	if err := client.Connect(); err != nil {
		return nil, fmt.Errorf("failed to create new connection: %v", err)
	}

	p.clients = append(p.clients, client)
	p.currentSize++

	return client, nil
}

// returnClient 将客户端返回到连接池
func (p *TTSConnectionPool) returnClient(client *WebSocketTTSClient) {
	if client.IsConnected() {
		select {
		case p.available <- client:
			// 成功返回到池中
		default:
			// 池已满，关闭连接
			client.Disconnect()
		}
	}
}

// Close 关闭连接池
func (p *TTSConnectionPool) Close() {
	close(p.stopChan)

	p.mu.Lock()
	defer p.mu.Unlock()

	for _, client := range p.clients {
		client.Disconnect()
	}

	close(p.available)
	close(p.requestChan)
}

// InitSimpleTTS 初始化简单TTS服务（可选调用）
func InitSimpleTTS() error {
	poolMutex.Lock()
	defer poolMutex.Unlock()

	if isInitialized && connectionPool != nil {
		return nil
	}

	connectionPool = NewTTSConnectionPool(5) // 最多5个连接
	isInitialized = true
	return nil
}

// TextToAudioBase64 将文本转换为base64编码的音频数据
func TextToAudioBase64(text string) (string, error) {
	// 确保连接池已初始化
	if err := ensureConnection(); err != nil {
		return "", err
	}

	voiceId := config.GetRandomVoiceID()
	log.Printf("[TTS] 使用音色: %s", voiceId)

	// 创建TTS请求
	request := TTSRequest{
		Text:     text,
		Voice:    voiceId,
		Encoding: config.DefaultEncoding,
		Cluster:  config.DefaultCluster,
	}

	// 通过连接池发送请求
	response, err := sendTTSRequestThroughPool(request)
	if err != nil {
		return "", fmt.Errorf("TTS request failed: %v", err)
	}

	if response.Error != "" {
		return "", fmt.Errorf("TTS error: %s", response.Error)
	}

	// 将音频数据编码为base64
	base64Audio := base64.StdEncoding.EncodeToString(response.AudioData)
	return base64Audio, nil
}

// TextToAudioBase64WithRandomVoice 将文本转换为base64编码的音频数据（随机音色）
func TextToAudioBase64WithRandomVoice(text string) (string, error) {
	// 这里需要从voice_engine包获取随机音色，但为了避免循环依赖，
	// 我们在voice_engine中直接调用TextToAudioBase64WithVoice
	// 这个函数主要是为了API的完整性，实际使用中建议直接在voice_engine中处理
	return "", fmt.Errorf("请使用voice_engine.TextToVoiceWithRandomVoice()函数")
}

// TextToAudioBase64WithVoice 将文本转换为base64编码的音频数据（指定音色）
func TextToAudioBase64WithVoice(text, voice string) (string, error) {
	// 确保连接池已初始化
	if err := ensureConnection(); err != nil {
		return "", err
	}

	// 创建TTS请求（使用指定音色）
	request := TTSRequest{
		Text:     text,
		Voice:    voice, // 使用指定的音色
		Encoding: config.DefaultEncoding,
		Cluster:  config.DefaultCluster,
	}

	// 通过连接池发送请求
	response, err := sendTTSRequestThroughPool(request)
	if err != nil {
		return "", fmt.Errorf("TTS request failed: %v", err)
	}

	if response.Error != "" {
		return "", fmt.Errorf("TTS error: %s", response.Error)
	}

	// 将音频数据编码为base64
	base64Audio := base64.StdEncoding.EncodeToString(response.AudioData)
	return base64Audio, nil
}

// sendTTSRequestThroughPool 通过连接池发送TTS请求
func sendTTSRequestThroughPool(request TTSRequest) (*TTSResponse, error) {
	if connectionPool == nil {
		return nil, fmt.Errorf("connection pool not initialized")
	}

	// 创建响应通道
	responseCh := make(chan *TTSPoolResponse, 1)

	// 创建池请求
	poolRequest := &TTSPoolRequest{
		Request:    request,
		ResponseCh: responseCh,
	}

	// 发送请求到池
	select {
	case connectionPool.requestChan <- poolRequest:
		// 请求已发送
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("timeout sending request to pool")
	}

	// 等待响应
	select {
	case response := <-responseCh:
		return response.Response, response.Error
	case <-time.After(30 * time.Second):
		return nil, fmt.Errorf("timeout waiting for TTS response")
	}
}

// ensureConnection 确保连接池已初始化
func ensureConnection() error {
	if !isInitialized || connectionPool == nil {
		if err := InitSimpleTTS(); err != nil {
			return err
		}
	}
	return nil
}

// CloseSimpleTTS 关闭简单TTS服务
func CloseSimpleTTS() error {
	poolMutex.Lock()
	defer poolMutex.Unlock()

	if connectionPool != nil {
		connectionPool.Close()
		connectionPool = nil
		isInitialized = false
	}

	return nil
}
