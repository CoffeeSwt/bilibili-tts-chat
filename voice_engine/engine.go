package voice_engine

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/CoffeeSwt/bilibili-tts-chat/config"
	"github.com/CoffeeSwt/bilibili-tts-chat/tts_api"

	"github.com/ebitengine/oto/v3"
	"github.com/hajimehoshi/go-mp3"
)

// 全局音频管理器
var (
	globalAudioManager *AudioManager
	audioManagerOnce   sync.Once
	DefaultVoice       = "zh_female_kefunvsheng_mars_bigtts"
)

// AudioTask 音频播放任务
type AudioTask struct {
	AudioData []byte
	Format    string // "mp3"
	Context   context.Context
	Done      chan error
}

// AudioManager 全局音频管理器
type AudioManager struct {
	otoContext *oto.Context
	taskQueue  chan *AudioTask
	mutex      sync.RWMutex
	isRunning  bool
}

// GetAudioManager 获取全局音频管理器实例（单例模式）
func GetAudioManager() *AudioManager {
	audioManagerOnce.Do(func() {
		globalAudioManager = &AudioManager{
			taskQueue: make(chan *AudioTask, 100), // 缓冲队列
			isRunning: false,
		}
		globalAudioManager.start()
	})
	return globalAudioManager
}

// start 启动音频管理器
func (am *AudioManager) start() {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	if am.isRunning {
		return
	}

	am.isRunning = true
	go am.processQueue()
}

// processQueue 处理音频播放队列
func (am *AudioManager) processQueue() {
	for task := range am.taskQueue {
		// 检查任务上下文是否已取消
		select {
		case <-task.Context.Done():
			task.Done <- task.Context.Err()
			continue
		default:
		}

		// 执行播放任务
		err := am.playAudioTask(task)
		task.Done <- err
	}
}

// playAudioTask 执行音频播放任务
func (am *AudioManager) playAudioTask(task *AudioTask) error {
	// 确保oto上下文已初始化
	if err := am.ensureOtoContext(task); err != nil {
		return err
	}

	// 根据格式播放音频
	switch task.Format {
	case "mp3":
		return am.playMP3WithContext(task.Context, task.AudioData)
	default:
		return fmt.Errorf("unsupported audio format: %s", task.Format)
	}
}

// ensureOtoContext 确保oto上下文已创建（只创建一次）
func (am *AudioManager) ensureOtoContext(task *AudioTask) error {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	if am.otoContext != nil {
		return nil
	}

	// 根据音频格式获取参数
	var sampleRate int
	var channelCount int

	switch task.Format {
	case "mp3":
		reader := bytes.NewReader(task.AudioData)
		decoder, err := mp3.NewDecoder(reader)
		if err != nil {
			return fmt.Errorf("failed to create MP3 decoder for context: %w", err)
		}
		sampleRate = decoder.SampleRate()
		channelCount = 2 // MP3通常是立体声
	default:
		// 默认参数
		sampleRate = 44100
		channelCount = 2
	}

	// 创建oto上下文
	ctx, readyChan, err := oto.NewContext(&oto.NewContextOptions{
		SampleRate:   sampleRate,
		ChannelCount: channelCount,
		Format:       oto.FormatSignedInt16LE,
	})
	if err != nil {
		return fmt.Errorf("failed to create oto context: %w", err)
	}

	// 等待oto准备就绪
	<-readyChan
	am.otoContext = ctx

	log.Println("[AudioManager] oto上下文初始化成功")
	return nil
}

