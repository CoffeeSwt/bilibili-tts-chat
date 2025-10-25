package workflow

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/CoffeeSwt/bilibili-tts-chat/config"
	"github.com/CoffeeSwt/bilibili-tts-chat/llm"
	"github.com/CoffeeSwt/bilibili-tts-chat/logger"
	"github.com/CoffeeSwt/bilibili-tts-chat/task_manager"
	"github.com/CoffeeSwt/bilibili-tts-chat/tts_api"
	"github.com/CoffeeSwt/bilibili-tts-chat/voice"
)

// WorkflowConfig 工作流配置
type WorkflowConfig struct {
	CheckInterval  time.Duration // 检查任务间隔
	MaxRetries     int           // 最大重试次数
	RetryDelay     time.Duration // 重试延迟
	Volume         int           // 音量设置 (1-100)
	LLMTimeout     time.Duration // LLM超时时间
	TTSTimeout     time.Duration // TTS超时时间
	MinTextLength  int           // 最小文本长度
	MaxTextLength  int           // 最大文本长度
	SystemPrompt   string        // 系统提示词
	EnableDebugLog bool          // 是否启用调试日志
}

// WorkflowStatus 工作流状态
type WorkflowStatus int

const (
	StatusStopped WorkflowStatus = iota // 已停止
	StatusRunning                       // 运行中
	StatusPaused                        // 暂停中
)

// WorkflowProcessor 工作流处理器
type WorkflowProcessor struct {
	mutex            sync.RWMutex              // 读写锁
	config           *WorkflowConfig           // 配置
	status           WorkflowStatus            // 当前状态
	ctx              context.Context           // 上下文
	cancel           context.CancelFunc        // 取消函数
	taskManager      *task_manager.TaskManager // 任务管理器
	llmClient        *llm.LLMClient            // LLM客户端
	voice            *config.Voice             // 语音配置
	taskCompleteChan chan struct{}             // 任务完成事件通道

	// 统计信息
	stats struct {
		TotalTasks      int64     // 总任务数
		SuccessfulTasks int64     // 成功任务数
		FailedTasks     int64     // 失败任务数
		LastTaskTime    time.Time // 最后任务时间
		StartTime       time.Time // 启动时间
	}
}

var (
	instance *WorkflowProcessor
	once     sync.Once
)

// GetInstance 获取工作流处理器单例实例
func GetInstance() *WorkflowProcessor {
	once.Do(func() {
		instance = &WorkflowProcessor{
			config: &WorkflowConfig{
				CheckInterval:  2 * time.Second,
				MaxRetries:     3,
				RetryDelay:     1 * time.Second,
				Volume:         80,
				LLMTimeout:     30 * time.Second,
				TTSTimeout:     15 * time.Second,
				MinTextLength:  1,
				MaxTextLength:  1000,
				SystemPrompt:   "你是一个智能助手，请根据用户的消息生成简洁、友好的回复。",
				EnableDebugLog: true,
			},
			status:           StatusStopped,
			taskManager:      task_manager.GetInstance(),
			llmClient:        llm.GetInstance(),
			taskCompleteChan: make(chan struct{}, 10), // 缓冲通道，避免阻塞
		}
		logger.Info("工作流处理器初始化完成")
	})
	return instance
}

// SetConfig 设置工作流配置
func (wp *WorkflowProcessor) SetConfig(config *WorkflowConfig) error {
	wp.mutex.Lock()
	defer wp.mutex.Unlock()

	if config == nil {
		return fmt.Errorf("配置不能为空")
	}

	// 验证配置参数
	if config.CheckInterval < 100*time.Millisecond {
		config.CheckInterval = 100 * time.Millisecond
	}
	if config.MaxRetries < 1 {
		config.MaxRetries = 1
	}
	if config.Volume < 1 || config.Volume > 100 {
		config.Volume = 80
	}
	if config.MinTextLength < 1 {
		config.MinTextLength = 1
	}
	if config.MaxTextLength < config.MinTextLength {
		config.MaxTextLength = config.MinTextLength + 100
	}

	wp.config = config
	logger.Info("工作流配置已更新")
	return nil
}

