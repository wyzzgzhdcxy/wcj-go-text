package myTTS

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/wyzzgzhdcxy/wcj-go-common/core"

	"wcj-go-text/golang/cmdWrapper"
)

// MiniMax TTS API 相关结构

type MiniMaxTTsReq struct {
	Text    string `json:"text"`    // 要转换的文本
	ApiKey  string `json:"apiKey"`  // MiniMax API Key
	VoiceID string `json:"voiceId"` // 音色 ID，默认 "longxia"
}

type MiniMaxTTSRes struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	OutputPath string `json:"outputPath"` // 输出文件路径
	Cost       string `json:"cost"`       // 耗时
}

// MiniMax TTS API 响应结构
type miniMaxAPIResp struct {
	BaseResp struct {
		StatusCode    int    `json:"status_code"`
		StatusMessage string `json:"status_msg"`
	} `json:"base_resp"`
	Data struct {
		Audio string `json:"audio"` // Base64 编码的音频
	} `json:"data"`
}

// TextToSpeech 使用 MiniMax TTS API 将文本转换为语音
func TextToSpeech(req MiniMaxTTsReq) MiniMaxTTSRes {
	startTime := time.Now()

	if req.Text == "" {
		return MiniMaxTTSRes{
			Success: false,
			Message: "请输入要转换的文本",
		}
	}

	apiKey := req.ApiKey
	if apiKey == "" {
		apiKey = os.Getenv("MINIMAX_API_KEY")
	}
	if apiKey == "" {
		return MiniMaxTTSRes{
			Success: false,
			Message: "请配置 MiniMax API Key (环境变量 MINIMAX_API_KEY)",
		}
	}

	// 获取缓存目录
	cacheDir := core.GetTempDir()
	ttsDir := filepath.Join(cacheDir, "tts")
	if err := os.MkdirAll(ttsDir, 0755); err != nil {
		return MiniMaxTTSRes{
			Success: false,
			Message: "创建缓存目录失败：" + err.Error(),
		}
	}

	// 生成输出文件路径
	timestamp := time.Now().Format("20060102_150405")
	outputPath := filepath.Join(ttsDir, fmt.Sprintf("tts_%s.mp3", timestamp))

	// 调用的 API URL (MiniMax TTS Pro API)
	apiURL := "https://api.minimaxi.com/v1/t2a_v2"
	// 构建请求体
	voiceID := req.VoiceID
	if voiceID == "" {
		voiceID = "male-qn-qingse" // 默认音色
	}
	// 模型和音色映射
	model := "speech-2.8-hd" // 高质量模型
	audioType := "mp3"

	requestBody := map[string]interface{}{
		"model":  model,
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
			"audio_type":  audioType,
			"sample_rate": 32000,
			"bitrate":     128000,
		},
	}

	// 发送请求
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return MiniMaxTTSRes{
			Success: false,
			Message: "请求体构建失败：" + err.Error(),
		}
	}

	httpReq, err := http.NewRequest("POST", apiURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return MiniMaxTTSRes{
			Success: false,
			Message: "创建请求失败：" + err.Error(),
		}
	}

	log.Printf("key:%s", apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	log.Printf("request url:%s\n", string(bodyBytes))

	resp, err := client.Do(httpReq)
	if err != nil {
		return MiniMaxTTSRes{
			Success: false,
			Message: "请求失败：" + err.Error(),
		}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return MiniMaxTTSRes{
			Success: false,
			Message: "读取响应失败：" + err.Error(),
		}
	}
	// 解析响应
	var apiResp miniMaxAPIResp
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return MiniMaxTTSRes{
			Success: false,
			Message: "解析响应失败：" + err.Error() + "，响应：" + string(respBody),
		}
	}

	// 检查 API 返回错误
	if apiResp.BaseResp.StatusCode != 0 {
		return MiniMaxTTSRes{
			Success: false,
			Message: fmt.Sprintf("API 返回错误：code=%d, msg=%s, resp=%s", apiResp.BaseResp.StatusCode, apiResp.BaseResp.StatusMessage, string(respBody)),
		}
	}

	if apiResp.Data.Audio == "" {
		return MiniMaxTTSRes{
			Success: false,
			Message: fmt.Sprintf("API 返回音频为空：resp=%s", string(respBody)),
		}
	}

	// 解码音频数据 (可能是 hex 或 base64 编码)
	var audioData []byte
	// 先尝试 hex 解码 (MiniMax 返回的是 hex 编码的音频)
	audioData, err = hex.DecodeString(apiResp.Data.Audio)
	if err != nil {
		return MiniMaxTTSRes{
			Success: false,
			Message: "音频解码失败：" + err.Error(),
		}
	}

	log.Printf("音频数据长度: %d bytes, 前20字节: %v", len(audioData), audioData[:20])

	// 检测音频格式并设置正确的文件扩展名
	actualExt := detectAudioFormat(audioData)
	if actualExt != ".mp3" {
		outputPath = strings.TrimSuffix(outputPath, ".mp3") + actualExt
		log.Printf("检测到音频格式为 %s，调整输出文件为: %s", actualExt, outputPath)
	}

	// 保存到文件
	if err := os.WriteFile(outputPath, audioData, 0644); err != nil {
		return MiniMaxTTSRes{
			Success: false,
			Message: "保存文件失败：" + err.Error(),
		}
	}

	// 验证文件
	info, _ := os.Stat(outputPath)
	log.Printf("文件已保存，大小: %d bytes", info.Size())

	elapsed := time.Since(startTime)

	return MiniMaxTTSRes{
		Success:    true,
		Message:    "转换成功",
		OutputPath: outputPath,
		Cost:       elapsed.String(),
	}
}

