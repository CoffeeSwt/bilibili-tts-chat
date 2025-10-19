package response

// 通用消息结构体
type LiveMessage struct {
	Cmd  string      `json:"cmd"`  // 消息类型
	Data interface{} `json:"data"` // 消息数据
}

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

// 礼物消息结构体 (LIVE_OPEN_PLATFORM_SEND_GIFT)
type GiftMessage struct {
	Cmd  string   `json:"cmd"`  // 消息类型，固定为 "LIVE_OPEN_PLATFORM_SEND_GIFT"
	Data GiftData `json:"data"` // 礼物数据
}

// 礼物数据结构体
type GiftData struct {
	RoomID                 int        `json:"room_id"`                   // 直播间(演播厅模式则为演播厅直播间,非演播厅模式则为收礼直播间)
	UID                    int        `json:"uid"`                       // 用户UID(已废弃，固定为0)
	OpenID                 string     `json:"open_id"`                   // 用户唯一标识
	UnionID                string     `json:"union_id"`                  // 用户在同一个开发者下的唯一标识
	UName                  string     `json:"uname"`                     // 送礼用户昵称
	UFace                  string     `json:"uface"`                     // 送礼用户头像
	GiftID                 int        `json:"gift_id"`                   // 道具id(盲盒:爆出道具id)
	GiftName               string     `json:"gift_name"`                 // 道具名(盲盒:爆出道具名)
	GiftNum                int        `json:"gift_num"`                  // 赠送道具数量
	Price                  int        `json:"price"`                     // 礼物单价(1000 = 1元 = 10电池),盲盒:爆出道具的价值
	Paid                   bool       `json:"paid"`                      // 是否是付费道具
	FansMedalLevel         int        `json:"fans_medal_level"`          // 对应房间勋章信息
	FansMedalName          string     `json:"fans_medal_name"`           // 粉丝勋章名
	FansMedalWearingStatus bool       `json:"fans_medal_wearing_status"` // 该房间粉丝勋章佩戴情况
	GuardLevel             int        `json:"guard_level"`               // room_id对应的大航海等级
	Timestamp              int64      `json:"timestamp"`                 // 收礼时间秒级时间戳
	MsgID                  string     `json:"msg_id"`                    // 消息唯一id
	AnchorInfo             AnchorInfo `json:"anchor_info"`               // 主播信息
	GiftIcon               string     `json:"gift_icon"`                 // 道具icon
	ComboGift              bool       `json:"combo_gift"`                // 是否是combo道具
	ComboInfo              ComboInfo  `json:"combo_info"`                // 连击信息
	BlindGift              BlindGift  `json:"blind_gift"`                // 盲盒信息
}

// 主播信息结构体
type AnchorInfo struct {
	UID     int    `json:"uid"`      // 收礼主播UID(即将废弃)
	OpenID  string `json:"open_id"`  // 主播唯一标识(2024-03-11后上线)
	UnionID string `json:"union_id"` // 用户在同一个开发者下的唯一标识
	UName   string `json:"uname"`    // 收礼主播昵称
	UFace   string `json:"uface"`    // 收礼主播头像
}

// 连击信息结构体
type ComboInfo struct {
	ComboBaseNum int    `json:"combo_base_num"` // 每次连击赠送的道具数量
	ComboCount   int    `json:"combo_count"`    // 连击次数
	ComboID      string `json:"combo_id"`       // 连击id
	ComboTimeout int    `json:"combo_timeout"`  // 连击有效期秒
}

// 盲盒信息结构体
type BlindGift struct {
	BlindGiftID int  `json:"blind_gift_id"` // 盲盒id
	Status      bool `json:"status"`        // 是否是盲盒
}

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
