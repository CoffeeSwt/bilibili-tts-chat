package live_end

import (
	"encoding/json"
	"fmt"

	"github.com/CoffeeSwt/bilibili-tts-chat/logger"
	"github.com/CoffeeSwt/bilibili-tts-chat/response"
)

func HandleLiveEnd(cmdData []byte) error {
	var msg response.LiveEndMessage
	if err := json.Unmarshal(cmdData, &msg); err != nil {
		logger.Error(fmt.Sprintf("[LiveEndHandler] 解析直播结束消息失败: %v", err))
		return err
	}

	logger.Info(fmt.Sprintf("[直播结束] 房间: %d, 结束时间: %d",
		msg.Data.RoomID, msg.Data.Timestamp))

	return nil
}
