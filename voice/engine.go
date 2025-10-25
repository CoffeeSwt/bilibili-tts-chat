package voice

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/CoffeeSwt/bilibili-tts-chat/logger"
	"github.com/ebitengine/oto/v3"
	"github.com/hajimehoshi/go-mp3"
)

// AudioTask 表示一个音频播放任务
type AudioTask struct {
	AudioData  []byte
	Volume     int
	Done       chan error
	Completion chan struct{} // 播放完成信号，可选
}

// AudioEngine 音频播放引擎
type AudioEngine struct {
	contexts  map[string]*oto.Context // 支持多个采样率的上下文
	taskQueue chan *AudioTask
	isRunning bool
	mutex     sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
}

var (
	instance *AudioEngine
	once     sync.Once
)

// getInstance 获取音频引擎单例实例
func getInstance() *AudioEngine {
	once.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		instance = &AudioEngine{
			contexts:  make(map[string]*oto.Context),
			taskQueue: make(chan *AudioTask, 100), // 缓冲队列，最多100个任务
			isRunning: false,
			ctx:       ctx,
			cancel:    cancel,
		}
		instance.init()
	})
	return instance
}

// init 初始化音频引擎
func (e *AudioEngine) init() {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if e.isRunning {
		return
	}

	// 启动播放处理协程
	go e.processQueue()
	e.isRunning = true
}

// getOrCreateContext 获取或创建指定采样率的音频上下文
func (e *AudioEngine) getOrCreateContext(sampleRate int, channelCount int) (*oto.Context, error) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	// 创建上下文键
	contextKey := fmt.Sprintf("%d_%d", sampleRate, channelCount)

	// 检查是否已存在
	if context, exists := e.contexts[contextKey]; exists {
		return context, nil
	}

	// 创建新的音频上下文
	op := &oto.NewContextOptions{
		SampleRate:   sampleRate,
		ChannelCount: channelCount,
		Format:       oto.FormatSignedInt16LE,
		BufferSize:   time.Millisecond * 100,
	}

	context, _, err := oto.NewContext(op)
	if err != nil {
		return nil, fmt.Errorf("创建音频上下文失败 (采样率: %d, 声道: %d): %v", sampleRate, channelCount, err)
	}

	// 缓存上下文
	e.contexts[contextKey] = context
	logger.Info(fmt.Sprintf("创建新的音频上下文: 采样率=%dHz, 声道=%d", sampleRate, channelCount))

	return context, nil
}

// processQueue 处理音频播放队列
func (e *AudioEngine) processQueue() {
	for {
		select {
		case <-e.ctx.Done():
			return
		case task := <-e.taskQueue:
			if task == nil {
				continue
			}
			err := e.playAudioInternal(task.AudioData, task.Volume)
			task.Done <- err
			close(task.Done)

			// 如果有完成信号channel，发送完成信号
			if task.Completion != nil {
				close(task.Completion)
			}
		}
	}
}

// playAudioInternal 内部音频播放实现
func (e *AudioEngine) playAudioInternal(audioData []byte, volume int) error {
	if len(audioData) == 0 {
		return fmt.Errorf("audio data is empty")
	}

	// 尝试解码 MP3 格式
	reader := bytes.NewReader(audioData)
	decoder, err := mp3.NewDecoder(reader)
	if err != nil {
		// 如果不是 MP3 格式，尝试作为 PCM 数据处理
		return e.playPCMData(audioData, volume)
	}

	// 播放 MP3 数据
	return e.playMP3Data(decoder, volume)
}

// playMP3Data 播放 MP3 格式音频
func (e *AudioEngine) playMP3Data(decoder *mp3.Decoder, volume int) error {
	// 获取 MP3 的真实音频参数
	sampleRate := decoder.SampleRate()
	channelCount := 2 // MP3 通常是立体声，但我们可以根据需要调整

	logger.Info(fmt.Sprintf("MP3音频信息: 采样率=%dHz, 声道=%d", sampleRate, channelCount))

	// 获取或创建对应的音频上下文
	context, err := e.getOrCreateContext(sampleRate, channelCount)
	if err != nil {
		return fmt.Errorf("获取音频上下文失败: %v", err)
	}

	// 创建播放器
	player := context.NewPlayer(decoder)
	if player == nil {
		return fmt.Errorf("无法创建MP3播放器")
	}
	defer player.Close()

	// 设置音量 (0.0 - 1.0)
	volumeFloat := float64(volume) / 100.0
	if volumeFloat > 1.0 {
		volumeFloat = 1.0
	} else if volumeFloat < 0.0 {
		volumeFloat = 0.0
	}
	player.SetVolume(volumeFloat)

	// 开始播放
	player.Play()

	// 等待播放完成
	for player.IsPlaying() {
		time.Sleep(10 * time.Millisecond)
	}

	return nil
}

