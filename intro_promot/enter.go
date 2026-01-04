package intro_promot

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/CoffeeSwt/bilibili-tts-chat/config"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// GetEnterPromot 用户进入房间，生成引导词
func GetEnterPromot() string {
	strategies := []func() string{
		strategyIntro,
		strategyVoiceRecommendation,
		strategyAnnouncement,
	}

	// 随机选择一个策略
	selectedStrategy := strategies[rand.Intn(len(strategies))]
	return selectedStrategy()
}

// 策略1：基础介绍
func strategyIntro() string {
	prompts := []string{
		"你好呀，我是这里的弹幕播报员小七，你可以试试发送弹幕:换，可以切换我为你的播报音色。",
	}
	return prompts[rand.Intn(len(prompts))]
}

// 策略2：音色推荐
func strategyVoiceRecommendation() string {
	// 获取随机3个推荐音色
	voices := getRandomVoices(3)
	if len(voices) == 0 {
		return strategyIntro()
	}

	voiceNames := ""
	for i, v := range voices {
		if i > 0 {
			voiceNames += "、"
		}
		voiceNames += v.Name
	}

	prompts := []string{
		"你好，我是小七。如果不喜欢现在的声音，可以发送“换音色”试试哦。",
		fmt.Sprintf("你好呀，我是可以变声的小七，试试发送“换 %s”来改变我的声音吧。", voices[0].Name),
		fmt.Sprintf("你好呀，我有好多有趣的声音，比如%s，快来体验一下自定义音色吧。", voiceNames),
	}
	return prompts[rand.Intn(len(prompts))]
}

// 策略3：引导看公告
func strategyAnnouncement() string {
	prompts := []string{
		"你好，我是小七。想要更多个性化设置吗？看看直播间公告，了解如何自定义我的声音。",
		"你好！发送弹幕就能播报。记得查看直播间公告，解锁更多隐藏玩法哦。",
		"你好呀，我是小七。除了播报弹幕，我还有很多本领，详情请看直播间简介。",
	}
	return prompts[rand.Intn(len(prompts))]
}

// 辅助函数：获取随机不重复的音色
func getRandomVoices(count int) []config.Voice {
	allVoices := config.GetVoices()
	if len(allVoices) <= count {
		return allVoices
	}

	// 随机打乱并取前count个
	shuffled := make([]config.Voice, len(allVoices))
	copy(shuffled, allVoices)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	return shuffled[:count]
}
