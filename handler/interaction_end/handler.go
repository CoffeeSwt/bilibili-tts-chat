package interaction_end

import (
	"encoding/json"
	"fmt"

	"github.com/CoffeeSwt/bilibili-tts-chat/logger"
	"github.com/CoffeeSwt/bilibili-tts-chat/response"
)

func HandleInteractionEnd(cmdData []byte) error {
	var msg response.InteractionEndMessage
	if err := json.Unmarshal(cmdData, &msg); err != nil {
		logger.Error(fmt.Sprintf("[InteractionEndHandler] 解析互动结束消息失败: %v", err))
		return err
	}

	logger.Info("互动结束")

	return nil
}
