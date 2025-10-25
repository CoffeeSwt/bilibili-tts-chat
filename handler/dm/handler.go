package dm

import (
	"encoding/json"
	"fmt"

	"github.com/CoffeeSwt/bilibili-tts-chat/logger"
	"github.com/CoffeeSwt/bilibili-tts-chat/response"
	"github.com/CoffeeSwt/bilibili-tts-chat/task_manager"
)

// HandleDanmaku 处理弹幕消息
// 这是弹幕消息的核心处理函数，负责解析和处理用户发送的弹幕
func HandleDanmaku(cmdData []byte) error {
	var msg response.DanmakuMessage
	if err := json.Unmarshal(cmdData, &msg); err != nil {
		logger.Error(fmt.Sprintf("[DanmakuHandler] 解析弹幕消息失败: %v", err))
		return err
	}

	logger.Info(fmt.Sprintf("收到弹幕: %s, 用户: %s, 房间: %d",
		msg.Data.Msg, msg.Data.UName, msg.Data.RoomID))

	// 构建结构化的事件描述，方便AI理解和回复
	eventDescription := fmt.Sprintf("【弹幕消息】用户 %s 发送了弹幕：%s", msg.Data.UName, msg.Data.Msg)

	// 添加用户等级信息
	if msg.Data.GuardLevel > 0 {
		guardLevels := map[int]string{1: "总督", 2: "提督", 3: "舰长"}
		if level, exists := guardLevels[msg.Data.GuardLevel]; exists {
			eventDescription += fmt.Sprintf("（%s）", level)
		}
	}

	// // 添加粉丝勋章信息
	// if msg.Data.FansMedalWearingStatus && msg.Data.FansMedalName != "" {
	// 	eventDescription += fmt.Sprintf("（佩戴勋章：%s %d级）", msg.Data.FansMedalName, msg.Data.FansMedalLevel)
	// }

	// 将事件描述添加到任务管理器
	if err := task_manager.AddText(eventDescription); err != nil {
		logger.Error(fmt.Sprintf("[DanmakuHandler] 添加事件到任务管理器失败: %v", err))
	}

	return nil
}
