package like

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

// HandleLike 处理点赞消息
// 这是点赞消息的核心处理函数，负责解析和处理用户的点赞行为
func HandleLike(cmdData []byte) error {
	var msg response.LikeMessage
	if err := json.Unmarshal(cmdData, &msg); err != nil {
		logger.Error(fmt.Sprintf("[LikeHandler] 解析点赞消息失败: %v", err))
		return err
	}

	logger.Info(fmt.Sprintf("[点赞] 用户: %s, 点赞数: %d, 房间: %d",
		msg.Data.UName, msg.Data.LikeCount, msg.Data.RoomID))

	usingLLMReply := config.GetUseLLMReplay()
	if usingLLMReply {
		eventDescription := fmt.Sprintf("【点赞】用户 %s 为直播间点了 %d 个赞",
			msg.Data.UName, msg.Data.LikeCount)
		if err := task_manager.AddText(eventDescription, task_manager.TextTypeNormal, user.GetUserVoice(msg.Data.UName)); err != nil {
			logger.Error(fmt.Sprintf("[LikeHandler] 添加事件到任务管理器失败: %v", err))
		}
	} else {
		reply := fmt.Sprintf("感谢%s的点赞", msg.Data.UName)
		if msg.Data.LikeCount > 5 {
			reply = fmt.Sprintf("感谢%s的点赞，%s", msg.Data.UName, common.RandomBlessing())
		}
		if err := task_manager.AddText(reply, task_manager.TextTypeNoLLMReply, user.GetUserVoice(msg.Data.UName)); err != nil {
			logger.Error(fmt.Sprintf("[LikeHandler] 添加事件到任务管理器失败: %v", err))
		}
	}

	return nil
}
