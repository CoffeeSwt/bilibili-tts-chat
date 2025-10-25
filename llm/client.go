package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/CoffeeSwt/bilibili-tts-chat/logger"
)

// ProviderType AI服务提供商类型
type ProviderType string

const (
	ProviderOpenAI    ProviderType = "openai"
	ProviderClaude    ProviderType = "claude"
	ProviderGemini    ProviderType = "gemini"
	ProviderOpenRouter ProviderType = "openrouter"
)

// Message 对话消息结构
type Message struct {
	Role    string `json:"role"`    // system, user, assistant
	Content string `json:"content"` // 消息内容
}

// Config LLM客户端配置
type Config struct {
	Provider     ProviderType `json:"provider"`      // 服务提供商
	APIKey       string       `json:"api_key"`       // API密钥
	BaseURL      string       `json:"base_url"`      // 基础URL
	Model        string       `json:"model"`         // 模型名称
	Temperature  float64      `json:"temperature"`   // 温度参数 0.0-2.0
	MaxTokens    int          `json:"max_tokens"`    // 最大token数
	SystemPrompt string       `json:"system_prompt"` // 系统提示词
	Timeout      time.Duration `json:"timeout"`      // 请求超时时间
	MaxRetries   int          `json:"max_retries"`   // 最大重试次数
}

// StreamResponse 流式响应结构
type StreamResponse struct {
	Content string `json:"content"` // 响应内容
	Done    bool   `json:"done"`    // 是否完成
	Error   string `json:"error"`   // 错误信息
}

// LLMClient AI对话客户端
type LLMClient struct {
	config     *Config
	httpClient *http.Client
	mutex      sync.RWMutex
	closed     bool
	cancelFunc context.CancelFunc
}

var (
	instance *LLMClient
	once     sync.Once
)

// GetInstance 获取LLM客户端单例实例
func GetInstance() *LLMClient {
	once.Do(func() {
		instance = &LLMClient{
			config: &Config{
				Provider:     ProviderOpenAI,
				BaseURL:      "https://api.openai.com/v1",
				Model:        "gpt-3.5-turbo",
				Temperature:  0.7,
				MaxTokens:    2048,
				Timeout:      30 * time.Second,
				MaxRetries:   3,
				SystemPrompt: "你是一个智能助手，请根据用户的问题提供有帮助的回答。",
			},
			httpClient: &http.Client{
				Timeout: 30 * time.Second,
			},
			closed: false,
		}
		logger.Info("LLM客户端初始化完成")
	})
	return instance
}

// Initialize 初始化客户端配置
func (c *LLMClient) Initialize(config *Config) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.closed {
		return fmt.Errorf("客户端已关闭")
	}

	if config == nil {
		return fmt.Errorf("配置不能为空")
	}

	if config.APIKey == "" {
		return fmt.Errorf("API密钥不能为空")
	}

	// 设置默认值
	if config.BaseURL == "" {
		switch config.Provider {
		case ProviderOpenAI:
			config.BaseURL = "https://api.openai.com/v1"
		case ProviderClaude:
			config.BaseURL = "https://api.anthropic.com/v1"
		case ProviderGemini:
			config.BaseURL = "https://generativelanguage.googleapis.com/v1"
		case ProviderOpenRouter:
			config.BaseURL = "https://openrouter.ai/api/v1"
		default:
			config.BaseURL = "https://api.openai.com/v1"
		}
	}

	if config.Model == "" {
		switch config.Provider {
		case ProviderOpenAI:
			config.Model = "gpt-3.5-turbo"
		case ProviderClaude:
			config.Model = "claude-3-sonnet-20240229"
		case ProviderGemini:
			config.Model = "gemini-pro"
		case ProviderOpenRouter:
			config.Model = "openai/gpt-3.5-turbo"
		default:
			config.Model = "gpt-3.5-turbo"
		}
	}

	if config.Temperature == 0 {
		config.Temperature = 0.7
	}

	if config.MaxTokens == 0 {
		config.MaxTokens = 2048
	}

	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}

	c.config = config
	c.httpClient = &http.Client{
		Timeout: config.Timeout,
	}

	logger.Info(fmt.Sprintf("LLM客户端配置更新: Provider=%s, Model=%s", config.Provider, config.Model))
	return nil
}

