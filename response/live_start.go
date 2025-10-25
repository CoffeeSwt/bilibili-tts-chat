package response

// 开始直播消息结构体 (LIVE_OPEN_PLATFORM_LIVE_START)
type LiveStartMessage struct {
	Cmd  string        `json:"cmd"`  // 消息类型，固定为 "LIVE_OPEN_PLATFORM_LIVE_START"
	Data LiveStartData `json:"data"` // 开始直播数据
}

// 开始直播数据结构体
type LiveStartData struct {
	AreaName  string `json:"area_name"` // 直播分区名称
	OpenID    string `json:"open_id"`   // 用户唯一标识
	UnionID   string `json:"union_id"`  // 用户在同一个开发者下的唯一标识
	RoomID    int    `json:"room_id"`   // 直播间id
	Timestamp int64  `json:"timestamp"` // 开始直播时间戳
	Title     string `json:"title"`     // 直播标题
}
