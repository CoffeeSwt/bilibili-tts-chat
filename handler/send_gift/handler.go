package send_gift

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

// HandleGift 处理礼物消息
// 这是礼物消息的核心处理函数，负责解析和处理用户发送的礼物
func HandleGift(cmdData []byte) error {
	var msg response.GiftMessage
	if err := json.Unmarshal(cmdData, &msg); err != nil {
		logger.Error(fmt.Sprintf("[GiftHandler] 解析礼物消息失败: %v", err))
		return err
	}

	logger.Info(fmt.Sprintf("[礼物] 用户: %s, 礼物: %s x%d, 价值: %d, 房间: %d",
		msg.Data.UName, msg.Data.GiftName, msg.Data.GiftNum, msg.Data.Price, msg.Data.RoomID))

	usingLLMReply := config.GetUseLLMReplay()
	if usingLLMReply {
		var eventDescription string
		if msg.Data.GiftNum > 1 {
			eventDescription = fmt.Sprintf("【礼物】用户 %s 送出了 %d个 %s（总价值：%d）",
				msg.Data.UName, msg.Data.GiftNum, msg.Data.GiftName, msg.Data.Price)
		} else {
			eventDescription = fmt.Sprintf("【礼物】用户 %s 送出了 %s（价值：%d）",
				msg.Data.UName, msg.Data.GiftName, msg.Data.Price)
		}
		if err := task_manager.AddText(eventDescription, task_manager.TextTypeNormal, user.GetUserVoice(msg.Data.UName)); err != nil {
			logger.Error(fmt.Sprintf("[GiftHandler] 添加事件到任务管理器失败: %v", err))
		}
	} else {
		var reply string
		if msg.Data.GiftNum > 1 {
			reply = fmt.Sprintf("感谢%s赠送了 %d个 %s，%s", msg.Data.UName, msg.Data.GiftNum, msg.Data.GiftName, common.RandomBlessing())
		} else {
			reply = fmt.Sprintf("感谢%s赠送了 %s，%s", msg.Data.UName, msg.Data.GiftName, common.RandomBlessing())
		}
		if err := task_manager.AddText(reply, task_manager.TextTypeNoLLMReply, user.GetUserVoice(msg.Data.UName)); err != nil {
			logger.Error(fmt.Sprintf("[GiftHandler] 添加事件到任务管理器失败: %v", err))
		}
	}

	return nil
}
