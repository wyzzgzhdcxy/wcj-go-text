package app

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"wcj-go-text/golang/myTTS"

	"github.com/wyzzgzhdcxy/wcj-go-common/core"
)

// GenerateImageReq 图片生成请求
type GenerateImageReq struct {
	Text               string `json:"text"`
	ApiKey             string `json:"apiKey"`
	NumImages          int    `json:"numImages"`
	Width              int    `json:"width"`
	Height             int    `json:"height"`
	ReferenceImagePath string `json:"referenceImagePath"`
}

// GenerateImageRes 图片生成结果
type GenerateImageRes struct {
	Success    bool     `json:"success"`
	Message    string   `json:"message"`
	OutputPath []string `json:"outputPath"`
	Cost       string   `json:"cost"`
}

// GenerateImage 使用 MiniMax API 生成图片
func (a *App) GenerateImage(req GenerateImageReq) GenerateImageRes {
	apiKey := req.ApiKey
	if apiKey == "" {
		apiKey = a.GetConfigKey("minimax_api_key")
	}

	imgReq := myTTS.MiniMaxImageReq{
		Text:               req.Text,
		ApiKey:             apiKey,
		NumImages:          req.NumImages,
		Width:              req.Width,
		Height:             req.Height,
		ReferenceImagePath: req.ReferenceImagePath,
	}

	result := myTTS.GenerateImage(imgReq)
	return GenerateImageRes{
		Success:    result.Success,
		Message:    result.Message,
		OutputPath: result.OutputPath,
		Cost:       result.Cost,
	}
}

// TextToSpeechReq 文字转语音请求
type TextToSpeechReq struct {
	Text    string `json:"text"`
	ApiKey  string `json:"apiKey"`
	VoiceID string `json:"voiceId"`
}

// TextToSpeechRes 文字转语音结果
type TextToSpeechRes struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	OutputPath string `json:"outputPath"`
	Cost       string `json:"cost"`
}

// TextToSpeech 使用 MiniMax TTS 将文本转换为语音
func (a *App) TextToSpeech(req TextToSpeechReq) TextToSpeechRes {
	apiKey := req.ApiKey
	if apiKey == "" {
		apiKey = a.GetConfigKey("minimax_api_key")
	}

	ttsReq := myTTS.MiniMaxTTsReq{
		Text:    req.Text,
		ApiKey:  apiKey,
		VoiceID: req.VoiceID,
	}

	result := myTTS.TextToSpeech(ttsReq)
	return TextToSpeechRes{
		Success:    result.Success,
		Message:    result.Message,
		OutputPath: result.OutputPath,
		Cost:       result.Cost,
	}
}

// GetAvailableVoices 获取可用的音色列表
func (a *App) GetAvailableVoices() []struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Language string `json:"language"`
} {
	return myTTS.GetAvailableVoices()
}

// OpenAudioFolder 打开音频文件夹
func (a *App) OpenAudioFolder() string {
	return myTTS.OpenAudioFolder()
}

// GetAudioDataUrl 获取音频文件的数据URL
func (a *App) GetAudioDataUrl(filePath string) string {
	return myTTS.GetAudioDataUrl(filePath)
}

// GetImageDataUrl 获取图片数据URL
func (a *App) GetImageDataUrl(filePath string) string {
	return myTTS.GetImageDataUrl(filePath)
}

// OpenImageFolder 打开图片文件夹
func (a *App) OpenImageFolder() error {
	cacheDir := core.GetTempDir()
	imgDir := filepath.Join(cacheDir, "tmp", "images", time.Now().Format("20060102"))
	os.MkdirAll(imgDir, 0755)
	a.OpenExplorer(imgDir)
	return nil
}

// CloneVoice 克隆音色
func (a *App) CloneVoice(req myTTS.CloneVoiceReq) myTTS.CloneVoiceRes {
	res := myTTS.CloneVoice(req)
	if res.Success {
		myTTS.SaveCustomVoice(req.VoiceID)
	}
	return res
}

// ListCustomVoices 列出已克隆的自定义音色
func (a *App) ListCustomVoices() myTTS.ListCustomVoicesRes {
	return myTTS.ListCustomVoices()
}

// SaveImagePrompt 保存图片提示词
func (a *App) SaveImagePrompt(prompt string) error {
	if settingsDb == nil {
		return fmt.Errorf("数据库未初始化")
	}
	if prompt == "" {
		return nil
	}
	_, err := settingsDb.Exec(`INSERT OR IGNORE INTO image_prompts(prompt) VALUES(?)`, prompt)
	return err
}

// GetImagePrompts 获取所有图片提示词
func (a *App) GetImagePrompts() ([]string, error) {
	if settingsDb == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}
	rows, err := settingsDb.Query(`SELECT prompt FROM image_prompts ORDER BY created_at DESC LIMIT 50`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prompts []string
	for rows.Next() {
		var prompt string
		if err := rows.Scan(&prompt); err != nil {
			continue
		}
		prompts = append(prompts, prompt)
	}
	return prompts, nil
}

// TextToSpeechLocal 使用 edge-tts 本地转语音（不需要 API Key）
func (a *App) TextToSpeechLocal(req myTTS.TextToSpeechLocalReq) myTTS.TextToSpeechLocalRes {
	return myTTS.TextToSpeechLocal(req)
}
