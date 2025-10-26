package live_room_enter

import (
	"encoding/json"
	"fmt"

	"github.com/CoffeeSwt/bilibili-tts-chat/config"
	"github.com/CoffeeSwt/bilibili-tts-chat/logger"
	"github.com/CoffeeSwt/bilibili-tts-chat/response"
	"github.com/CoffeeSwt/bilibili-tts-chat/task_manager"
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

	// 构建结构化的事件描述，方便AI理解和回复
	eventDescription := fmt.Sprintf("【进入房间】用户 %s 进入了直播间", msg.Data.UName)

	// 将事件描述添加到任务管理器
	if err := task_manager.AddText(eventDescription, task_manager.TextTypeNormal, config.GetRandomVoice()); err != nil {
		logger.Error(fmt.Sprintf("[RoomHandler] 添加事件到任务管理器失败: %v", err))
	}

	return nil
}
