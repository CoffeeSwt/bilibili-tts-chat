package voice_engine

import (
	"fmt"
	"github.com/CoffeeSwt/bilibili-tts-chat/response"
	user_voice "github.com/CoffeeSwt/bilibili-tts-chat/user"
	"log"
	"math/rand"
	"time"
)

// 初始化随机种子
func init() {
	rand.Seed(time.Now().UnixNano())
}

// 大航海等级名称映射
var guardLevelNames = map[int]string{
	1: "总督",
	2: "提督",
	3: "舰长",
}

// 舰长感谢语模板
var captainTemplates = []string{
	"感谢%s开通舰长！欢迎加入舰队，愿你一帆风顺！",
	"感谢%s成为舰长！你的支持是我前进的动力！",
	"感谢%s开通舰长！祝你事业有成，身体健康！",
	"感谢%s成为舰长！有你的陪伴真是太幸福了！",
	"感谢%s开通舰长！愿你每天都开开心心！",
	"感谢%s成为舰长！你就是我们舰队的骄傲！",
	"感谢%s开通舰长！祝你工作顺利，万事如意！",
	"感谢%s成为舰长！你的温暖让直播间更有爱！",
}

// 提督感谢语模板
var admiralTemplates = []string{
	"感谢%s开通提督！真是太豪气了，祝你财源滚滚！",
	"感谢%s成为提督！你就是我们的大英雄！",
	"感谢%s开通提督！愿你好运连连，心想事成！",
	"感谢%s成为提督！你的慷慨让人感动！",
	"感谢%s开通提督！祝你生意兴隆，财运亨通！",
	"感谢%s成为提督！你就是传说中的大佬！",
	"感谢%s开通提督！愿你身体健康，家庭幸福！",
	"感谢%s成为提督！你的支持让我充满力量！",
}

// 总督感谢语模板
var governorTemplates = []string{
	"感谢%s开通总督！哇，真正的大佬来了！",
	"感谢%s成为总督！你就是传说中的神豪！",
	"感谢%s开通总督！祝你万事如意，富贵满堂！",
	"感谢%s成为总督！简直是土豪中的土豪！",
	"感谢%s开通总督！你就是我们的守护神！",
	"感谢%s成为总督！祝你财源广进，福寿安康！",
	"感谢%s开通总督！你的豪气震撼全场！",
	"感谢%s成为总督！愿你永远幸福快乐！",
}

// 通用感谢语模板（当等级未知时使用）
var generalTemplates = []string{
	"感谢%s开通大航海！你真是太棒了！",
	"感谢%s的大航海支持！你就是我们的贵人！",
	"感谢%s开通大航海！祝你好运连连！",
	"感谢%s的大航海！你的支持让我感动！",
}

// 续费感谢语模板
var renewalTemplates = []string{
	"感谢%s续费%s！老朋友的支持最暖心了！",
	"感谢%s继续开通%s！有你真好！",
	"感谢%s再次开通%s！你就是我们的老铁！",
}

// 获取大航海等级名称
func getGuardLevelName(level int) string {
	if name, exists := guardLevelNames[level]; exists {
		return name
	}
	return "大航海"
}

// 获取随机感谢语模板
func getRandomGuardTemplate(level int, isRenewal bool) []string {
	// 如果是续费，优先使用续费模板
	if isRenewal {
		return renewalTemplates
	}

	switch level {
	case 3: // 舰长
		return captainTemplates
	case 2: // 提督
		return admiralTemplates
	case 1: // 总督
		return governorTemplates
	default:
		return generalTemplates
	}
}

// 检测是否为续费（简单逻辑，可以根据实际需求调整）
func isRenewalGuard(msg response.GuardData) bool {
	// 这里可以根据实际业务逻辑判断是否为续费
	// 比如检查用户历史记录、时间间隔等
	// 目前简单返回false，表示都是新开通
	return false
}

func PlayGuardVoice(msg response.GuardData) {
	voiceText := getGuardText(msg)
	// 获取用户音色
	voiceID := user_voice.GetUserVoice(msg.UserInfo.UName)

	go func() {
		if err := TextToVoiceAsync(voiceText, voiceID); err != nil {
			log.Printf("[TTS] 守护语音播放失败: %v, 内容: %s", err, voiceText)
		} else {
			log.Printf("[TTS] 守护语音播放成功: %s (等级:%d)", voiceText, msg.GuardLevel)
		}
	}()
}

// GetGuardText 导出函数用于测试
func GetGuardText(msg response.GuardData) string {
	return getGuardText(msg)
}

func getGuardText(msg response.GuardData) string {
	// 获取大航海等级名称
	levelName := getGuardLevelName(msg.GuardLevel)

	// 检测是否为续费
	isRenewal := isRenewalGuard(msg)

	// 获取对应的感谢语模板
	templates := getRandomGuardTemplate(msg.GuardLevel, isRenewal)

	// 随机选择一个模板
	template := templates[rand.Intn(len(templates))]

	// 生成感谢语
	var voiceText string
	if isRenewal {
		// 续费感谢语格式
		voiceText = fmt.Sprintf(template, msg.UserInfo.UName, levelName)
	} else {
		// 新开通感谢语格式
		voiceText = fmt.Sprintf(template, msg.UserInfo.UName)
	}

	return voiceText
}
