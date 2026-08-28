package rpc

import (
	"context"
	"fmt"
	"os"
	"time"
)

// MiniMax API 相关常量
const (
	MiniMaxBaseURL       = "https://api.minimaxi.com"
	MiniMaxTTSEndpoint   = "/v1/t2a_v2"
	MiniMaxImageEndpoint = "/v1/image_generation"
	MiniMaxCloneEndpoint = "/v1/voice_clone"

	EnvMiniMaxAPIKey = "MINIMAX_API_KEY"
)

// MiniMaxClient 是对 MiniMax API 的 RPC 客户端封装。
type MiniMaxClient struct {
	client *Client
}

// MiniMaxBaseResp 是 MiniMax API 通用响应包装。
type MiniMaxBaseResp struct {
	BaseResp struct {
		StatusCode    int    `json:"status_code"`
		StatusMessage string `json:"status_msg"`
	} `json:"base_resp"`
}

// IsSuccess 判断 MiniMax API 响应是否成功。
func (r *MiniMaxBaseResp) IsSuccess() bool {
	return r.BaseResp.StatusCode == 0
}

// NewMiniMaxClient 创建一个新的 MiniMax RPC 客户端。
func NewMiniMaxClient() *MiniMaxClient {
	return &MiniMaxClient{
		client: NewClient(WithTimeout(60 * time.Second)),
	}
}

// apiKey 获取 API Key。优先读取环境变量。
func (c *MiniMaxClient) apiKey() (string, error) {
	key := os.Getenv(EnvMiniMaxAPIKey)
	if key == "" {
		return "", fmt.Errorf("未配置 MiniMax API Key (环境变量 %s)", EnvMiniMaxAPIKey)
	}
	return key, nil
}

// TTSRequest 文字转语音请求参数。
type TTSRequest struct {
	Text    string `json:"text"`
	VoiceID string `json:"voiceId"`
}

// TTSResponse 文字转语音响应结果。
type TTSResponse struct {
	// Audio 返回原始音频字节（已解码）。
	Audio []byte
	// AudioFormat 识别到的音频格式扩展名（如 ".mp3", ".wav"）。
	AudioFormat string
}

// miniMaxTTSRawResponse MiniMax TTS API 原始响应。
type miniMaxTTSRawResponse struct {
	MiniMaxBaseResp
	Data struct {
		Audio string `json:"audio"` // hex 编码的音频
	} `json:"data"`
}

// TextToSpeech 调用 MiniMax T2A_v2 接口将文本转换为语音。
// 返回的音频数据为已解码的二进制。
func (c *MiniMaxClient) TextToSpeech(ctx context.Context, req TTSRequest) (*TTSResponse, error) {
	apiKey, err := c.apiKey()
	if err != nil {
		return nil, err
	}

	voiceID := req.VoiceID
	if voiceID == "" {
		voiceID = "male-qn-qingse"
	}

	body := map[string]interface{}{
		"model":  "speech-2.8-hd",
		"text":   req.Text,
		"stream": false,
		"voice_setting": map[string]interface{}{
			"voice_id":          voiceID,
			"speed":             1.0,
			"volume":            1.0,
			"pitch":             0,
			"emotion":           "happy",
			"use_speaker_boost": false,
		},
		"audio_setting": map[string]interface{}{
			"audio_type":  "mp3",
			"sample_rate": 32000,
			"bitrate":     128000,
		},
	}

	headers := map[string]string{
		"Authorization": "Bearer " + apiKey,
	}

	var rawResp miniMaxTTSRawResponse
	_, _, err = c.client.Request(ctx, "POST", MiniMaxBaseURL+MiniMaxTTSEndpoint, headers, body, &rawResp)
	if err != nil {
		return nil, err
	}

	if !rawResp.IsSuccess() {
		return nil, fmt.Errorf("MiniMax TTS 返回错误：code=%d, msg=%s", rawResp.BaseResp.StatusCode, rawResp.BaseResp.StatusMessage)
	}

	if rawResp.Data.Audio == "" {
		return nil, fmt.Errorf("MiniMax TTS 返回音频为空")
	}

	// 解码音频数据（MiniMax 返回 hex 编码）
	audioBytes, err := decodeHex(rawResp.Data.Audio)
	if err != nil {
		return nil, fmt.Errorf("音频解码失败：%w", err)
	}

	return &TTSResponse{
		Audio:       audioBytes,
		AudioFormat: detectAudioFormat(audioBytes),
	}, nil
}

// ImageRequest 图片生成请求参数。
type ImageRequest struct {
	Text               string `json:"text"`
	NumImages          int    `json:"numImages"`
	Width              int    `json:"width"`
	Height             int    `json:"height"`
	ReferenceImagePath string `json:"referenceImagePath"`
}