// GetAvailableVoices 获取可用的音色列表
func GetAvailableVoices() []struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Language string `json:"language"`
} {
	return []struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Language string `json:"language"`
	}{
		// 中文 (普通话)
		{ID: "male-qn-qingse", Name: "青涩青年音色", Language: "中文(普通话)"},
		{ID: "male-qn-jingying", Name: "精英青年音色", Language: "中文(普通话)"},
		{ID: "male-qn-badao", Name: "霸道青年音色", Language: "中文(普通话)"},
		{ID: "male-qn-daxuesheng", Name: "青年大学生音色", Language: "中文(普通话)"},
		{ID: "female-shaonv", Name: "少女音色", Language: "中文(普通话)"},
		{ID: "female-yujie", Name: "御姐音色", Language: "中文(普通话)"},
		{ID: "female-chengshu", Name: "成熟女性音色", Language: "中文(普通话)"},
		{ID: "female-tianmei", Name: "甜美女性音色", Language: "中文(普通话)"},
		{ID: "male-qn-qingse-jingpin", Name: "青涩青年音色-beta", Language: "中文(普通话)"},
		{ID: "male-qn-jingying-jingpin", Name: "精英青年音色-beta", Language: "中文(普通话)"},
		{ID: "male-qn-badao-jingpin", Name: "霸道青年音色-beta", Language: "中文(普通话)"},
		{ID: "male-qn-daxuesheng-jingpin", Name: "青年大学生音色-beta", Language: "中文(普通话)"},
		{ID: "female-shaonv-jingpin", Name: "少女音色-beta", Language: "中文(普通话)"},
		{ID: "female-yujie-jingpin", Name: "御姐音色-beta", Language: "中文(普通话)"},
		{ID: "female-chengshu-jingpin", Name: "成熟女性音色-beta", Language: "中文(普通话)"},
		{ID: "female-tianmei-jingpin", Name: "甜美女性音色-beta", Language: "中文(普通话)"},
		{ID: "clever_boy", Name: "聪明男童", Language: "中文(普通话)"},
		{ID: "cute_boy", Name: "可爱男童", Language: "中文(普通话)"},
		{ID: "lovely_girl", Name: "萌萌女童", Language: "中文(普通话)"},
		{ID: "cartoon_pig", Name: "卡通猪小琪", Language: "中文(普通话)"},
		{ID: "bingjiao_didi", Name: "病娇弟弟", Language: "中文(普通话)"},
		{ID: "junlang_nanyou", Name: "俊朗男友", Language: "中文(普通话)"},
		{ID: "chunzhen_xuedi", Name: "纯真学弟", Language: "中文(普通话)"},
		{ID: "lengdan_xiongzhang", Name: "冷淡学长", Language: "中文(普通话)"},
		{ID: "badao_shaoye", Name: "霸道少爷", Language: "中文(普通话)"},
		{ID: "tianxin_xiaoling", Name: "甜心小玲", Language: "中文(普通话)"},
		{ID: "qiaopi_mengmei", Name: "俏皮萌妹", Language: "中文(普通话)"},
		{ID: "wumei_yujie", Name: "妩媚御姐", Language: "中文(普通话)"},
		{ID: "diadia_xuemei", Name: "嗲嗲学妹", Language: "中文(普通话)"},
		{ID: "danya_xuejie", Name: "淡雅学姐", Language: "中文(普通话)"},
		{ID: "Chinese (Mandarin)_Reliable_Executive", Name: "沉稳高管", Language: "中文(普通话)"},
		{ID: "Chinese (Mandarin)_News_Anchor", Name: "新闻女声", Language: "中文(普通话)"},
		{ID: "Chinese (Mandarin)_Mature_Woman", Name: "傲娇御姐", Language: "中文(普通话)"},
		{ID: "Chinese (Mandarin)_Unrestrained_Young_Man", Name: "不羁青年", Language: "中文(普通话)"},
		{ID: "Arrogant_Miss", Name: "嚣张小姐", Language: "中文(普通话)"},
		{ID: "Robot_Armor", Name: "机械战甲", Language: "中文(普通话)"},
		{ID: "Chinese (Mandarin)_Kind-hearted_Antie", Name: "热心大婶", Language: "中文(普通话)"},
		{ID: "Chinese (Mandarin)_HK_Flight_Attendant", Name: "港普空姐", Language: "中文(普通话)"},
		{ID: "Chinese (Mandarin)_Humorous_Elder", Name: "搞笑大爷", Language: "中文(普通话)"},
		{ID: "Chinese (Mandarin)_Gentleman", Name: "温润男声", Language: "中文(普通话)"},
		{ID: "Chinese (Mandarin)_Warm_Bestie", Name: "温暖闺蜜", Language: "中文(普通话)"},
		{ID: "Chinese (Mandarin)_Male_Announcer", Name: "播报男声", Language: "中文(普通话)"},
		{ID: "Chinese (Mandarin)_Sweet_Lady", Name: "甜美女声", Language: "中文(普通话)"},
		{ID: "Chinese (Mandarin)_Southern_Young_Man", Name: "南方小哥", Language: "中文(普通话)"},
		{ID: "Chinese (Mandarin)_Wise_Women", Name: "阅历姐姐", Language: "中文(普通话)"},
		{ID: "Chinese (Mandarin)_Gentle_Youth", Name: "温润青年", Language: "中文(普通话)"},
		{ID: "Chinese (Mandarin)_Warm_Girl", Name: "温暖少女", Language: "中文(普通话)"},
		{ID: "Chinese (Mandarin)_Kind-hearted_Elder", Name: "花甲奶奶", Language: "中文(普通话)"},
		{ID: "Chinese (Mandarin)_Cute_Spirit", Name: "憨憨萌兽", Language: "中文(普通话)"},
		{ID: "Chinese (Mandarin)_Radio_Host", Name: "电台男主播", Language: "中文(普通话)"},
		{ID: "Chinese (Mandarin)_Lyrical_Voice", Name: "抒情男声", Language: "中文(普通话)"},
		{ID: "Chinese (Mandarin)_Straightforward_Boy", Name: "率真弟弟", Language: "中文(普通话)"},
		{ID: "Chinese (Mandarin)_Sincere_Adult", Name: "真诚青年", Language: "中文(普通话)"},
		{ID: "Chinese (Mandarin)_Gentle_Senior", Name: "温柔学姐", Language: "中文(普通话)"},
		{ID: "Chinese (Mandarin)_Stubborn_Friend", Name: "嘴硬竹马", Language: "中文(普通话)"},
		{ID: "Chinese (Mandarin)_Crisp_Girl", Name: "清脆少女", Language: "中文(普通话)"},
		{ID: "Chinese (Mandarin)_Pure-hearted_Boy", Name: "清澈邻家弟弟", Language: "中文(普通话)"},
		{ID: "Chinese (Mandarin)_Soft_Girl", Name: "柔和少女", Language: "中文(普通话)"},
		// 中文 (粤语)
		{ID: "Cantonese_ProfessionalHost（F)", Name: "专业女主持(粤语)", Language: "中文(粤语)"},
		{ID: "Cantonese_GentleLady", Name: "温柔女声(粤语)", Language: "中文(粤语)"},
		{ID: "Cantonese_ProfessionalHost（M)", Name: "专业男主持(粤语)", Language: "中文(粤语)"},
		{ID: "Cantonese_PlayfulMan", Name: "活泼男声(粤语)", Language: "中文(粤语)"},
		{ID: "Cantonese_CuteGirl", Name: "可爱女孩(粤语)", Language: "中文(粤语)"},
		{ID: "Cantonese_KindWoman", Name: "善良女声(粤语)", Language: "中文(粤语)"},
	}
}

