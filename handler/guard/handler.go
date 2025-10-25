package guard

import (
	"encoding/json"
	"fmt"

	"github.com/CoffeeSwt/bilibili-tts-chat/logger"
	"github.com/CoffeeSwt/bilibili-tts-chat/response"
)

// HandleGuard 处理大航海消息
// 这是大航海消息的核心处理函数，负责解析和处理用户购买大航海的消息
func HandleGuard(cmdData []byte) error {
	var msg response.GuardMessage
	if err := json.Unmarshal(cmdData, &msg); err != nil {
		logger.Error(fmt.Sprintf("[GuardHandler] 解析大航海消息失败: %v", err))
		return err
	}

	logger.Info(fmt.Sprintf("收到大航海: %s, 等级: %d, 数量: %d%s, 价格: %d",
		msg.Data.UserInfo.UName, msg.Data.GuardLevel, msg.Data.GuardNum, msg.Data.GuardUnit,
		msg.Data.Price))

	return nil
}
