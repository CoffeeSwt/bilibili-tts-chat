package handler

import (
	"encoding/json"
	"log"

	"github.com/CoffeeSwt/bilibili-tts-chat/response"
	"github.com/CoffeeSwt/bilibili-tts-chat/voice_engine"
)

// HandleSuperChat 处理付费留言消息
// 这是付费留言消息的核心处理函数，负责解析和处理用户发送的付费留言
func (h *MessageHandler) HandleSuperChat(cmdData []byte) error {
	var msg response.SuperChatMessage
	if err := json.Unmarshal(cmdData, &msg); err != nil {
		log.Printf("[SuperChatHandler] 解析付费留言消息失败: %v", err)
		return err
	}

	log.Printf("[付费留言] 用户: %s, 内容: %s, 金额: %d元, 房间: %d",
		msg.Data.UName, msg.Data.Message, msg.Data.RMB, msg.Data.RoomID)

	voice_engine.PlaySuperChatVoice(msg.Data)

	return nil
}

// HandleSuperChatDel 处理付费留言下线消息
// 当付费留言需要被移除时调用此函数
func (h *MessageHandler) HandleSuperChatDel(cmdData []byte) error {
	var msg response.SuperChatDelMessage
	if err := json.Unmarshal(cmdData, &msg); err != nil {
		log.Printf("[SuperChatHandler] 解析付费留言下线消息失败: %v", err)
		return err
	}

	log.Printf("[付费留言下线] 房间: %d, 删除留言ID: %v",
		msg.Data.RoomID, msg.Data.MessageIDs)

	return nil
}
