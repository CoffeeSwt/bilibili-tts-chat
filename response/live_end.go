package response

// 结束直播消息结构体 (LIVE_OPEN_PLATFORM_LIVE_END)
type LiveEndMessage struct {
	Cmd  string      `json:"cmd"`  // 消息类型，固定为 "LIVE_OPEN_PLATFORM_LIVE_END"
	Data LiveEndData `json:"data"` // 结束直播数据
}

// 结束直播数据结构体
type LiveEndData struct {
	AreaName  string `json:"area_name"` // 直播分区名称
	OpenID    string `json:"open_id"`   // 用户唯一标识
	UnionID   string `json:"union_id"`  // 用户在同一个开发者下的唯一标识
	RoomID    int    `json:"room_id"`   // 直播间id
	Timestamp int64  `json:"timestamp"` // 结束直播时间戳
	Title     string `json:"title"`     // 直播标题
}
