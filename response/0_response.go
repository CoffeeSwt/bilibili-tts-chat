package response

// 通用消息结构体
type LiveMessage struct {
	Cmd  string `json:"cmd"`  // 消息类型
	Data any    `json:"data"` // 消息数据
}
