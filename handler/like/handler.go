package like

import (
	"encoding/json"
	"fmt"

	"github.com/CoffeeSwt/bilibili-tts-chat/config"
	"github.com/CoffeeSwt/bilibili-tts-chat/logger"
	"github.com/CoffeeSwt/bilibili-tts-chat/response"
	"github.com/CoffeeSwt/bilibili-tts-chat/task_manager"
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

	// 构建结构化的事件描述，方便AI理解和回复
	eventDescription := fmt.Sprintf("【点赞】用户 %s 为直播间点了 %d 个赞",
		msg.Data.UName, msg.Data.LikeCount)

	// 将事件描述添加到任务管理器
	if err := task_manager.AddText(eventDescription, task_manager.TextTypeNormal, config.GetRandomVoice()); err != nil {
		logger.Error(fmt.Sprintf("[LikeHandler] 添加事件到任务管理器失败: %v", err))
	}

	return nil
}
