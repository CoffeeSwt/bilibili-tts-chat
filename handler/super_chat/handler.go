package super_chat

import (
	"encoding/json"
	"fmt"

	"github.com/CoffeeSwt/bilibili-tts-chat/config"
	"github.com/CoffeeSwt/bilibili-tts-chat/handler/common"
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

	usingLLMReply := config.GetUseLLMReplay()
	if usingLLMReply {
		eventDescription := fmt.Sprintf("【付费留言】用户 %s 发送了 %d元 的付费留言：%s",
			msg.Data.UName, msg.Data.RMB, msg.Data.Message)
		if err := task_manager.AddText(eventDescription, task_manager.TextTypeNormal, user.GetUserVoice(msg.Data.UName)); err != nil {
			logger.Error(fmt.Sprintf("[SuperChatHandler] 添加事件到任务管理器失败: %v", err))
		}
	} else {
		reply := fmt.Sprintf("%s的付费留言（%d元）：%s，%s", msg.Data.UName, msg.Data.RMB, msg.Data.Message, common.RandomBlessing())
		if err := task_manager.AddText(reply, task_manager.TextTypeNoLLMReply, user.GetUserVoice(msg.Data.UName)); err != nil {
			logger.Error(fmt.Sprintf("[SuperChatHandler] 添加事件到任务管理器失败: %v", err))
		}
	}

	return nil
}
