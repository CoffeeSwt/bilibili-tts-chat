package response

// 付费大航海消息结构体 (LIVE_OPEN_PLATFORM_GUARD)
type GuardMessage struct {
	Cmd  string    `json:"cmd"`  // 消息类型，固定为 "LIVE_OPEN_PLATFORM_GUARD"
	Data GuardData `json:"data"` // 付费大航海数据
}

// 付费大航海数据结构体
type GuardData struct {
	UserInfo               UserInfo `json:"user_info"`                 // 用户信息
	GuardLevel             int      `json:"guard_level"`               // 对应的大航海等级 1总督 2提督 3舰长
	GuardNum               int      `json:"guard_num"`                 // 大航海数量
	GuardUnit              string   `json:"guard_unit"`                // 大航海单位(正常单位为"月")
	Price                  int      `json:"price"`                     // 价格
	FansMedalLevel         int      `json:"fans_medal_level"`          // 粉丝勋章等级
	FansMedalName          string   `json:"fans_medal_name"`           // 粉丝勋章名
	FansMedalWearingStatus bool     `json:"fans_medal_wearing_status"` // 该房间粉丝勋章佩戴情况
	Timestamp              int64    `json:"timestamp"`                 // 时间戳
	RoomID                 int      `json:"room_id"`                   // 直播间id
	MsgID                  string   `json:"msg_id"`                    // 消息唯一id
}

// 用户信息结构体 (用于大航海消息)
type UserInfo struct {
	UID     int    `json:"uid"`      // 用户UID(已废弃，固定为0)
	OpenID  string `json:"open_id"`  // 用户唯一标识
	UnionID string `json:"union_id"` // 用户在同一个开发者下的唯一标识
	UName   string `json:"uname"`    // 用户昵称
	UFace   string `json:"uface"`    // 用户头像
}
