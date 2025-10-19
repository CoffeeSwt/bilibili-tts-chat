package handler

import (
	"encoding/json"
	"log"

	"github.com/CoffeeSwt/bilibili-tts-chat/response"
	"github.com/CoffeeSwt/bilibili-tts-chat/voice_engine"
)

// HandleGuard 处理大航海消息
// 这是大航海消息的核心处理函数，负责解析和处理用户购买大航海的消息
func (h *MessageHandler) HandleGuard(cmdData []byte) error {
	var msg response.GuardMessage
	if err := json.Unmarshal(cmdData, &msg); err != nil {
		log.Printf("[GuardHandler] 解析大航海消息失败: %v", err)
		return err
	}

	log.Printf("[大航海] 用户: %s, 等级: %d, 数量: %d%s, 价格: %d, 房间: %d",
		msg.Data.UserInfo.UName, msg.Data.GuardLevel, msg.Data.GuardNum, msg.Data.GuardUnit,
		msg.Data.Price, msg.Data.RoomID)

	// 播放守护语音
	voice_engine.PlayGuardVoice(msg.Data)
	return nil
}
