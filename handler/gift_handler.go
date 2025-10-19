package handler

import (
	"encoding/json"
	"log"

	"github.com/CoffeeSwt/bilibili-tts-chat/response"
	"github.com/CoffeeSwt/bilibili-tts-chat/voice_engine"
)

// HandleGift 处理礼物消息
// 这是礼物消息的核心处理函数，负责解析和处理用户发送的礼物
func (h *MessageHandler) HandleGift(cmdData []byte) error {
	var msg response.GiftMessage
	if err := json.Unmarshal(cmdData, &msg); err != nil {
		log.Printf("[GiftHandler] 解析礼物消息失败: %v", err)
		return err
	}

	log.Printf("[礼物] 用户: %s, 礼物: %s x%d, 价值: %d, 房间: %d",
		msg.Data.UName, msg.Data.GiftName, msg.Data.GiftNum, msg.Data.Price, msg.Data.RoomID)

	// 播放礼物语音
	voice_engine.PlayGiftVoice(msg.Data)

	return nil
}
