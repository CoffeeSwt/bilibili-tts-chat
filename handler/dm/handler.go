package dm

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/CoffeeSwt/bilibili-tts-chat/config"
	"github.com/CoffeeSwt/bilibili-tts-chat/logger"
	"github.com/CoffeeSwt/bilibili-tts-chat/response"
	"github.com/CoffeeSwt/bilibili-tts-chat/task_manager"
	"github.com/CoffeeSwt/bilibili-tts-chat/user"
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

	// 处理音色相关指令
	isVoiceCommand := false
	if msg.Data.Msg == "我的音色" {
		// 查询当前用户的音色
		currentVoice := user.GetUserVoice(msg.Data.UName)
		var queryMessage string

		if currentVoice != nil {
			queryMessage = fmt.Sprintf("%s 当前使用的音色是：%s", msg.Data.UName, currentVoice.Name)
			logger.Info(fmt.Sprintf("[DanmakuHandler] 用户 %s 查询当前音色: %s", msg.Data.UName, currentVoice.Name))
		} else {
			queryMessage = fmt.Sprintf("%s 当前没有设置音色，将为您分配默认音色", msg.Data.UName)
			logger.Info(fmt.Sprintf("[DanmakuHandler] 用户 %s 查询音色时发现未设置，将分配默认音色", msg.Data.UName))
			// 为用户分配一个默认音色
			currentVoice = user.GetUserVoice(msg.Data.UName)
		}

		// 更新用户活跃时间
		user.UpdateUserActivity(msg.Data.UName)
		// 将查询结果添加到任务管理器
		if err := task_manager.AddText(queryMessage, task_manager.TextTypeCommand, currentVoice); err != nil {
			logger.Error(fmt.Sprintf("[DanmakuHandler] 添加音色查询结果到任务管理器失败: %v", err))
		}
		isVoiceCommand = true
	} else if msg.Data.Msg == "换音色" || strings.HasPrefix(msg.Data.Msg, "换") {
		var v *config.Voice
		var switchMessage string

		if msg.Data.Msg == "换音色" {
			// 随机切换音色
			v = config.GetRandomVoice()
			switchMessage = fmt.Sprintf("%s 的播报音色已随机切换为 %s", msg.Data.UName, v.Name)
			logger.Info(fmt.Sprintf("[DanmakuHandler] 用户 %s 随机切换音色为: %s", msg.Data.UName, v.Name))
			isVoiceCommand = true
		} else if strings.HasPrefix(msg.Data.Msg, "换") && len(msg.Data.Msg) > 1 {
			// 指定音色名称切换
			voiceName := strings.TrimPrefix(msg.Data.Msg, "换")
			voiceName = strings.TrimSpace(voiceName) // 去除可能的空格

			if voiceName != "" {
				// 尝试根据名称查找音色
				targetVoice := config.GetVoiceByName(voiceName)
				if targetVoice != nil {
					v = targetVoice
					switchMessage = fmt.Sprintf("%s 的播报音色已切换为 %s", msg.Data.UName, v.Name)
					logger.Info(fmt.Sprintf("[DanmakuHandler] 用户 %s 指定切换音色为: %s", msg.Data.UName, v.Name))
					isVoiceCommand = true
				} else {
					// 找不到指定音色，回退到随机切换
					v = config.GetRandomVoice()
					switchMessage = fmt.Sprintf("%s 指定的音色 \"%s\" 不存在，已随机切换为 %s", msg.Data.UName, voiceName, v.Name)
					logger.Info(fmt.Sprintf("[DanmakuHandler] 用户 %s 指定的音色 \"%s\" 不存在，随机切换为: %s", msg.Data.UName, voiceName, v.Name))
					isVoiceCommand = true
				}
			} else {
				// 找不到指定音色，回退到随机切换
				v = config.GetRandomVoice()
				switchMessage = fmt.Sprintf("%s 指定的音色不存在，已随机切换为 %s", msg.Data.UName, v.Name)
				logger.Info(fmt.Sprintf("[DanmakuHandler] 用户 %s 指定音色不存在，随机切换为: %s", msg.Data.UName, v.Name))
				isVoiceCommand = true
			}
		}

		// 设置用户音色（仅对音色切换操作，不包括查询操作）
		if isVoiceCommand && v != nil && msg.Data.Msg != "我的音色" {
			user.SetUserVoice(msg.Data.UName, v.VoiceType)
			// 更新用户活跃时间
			user.UpdateUserActivity(msg.Data.UName)
			// 将事件描述添加到任务管理器
			if err := task_manager.AddText(switchMessage, task_manager.TextTypeCommand, v); err != nil {
				logger.Error(fmt.Sprintf("[DanmakuHandler] 添加事件到任务管理器失败: %v", err))
			}
		}
	}

	// 如果不是音色相关指令，按普通弹幕处理
	if !isVoiceCommand {
		// 将事件描述添加到任务管理器
		if err := task_manager.AddText(eventDescription, task_manager.TextTypeNormal, user.GetUserVoice(msg.Data.UName)); err != nil {
			logger.Error(fmt.Sprintf("[DanmakuHandler] 添加事件到任务管理器失败: %v", err))
		}
	}

	return nil
}
