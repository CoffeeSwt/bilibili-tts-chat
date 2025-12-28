package live_room_enter

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/CoffeeSwt/bilibili-tts-chat/config"
	"github.com/CoffeeSwt/bilibili-tts-chat/intro_promot"
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

	voice := user.GetUserVoice(msg.Data.UName)
	logger.Info(fmt.Sprintf("[进入房间][%s][%s] (OpenID: %s)",
		msg.Data.UName, voice.Name, msg.Data.OpenID))

	usingLLMReply := config.GetUseLLMReplay()
	if usingLLMReply {
		eventDescription := fmt.Sprintf("【进入房间】用户 %s 进入了直播间", msg.Data.UName)
		if err := task_manager.AddText(eventDescription, task_manager.TextTypeNormal, voice); err != nil {
			logger.Error(fmt.Sprintf("[RoomHandler] 添加事件到任务管理器失败: %v", err))
		}
	} else {
		// 80% 概率触发引导词
		rand.Seed(time.Now().UnixNano())
		enterPromot := ""
		if rand.Float64() < 0.8 {
			enterPromot = "，" + intro_promot.GetEnterPromot()
		}

		reply := fmt.Sprintf("欢迎%s进入直播间%s", msg.Data.UName, enterPromot)
		if err := task_manager.AddText(reply, task_manager.TextTypeNoLLMReply, voice); err != nil {
			logger.Error(fmt.Sprintf("[RoomHandler] 添加事件到任务管理器失败: %v", err))
		}
	}

	return nil
}