// playMP3WithContext 使用共享上下文播放MP3
func (am *AudioManager) playMP3WithContext(ctx context.Context, audioData []byte) error {
	reader := bytes.NewReader(audioData)
	decoder, err := mp3.NewDecoder(reader)
	if err != nil {
		return fmt.Errorf("failed to create MP3 decoder: %w", err)
	}

	// 使用共享的oto上下文创建播放器
	player := am.otoContext.NewPlayer(decoder)
	defer player.Close()

	// 开始播放
	player.Play()

	// 等待播放完成或上下文取消
	for player.IsPlaying() {
		select {
		case <-ctx.Done():
			player.Close()
			return ctx.Err()
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}

	return nil
}



// PlayAudioWithQueue 通过队列播放音频（线程安全）
func (am *AudioManager) PlayAudioWithQueue(ctx context.Context, audioData []byte, format string) error {
	task := &AudioTask{
		AudioData: audioData,
		Format:    format,
		Context:   ctx,
		Done:      make(chan error, 1),
	}

	// 将任务加入队列
	select {
	case am.taskQueue <- task:
		// 等待任务完成
		return <-task.Done
	case <-ctx.Done():
		return ctx.Err()
	}
}

// GetRandomVoice 随机选择一个音色
func GetRandomVoice() string {
	return config.GetRandomVoiceID()
}

// GetVoiceDescription 获取音色描述
func GetVoiceDescription(voice string) string {
	voiceInfo := config.GetVoiceInfoByID(voice)
	return fmt.Sprintf("%s（%s）", voiceInfo.Name, voiceInfo.Description)
}

// PlayVoice 播放base64编码的音频数据
func PlayVoice(base64Data string) error {
	// 1. 解码base64数据
	audioData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return fmt.Errorf("failed to decode base64 data: %w", err)
	}

	// 2. 创建临时文件
	tempDir := os.TempDir()
	tempFile := filepath.Join(tempDir, fmt.Sprintf("voice_%d.mp3", time.Now().UnixNano()))

	// 写入临时文件
	err = os.WriteFile(tempFile, audioData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// 确保清理临时文件
	defer func() {
		os.Remove(tempFile)
	}()

	// 3. 播放音频文件
	err = playAudioFile(tempFile)
	if err != nil {
		return fmt.Errorf("failed to play audio: %w", err)
	}

	return nil
}

// TextToVoice 直接从文本生成并播放语音
func TextToVoice(text string) error {
	log.Printf("开始TTS转换: %s", text)

	// 调用TTS API获取base64音频数据
	base64Data, err := tts_api.TextToAudioBase64(text)
	if err != nil {
		return fmt.Errorf("TTS转换失败: %w", err)
	}

	log.Printf("TTS转换成功，音频数据长度: %d", len(base64Data))

	// 解码base64数据
	audioData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return fmt.Errorf("base64解码失败: %w", err)
	}

	// 直接播放音频字节数据（支持MP3格式）
	err = PlayVoiceFromBytes(audioData)
	if err != nil {
		return fmt.Errorf("音频播放失败: %w", err)
	}

	log.Println("TTS播放完成")
	return nil
}

// TextToVoiceAsync 异步版本，不阻塞调用者
func TextToVoiceAsync(text, voiceID string) error {
	go func() {
		err := TextToVoiceWithVoice(text, voiceID)
		if err != nil {
			log.Printf("异步TTS播放错误: %v", err)
		}
	}()
	return nil
}

// TextToVoiceWithContext 支持上下文取消的版本
func TextToVoiceWithContext(ctx context.Context, text string) error {
	// 检查上下文是否已取消
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	log.Printf("开始TTS转换 (带上下文): %s", text)

	// 调用TTS API获取base64音频数据
	base64Data, err := tts_api.TextToAudioBase64(text)
	if err != nil {
		return fmt.Errorf("TTS转换失败: %w", err)
	}

	// 再次检查上下文
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	log.Printf("TTS转换成功，音频数据长度: %d", len(base64Data))

	// 解码base64数据
	audioData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return fmt.Errorf("base64解码失败: %w", err)
	}

	// 使用带上下文的播放函数
	err = PlayVoiceFromBytesWithContext(ctx, audioData)
	if err != nil {
		return fmt.Errorf("音频播放失败: %w", err)
	}

	log.Println("TTS播放完成 (带上下文)")
	return nil
}

// TextToVoiceWithVoice 指定音色的同步版本
func TextToVoiceWithVoice(text, voice string) error {
	log.Printf("开始TTS转换 (指定音色: %s): %s", voice, text)

	// 调用TTS API获取base64音频数据（使用指定音色）
	base64Data, err := tts_api.TextToAudioBase64WithVoice(text, voice)
	if err != nil {
		return fmt.Errorf("TTS转换失败: %w", err)
	}

	log.Printf("TTS转换成功，音频数据长度: %d", len(base64Data))

	// 解码base64数据
	audioData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return fmt.Errorf("base64解码失败: %w", err)
	}

	// 直接播放音频字节数据
	err = PlayVoiceFromBytes(audioData)
	if err != nil {
		return fmt.Errorf("音频播放失败: %w", err)
	}

	log.Println("TTS播放完成 (指定音色)")
	return nil
}

// TextToVoiceAsyncWithVoice 指定音色的异步版本
func TextToVoiceAsyncWithVoice(text, voice string) error {
	go func() {
		err := TextToVoiceWithVoice(text, voice)
		if err != nil {
			log.Printf("异步TTS播放错误 (指定音色): %v", err)
		}
	}()
	return nil
}

// TextToVoiceWithContextAndVoice 指定音色的上下文版本
func TextToVoiceWithContextAndVoice(ctx context.Context, text, voice string) error {
	// 检查上下文是否已取消
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	log.Printf("开始TTS转换 (带上下文，指定音色: %s): %s", voice, text)

	// 调用TTS API获取base64音频数据（使用指定音色）
	base64Data, err := tts_api.TextToAudioBase64WithVoice(text, voice)
	if err != nil {
		return fmt.Errorf("TTS转换失败: %w", err)
	}

	// 再次检查上下文
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	log.Printf("TTS转换成功，音频数据长度: %d", len(base64Data))

	// 解码base64数据
	audioData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return fmt.Errorf("base64解码失败: %w", err)
	}

	// 使用带上下文的播放函数
	err = PlayVoiceFromBytesWithContext(ctx, audioData)
	if err != nil {
		return fmt.Errorf("音频播放失败: %w", err)
	}

	log.Println("TTS播放完成 (带上下文，指定音色)")
	return nil
}