// detectPCMFormat 尝试检测PCM数据的格式
func (e *AudioEngine) detectPCMFormat(audioData []byte) (sampleRate int, channelCount int) {
	// 根据数据长度和常见的TTS音频格式进行推测
	dataLength := len(audioData)

	// 常见的TTS采样率
	commonSampleRates := []int{24000, 22050, 16000, 8000, 44100, 48000}

	// 假设是16-bit音频
	bytesPerSample := 2

	// 尝试不同的声道数和采样率组合
	for _, rate := range commonSampleRates {
		for _, channels := range []int{1, 2} {
			// 计算预期的数据长度（假设1秒音频）
			expectedBytesPerSecond := rate * channels * bytesPerSample

			// 如果数据长度接近某个时长的音频，则可能是这个格式
			if dataLength >= expectedBytesPerSecond/10 && dataLength <= expectedBytesPerSecond*10 {
				// 优先选择常见的TTS格式
				if rate == 24000 || rate == 22050 || rate == 16000 {
					return rate, channels
				}
			}
		}
	}

	// 如果无法检测，返回常见的TTS默认值
	return 24000, 1
}

// playPCMData 播放 PCM 格式音频数据
func (e *AudioEngine) playPCMData(audioData []byte, volume int) error {
	// 检查数据长度，确保至少有足够的数据
	if len(audioData) < 4 {
		return fmt.Errorf("音频数据太短，至少需要4字节")
	}

	// 尝试检测PCM格式
	sampleRate, channelCount := e.detectPCMFormat(audioData)

	logger.Info(fmt.Sprintf("PCM音频信息（检测）: 采样率=%dHz, 声道=%d, 数据长度=%d字节",
		sampleRate, channelCount, len(audioData)))

	// 获取或创建对应的音频上下文
	context, err := e.getOrCreateContext(sampleRate, channelCount)
	if err != nil {
		return fmt.Errorf("获取音频上下文失败: %v", err)
	}

	// 创建播放器
	reader := bytes.NewReader(audioData)
	player := context.NewPlayer(reader)
	if player == nil {
		return fmt.Errorf("无法创建PCM播放器")
	}
	defer player.Close()

	// 设置音量
	volumeFloat := float64(volume) / 100.0
	if volumeFloat > 1.0 {
		volumeFloat = 1.0
	} else if volumeFloat < 0.0 {
		volumeFloat = 0.0
	}
	player.SetVolume(volumeFloat)

	// 开始播放
	player.Play()

	// 等待播放完成
	for player.IsPlaying() {
		time.Sleep(10 * time.Millisecond)
	}

	return nil
}

// PlayAudio 播放音频的公共接口
// audioData: 音频数据（支持 MP3 或 PCM 格式）
// volume: 音量 (1-100)
// 返回: error
func PlayAudio(audioData []byte, volume int) error {
	// 参数验证
	if len(audioData) == 0 {
		return fmt.Errorf("audio data cannot be empty")
	}
	if volume < 1 {
		volume = 1
		logger.Warn(fmt.Sprintf("收到无效音量值 %d，设置为默认值 1", volume))
	}
	if volume > 100 {
		volume = 100
		logger.Warn(fmt.Sprintf("收到无效音量值 %d，设置为默认值 100", volume))
	}

	// 获取引擎实例
	engine := getInstance()

	// 创建任务
	task := &AudioTask{
		AudioData:  audioData,
		Volume:     volume,
		Done:       make(chan error, 1),
		Completion: nil, // 普通播放不需要完成信号
	}

	// 提交任务到队列
	select {
	case engine.taskQueue <- task:
		// 等待任务完成
		return <-task.Done
	case <-time.After(5 * time.Second):
		return fmt.Errorf("audio playback request timeout")
	}
}

