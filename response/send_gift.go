package response

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
