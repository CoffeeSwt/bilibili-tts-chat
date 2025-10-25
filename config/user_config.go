package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type UserConfig struct {
	RoomIDCode      string `json:"room_id_code"`     // 直播间身份码
	RoomDescription string `json:"room_description"` // 直播间描述 //用于为主播定制上下文，例如主播叫什么，正在播什么，喜欢什么，擅长什么，有什么特点
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
