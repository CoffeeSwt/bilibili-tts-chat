package live_start

import (
	"encoding/json"
	"fmt"

	"github.com/CoffeeSwt/bilibili-tts-chat/logger"
	"github.com/CoffeeSwt/bilibili-tts-chat/response"
)

func HandleLiveStart(cmdData []byte) error {
	var msg response.LiveStartMessage
	if err := json.Unmarshal(cmdData, &msg); err != nil {
		logger.Error(fmt.Sprintf("[LiveStartHandler] 解析直播开始消息失败: %v", err))
		return err
	}

	logger.Info(fmt.Sprintf("[直播开始] 房间: %d, 开始时间: %d",
		msg.Data.RoomID, msg.Data.Timestamp))

	return nil
}
