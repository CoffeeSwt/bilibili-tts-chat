package response

// 付费留言下线消息结构体 (LIVE_OPEN_PLATFORM_SUPER_CHAT_DEL)
type SuperChatDelMessage struct {
	Cmd  string           `json:"cmd"`  // 消息类型，固定为 "LIVE_OPEN_PLATFORM_SUPER_CHAT_DEL"
	Data SuperChatDelData `json:"data"` // 付费留言下线数据
}

// 付费留言下线数据结构体
type SuperChatDelData struct {
	RoomID     int    `json:"room_id"`     // 直播间id
	MessageIDs []int  `json:"message_ids"` // 留言id列表
	MsgID      string `json:"msg_id"`      // 消息唯一id
}
