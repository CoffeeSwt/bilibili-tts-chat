package handler

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/CoffeeSwt/bilibili-tts-chat/response"
)

// MessageHandler 消息处理器结构体
type MessageHandler struct {
	// 可以添加一些配置或状态字段
}

// NewMessageHandler 创建新的消息处理器
func NewMessageHandler() *MessageHandler {
	return &MessageHandler{}
}

// HandleMessage 统一的消息分发函数
func (h *MessageHandler) HandleMessage(cmdData []byte) error {
	// 首先解析出基本的消息结构，获取cmd字段
	var baseMsg response.LiveMessage
	if err := json.Unmarshal(cmdData, &baseMsg); err != nil {
		log.Printf("[MessageHandler] 解析消息失败: %v, 原始数据: %s", err, string(cmdData))
		return err
	}

	log.Printf("[MessageHandler] 收到消息类型: %s", baseMsg.Cmd)

	// 根据cmd类型分发到对应的处理函数
	switch baseMsg.Cmd {
	case "LIVE_OPEN_PLATFORM_DM":
		return h.HandleDanmaku(cmdData)
	case "LIVE_OPEN_PLATFORM_SEND_GIFT":
		return h.HandleGift(cmdData)
	case "LIVE_OPEN_PLATFORM_SUPER_CHAT":
		return h.HandleSuperChat(cmdData)
	case "LIVE_OPEN_PLATFORM_SUPER_CHAT_DEL":
		return h.HandleSuperChatDel(cmdData)
	case "LIVE_OPEN_PLATFORM_GUARD":
		return h.HandleGuard(cmdData)
	case "LIVE_OPEN_PLATFORM_LIKE":
		return h.HandleLike(cmdData)
	case "LIVE_OPEN_PLATFORM_LIVE_ROOM_ENTER":
		return h.HandleRoomEnter(cmdData)
	case "LIVE_OPEN_PLATFORM_LIVE_START":
		return h.HandleLiveStart(cmdData)
	case "LIVE_OPEN_PLATFORM_LIVE_END":
		return h.HandleLiveEnd(cmdData)
	// case "LIVE_OPEN_PLATFORM_INTERACTION_END":
	// 	return h.HandleInteractionEnd(cmdData)
	default:
		log.Printf("[MessageHandler] 未知的消息类型: %s", baseMsg.Cmd)
		return fmt.Errorf("未知的消息类型: %s", baseMsg.Cmd)
	}
}
