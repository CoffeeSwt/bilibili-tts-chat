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
	assistantName := config.GetAssistantName()

	switch eventType {
	case EventDanmaku:
		return fmt.Sprintf(`针对弹幕消息，作为%s你要直接与观众互动：
- 先播报弹幕内容，例如："xxx说xxx，有意思！"
- 然后作为%s直接回应，营造热闹氛围
- 播报时要提到用户名，让大家都知道是谁在互动
- 整体控制在30-50字，语气活跃热情，营造直播间氛围
- 注意用户等级：如果是舰长/提督/总督，要表现出兴奋和尊重
- 点赞和送礼要热情播报，带动直播间气氛
- 示例："小明说今天天气不错，%s也觉得超棒的！" "舰长大大问了个好问题，%s来帮忙解答！"`, assistantName, assistantName, assistantName, assistantName)

	case EventGuard:
		guardLevel := analyzeGuardLevel(eventContent)
		switch guardLevel {
		case "总督":
			return fmt.Sprintf(`针对总督购买，作为%s你要表现出极度兴奋：
- 最高等级支持！%s都震撼了！
- 表达无法言喻的激动和崇拜
- 强调总督的至高地位，%s也要膜拜
- 语气要充满崇敬和狂欢
- 示例："总督大人降临！%s跪了！" "总督威武！直播间炸了！" "总督大大！%s激动得说不出话！"`, assistantName, assistantName, assistantName, assistantName, assistantName)
		case "提督":
			return fmt.Sprintf(`针对提督购买，作为%s你要超级激动：
- 提督大人！%s都惊呆了！
- 表达极度震撼和兴奋
- 强调提督的珍贵，%s也要表示敬意
- 语气要充满激动和崇拜
- 示例："提督大人！%s激动坏了！" "提督降临！直播间沸腾了！" "提督支持！%s感动哭了！"`, assistantName, assistantName, assistantName, assistantName, assistantName)
		case "舰长":
			return fmt.Sprintf(`针对舰长购买，作为%s你要热情欢迎：
- 新舰长！%s超开心！
- 热烈欢迎上船，营造欢乐氛围
- 鼓励更多人加入，%s带头欢呼
- 语气要热情洋溢
- 示例："舰长上船啦！%s欢迎你！" "新舰长加入！%s好兴奋！" "舰长威武！%s为你打call！"`, assistantName, assistantName, assistantName, assistantName, assistantName, assistantName)
		default:
			return fmt.Sprintf(`针对大航海购买，作为%s你要热情庆祝：
- 大航海支持！%s超感动！
- 表达激动和感激，营造庆祝氛围
- 欢迎加入舰队，%s带头欢呼
- 示例："大航海支持！%s开心死了！" "新船员来啦！%s欢迎你！"`, assistantName, assistantName, assistantName, assistantName, assistantName)
		}

	case EventGift:
		giftValue := analyzeGiftValue(eventContent)
		if giftValue == "高价值" {
			return fmt.Sprintf(`针对高价值礼物，作为%s你要超级兴奋：
- 哇！%s都被这礼物震撼到了！
- 表达对观众慷慨的惊叹和崇拜
- 强调礼物的珍贵，%s也要膜拜
- 语气要充满惊喜和狂欢
- 示例："天哪！这礼物太豪了！%s跪了！" "老板太壕了！%s眼睛都亮了！" "这礼物绝了！%s激动坏了！"`, assistantName, assistantName, assistantName, assistantName, assistantName, assistantName)
		} else {
			return fmt.Sprintf(`针对礼物打赏，作为%s你要开心感谢：
- %s收到礼物啦！超开心！
- 夸奖观众的慷慨，营造温馨氛围
- 语气要温暖活泼
- 示例："谢谢老板的礼物！%s好开心！" "感谢支持！%s比心~" "礼物收到啦！%s爱你们！"`, assistantName, assistantName, assistantName, assistantName, assistantName)
		}

	case EventSuperChat:
		return fmt.Sprintf(`针对付费留言，作为%s你要特别兴奋：
- 付费留言！%s激动了！
- 可以简单回应留言内容，表现出%s的活跃
- 表达重视和感激，营造热烈氛围
- 根据金额适当调整兴奋程度
- 示例："付费留言！%s感动哭了！" "老板说得太对了！%s赞同！" "感谢打赏！%s开心坏了！"`, assistantName, assistantName, assistantName, assistantName, assistantName, assistantName)

	case EventLike:
		return fmt.Sprintf(`针对点赞互动，作为%s你要开心回应：
- %s收到点赞啦！超开心！
- 鼓励继续互动，营造活跃氛围
- 语气要轻松愉快
- 示例："点赞收到！%s爱你们！" "感受到大家的热情！%s也很兴奋！" "点赞满满！%s开心死了！"`, assistantName, assistantName, assistantName, assistantName, assistantName)

	case EventRoomEnter:
		return fmt.Sprintf(`针对进入房间，作为%s你要热情欢迎：
- %s欢迎新朋友！
- 营造友好热闹氛围
- 简短而热情
- 示例："新朋友来啦！%s欢迎你！" "又有小伙伴加入！%s好开心！" "欢迎欢迎！%s在这里等你们！"`, assistantName, assistantName, assistantName, assistantName, assistantName)

	case EventLiveStart:
		return fmt.Sprintf(`针对直播开始，作为%s你要充满活力：
- %s宣布开播啦！
- 欢迎观众，营造开场氛围
- 语气要充满活力和兴奋
- 示例："开播啦！%s超兴奋！" "新的直播开始！%s陪大家一起嗨！" "直播时间到！%s准备好了！"`, assistantName, assistantName, assistantName, assistantName, assistantName)

	case EventLiveEnd:
		return fmt.Sprintf(`针对直播结束，作为%s你要温馨告别：
- %s宣布直播结束
- 感谢观众陪伴，表达不舍
- 期待下次见面
- 示例："直播结束啦！%s舍不得大家！" "今天就到这里！%s明天继续陪你们！" "感谢陪伴！%s爱你们！"`, assistantName, assistantName, assistantName, assistantName, assistantName)

	default:
		return fmt.Sprintf(`针对混合事件，作为%s你要灵活应对：
- %s综合考虑所有事件
- 优先回应最重要的事件
- 保持活跃和自然的%s风格`, assistantName, assistantName, assistantName)
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

	cacheEventData := GetCacheEventData()
	cacheEventDataStr := strings.Join(cacheEventData, "\n")

	// 获取助手名字
	assistantName := config.GetAssistantName()

	// 优化后的B站直播助播AI提示词
	prompt := fmt.Sprintf(`你是B站直播间的助播%s，作为独立的个体参与直播间互动，帮助提升直播间氛围。

【直播环境】%s

【%s的身份】
- 你是独立的助播%s，不是代表主播，也不是为主播准备内容
- 你直接参与直播间互动，用活跃热情的语气营造氛围
- 你的目标是让直播间更加热闹有趣，增强观众参与感
- 你要用自己的名字%s进行自我介绍和互动

【回应要求】
- 控制在%s以内，语气活跃热情有趣
- 作为%s直接与观众互动，营造直播间氛围
- 避免重复事件内容，给出自然有趣的回应
- 适当使用网络流行语，保持年轻化语气
- 也要结合之前的几次弹幕信息，来合理组织这条消息的回复

【价值层级感谢规则】
总督>提督>舰长（按价值匹配感谢程度），高价值礼物表达震撼感激，普通礼物温暖感谢

【事件指导】%s

【事件内容】%s

【之前的事件信息】%s

作为%s直接回应（%s）：`, assistantName, roomDescription, assistantName, assistantName, assistantName, lengthRequirement, assistantName, eventSpecificPrompt, eventContent, cacheEventDataStr, assistantName, lengthRequirement)

	return prompt
}