// detectAudioFormat 检测音频格式
func detectAudioFormat(data []byte) string {
	if len(data) < 12 {
		return ".mp3" // 默认
	}
	// MP3: FF FB 或 FF F3 或 ID3
	if data[0] == 0xFF && (data[1]&0xE0) == 0xE0 {
		return ".mp3"
	}
	// WAV: RIFF....WAVE
	if len(data) >= 12 && string(data[0:4]) == "RIFF" && string(data[8:12]) == "WAVE" {
		return ".wav"
	}
	// OGG: OggS
	if len(data) >= 4 && string(data[0:4]) == "OggS" {
		return ".ogg"
	}
	// FLAC: fLaC
	if len(data) >= 4 && string(data[0:4]) == "fLaC" {
		return ".flac"
	}
	// M4A/AAC: ftyp
	if len(data) >= 8 && string(data[4:8]) == "ftyp" {
		return ".m4a"
	}
	return ".mp3" // 默认
}

// GetAudioDataUrl 读取音频文件并转换为 data URL (用于前端播放)
func GetAudioDataUrl(filePath string) string {
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("读取音频文件失败: %v", err)
		return ""
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	var mimeType string
	switch ext {
	case ".mp3":
		mimeType = "audio/mpeg"
	case ".wav":
		mimeType = "audio/wav"
	case ".ogg":
		mimeType = "audio/ogg"
	case ".m4a":
		mimeType = "audio/mp4"
	case ".flac":
		mimeType = "audio/flac"
	default:
		mimeType = "audio/mpeg"
	}

	// 使用 base64 编码
	encoded := base64.StdEncoding.EncodeToString(data)
	return fmt.Sprintf("data:%s;base64,%s", mimeType, encoded)
}

