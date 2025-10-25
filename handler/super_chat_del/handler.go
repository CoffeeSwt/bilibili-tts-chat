package super_chat_del

import (
	"encoding/json"
	"fmt"

	"github.com/CoffeeSwt/bilibili-tts-chat/logger"
	"github.com/CoffeeSwt/bilibili-tts-chat/response"
)

// HandleSuperChat 处理付费留言消息
// 这是付费留言消息的核心处理函数，负责解析和处理用户发送的付费留言
func HandleSuperChatDel(cmdData []byte) error {
	var msg response.SuperChatDelMessage
	if err := json.Unmarshal(cmdData, &msg); err != nil {
		logger.Error(fmt.Sprintf("[SuperChatDelHandler] 解析付费留言消息失败: %v", err))
		return err
	}

	logger.Info(fmt.Sprintf("[付费留言下线] 房间: %d, 删除留言ID: %v",
		msg.Data.RoomID, msg.Data.MessageIDs))
	return nil
}
