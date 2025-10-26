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

	// HashMap 索引，用于快速查找
	idMap        map[uint]*Voice   // ID -> Voice 映射
	nameMap      map[string]*Voice // Name -> Voice 映射
	typeMap      map[string]*Voice // VoiceType -> Voice 映射
	maleVoices   []Voice           // 男性音色列表
	femaleVoices []Voice           // 女性音色列表
}

// 全局配置实例
var (
	_voiceConfig *VoiceConfig
	onceVoice    sync.Once
	voiceMutex   sync.RWMutex // 读写锁，确保线程安全
)

// loadConfig 加载配置并构建索引
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

	// 构建 HashMap 索引
	buildVoiceIndexes(&voiceConfig)

	_voiceConfig = &voiceConfig
}

// buildVoiceIndexes 构建各种索引以提高查询效率
func buildVoiceIndexes(config *VoiceConfig) {
	// 初始化 map
	config.idMap = make(map[uint]*Voice)
	config.nameMap = make(map[string]*Voice)
	config.typeMap = make(map[string]*Voice)
	config.maleVoices = make([]Voice, 0)
	config.femaleVoices = make([]Voice, 0)

	// 遍历所有音色，构建索引
	for i := range config.Voices {
		voice := &config.Voices[i]

		// 构建 ID 映射
		config.idMap[voice.ID] = voice

		// 构建 Name 映射
		config.nameMap[voice.Name] = voice

		// 构建 VoiceType 映射
		config.typeMap[voice.VoiceType] = voice

		// 按性别分组
		switch voice.Gender {
		case "male":
			config.maleVoices = append(config.maleVoices, *voice)
		case "female":
			config.femaleVoices = append(config.femaleVoices, *voice)
		}
	}
}

func GetVoiceConfig() *VoiceConfig {
	onceVoice.Do(func() {
		loadVoiceConfig()
	})
	return _voiceConfig
}

func GetVoices() []Voice {
	voiceMutex.RLock()
	defer voiceMutex.RUnlock()
	return GetVoiceConfig().Voices
}

// GetVoiceByID 通过 ID 快速查找音色 - O(1) 时间复杂度
func GetVoiceByID(id uint) *Voice {
	voiceMutex.RLock()
	defer voiceMutex.RUnlock()

	config := GetVoiceConfig()
	if config == nil || config.idMap == nil {
		return GetRandomVoice()
	}

	voice, exists := config.idMap[id]
	if !exists {
		return nil
	}
	return voice
}

// GetVoiceByName 通过名称快速查找音色 - O(1) 时间复杂度
func GetVoiceByName(name string) *Voice {
	voiceMutex.RLock()
	defer voiceMutex.RUnlock()

	config := GetVoiceConfig()
	if config == nil || config.nameMap == nil {
		return GetRandomVoice()
	}

	voice, exists := config.nameMap[name]
	if !exists {
		return GetRandomVoice()
	}
	return voice
}

// GetVoiceByType 通过类型快速查找音色 - O(1) 时间复杂度
func GetVoiceByType(voiceType string) *Voice {
	voiceMutex.RLock()
	defer voiceMutex.RUnlock()

	config := GetVoiceConfig()
	if config == nil || config.typeMap == nil {
		return GetRandomVoice()
	}

	voice, exists := config.typeMap[voiceType]
	if !exists {
		return GetRandomVoice()
	}
	return voice
}

// GetRandomVoice 获取随机音色
func GetRandomVoice() *Voice {
	voiceMutex.RLock()
	defer voiceMutex.RUnlock()

	config := GetVoiceConfig()
	if config == nil || len(config.Voices) == 0 {
		return nil
	}

	randIndex := rand.Intn(len(config.Voices))
	return &config.Voices[randIndex]
}

// GetMaleVoices 获取所有男性音色 - O(1) 时间复杂度
func GetMaleVoices() []Voice {
	voiceMutex.RLock()
	defer voiceMutex.RUnlock()

	config := GetVoiceConfig()
	if config == nil {
		return []Voice{}
	}

	// 返回预先构建的男性音色列表的副本
	result := make([]Voice, len(config.maleVoices))
	copy(result, config.maleVoices)
	return result
}

// GetFemaleVoices 获取所有女性音色 - O(1) 时间复杂度
func GetFemaleVoices() []Voice {
	voiceMutex.RLock()
	defer voiceMutex.RUnlock()

	config := GetVoiceConfig()
	if config == nil {
		return []Voice{}
	}

	// 返回预先构建的女性音色列表的副本
	result := make([]Voice, len(config.femaleVoices))
	copy(result, config.femaleVoices)
	return result
}

// GetRandomMaleVoice 获取随机男性音色 - O(1) 时间复杂度
func GetRandomMaleVoice() *Voice {
	voiceMutex.RLock()
	defer voiceMutex.RUnlock()

	config := GetVoiceConfig()
	if config == nil || len(config.maleVoices) == 0 {
		return nil
	}

	randIndex := rand.Intn(len(config.maleVoices))
	return &config.maleVoices[randIndex]
}

// GetRandomFemaleVoice 获取随机女性音色 - O(1) 时间复杂度
func GetRandomFemaleVoice() *Voice {
	voiceMutex.RLock()
	defer voiceMutex.RUnlock()

	config := GetVoiceConfig()
	if config == nil || len(config.femaleVoices) == 0 {
		return nil
	}

	randIndex := rand.Intn(len(config.femaleVoices))
	return &config.femaleVoices[randIndex]
}