// GetImageDataUrl 读取图片文件并转换为 HTTP URL
func GetImageDataUrl(filePath string) string {
	filePath = strings.ReplaceAll(filePath, "\\", "/")
	encodedPath := url.QueryEscape(filePath)
	return fmt.Sprintf("http://localhost:45670/file?path=%s", encodedPath)
}

// OpenAudioFolder 打开音频文件夹
func OpenAudioFolder() string {
	cacheDir := core.GetTempDir()
	ttsDir := filepath.Join(cacheDir, "tts")
	if err := os.MkdirAll(ttsDir, 0755); err != nil {
		// 创建失败也尝试打开
	}
	// 打开文件夹
	go func() {
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "windows":
			cmd = cmdWrapper.Command("explorer", ttsDir)
		case "darwin":
			cmd = cmdWrapper.Command("open", ttsDir)
		case "linux":
			cmd = cmdWrapper.Command("xdg-open", ttsDir)
		}
		if cmd != nil {
			cmd.Start()
		}
	}()
	return ttsDir
}

// MiniMax Image Gen API 相关结构

type MiniMaxImageReq struct {
	Text               string `json:"text"`                 // 图片描述文本
	ApiKey             string `json:"apiKey"`               // MiniMax API Key
	NumImages          int    `json:"numImages"`            // 生成图片数量，默认9
	Width              int    `json:"width"`                // 生成图片宽度（像素），取值范围[512, 2048]，必须是8的倍数
	Height             int    `json:"height"`               // 生成图片高度（像素），取值范围[512, 2048]，必须是8的倍数
	ReferenceImagePath string `json:"reference_image_path"` // 参考图文件路径
}

