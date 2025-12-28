package bili

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/CoffeeSwt/bilibili-tts-chat/config"
	"github.com/CoffeeSwt/bilibili-tts-chat/logger"
	"github.com/CoffeeSwt/bilibili-tts-chat/task_manager"
)

// AppManager 应用管理器，封装所有B站相关的逻辑
type AppManager struct {
	gameID        string
	appID         int64
	heartbeatStop chan struct{}
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	isRunning     bool
	mu            sync.RWMutex
}

// StartAppRequest 启动应用请求
type StartAppRequest struct {
	Code  string `json:"code"`
	AppId int64  `json:"app_id"`
}

// StartAppRespData 启动应用响应数据
type StartAppRespData struct {
	GameInfo      GameInfo      `json:"game_info"`
	WebsocketInfo WebSocketInfo `json:"websocket_info"`
	AnchorInfo    AnchorInfo    `json:"anchor_info"`
}

// GameInfo 场次信息
type GameInfo struct {
	GameId string `json:"game_id"`
}

// WebSocketInfo 长连信息
type WebSocketInfo struct {
	AuthBody string   `json:"auth_body"`
	WssLink  []string `json:"wss_link"`
}

// AnchorInfo 主播信息
type AnchorInfo struct {
	RoomId int64  `json:"room_id"`
	Uname  string `json:"uname"`
	Uface  string `json:"uface"`
	Uid    int64  `json:"uid"`
	OpenId string `json:"open_id"`
}

// EndAppRequest 关闭应用请求
type EndAppRequest struct {
	GameId string `json:"game_id"`
	AppId  int64  `json:"app_id"`
}

// AppHeartbeatReq 应用心跳请求
type AppHeartbeatReq struct {
	GameId string `json:"game_id"`
}

// NewAppManager 创建新的应用管理器
func NewAppManager() *AppManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &AppManager{
		appID:         int64(config.GetBiliAppID()),
		heartbeatStop: make(chan struct{}),
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Start 启动应用
func (am *AppManager) Start() error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if am.isRunning {
		return fmt.Errorf("应用已经在运行中")
	}

	logger.Info("正在启动应用管理器...")

	// 启动B站应用
	startAppResp, err := am.startApp()
	if err != nil {
		return fmt.Errorf("启动B站应用失败: %w", err)
	}

	// 启动心跳
	am.startHeartbeat()

	// 启动WebSocket连接
	if err := am.startWebSocket(startAppResp); err != nil {
		return fmt.Errorf("启动WebSocket连接失败: %w", err)
	}

	// 启动事件驱动任务处理器
	am.startEventDrivenTaskProcessor()

	am.isRunning = true
	logger.Info("应用管理器启动成功")
	return nil
}

// Stop 停止应用
func (am *AppManager) Stop() error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if !am.isRunning {
		return nil
	}

	logger.Info("正在停止应用管理器...")

	// 停止心跳
	close(am.heartbeatStop)

	// 取消上下文
	am.cancel()

	// 等待所有goroutine结束
	am.wg.Wait()

	// 关闭B站应用
	if am.gameID != "" {
		if err := am.endApp(); err != nil {
			log.Printf("关闭B站应用失败: %v", err)
		}
	}

	// 显式停止WebSocket连接
	if err := StopWebsocket(); err != nil {
		logger.Error("停止WebSocket连接失败", err)
	}

	// 清理任务管理器状态
	task_manager.ClearTasks()

	am.isRunning = false
	logger.Info("应用管理器已停止")
	return nil
}

// WaitForShutdown 等待关闭信号
func (am *AppManager) WaitForShutdown() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)

	for {
		s := <-c
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			logger.Info("收到关闭信号，正在退出...")
			return
		case syscall.SIGHUP:
			logger.Info("收到SIGHUP信号")
		default:
			return
		}
	}
}

// startApp 启动B站应用
func (am *AppManager) startApp() (*StartAppRespData, error) {
	logger.Info("正在启动B站应用...")
	logger.Info(fmt.Sprintf("StartApp请求参数: Code=%s, AppId=%d", config.GetRoomIDCode(), am.appID))
	resp, err := am.StartApp(config.GetRoomIDCode(), am.appID)
	if err != nil {
		return nil, fmt.Errorf("StartApp API调用失败: %w", err)
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("StartApp API返回错误: Code=%d, Message=%s", resp.Code, resp.Message)
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("StartApp API返回的Data为空")
	}

	startAppRespData := &StartAppRespData{}
	err = json.Unmarshal(resp.Data, startAppRespData)
	if err != nil {
		return nil, fmt.Errorf("解析StartApp响应数据失败: %w", err)
	}

	am.gameID = startAppRespData.GameInfo.GameId
	logger.Info("B站应用启动成功")
	return startAppRespData, nil
}