// SetVoiceConfig 设置语音配置
func (wp *WorkflowProcessor) SetVoiceConfig(voice *config.Voice) error {
	wp.mutex.Lock()
	defer wp.mutex.Unlock()

	if voice == nil {
		return fmt.Errorf("语音配置不能为空")
	}

	wp.voice = voice
	logger.Info(fmt.Sprintf("语音配置已设置: %s", voice.VoiceType))
	return nil
}

// Start 启动工作流处理器
func (wp *WorkflowProcessor) Start() error {
	wp.mutex.Lock()
	defer wp.mutex.Unlock()

	if wp.status == StatusRunning {
		return fmt.Errorf("工作流处理器已在运行中")
	}

	// 检查必要的配置
	if !wp.llmClient.IsReady() {
		return fmt.Errorf("LLM客户端未就绪，请先设置LLM配置")
	}

	if wp.voice == nil {
		return fmt.Errorf("语音配置未设置，请先设置语音配置")
	}

	// 创建上下文
	wp.ctx, wp.cancel = context.WithCancel(context.Background())
	wp.status = StatusRunning
	wp.stats.StartTime = time.Now()

	// 启动主处理循环
	go wp.processLoop()

	// 触发初始任务检查事件
	go func() {
		time.Sleep(100 * time.Millisecond) // 稍等一下确保processLoop已启动
		select {
		case wp.taskCompleteChan <- struct{}{}:
		default:
		}
	}()

	logger.Info("工作流处理器已启动 (事件驱动模式)")
	return nil
}

// Stop 停止工作流处理器
func (wp *WorkflowProcessor) Stop() error {
	wp.mutex.Lock()
	defer wp.mutex.Unlock()

	if wp.status == StatusStopped {
		return fmt.Errorf("工作流处理器已停止")
	}

	if wp.cancel != nil {
		wp.cancel()
	}

	wp.status = StatusStopped
	logger.Info("工作流处理器已停止")
	return nil
}

// IsRunning 检查是否运行中
func (wp *WorkflowProcessor) IsRunning() bool {
	wp.mutex.RLock()
	defer wp.mutex.RUnlock()
	return wp.status == StatusRunning
}

// GetStatus 获取当前状态
func (wp *WorkflowProcessor) GetStatus() WorkflowStatus {
	wp.mutex.RLock()
	defer wp.mutex.RUnlock()
	return wp.status
}

// GetStats 获取统计信息
func (wp *WorkflowProcessor) GetStats() map[string]interface{} {
	wp.mutex.RLock()
	defer wp.mutex.RUnlock()

	uptime := time.Duration(0)
	if !wp.stats.StartTime.IsZero() {
		uptime = time.Since(wp.stats.StartTime)
	}

	return map[string]interface{}{
		"status":           wp.status,
		"total_tasks":      wp.stats.TotalTasks,
		"successful_tasks": wp.stats.SuccessfulTasks,
		"failed_tasks":     wp.stats.FailedTasks,
		"last_task_time":   wp.stats.LastTaskTime,
		"uptime":           uptime,
		"success_rate":     wp.calculateSuccessRate(),
	}
}

// calculateSuccessRate 计算成功率
func (wp *WorkflowProcessor) calculateSuccessRate() float64 {
	if wp.stats.TotalTasks == 0 {
		return 0.0
	}
	return float64(wp.stats.SuccessfulTasks) / float64(wp.stats.TotalTasks) * 100.0
}

// processLoop 主处理循环
func (wp *WorkflowProcessor) processLoop() {
	logger.Info("工作流处理循环已启动 (事件驱动模式)")

	// 启动时检查一次是否有待处理任务
	wp.checkAndProcessTasks()

	for {
		select {
		case <-wp.ctx.Done():
			logger.Info("工作流处理循环已退出")
			return
		case <-wp.taskCompleteChan:
			// 收到任务完成事件，立即检查新任务
			if wp.config.EnableDebugLog {
				logger.Debug("收到任务完成事件，检查新任务")
			}
			wp.checkAndProcessTasks()
		}
	}
}

