package command

import (
	"fmt"
	"strings"

	"github.com/CoffeeSwt/bilibili-tts-chat/config"
	"github.com/CoffeeSwt/bilibili-tts-chat/logger"
	"github.com/CoffeeSwt/bilibili-tts-chat/response"
	"github.com/CoffeeSwt/bilibili-tts-chat/task_manager"
	"github.com/CoffeeSwt/bilibili-tts-chat/user"
)

type Handler func(msg *response.DanmakuMessage, arg string) error

var Exact = map[string]Handler{
	"我的音色": handleQueryVoice,
	"换音色":  handleRandomSwitchVoice,
}

var Prefix = map[string]Handler{
	"换": handleSwitchVoiceByName,
}

func handleQueryVoice(msg *response.DanmakuMessage, _ string) error {
	currentVoice := user.GetUserVoice(msg.Data.UName)
	var queryMessage string
	if currentVoice != nil {
		queryMessage = fmt.Sprintf("%s 当前使用的音色是：%s", msg.Data.UName, currentVoice.Name)
		logger.Info(fmt.Sprintf("[DanmakuHandler] 用户 %s 查询当前音色: %s", msg.Data.UName, currentVoice.Name))
	} else {
		queryMessage = fmt.Sprintf("%s 当前没有设置音色，将为您分配默认音色", msg.Data.UName)
		logger.Info(fmt.Sprintf("[DanmakuHandler] 用户 %s 查询音色时发现未设置，将分配默认音色", msg.Data.UName))
		currentVoice = user.GetUserVoice(msg.Data.UName)
	}
	user.UpdateUserActivity(msg.Data.UName)
	if err := task_manager.AddText(queryMessage, task_manager.TextTypeCommand, currentVoice); err != nil {
		logger.Error(fmt.Sprintf("[DanmakuHandler] 添加音色查询结果到任务管理器失败: %v", err))
	}
	return nil
}

func handleRandomSwitchVoice(msg *response.DanmakuMessage, _ string) error {
	v := config.GetRandomVoice()
	switchMessage := fmt.Sprintf("%s 的播报音色已随机切换为 %s", msg.Data.UName, v.Name)
	logger.Info(fmt.Sprintf("[DanmakuHandler] 用户 %s 随机切换音色为: %s", msg.Data.UName, v.Name))
	user.SetUserVoice(msg.Data.UName, v.VoiceType)
	user.UpdateUserActivity(msg.Data.UName)
	if err := task_manager.AddText(switchMessage, task_manager.TextTypeCommand, v); err != nil {
		logger.Error(fmt.Sprintf("[DanmakuHandler] 添加事件到任务管理器失败: %v", err))
	}
	return nil
}

func handleSwitchVoiceByName(msg *response.DanmakuMessage, arg string) error {
	var v *config.Voice
	var switchMessage string
	voiceName := strings.TrimSpace(arg)
	if voiceName != "" {
		targetVoice := config.GetVoiceByName(voiceName)
		if targetVoice != nil {
			v = targetVoice
			switchMessage = fmt.Sprintf("%s 的播报音色已切换为 %s", msg.Data.UName, v.Name)
			logger.Info(fmt.Sprintf("[DanmakuHandler] 用户 %s 指定切换音色为: %s", msg.Data.UName, v.Name))
		} else {
			v = config.GetRandomVoice()
			switchMessage = fmt.Sprintf("%s 的播报音色已切换为 %s", msg.Data.UName, v.Name)
			logger.Info(fmt.Sprintf("[DanmakuHandler] 用户 %s 指定的音色 \"%s\" 不存在，随机切换为: %s", msg.Data.UName, voiceName, v.Name))
		}
	} else {
		v = config.GetRandomVoice()
		switchMessage = fmt.Sprintf("%s 指定的音色不存在，已随机切换为 %s", msg.Data.UName, v.Name)
		logger.Info(fmt.Sprintf("[DanmakuHandler] 用户 %s 指定音色不存在，随机切换为: %s", msg.Data.UName, v.Name))
	}
	if v != nil {
		user.SetUserVoice(msg.Data.UName, v.VoiceType)
		user.UpdateUserActivity(msg.Data.UName)
		if err := task_manager.AddText(switchMessage, task_manager.TextTypeCommand, v); err != nil {
			logger.Error(fmt.Sprintf("[DanmakuHandler] 添加事件到任务管理器失败: %v", err))
		}
	}
	return nil
}