// startHeartbeat 启动心跳
func (am *AppManager) startHeartbeat() {
	am.wg.Add(1)
	go func() {
		defer am.wg.Done()
		ticker := time.NewTicker(20 * time.Second)
		defer ticker.Stop()

		logger.Info("正在启动心跳服务...")
		for {
			select {
			case <-ticker.C:
				if am.gameID != "" {
					_, err := am.AppHeart(am.gameID)
					if err != nil {
						logger.Error("心跳发送失败", err)
					}
				}
			case <-am.heartbeatStop:
				logger.Info("心跳服务已停止")
				return
			case <-am.ctx.Done():
				logger.Info("心跳服务因上下文取消而停止")
				return
			}
		}
	}()
}

// startWebSocket 启动WebSocket连接
func (am *AppManager) startWebSocket(startAppResp *StartAppRespData) error {
	if len(startAppResp.WebsocketInfo.WssLink) == 0 {
		return fmt.Errorf("WebSocket链接为空")
	}

	logger.Info("正在启动WebSocket连接...")
	err := StartWebsocket(startAppResp.WebsocketInfo.WssLink[0], startAppResp.WebsocketInfo.AuthBody)
	if err != nil {
		return fmt.Errorf("启动WebSocket失败: %w", err)
	}

	logger.Info("WebSocket连接启动成功")
	return nil
}

// endApp 关闭B站应用
func (am *AppManager) endApp() error {
	logger.Info("正在关闭B站应用...")
	_, err := am.EndApp(am.gameID, am.appID)
	if err != nil {
		return err
	}
	logger.Info("B站应用关闭成功")
	return nil
}

// StartApp 启动app
func (am *AppManager) StartApp(code string, appId int64) (resp BaseResp, err error) {
	startAppReq := StartAppRequest{
		Code:  code,
		AppId: appId,
	}
	reqJson, err := json.Marshal(startAppReq)
	if err != nil {
		return resp, fmt.Errorf("序列化StartApp请求参数失败: %w", err)
	}
	return ApiRequest(string(reqJson), "/v2/app/start")
}

// AppHeart app心跳
func (am *AppManager) AppHeart(gameId string) (resp BaseResp, err error) {
	appHeartbeatReq := AppHeartbeatReq{
		GameId: gameId,
	}
	reqJson, _ := json.Marshal(appHeartbeatReq)
	return ApiRequest(string(reqJson), "/v2/app/heartbeat")
}

// EndApp 关闭app
func (am *AppManager) EndApp(gameId string, appId int64) (resp BaseResp, err error) {
	endAppReq := EndAppRequest{
		GameId: gameId,
		AppId:  appId,
	}
	reqJson, _ := json.Marshal(endAppReq)
	return ApiRequest(string(reqJson), "/v2/app/end")
}

// startEventDrivenTaskProcessor 启动事件驱动任务处理器
func (am *AppManager) startEventDrivenTaskProcessor() {
	am.wg.Add(1)
	go func() {
		defer am.wg.Done()

		logger.Info("事件驱动任务处理器已启动")
		taskNotify := task_manager.GetTaskNotifyChannel()

		for {
			select {
			case <-am.ctx.Done():
				logger.Info("事件驱动任务处理器收到停止信号，正在退出...")
				return
			case <-taskNotify:
				// 收到任务通知，开始处理循环
				logger.Info("收到任务通知，开始处理任务...")
				am.processTaskLoop()
			}
		}
	}()
}

// processTaskLoop 处理任务循环
func (am *AppManager) processTaskLoop() {
	for {
		// 检查上下文是否已取消
		select {
		case <-am.ctx.Done():
			logger.Info("任务处理循环收到停止信号，正在退出...")
			return
		default:
		}

		// 检查是否有任务需要处理
		if task_manager.IsTaskRunning() {
			logger.Info("检测到有任务需要处理，开始执行...")
			task_manager.PlayEventTasks(am.ctx)
			// PlayEventTasks 完成后，立即检查是否还有新任务
			logger.Info("任务执行完成，检查是否有新任务...")
		} else {
			// 没有任务了，退出循环等待下一个通知
			logger.Info("没有更多任务，等待下一个通知...")
			break
		}
	}
}
