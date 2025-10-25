package like

import (
	"encoding/json"
	"fmt"

	"github.com/CoffeeSwt/bilibili-tts-chat/logger"
	"github.com/CoffeeSwt/bilibili-tts-chat/response"
)

// HandleLike 处理点赞消息
// 这是点赞消息的核心处理函数，负责解析和处理用户的点赞行为
func HandleLike(cmdData []byte) error {
	var msg response.LikeMessage
	if err := json.Unmarshal(cmdData, &msg); err != nil {
		logger.Error(fmt.Sprintf("[LikeHandler] 解析点赞消息失败: %v", err))
		return err
	}

	logger.Info(fmt.Sprintf("[点赞] 用户: %s, 点赞数: %d, 房间: %d",
		msg.Data.UName, msg.Data.LikeCount, msg.Data.RoomID))

	return nil
}
