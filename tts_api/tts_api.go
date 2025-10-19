package tts_api

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/CoffeeSwt/bilibili-tts-chat/config"
	"strings"

	"github.com/google/uuid"
	"github.com/monaco-io/request"
)

// v3 API 请求结构体定义
type TTSV3Request struct {
	User      UserInfo   `json:"user"`
	Namespace string     `json:"namespace"`
	ReqParams ReqParams  `json:"req_params"`
}

type UserInfo struct {
	UID string `json:"uid"`
}

type ReqParams struct {
	Text        string      `json:"text"`
	Speaker     string      `json:"speaker"`
	AudioParams AudioParams `json:"audio_params"`
}

type AudioParams struct {
	Format     string `json:"format"`
	SampleRate int    `json:"sample_rate"`
	BitRate    int    `json:"bit_rate,omitempty"`
}

// v3 API 响应结构体定义
type TTSV3Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

// TTS 函数实现 - v3 API 流式处理
func TTS(text string) (string, error) {
	// 生成唯一的请求ID
	reqID := uuid.New().String()

	// 构建v3 API请求体
	ttsReq := TTSV3Request{
		User: UserInfo{
			UID: config.TTSUID,
		},
		Namespace: config.TTSNamespace,
		ReqParams: ReqParams{
			Text:    text,
			Speaker: config.DefaultVoice, // 使用配置中的默认音色
			AudioParams: AudioParams{
				Format:     config.DefaultEncoding,
				SampleRate: 24000,
				BitRate:    64000,
			},
		},
	}

	// 发送HTTP POST请求 - v3 API使用新的请求头
	client := request.Client{
		URL:    config.TTSAPIUrl,
		Method: "POST",
		Header: map[string]string{
			"Content-Type":       "application/json",
			"X-Api-App-Id":       config.GetTTSAppID(),
		"X-Api-Access-Key":   config.GetTTSAccessKey(),
			"X-Api-Resource-Id":  config.TTSResourceID,
			"X-Api-Request-Id":   reqID,
		},
		JSON: ttsReq,
	}

	resp := client.Send()

	// 检查请求是否成功
	if !resp.OK() {
		return "", fmt.Errorf("HTTP request failed: %v", resp.Error())
	}

	// v3 API返回流式数据，需要逐行解析
	var audioDataParts []string
	responseBody := resp.String()
	
	// 按行分割响应
	scanner := bufio.NewScanner(strings.NewReader(responseBody))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// 解析每行JSON
		var ttsResp TTSV3Response
		if err := json.Unmarshal([]byte(line), &ttsResp); err != nil {
			// 跳过无法解析的行
			continue
		}

		// 检查错误状态码
		if ttsResp.Code != 0 && ttsResp.Code != 20000000 {
			return "", fmt.Errorf("TTS API error: code=%d, message=%s", ttsResp.Code, ttsResp.Message)
		}

		// 收集音频数据
		if ttsResp.Data != "" {
			audioDataParts = append(audioDataParts, ttsResp.Data)
		}

		// 检查是否为结束标志
		if ttsResp.Code == 20000000 {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	// 拼接所有音频数据
	if len(audioDataParts) == 0 {
		return "", fmt.Errorf("no audio data returned from TTS API")
	}

	// 将所有base64音频数据拼接
	fullAudioData := strings.Join(audioDataParts, "")
	
	return fullAudioData, nil
}
