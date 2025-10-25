package response

// 消息推送结束通知消息结构体 (LIVE_OPEN_PLATFORM_INTERACTION_END)
type InteractionEndMessage struct {
	Cmd  string             `json:"cmd"`  // 消息类型，固定为 "LIVE_OPEN_PLATFORM_INTERACTION_END"
	Data InteractionEndData `json:"data"` // 消息推送结束通知数据
}

// 消息推送结束通知数据结构体
type InteractionEndData struct {
	GameID    string `json:"game_id"`   // 游戏id/应用id
	Timestamp int64  `json:"timestamp"` // 结束时间戳
}
