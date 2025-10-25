package live_end

import (
	"encoding/json"
	"fmt"

	"github.com/CoffeeSwt/bilibili-tts-chat/logger"
	"github.com/CoffeeSwt/bilibili-tts-chat/response"
	"github.com/CoffeeSwt/bilibili-tts-chat/task_manager"
)

func HandleLiveEnd(cmdData []byte) error {
	var msg response.LiveEndMessage
	if err := json.Unmarshal(cmdData, &msg); err != nil {
		logger.Error(fmt.Sprintf("[LiveEndHandler] 解析直播结束消息失败: %v", err))
		return err
	}

	logger.Info(fmt.Sprintf("[直播结束] 房间: %d, 结束时间: %d",
		msg.Data.RoomID, msg.Data.Timestamp))

	// 构建结构化的事件描述，方便AI理解和回复
	eventDescription := fmt.Sprintf("【直播结束】主播结束了直播，房间号：%d", msg.Data.RoomID)

	// 将事件描述添加到任务管理器
	if err := task_manager.AddText(eventDescription); err != nil {
		logger.Error(fmt.Sprintf("[LiveEndHandler] 添加事件到任务管理器失败: %v", err))
	}

	return nil
}
