package response

// 点赞消息结构体 (LIVE_OPEN_PLATFORM_LIKE)
type LikeMessage struct {
	Cmd  string   `json:"cmd"`  // 消息类型，固定为 "LIVE_OPEN_PLATFORM_LIKE"
	Data LikeData `json:"data"` // 点赞数据
}

// 点赞数据结构体
type LikeData struct {
	UName                  string `json:"uname"`                     // 用户昵称
	UID                    int    `json:"uid"`                       // 用户UID(已废弃，固定为0)
	OpenID                 string `json:"open_id"`                   // 用户唯一标识
	UnionID                string `json:"union_id"`                  // 用户在同一个开发者下的唯一标识
	UFace                  string `json:"uface"`                     // 用户头像
	Timestamp              int64  `json:"timestamp"`                 // 时间戳
	LikeText               string `json:"like_text"`                 // 点赞文本，如"为主播点赞了"
	LikeCount              int    `json:"like_count"`                // 点赞数量
	FansMedalWearingStatus bool   `json:"fans_medal_wearing_status"` // 该房间粉丝勋章佩戴情况
	FansMedalName          string `json:"fans_medal_name"`           // 粉丝勋章名
	FansMedalLevel         int    `json:"fans_medal_level"`          // 粉丝勋章等级
	MsgID                  string `json:"msg_id"`                    // 消息唯一id
	RoomID                 int    `json:"room_id"`                   // 直播间id
}