// SetSystemPrompt 设置系统提示词
func (c *LLMClient) SetSystemPrompt(prompt string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.closed {
		return fmt.Errorf("客户端已关闭")
	}

	c.config.SystemPrompt = prompt
	logger.Info("系统提示词已更新")
	return nil
}

// ChatStream 流式对话接口
func (c *LLMClient) ChatStream(messages []Message) (<-chan StreamResponse, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if c.closed {
		return nil, fmt.Errorf("客户端已关闭")
	}

	if len(messages) == 0 {
		return nil, fmt.Errorf("消息列表不能为空")
	}

	// 创建响应通道
	responseChan := make(chan StreamResponse, 100)

	// 准备消息列表（添加系统提示词）
	fullMessages := c.prepareMessages(messages)

	// 启动goroutine处理流式响应
	go func() {
		defer close(responseChan)
		
		err := c.performStreamRequest(fullMessages, responseChan)
		if err != nil {
			responseChan <- StreamResponse{
				Content: "",
				Done:    true,
				Error:   err.Error(),
			}
		}
	}()

	return responseChan, nil
}

// Chat 普通对话接口
func (c *LLMClient) Chat(messages []Message) (string, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if c.closed {
		return "", fmt.Errorf("客户端已关闭")
	}

	if len(messages) == 0 {
		return "", fmt.Errorf("消息列表不能为空")
	}

	// 准备消息列表
	fullMessages := c.prepareMessages(messages)

	// 执行请求
	response, err := c.performRequest(fullMessages)
	if err != nil {
		return "", err
	}

	return response, nil
}

// prepareMessages 准备消息列表（添加系统提示词）
func (c *LLMClient) prepareMessages(messages []Message) []Message {
	fullMessages := make([]Message, 0, len(messages)+1)
	
	// 添加系统提示词
	if c.config.SystemPrompt != "" {
		fullMessages = append(fullMessages, Message{
			Role:    "system",
			Content: c.config.SystemPrompt,
		})
	}
	
	// 添加用户消息
	fullMessages = append(fullMessages, messages...)
	
	return fullMessages
}

// performStreamRequest 执行流式请求
func (c *LLMClient) performStreamRequest(messages []Message, responseChan chan<- StreamResponse) error {
	var lastErr error
	
	for attempt := 0; attempt < c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			logger.Info(fmt.Sprintf("重试流式请求，第 %d 次尝试", attempt+1))
			time.Sleep(time.Duration(attempt) * time.Second)
		}

		err := c.doStreamRequest(messages, responseChan)
		if err == nil {
			return nil
		}
		
		lastErr = err
		logger.Warn(fmt.Sprintf("流式请求失败: %v", err))
	}
	
	return fmt.Errorf("流式请求失败，已重试 %d 次: %v", c.config.MaxRetries, lastErr)
}

// doStreamRequest 执行单次流式请求
func (c *LLMClient) doStreamRequest(messages []Message, responseChan chan<- StreamResponse) error {
	// 构建请求体
	requestBody := map[string]interface{}{
		"model":       c.config.Model,
		"messages":    messages,
		"temperature": c.config.Temperature,
		"max_tokens":  c.config.MaxTokens,
		"stream":      true,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("序列化请求体失败: %v", err)
	}

	// 创建HTTP请求
	url := c.getEndpointURL()
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	c.setRequestHeaders(req)

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 处理流式响应
	return c.processStreamResponse(resp.Body, responseChan)
}

