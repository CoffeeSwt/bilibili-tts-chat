package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type UserConfig struct {
	RoomIDCode      string `json:"room_id_code"`      // 直播间身份码
	RoomDescription string `json:"room_description"`  // 直播间描述 //用于为主播定制上下文，例如主播叫什么，正在播什么，喜欢什么，擅长什么，有什么特点
	MaxUserDataLen  int    `json:"max_user_data_len"` // 最大用户数据长度 // 用于限制用户数据的长度，防止占用过多内存
	CleanupInterval int    `json:"cleanup_interval"`  // 清理间隔 // 用于指定清理不活跃用户的时间间隔，单位为天
	Volume          int    `json:"volume"`            // 音量 // 用于指定播放音频的音量，范围为1-100
}

// 全局配置实例
var (
	_userConfig *UserConfig
	onceUser    sync.Once
)

// loadConfig 加载配置
func loadUserConfig() {
	// 获取当前工作目录
	wd, _ := os.Getwd()
	configPath := filepath.Join(wd, "user.json")

	// 打开文件
	file, err := os.Open(configPath)
	if err != nil {
		ErrorInit(fmt.Sprintf("打开 user.json 文件失败: %v", err))
		return
	}
	defer file.Close()

	// 读取文件内容
	content, err := os.ReadFile(configPath)
	if err != nil {
		ErrorInit(fmt.Sprintf("读取 user.json 文件失败: %v", err))
		return
	}
	var userConfig UserConfig
	err = json.Unmarshal(content, &userConfig)
	if err != nil {
		ErrorInit(fmt.Sprintf("解析 user.json 文件失败: %v", err))
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
