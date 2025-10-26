package super_chat

import (
	"encoding/json"
	"fmt"

	"github.com/CoffeeSwt/bilibili-tts-chat/logger"
	"github.com/CoffeeSwt/bilibili-tts-chat/response"
	"github.com/CoffeeSwt/bilibili-tts-chat/task_manager"
	"github.com/CoffeeSwt/bilibili-tts-chat/user"
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

	// 构建结构化的事件描述，方便AI理解和回复
	eventDescription := fmt.Sprintf("【付费留言】用户 %s 发送了 %d元 的付费留言：%s",
		msg.Data.UName, msg.Data.RMB, msg.Data.Message)

	// 将事件描述添加到任务管理器
	if err := task_manager.AddText(eventDescription, task_manager.TextTypeNormal, user.GetUserVoice(msg.Data.UName)); err != nil {
		logger.Error(fmt.Sprintf("[SuperChatHandler] 添加事件到任务管理器失败: %v", err))
	}

	return nil
}