// checkAndProcessTasks 检查并处理任务
func (wp *WorkflowProcessor) checkAndProcessTasks() {
	// 检查是否有正在运行的任务
	if !wp.taskManager.IsTaskRunning() {
		return
	}

	// 获取当前任务的文本
	texts := wp.taskManager.GetCurrentTexts()
	if len(texts) == 0 {
		return
	}

	// 检查文本长度
	totalLength := wp.calculateTotalTextLength(texts)
	if totalLength < wp.config.MinTextLength {
		if wp.config.EnableDebugLog {
			logger.Debug(fmt.Sprintf("文本长度不足，跳过处理: %d < %d", totalLength, wp.config.MinTextLength))
		}
		return
	}

	// 完成当前任务并获取所有文本
	allTexts := wp.taskManager.CompleteTask()
	if len(allTexts) == 0 {
		return
	}

	// 处理任务
	go wp.processTask(allTexts)
}

// calculateTotalTextLength 计算文本总长度
func (wp *WorkflowProcessor) calculateTotalTextLength(texts []string) int {
	total := 0
	for _, text := range texts {
		total += len(strings.TrimSpace(text))
	}
	return total
}

// processTask 处理单个任务
func (wp *WorkflowProcessor) processTask(texts []string) {
	wp.mutex.Lock()
	wp.stats.TotalTasks++
	wp.stats.LastTaskTime = time.Now()
	taskID := wp.stats.TotalTasks
	wp.mutex.Unlock()

	logger.Info(fmt.Sprintf("开始处理任务 #%d，文本数量: %d", taskID, len(texts)))

	// 重试机制
	var lastError error
	for attempt := 1; attempt <= wp.config.MaxRetries; attempt++ {
		if wp.config.EnableDebugLog {
			logger.Debug(fmt.Sprintf("任务 #%d 第 %d 次尝试", taskID, attempt))
		}

		err := wp.processTaskInternal(taskID, texts)
		if err == nil {
			wp.mutex.Lock()
			wp.stats.SuccessfulTasks++
			wp.mutex.Unlock()
			logger.Info(fmt.Sprintf("任务 #%d 处理成功", taskID))
			return
		}

		lastError = err
		logger.Warn(fmt.Sprintf("任务 #%d 第 %d 次尝试失败: %v", taskID, attempt, err))

		// 如果不是最后一次尝试，等待重试延迟
		if attempt < wp.config.MaxRetries {
			time.Sleep(wp.config.RetryDelay)
		}
	}

	// 所有重试都失败
	wp.mutex.Lock()
	wp.stats.FailedTasks++
	wp.mutex.Unlock()
	logger.Error(fmt.Sprintf("任务 #%d 处理失败，已达到最大重试次数: %v", taskID, lastError))
}

// processTaskInternal 内部任务处理逻辑
func (wp *WorkflowProcessor) processTaskInternal(taskID int64, texts []string) error {
	// 1. 准备消息
	messages, err := wp.prepareMessages(texts)
	if err != nil {
		return fmt.Errorf("准备消息失败: %v", err)
	}

	// 2. 调用LLM生成回复
	response, err := wp.callLLM(taskID, messages)
	if err != nil {
		return fmt.Errorf("LLM调用失败: %v", err)
	}

	// AI生成完成，立即触发下一个任务检查事件
	select {
	case wp.taskCompleteChan <- struct{}{}:
		if wp.config.EnableDebugLog {
			logger.Debug(fmt.Sprintf("任务 #%d AI生成完成，已触发下一任务检查事件", taskID))
		}
	default:
		// 通道已满，不阻塞
	}

	// 4. 播放语音（异步进行，不影响下一个任务处理）
	go func(resp string) {
		// 3. 转换为语音
		audioData, err := wp.generateSpeech(taskID, resp)
		if err != nil {
			logger.Error(fmt.Sprintf("任务 #%d 语音生成失败: %v", taskID, err))
			return
		}

		err = wp.playAudio(taskID, audioData)
		if err != nil {
			logger.Error(fmt.Sprintf("任务 #%d 音频播放失败: %v", taskID, err))
		}
	}(response)

	return nil
}

// prepareMessages 准备LLM消息
func (wp *WorkflowProcessor) prepareMessages(texts []string) ([]llm.Message, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("文本列表为空")
	}

	// 合并所有文本
	combinedText := strings.Join(texts, "\n")

	// 检查文本长度
	if len(combinedText) > wp.config.MaxTextLength {
		combinedText = combinedText[:wp.config.MaxTextLength] + "..."
	}

	// 创建用户消息
	messages := []llm.Message{
		{
			Role:    "user",
			Content: combinedText,
		},
	}

	return messages, nil
}