// PlayAudioWithCompletion 播放音频的公共接口（带完成信号）
// audioData: 音频数据（支持 MP3 或 PCM 格式）
// volume: 音量 (1-100)
// 返回: error 和完成信号channel
func PlayAudioWithCompletion(audioData []byte, volume int) (<-chan struct{}, error) {
	// 参数验证
	if len(audioData) == 0 {
		return nil, fmt.Errorf("audio data cannot be empty")
	}
	if volume < 1 {
		volume = 1
		logger.Warn(fmt.Sprintf("收到无效音量值 %d，设置为默认值 1", volume))
	}
	if volume > 100 {
		volume = 100
		logger.Warn(fmt.Sprintf("收到无效音量值 %d，设置为默认值 100", volume))
	}

	// 获取引擎实例
	engine := getInstance()

	// 创建完成信号channel
	completion := make(chan struct{})

	// 创建任务
	task := &AudioTask{
		AudioData:  audioData,
		Volume:     volume,
		Done:       make(chan error, 1),
		Completion: completion,
	}

	// 提交任务到队列
	select {
	case engine.taskQueue <- task:
		// 等待任务开始并返回错误（如果有）和完成信号
		err := <-task.Done
		return completion, err
	case <-time.After(30 * time.Second):
		close(completion) // 超时时关闭完成信号
		return completion, fmt.Errorf("audio playback request timeout")
	}
}

// PlayAudioWithFormat 播放音频的扩展接口，允许指定音频格式
// audioData: 音频数据
// volume: 音量 (1-100)
// sampleRate: 采样率 (如24000, 44100等)
// channelCount: 声道数 (1=单声道, 2=立体声)
// 返回: error
func PlayAudioWithFormat(audioData []byte, volume int, sampleRate int, channelCount int) error {
	// 参数验证
	if len(audioData) == 0 {
		return fmt.Errorf("audio data cannot be empty")
	}
	if volume < 1 || volume > 100 {
		return fmt.Errorf("volume must be between 1 and 100")
	}
	if sampleRate <= 0 {
		return fmt.Errorf("sample rate must be positive")
	}
	if channelCount != 1 && channelCount != 2 {
		return fmt.Errorf("channel count must be 1 or 2")
	}

	// 获取引擎实例
	engine := getInstance()

	// 直接播放PCM数据，使用指定的格式
	context, err := engine.getOrCreateContext(sampleRate, channelCount)
	if err != nil {
		return fmt.Errorf("获取音频上下文失败: %v", err)
	}

	reader := bytes.NewReader(audioData)
	player := context.NewPlayer(reader)
	if player == nil {
		return fmt.Errorf("无法创建播放器")
	}
	defer player.Close()

	// 设置音量
	volumeFloat := float64(volume) / 100.0
	player.SetVolume(volumeFloat)

	// 开始播放
	player.Play()

	// 等待播放完成
	for player.IsPlaying() {
		time.Sleep(10 * time.Millisecond)
	}

	return nil
}

// Shutdown 关闭音频引擎
func Shutdown() {
	if instance != nil {
		instance.mutex.Lock()
		defer instance.mutex.Unlock()

		if instance.cancel != nil {
			instance.cancel()
		}

		// 关闭所有音频上下文
		for key, context := range instance.contexts {
			if context != nil {
				context.Suspend()
				logger.Info(fmt.Sprintf("关闭音频上下文: %s", key))
			}
		}
		instance.contexts = make(map[string]*oto.Context)
		instance.isRunning = false
	}
}

// GetQueueLength 获取当前队列长度（用于调试）
func GetQueueLength() int {
	engine := getInstance()
	return len(engine.taskQueue)
}

// GetActiveContexts 获取当前活跃的音频上下文信息（用于调试）
func GetActiveContexts() []string {
	engine := getInstance()
	engine.mutex.RLock()
	defer engine.mutex.RUnlock()

	var contexts []string
	for key := range engine.contexts {
		contexts = append(contexts, key)
	}
	return contexts
}