// ImageResponse 图片生成响应。
type ImageResponse struct {
	ImageURLs []string `json:"image_urls"`
}

// miniMaxImageRawResponse MiniMax 图片生成 API 原始响应。
type miniMaxImageRawResponse struct {
	MiniMaxBaseResp
	Data struct {
		ImageURLs []string `json:"image_urls"`
	} `json:"data"`
}

// GenerateImage 调用 MiniMax 图片生成接口，返回生成的图片 URL 列表。
func (c *MiniMaxClient) GenerateImage(ctx context.Context, req ImageRequest) (*ImageResponse, error) {
	apiKey, err := c.apiKey()
	if err != nil {
		return nil, err
	}

	// 默认参数归一化
	numImages := req.NumImages
	if numImages <= 0 || numImages > 9 {
		numImages = 9
	}
	width := req.Width
	if width < 512 || width > 2048 || width%8 != 0 {
		width = 1024
	}
	height := req.Height
	if height < 512 || height > 2048 || height%8 != 0 {
		height = 1024
	}

	body := map[string]interface{}{
		"model":           "image-01",
		"prompt":          req.Text,
		"n":               numImages,
		"width":           width,
		"height":          height,
		"response_format": "url",
	}

	// 可选参考图
	if req.ReferenceImagePath != "" {
		dataURL, err := encodeImageAsDataURL(req.ReferenceImagePath)
		if err != nil {
			return nil, fmt.Errorf("读取参考图失败：%w", err)
		}
		body["subject_reference"] = []map[string]interface{}{
			{
				"type":       "character",
				"image_file": dataURL,
			},
		}
	}

	headers := map[string]string{
		"Authorization": "Bearer " + apiKey,
	}

	var rawResp miniMaxImageRawResponse
	_, _, err = c.client.Request(ctx, "POST", MiniMaxBaseURL+MiniMaxImageEndpoint, headers, body, &rawResp)
	if err != nil {
		return nil, err
	}

	if !rawResp.IsSuccess() {
		return nil, fmt.Errorf("MiniMax Image 返回错误：code=%d, msg=%s", rawResp.BaseResp.StatusCode, rawResp.BaseResp.StatusMessage)
	}

	if len(rawResp.Data.ImageURLs) == 0 || rawResp.Data.ImageURLs[0] == "" {
		return nil, fmt.Errorf("MiniMax Image 返回图片 URL 为空")
	}

	return &ImageResponse{
		ImageURLs: rawResp.Data.ImageURLs,
	}, nil
}

// DownloadImageBytes 下载图片 URL，返回原始字节。
func (c *MiniMaxClient) DownloadImageBytes(ctx context.Context, url string) ([]byte, error) {
	return c.client.GetBytes(ctx, url, nil)
}

// CloneRequest 音色克隆请求参数。
type CloneRequest struct {
	SourceFileID    string `json:"source_file_id"`
	ReferenceText   string `json:"reference_text"`
	ReferenceFileID string `json:"reference_file_id"`
	VoiceID         string `json:"voice_id"`
}

// CloneResponse 音色克隆响应。
type CloneResponse struct{}

// CloneVoice 调用 MiniMax 音色克隆接口。
func (c *MiniMaxClient) CloneVoice(ctx context.Context, req CloneRequest) (*CloneResponse, error) {
	apiKey, err := c.apiKey()
	if err != nil {
		return nil, err
	}

	if req.SourceFileID == "" {
		return nil, fmt.Errorf("请先上传待克隆音频")
	}
	if req.VoiceID == "" {
		return nil, fmt.Errorf("请输入自定义音色ID")
	}

	body := map[string]interface{}{
		"model": "speech-01",
		"input": map[string]interface{}{
			"source": req.SourceFileID,
		},
		"voice_id": req.VoiceID,
	}
	if req.ReferenceText != "" {
		body["reference_text"] = req.ReferenceText
	}
	if req.ReferenceFileID != "" {
		body["reference"] = req.ReferenceFileID
	}

	headers := map[string]string{
		"Authorization": "Bearer " + apiKey,
	}

	var rawResp MiniMaxBaseResp
	_, _, err = c.client.Request(ctx, "POST", MiniMaxBaseURL+MiniMaxCloneEndpoint, headers, body, &rawResp)
	if err != nil {
		return nil, err
	}

	if !rawResp.IsSuccess() {
		return nil, fmt.Errorf("MiniMax CloneVoice 返回错误：code=%d, msg=%s", rawResp.BaseResp.StatusCode, rawResp.BaseResp.StatusMessage)
	}

	return &CloneResponse{}, nil
}