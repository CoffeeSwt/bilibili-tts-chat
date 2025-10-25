package llm

import (
	"testing"
	"time"
)

func TestLLMClientSingleton(t *testing.T) {
	// 测试单例模式
	client1 := GetInstance()
	client2 := GetInstance()
	
	if client1 != client2 {
		t.Error("LLMClient应该是单例模式")
	}
}

func TestInitialize(t *testing.T) {
	client := GetInstance()
	
	// 测试无效配置
	err := client.Initialize(&Config{})
	if err == nil {
		t.Error("空API密钥应该返回错误")
	}
	
	// 测试有效配置
	config := &Config{
		Provider:     ProviderOpenAI,
		APIKey:       "test-api-key",
		Model:        "gpt-3.5-turbo",
		Temperature:  0.7,
		MaxTokens:    1000,
		SystemPrompt: "你是一个测试助手",
	}
	
	err = client.Initialize(config)
	if err != nil {
		t.Errorf("初始化失败: %v", err)
	}
	
	// 验证配置
	currentConfig := client.GetConfig()
	if currentConfig.APIKey != config.APIKey {
		t.Error("API密钥配置不正确")
	}
	
	if currentConfig.Model != config.Model {
		t.Error("模型配置不正确")
	}
}

func TestSetSystemPrompt(t *testing.T) {
	client := GetInstance()
	
	// 初始化客户端
	config := &Config{
		Provider: ProviderOpenAI,
		APIKey:   "test-api-key",
	}
	client.Initialize(config)
	
	// 设置系统提示词
	newPrompt := "你是一个专业的编程助手"
	err := client.SetSystemPrompt(newPrompt)
	if err != nil {
		t.Errorf("设置系统提示词失败: %v", err)
	}
	
	// 验证系统提示词
	currentConfig := client.GetConfig()
	if currentConfig.SystemPrompt != newPrompt {
		t.Error("系统提示词设置不正确")
	}
}

func TestIsReady(t *testing.T) {
	client := GetInstance()
	
	// 确保客户端是打开状态并清除之前的配置
	client.Reopen()
	client.config = nil
	
	// 未配置API密钥时应该不就绪（Initialize会失败，config保持nil）
	err := client.Initialize(&Config{Provider: ProviderOpenAI})
	if err == nil {
		t.Error("未配置API密钥时Initialize应该失败")
	}
	if client.IsReady() {
		t.Error("未配置API密钥时应该不就绪")
	}
	
	// 配置API密钥后应该就绪
	err = client.Initialize(&Config{
		Provider: ProviderOpenAI,
		APIKey:   "test-api-key",
	})
	if err != nil {
		t.Errorf("配置API密钥后Initialize应该成功: %v", err)
	}
	if !client.IsReady() {
		t.Error("配置API密钥后应该就绪")
	}
}

func TestProviderDefaults(t *testing.T) {
	client := GetInstance()
	
	// 测试OpenAI默认配置
	config := &Config{
		Provider: ProviderOpenAI,
		APIKey:   "test-key",
	}
	client.Initialize(config)
	
	currentConfig := client.GetConfig()
	if currentConfig.BaseURL != "https://api.openai.com/v1" {
		t.Error("OpenAI默认BaseURL不正确")
	}
	if currentConfig.Model != "gpt-3.5-turbo" {
		t.Error("OpenAI默认模型不正确")
	}
	
	// 测试Claude默认配置
	config = &Config{
		Provider: ProviderClaude,
		APIKey:   "test-key",
	}
	client.Initialize(config)
	
	currentConfig = client.GetConfig()
	if currentConfig.BaseURL != "https://api.anthropic.com/v1" {
		t.Error("Claude默认BaseURL不正确")
	}
	if currentConfig.Model != "claude-3-sonnet-20240229" {
		t.Error("Claude默认模型不正确")
	}
}

func TestMessageValidation(t *testing.T) {
	client := GetInstance()
	
	// 初始化客户端
	client.Initialize(&Config{
		Provider: ProviderOpenAI,
		APIKey:   "test-api-key",
	})
	
	// 测试空消息列表
	_, err := client.ChatStream([]Message{})
	if err == nil {
		t.Error("空消息列表应该返回错误")
	}
	
	_, err = client.Chat([]Message{})
	if err == nil {
		t.Error("空消息列表应该返回错误")
	}
}

