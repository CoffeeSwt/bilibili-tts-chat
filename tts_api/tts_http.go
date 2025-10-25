package tts_api

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/CoffeeSwt/bilibili-tts-chat/config"
)

var (
	ttsClient *TTSClient
	onceTTS   sync.Once
)

// TTSClient 字节跳动TTS HTTP客户端
type TTSClient struct {
	AppID      string
	AccessKey  string
	BaseURL    string
	HTTPClient *http.Client
}

// TTSConfig TTS配置
type TTSConfig struct {
	AppID     string `json:"app_id"`
	AccessKey string `json:"access_key"`
	BaseURL   string `json:"base_url"`
}

// AudioParams 音频参数
type AudioParams struct {
	Format          string `json:"format"`
	SampleRate      int    `json:"sample_rate"`
	EnableTimestamp bool   `json:"enable_timestamp"`
}

// ReqParams 请求参数
type ReqParams struct {
	Text        string      `json:"text"`
	Speaker     string      `json:"speaker"`
	AudioParams AudioParams `json:"audio_params"`
	Additions   string      `json:"additions"`
}

// User 用户信息
type User struct {
	UID string `json:"uid"`
}

// TTSRequest TTS请求结构
type TTSRequest struct {
	User      User      `json:"user"`
	ReqParams ReqParams `json:"req_params"`
}

// TTSResponse TTS响应结构
type TTSResponse struct {
	Code     int    `json:"code"`
	Data     string `json:"data,omitempty"`
	Sentence string `json:"sentence,omitempty"`
	Message  string `json:"message,omitempty"`
}

// TTSResult TTS结果
type TTSResult struct {
	AudioData []byte
	DataSize  int64
	LogID     string
}

// getTTSClient 获取TTS客户端单例
func getTTSClient() *TTSClient {
	onceTTS.Do(func() {
		ttsClient = &TTSClient{
			AppID:     config.GetTTSXApiAppID(),
			AccessKey: config.GetTTSXApiAccessKey(),
			BaseURL:   config.TTSHttpV3Host,
			HTTPClient: &http.Client{
				Timeout: 30 * time.Second,
			},
		}
	})
	return ttsClient
}

// GenerateSpeech 生成语音 - 对外暴露的唯一方法
// !!!注意这里没有做并发控制
func GenerateSpeech(text string, voice *config.Voice) (*TTSResult, error) {
	if voice == nil {
		return nil, fmt.Errorf("voice参数不能为空")
	}

	if text == "" {
		return nil, fmt.Errorf("text参数不能为空")
	}

	client := getTTSClient()

	// 构建请求参数
	request := TTSRequest{
		User: User{
			UID: config.GetRoomIDCode(),
		},
		ReqParams: ReqParams{
			Text:    text,
			Speaker: voice.VoiceType,
			AudioParams: AudioParams{
				Format:          "mp3",
				SampleRate:      24000,
				EnableTimestamp: true,
			},
			Additions: `{"explicit_language": "zh","disable_markdown_filter":true, "enable_timestamp":true}`,
		},
	}

	return client.processRequest(request, voice.ApiResourceID)
}

// processRequest 处理TTS请求
func (c *TTSClient) processRequest(request TTSRequest, resourceID string) (*TTSResult, error) {
	// 序列化请求体
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("序列化请求参数失败: %v", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", c.BaseURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	// 设置请求头
	headers := c.buildHeaders(resourceID)
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// 发送请求
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP请求失败，状态码: %d, 响应体: %s", resp.StatusCode, string(bodyBytes))
	}

	logID := resp.Header.Get("X-Tt-Logid")

	// 处理流式响应
	audioData, err := c.processStreamResponse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("处理流式响应失败: %v", err)
	}

	return &TTSResult{
		AudioData: audioData,
		DataSize:  int64(len(audioData)),
		LogID:     logID,
	}, nil
}

// buildHeaders 构建请求头
func (c *TTSClient) buildHeaders(resourceID string) map[string]string {
	return map[string]string{
		"X-Api-App-Id":      c.AppID,
		"X-Api-Access-Key":  c.AccessKey,
		"X-Api-Resource-Id": resourceID,
		"Content-Type":      "application/json",
		"Connection":        "keep-alive",
	}
}

// processStreamResponse 处理流式响应
func (c *TTSClient) processStreamResponse(body io.Reader) ([]byte, error) {
	var audioData []byte
	scanner := bufio.NewScanner(body)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var response TTSResponse
		err := json.Unmarshal([]byte(line), &response)
		if err != nil {
			continue // 跳过无法解析的行
		}

		// 处理音频数据
		if response.Code == 0 && response.Data != "" {
			chunkAudio, err := base64.StdEncoding.DecodeString(response.Data)
			if err != nil {
				continue // 跳过解码失败的数据
			}
			audioData = append(audioData, chunkAudio...)
			continue
		}

		// 处理结束标志
		if response.Code == 20000000 {
			break
		}

		// 处理错误
		if response.Code > 0 && response.Code != 20000000 {
			continue // 跳过错误数据
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取响应流失败: %v", err)
	}

	if len(audioData) == 0 {
		return nil, fmt.Errorf("未获取到音频数据")
	}

	return audioData, nil
}
