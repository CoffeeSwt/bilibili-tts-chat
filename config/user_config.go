package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type UserConfig struct {
	RoomIDCode          string `json:"room_id_code"`          // 直播间身份码
	RoomDescription     string `json:"room_description"`      // 直播间描述 //用于为主播定制上下文，例如主播叫什么，正在播什么，喜欢什么，擅长什么，有什么特点
	AssistantName       string `json:"assistant_name"`        // 助手名称 // 用于指定助手的名称，例如"助手"、"小助手"等
	MaxUserDataLen      int    `json:"max_user_data_len"`     // 最大用户数据长度 // 用于限制用户数据的长度，防止占用过多内存
	CleanupInterval     int    `json:"cleanup_interval"`      // 清理间隔 // 用于指定清理不活跃用户的时间间隔，单位为天
	Volume              int    `json:"volume"`                // 音量 // 用于指定播放音频的音量，范围为1-100
	SpeechRate          int    `json:"speech_rate"`           // 广播语速 // 用于指定广播消息的语速，范围为[-50,100]，100代表2.0倍速，-50代表0.5倍数
	AssistantMemorySize int    `json:"assistant_memory_size"` // 助手的记忆大小
	UseLLMReplay        bool   `json:"use_llm_replay"`        // 是否使用LLM回复 // 用于指定是否使用LLM模型回复用户消息，为true时表示使用，为false时表示不使用
	FirstStart          bool   `json:"first_start"`           // 是否第一次启动 // 用于指定是否第一次启动程序，为true时表示第一次启动，为false时表示不是第一次启动，第一次启动用于初始化配置
}

// 全局配置实例
var (
	_userConfig *UserConfig
	onceUser    sync.Once
)

// loadConfig 加载配置
func loadUserConfig() {
	wd, _ := os.Getwd()
	configPath, ok := findFileUpwards(wd, "user.json")
	if !ok {
		// 回退到示例配置
		if examplePath, ok2 := findFileUpwards(wd, "user.example.json"); ok2 {
			configPath = examplePath
			fmt.Println("⚠️ 未找到 user.json，已回退到 user.example.json")
		} else {
			ErrorInit(fmt.Sprintf("未找到 user.json 或 user.example.json"))
			return
		}
	}

	// 读取文件内容
	content, err := os.ReadFile(configPath)
	if err != nil {
		ErrorInit(fmt.Sprintf("读取用户配置文件失败: %v", err))
		return
	}
	var userConfig UserConfig
	err = json.Unmarshal(content, &userConfig)
	if err != nil {
		ErrorInit(fmt.Sprintf("解析用户配置文件失败: %v", err))
		return
	}
	_userConfig = &userConfig
}

func GetUserConfig() *UserConfig {
	onceUser.Do(func() {
		loadUserConfig()
	})
	return _userConfig
}

func GetRoomIDCode() string {
	return GetUserConfig().RoomIDCode
}

func GetRoomDescription() string {
	return GetUserConfig().RoomDescription
}

func GetMaxUserDataLen() int {
	return GetUserConfig().MaxUserDataLen
}

func GetCleanupInterval() int {
	return GetUserConfig().CleanupInterval
}

func GetVolume() int {
	volume := GetUserConfig().Volume
	if volume < 1 {
		volume = 1
	}
	if volume > 100 {
		volume = 100
	}
	return volume
}

func GetAssistantName() string {
	return GetUserConfig().AssistantName
}

func GetAssistantMemorySize() int {
	return GetUserConfig().AssistantMemorySize
}

func GetSpeechRate() int {
	rate := min(max(GetUserConfig().SpeechRate, -50), 100)
	return rate
}

func GetUseLLMReplay() bool {
	return GetUserConfig().UseLLMReplay
}

func IfFirstStart() bool {
	return GetUserConfig().FirstStart
}

// SaveUserConfig 保存用户配置
func SaveUserConfig(cfg UserConfig) error {
	_userConfig = &cfg

	data, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	wd, _ := os.Getwd()
	configPath := filepath.Join(wd, "user.json")

	// 尝试查找现有的 user.json
	if existingPath, ok := findFileUpwards(wd, "user.json"); ok {
		configPath = existingPath
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置失败: %v", err)
	}
	return nil
}
