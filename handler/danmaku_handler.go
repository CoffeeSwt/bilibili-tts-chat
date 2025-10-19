package handler

import (
	"encoding/json"
	"log"

	"github.com/CoffeeSwt/bilibili-tts-chat/response"
	"github.com/CoffeeSwt/bilibili-tts-chat/voice_engine"
)

// HandleDanmaku 处理弹幕消息
// 这是弹幕消息的核心处理函数，负责解析和处理用户发送的弹幕
func (h *MessageHandler) HandleDanmaku(cmdData []byte) error {
	var msg response.DanmakuMessage
	if err := json.Unmarshal(cmdData, &msg); err != nil {
		log.Printf("[DanmakuHandler] 解析弹幕消息失败: %v", err)
		return err
	}

	log.Printf("[弹幕] 用户: %s, 内容: %s, 房间: %d",
		msg.Data.UName, msg.Data.Msg, msg.Data.RoomID)

	// TTS语音播放弹幕内容
	voice_engine.PlayDanmakuVoice(msg.Data)

	return nil
}
