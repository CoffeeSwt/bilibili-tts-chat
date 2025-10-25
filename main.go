package main

import (
	"log"

	"github.com/CoffeeSwt/bilibili-tts-chat/bili"
	"github.com/CoffeeSwt/bilibili-tts-chat/logger"
)

func main() {
	// 初始化文件日志系统
	if err := logger.InitFileLogging(); err != nil {
		log.Fatalf("初始化日志系统失败: %v", err)
	}
	logger.Info("日志系统初始化成功")

	// 创建应用管理器
	app := bili.NewAppManager()

	// 启动应用
	if err := app.Start(); err != nil {
		log.Fatalf("启动应用失败: %v", err)
	}

	// 确保应用正常关闭
	defer func() {
		logger.Info("正在关闭应用...")
		if err := app.Stop(); err != nil {
			logger.Error("停止应用失败", "error", err)
		}
		
		// 刷新并关闭日志文件
		if err := logger.FlushLogs(); err != nil {
			log.Printf("刷新日志失败: %v", err)
		}
		
		// 关闭文件写入器
		if fw := logger.GetFileWriter(); fw != nil {
			if err := fw.Close(); err != nil {
				log.Printf("关闭日志文件失败: %v", err)
			}
		}
		logger.Info("应用已安全关闭")
	}()

	// 等待关闭信号
	app.WaitForShutdown()
}
