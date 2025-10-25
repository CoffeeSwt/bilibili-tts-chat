package super_chat

import (
	"encoding/json"
	"fmt"

	"github.com/CoffeeSwt/bilibili-tts-chat/logger"
	"github.com/CoffeeSwt/bilibili-tts-chat/response"
)

// HandleSuperChat 处理付费留言消息
// 这是付费留言消息的核心处理函数，负责解析和处理用户发送的付费留言
func HandleSuperChat(cmdData []byte) error {
	var msg response.SuperChatMessage
	if err := json.Unmarshal(cmdData, &msg); err != nil {
		logger.Error(fmt.Sprintf("[SuperChatHandler] 解析付费留言消息失败: %v", err))
		return err
	}

	logger.Info(fmt.Sprintf("[付费留言] 用户: %s, 内容: %s, 金额: %d元, 房间: %d",
		msg.Data.UName, msg.Data.Message, msg.Data.RMB, msg.Data.RoomID))

	return nil
}
