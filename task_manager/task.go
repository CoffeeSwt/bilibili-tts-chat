package task_manager

import (
	"context"
	"fmt"
	"strings"

	"github.com/CoffeeSwt/bilibili-tts-chat/config"
	"github.com/CoffeeSwt/bilibili-tts-chat/llm"
	"github.com/CoffeeSwt/bilibili-tts-chat/logger"
	"github.com/CoffeeSwt/bilibili-tts-chat/tts_api"
	"github.com/CoffeeSwt/bilibili-tts-chat/voice"
)

// PlayEventTasks 处理事件任务：从task_manager获取文本、调用LLM、生成语音、播报
func PlayEventTasks(ctx context.Context) {
	// 检查是否有正在运行的任务
	if !IsTaskRunning() {
		logger.Warn("PlayEventTasks: 没有正在运行的任务")
		return
	}

	// 从task_manager获取文本并完成任务
	texts := CompleteTask()

	// 参数验证
	if len(texts) == 0 {
		logger.Warn("PlayEventTasks: 从任务管理器获取的文本列表为空")
		return
	}

	logger.Info("PlayEventTasks: 开始处理任务", "texts_count", len(texts))

	// 2. 生成提示词并调用大模型
	prompt := llm.GeneratePrompt(texts)
	logger.Info("PlayEventTasks: 提示词生成完成", "prompt_length", len(prompt))

	// 检查上下文是否已取消
	select {
	case <-ctx.Done():
		logger.Info("PlayEventTasks: 任务被取消")
		return
	default:
	}

	// 3. 调用LLM流式对话
	llmResponse, err := callLLMStream(ctx, prompt)
	if err != nil {
		logger.Error("PlayEventTasks: LLM调用失败", "error", err)
		return
	}

	if llmResponse == "" {
		logger.Warn("PlayEventTasks: LLM返回空响应")
		return
	}

	logger.Info("PlayEventTasks: LLM响应获取完成", "response_length", len(llmResponse))

	// 检查上下文是否已取消
	select {
	case <-ctx.Done():
		logger.Info("PlayEventTasks: 任务被取消")
		return
	default:
	}

	// 4. 将大模型返回的内容转换为语音
	audioData, err := generateSpeech(llmResponse)
	if err != nil {
		logger.Error("PlayEventTasks: 语音生成失败", "error", err)
		return
	}

	logger.Info("PlayEventTasks: 语音生成完成", "audio_size", len(audioData))

	// 检查上下文是否已取消
	select {
	case <-ctx.Done():
		logger.Info("PlayEventTasks: 任务被取消")
		return
	default:
	}

	// 5. 播报语音并等待播报完成
	err = playAudioAndWait(ctx, audioData)
	if err != nil {
		logger.Error("PlayEventTasks: 音频播放失败", "error", err)
		return
	}

	logger.Info("PlayEventTasks: 任务完成")
}

// callLLMStream 调用LLM流式对话并收集完整响应
func callLLMStream(ctx context.Context, prompt string) (string, error) {
	// 构建消息
	messages := []llm.Message{
		{
			Role:    "user",
			Content: prompt,
		},
	}

	var responseChan <-chan llm.StreamResponse
	var err error

	// 检查是否启用Mock模式
	if config.GetLLMMockEnabled() {
		// Mock模式下直接调用模拟函数，不需要检查客户端就绪状态
		responseChan, err = llm.ChatStreamWithMock(messages)
		if err != nil {
			return "", fmt.Errorf("启动模拟流式对话失败: %v", err)
		}
	} else {
		// 非Mock模式下检查LLM客户端是否就绪
		if !llm.IsReady() {
			return "", fmt.Errorf("LLM客户端未就绪")
		}
		// 调用真实的LLM流式对话
		responseChan, err = llm.ChatStream(messages)
		if err != nil {
			return "", fmt.Errorf("启动流式对话失败: %v", err)
		}
	}

	// 收集流式响应
	var responseBuilder strings.Builder
	for {
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("LLM调用被取消")
		case response, ok := <-responseChan:
			if !ok {
				// 通道已关闭，返回收集到的响应
				return responseBuilder.String(), nil
			}

			if response.Error != "" {
				return "", fmt.Errorf("LLM响应错误: %s", response.Error)
			}

			if response.Content != "" {
				responseBuilder.WriteString(response.Content)
			}

			if response.Done {
				// 响应完成
				return responseBuilder.String(), nil
			}
		}
	}
}

// generateSpeech 生成语音
func generateSpeech(text string) ([]byte, error) {
	// 获取随机音色
	voice := config.GetRandomVoice()
	if voice == nil {
		return nil, fmt.Errorf("无法获取音色配置")
	}

	logger.Info("generateSpeech: 使用音色", "voice_name", voice.Name)

	// 调用TTS API生成语音
	ttsResult, err := tts_api.GenerateSpeech(text, voice)
	if err != nil {
		return nil, fmt.Errorf("TTS生成失败: %v", err)
	}

	if ttsResult == nil || len(ttsResult.AudioData) == 0 {
		return nil, fmt.Errorf("TTS返回空音频数据")
	}

	return ttsResult.AudioData, nil
}

// playAudioAndWait 播放音频并等待完成
func playAudioAndWait(ctx context.Context, audioData []byte) error {
	// 使用带完成信号的播放接口
	completionChan, err := voice.PlayAudioWithCompletion(audioData, 80) // 音量设置为80%
	if err != nil {
		return fmt.Errorf("启动音频播放失败: %v", err)
	}

	// 等待播放完成或上下文取消
	select {
	case <-ctx.Done():
		return fmt.Errorf("音频播放被取消")
	case <-completionChan:
		// 播放完成
		logger.Info("playAudioAndWait: 音频播放完成")
		return nil
	}
}
