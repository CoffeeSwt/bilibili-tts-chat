package main

import (
	"encoding/json"
	"fmt"
	"github.com/CoffeeSwt/bilibili-tts-chat/net"
	"github.com/CoffeeSwt/bilibili-tts-chat/tts_api"
	user_voice "github.com/CoffeeSwt/bilibili-tts-chat/user"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/CoffeeSwt/bilibili-tts-chat/config"
)

type StartAppRequest struct {
	// 主播身份码
	Code string `json:"code"`
	// 项目id
	AppId int64 `json:"app_id"`
}

type StartAppRespData struct {
	// 场次信息
	GameInfo GameInfo `json:"game_info"`
	// 长连信息
	WebsocketInfo WebSocketInfo `json:"websocket_info"`
	// 主播信息
	AnchorInfo AnchorInfo `json:"anchor_info"`
}

type GameInfo struct {
	GameId string `json:"game_id"`
}

type WebSocketInfo struct {
	//  长连使用的请求json体 第三方无需关注内容,建立长连时使用即可
	AuthBody string `json:"auth_body"`
	//  wss 长连地址
	WssLink []string `json:"wss_link"`
}

type AnchorInfo struct {
	//主播房间号
	RoomId int64 `json:"room_id"`
	//主播昵称
	Uname string `json:"uname"`
	//主播头像
	Uface string `json:"uface"`
	//主播uid
	Uid int64 `json:"uid"`
	//主播open_id
	OpenId string `json:"open_id"`
}

type EndAppRequest struct {
	// 场次id
	GameId string `json:"game_id"`
	// 项目id
	AppId int64 `json:"app_id"`
}

type AppHeartbeatReq struct {
	// 主播身份码
	GameId string `json:"game_id"`
}

func main() {
	// 加载用户音色配置
	log.Println("正在加载用户音色配置...")
	err := user_voice.LoadUserVoices()
	if err != nil {
		log.Printf("加载用户音色配置失败: %v", err)
		// 不返回，继续运行程序
	} else {
		log.Println("用户音色配置加载成功")
	}

	// 初始化TTS服务
	log.Println("正在初始化TTS服务...")
	err = tts_api.InitSimpleTTS()
	if err != nil {
		log.Printf("TTS服务初始化失败: %v", err)
		return
	}
	log.Println("TTS服务初始化成功")

	// 开启应用
	resp, err := StartApp(config.GetIdCode(), config.AppId)
	if err != nil {
		log.Printf("StartApp API调用失败: %v", err)
		return
	}

	log.Printf("StartApp API响应: Code=%d, Message=%s, Data=%s", resp.Code, resp.Message, string(resp.Data))

	// 检查API响应状态
	if resp.Code != 0 {
		log.Printf("StartApp API返回错误: Code=%d, Message=%s", resp.Code, resp.Message)
		return
	}

	// 检查Data是否为空
	if len(resp.Data) == 0 {
		log.Println("StartApp API返回的Data为空")
		return
	}

	// 解析返回值
	startAppRespData := &StartAppRespData{}
	err = json.Unmarshal(resp.Data, &startAppRespData)
	if err != nil {
		log.Printf("解析StartApp响应数据失败: %v, 原始数据: %s", err, string(resp.Data))
		return
	}

	if startAppRespData == nil {
		log.Println("start app get msg err")
		return
	}

	defer func() {
		// 保存用户音色配置
		log.Println("正在保存用户音色配置...")
		err = user_voice.SaveUserVoices()
		if err != nil {
			log.Printf("保存用户音色配置失败: %v", err)
		} else {
			log.Println("用户音色配置保存成功")
		}

		// 关闭应用
		_, err = EndApp(startAppRespData.GameInfo.GameId, config.AppId)
		if err != nil {
			fmt.Println(err)
			return
		}

		// 清理TTS服务
		log.Println("正在清理TTS服务...")
		tts_api.CloseSimpleTTS()
		log.Println("TTS服务已清理")
	}()

	if len(startAppRespData.WebsocketInfo.WssLink) == 0 {
		return
	}

	go func(gameId string) {
		for {
			time.Sleep(time.Second * 20)
			_, _ = AppHeart(gameId)
		}
	}(startAppRespData.GameInfo.GameId)

	// 开启长连
	err = net.StartWebsocket(startAppRespData.WebsocketInfo.WssLink[0], startAppRespData.WebsocketInfo.AuthBody)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 退出
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			log.Println("WebsocketClient exit")
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}

// StartApp 启动app
func StartApp(code string, appId int64) (resp net.BaseResp, err error) {
	startAppReq := StartAppRequest{
		Code:  code,
		AppId: appId,
	}
	reqJson, _ := json.Marshal(startAppReq)
	return net.ApiRequest(string(reqJson), "/v2/app/start")
}

// AppHeart app心跳
func AppHeart(gameId string) (resp net.BaseResp, err error) {
	appHeartbeatReq := AppHeartbeatReq{
		GameId: gameId,
	}
	reqJson, _ := json.Marshal(appHeartbeatReq)
	return net.ApiRequest(string(reqJson), "/v2/app/heartbeat")
}

// EndApp 关闭app
func EndApp(gameId string, appId int64) (resp net.BaseResp, err error) {
	endAppReq := EndAppRequest{
		GameId: gameId,
		AppId:  appId,
	}
	reqJson, _ := json.Marshal(endAppReq)
	return net.ApiRequest(string(reqJson), "/v2/app/end")
}
