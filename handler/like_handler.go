package handler

import (
	"encoding/json"
	"log"

	"github.com/CoffeeSwt/bilibili-tts-chat/response"
	"github.com/CoffeeSwt/bilibili-tts-chat/voice_engine"
)

// HandleLike 处理点赞消息
// 这是点赞消息的核心处理函数，负责解析和处理用户的点赞行为
func (h *MessageHandler) HandleLike(cmdData []byte) error {
	var msg response.LikeMessage
	if err := json.Unmarshal(cmdData, &msg); err != nil {
		log.Printf("[LikeHandler] 解析点赞消息失败: %v", err)
		return err
	}

	log.Printf("[点赞] 用户: %s, 点赞数: %d, 房间: %d",
		msg.Data.UName, msg.Data.LikeCount, msg.Data.RoomID)

	// 播放点赞语音
	voice_engine.PlayLikeVoice(msg.Data)

	return nil
}
