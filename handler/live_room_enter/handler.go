package live_room_enter

import (
	"encoding/json"
	"fmt"

	"github.com/CoffeeSwt/bilibili-tts-chat/logger"
	"github.com/CoffeeSwt/bilibili-tts-chat/response"
)

// HandleRoomEnter 处理用户进入房间消息
// 当有用户进入直播间时触发，可以用于欢迎消息、统计等
func HandleRoomEnter(cmdData []byte) error {
	var msg response.RoomEnterMessage
	if err := json.Unmarshal(cmdData, &msg); err != nil {
		logger.Error(fmt.Sprintf("[RoomHandler] 解析进入房间消息失败: %v", err))
		return err
	}

	logger.Info(fmt.Sprintf("[进入房间] 用户: %s (OpenID: %s), 房间: %d",
		msg.Data.UName, msg.Data.OpenID, msg.Data.RoomID))
	// 播放欢迎语音
	return nil
}
