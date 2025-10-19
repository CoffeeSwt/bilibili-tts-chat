package handler

import (
	"encoding/json"
	"log"

	"github.com/CoffeeSwt/bilibili-tts-chat/response"
	"github.com/CoffeeSwt/bilibili-tts-chat/voice_engine"
)

// HandleRoomEnter 处理用户进入房间消息
// 当有用户进入直播间时触发，可以用于欢迎消息、统计等
func (h *MessageHandler) HandleRoomEnter(cmdData []byte) error {
	var msg response.RoomEnterMessage
	if err := json.Unmarshal(cmdData, &msg); err != nil {
		log.Printf("[RoomHandler] 解析进入房间消息失败: %v", err)
		return err
	}

	log.Printf("[进入房间] 用户: %s (OpenID: %s), 房间: %d",
		msg.Data.UName, msg.Data.OpenID, msg.Data.RoomID)
	// 播放欢迎语音
	voice_engine.PlayUserEnterVoice(msg.Data)
	return nil
}

// HandleLiveStart 处理直播开始消息
// 当主播开始直播时触发，用于通知、统计、初始化等
func (h *MessageHandler) HandleLiveStart(cmdData []byte) error {
	var msg response.LiveStartMessage
	if err := json.Unmarshal(cmdData, &msg); err != nil {
		log.Printf("[RoomHandler] 解析直播开始消息失败: %v", err)
		return err
	}

	log.Printf("[直播开始] 房间: %d, 主播: %s",
		msg.Data.RoomID, msg.Data.OpenID)

	return nil
}

// HandleLiveEnd 处理直播结束消息
// 当主播结束直播时触发，用于清理、统计、通知等
func (h *MessageHandler) HandleLiveEnd(cmdData []byte) error {
	var msg response.LiveEndMessage
	if err := json.Unmarshal(cmdData, &msg); err != nil {
		log.Printf("[RoomHandler] 解析直播结束消息失败: %v", err)
		return err
	}

	log.Printf("[直播结束] 房间: %d", msg.Data.RoomID)

	return nil
}
