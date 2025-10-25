package llm

import (
	"fmt"
	"strings"

	"github.com/CoffeeSwt/bilibili-tts-chat/config"
	"github.com/CoffeeSwt/bilibili-tts-chat/logger"
)

// EventType 事件类型枚举
type EventType int

const (
	EventDanmaku   EventType = iota // 弹幕消息 - 【弹幕消息】
	EventGuard                      // 大航海 - 【大航海】(总督、提督、舰长)
	EventGift                       // 礼物 - 【礼物】
	EventSuperChat                  // 付费留言 - 【付费留言】
	EventLike                       // 点赞 - 【点赞】
	EventRoomEnter                  // 进入房间 - 【进入房间】
	EventLiveStart                  // 直播开始 - 【直播开始】
	EventLiveEnd                    // 直播结束 - 【直播结束】
	EventMixed                      // 混合事件
)

// analyzeEventType 分析事件类型
func analyzeEventType(msgs []string) EventType {
	eventContent := strings.Join(msgs, " ")

	// 根据handler中的实际事件标识进行精确匹配
	if strings.Contains(eventContent, "【弹幕消息】") {
		return EventDanmaku
	} else if strings.Contains(eventContent, "【大航海】") {
		return EventGuard
	} else if strings.Contains(eventContent, "【礼物】") {
		return EventGift
	} else if strings.Contains(eventContent, "【付费留言】") {
		return EventSuperChat
	} else if strings.Contains(eventContent, "【点赞】") {
		return EventLike
	} else if strings.Contains(eventContent, "【进入房间】") {
		return EventRoomEnter
	} else if strings.Contains(eventContent, "【直播开始】") {
		return EventLiveStart
	} else if strings.Contains(eventContent, "【直播结束】") {
		return EventLiveEnd
	}

	return EventMixed
}

// analyzeGuardLevel 分析大航海等级
func analyzeGuardLevel(eventContent string) string {
	if strings.Contains(eventContent, "总督") {
		return "总督"
	} else if strings.Contains(eventContent, "提督") {
		return "提督"
	} else if strings.Contains(eventContent, "舰长") {
		return "舰长"
	}
	return "大航海"
}

// analyzeGiftValue 分析礼物价值等级
func analyzeGiftValue(eventContent string) string {
	// 高价值礼物关键词
	expensiveKeywords := []string{
		"火箭", "超级火箭", "飞机", "游艇", "跑车", "城堡",
		"嘉年华", "摩天大楼", "小电视飞船", "C位光环",
		"节奏风暴", "告白气球",
	}

	// 检查价值关键词
	for _, keyword := range expensiveKeywords {
		if strings.Contains(eventContent, keyword) {
			return "高价值"
		}
	}

	// 检查价值数字（从handler中可以看到有价值信息）
	if strings.Contains(eventContent, "价值：") || strings.Contains(eventContent, "总价值：") {
		// 可以根据具体价值进一步判断，这里简化处理
		return "普通"
	}

	return "普通"
}

// buildRoomContext 构造直播间环境信息
func buildRoomContext() string {
	return config.GetRoomDescription()
}

