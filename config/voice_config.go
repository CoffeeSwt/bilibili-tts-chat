package config

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
)

type Voice struct {
	ID            uint   `json:"id"`              // 音色ID，自用
	Name          string `json:"name"`            // 音色名称
	VoiceType     string `json:"voice_type"`      //音色的Type号
	Gender        string `json:"gender"`          //性别
	ApiResourceID string `json:"api_resource_id"` // 火山引擎TTS，调用服务的资源信息 ID https://www.volcengine.com/docs/6561/1598757
}

// Config 配置管理结构体
type VoiceConfig struct {
	Voices []Voice `json:"voices"`
}

// 全局配置实例
var (
	_voiceConfig *VoiceConfig
	onceVoice    sync.Once
)

// loadConfig 加载配置
func loadVoiceConfig() {
	// 获取当前工作目录
	wd, _ := os.Getwd()
	configPath := filepath.Join(wd, "voices.json")

	// 打开文件
	file, err := os.Open(configPath)
	if err != nil {
		ErrorInit(fmt.Sprintf("打开 voices.json 文件失败: %v", err))
		return
	}
	defer file.Close()

	// 读取文件内容
	content, err := os.ReadFile(configPath)
	if err != nil {
		ErrorInit(fmt.Sprintf("读取 voices.json 文件失败: %v", err))
		return
	}
	var voiceConfig VoiceConfig
	err = json.Unmarshal(content, &voiceConfig)
	if err != nil {
		ErrorInit(fmt.Sprintf("解析 voices.json 文件失败: %v", err))
		return
	}
	_voiceConfig = &voiceConfig
}

func GetVoiceConfig() *VoiceConfig {
	onceVoice.Do(func() {
		loadVoiceConfig()
	})
	return _voiceConfig
}

func GetVoices() []Voice {
	return GetVoiceConfig().Voices
}

func GetVoiceByID(id uint) *Voice {
	for _, voice := range GetVoices() {
		if voice.ID == id {
			return &voice
		}
	}
	return nil
}

func GetVoiceByName(name string) *Voice {
	for _, voice := range GetVoices() {
		if voice.Name == name {
			return &voice
		}
	}
	return nil
}

func GetRandomVoice() *Voice {
	voices := GetVoices()
	randIndex := rand.Intn(len(voices))
	return &voices[randIndex]
}

// GetMaleVoices 获取所有男性音色
func GetMaleVoices() []Voice {
	var maleVoices []Voice
	for _, voice := range GetVoices() {
		if voice.Gender == "male" {
			maleVoices = append(maleVoices, voice)
		}
	}
	return maleVoices
}

// GetFemaleVoices 获取所有女性音色
func GetFemaleVoices() []Voice {
	var femaleVoices []Voice
	for _, voice := range GetVoices() {
		if voice.Gender == "female" {
			femaleVoices = append(femaleVoices, voice)
		}
	}
	return femaleVoices
}

// GetRandomMaleVoice 获取随机男性音色
func GetRandomMaleVoice() *Voice {
	maleVoices := GetMaleVoices()
	if len(maleVoices) == 0 {
		return nil
	}
	randIndex := rand.Intn(len(maleVoices))
	return &maleVoices[randIndex]
}

// GetRandomFemaleVoice 获取随机女性音色
func GetRandomFemaleVoice() *Voice {
	femaleVoices := GetFemaleVoices()
	if len(femaleVoices) == 0 {
		return nil
	}
	randIndex := rand.Intn(len(femaleVoices))
	return &femaleVoices[randIndex]
}
