package handler

import (
	"encoding/json"
	"fmt"

	"github.com/CoffeeSwt/bilibili-tts-chat/handler/dm"
	"github.com/CoffeeSwt/bilibili-tts-chat/handler/guard"
	"github.com/CoffeeSwt/bilibili-tts-chat/handler/like"
	"github.com/CoffeeSwt/bilibili-tts-chat/handler/live_end"
	"github.com/CoffeeSwt/bilibili-tts-chat/handler/live_room_enter"
	"github.com/CoffeeSwt/bilibili-tts-chat/handler/live_start"
	"github.com/CoffeeSwt/bilibili-tts-chat/handler/send_gift"
	"github.com/CoffeeSwt/bilibili-tts-chat/handler/super_chat"
	"github.com/CoffeeSwt/bilibili-tts-chat/handler/super_chat_del"
	"github.com/CoffeeSwt/bilibili-tts-chat/logger"
	"github.com/CoffeeSwt/bilibili-tts-chat/response"
)

// MessageHandler 消息处理器结构体
type MessageHandler struct {
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
		logger.Error(fmt.Sprintf("解析消息失败: %v, 原始数据: %s", err, string(cmdData)))
		return err
	}

	logger.Info(fmt.Sprintf("收到消息类型: %s", baseMsg.Cmd))

	// 根据cmd类型分发到对应的处理函数
	switch baseMsg.Cmd {
	case "LIVE_OPEN_PLATFORM_DM":
		return dm.HandleDanmaku(cmdData)
	case "LIVE_OPEN_PLATFORM_SEND_GIFT":
		return send_gift.HandleGift(cmdData)
	case "LIVE_OPEN_PLATFORM_SUPER_CHAT":
		return super_chat.HandleSuperChat(cmdData)
	case "LIVE_OPEN_PLATFORM_SUPER_CHAT_DEL":
		return super_chat_del.HandleSuperChatDel(cmdData)
	case "LIVE_OPEN_PLATFORM_GUARD":
		return guard.HandleGuard(cmdData)
	case "LIVE_OPEN_PLATFORM_LIKE":
		return like.HandleLike(cmdData)
	case "LIVE_OPEN_PLATFORM_LIVE_ROOM_ENTER":
		return live_room_enter.HandleRoomEnter(cmdData)
	case "LIVE_OPEN_PLATFORM_LIVE_START":
		return live_start.HandleLiveStart(cmdData)
	case "LIVE_OPEN_PLATFORM_LIVE_END":
		return live_end.HandleLiveEnd(cmdData)
	// case "LIVE_OPEN_PLATFORM_INTERACTION_END":
	// 	return h.HandleInteractionEnd(cmdData)
	default:
		logger.Warn(fmt.Sprintf("未知的消息类型: %s", baseMsg.Cmd))
		return fmt.Errorf("未知的消息类型: %s", baseMsg.Cmd)
	}
}