// TextToVoiceWithRandomVoice 使用随机音色的同步版本
func TextToVoiceWithRandomVoice(text string) error {
	// 随机选择音色
	voice := GetRandomVoice()
	voiceDesc := GetVoiceDescription(voice)

	log.Printf("开始TTS转换 (随机音色: %s): %s", voiceDesc, text)

	// 调用TTS API获取base64音频数据（使用随机音色）
	base64Data, err := tts_api.TextToAudioBase64WithVoice(text, voice)
	if err != nil {
		return fmt.Errorf("TTS转换失败: %w", err)
	}

	log.Printf("TTS转换成功，音频数据长度: %d", len(base64Data))

	// 解码base64数据
	audioData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return fmt.Errorf("base64解码失败: %w", err)
	}

	// 直接播放音频字节数据
	err = PlayVoiceFromBytes(audioData)
	if err != nil {
		return fmt.Errorf("音频播放失败: %w", err)
	}

	log.Printf("TTS播放完成 (随机音色: %s)", voiceDesc)
	return nil
}

// TextToVoiceAsyncWithRandomVoice 使用随机音色的异步版本
func TextToVoiceAsyncWithRandomVoice(text string) error {
	go func() {
		err := TextToVoiceWithRandomVoice(text)
		if err != nil {
			log.Printf("异步TTS播放错误 (随机音色): %v", err)
		}
	}()
	return nil
}

// TextToVoiceWithContextAndRandomVoice 使用随机音色的上下文版本
func TextToVoiceWithContextAndRandomVoice(ctx context.Context, text string) error {
	// 检查上下文是否已取消
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// 随机选择音色
	voice := GetRandomVoice()
	voiceDesc := GetVoiceDescription(voice)

	log.Printf("开始TTS转换 (带上下文，随机音色: %s): %s", voiceDesc, text)

	// 调用TTS API获取base64音频数据（使用随机音色）
	base64Data, err := tts_api.TextToAudioBase64WithVoice(text, voice)
	if err != nil {
		return fmt.Errorf("TTS转换失败: %w", err)
	}

	// 再次检查上下文
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	log.Printf("TTS转换成功，音频数据长度: %d", len(base64Data))

	// 解码base64数据
	audioData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return fmt.Errorf("base64解码失败: %w", err)
	}

	// 使用带上下文的播放函数
	err = PlayVoiceFromBytesWithContext(ctx, audioData)
	if err != nil {
		return fmt.Errorf("音频播放失败: %w", err)
	}

	log.Printf("TTS播放完成 (带上下文，随机音色: %s)", voiceDesc)
	return nil
}

// PlayVoiceFromBytesWithContext 支持上下文取消的字节数据播放
func PlayVoiceFromBytesWithContext(ctx context.Context, audioData []byte) error {
	// 检查上下文是否已取消
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// 使用全局音频管理器播放
	manager := GetAudioManager()

	// 只支持MP3格式
	format := "mp3"

	// 使用队列播放音频
	return manager.PlayAudioWithQueue(ctx, audioData, format)
}

// playAudioFile 播放MP3音频文件
func playAudioFile(filePath string) error {
	// 读取文件内容
	audioData, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read audio file: %w", err)
	}

	// 使用全局音频管理器播放MP3
	manager := GetAudioManager()
	ctx := context.Background()
	return manager.PlayAudioWithQueue(ctx, audioData, "mp3")
}

// PlayVoiceFromBytes 直接从字节数据播放音频（仅支持MP3格式）
func PlayVoiceFromBytes(audioData []byte) error {
	// 使用全局音频管理器播放
	manager := GetAudioManager()

	// 只支持MP3格式
	format := "mp3"

	// 使用队列播放音频
	ctx := context.Background()
	return manager.PlayAudioWithQueue(ctx, audioData, format)
}

// PlayVoiceAsync 异步播放音频（非阻塞）
func PlayVoiceAsync(base64Data string) error {
	go func() {
		err := PlayVoice(base64Data)
		if err != nil {
			// 这里可以添加日志记录
			fmt.Printf("Error playing voice: %v\n", err)
		}
	}()
	return nil
}

// PlayVoiceWithContext 支持上下文取消的音频播放
func PlayVoiceWithContext(ctx context.Context, base64Data string) error {
	// 解码base64数据
	audioData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return fmt.Errorf("failed to decode base64 data: %w", err)
	}

	// 使用全局音频管理器播放MP3音频
	audioManager := GetAudioManager()
	return audioManager.PlayAudioWithQueue(ctx, audioData, "mp3")
}
