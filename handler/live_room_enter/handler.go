package live_room_enter

import (
	"encoding/json"
	"fmt"

	"github.com/CoffeeSwt/bilibili-tts-chat/config"
	"github.com/CoffeeSwt/bilibili-tts-chat/logger"
	"github.com/CoffeeSwt/bilibili-tts-chat/response"
	"github.com/CoffeeSwt/bilibili-tts-chat/task_manager"
	"github.com/CoffeeSwt/bilibili-tts-chat/user"
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

	usingLLMReply := config.GetUseLLMReplay()
	if usingLLMReply {
		eventDescription := fmt.Sprintf("【进入房间】用户 %s 进入了直播间", msg.Data.UName)
		if err := task_manager.AddText(eventDescription, task_manager.TextTypeNormal, user.GetUserVoice(msg.Data.UName)); err != nil {
			logger.Error(fmt.Sprintf("[RoomHandler] 添加事件到任务管理器失败: %v", err))
		}
	} else {
		reply := fmt.Sprintf("欢迎%s进入直播间", msg.Data.UName)
		if err := task_manager.AddText(reply, task_manager.TextTypeNoLLMReply, user.GetUserVoice(msg.Data.UName)); err != nil {
			logger.Error(fmt.Sprintf("[RoomHandler] 添加事件到任务管理器失败: %v", err))
		}
	}

	return nil
}
