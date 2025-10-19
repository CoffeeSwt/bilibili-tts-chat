package voice_engine

import (
	"fmt"
	"github.com/CoffeeSwt/bilibili-tts-chat/response"
	user_voice "github.com/CoffeeSwt/bilibili-tts-chat/user"
	"log"
	"math/rand"
	"time"
)

func PlayUserEnterVoice(msg response.RoomEnterData) {
	voiceText, voiceID := GetUserEnterVoiceText(msg)
	// 异步调用TTS，避免阻塞弹幕处理流程
	go func() {
		if err := TextToVoiceAsync(voiceText, voiceID); err != nil {
			log.Printf("[TTS] 弹幕语音播放失败: %v, 内容: %s", err, voiceText)
		} else {
			log.Printf("[TTS] 弹幕语音播放成功: %s", voiceText)
		}
	}()
}

// 初始化随机种子
func init() {
	rand.Seed(time.Now().UnixNano())
}

// 欢迎语模板数组
var welcomeTemplates = []string{
	// 热情欢迎类
	"欢迎%s来到直播间！",
	"热烈欢迎%s的到来！",
	"欢迎%s加入我们！",

	// 幽默风趣类
	"哇！%s闪亮登场啦！",
	"看！%s华丽丽地出现了！",
	"哎呀！%s来了来了！",

	// 温馨问候类
	"%s你好呀，欢迎来玩！",
	"%s，很高兴见到你！",
	"嗨%s，欢迎回来！",

	// 活泼可爱类
	"%s小可爱来啦！",
	"%s宝贝进入直播间！",
	"萌萌的%s出现了！",

	// 正式礼貌类
	"尊贵的%s，欢迎光临！",
	"欢迎%s莅临直播间！",
	"感谢%s的到来！",

	// 互动引导类
	"%s快来和大家打个招呼吧！",
	"%s，大家都在等你呢！",
	"%s，来聊聊天吧！",

	// 特色创意类
	"叮咚！%s已上线！",
	"系统提示：%s进入房间！",
	"新朋友%s报到！",
}

// 获取时间问候语
func getTimeGreeting() string {
	hour := time.Now().Hour()
	switch {
	case hour >= 5 && hour < 12:
		return "早上好！"
	case hour >= 12 && hour < 18:
		return "下午好！"
	case hour >= 18 && hour < 23:
		return "晚上好！"
	default:
		return "夜深了，"
	}
}

// 获取随机欢迎语模板
func getRandomWelcomeTemplate() string {
	return welcomeTemplates[rand.Intn(len(welcomeTemplates))]
}

func GetUserEnterVoiceText(msg response.RoomEnterData) (voiceText string, voiceID string) {
	// 获取时间问候
	timeGreeting := getTimeGreeting()

	// 获取随机欢迎语模板

	welcomeTemplate := getRandomWelcomeTemplate()
	// 生成完整的欢迎语
	welcomeText := fmt.Sprintf(welcomeTemplate, msg.UName)

	// 组合时间问候和欢迎语
	if timeGreeting == "夜深了，" {
		voiceText = timeGreeting + welcomeText
	} else {
		voiceText = timeGreeting + welcomeText
	}

	// 获取用户音色
	voiceID = user_voice.GetUserVoice(msg.UName)

	return voiceText, voiceID
}