type MiniMaxImageRes struct {
	Success    bool     `json:"success"`
	Message    string   `json:"message"`
	OutputPath []string `json:"outputPath"` // 输出文件路径数组
	Cost       string   `json:"cost"`       // 耗时
}

// MiniMaxImageResp API 响应结构
type miniMaxImageResp struct {
	BaseResp struct {
		StatusCode    int    `json:"status_code"`
		StatusMessage string `json:"status_msg"`
	} `json:"base_resp"`
	Data struct {
		ImageURLs []string `json:"image_urls"` // 图片 URL 数组
	}
}

// GenerateImage 使用 MiniMax API 根据文本生成图片
func GenerateImage(req MiniMaxImageReq) MiniMaxImageRes {
	startTime := time.Now()

	// 打印收到的请求参数
	reqJSON, _ := json.Marshal(req)
	log.Printf("[GenerateImage] 收到请求: %s", string(reqJSON))

	if req.Text == "" {
		return MiniMaxImageRes{
			Success: false,
			Message: "请输入图片描述文本",
		}
	}

	apiKey := req.ApiKey
	if apiKey == "" {
		apiKey = os.Getenv("MINIMAX_API_KEY")
	}
	if apiKey == "" {
		return MiniMaxImageRes{
			Success: false,
			Message: "请配置 MiniMax API Key (环境变量 MINIMAX_API_KEY)",
		}
	}

	// 获取缓存目录
	cacheDir := core.GetTempDir()
	imgDir := filepath.Join(cacheDir, "tmp", "images", time.Now().Format("20060102"))
	if err := os.MkdirAll(imgDir, 0755); err != nil {
		return MiniMaxImageRes{
			Success: false,
			Message: "创建缓存目录失败：" + err.Error(),
		}
	}

	// 生成输出文件路径
	timestamp := time.Now().Format("20060102_150405")

	// 调用的 API URL (MiniMax Image Generation API)
	apiURL := "https://api.minimaxi.com/v1/image_generation"

	// 图片数量，默认9
	numImages := req.NumImages
	if numImages <= 0 || numImages > 9 {
		numImages = 9
	}

	// 图片宽度，默认1024，取值范围[512, 2048]，必须是8的倍数
	width := req.Width
	if width < 512 || width > 2048 || width%8 != 0 {
		width = 1024
	}

	// 图片高度，默认1024，取值范围[512, 2048]，必须是8的倍数
	height := req.Height
	if height < 512 || height > 2048 || height%8 != 0 {
		height = 1024
	}

	// 请求体
	requestBody := map[string]any{
		"model":           "image-01",
		"prompt":          req.Text,
		"n":               numImages,
		"width":           width,
		"height":          height,
		"response_format": "url", // 返回 URL 格式
	}

	// 如果有参考图，读取文件并添加到请求体
	if req.ReferenceImagePath != "" {
		imageData, err := os.ReadFile(req.ReferenceImagePath)
		if err != nil {
			return MiniMaxImageRes{
				Success: false,
				Message: "读取参考图失败：" + err.Error(),
			}
		}
		base64Data := base64.StdEncoding.EncodeToString(imageData)
		mimeType := "image/jpeg"
		if strings.HasSuffix(strings.ToLower(req.ReferenceImagePath), ".png") {
			mimeType = "image/png"
		}
		requestBody["subject_reference"] = []any{
			map[string]any{
				"type":       "character",
				"image_file": fmt.Sprintf("data:%s;base64,%s", mimeType, base64Data),
			},
		}
	}

	// 打印请求信息
	requestJSON, _ := json.Marshal(requestBody)
	reqStr := string(requestJSON)
	if len(reqStr) > 2000 {
		reqStr = reqStr[:2000] + "...(省略 " + fmt.Sprintf("%d", len(reqStr)-2000) + " 字符)"
	}
	log.Printf("[GenerateImage] 请求信息: %s", reqStr)

	// 发送请求
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return MiniMaxImageRes{
			Success: false,
			Message: "请求体构建失败：" + err.Error(),
		}
	}

	httpReq, err := http.NewRequest("POST", apiURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return MiniMaxImageRes{
			Success: false,
			Message: "创建请求失败：" + err.Error(),
		}
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{
		Timeout: 120 * time.Second,
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return MiniMaxImageRes{
			Success: false,
			Message: "请求失败：" + err.Error(),
		}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return MiniMaxImageRes{
			Success: false,
			Message: "读取响应失败：" + err.Error(),
		}
	}

	log.Printf("image resp:%s", string(respBody))

	// 解析响应
	var apiResp miniMaxImageResp
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return MiniMaxImageRes{
			Success: false,
			Message: "解析响应失败：" + err.Error() + "，响应：" + string(respBody),
		}
	}

	// 检查 API 返回错误
	if apiResp.BaseResp.StatusCode != 0 {
		return MiniMaxImageRes{
			Success: false,
			Message: fmt.Sprintf("API 返回错误：code=%d, msg=%s", apiResp.BaseResp.StatusCode, apiResp.BaseResp.StatusMessage),
		}
	}

	// 下载图片
	if len(apiResp.Data.ImageURLs) == 0 || apiResp.Data.ImageURLs[0] == "" {
		return MiniMaxImageRes{
			Success: false,
			Message: "API 返回图片 URL 为空",
		}
	}

	// 下载所有图片
	var outputPaths []string
	for i, imgURL := range apiResp.Data.ImageURLs {
		// 生成输出文件路径
		ext := ".jpeg"
		if strings.Contains(imgURL, ".png") {
			ext = ".png"
		}
		outputPath := filepath.Join(imgDir, fmt.Sprintf("img_%s_%d%s", timestamp, i+1, ext))

		// 下载图片
		imgResp, err := client.Get(imgURL)
		if err != nil {
			log.Printf("下载第 %d 张图片失败: %v", i+1, err)
			continue
		}

		imgData, err := io.ReadAll(imgResp.Body)
		imgResp.Body.Close()
		if err != nil {
			log.Printf("读取第 %d 张图片数据失败: %v", i+1, err)
			continue
		}

		// 保存到文件
		if err := os.WriteFile(outputPath, imgData, 0644); err != nil {
			log.Printf("保存第 %d 张图片失败: %v", i+1, err)
			continue
		}

		outputPaths = append(outputPaths, outputPath)
	}

	if len(outputPaths) == 0 {
		return MiniMaxImageRes{
			Success: false,
			Message: "没有成功下载任何图片",
		}
	}

	elapsed := time.Since(startTime)

	return MiniMaxImageRes{
		Success:    true,
		Message:    fmt.Sprintf("生成成功，共 %d 张图片", len(outputPaths)),
		OutputPath: outputPaths,
		Cost:       elapsed.String(),
	}
}

