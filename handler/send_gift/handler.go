package send_gift

import (
	"encoding/json"
	"fmt"

	"github.com/CoffeeSwt/bilibili-tts-chat/logger"
	"github.com/CoffeeSwt/bilibili-tts-chat/response"
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

	return nil
}
