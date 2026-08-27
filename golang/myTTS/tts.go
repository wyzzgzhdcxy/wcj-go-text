package myTTS

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/wyzzgzhdcxy/wcj-go-common/core"

	"wcj-go-text/golang/cmdWrapper"
	"wcj-go-text/golang/rpc"
)

// MiniMax TTS API 相关结构

// maskKey 返回 API Key 的掩码字符串，只保留末尾 8 位，避免完整密钥写入日志。
func maskKey(s string) string {
	if len(s) <= 8 {
		return "****"
	}
	return "****" + s[len(s)-8:]
}

type MiniMaxTTsReq struct {
	Text    string `json:"text"`    // 要转换的文本
	VoiceID string `json:"voiceId"` // 音色 ID，默认 "longxia"
}

type MiniMaxTTSRes struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	OutputPath string `json:"outputPath"` // 输出文件路径
	Cost       string `json:"cost"`       // 耗时
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

	// 调用 RPC 层获取音频数据
	client := rpc.NewMiniMaxClient()
	resp, err := client.TextToSpeech(context.Background(), rpc.TTSRequest{
		Text:    req.Text,
		VoiceID: req.VoiceID,
	})
	if err != nil {
		return MiniMaxTTSRes{
			Success: false,
			Message: err.Error(),
		}
	}

	// 根据实际格式调整扩展名
	if resp.AudioFormat != ".mp3" {
		outputPath = strings.TrimSuffix(outputPath, ".mp3") + resp.AudioFormat
		log.Printf("检测到音频格式为 %s，调整输出文件为: %s", resp.AudioFormat, outputPath)
	}

	// 保存到文件
	if err := os.WriteFile(outputPath, resp.Audio, 0644); err != nil {
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
	encodedPath := strings.ReplaceAll(filePath, " ", "%20")
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

	// 通过 RPC 层获取图片 URL 列表
	client := rpc.NewMiniMaxClient()
	resp, err := client.GenerateImage(context.Background(), rpc.ImageRequest{
		Text:               req.Text,
		NumImages:          req.NumImages,
		Width:              req.Width,
		Height:             req.Height,
		ReferenceImagePath: req.ReferenceImagePath,
	})
	if err != nil {
		return MiniMaxImageRes{
			Success: false,
			Message: err.Error(),
		}
	}

	// 下载所有图片
	var outputPaths []string
	for i, imgURL := range resp.ImageURLs {
		ext := ".jpeg"
		if strings.Contains(imgURL, ".png") {
			ext = ".png"
		}
		outputPath := filepath.Join(imgDir, fmt.Sprintf("img_%s_%d%s", timestamp, i+1, ext))

		imgData, err := client.DownloadImageBytes(context.Background(), imgURL)
		if err != nil {
			log.Printf("下载第 %d 张图片失败: %v", i+1, err)
			continue
		}

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

	client := rpc.NewMiniMaxClient()
	_, err := client.CloneVoice(context.Background(), rpc.CloneRequest{
		SourceFileID:    req.SourceFileID,
		ReferenceText:   req.ReferenceText,
		ReferenceFileID: req.ReferenceFileID,
		VoiceID:         req.VoiceID,
	})
	if err != nil {
		return CloneVoiceRes{
			Success: false,
			Message: err.Error(),
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