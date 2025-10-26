package user

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/CoffeeSwt/bilibili-tts-chat/config"
	"github.com/CoffeeSwt/bilibili-tts-chat/logger"
	"gopkg.in/yaml.v2"
)

var (
	userVoices = UserVoice{
		UserVoices: make(UserVoiceMap),
	}
	once       sync.Once
	voiceMutex sync.RWMutex // 读写锁，保护并发访问
	configPath string       // 配置文件的绝对路径
)

// UserVoiceInfo 用户音色信息，包含音色类型和最后活跃时间
type UserVoiceInfo struct {
	VoiceType      string    `yaml:"voice_type"`
	LastActiveTime time.Time `yaml:"last_active_time"`
}

type UserVoice struct {
	UserVoices UserVoiceMap `yaml:"user_voices"`
}

type UserVoiceMap map[string]UserVoiceInfo

// LegacyUserVoice 用于向后兼容的旧格式结构
type LegacyUserVoice struct {
	UserVoices map[string]string `yaml:"user_voices"`
}

// GetUserVoice 获取用户的音色配置，线程安全
func GetUserVoice(userName string) *config.Voice {
	loadUserVoices()

	voiceMutex.Lock()
	userInfo, exists := userVoices.UserVoices[userName]

	if !exists {
		// 为新用户分配随机音色
		v := config.GetRandomVoice()
		if v == nil {
			voiceMutex.Unlock()
			logger.Error("[GetUserVoice] 无法获取随机音色")
			return nil
		}

		// 创建新用户信息
		userInfo = UserVoiceInfo{
			VoiceType:      v.VoiceType,
			LastActiveTime: time.Now(),
		}
		userVoices.UserVoices[userName] = userInfo
		voiceMutex.Unlock()

		return v
	}

	// 更新已存在用户的活跃时间
	userInfo.LastActiveTime = time.Now()
	userVoices.UserVoices[userName] = userInfo
	voiceMutex.Unlock()

	return config.GetVoiceByType(userInfo.VoiceType)
}

// loadUserVoices 加载用户音色配置，只执行一次
func loadUserVoices() {
	once.Do(func() {
		// 获取配置文件的绝对路径
		configPath = getConfigFilePath()

		data, err := os.ReadFile(configPath)
		if err != nil {
			if os.IsNotExist(err) {
				logger.Info(fmt.Sprintf("[loadUserVoices] 用户音色配置文件不存在，将创建新文件: %s", configPath))
				// 创建空的配置文件
				createEmptyConfigFile()
				return
			}
			logger.Error(fmt.Sprintf("[loadUserVoices] 读取用户音色配置文件失败: %v", err))
			return
		}

		voiceMutex.Lock()
		defer voiceMutex.Unlock()

		// 首先尝试解析新格式
		if err := yaml.Unmarshal(data, &userVoices); err != nil {
			// 如果新格式解析失败，尝试解析旧格式
			var legacyVoices LegacyUserVoice
			if legacyErr := yaml.Unmarshal(data, &legacyVoices); legacyErr != nil {
				logger.Error(fmt.Sprintf("[loadUserVoices] 解析用户音色配置文件失败: %v", err))
				return
			}

			// 转换旧格式到新格式
			logger.Info("[loadUserVoices] 检测到旧格式配置文件，正在转换为新格式")
			userVoices.UserVoices = make(UserVoiceMap)
			now := time.Now()
			for userName, voiceType := range legacyVoices.UserVoices {
				userVoices.UserVoices[userName] = UserVoiceInfo{
					VoiceType:      voiceType,
					LastActiveTime: now, // 旧数据设置为当前时间
				}
			}

			// 立即保存转换后的新格式配置
			if err := saveUserVoicesInternal(); err != nil {
				logger.Error(fmt.Sprintf("[loadUserVoices] 保存转换后的配置文件失败: %v", err))
			} else {
				logger.Info("[loadUserVoices] 已成功保存转换后的新格式配置文件")
			}
		}

		// 确保 UserVoices map 已初始化
		if userVoices.UserVoices == nil {
			userVoices.UserVoices = make(UserVoiceMap)
		}

		logger.Info(fmt.Sprintf("[loadUserVoices] 成功加载 %d 个用户音色配置", len(userVoices.UserVoices)))

		maxUserLen := config.GetMaxUserDataLen()
		if maxUserLen <= 0 {
			maxUserLen = 1000
		}
		// 检查是否需要清理不活跃用户
		if len(userVoices.UserVoices) > maxUserLen {
			cleanupCount := cleanupInactiveUsersInternal()
			if cleanupCount > 0 {
				logger.Info(fmt.Sprintf("[loadUserVoices] 清理了 %d 个不活跃用户配置", cleanupCount))
			}
		}
	})
}

