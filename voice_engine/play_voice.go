package voice_engine

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/CoffeeSwt/bilibili-tts-chat/config"
	"github.com/CoffeeSwt/bilibili-tts-chat/response"
	user_voice "github.com/CoffeeSwt/bilibili-tts-chat/user"
)

// playDanmakuVoice 异步播放弹幕语音
func PlayDanmakuVoice(dataDanmu response.DanmakuData) {
	// 过滤空内容
	content := strings.TrimSpace(dataDanmu.Msg)
	if content == "" {
		return
	}

	// 限制弹幕长度，避免过长的内容影响体验
	const maxLength = 100
	if len([]rune(content)) > maxLength {
		log.Printf("[TTS] 弹幕内容过长，跳过语音播放: %s (长度: %d)", content, len([]rune(content)))
		return
	}

	voiceText, voiceID := GetUserContentWithVoice(dataDanmu.UName, content, dataDanmu.ReplyUName)

	// 异步调用TTS，避免阻塞弹幕处理流程
	go func() {
		if err := TextToVoiceAsync(voiceText, voiceID); err != nil {
			log.Printf("[TTS] 弹幕语音播放失败: %v, 内容: %s", err, voiceText)
		} else {
			log.Printf("[TTS] 弹幕语音播放成功: %s", voiceText)
		}
	}()
}

// switchVoiceAndNotify 公共的音色切换和通知逻辑
func switchVoiceAndNotify(username, newVoiceID string, voiceInfo config.VoiceInfo, voiceIndex int, switchType string) (string, string) {
	user_voice.SetUserVoice(username, newVoiceID)
	log.Printf("[DanmakuHandler] 用户 %s %s切换音色: %s (%s)",
		username, switchType, voiceInfo.Name, newVoiceID)
	return fmt.Sprintf("%s已切换为%d号音色%s", username, voiceIndex, voiceInfo.Name), newVoiceID
}

func GetUserContentWithVoice(username, content, replyUsername string) (string, string) {
	// 获取用户音色
	voiceID := user_voice.GetUserVoice(username)
	if voiceID == "" {
		voiceID = config.GetRandomVoiceID()
		user_voice.SetUserVoice(username, voiceID) //写入缓存
	}

	// 1. 随机切换音色："换音色"
	if content == "换音色" {
		newVoiceID := config.GetRandomVoiceID()
		voiceInfo := config.GetVoiceInfoByID(newVoiceID)
		voiceIndex := config.GetVoiceIndexByID(newVoiceID)
		return switchVoiceAndNotify(username, newVoiceID, voiceInfo, voiceIndex, "随机")
	}

	// 2. 按编号切换音色："换xxx号音色"
	voiceNumberRegex := regexp.MustCompile(`换(\d+)号音色`)
	numberMatches := voiceNumberRegex.FindStringSubmatch(content)
	if len(numberMatches) > 1 {
		voiceNumberStr := numberMatches[1]
		voiceNumber, err := strconv.Atoi(voiceNumberStr)
		if err != nil {
			log.Printf("[DanmakuHandler] 音色号码转换失败: %v, 输入: %s", err, voiceNumberStr)
			return fmt.Sprintf("%s的音色切换失败，将使用原来的音色", username), ""
		}

		// 使用GetVoiceByIndex函数根据编号获取音色
		voiceInfo := config.GetVoiceByIndex(voiceNumber)
		if voiceInfo.ID == "" {
			// 如果没有找到对应编号的音色，使用随机音色
			randomVoiceID := config.GetRandomVoiceID()
			randomVoiceInfo := config.GetVoiceInfoByID(randomVoiceID)
			randomVoiceIndex := config.GetVoiceIndexByID(randomVoiceID)
			log.Printf("[DanmakuHandler] 用户 %s 请求的 %d 号音色不存在，使用随机音色: %s (%s)",
				username, voiceNumber, randomVoiceInfo.Name, randomVoiceID)
			return switchVoiceAndNotify(username, randomVoiceID, randomVoiceInfo, randomVoiceIndex,
				fmt.Sprintf("请求的%d号音色不存在，随机", voiceNumber))
		}

		return switchVoiceAndNotify(username, voiceInfo.ID, voiceInfo, voiceNumber, "按编号")
	}

	// 3. 按名称切换音色："换xxx音色"
	voiceNameRegex := regexp.MustCompile(`换(.+?)音色`)
	nameMatches := voiceNameRegex.FindStringSubmatch(content)
	if len(nameMatches) > 1 {
		voiceName := strings.TrimSpace(nameMatches[1])
		// 过滤掉数字，避免与按编号切换冲突
		if !regexp.MustCompile(`^\d+$`).MatchString(voiceName) {
			voiceInfo := config.GetVoiceInfoByName(voiceName)
			// 检查是否找到了匹配的音色（通过检查ID是否为默认值）
			if voiceInfo.Name == voiceName {
				voiceIndex := config.GetVoiceIndexByID(voiceInfo.ID)
				log.Printf("[DanmakuHandler] 用户 %s 按名称切换音色: %s -> %s (%s)",
					username, voiceName, voiceInfo.Name, voiceInfo.ID)
				return switchVoiceAndNotify(username, voiceInfo.ID, voiceInfo, voiceIndex, "按名称")
			} else {
				// 没有找到匹配的音色名称
				log.Printf("[DanmakuHandler] 用户 %s 请求的音色名称 '%s' 不存在", username, voiceName)
				return fmt.Sprintf("%s请求的音色'%s'不存在，请检查音色名称", username, voiceName), ""
			}
		}
	}

	// 4. 普通聊天内容
	if replyUsername != "" {
		return username + "对" + replyUsername + "说：" + content, voiceID
	} else {
		return username + "说：" + content, voiceID
	}
}