// OpenImageFolder 打开图片文件夹
func OpenImageFolder() string {
	cacheDir := core.GetTempDir()
	imgDir := filepath.Join(cacheDir, "tmp", "images", time.Now().Format("20060102"))
	if err := os.MkdirAll(imgDir, 0755); err != nil {
		return imgDir
	}
	return imgDir
}

// TextToSpeechLocalReq 本地 TTS 请求（使用 edge-tts 微软接口）
type TextToSpeechLocalReq struct {
	Text       string `json:"text"`
	Voice      string `json:"voice"` // 例如 zh-CN-XiaoxiaoNeural
	OutputPath string `json:"outputPath"`
}

type TextToSpeechLocalRes struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	OutputPath string `json:"outputPath"`
	Cost       string `json:"cost"`
}

// TextToSpeechLocal 使用 edge-tts 本地转语音（不需要 API Key）
func TextToSpeechLocal(req TextToSpeechLocalReq) TextToSpeechLocalRes {
	startTime := time.Now()

	if req.Text == "" {
		return TextToSpeechLocalRes{
			Success: false,
			Message: "请输入要转换的文本",
		}
	}

	// 默认音色
	voice := req.Voice
	if voice == "" {
		voice = "zh-CN-XiaoxiaoNeural"
	}

	// 获取缓存目录
	cacheDir := core.GetTempDir()
	ttsDir := filepath.Join(cacheDir, "tts")
	if err := os.MkdirAll(ttsDir, 0755); err != nil {
		return TextToSpeechLocalRes{
			Success: false,
			Message: "创建缓存目录失败：" + err.Error(),
		}
	}

	// 生成输出文件路径
	timestamp := time.Now().Format("20060102_150405")
	outputPath := req.OutputPath
	if outputPath == "" {
		outputPath = filepath.Join(ttsDir, fmt.Sprintf("tts_%s.mp3", timestamp))
	}

	// edge-tts 命令
	// 格式: edge-tts --text "文本" --voice "音色" --write-media "输出路径"
	cmd := cmdWrapper.Command("edge-tts",
		"--text", req.Text,
		"--voice", voice,
		"--write-media", outputPath)

	err := cmd.Run()
	if err != nil {
		// 如果 edge-tts 不可用，尝试使用 ffmpeg 方式
		return TextToSpeechLocalRes{
			Success: false,
			Message: "edge-tts 不可用：" + err.Error(),
		}
	}

	elapsed := time.Since(startTime)

	return TextToSpeechLocalRes{
		Success:    true,
		Message:    "转换成功",
		OutputPath: outputPath,
		Cost:       elapsed.String(),
	}
}