// getConfigFilePath 获取配置文件的绝对路径
func getConfigFilePath() string {
	// 获取当前工作目录
	wd, err := os.Getwd()
	if err != nil {
		logger.Error(fmt.Sprintf("[getConfigFilePath] 获取工作目录失败: %v", err))
		// 回退到相对路径
		return "user_voices.yaml"
	}
	return filepath.Join(wd, "user_voices.yaml")
}

// createEmptyConfigFile 创建空的配置文件
func createEmptyConfigFile() {
	voiceMutex.Lock()
	defer voiceMutex.Unlock()

	userVoices.UserVoices = make(UserVoiceMap)

	if err := saveUserVoicesInternal(); err != nil {
		logger.Error(fmt.Sprintf("[createEmptyConfigFile] 创建配置文件失败: %v", err))
	}
}

// saveUserVoicesInternal 内部保存函数，需要在锁保护下调用
func saveUserVoicesInternal() error {
	data, err := yaml.Marshal(&userVoices)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	// 确保目录存在
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	// 写入文件
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	return nil
}

// SaveUserVoices 手动保存用户音色配置（公开接口）
func SaveUserVoices() error {
	voiceMutex.Lock()
	defer voiceMutex.Unlock()

	return saveUserVoicesInternal()
}

// SetUserVoice 设置用户的音色配置
func SetUserVoice(userName, voiceType string) error {
	loadUserVoices()

	// 验证音色类型是否存在
	voice := config.GetVoiceByType(voiceType)
	if voice == nil {
		return fmt.Errorf("音色类型 %s 不存在", voiceType)
	}

	voiceMutex.Lock()
	userVoices.UserVoices[userName] = UserVoiceInfo{
		VoiceType:      voice.VoiceType,
		LastActiveTime: time.Now(),
	}

	// 立即保存配置，确保音色切换后立即持久化
	if err := saveUserVoicesInternal(); err != nil {
		voiceMutex.Unlock()
		logger.Error(fmt.Sprintf("[SetUserVoice] 保存用户音色配置失败: %v", err))
		return fmt.Errorf("保存配置失败: %v", err)
	}
	voiceMutex.Unlock()

	logger.Info(fmt.Sprintf("[SetUserVoice] 用户 %s 音色已切换为 %s 并保存", userName, voice.Name))
	return nil
}

// GetAllUserVoices 获取所有用户音色配置的副本
func GetAllUserVoices() UserVoiceMap {
	loadUserVoices()

	voiceMutex.RLock()
	defer voiceMutex.RUnlock()

	// 返回副本，避免外部修改
	result := make(UserVoiceMap)
	for k, v := range userVoices.UserVoices {
		result[k] = v
	}
	return result
}

// GetUserCount 获取当前用户数量
func GetUserCount() int {
	loadUserVoices()

	voiceMutex.RLock()
	defer voiceMutex.RUnlock()

	return len(userVoices.UserVoices)
}

// UpdateUserActivity 更新用户的最后活跃时间（供外部调用）
func UpdateUserActivity(userName string) {
	loadUserVoices()

	voiceMutex.Lock()
	defer voiceMutex.Unlock()

	userInfo, exists := userVoices.UserVoices[userName]
	if exists {
		userInfo.LastActiveTime = time.Now()
		userVoices.UserVoices[userName] = userInfo

		logger.Debug(fmt.Sprintf("[UpdateUserActivity] 更新用户 %s 的活跃时间", userName))
	}
}

// CleanupInactiveUsers 手动清理不活跃用户的接口
func CleanupInactiveUsers() int {
	loadUserVoices()

	voiceMutex.Lock()
	defer voiceMutex.Unlock()

	cleanupCount := cleanupInactiveUsersInternal()
	if cleanupCount > 0 {
		logger.Info(fmt.Sprintf("[CleanupInactiveUsers] 清理了 %d 个不活跃用户配置", cleanupCount))
	}

	return cleanupCount
}

// cleanupInactiveUsersInternal 内部清理函数，需要在锁保护下调用
func cleanupInactiveUsersInternal() int {
	if userVoices.UserVoices == nil {
		return 0
	}

	cleanupDays := config.GetCleanupInterval()
	if cleanupDays <= 0 {
		cleanupDays = 30
	}
	timeAgo := time.Now().AddDate(0, 0, -cleanupDays)
	cleanupCount := 0

	// 收集需要删除的用户名
	var usersToDelete []string
	for userName, userInfo := range userVoices.UserVoices {
		if userInfo.LastActiveTime.Before(timeAgo) {
			usersToDelete = append(usersToDelete, userName)
		}
	}

	// 删除不活跃用户
	for _, userName := range usersToDelete {
		delete(userVoices.UserVoices, userName)
		cleanupCount++
		logger.Debug(fmt.Sprintf("[cleanupInactiveUsersInternal] 删除不活跃用户: %s (最后活跃: %s)",
			userName, userVoices.UserVoices[userName].LastActiveTime.Format("2006-01-02 15:04:05")))
	}

	return cleanupCount
}
