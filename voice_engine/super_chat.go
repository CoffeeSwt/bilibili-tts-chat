package voice_engine

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/CoffeeSwt/bilibili-tts-chat/response"
	user_voice "github.com/CoffeeSwt/bilibili-tts-chat/user"
)

func PlaySuperChatVoice(msg response.SuperChatData) {
	voiceText := getSuperChatText(msg)
	// 获取用户音色
	voiceID := user_voice.GetUserVoice(msg.UName)

	go func() {
		if err := TextToVoiceAsync(voiceText, voiceID); err != nil {
			log.Printf("[TTS] 付费留言语音播放失败: %v, 内容: %s", err, voiceText)
		} else {
			log.Printf("[TTS] 付费留言语音播放成功: %s (金额:%d元)", voiceText, msg.RMB)
		}
	}()
}

// 初始化随机数种子
func init() {
	rand.Seed(time.Now().UnixNano())
}

// 统一的感谢语模板
var (
	// 统一格式的感谢语模板
	thankTemplate = "感谢%s对主播的%d元留言"

	// 祝福语开场白（可选）
	blessingPrefixes = []string{
		"祝您好运连连",
		"祝您心想事成",
		"祝您身体健康",
		"祝您工作顺利",
		"祝您天天开心",
		"祝您财源广进",
	}
)

// 获取统一的感谢语
func getThankText(username string, amount int) string {
	return fmt.Sprintf(thankTemplate, username, amount)
}

// getRandomBlessingPrefix 根据金额获取随机祝福语开场白（33%概率）
func getRandomBlessingPrefix() string {
	return blessingPrefixes[rand.Intn(len(blessingPrefixes))]
}

// 格式化留言内容
func formatMessage(message string) string {
	// 去除首尾空格
	message = strings.TrimSpace(message)

	// 如果留言过长，适当截取（保留前150个字符）
	if len([]rune(message)) > 150 {
		runes := []rune(message)
		message = string(runes[:150]) + "..."
	}

	return message
}

// GetSuperChatText 导出函数用于测试
func GetSuperChatText(msg response.SuperChatData) string {
	return getSuperChatText(msg)
}

func getSuperChatText(msg response.SuperChatData) string {
	// 获取统一的感谢语
	thankText := getThankText(msg.UName, msg.RMB)

	// 获取随机祝福语
	blessingPrefix := getRandomBlessingPrefix()

	// 处理留言内容
	formattedMessage := formatMessage(msg.Message)

	// 组合最终文本 - 新顺序：感谢 → 祝福语 → 内容
	var result string

	// 1. 先感谢
	result = thankText

	// 2. 再说祝福语
	if blessingPrefix != "" {
		result += "，" + blessingPrefix
	}

	// 3. 最后播报内容
	if formattedMessage != "" {
		// 有留言内容
		result += " 他对主播说：" + formattedMessage
	} else {
		// 没有留言内容
		result += " 但没有留言内容"
	}

	return result
}
