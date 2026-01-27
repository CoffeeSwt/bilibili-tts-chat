package config

import (
	"bufio"
	"embed"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

type Mode string

const (
	Release Mode = "release"
	Dev     Mode = "dev"
)

//go:embed .env
var embeddedEnv embed.FS

// EnvConfig 配置管理结构体
type EnvConfig struct {
	Mode                Mode   `json:"mode"`                   //运行模式，可选值：dev, release
	TTS_XApiAppID       string `json:"tts_x_api_app_id"`       //火山引擎TTS服务的App ID
	TTS_XApiAccessKey   string `json:"tts_x_api_access_key"`   //火山引擎TTS服务的Access Key
	BiliAppID           string `json:"bili_app_id"`            //B站开放平台App ID
	BiliAccessKey       string `json:"bili_access_key"`        //B站开放平台Access Key
	BiliSecretKey       string `json:"bili_secret_key"`        //B站开放平台Access Key Secret
	LLMMockEnabled      bool   `json:"llm_mock_enabled"`       //是否启用LLM Mock模式，用于测试
	LLMVolcengineAPIKey string `json:"llm_volcengine_api_key"` //火山引擎LLM服务的API Key
	LLMVolcengineModel  string `json:"llm_volcengine_model"`   //火山引擎LLM服务的模型名称
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
	envMap := make(map[string]string)

	// 尝试读取嵌入的 .env 文件
	content, err := embeddedEnv.ReadFile(".env")
	if err == nil {
		fmt.Println("🔒 使用内置配置启动")
		parseEnvContent(string(content), envMap)
	} else {
		// 读取本地 .env 文件，支持向上查找以及回退到 .env.example
		wd, _ := os.Getwd()
		envPath, ok := findFileUpwards(wd, ".env")
		if !ok {
			if examplePath, ok2 := findFileUpwards(wd, ".env.example"); ok2 {
				envPath = examplePath
			} else {
				fmt.Println("⚠️ 未找到 .env 或 .env.example，将使用默认配置")
				// 使用默认空配置
				envConfig = &EnvConfig{Mode: Dev}
				return
			}
		}

		content, err := os.ReadFile(envPath)
		if err == nil {
			parseEnvContent(string(content), envMap)
		}
	}

	// 创建配置实例
	envConfig = &EnvConfig{
		Mode:                getWithDefault(envMap, "mode", Dev),
		TTS_XApiAppID:       getWithDefault(envMap, "tts_x_api_app_id", ""),
		TTS_XApiAccessKey:   getWithDefault(envMap, "tts_x_api_access_key", ""),
		BiliAppID:           getWithDefault(envMap, "bili_app_id", ""),
		BiliAccessKey:       getWithDefault(envMap, "bili_access_key", ""),
		BiliSecretKey:       getWithDefault(envMap, "bili_secret_key", ""),
		LLMMockEnabled:      getWithDefault(envMap, "llm_mock_enabled", false),
		LLMVolcengineAPIKey: getWithDefault(envMap, "llm_volcengine_api_key", ""),
		LLMVolcengineModel:  getWithDefault(envMap, "llm_volcengine_model", ""),
	}
}

func parseEnvContent(content string, envMap map[string]string) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if len(value) >= 2 {
				if (value[0] == '"' && value[len(value)-1] == '"') ||
					(value[0] == '\'' && value[len(value)-1] == '\'') {
					value = value[1 : len(value)-1]
				}
			}
			envMap[key] = value
		}
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

// GetLLMMockEnabled 获取LLM Mock模式配置
func GetLLMMockEnabled() bool {
	return GetEnvConfig().LLMMockEnabled
}

// GetLLMVolcengineAPIKey 获取火山引擎LLM服务的API Key
func GetLLMVolcengineAPIKey() string {
	return GetEnvConfig().LLMVolcengineAPIKey
}

// GetLLMVolcengineModel 获取火山引擎LLM服务的模型名称
func GetLLMVolcengineModel() string {
	return GetEnvConfig().LLMVolcengineModel
}
