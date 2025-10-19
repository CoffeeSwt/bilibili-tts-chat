package voice_engine

import (
	"fmt"
	"github.com/CoffeeSwt/bilibili-tts-chat/response"
	user_voice "github.com/CoffeeSwt/bilibili-tts-chat/user"
	"log"
	"math/rand"
	"sync"
	"time"
)

// 初始化随机种子
func init() {
	rand.Seed(time.Now().UnixNano())
}

// 用户连击点赞记录
var (
	userLikeCount = make(map[string]int)
	likeMutex     sync.RWMutex
)

// 基础点赞感谢语模板数组
var likeTemplates = []string{
	// 基础感谢类
	"感谢%s的点赞！",
	"谢谢%s的支持！",
	"感谢%s给的赞！",

	// 温馨夸赞类
	"感谢%s的点赞，你真是太棒了！",
	"谢谢%s的点赞，你的品味真好！",
	"感谢%s的点赞，你真是个小天才！",

	// 互动鼓励类
	"感谢%s的点赞，你的支持是我最大的动力！",
	"谢谢%s的点赞，有你真好！",
	"感谢%s的点赞，让我们一起努力！",

	// 可爱俏皮类
	"哇！%s给了个赞，爱你哟！",
	"嘻嘻，%s点赞了，么么哒！",
	"%s的小手点了个赞，好开心！",

	// 热情洋溢类
	"感谢%s的点赞，你的眼光真不错！",
	"谢谢%s的点赞，你真是太有眼光了！",
	"感谢%s的点赞，你的审美真棒！",

	// 幽默风趣类
	"感谢%s的点赞，你的手指真是太灵活了！",
	"谢谢%s的点赞，你的拇指真给力！",
	"感谢%s的点赞，你点赞的姿势真帅！",

	// 贴心关怀类
	"感谢%s的点赞，希望你今天心情美美哒！",
	"谢谢%s的点赞，愿你每天都开心！",
	"感谢%s的点赞，祝你天天好心情！",

	// 励志正能量类
	"感谢%s的点赞，让我们一起加油！",
	"谢谢%s的点赞，正能量满满！",
	"感谢%s的点赞，我们都要努力哦！",

	// 甜美温柔类
	"感谢%s的点赞，你就像小天使一样！",
	"谢谢%s的点赞，你真是个小可爱！",
	"感谢%s的点赞，你的心真善良！",

	// 活泼开朗类
	"感谢%s的点赞，你让直播间更有活力！",
	"谢谢%s的点赞，气氛瞬间活跃了！",
	"感谢%s的点赞，你真是开心果！",
}

// 连击点赞特殊感谢语
var comboLikeTemplates = []string{
	"哇！%s连续点赞，太给力了！",
	"%s的连击点赞，简直是点赞狂魔！",
	"感谢%s的疯狂点赞，你太热情了！",
	"%s连续点赞，手速真快！",
	"哇塞！%s的连击点赞，爱死你了！",
	"感谢%s的连续支持，你真是我的小粉丝！",
	"%s点赞不停，简直是点赞机器！",
	"哇！%s的连击，这节奏太棒了！",
}

// 超级连击感谢语（5次以上）
var superComboTemplates = []string{
	"天哪！%s的超级连击，你是点赞之王！",
	"哇！%s疯狂点赞，简直是点赞风暴！",
	"感谢%s的超级连击，你太疯狂了！",
	"%s的点赞连击，创造了新纪录！",
	"哇塞！%s的超级连击，你是点赞大神！",
}

// 获取用户连击次数并更新
func getUserLikeCombo(username string) int {
	likeMutex.Lock()
	defer likeMutex.Unlock()

	userLikeCount[username]++
	return userLikeCount[username]
}

// 重置用户连击次数（可以定期调用清理）
func resetUserLikeCombo(username string) {
	likeMutex.Lock()
	defer likeMutex.Unlock()

	delete(userLikeCount, username)
}

// 获取随机点赞感谢语
func getRandomLikeTemplate(username string, comboCount int) string {
	// 超级连击（5次以上）
	if comboCount >= 5 {
		return superComboTemplates[rand.Intn(len(superComboTemplates))]
	}

	// 连击（2-4次）
	if comboCount >= 2 {
		return comboLikeTemplates[rand.Intn(len(comboLikeTemplates))]
	}

	// 普通点赞
	return likeTemplates[rand.Intn(len(likeTemplates))]
}

// 添加语气词和表情描述
func addEmotionToText(text string) string {
	emotions := []string{
		"", // 有时候不加语气词
	}

	// 随机添加语气词
	if rand.Float32() < 0.6 { // 60%的概率添加语气词
		emotion := emotions[rand.Intn(len(emotions))]
		if emotion != "" {
			text = text + emotion
		}
	}

	return text
}

func PlayLikeVoice(msg response.LikeData) {
	// 获取用户连击次数
	comboCount := getUserLikeCombo(msg.UName)

	// 获取随机感谢语模板
	template := getRandomLikeTemplate(msg.UName, comboCount)

	// 生成基础感谢语
	voiceText := fmt.Sprintf(template, msg.UName)

	// 添加语气词和表情描述
	voiceText = addEmotionToText(voiceText)

	// 获取用户音色
	voiceID := user_voice.GetUserVoice(msg.UName)

	go func() {
		if err := TextToVoiceAsync(voiceText, voiceID); err != nil {
			log.Printf("[TTS] 点赞语音播放失败: %v, 内容: %s", err, voiceText)
		} else {
			log.Printf("[TTS] 点赞语音播放成功: %s (连击:%d)", voiceText, comboCount)
		}
	}()

	// 定期清理连击记录（避免内存泄漏）
	go func() {
		time.Sleep(30 * time.Second) // 30秒后清理该用户的连击记录
		resetUserLikeCombo(msg.UName)
	}()
}
