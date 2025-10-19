package voice_engine

import (
	"github.com/CoffeeSwt/bilibili-tts-chat/response"
	"log"
	"strconv"
	"strings"
)

func PlayGiftVoice(msg response.GiftData) {
	voiceText, voiceID := GetGiftVoiceText(msg)
	// 异步调用TTS，避免阻塞弹幕处理流程
	go func() {
		if err := TextToVoiceAsync(voiceText, voiceID); err != nil {
			log.Printf("[TTS] 弹幕语音播放失败: %v, 内容: %s", err, voiceText)
		} else {
			log.Printf("[TTS] 弹幕语音播放成功: %s", voiceText)
		}
	}()
}

func GetGiftVoiceText(msg response.GiftData) (voiceText string, voiceID string) {
	if msg.Paid {
		// 格式化为两位小数
		giftPrice := strconv.FormatFloat(float64(msg.Price)/1000, 'f', 2, 64)
		// 去掉尾部的零
		giftPrice = strings.TrimRight(giftPrice, "0")
		// 去掉小数点后面的0，如果小数点后只有0了
		giftPrice = strings.TrimRight(giftPrice, ".")
		voiceText = msg.UName + "赠送了 " + strconv.Itoa(msg.GiftNum) + "个" + msg.GiftName + "，价值" + giftPrice + "元"
	} else {
		voiceText = msg.UName + "赠送了 " + strconv.Itoa(msg.GiftNum) + "个" + msg.GiftName
	}
	return voiceText, voiceID
}