// performRequest 执行普通请求
func (c *LLMClient) performRequest(messages []Message) (string, error) {
	var lastErr error
	
	for attempt := 0; attempt < c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			logger.Info(fmt.Sprintf("重试请求，第 %d 次尝试", attempt+1))
			time.Sleep(time.Duration(attempt) * time.Second)
		}

		response, err := c.doRequest(messages)
		if err == nil {
			return response, nil
		}
		
		lastErr = err
		logger.Warn(fmt.Sprintf("请求失败: %v", err))
	}
	
	return "", fmt.Errorf("请求失败，已重试 %d 次: %v", c.config.MaxRetries, lastErr)
}

// doRequest 执行单次请求
func (c *LLMClient) doRequest(messages []Message) (string, error) {
	// 构建请求体
	requestBody := map[string]interface{}{
		"model":       c.config.Model,
		"messages":    messages,
		"temperature": c.config.Temperature,
		"max_tokens":  c.config.MaxTokens,
		"stream":      false,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("序列化请求体失败: %v", err)
	}

	// 创建HTTP请求
	url := c.getEndpointURL()
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	c.setRequestHeaders(req)

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}

	// 提取内容
	return c.extractContent(response)
}

// getEndpointURL 获取API端点URL
func (c *LLMClient) getEndpointURL() string {
	switch c.config.Provider {
	case ProviderOpenAI, ProviderOpenRouter:
		return c.config.BaseURL + "/chat/completions"
	case ProviderClaude:
		return c.config.BaseURL + "/messages"
	case ProviderGemini:
		return c.config.BaseURL + "/models/" + c.config.Model + ":generateContent"
	default:
		return c.config.BaseURL + "/chat/completions"
	}
}

// setRequestHeaders 设置请求头
func (c *LLMClient) setRequestHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	
	switch c.config.Provider {
	case ProviderOpenAI, ProviderOpenRouter:
		req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	case ProviderClaude:
		req.Header.Set("x-api-key", c.config.APIKey)
		req.Header.Set("anthropic-version", "2023-06-01")
	case ProviderGemini:
		req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	default:
		req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	}
}

// processStreamResponse 处理流式响应
func (c *LLMClient) processStreamResponse(body io.Reader, responseChan chan<- StreamResponse) error {
	scanner := bufio.NewScanner(body)
	
	for scanner.Scan() {
		line := scanner.Text()
		
		// 跳过空行和非数据行
		if line == "" || !strings.HasPrefix(line, "data: ") {
			continue
		}
		
		// 移除 "data: " 前缀
		data := strings.TrimPrefix(line, "data: ")
		
		// 检查是否为结束标记
		if data == "[DONE]" {
			responseChan <- StreamResponse{
				Content: "",
				Done:    true,
				Error:   "",
			}
			return nil
		}
		
		// 解析JSON数据
		var streamData map[string]interface{}
		if err := json.Unmarshal([]byte(data), &streamData); err != nil {
			logger.Warn(fmt.Sprintf("解析流式数据失败: %v, 数据: %s", err, data))
			continue
		}
		
		// 提取内容
		content := c.extractStreamContent(streamData)
		if content != "" {
			responseChan <- StreamResponse{
				Content: content,
				Done:    false,
				Error:   "",
			}
		}
	}
	
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取流式响应失败: %v", err)
	}
	
	// 发送完成信号
	responseChan <- StreamResponse{
		Content: "",
		Done:    true,
		Error:   "",
	}
	
	return nil
}

