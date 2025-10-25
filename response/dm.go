package response

// 弹幕消息结构体 (LIVE_OPEN_PLATFORM_DM)
type DanmakuMessage struct {
	Cmd  string      `json:"cmd"`  // 消息类型，固定为 "LIVE_OPEN_PLATFORM_DM"
	Data DanmakuData `json:"data"` // 弹幕数据
}

// 弹幕数据结构体
type DanmakuData struct {
	RoomID                 int    `json:"room_id"`                   // 弹幕接收的直播间
	UID                    int    `json:"uid"`                       // 用户UID(已废弃，固定为0)
	OpenID                 string `json:"open_id"`                   // 用户唯一标识
	UnionID                string `json:"union_id"`                  // 用户在同一个开发者下的唯一标识
	UName                  string `json:"uname"`                     // 用户昵称
	Msg                    string `json:"msg"`                       // 弹幕内容
	MsgID                  string `json:"msg_id"`                    // 消息唯一id
	FansMedalLevel         int    `json:"fans_medal_level"`          // 对应房间勋章信息
	FansMedalName          string `json:"fans_medal_name"`           // 粉丝勋章名
	FansMedalWearingStatus bool   `json:"fans_medal_wearing_status"` // 该房间粉丝勋章佩戴情况
	GuardLevel             int    `json:"guard_level"`               // 对应房间大航海 1总督 2提督 3舰长
	Timestamp              int64  `json:"timestamp"`                 // 弹幕发送时间秒级时间戳
	UFace                  string `json:"uface"`                     // 用户头像
	EmojiImgURL            string `json:"emoji_img_url"`             // 表情包图片地址
	DmType                 int    `json:"dm_type"`                   // 弹幕类型 0：普通弹幕 1：表情包弹幕
	GloryLevel             int    `json:"glory_level"`               // 直播荣耀等级
	ReplyOpenID            string `json:"reply_open_id"`             // 被at用户唯一标识
	ReplyUName             string `json:"reply_uname"`               // 被at的用户昵称
	IsAdmin                int    `json:"is_admin"`                  // 发送弹幕的用户是否为房管
}