func TestPrepareMessages(t *testing.T) {
	client := GetInstance()
	
	// 设置系统提示词
	client.Initialize(&Config{
		Provider:     ProviderOpenAI,
		APIKey:       "test-api-key",
		SystemPrompt: "你是一个助手",
	})
	
	// 准备消息
	userMessages := []Message{
		{Role: "user", Content: "你好"},
	}
	
	fullMessages := client.prepareMessages(userMessages)
	
	// 应该包含系统消息和用户消息
	if len(fullMessages) != 2 {
		t.Error("准备的消息数量不正确")
	}
	
	if fullMessages[0].Role != "system" {
		t.Error("第一条消息应该是系统消息")
	}
	
	if fullMessages[1].Role != "user" {
		t.Error("第二条消息应该是用户消息")
	}
}

func TestGetEndpointURL(t *testing.T) {
	client := GetInstance()
	
	// 测试不同提供商的端点URL
	testCases := []struct {
		provider    ProviderType
		expectedURL string
	}{
		{ProviderOpenAI, "https://api.openai.com/v1/chat/completions"},
		{ProviderClaude, "https://api.anthropic.com/v1/messages"},
		{ProviderOpenRouter, "https://openrouter.ai/api/v1/chat/completions"},
	}
	
	for _, tc := range testCases {
		client.Initialize(&Config{
			Provider: tc.provider,
			APIKey:   "test-key",
		})
		
		url := client.getEndpointURL()
		if url != tc.expectedURL {
			t.Errorf("Provider %s 的端点URL不正确，期望: %s, 实际: %s", 
				tc.provider, tc.expectedURL, url)
		}
	}
}

func TestClose(t *testing.T) {
	client := GetInstance()
	
	// 初始化客户端
	client.Initialize(&Config{
		Provider: ProviderOpenAI,
		APIKey:   "test-api-key",
	})
	
	// 关闭客户端
	err := client.Close()
	if err != nil {
		t.Errorf("关闭客户端失败: %v", err)
	}
	
	// 关闭后的操作应该失败
	err = client.SetSystemPrompt("test")
	if err == nil {
		t.Error("关闭后设置系统提示词应该失败")
	}
	
	_, err = client.ChatStream([]Message{{Role: "user", Content: "test"}})
	if err == nil {
		t.Error("关闭后流式对话应该失败")
	}
	
	_, err = client.Chat([]Message{{Role: "user", Content: "test"}})
	if err == nil {
		t.Error("关闭后普通对话应该失败")
	}
}

func TestPackageLevelFunctions(t *testing.T) {
	// 重新打开客户端
	GetInstance().Reopen()
	
	// 测试包级别函数
	config := &Config{
		Provider: ProviderOpenAI,
		APIKey:   "test-api-key",
	}
	
	err := Initialize(config)
	if err != nil {
		t.Errorf("包级别初始化失败: %v", err)
	}
	
	if !IsReady() {
		t.Error("包级别IsReady应该返回true")
	}
	
	err = SetSystemPrompt("测试提示词")
	if err != nil {
		t.Errorf("包级别设置系统提示词失败: %v", err)
	}
	
	currentConfig := GetConfig()
	if currentConfig.SystemPrompt != "测试提示词" {
		t.Error("包级别获取配置不正确")
	}
}

func TestConfigDefaults(t *testing.T) {
	client := GetInstance()
	
	// 测试默认值设置
	config := &Config{
		Provider: ProviderOpenAI,
		APIKey:   "test-key",
		// 其他字段使用默认值
	}
	
	client.Initialize(config)
	currentConfig := client.GetConfig()
	
	if currentConfig.Temperature != 0.7 {
		t.Error("默认温度值不正确")
	}
	
	if currentConfig.MaxTokens != 2048 {
		t.Error("默认最大token数不正确")
	}
	
	if currentConfig.Timeout != 30*time.Second {
		t.Error("默认超时时间不正确")
	}
	
	if currentConfig.MaxRetries != 3 {
		t.Error("默认重试次数不正确")
	}
}

func TestConcurrentAccess(t *testing.T) {
	client := GetInstance()
	
	// 初始化客户端
	client.Initialize(&Config{
		Provider: ProviderOpenAI,
		APIKey:   "test-api-key",
	})
	
	// 并发访问测试
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func(id int) {
			// 并发设置系统提示词
			client.SetSystemPrompt("并发测试提示词")
			
			// 并发获取配置
			config := client.GetConfig()
			if config == nil {
				t.Error("并发获取配置失败")
			}
			
			// 并发检查就绪状态
			client.IsReady()
			
			done <- true
		}(i)
	}
	
	// 等待所有goroutine完成
	for i := 0; i < 10; i++ {
		<-done
	}
}