// ListTTsFiles 列出缓存目录中的 TTS 文件
func ListTTsFiles() ([]string, error) {
	cacheDir := core.GetTempDir()
	ttsDir := filepath.Join(cacheDir, "tts")

	if _, err := os.Stat(ttsDir); os.IsNotExist(err) {
		return []string{}, nil
	}

	entries, err := os.ReadDir(ttsDir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".mp3") {
			files = append(files, filepath.Join(ttsDir, entry.Name()))
		}
	}

	return files, nil
}

// ========== 音色克隆相关 ==========

// UploadAudioForCloneReq 上传音频用于克隆的请求
type UploadAudioForCloneReq struct {
	FilePath string `json:"filePath"` // 音频文件路径
}

// UploadAudioForCloneRes 上传音频用于克隆的响应
type UploadAudioForCloneRes struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	FileID  string `json:"file_id"` // 文件ID
}

// UploadReferenceAudioReq 上传参考音频的请求
type UploadReferenceAudioReq struct {
	FilePath string `json:"filePath"` // 音频文件路径
}

// UploadReferenceAudioRes 上传参考音频的响应
type UploadReferenceAudioRes struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	FileID  string `json:"file_id"` // 文件ID
}

// CloneVoiceReq 克隆音色的请求
type CloneVoiceReq struct {
	SourceFileID    string `json:"source_file_id"`    // 源音频文件ID
	ReferenceText   string `json:"reference_text"`    // 参考文本
	ReferenceFileID string `json:"reference_file_id"` // 参考音频文件ID (可选)
	VoiceID         string `json:"voice_id"`          // 自定义的音色ID
}

