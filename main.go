package main

import (
	"log"

	"github.com/CoffeeSwt/bilibili-tts-chat/bili"
)

func main() {
	// 创建应用管理器
	app := bili.NewAppManager()

	// 启动应用
	if err := app.Start(); err != nil {
		log.Fatalf("启动应用失败: %v", err)
	}

	// 确保应用正常关闭
	defer func() {
		if err := app.Stop(); err != nil {
			log.Printf("停止应用失败: %v", err)
		}
	}()

	// 等待关闭信号
	app.WaitForShutdown()
}
