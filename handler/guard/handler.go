package guard

import (
	"encoding/json"
	"fmt"

	"github.com/CoffeeSwt/bilibili-tts-chat/logger"
	"github.com/CoffeeSwt/bilibili-tts-chat/response"
	"github.com/CoffeeSwt/bilibili-tts-chat/task_manager"
)

// HandleGuard 处理大航海消息
// 这是大航海消息的核心处理函数，负责解析和处理用户购买大航海的消息
func HandleGuard(cmdData []byte) error {
	var msg response.GuardMessage
	if err := json.Unmarshal(cmdData, &msg); err != nil {
		logger.Error(fmt.Sprintf("[GuardHandler] 解析大航海消息失败: %v", err))
		return err
	}

	logger.Info(fmt.Sprintf("收到大航海: %s, 等级: %d, 数量: %d%s, 价格: %d",
		msg.Data.UserInfo.UName, msg.Data.GuardLevel, msg.Data.GuardNum, msg.Data.GuardUnit,
		msg.Data.Price))

	// 构建结构化的事件描述，方便AI理解和回复
	guardLevels := map[int]string{1: "总督", 2: "提督", 3: "舰长"}
	guardName := "大航海"
	if level, exists := guardLevels[msg.Data.GuardLevel]; exists {
		guardName = level
	}

	var eventDescription string
	if msg.Data.GuardNum > 1 {
		eventDescription = fmt.Sprintf("【大航海】用户 %s 购买了 %d%s %s（总价值：%d）", 
			msg.Data.UserInfo.UName, msg.Data.GuardNum, msg.Data.GuardUnit, guardName, msg.Data.Price)
	} else {
		eventDescription = fmt.Sprintf("【大航海】用户 %s 购买了 %s（价值：%d）", 
			msg.Data.UserInfo.UName, guardName, msg.Data.Price)
	}

	// 将事件描述添加到任务管理器
	if err := task_manager.AddText(eventDescription); err != nil {
		logger.Error(fmt.Sprintf("[GuardHandler] 添加事件到任务管理器失败: %v", err))
	}

	return nil
}
