package user_voice

import (
	"github.com/CoffeeSwt/bilibili-tts-chat/config"
	"io/ioutil"
	"log"
	"os"
	"sync"

	"gopkg.in/yaml.v2"
)

var (
	UserVoiceMap = make(map[string]string)
	DefaultVoice = "zh_female_kefunvsheng_mars_bigtts"
	mutex        sync.RWMutex
)

// UserVoiceConfig 用于YAML序列化的结构体
type UserVoiceConfig struct {
	UserVoices map[string]string `yaml:"user_voices"`
}

const UserVoiceFile = "user_voices.yaml"

func SetUserVoice(userID, voiceID string) {
	UserVoiceMap[userID] = voiceID
}

func GetUserVoice(userID string) string {
	if voiceID, ok := UserVoiceMap[userID]; ok {
		return voiceID
	}
	randV := config.GetRandomVoiceID()
	SetUserVoice(userID, randV)
	return randV
}

// SaveUserVoices 将用户音色映射保存到YAML文件
func SaveUserVoices() error {
	mutex.RLock()
	defer mutex.RUnlock()

	config := UserVoiceConfig{
		UserVoices: make(map[string]string),
	}

	// 复制当前的用户音色映射
	for userID, voiceID := range UserVoiceMap {
		config.UserVoices[userID] = voiceID
	}

	// 序列化为YAML
	data, err := yaml.Marshal(&config)
	if err != nil {
		log.Printf("Failed to marshal user voices to YAML: %v", err)
		return err
	}

	// 写入文件
	err = ioutil.WriteFile(UserVoiceFile, data, 0644)
	if err != nil {
		log.Printf("Failed to write user voices to file %s: %v", UserVoiceFile, err)
		return err
	}

	log.Printf("Successfully saved %d user voice mappings to %s", len(config.UserVoices), UserVoiceFile)
	return nil
}

// LoadUserVoices 从YAML文件加载用户音色映射
func LoadUserVoices() error {
	mutex.Lock()
	defer mutex.Unlock()

	// 检查文件是否存在
	if _, err := os.Stat(UserVoiceFile); os.IsNotExist(err) {
		log.Printf("User voice file %s does not exist, starting with empty mapping", UserVoiceFile)
		return nil
	}

	// 读取文件
	data, err := ioutil.ReadFile(UserVoiceFile)
	if err != nil {
		log.Printf("Failed to read user voices from file %s: %v", UserVoiceFile, err)
		return err
	}

	// 反序列化YAML
	var config UserVoiceConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Printf("Failed to unmarshal user voices from YAML: %v", err)
		return err
	}

	// 加载到内存映射
	if config.UserVoices != nil {
		for userID, voiceID := range config.UserVoices {
			UserVoiceMap[userID] = voiceID
		}
		log.Printf("Successfully loaded %d user voice mappings from %s", len(config.UserVoices), UserVoiceFile)
	} else {
		log.Printf("No user voice mappings found in %s", UserVoiceFile)
	}

	return nil
}
