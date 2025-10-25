package dm

import (
	"encoding/json"
	"fmt"

	"github.com/CoffeeSwt/bilibili-tts-chat/logger"
	"github.com/CoffeeSwt/bilibili-tts-chat/response"
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

	return nil
}
