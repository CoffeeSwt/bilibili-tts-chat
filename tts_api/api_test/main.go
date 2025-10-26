package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/CoffeeSwt/bilibili-tts-chat/config"
	"github.com/CoffeeSwt/bilibili-tts-chat/logger"
	"github.com/CoffeeSwt/bilibili-tts-chat/tts_api"
	"github.com/CoffeeSwt/bilibili-tts-chat/voice"
)

func main() {
	wg := sync.WaitGroup{}
	for _, v := range config.GetVoices() {
		wg.Add(1)
		go func(v config.Voice) {
			defer wg.Done()
			logger.Info("开始生成语音", "voice", v.Name)
			ttsRes, err := GenerateSpeech(v)
			if err != nil {
				logger.Error("生成语音失败", "err", err)
				return
			}

			voiceErr := voice.PlayAudio(ttsRes.AudioData)
			if voiceErr != nil {
				logger.Error("播放语音失败", "err", voiceErr)
				return
			}

			filename := v.Name + "_" + time.Now().Format("20060102150405") + ".mp3"
			// 保存语音文件
			err = SaveSpeech(ttsRes, filename)
			if err != nil {
				logger.Error("保存语音文件失败", "err", err)
				return
			}
			logger.Info("保存语音文件成功", "filename", filename)
		}(v)
	}
	wg.Wait()
}

func GenerateSpeech(v config.Voice) (*tts_api.TTSResult, error) {
	ttsRes, err := tts_api.GenerateSpeech("你好，我是测试音色"+v.Name, &v)
	if err != nil {
		return nil, err
	}
	if ttsRes.DataSize <= 0 {
		return nil, fmt.Errorf("生成的语音数据大小为0")
	}
	return ttsRes, nil
}

func SaveSpeech(ttsRes *tts_api.TTSResult, filename string) error {
	if err := os.MkdirAll("output", 0755); err != nil {
		return err
	}
	if err := os.WriteFile("output/"+filename, []byte(ttsRes.AudioData), 0644); err != nil {
		return err
	}
	return nil
}