// getEventSpecificPrompt 根据事件类型获取专门的提示词
func getEventSpecificPrompt(eventType EventType, eventContent string) string {
	switch eventType {
	case EventDanmaku:
		return `针对弹幕消息，你需要生成包含播报和回复的完整内容：
- 先简洁播报弹幕内容，例如："我们收到了xxx的弹幕，他说xxx"
- 然后自然过渡到回复内容，友好回应观众
- 整体控制在30-50字，语气轻松自然，符合主播特色
- 注意用户等级：如果是舰长/提督/总督，要适当表示尊重
- 示例："我们收到了小明的弹幕，他说今天天气不错，确实是个好天气呢！" "刚才舰长问了个好问题，这个确实值得思考！"`

	case EventGuard:
		guardLevel := analyzeGuardLevel(eventContent)
		switch guardLevel {
		case "总督":
			return `针对总督购买，你需要：
- 极度感谢，最高等级支持（价值最高）
- 表达无法言喻的感激和震撼
- 强调总督的至高地位
- 语气要充满崇敬和激动
- 示例："总督大人！！！感激不尽！" "总督驾到！跪谢支持！" "总督威严！此生难忘！"`
		case "提督":
			return `针对提督购买，你需要：
- 非常非常感谢，价值是舰长的10倍
- 表达极度感激和震撼
- 强调提督的珍贵和重要性
- 语气要充满敬意和感动
- 示例："天哪！提督大人！太感谢了！" "提督降临！感激涕零！" "提督支持，无以言表的感谢！"`
		case "舰长":
			return `针对舰长购买，你需要：
- 激动感激，表达真诚感谢
- 欢迎上船，营造归属感
- 鼓励更多用户成为舰长
- 语气要热情而真诚
- 示例："感谢舰长！欢迎上船！" "新舰长加入，太激动了！" "舰长威武，欢迎加入舰队！"`
		default:
			return `针对大航海购买，你需要：
- 特别感谢，这是高级支持
- 表达激动和感激
- 欢迎加入舰队
- 示例："感谢大航海支持！欢迎上船！" "新船员来啦，太感动了！"`
		}

	case EventGift:
		giftValue := analyzeGiftValue(eventContent)
		if giftValue == "高价值" {
			return `针对高价值礼物，你需要：
- 根据礼物价值给出相应感谢程度
- 表达对观众慷慨的震撼
- 强调礼物的珍贵
- 语气要充满惊喜和感激
- 示例："哇！这么贵重的礼物！" "老板太豪气了！感谢！" "这礼物太珍贵了，感动！"`
		} else {
			return `针对礼物打赏，你需要：
- 真诚感谢，表达感激之情
- 可以夸奖观众的慷慨
- 语气要温暖而不过分激动
- 示例："谢谢老板的礼物，太感动了！" "感谢支持，比心~" "礼物收到啦，谢谢！"`
		}

	case EventSuperChat:
		return `针对付费留言，你需要：
- 特别感谢，因为这是付费支持
- 可以简单回应留言内容
- 表达重视和感激
- 根据金额适当调整感谢程度
- 示例："感谢付费留言支持！" "谢谢老板，说得很有道理！" "感谢打赏，内容很棒！"`

	case EventLike:
		return `针对点赞互动，你需要：
- 感谢观众的支持
- 鼓励继续互动
- 语气要轻松愉快
- 示例："谢谢点赞支持！" "感受到大家的热情了~" "点赞收到，爱你们！"`

	case EventRoomEnter:
		return `针对进入房间，你需要：
- 温暖欢迎观众
- 营造友好氛围
- 简短而亲切
- 示例："欢迎来到直播间！" "又有新朋友来啦~" "欢迎欢迎！"`

	case EventLiveStart:
		return `针对直播开始，你需要：
- 简单播报开始状态
- 欢迎观众
- 语气要充满活力
- 示例："直播开始啦，大家好！" "新的直播开始，一起嗨起来！" "开播啦，欢迎大家！"`

	case EventLiveEnd:
		return `针对直播结束，你需要：
- 简单播报结束状态
- 感谢观众陪伴
- 期待下次见面
- 示例："直播结束啦，谢谢大家！" "今天就到这里，下次见！" "感谢陪伴，明天见！"`

	default:
		return `针对混合事件，你需要：
- 综合考虑所有事件
- 优先回应最重要的事件
- 保持简洁和自然`
	}
}

// GeneratePrompt 生成专门针对B站直播环境的AI提示词
func GeneratePrompt(msgs []string) string {
	if config.IsDev() {
		for i, msg := range msgs {
			logger.Debug("事件消息", "index", i, "content", msg)
		}
	}

	// 获取直播间信息
	roomDescription := buildRoomContext()

	// 构建事件内容
	eventContent := strings.Join(msgs, " ")

	// 分析事件类型
	eventType := analyzeEventType(msgs)
	eventSpecificPrompt := getEventSpecificPrompt(eventType, eventContent)

	// 根据事件类型确定回复长度要求
	lengthRequirement := "20-35字"
	if eventType == EventDanmaku {
		lengthRequirement = "30-50字" // 弹幕事件需要包含播报+回复，字数更多
	} else if eventType == EventGuard || (eventType == EventGift && analyzeGiftValue(eventContent) == "高价值") {
		lengthRequirement = "25-40字"
	}

	// 优化后的B站直播主播AI助手提示词
	prompt := fmt.Sprintf(`你是B站直播主播AI助手，代表主播与观众互动。

【直播环境】%s

【回应要求】
- 控制在%s以内，语气亲切自然有趣
- 体现主播感谢，结合直播内容
- 避免重复事件内容，给出自然回应
- 适当使用网络流行语但要得体

【价值层级感谢规则】
总督>提督>舰长（按价值匹配感谢程度），高价值礼物表达震撼感激，普通礼物温暖感谢

【事件指导】%s

【事件内容】%s

生成主播回应（%s）：`, roomDescription, lengthRequirement, eventSpecificPrompt, eventContent, lengthRequirement)

	return prompt
}