// CloneVoiceRes 克隆音色的响应
type CloneVoiceRes struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// CloneVoice 音色克隆
func CloneVoice(req CloneVoiceReq) CloneVoiceRes {
	apiKey := os.Getenv("MINIMAX_API_KEY")
	if apiKey == "" {
		return CloneVoiceRes{
			Success: false,
			Message: "请配置 MiniMax API Key (环境变量 MINIMAX_API_KEY)",
		}
	}

	if req.SourceFileID == "" {
		return CloneVoiceRes{
			Success: false,
			Message: "请先上传待克隆音频",
		}
	}

	if req.VoiceID == "" {
		return CloneVoiceRes{
			Success: false,
			Message: "请输入自定义音色ID",
		}
	}

	// 调用音色克隆 API
	apiURL := "https://api.minimaxi.com/v1/voice_clone"

	requestBody := map[string]interface{}{
		"model": "speech-01",
		"input": map[string]interface{}{
			"source": req.SourceFileID,
		},
		"voice_id": req.VoiceID,
	}

	if req.ReferenceText != "" {
		requestBody["reference_text"] = req.ReferenceText
	}
	if req.ReferenceFileID != "" {
		requestBody["reference"] = req.ReferenceFileID
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return CloneVoiceRes{
			Success: false,
			Message: "请求体构建失败：" + err.Error(),
		}
	}

	httpReq, err := http.NewRequest("POST", apiURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return CloneVoiceRes{
			Success: false,
			Message: "创建请求失败：" + err.Error(),
		}
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return CloneVoiceRes{
			Success: false,
			Message: "请求失败：" + err.Error(),
		}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return CloneVoiceRes{
			Success: false,
			Message: "读取响应失败：" + err.Error(),
		}
	}

	log.Printf("clone voice resp:%s", string(respBody))

	// 解析响应
	var apiResp struct {
		BaseResp struct {
			StatusCode    int    `json:"status_code"`
			StatusMessage string `json:"status_msg"`
		} `json:"base_resp"`
	}
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return CloneVoiceRes{
			Success: false,
			Message: "解析响应失败：" + err.Error(),
		}
	}

	if apiResp.BaseResp.StatusCode != 0 {
		return CloneVoiceRes{
			Success: false,
			Message: fmt.Sprintf("API 返回错误：code=%d, msg=%s", apiResp.BaseResp.StatusCode, apiResp.BaseResp.StatusMessage),
		}
	}

	return CloneVoiceRes{
		Success: true,
		Message: "音色克隆成功！音色ID: " + req.VoiceID,
	}
}

// ListCustomVoicesReq 列出自定义音色的请求
type ListCustomVoicesReq struct{}

// ListCustomVoicesRes 列出自定义音色的响应
type ListCustomVoicesRes struct {
	Success      bool     `json:"success"`
	Message      string   `json:"message"`
	CustomVoices []string `json:"custom_voices"` // 自定义音色ID列表
}

// ListCustomVoices 列出已克隆的自定义音色
func ListCustomVoices() ListCustomVoicesRes {
	cacheDir := core.GetTempDir()
	voicesFile := filepath.Join(cacheDir, "custom_voices.json")

	if _, err := os.Stat(voicesFile); os.IsNotExist(err) {
		return ListCustomVoicesRes{
			Success:      true,
			CustomVoices: []string{},
		}
	}

	data, err := os.ReadFile(voicesFile)
	if err != nil {
		return ListCustomVoicesRes{
			Success: true,
			Message: "读取失败",
		}
	}

	var voices []string
	if err := json.Unmarshal(data, &voices); err != nil {
		return ListCustomVoicesRes{
			Success: true,
			Message: "解析失败",
		}
	}

	return ListCustomVoicesRes{
		Success:      true,
		CustomVoices: voices,
	}
}

// SaveCustomVoice 保存自定义音色ID到列表
func SaveCustomVoice(voiceID string) error {
	cacheDir := core.GetTempDir()
	voicesFile := filepath.Join(cacheDir, "custom_voices.json")

	var voices []string
	if _, err := os.Stat(voicesFile); err == nil {
		data, _ := os.ReadFile(voicesFile)
		json.Unmarshal(data, &voices)
	}

	// 检查是否已存在
	for _, v := range voices {
		if v == voiceID {
			return nil
		}
	}

	voices = append(voices, voiceID)
	data, _ := json.Marshal(voices)
	return os.WriteFile(voicesFile, data, 0644)
}
