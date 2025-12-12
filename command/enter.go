package command

import (
	"strings"

	"github.com/CoffeeSwt/bilibili-tts-chat/response"
)

func CheckIfCommandAndUseHandler(msg *response.DanmakuMessage) (func(msg *response.DanmakuMessage) error, bool) {
	text := msg.Data.Msg
	if h, ok := Exact[text]; ok {
		return func(m *response.DanmakuMessage) error { return h(m, "") }, true
	}
	for prefix, h := range Prefix {
		if strings.HasPrefix(text, prefix) {
			arg := strings.TrimSpace(strings.TrimPrefix(text, prefix))
			return func(m *response.DanmakuMessage) error { return h(m, arg) }, true
		}
	}
	return nil, false
}