// extractContent 从响应中提取内容
func (c *LLMClient) extractContent(response map[string]interface{}) (string, error) {
	switch c.config.Provider {
	case ProviderOpenAI, ProviderOpenRouter:
		if choices, ok := response["choices"].([]interface{}); ok && len(choices) > 0 {
			if choice, ok := choices[0].(map[string]interface{}); ok {
				if message, ok := choice["message"].(map[string]interface{}); ok {
					if content, ok := message["content"].(string); ok {
						return content, nil
					}
				}
			}
		}
	case ProviderClaude:
		if content, ok := response["content"].([]interface{}); ok && len(content) > 0 {
			if item, ok := content[0].(map[string]interface{}); ok {
				if text, ok := item["text"].(string); ok {
					return text, nil
				}
			}
		}
	case ProviderGemini:
		if candidates, ok := response["candidates"].([]interface{}); ok && len(candidates) > 0 {
			if candidate, ok := candidates[0].(map[string]interface{}); ok {
				if content, ok := candidate["content"].(map[string]interface{}); ok {
					if parts, ok := content["parts"].([]interface{}); ok && len(parts) > 0 {
						if part, ok := parts[0].(map[string]interface{}); ok {
							if text, ok := part["text"].(string); ok {
								return text, nil
							}
						}
					}
				}
			}
		}
	}
	
	return "", fmt.Errorf("无法从响应中提取内容")
}

// extractStreamContent 从流式响应中提取内容
func (c *LLMClient) extractStreamContent(data map[string]interface{}) string {
	switch c.config.Provider {
	case ProviderOpenAI, ProviderOpenRouter:
		if choices, ok := data["choices"].([]interface{}); ok && len(choices) > 0 {
			if choice, ok := choices[0].(map[string]interface{}); ok {
				if delta, ok := choice["delta"].(map[string]interface{}); ok {
					if content, ok := delta["content"].(string); ok {
						return content
					}
				}
			}
		}
	case ProviderClaude:
		if delta, ok := data["delta"].(map[string]interface{}); ok {
			if text, ok := delta["text"].(string); ok {
				return text
			}
		}
	case ProviderGemini:
		if candidates, ok := data["candidates"].([]interface{}); ok && len(candidates) > 0 {
			if candidate, ok := candidates[0].(map[string]interface{}); ok {
				if content, ok := candidate["content"].(map[string]interface{}); ok {
					if parts, ok := content["parts"].([]interface{}); ok && len(parts) > 0 {
						if part, ok := parts[0].(map[string]interface{}); ok {
							if text, ok := part["text"].(string); ok {
								return text
							}
						}
					}
				}
			}
		}
	}
	
	return ""
}

// GetConfig 获取当前配置
func (c *LLMClient) GetConfig() *Config {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	// 返回配置的副本
	configCopy := *c.config
	return &configCopy
}

// IsReady 检查客户端是否就绪
func (c *LLMClient) IsReady() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	return !c.closed && c.config != nil && c.config.APIKey != ""
}

// Close 关闭客户端
func (c *LLMClient) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	if c.closed {
		return nil
	}
	
	c.closed = true
	if c.cancelFunc != nil {
		c.cancelFunc()
	}
	
	logger.Info("LLM客户端已关闭")
	return nil
}

// Reopen 重新打开客户端（主要用于测试）
func (c *LLMClient) Reopen() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.closed = false
	c.cancelFunc = nil
	
	logger.Info("LLM客户端已重新打开")
	return nil
}

// 包级别的便利函数

// Initialize 初始化全局LLM客户端
func Initialize(config *Config) error {
	return GetInstance().Initialize(config)
}

// SetSystemPrompt 设置全局系统提示词
func SetSystemPrompt(prompt string) error {
	return GetInstance().SetSystemPrompt(prompt)
}

// ChatStream 全局流式对话
func ChatStream(messages []Message) (<-chan StreamResponse, error) {
	return GetInstance().ChatStream(messages)
}

// Chat 全局普通对话
func Chat(messages []Message) (string, error) {
	return GetInstance().Chat(messages)
}

// IsReady 检查全局客户端是否就绪
func IsReady() bool {
	return GetInstance().IsReady()
}

// GetConfig 获取全局客户端配置
func GetConfig() *Config {
	return GetInstance().GetConfig()
}

// Close 关闭全局客户端
func Close() error {
	return GetInstance().Close()
}
