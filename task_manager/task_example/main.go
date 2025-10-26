package main

import (
	"github.com/CoffeeSwt/bilibili-tts-chat/config"
	"github.com/CoffeeSwt/bilibili-tts-chat/task_manager"
)

func main() {
	task_manager.AddText("你好", task_manager.TextTypeNormal, config.GetRandomVoice())
}
