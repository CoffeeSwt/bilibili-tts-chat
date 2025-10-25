package response

// 付费留言消息结构体 (LIVE_OPEN_PLATFORM_SUPER_CHAT)
type SuperChatMessage struct {
	Cmd  string        `json:"cmd"`  // 消息类型，固定为 "LIVE_OPEN_PLATFORM_SUPER_CHAT"
	Data SuperChatData `json:"data"` // 付费留言数据
}

// 付费留言数据结构体
type SuperChatData struct {
	RoomID                 int    `json:"room_id"`                   // 直播间id
	UID                    int    `json:"uid"`                       // 购买用户UID(已废弃，固定为0)
	OpenID                 string `json:"open_id"`                   // 购买用户唯一标识
	UnionID                string `json:"union_id"`                  // 用户在同一个开发者下的唯一标识
	UName                  string `json:"uname"`                     // 购买的用户昵称
	UFace                  string `json:"uface"`                     // 购买用户头像
	MessageID              int    `json:"message_id"`                // 留言id(风控场景下撤回留言需要)
	Message                string `json:"message"`                   // 留言内容
	MsgID                  string `json:"msg_id"`                    // 消息唯一id
	RMB                    int    `json:"rmb"`                       // 支付金额(元)
	Timestamp              int64  `json:"timestamp"`                 // 赠送时间秒级
	StartTime              int64  `json:"start_time"`                // 生效开始时间
	EndTime                int64  `json:"end_time"`                  // 生效结束时间
	GuardLevel             int    `json:"guard_level"`               // 对应房间大航海等级
	FansMedalLevel         int    `json:"fans_medal_level"`          // 对应房间勋章信息
	FansMedalName          string `json:"fans_medal_name"`           // 对应房间勋章名字
	FansMedalWearingStatus bool   `json:"fans_medal_wearing_status"` // 该房间粉丝勋章佩戴情况
}
