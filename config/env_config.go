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
	embeddedMap := make(map[string]string)
	localMap := make(map[string]string)

	// 1. 读取嵌入的 .env (构建时注入，优先级最高)
	if content, err := embeddedEnv.ReadFile(".env"); err == nil {
		fmt.Println("🔒 已加载内置配置")
		parseEnvContent(string(content), embeddedMap)
	}

	// 2. 读取本地 .env (用于开发或用户覆盖非敏感配置)
	wd, _ := os.Getwd()
	if envPath, ok := findFileUpwards(wd, ".env"); ok {
		if content, err := os.ReadFile(envPath); err == nil {
			parseEnvContent(string(content), localMap)
		}
	} else if examplePath, ok := findFileUpwards(wd, ".env.example"); ok {
		// 回退到 example
		if content, err := os.ReadFile(examplePath); err == nil {
			parseEnvContent(string(content), localMap)
		}
	}

	// 辅助函数：优先从 embeddedMap 取，取不到再从 localMap 取
	// 对于 B站凭证，我们强制优先使用 embeddedMap 中的值（如果有），防止用户通过本地 .env 覆盖
	// 如果 embeddedMap 中没有（例如开发环境且 config/.env 是空的），则回退到 localMap
	getVal := func(key string, sensitive bool) string {
		if val, ok := embeddedMap[key]; ok && val != "" {
			return val
		}

		// 敏感配置（B站凭证）不允许通过本地 .env 覆盖
		// 必须通过构建脚本注入到 embeddedMap 中，或者在开发时手动放置到 config/.env
		if sensitive {
			return ""
		}

		if val, ok := localMap[key]; ok {
			return val
		}
		return ""
	}

	// 创建配置实例
	envConfig = &EnvConfig{
		Mode:              Mode(getVal("mode", false)),
		TTS_XApiAppID:     getVal("tts_x_api_app_id", false),
		TTS_XApiAccessKey: getVal("tts_x_api_access_key", false),

		// B站凭证
		BiliAppID:     getVal("bili_app_id", true),
		BiliAccessKey: getVal("bili_access_key", true),
		BiliSecretKey: getVal("bili_secret_key", true),

		LLMMockEnabled:      getVal("llm_mock_enabled", false) == "true",
		LLMVolcengineAPIKey: getVal("llm_volcengine_api_key", false),
		LLMVolcengineModel:  getVal("llm_volcengine_model", false),
	}

	// 如果 mode 未设置，默认为 dev
	if envConfig.Mode == "" {
		envConfig.Mode = Dev
	}

	if envConfig.BiliAppID == "" {
		fmt.Println("⚠️ 未检测到B站官方授权凭证")
		fmt.Println("👉 请联系作者 CoffeeSwt 获取授权，或使用官方构建版本")
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
