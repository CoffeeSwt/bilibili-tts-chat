package guard

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

	guardLevels := map[int]string{1: "总督", 2: "提督", 3: "舰长"}
	guardName := "大航海"
	if level, exists := guardLevels[msg.Data.GuardLevel]; exists {
		guardName = level
	}

	usingLLMReply := config.GetUseLLMReplay()
	if usingLLMReply {
		var eventDescription string
		if msg.Data.GuardNum > 1 {
			eventDescription = fmt.Sprintf("【大航海】用户 %s 购买了 %d%s %s（总价值：%d）",
				msg.Data.UserInfo.UName, msg.Data.GuardNum, msg.Data.GuardUnit, guardName, msg.Data.Price)
		} else {
			eventDescription = fmt.Sprintf("【大航海】用户 %s 购买了 %s（价值：%d）",
				msg.Data.UserInfo.UName, guardName, msg.Data.Price)
		}
		if err := task_manager.AddText(eventDescription, task_manager.TextTypeNormal, user.GetUserVoice(msg.Data.UserInfo.UName)); err != nil {
			logger.Error(fmt.Sprintf("[GuardHandler] 添加事件到任务管理器失败: %v", err))
		}
	} else {
		var durationPrefix string
		if msg.Data.GuardUnit == "月" {
			if msg.Data.GuardNum == 1 {
				durationPrefix = "一个月的"
			} else {
				durationPrefix = fmt.Sprintf("%d个月的", msg.Data.GuardNum)
			}
		} else if msg.Data.GuardUnit != "" {
			if msg.Data.GuardNum == 1 {
				durationPrefix = fmt.Sprintf("一%s的", msg.Data.GuardUnit)
			} else {
				durationPrefix = fmt.Sprintf("%d%s的", msg.Data.GuardNum, msg.Data.GuardUnit)
			}
		} else {
			if msg.Data.GuardNum == 1 {
				durationPrefix = "一个"
			} else {
				durationPrefix = fmt.Sprintf("%d个", msg.Data.GuardNum)
			}
		}

		reply := fmt.Sprintf("感谢%s给主播赠送了%s%s，%s", msg.Data.UserInfo.UName, durationPrefix, guardName, common.RandomBlessing())
		if err := task_manager.AddText(reply, task_manager.TextTypeNoLLMReply, user.GetUserVoice(msg.Data.UserInfo.UName)); err != nil {
			logger.Error(fmt.Sprintf("[GuardHandler] 添加事件到任务管理器失败: %v", err))
		}
	}

	return nil
}
