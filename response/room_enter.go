package response

// 进入房间消息结构体 (LIVE_OPEN_PLATFORM_LIVE_ROOM_ENTER)
type RoomEnterMessage struct {
	Cmd  string        `json:"cmd"`  // 消息类型，固定为 "LIVE_OPEN_PLATFORM_LIVE_ROOM_ENTER"
	Data RoomEnterData `json:"data"` // 进入房间数据
}

// 进入房间数据结构体
type RoomEnterData struct {
	RoomID    int    `json:"room_id"`   // 直播间id
	UFace     string `json:"uface"`     // 用户头像
	UName     string `json:"uname"`     // 用户昵称
	OpenID    string `json:"open_id"`   // 用户唯一标识
	UnionID   string `json:"union_id"`  // 用户在同一个开发者下的唯一标识
	Timestamp int64  `json:"timestamp"` // 时间戳
}