// callLLM 调用LLM生成回复
func (wp *WorkflowProcessor) callLLM(taskID int64, messages []llm.Message) (string, error) {
	if wp.config.EnableDebugLog {
		logger.Debug(fmt.Sprintf("任务 #%d 调用LLM，消息数量: %d", taskID, len(messages)))
	}

	// 创建超时上下文
	ctx, cancel := context.WithTimeout(wp.ctx, wp.config.LLMTimeout)
	defer cancel()

	// 调用ChatStream接口
	streamChan, err := wp.llmClient.ChatStream(messages)
	if err != nil {
		return "", fmt.Errorf("启动ChatStream失败: %v", err)
	}

	// 收集流式响应
	var responseBuilder strings.Builder
	for {
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("LLM调用超时")
		case response, ok := <-streamChan:
			if !ok {
				// 流结束
				result := responseBuilder.String()
				if result == "" {
					return "", fmt.Errorf("LLM返回空响应")
				}
				if wp.config.EnableDebugLog {
					logger.Debug(fmt.Sprintf("任务 #%d LLM响应: %s", taskID, result))
				}
				return result, nil
			}

			if response.Error != "" {
				return "", fmt.Errorf("LLM流式响应错误: %v", response.Error)
			}

			responseBuilder.WriteString(response.Content)
		}
	}
}

// generateSpeech 生成语音
func (wp *WorkflowProcessor) generateSpeech(taskID int64, text string) ([]byte, error) {
	if wp.config.EnableDebugLog {
		logger.Debug(fmt.Sprintf("任务 #%d 生成语音，文本长度: %d", taskID, len(text)))
	}

	// 创建超时上下文
	ctx, cancel := context.WithTimeout(wp.ctx, wp.config.TTSTimeout)
	defer cancel()

	// 使用goroutine调用TTS API
	resultChan := make(chan *tts_api.TTSResult, 1)
	errorChan := make(chan error, 1)

	go func() {
		result, err := tts_api.GenerateSpeech(text, wp.voice)
		if err != nil {
			errorChan <- err
			return
		}
		resultChan <- result
	}()

	// 等待结果或超时
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("TTS调用超时")
	case err := <-errorChan:
		return nil, err
	case result := <-resultChan:
		if result == nil || len(result.AudioData) == 0 {
			return nil, fmt.Errorf("TTS返回空音频数据")
		}
		if wp.config.EnableDebugLog {
			logger.Debug(fmt.Sprintf("任务 #%d 语音生成成功，音频大小: %d 字节", taskID, len(result.AudioData)))
		}
		return result.AudioData, nil
	}
}

// playAudio 播放音频
func (wp *WorkflowProcessor) playAudio(taskID int64, audioData []byte) error {
	if wp.config.EnableDebugLog {
		logger.Debug(fmt.Sprintf("任务 #%d 开始播放音频，大小: %d 字节", taskID, len(audioData)))
	}

	// 使用带完成信号的播放方法
	completion, err := voice.PlayAudioWithCompletion(audioData, wp.config.Volume)
	if err != nil {
		return fmt.Errorf("启动音频播放失败: %v", err)
	}

	// 等待播放完成
	select {
	case <-wp.ctx.Done():
		return fmt.Errorf("音频播放被中断")
	case <-completion:
		if wp.config.EnableDebugLog {
			logger.Debug(fmt.Sprintf("任务 #%d 音频播放完成", taskID))
		}
		return nil
	}
}

// 便利函数，直接使用单例实例

// Start 启动工作流处理器
func Start() error {
	return GetInstance().Start()
}

// Stop 停止工作流处理器
func Stop() error {
	return GetInstance().Stop()
}

// SetVoiceConfig 设置语音配置
func SetVoiceConfig(voice *config.Voice) error {
	return GetInstance().SetVoiceConfig(voice)
}

// SetConfig 设置工作流配置
func SetConfig(config *WorkflowConfig) error {
	return GetInstance().SetConfig(config)
}

// IsRunning 检查是否运行中
func IsRunning() bool {
	return GetInstance().IsRunning()
}

// GetStatus 获取当前状态
func GetStatus() WorkflowStatus {
	return GetInstance().GetStatus()
}

// GetStats 获取统计信息
func GetStats() map[string]interface{} {
	return GetInstance().GetStats()
}
