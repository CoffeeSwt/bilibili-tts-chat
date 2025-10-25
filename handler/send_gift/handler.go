package send_gift

import (
	"encoding/json"
	"fmt"

	"github.com/CoffeeSwt/bilibili-tts-chat/logger"
	"github.com/CoffeeSwt/bilibili-tts-chat/response"
	"github.com/CoffeeSwt/bilibili-tts-chat/task_manager"
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

	// 构建结构化的事件描述，方便AI理解和回复
	var eventDescription string
	if msg.Data.GiftNum > 1 {
		eventDescription = fmt.Sprintf("【礼物】用户 %s 送出了 %d个 %s（总价值：%d）", 
			msg.Data.UName, msg.Data.GiftNum, msg.Data.GiftName, msg.Data.Price)
	} else {
		eventDescription = fmt.Sprintf("【礼物】用户 %s 送出了 %s（价值：%d）", 
			msg.Data.UName, msg.Data.GiftName, msg.Data.Price)
	}

	// 将事件描述添加到任务管理器
	if err := task_manager.AddText(eventDescription); err != nil {
		logger.Error(fmt.Sprintf("[GiftHandler] 添加事件到任务管理器失败: %v", err))
	}

	return nil
}
