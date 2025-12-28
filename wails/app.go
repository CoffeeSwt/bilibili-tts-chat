package main

import (
	"context"
	"fmt"
	"log"

	"github.com/CoffeeSwt/bilibili-tts-chat/bili"
	"github.com/CoffeeSwt/bilibili-tts-chat/config"
	"github.com/CoffeeSwt/bilibili-tts-chat/logger"
	"github.com/CoffeeSwt/bilibili-tts-chat/user"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx        context.Context
	appManager *bili.AppManager
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	if err := logger.InitFileLogging(); err != nil {
		log.Fatalf("初始化日志系统失败: %v", err)
	}
	logger.Info("日志系统初始化成功")

	// 设置日志回调，将日志发送到前端
	logger.SetLogCallback(func(level logger.LogLevel, timestamp, location, message string, keyvals ...interface{}) {
		runtime.EventsEmit(ctx, "log", map[string]interface{}{
			"level":     string(level),
			"timestamp": timestamp,
			"location":  location,
			"message":   message,
		})
	})

	a.appManager = bili.NewAppManager()
	if err := a.appManager.Start(); err != nil {
		logger.Error("启动应用失败", "error", err)
		// 不再退出，允许前端查看日志
	}
}

func (a *App) shutdown(ctx context.Context) {
	_ = ctx
	logger.Info("正在关闭应用...")
	if a.appManager != nil {
		if err := a.appManager.Stop(); err != nil {
			logger.Error("停止应用失败", "error", err)
		}
	}

	logger.Info("正在保存用户音频配置...")
	if err := user.SaveUserVoices(); err != nil {
		logger.Error("保存用户音频配置失败", "error", err)
		log.Printf("保存用户音频配置失败: %v", err)
	} else {
		logger.Info("用户音频配置保存成功")
	}

	if err := logger.FlushLogs(); err != nil {
		log.Printf("刷新日志失败: %v", err)
	}

	if fw := logger.GetFileWriter(); fw != nil {
		if err := fw.Close(); err != nil {
			log.Printf("关闭日志文件失败: %v", err)
		}
	}
	logger.Info("应用已安全关闭")
}

func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// GetConfig 获取当前配置
func (a *App) GetConfig() *config.UserConfig {
	return config.GetUserConfig()
}

// SaveConfig 保存配置
func (a *App) SaveConfig(cfg config.UserConfig) error {
	return config.SaveUserConfig(cfg)
}

// RestartApp 重启应用逻辑（停止旧实例并启动新实例）
func (a *App) RestartApp() error {
	logger.Info("前端触发应用重启...")
	
	if a.appManager != nil {
		if err := a.appManager.Stop(); err != nil {
			logger.Error("重启前停止应用失败", "error", err)
		}
	}

	a.appManager = bili.NewAppManager()
	if err := a.appManager.Start(); err != nil {
		logger.Error("重启应用失败", "error", err)
		return fmt.Errorf("重启失败: %v", err)
	}
	
	logger.Info("应用重启成功")
	return nil
}
