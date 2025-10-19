package voice_engine

import (
	"fmt"
	"github.com/CoffeeSwt/bilibili-tts-chat/config"
	"github.com/CoffeeSwt/bilibili-tts-chat/response"
	user_voice "github.com/CoffeeSwt/bilibili-tts-chat/user"
	"log"
	"regexp"
	"strconv"
	"strings"
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

func GetUserContentWithVoice(username, content, replyUsername string) (string, string) {
	// 获取用户音色
	voiceID := user_voice.GetUserVoice(username)
	if voiceID == "" {
		voiceID = config.GetRandomVoiceID()
		user_voice.SetUserVoice(username, voiceID) //写入缓存
	}

	if content == "换音色" {
		randVoice := config.GetRandomVoiceID()
		user_voice.SetUserVoice(username, randVoice) //写入缓存
		return fmt.Sprintf("%s的音色已切换为现在这个音色", username), randVoice
	}
	// 使用正则表达式匹配"切换为xxx号音色"的模式
	voiceSwitchRegex := regexp.MustCompile(`换(\d+)号音色`)
	matches := voiceSwitchRegex.FindStringSubmatch(content)

	if len(matches) > 1 {
		// 提取数字
		voiceNumberStr := matches[1]
		voiceNumber, err := strconv.Atoi(voiceNumberStr)
		if err != nil {
			log.Printf("[DanmakuHandler] 音色号码转换失败: %v, 输入: %s", err, voiceNumberStr)
			return fmt.Sprintf("%s的音色切换失败，将使用原来的音色", username), ""
		}

		// 获取所有音色
		voices := config.GetVoices()
		if len(voices) == 0 {
			log.Printf("[DanmakuHandler] 音色数组为空，无法切换音色")
			return fmt.Sprintf("%s的音色切换失败，将使用原来的音色", username), voiceID
		}

		// 使用取模运算确保索引在数组范围内（1号对应索引0）
		voiceIndex := (voiceNumber - 1) % len(voices)
		if voiceIndex < 0 {
			voiceID = config.GetRandomVoiceID()
			user_voice.SetUserVoice(username, voiceID) //写入缓存
			return fmt.Sprintf("%s的音色切换失败，将使用随机音色", username), voiceID
		}

		// 设置用户音色
		user_voice.SetUserVoice(username, voices[voiceIndex].ID)

		voiceInfo := config.GetVoiceInfoByID(voices[voiceIndex].ID)
		log.Printf("[DanmakuHandler] 用户 %s 切换到 %d 号音色: %s (%s)",
			username, voiceNumber, voiceInfo.Name, voices[voiceIndex].ID)

		return fmt.Sprintf("%s已切换为%d号音色%s", username, voiceNumber, voiceInfo.Name), voices[voiceIndex].ID
	} else if replyUsername != "" {
		return username + "对" + replyUsername + "说：" + content, voiceID
	} else {
		return username + "说：" + content, voiceID
	}
}
