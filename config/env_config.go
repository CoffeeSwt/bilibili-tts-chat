package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

type Mode string

const (
	Release Mode = "release"
	Dev     Mode = "dev"
)

// EnvConfig 配置管理结构体
type EnvConfig struct {
	Mode              Mode   `json:"mode"`                 //运行模式，可选值：dev, release
	TTS_XApiAppID     string `json:"tts_x_api_app_id"`     //火山引擎TTS服务的App ID
	TTS_XApiAccessKey string `json:"tts_x_api_access_key"` //火山引擎TTS服务的Access Key
	BiliAppID         string `json:"bili_app_id"`          //B站开放平台App ID
	BiliAccessKey     string `json:"bili_access_key"`      //B站开放平台Access Key
	BiliSecretKey     string `json:"bili_secret_key"`      //B站开放平台Access Key Secret
}

// 全局配置实例
var (
	envConfig *EnvConfig
	once      sync.Once
)

// getWithDefault 使用泛型的配置获取函数（内部使用）
func getWithDefault[T any](envMap map[string]string, key string, defaultValue T) T {
	value, exists := envMap[key]
	if !exists {
		return defaultValue
	}

	// 根据默认值的类型进行转换
	switch any(defaultValue).(type) {
	case string:
		return any(value).(T)
	case int:
		if intVal, err := strconv.Atoi(value); err == nil {
			return any(intVal).(T)
		}
		return defaultValue
	case bool:
		lowerValue := strings.ToLower(value)
		boolVal := lowerValue == "true" || lowerValue == "1" || lowerValue == "yes" || lowerValue == "on"
		return any(boolVal).(T)
	case float64:
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return any(floatVal).(T)
		}
		return defaultValue
	default:
		return defaultValue
	}
}

// loadConfig 加载配置
func loadEnvConfig() {
	// 读取 .env 文件

	envMap := make(map[string]string)
	// 获取当前工作目录
	wd, _ := os.Getwd()
	// 构建 .env 文件路径
	envPath := filepath.Join(wd, ".env")

	// 打开文件
	file, err := os.Open(envPath)
	if err != nil {
		ErrorInit(fmt.Sprintf("打开 .env 文件失败: %v", err))
	}
	defer file.Close()

	// 逐行读取文件
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// 跳过空行和注释行
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// 解析键值对
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			// 移除值两端的引号（如果有）
			if len(value) >= 2 {
				if (value[0] == '"' && value[len(value)-1] == '"') ||
					(value[0] == '\'' && value[len(value)-1] == '\'') {
					value = value[1 : len(value)-1]
				}
			}

			envMap[key] = value
		}
	}

	// 创建配置实例
	envConfig = &EnvConfig{
		Mode:              getWithDefault(envMap, "mode", Dev),
		TTS_XApiAppID:     getWithDefault(envMap, "tts_x_api_app_id", ""),
		TTS_XApiAccessKey: getWithDefault(envMap, "tts_x_api_access_key", ""),
		BiliAppID:         getWithDefault(envMap, "bili_app_id", ""),
		BiliAccessKey:     getWithDefault(envMap, "bili_access_key", ""),
		BiliSecretKey:     getWithDefault(envMap, "bili_secret_key", ""),
	}
}

func GetEnvConfig() *EnvConfig {
	once.Do(func() {
		loadEnvConfig()
	})
	return envConfig
}

func GetMode() Mode {
	return GetEnvConfig().Mode
}

func IsDev() bool {
	return GetMode() == Dev
}

func GetTTSXApiAppID() string {
	return GetEnvConfig().TTS_XApiAppID
}

func GetTTSXApiAccessKey() string {
	return GetEnvConfig().TTS_XApiAccessKey
}

// BiliAppID B站开放平台App ID
func GetBiliAppID() int {
	if appID, err := strconv.Atoi(GetEnvConfig().BiliAppID); err == nil {
		return appID
	}
	return 0
}

func GetBiliAccessKey() string {
	return GetEnvConfig().BiliAccessKey
}

func GetBiliSecretKey() string {
	return GetEnvConfig().BiliSecretKey
}
