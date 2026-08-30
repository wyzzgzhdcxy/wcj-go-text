package myVideo

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/wyzzgzhdcxy/wcj-go-common/core"
)

// BatchExtractAudioReq 批量提取音频请求
type BatchExtractAudioReq struct {
	DirPath     string `json:"dirPath"`     // 视频目录路径
	Format      string `json:"format"`      // 输出格式: mp3, aac, wav, flac, 默认 mp3
	ThreadCount int    `json:"threadCount"` // 并行线程数，默认 4
}

// BatchExtractAudioRes 批量提取音频结果
type BatchExtractAudioRes struct {
	Success      bool     `json:"success"`
	Message      string   `json:"message"`
	TotalCount   int      `json:"totalCount"`   // 总视频数
	SuccessCount int      `json:"successCount"` // 成功数
	FailedCount  int      `json:"failedCount"`  // 失败数
	FailedFiles  []string `json:"failedFiles"`  // 失败文件列表
	OutputDir    string   `json:"outputDir"`    // 输出目录
	TotalCost    string   `json:"totalCost"`    // 总耗时
}

// BatchExtractAudio 批量从视频中提取音频（多线程）
func BatchExtractAudio(req BatchExtractAudioReq, callback func([]byte)) BatchExtractAudioRes {
	startTime := time.Now()
	callback([]byte("开始批量提取音频..."))

	if req.DirPath == "" {
		return BatchExtractAudioRes{Success: false, Message: "请选择视频目录"}
	}

	if !core.FileExist(req.DirPath) {
		return BatchExtractAudioRes{Success: false, Message: "目录不存在：" + req.DirPath}
	}

	format := req.Format
	if format == "" {
		format = "mp3"
	}

	threadCount := req.ThreadCount
	if threadCount <= 0 {
		threadCount = 4
	}

	// 创建输出目录
	outputDir := filepath.Join(req.DirPath, "music")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return BatchExtractAudioRes{Success: false, Message: "创建输出目录失败：" + err.Error()}
	}

	callback([]byte(fmt.Sprintf("输出目录: %s，并行线程数: %d", outputDir, threadCount)))

	// 常见视频格式扩展名
	videoExts := map[string]bool{
		".mp4":  true,
		".avi":  true,
		".mkv":  true,
		".mov":  true,
		".wmv":  true,
		".flv":  true,
		".m4v":  true,
		".ts":   true,
		".rmvb": true,
		".webm": true,
	}

	// 扫描目录下所有视频文件
	entries, err := os.ReadDir(req.DirPath)
	if err != nil {
		return BatchExtractAudioRes{Success: false, Message: "无法读取目录：" + err.Error()}
	}

	var videoFiles []string
	var extCountMap = make(map[string]int)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		fileName := entry.Name()
		ext := strings.ToLower(filepath.Ext(fileName))
		if !videoExts[ext] {
			continue
		}
		videoFiles = append(videoFiles, filepath.Join(req.DirPath, fileName))
		extCountMap[ext]++
	}

	if len(videoFiles) == 0 {
		return BatchExtractAudioRes{Success: false, Message: "目录中没有找到视频文件"}
	}

	callback([]byte(fmt.Sprintf("共找到 %d 个视频文件", len(videoFiles))))

	// 统计各类型文件数
	var extStats []string
	for ext, count := range extCountMap {
		extStats = append(extStats, fmt.Sprintf("%s: %d个", ext, count))
	}
	callback([]byte("文件类型分布: " + strings.Join(extStats, ", ")))

	// 使用多线程处理
	var mu sync.Mutex
	var successCount, failedCount int
	var failedFiles []string
	var wg sync.WaitGroup

	// 控制并发数
	semaphore := make(chan struct{}, threadCount)

	for i, videoPath := range videoFiles {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(idx int, path string) {
			defer wg.Done()
			defer func() { <-semaphore }()

			callback([]byte(fmt.Sprintf("[%d/%d] 正在处理: %s", idx+1, len(videoFiles), filepath.Base(path))))

			// 生成输出路径
			baseName := filepath.Base(path)
			dotIdx := strings.LastIndex(baseName, ".")
			if dotIdx > 0 {
				baseName = baseName[:dotIdx]
			}
			outputPath := filepath.Join(outputDir, baseName+"."+format)

			// 构建 ffmpeg 命令
			var execCmd *exec.Cmd
			switch format {
			case "wav":
				execCmd = core.Command("ffmpeg", "-i", path, "-vn", "-acodec", "pcm_s16le", "-y", outputPath)
			case "aac":
				execCmd = core.Command("ffmpeg", "-i", path, "-vn", "-acodec", "copy", "-y", outputPath)
			case "flac":
				execCmd = core.Command("ffmpeg", "-i", path, "-vn", "-acodec", "flac", "-y", outputPath)
			default: // mp3
				execCmd = core.Command("ffmpeg", "-i", path, "-vn", "-acodec", "libmp3lame", "-q:a", "2", "-y", outputPath)
			}

			// 记录完整命令行
			cmdStr := "ffmpeg"
			for _, arg := range execCmd.Args {
				cmdStr += " " + arg
			}
			callback([]byte("执行命令: " + cmdStr))


			execCmd.Stdout = os.Stdout
			execCmd.Stderr = os.Stderr

			if err := execCmd.Run(); err != nil {
				mu.Lock()
				failedCount++
				failedFiles = append(failedFiles, path+" (错误: "+err.Error()+")")
				mu.Unlock()
				callback([]byte(fmt.Sprintf("❌ [%d/%d] 失败: %s", idx+1, len(videoFiles), filepath.Base(path))))
			} else {
				mu.Lock()
				successCount++
				mu.Unlock()
				callback([]byte(fmt.Sprintf("✅ [%d/%d] 完成: %s", idx+1, len(videoFiles), filepath.Base(path))))
			}
		}(i, videoPath)
	}

	wg.Wait()

	elapsed := time.Since(startTime)
	totalCost := fmt.Sprintf("%.1f秒", elapsed.Seconds())

	callback([]byte(fmt.Sprintf("批量提取完成！成功: %d，失败: %d，总耗时: %s", successCount, failedCount, totalCost)))

	return BatchExtractAudioRes{
		Success:      failedCount == 0,
		Message:      fmt.Sprintf("批量提取完成，成功: %d，失败: %d", successCount, failedCount),
		TotalCount:   len(videoFiles),
		SuccessCount: successCount,
		FailedCount:  failedCount,
		FailedFiles:  failedFiles,
		OutputDir:    outputDir,
		TotalCost:    totalCost,
	}
}

// BatchExtractAudioByFilesReq 按文件列表提取音频请求
type BatchExtractAudioByFilesReq struct {
	DirPath     string   `json:"dirPath"`     // 视频目录路径
	FileNames   []string `json:"fileNames"`   // 要处理的文件名列表
	Format      string   `json:"format"`      // 输出格式: mp3, aac, wav, flac, 默认 mp3
	ThreadCount int      `json:"threadCount"` // 并行线程数，默认 4
}

// BatchExtractAudioByFiles 按指定文件列表提取音频
func BatchExtractAudioByFiles(req BatchExtractAudioByFilesReq, callback func([]byte)) BatchExtractAudioRes {
	startTime := time.Now()
	callback([]byte("开始提取音频..."))

	if req.DirPath == "" {
		return BatchExtractAudioRes{Success: false, Message: "请选择视频目录"}
	}

	if !core.FileExist(req.DirPath) {
		return BatchExtractAudioRes{Success: false, Message: "目录不存在：" + req.DirPath}
	}

	if len(req.FileNames) == 0 {
		return BatchExtractAudioRes{Success: false, Message: "没有选择要处理的文件"}
	}

	format := req.Format
	if format == "" {
		format = "mp3"
	}

	threadCount := req.ThreadCount
	if threadCount <= 0 {
		threadCount = 4
	}

	// 创建输出目录
	outputDir := filepath.Join(req.DirPath, "music")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return BatchExtractAudioRes{Success: false, Message: "创建输出目录失败：" + err.Error()}
	}

	callback([]byte(fmt.Sprintf("输出目录: %s，并行线程数: %d", outputDir, threadCount)))
	callback([]byte(fmt.Sprintf("要处理的文件数: %d", len(req.FileNames))))

	// 构建文件路径集合
	fileNameSet := make(map[string]bool)
	for _, name := range req.FileNames {
		fileNameSet[name] = true
	}

	// 只处理选中的文件
	var videoFiles []string
	for _, name := range req.FileNames {
		videoFiles = append(videoFiles, filepath.Join(req.DirPath, name))
	}

	callback([]byte(fmt.Sprintf("共找到 %d 个视频文件", len(videoFiles))))

	// 使用多线程处理
	var mu sync.Mutex
	var successCount, failedCount int
	var failedFiles []string
	var wg sync.WaitGroup

	// 控制并发数
	semaphore := make(chan struct{}, threadCount)

	for i, videoPath := range videoFiles {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(idx int, path string) {
			defer wg.Done()
			defer func() { <-semaphore }()

			callback([]byte(fmt.Sprintf("[%d/%d] 正在处理: %s", idx+1, len(videoFiles), filepath.Base(path))))

			// 生成输出路径
			baseName := filepath.Base(path)
			dotIdx := strings.LastIndex(baseName, ".")
			if dotIdx > 0 {
				baseName = baseName[:dotIdx]
			}
			outputPath := filepath.Join(outputDir, baseName+"."+format)

			// 构建 ffmpeg 命令
			var execCmd *exec.Cmd
			switch format {
			case "wav":
				execCmd = core.Command("ffmpeg", "-i", path, "-vn", "-acodec", "pcm_s16le", "-y", outputPath)
			case "aac":
				execCmd = core.Command("ffmpeg", "-i", path, "-vn", "-acodec", "copy", "-y", outputPath)
			case "flac":
				execCmd = core.Command("ffmpeg", "-i", path, "-vn", "-acodec", "flac", "-y", outputPath)
			default: // mp3
				execCmd = core.Command("ffmpeg", "-i", path, "-vn", "-acodec", "libmp3lame", "-q:a", "2", "-y", outputPath)
			}

			// 记录完整命令行
			cmdStr := "ffmpeg"
			for _, arg := range execCmd.Args {
				cmdStr += " " + arg
			}
			callback([]byte("执行命令: " + cmdStr))


			execCmd.Stdout = os.Stdout
			execCmd.Stderr = os.Stderr

			if err := execCmd.Run(); err != nil {
				mu.Lock()
				failedCount++
				failedFiles = append(failedFiles, path+" (错误: "+err.Error()+")")
				mu.Unlock()
				callback([]byte(fmt.Sprintf("❌ [%d/%d] 失败: %s", idx+1, len(videoFiles), filepath.Base(path))))
			} else {
				mu.Lock()
				successCount++
				mu.Unlock()
				callback([]byte(fmt.Sprintf("✅ [%d/%d] 完成: %s", idx+1, len(videoFiles), filepath.Base(path))))
			}
		}(i, videoPath)
	}

	wg.Wait()

	elapsed := time.Since(startTime)
	totalCost := fmt.Sprintf("%.1f秒", elapsed.Seconds())

	callback([]byte(fmt.Sprintf("提取完成！成功: %d，失败: %d，总耗时: %s", successCount, failedCount, totalCost)))

	return BatchExtractAudioRes{
		Success:      failedCount == 0,
		Message:      fmt.Sprintf("提取完成，成功: %d，失败: %d", successCount, failedCount),
		TotalCount:   len(videoFiles),
		SuccessCount: successCount,
		FailedCount:  failedCount,
		FailedFiles:  failedFiles,
		OutputDir:    outputDir,
		TotalCost:    totalCost,
	}
}

// RotateVideoReq 视频旋转请求
type RotateVideoReq struct {
	DirPath     string `json:"dirPath"`
	FileName    string `json:"fileName"`
	Angle       int    `json:"angle"`       // 旋转角度: 90, 180, 270
	Clockwise   bool   `json:"clockwise"`   // true: 顺时针, false: 逆时针
	FastRotate  bool   `json:"fastRotate"`  // true: 快速旋转（只转头部，无损）
}

// RotateVideoRes 视频旋转结果
type RotateVideoRes struct {
	Success    bool   `json:"success"`
	OutputPath string `json:"outputPath"`
	Message    string `json:"message"`
	Cost       string `json:"cost"`
}

// RotateVideo 旋转视频文件
func RotateVideo(req RotateVideoReq, callback func([]byte)) RotateVideoRes {
	startTime := time.Now()
	filePath := filepath.Join(req.DirPath, req.FileName)
	callback([]byte("开始旋转视频: " + filePath))

	ext := strings.ToLower(filepath.Ext(filePath))
	basePath := filePath[:len(filePath)-len(ext)]

	direction := "clockwise"
	if !req.Clockwise {
		direction = "anticlockwise"
	}

	var vfFilter string
	var outputPath string
	rotateDegree := 0

	switch req.Angle {
	case 90:
		if req.Clockwise {
			vfFilter = "transpose=1"
			rotateDegree = 90
		} else {
			vfFilter = "transpose=2"
			rotateDegree = 270
		}
		outputPath = fmt.Sprintf("%s_rotate90_%s%s", basePath, direction, ext)
	case 180:
		vfFilter = "transpose=1,transpose=1"
		rotateDegree = 180
		outputPath = fmt.Sprintf("%s_rotate180%s", basePath, ext)
	case 270:
		if req.Clockwise {
			vfFilter = "transpose=2"
			rotateDegree = 270
		} else {
			vfFilter = "transpose=1"
			rotateDegree = 90
		}
		outputPath = fmt.Sprintf("%s_rotate270_%s%s", basePath, direction, ext)
	default:
		return RotateVideoRes{
			Success: false,
			Message: "不支持的旋转角度，请选择90、180或270度",
		}
	}

	callback([]byte(fmt.Sprintf("输出文件: %s", outputPath)))
	callback([]byte(fmt.Sprintf("模式: %s", map[bool]string{true: "快速旋转（无损）", false: "普通旋转（重新编码)"}[req.FastRotate])))

	var execCmd *exec.Cmd

	if req.FastRotate {
		if !isFastRotateExtSupported(ext) {
			return RotateVideoRes{
				Success: false,
				Message: "当前格式不支持可靠的快速旋转，仅支持 MP4/MOV/M4V/3GP，请关闭快速旋转",
			}
		}

		displayRotation := clockwiseDegreesToDisplayRotation(rotateDegree)

		// 快速旋转：只修改显示矩阵metadata，不重新编码
		execCmd = core.Command("ffmpeg", "-display_rotation:v:0", strconv.Itoa(displayRotation), "-i", filePath, "-map", "0", "-c", "copy", "-y", outputPath)
		rotateDegree = displayRotation
	} else {
		// 普通旋转：重新编码
		execCmd = core.Command("ffmpeg", "-i", filePath, "-vf", vfFilter, "-c:a", "copy", "-y", outputPath)
	}

	// 记录完整命令行
	cmdStr := "ffmpeg"
	for _, arg := range execCmd.Args {
		cmdStr += " " + arg
	}
	callback([]byte("执行命令: " + cmdStr))

	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	if err := execCmd.Run(); err != nil {
		log.Printf("ffmpeg 执行失败: %v", err)
		return RotateVideoRes{
			Success: false,
			Message: "ffmpeg 执行失败: " + err.Error(),
		}
	}

	if req.FastRotate {
		ok, reason := verifyVideoRotateMetadata(outputPath, rotateDegree)
		if !ok {
			// 清理临时文件
			os.Remove(outputPath)
			return RotateVideoRes{
				Success: false,
				OutputPath: outputPath,
				Message: "快速旋转未生效: " + reason + "，请关闭快速旋转使用普通旋转",
			}
		}
	}

	// 旋转成功，删除原文件
	callback([]byte("删除原文件: " + filePath))
	if err := os.Remove(filePath); err != nil {
		callback([]byte("警告: 删除原文件失败: " + err.Error() + "，输出文件保留在: " + outputPath))
	} else {
		// 重命名输出文件为原文件名
		callback([]byte("重命名文件: " + outputPath + " -> " + filePath))
		if err := os.Rename(outputPath, filePath); err != nil {
			callback([]byte("警告: 重命名失败: " + err.Error() + "，输出文件保留在: " + outputPath))
		} else {
			outputPath = filePath
		}
	}

	elapsed := time.Since(startTime)
	callback([]byte("旋转完成，耗时: " + elapsed.String()))

	return RotateVideoRes{
		Success:    true,
		OutputPath: filePath,
		Message:    "视频旋转完成，已替换原文件",
		Cost:       elapsed.String(),
	}
}

func clockwiseDegreesToDisplayRotation(clockwise int) int {
	clockwise = normalizeRotation(clockwise)
	if clockwise == 0 {
		return 0
	}
	return -clockwise
}

func isFastRotateExtSupported(ext string) bool {
	switch strings.ToLower(ext) {
	case ".mp4", ".mov", ".m4v", ".3gp":
		return true
	default:
		return false
	}
}

func verifyVideoRotateMetadata(filePath string, expected int) (bool, string) {
	cmd := core.Command("ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_streams",
		"-of", "json",
		filePath)


	output, err := cmd.Output()
	if err != nil {
		return false, "无法校验旋转元数据: " + err.Error()
	}

	var result struct {
		Streams []struct {
			Tags map[string]string `json:"tags"`
			SideDataList []struct {
				Rotation float64 `json:"rotation"`
			} `json:"side_data_list"`
		} `json:"streams"`
	}

	if err := json.Unmarshal(output, &result); err != nil {
		return false, "解析旋转元数据失败"
	}

	expectedNorm := normalizeRotation(expected)
	for _, stream := range result.Streams {
		if rotateStr, ok := stream.Tags["rotate"]; ok {
			if rotateValue, err := strconv.Atoi(strings.TrimSpace(rotateStr)); err == nil && normalizeRotation(rotateValue) == expectedNorm {
				return true, ""
			}
		}
		for _, sideData := range stream.SideDataList {
			if normalizeRotation(int(sideData.Rotation)) == expectedNorm {
				return true, ""
			}
		}
	}

	return false, "未检测到期望旋转元数据"
}

func normalizeRotation(value int) int {
	value %= 360
	if value < 0 {
		value += 360
	}
	return value
}

// ExtractFramesByFilesReq 批量抽帧请求
type ExtractFramesByFilesReq struct {
	DirPath   string   `json:"dirPath"`
	FileNames []string `json:"fileNames"`
	Count     int      `json:"count"` // 每个视频抽帧数量
}

// ExtractFramesByFilesRes 批量抽帧结果
type ExtractFramesByFilesRes struct {
	Success      bool     `json:"success"`
	Message      string   `json:"message"`
	OutputDir    string   `json:"outputDir"`
	TotalCount   int      `json:"totalCount"`
	SuccessCount int      `json:"successCount"`
	FailedCount  int      `json:"failedCount"`
	FailedFiles  []string `json:"failedFiles"`
	TotalCost    string   `json:"totalCost"`
}

// ExtractFramesByFiles 批量抽帧
func ExtractFramesByFiles(req ExtractFramesByFilesReq, callback func([]byte)) ExtractFramesByFilesRes {
	startTime := time.Now()
	callback([]byte(fmt.Sprintf("开始批量抽帧，共 %d 个视频，每视频抽 %d 帧", len(req.FileNames), req.Count)))

	if req.Count < 1 || req.Count > 100 {
		return ExtractFramesByFilesRes{Success: false, Message: "抽取数量必须在 1-100 之间"}
	}

	outputDir := filepath.Join(req.DirPath, "frames")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return ExtractFramesByFilesRes{Success: false, Message: "创建输出目录失败：" + err.Error()}
	}

	var mu sync.Mutex
	var successCount, failedCount int
	var failedFiles []string

	for i, fileName := range req.FileNames {
		callback([]byte(fmt.Sprintf("[%d/%d] 处理中: %s", i+1, len(req.FileNames), fileName)))

		filePath := filepath.Join(req.DirPath, fileName)

		// 获取视频时长
		duration, err := getVideoDuration(filePath)
		if err != nil {
			mu.Lock()
			failedCount++
			failedFiles = append(failedFiles, fileName+" (获取时长失败)")
			mu.Unlock()
			callback([]byte(fmt.Sprintf("❌ [%d/%d] 获取时长失败: %s", i+1, len(req.FileNames), fileName)))
			continue
		}

		// 生成随机时间点
		timestamps := generateRandomTimestamps(duration, req.Count)

		// 处理每个时间点
		frameBaseName := fileName
		if idx := strings.LastIndex(fileName, "."); idx > 0 {
			frameBaseName = fileName[:idx]
		}
		frameDir := filepath.Join(outputDir, frameBaseName)
		if err := os.MkdirAll(frameDir, 0755); err != nil {
			mu.Lock()
			failedCount++
			failedFiles = append(failedFiles, fileName+" (创建帧目录失败)")
			mu.Unlock()
			continue
		}

		framesSaved := 0
		for j, ts := range timestamps {
			framePath := filepath.Join(frameDir, fmt.Sprintf("%s_%03d.jpg", frameBaseName, j+1))

			cmd := core.Command("ffmpeg",
				"-ss", fmt.Sprintf("%.3f", ts),
				"-i", filePath,
				"-vf", "select=eq(pict_type\\,I)",
				"-vsync", "0",
				"-vframes", "1",
				"-q:v", "2",
				"-y",
				framePath)

			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			if err := cmd.Run(); err == nil {
				framesSaved++
			}
		}

		mu.Lock()
		if framesSaved > 0 {
			successCount++
			callback([]byte(fmt.Sprintf("✅ [%d/%d] 完成: %s (%d帧)", i+1, len(req.FileNames), fileName, framesSaved)))
		} else {
			_ = os.RemoveAll(frameDir)
			failedCount++
			failedFiles = append(failedFiles, fileName+" (抽帧失败)")
			callback([]byte(fmt.Sprintf("❌ [%d/%d] 失败: %s", i+1, len(req.FileNames), fileName)))
		}
		mu.Unlock()
	}

	elapsed := time.Since(startTime)
	callback([]byte(fmt.Sprintf("抽帧完成！成功: %d，失败: %d，总耗时: %s", successCount, failedCount, elapsed.String())))

	return ExtractFramesByFilesRes{
		Success:      failedCount == 0,
		Message:      fmt.Sprintf("成功: %d，失败: %d", successCount, failedCount),
		OutputDir:    outputDir,
		TotalCount:   len(req.FileNames),
		SuccessCount: successCount,
		FailedCount:  failedCount,
		FailedFiles:  failedFiles,
		TotalCost:    elapsed.String(),
	}
}

// ExtractAudioReq 声音分离请求
type ExtractAudioReq struct {
	FilePath string `json:"filePath"`
	Format   string `json:"format"` // 输出格式: mp3, aac, wav, flac, 默认 mp3
}

// ExtractAudioRes 声音分离结果
type ExtractAudioRes struct {
	Success    bool   `json:"success"`
	OutputPath string `json:"outputPath"`
	Message    string `json:"message"`
	Cost       string `json:"cost"`
}

// ExtractAudio 从视频中分离音频
func ExtractAudio(req ExtractAudioReq, callback func([]byte)) ExtractAudioRes {
	startTime := time.Now()
	callback([]byte("开始提取音频: " + req.FilePath))

	if req.FilePath == "" {
		return ExtractAudioRes{Success: false, Message: "请选择视频文件"}
	}

	format := req.Format
	if format == "" {
		format = "mp3"
	}

	// 生成输出路径
	basePath := req.FilePath
	if idx := strings.LastIndex(req.FilePath, "."); idx != -1 {
		basePath = req.FilePath[:idx]
	}
	outputPath := fmt.Sprintf("%s_audio.%s", basePath, format)

	callback([]byte(fmt.Sprintf("输出格式: %s，输出路径: %s", format, outputPath)))

	// 使用 exec.Command 直接传参，避免路径含中文/空格解析失败
	// -vn: 不包含视频流
	// -acodec: 指定音频编解码器
	var execCmd *exec.Cmd
	switch format {
	case "wav":
		execCmd = core.Command("ffmpeg", "-i", req.FilePath, "-vn", "-acodec", "pcm_s16le", "-y", outputPath)
	case "aac":
		execCmd = core.Command("ffmpeg", "-i", req.FilePath, "-vn", "-acodec", "copy", "-y", outputPath)
	case "flac":
		execCmd = core.Command("ffmpeg", "-i", req.FilePath, "-vn", "-acodec", "flac", "-y", outputPath)
	default: // mp3
		execCmd = core.Command("ffmpeg", "-i", req.FilePath, "-vn", "-acodec", "libmp3lame", "-q:a", "2", "-y", outputPath)
	}

	// 记录完整命令行
	cmdStr := "ffmpeg"
	for _, arg := range execCmd.Args {
		cmdStr += " " + arg
	}
	callback([]byte("执行命令: " + cmdStr))

	// Windows 下隐藏命令行窗口

	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr
	if err := execCmd.Run(); err != nil {
		log.Printf("ffmpeg 音频提取失败: %v", err)
		return ExtractAudioRes{
			Success: false,
			Message: "ffmpeg 音频提取失败: " + err.Error(),
		}
	}

	elapsed := time.Since(startTime)
	callback([]byte("音频提取完成，耗时: " + elapsed.String()))

	return ExtractAudioRes{
		Success:    true,
		OutputPath: outputPath,
		Message:    "音频提取完成",
		Cost:       elapsed.String(),
	}
}

// VideoInfo 视频文件信息
type VideoInfo struct {
	FilePath    string  `json:"filePath"`    // 完整路径
	FileName    string  `json:"fileName"`    // 文件名（含扩展名）
	FileSizeMB  float64 `json:"fileSizeMB"`  // 文件大小 MB
	Resolution  string  `json:"resolution"` // 分辨率：宽 x 高
	Width       int     `json:"width"`       // 宽度
	Height      int     `json:"height"`      // 高度
	FrameRate   string  `json:"frameRate"`   // 帧率
	BitRate     string  `json:"bitRate"`     // 码率
	Codec       string  `json:"codec"`       // 视频编码
	Duration    float64 `json:"duration"`    // 时长（秒）
}

// ScanVideosReq 扫描视频请求
type ScanVideosReq struct {
	DirPath string `json:"dirPath"` // 要扫描的目录路径
}

// ScanVideosRes 扫描视频结果
type ScanVideosRes struct {
	Success        bool        `json:"success"`
	Message        string      `json:"message"`
	AllVideos      []VideoInfo `json:"allVideos"`      // 所有视频列表
	VerticalVideos []VideoInfo `json:"verticalVideos"` // 竖屏视频（高 > 宽）
}

// ScanVideos 扫描指定目录中的所有视频文件（不扫描子目录）
func ScanVideos(req ScanVideosReq, callback func([]byte)) ScanVideosRes {
	callback([]byte("开始扫描目录：" + req.DirPath))

	if req.DirPath == "" {
		return ScanVideosRes{
			Success: false,
			Message: "请选择要扫描的目录",
		}
	}

	if !core.FileExist(req.DirPath) {
		return ScanVideosRes{
			Success: false,
			Message: "目录不存在：" + req.DirPath,
		}
	}

	// 常见视频格式扩展名
	videoExts := map[string]bool{
		".mp4":  true,
		".avi":  true,
		".mkv":  true,
		".mov":  true,
		".wmv":  true,
		".flv":  true,
		".m4v":  true,
		".ts":   true,
		".rmvb": true,
		".webm": true,
	}

	var allVideos []VideoInfo
	var verticalVideos []VideoInfo

	// 只遍历当前目录，不递归子目录
	entries, err := os.ReadDir(req.DirPath)
	if err != nil {
		return ScanVideosRes{
			Success: false,
			Message: "无法读取目录：" + err.Error(),
		}
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue // 跳过子目录
		}

		fileName := entry.Name()
		ext := strings.ToLower(filepath.Ext(fileName))
		if !videoExts[ext] {
			continue // 只处理视频文件
		}

		filePath := filepath.Join(req.DirPath, fileName)

		// 获取文件信息
		info, err := entry.Info()
		if err != nil {
			log.Printf("无法获取文件信息：%s, 错误：%v", filePath, err)
			continue
		}

		// 获取文件大小（MB）
		fileSizeMB := float64(info.Size()) / 1024.0 / 1024.0

		// 使用 ffprobe 获取视频详细信息
		detail, err := getVideoDetail(filePath)
		if err != nil {
			log.Printf("无法获取视频信息：%s, 错误：%v", filePath, err)
			continue
		}

		resolution := fmt.Sprintf("%dx%d", detail.Width, detail.Height)

		videoInfo := VideoInfo{
			FilePath:   filePath,
			FileName:   fileName,
			FileSizeMB: fileSizeMB,
			Resolution: resolution,
			Width:      detail.Width,
			Height:     detail.Height,
			FrameRate:  detail.FrameRate,
			BitRate:    detail.BitRate,
			Codec:      detail.Codec,
			Duration:   detail.Duration,
		}

		allVideos = append(allVideos, videoInfo)

		// 判断是否为竖屏视频（高 > 宽）
		isVertical := ""
		if detail.Height > detail.Width {
			verticalVideos = append(verticalVideos, videoInfo)
			isVertical = " [竖屏]"
		}

		// 每扫描一个文件，发送一条日志到前端
		logMsg := fmt.Sprintf("找到：%s | 大小：%.2f MB | 分辨率：%s | 帧率：%s | 码率：%s | 编码：%s | 时长：%.1f秒%s",
			fileName, fileSizeMB, resolution, detail.FrameRate, detail.BitRate, detail.Codec, detail.Duration, isVertical)
		callback([]byte(logMsg))
	}

	if err != nil {
		return ScanVideosRes{
			Success: false,
			Message: "扫描目录失败：" + err.Error(),
		}
	}

	callback([]byte(fmt.Sprintf("扫描完成，共找到 %d 个视频文件，其中竖屏视频 %d 个", len(allVideos), len(verticalVideos))))

	return ScanVideosRes{
		Success:        true,
		Message:        fmt.Sprintf("找到 %d 个视频文件", len(allVideos)),
		AllVideos:      allVideos,
		VerticalVideos: verticalVideos,
	}
}

// VideoDetailInfo 视频详细信息
type VideoDetailInfo struct {
	Width     int     // 宽度
	Height    int     // 高度
	FrameRate string  // 帧率
	BitRate   string  // 码率
	Codec     string  // 视频编码
	Duration  float64 // 时长（秒）
}

// getVideoDetail 使用 ffprobe 获取视频的详细信息
func getVideoDetail(filePath string) (*VideoDetailInfo, error) {
	cmd := core.Command("ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_streams",
		"-show_format",
		filePath)

	// Windows 下隐藏命令行窗口

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, err
	}

	info := &VideoDetailInfo{}

	streams, ok := result["streams"].([]interface{})
	if !ok || len(streams) == 0 {
		return nil, fmt.Errorf("未找到视频流")
	}

	// 查找第一个视频流
	for _, s := range streams {
		stream, ok := s.(map[string]interface{})
		if !ok {
			continue
		}
		codecType, _ := stream["codec_type"].(string)
		if codecType == "video" {
			if widthFloat, ok := stream["width"].(float64); ok {
				info.Width = int(widthFloat)
			}
			if heightFloat, ok := stream["height"].(float64); ok {
				info.Height = int(heightFloat)
			}
			if codec, ok := stream["codec_name"].(string); ok {
				info.Codec = codec
			}
			// 帧率
			if rFrameRate, ok := stream["r_frame_rate"].(string); ok {
				info.FrameRate = rFrameRate
			}
			if bitRate, ok := stream["bit_rate"].(string); ok {
				info.BitRate = formatBitRate(bitRate)
			}
		}
	}

	// 获取时长
	if format, ok := result["format"].(map[string]interface{}); ok {
		if durationStr, ok := format["duration"].(string); ok {
			if duration, err := strconv.ParseFloat(durationStr, 64); err == nil {
				info.Duration = duration
			}
		}
	}

	return info, nil
}

// formatBitRate 将比特率转换为可读格式
func formatBitRate(bitRate string) string {
	br, err := strconv.ParseFloat(bitRate, 64)
	if err != nil {
		return bitRate
	}
	if br >= 1000000 {
		return fmt.Sprintf("%.2f Mbps", br/1000000)
	} else if br >= 1000 {
		return fmt.Sprintf("%.2f Kbps", br/1000)
	}
	return bitRate
}

// ExtractFramesReq 抽帧请求
type ExtractFramesReq struct {
	FilePath string `json:"filePath"` // 视频文件路径
	Count    int    `json:"count"`    // 抽取帧数量
}

// ExtractFramesRes 抽帧结果
type ExtractFramesRes struct {
	Success    bool     `json:"success"`
	Message    string   `json:"message"`
	OutputDir  string   `json:"outputDir"`  // 输出目录
	FramePaths []string `json:"framePaths"` // 所有帧的路径
	FrameCount int      `json:"frameCount"` // 实际抽取的帧数
	Cost       string   `json:"cost"`       // 耗时
}

// ExtractFrames 从视频中随机抽取 N 帧
func ExtractFrames(req ExtractFramesReq, callback func([]byte)) ExtractFramesRes {
	callback([]byte("开始抽取视频帧：" + req.FilePath))

	if req.FilePath == "" {
		return ExtractFramesRes{
			Success: false,
			Message: "请选择视频文件",
		}
	}

	if req.Count < 1 || req.Count > 100 {
		return ExtractFramesRes{
			Success: false,
			Message: "抽取数量必须在 1-100 之间",
		}
	}

	startTime := time.Now()

	// 获取视频时长
	duration, err := getVideoDuration(req.FilePath)
	if err != nil {
		return ExtractFramesRes{
			Success: false,
			Message: "获取视频时长失败：" + err.Error(),
		}
	}

	callback([]byte(fmt.Sprintf("视频时长：%.2f 秒", duration)))

	// 生成输出目录
	outputDir := req.FilePath + "_frames"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return ExtractFramesRes{
			Success: false,
			Message: "创建输出目录失败：" + err.Error(),
		}
	}

	// 生成随机时间点
	timestamps := generateRandomTimestamps(duration, req.Count)
	callback([]byte(fmt.Sprintf("生成 %d 个随机时间点", len(timestamps))))

	var framePaths []string

	// 抽取每一帧
	for i, ts := range timestamps {
		framePath := filepath.Join(outputDir, fmt.Sprintf("frame_%03d.jpg", i+1))

		// 使用 ffmpeg 抽取单帧
		cmd := core.Command("ffmpeg",
			"-ss", fmt.Sprintf("%.3f", ts),
			"-i", req.FilePath,
			"-vf", "select=eq(pict_type\\,I)", // 视频过滤器：只选择帧类型为I的帧
			"-vsync", "0", // 防止帧数同步
			"-vframes", "1",
			"-q:v", "2",
			"-y",
			framePath)

		// Windows 下隐藏命令行窗口

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			log.Printf("抽取第 %d 帧失败：%v", i+1, err)
			continue
		}

		framePaths = append(framePaths, framePath)
		callback([]byte(fmt.Sprintf("✅ 已抽取第 %d/%d 帧", i+1, len(timestamps))))
	}

	elapsed := time.Since(startTime)
	frameCount := len(framePaths)
	success := frameCount > 0
	if success {
		callback([]byte(fmt.Sprintf("抽帧完成，共提取 %d 张图片，耗时：%s", frameCount, elapsed.String())))
	} else {
		callback([]byte(fmt.Sprintf("抽帧完成，但未提取到关键帧，耗时：%s", elapsed.String())))
	}

	message := fmt.Sprintf("成功提取 %d 张图片", frameCount)
	if !success {
		message = "未提取到关键帧"
	}

	return ExtractFramesRes{
		Success:    success,
		Message:    message,
		OutputDir:  outputDir,
		FramePaths: framePaths,
		FrameCount: frameCount,
		Cost:       elapsed.String(),
	}
}

// getVideoDuration 使用 ffprobe 获取视频时长（秒）
func getVideoDuration(filePath string) (float64, error) {
	cmd := core.Command("ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		filePath)

	// Windows 下隐藏命令行窗口

	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return 0, err
	}

	format, ok := result["format"].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("无法解析视频格式信息")
	}

	durationStr, _ := format["duration"].(string)
	if durationStr == "" {
		return 0, fmt.Errorf("无法获取视频时长")
	}

	parsedDuration, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, fmt.Errorf("解析视频时长失败: %w", err)
	}
	return parsedDuration, nil
}

// generateRandomTimestamps 生成 N 个随机的时间点（秒）
func generateRandomTimestamps(duration float64, count int) []float64 {
	rand.Seed(time.Now().UnixNano())
	timestamps := make([]float64, count)

	if duration <= 0 {
		return timestamps
	}

	start := duration * 0.05
	end := duration * 0.95
	if end <= start {
		mid := duration / 2
		for i := 0; i < count; i++ {
			timestamps[i] = mid
		}
		return timestamps
	}

	segment := (end - start) / float64(count)
	if segment <= 0 {
		for i := 0; i < count; i++ {
			timestamps[i] = start
		}
		return timestamps
	}

	for i := 0; i < count; i++ {
		left := start + float64(i)*segment
		right := left + segment
		timestamps[i] = left + rand.Float64()*(right-left)
	}

	sort.Float64s(timestamps)
	return timestamps
}

// ExtractVideoThumbnailReq 提取视频缩略图请求
type ExtractVideoThumbnailReq struct {
	FilePath  string  `json:"filePath"`  // 视频文件路径
	Timestamp float64 `json:"timestamp"` // 提取帧的时间点（秒），默认 1.0
}

// ExtractVideoThumbnailRes 提取视频缩略图结果
type ExtractVideoThumbnailRes struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Thumbnail string `json:"thumbnail"` // Base64 编码的缩略图
	MimeType  string `json:"mimeType"`  // 图片 MIME 类型
}

// ExtractVideoThumbnail 从视频中提取一帧作为缩略图
func ExtractVideoThumbnail(req ExtractVideoThumbnailReq) ExtractVideoThumbnailRes {
	if req.FilePath == "" {
		return ExtractVideoThumbnailRes{
			Success: false,
			Message: "请选择视频文件",
		}
	}

	if !core.FileExist(req.FilePath) {
		return ExtractVideoThumbnailRes{
			Success: false,
			Message: "视频文件不存在",
		}
	}

	// 默认从第1秒提取
	timestamp := req.Timestamp
	if timestamp <= 0 {
		timestamp = 1.0
	}

	duration, err := getVideoDuration(req.FilePath)
	if err != nil {
		return ExtractVideoThumbnailRes{
			Success: false,
			Message: "获取视频时长失败: " + err.Error(),
		}
	}

	if duration <= 0 {
		return ExtractVideoThumbnailRes{
			Success: false,
			Message: "视频时长无效",
		}
	}

	if timestamp < 0 {
		timestamp = 0
	}
	maxTimestamp := duration - 0.001
	if maxTimestamp < 0 {
		maxTimestamp = 0
	}
	if timestamp > maxTimestamp {
		timestamp = maxTimestamp
	}

	// 生成临时缩略图文件路径
	tempDir := core.GetTempDir()
	thumbnailPath := filepath.Join(tempDir, fmt.Sprintf("thumb_%d.jpg", time.Now().UnixNano()))
	defer os.Remove(thumbnailPath)

	// 使用 ffmpeg 提取单帧
	cmd := core.Command("ffmpeg",
		"-ss", fmt.Sprintf("%.3f", timestamp),
		"-i", req.FilePath,
		"-vframes", "1",
		"-q:v", "2",
		"-y",
		thumbnailPath)

	// Windows 下隐藏命令行窗口

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Printf("缩略图提取失败: %v", err)
		return ExtractVideoThumbnailRes{
			Success: false,
			Message: "缩略图提取失败: " + err.Error(),
		}
	}

	// 读取缩略图并转为 base64
	fileData, err := os.ReadFile(thumbnailPath)
	if err != nil {
		log.Printf("缩略图读取失败: %v", err)
		return ExtractVideoThumbnailRes{
			Success: false,
			Message: "缩略图读取失败: " + err.Error(),
		}
	}

	// base64 编码
	encoded := base64.StdEncoding.EncodeToString(fileData)

	return ExtractVideoThumbnailRes{
		Success:   true,
		Message:   "缩略图提取成功",
		Thumbnail: encoded,
		MimeType:  "image/jpeg",
	}
}

// TrimVideoStartReq 去除片头请求
type TrimVideoStartReq struct {
	DirPath  string  `json:"dirPath"`  // 视频目录路径
	FileName string  `json:"fileName"` // 文件名
	Duration float64 `json:"duration"` // 要切除的片头时长（秒）
}

// TrimVideoStartRes 去除片头结果
type TrimVideoStartRes struct {
	Success    bool   `json:"success"`
	OutputPath string `json:"outputPath"`
	Message    string `json:"message"`
	Cost       string `json:"cost"`
}

// TrimVideoStart 去除视频片头（无损切割）
func TrimVideoStart(req TrimVideoStartReq, callback func([]byte)) TrimVideoStartRes {
	startTime := time.Now()
	filePath := filepath.Join(req.DirPath, req.FileName)
	callback([]byte("开始处理视频: " + filePath))

	if req.Duration <= 0 {
		return TrimVideoStartRes{
			Success: false,
			Message: "片头时长必须大于0",
		}
	}

	// 获取视频总时长
	duration, err := getVideoDuration(filePath)
	if err != nil {
		return TrimVideoStartRes{
			Success: false,
			Message: "获取视频时长失败: " + err.Error(),
		}
	}

	if req.Duration >= duration {
		return TrimVideoStartRes{
			Success: false,
			Message: "片头时长不能大于视频总时长",
		}
	}

	remainDuration := duration - req.Duration
	callback([]byte(fmt.Sprintf("视频总时长: %.2f 秒", duration)))
	callback([]byte(fmt.Sprintf("要去除片头: %.2f 秒", req.Duration)))
	callback([]byte(fmt.Sprintf("保留时长: %.2f 秒", remainDuration)))

	ext := strings.ToLower(filepath.Ext(filePath))
	basePath := filePath[:len(filePath)-len(ext)]
	outputPath := fmt.Sprintf("%s_cut%s", basePath, ext)

	// 无损切割：-ss 放在 -i 之前会自动从最近关键帧开始，-avoid_negative_ts 处理时间戳
	execCmd := core.Command("ffmpeg",
		"-ss", fmt.Sprintf("%.3f", req.Duration),
		"-i", filePath,
		"-c", "copy",
		"-avoid_negative_ts", "make_zero",
		"-y",
		outputPath)

	// 记录完整命令行
	cmdStr := "ffmpeg"
	for _, arg := range execCmd.Args {
		cmdStr += " " + arg
	}
	callback([]byte("执行命令: " + cmdStr))

	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	if err := execCmd.Run(); err != nil {
		log.Printf("ffmpeg 执行失败: %v", err)
		return TrimVideoStartRes{
			Success: false,
			Message: "ffmpeg 执行失败: " + err.Error(),
		}
	}

	// 处理成功，删除原文件
	callback([]byte("删除原文件: " + filePath))
	if err := os.Remove(filePath); err != nil {
		callback([]byte("警告: 删除原文件失败: " + err.Error() + "，输出文件保留在: " + outputPath))
	} else {
		// 重命名输出文件为原文件名
		callback([]byte("重命名文件: " + outputPath + " -> " + filePath))
		if err := os.Rename(outputPath, filePath); err != nil {
			callback([]byte("警告: 重命名失败: " + err.Error() + "，输出文件保留在: " + outputPath))
		} else {
			outputPath = filePath
		}
	}

	elapsed := time.Since(startTime)
	callback([]byte("处理完成，耗时: " + elapsed.String()))

	return TrimVideoStartRes{
		Success:    true,
		OutputPath: filePath,
		Message:    "片头去除完成，已替换原文件",
		Cost:       elapsed.String(),
	}
}

// TrimVideoEndReq 去除片尾请求
type TrimVideoEndReq struct {
	DirPath    string  `json:"dirPath"`    // 视频目录路径
	FileName   string  `json:"fileName"`  // 文件名
	Duration   float64 `json:"duration"`  // 要切除的片尾时长（秒）
}

// TrimVideoEndRes 去除片尾结果
type TrimVideoEndRes struct {
	Success    bool   `json:"success"`
	OutputPath string `json:"outputPath"`
	Message    string `json:"message"`
	Cost       string `json:"cost"`
}

// TrimVideoEnd 去除视频片尾（无损切割）
func TrimVideoEnd(req TrimVideoEndReq, callback func([]byte)) TrimVideoEndRes {
	startTime := time.Now()
	filePath := filepath.Join(req.DirPath, req.FileName)
	callback([]byte("开始处理视频: " + filePath))

	if req.Duration <= 0 {
		return TrimVideoEndRes{
			Success: false,
			Message: "片尾时长必须大于0",
		}
	}

	// 获取视频总时长
	duration, err := getVideoDuration(filePath)
	if err != nil {
		return TrimVideoEndRes{
			Success: false,
			Message: "获取视频时长失败: " + err.Error(),
		}
	}

	if req.Duration >= duration {
		return TrimVideoEndRes{
			Success: false,
			Message: "片尾时长不能大于视频总时长",
		}
	}

	newDuration := duration - req.Duration
	callback([]byte(fmt.Sprintf("视频总时长: %.2f 秒", duration)))
	callback([]byte(fmt.Sprintf("要去除片尾: %.2f 秒", req.Duration)))
	callback([]byte(fmt.Sprintf("保留时长: %.2f 秒", newDuration)))

	ext := strings.ToLower(filepath.Ext(filePath))
	basePath := filePath[:len(filePath)-len(ext)]
	outputPath := fmt.Sprintf("%s_cut%s", basePath, ext)

	// 无损切割：使用 -c copy 直接复制流，不重新编码
	execCmd := core.Command("ffmpeg",
		"-i", filePath,
		"-t", fmt.Sprintf("%.3f", newDuration),
		"-c", "copy",
		"-y",
		outputPath)

	// 记录完整命令行
	cmdStr := "ffmpeg"
	for _, arg := range execCmd.Args {
		cmdStr += " " + arg
	}
	callback([]byte("执行命令: " + cmdStr))

	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	if err := execCmd.Run(); err != nil {
		log.Printf("ffmpeg 执行失败: %v", err)
		return TrimVideoEndRes{
			Success: false,
			Message: "ffmpeg 执行失败: " + err.Error(),
		}
	}

	// 处理成功，删除原文件
	callback([]byte("删除原文件: " + filePath))
	if err := os.Remove(filePath); err != nil {
		callback([]byte("警告: 删除原文件失败: " + err.Error() + "，输出文件保留在: " + outputPath))
	} else {
		// 重命名输出文件为原文件名
		callback([]byte("重命名文件: " + outputPath + " -> " + filePath))
		if err := os.Rename(outputPath, filePath); err != nil {
			callback([]byte("警告: 重命名失败: " + err.Error() + "，输出文件保留在: " + outputPath))
		} else {
			outputPath = filePath
		}
	}

	elapsed := time.Since(startTime)
	callback([]byte("处理完成，耗时: " + elapsed.String()))

	return TrimVideoEndRes{
		Success:    true,
		OutputPath: filePath,
		Message:    "片尾去除完成，已替换原文件",
		Cost:       elapsed.String(),
	}
}

// TrimVideoEndByFilesReq 批量去除片尾请求
type TrimVideoEndByFilesReq struct {
	DirPath     string  `json:"dirPath"`     // 视频目录路径
	FileNames   []string `json:"fileNames"` // 要处理的文件名列表
	Duration    float64  `json:"duration"`   // 要切除的片尾时长（秒）
	ThreadCount int     `json:"threadCount"` // 并行线程数，默认 4
}

// TrimVideoEndByFilesRes 批量去除片尾结果
type TrimVideoEndByFilesRes struct {
	Success      bool     `json:"success"`
	Message      string   `json:"message"`
	TotalCount   int      `json:"totalCount"`
	SuccessCount int      `json:"successCount"`
	FailedCount  int      `json:"failedCount"`
	FailedFiles  []string `json:"failedFiles"`
	TotalCost    string   `json:"totalCost"`
}

// TrimVideoEndByFiles 批量去除视频片尾（并发执行）
func TrimVideoEndByFiles(req TrimVideoEndByFilesReq, callback func([]byte)) TrimVideoEndByFilesRes {
	startTime := time.Now()
	callback([]byte(fmt.Sprintf("开始批量去除片尾，共 %d 个视频，时长 %.2f 秒", len(req.FileNames), req.Duration)))

	if req.DirPath == "" {
		return TrimVideoEndByFilesRes{Success: false, Message: "请选择视频目录"}
	}

	if !core.FileExist(req.DirPath) {
		return TrimVideoEndByFilesRes{Success: false, Message: "目录不存在：" + req.DirPath}
	}

	if len(req.FileNames) == 0 {
		return TrimVideoEndByFilesRes{Success: false, Message: "没有选择要处理的文件"}
	}

	if req.Duration <= 0 {
		return TrimVideoEndByFilesRes{Success: false, Message: "片尾时长必须大于0"}
	}

	threadCount := req.ThreadCount
	if threadCount <= 0 {
		threadCount = 4
	}
	callback([]byte(fmt.Sprintf("并行线程数: %d", threadCount)))

	var mu sync.Mutex
	var successCount, failedCount int
	var failedFiles []string
	var wg sync.WaitGroup

	// 控制并发数
	semaphore := make(chan struct{}, threadCount)

	for i, fileName := range req.FileNames {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(idx int, fname string) {
			defer wg.Done()
			defer func() { <-semaphore }()

			callback([]byte(fmt.Sprintf("[%d/%d] 正在处理: %s", idx+1, len(req.FileNames), fname)))

			filePath := filepath.Join(req.DirPath, fname)

			// 获取视频总时长
			duration, err := getVideoDuration(filePath)
			if err != nil {
				mu.Lock()
				failedCount++
				failedFiles = append(failedFiles, fname+" (获取时长失败)")
				mu.Unlock()
				callback([]byte(fmt.Sprintf("❌ [%d/%d] 获取时长失败: %s", idx+1, len(req.FileNames), fname)))
				return
			}

			if req.Duration >= duration {
				mu.Lock()
				failedCount++
				failedFiles = append(failedFiles, fname+" (片尾时长大于视频总时长)")
				mu.Unlock()
				callback([]byte(fmt.Sprintf("❌ [%d/%d] 片尾时长大于视频总时长: %s", idx+1, len(req.FileNames), fname)))
				return
			}

			newDuration := duration - req.Duration

			ext := strings.ToLower(filepath.Ext(filePath))
			basePath := filePath[:len(filePath)-len(ext)]
			outputPath := fmt.Sprintf("%s_cut%s", basePath, ext)

			// 无损切割
			execCmd := core.Command("ffmpeg",
				"-i", filePath,
				"-t", fmt.Sprintf("%.3f", newDuration),
				"-c", "copy",
				"-y",
				outputPath)

			// 记录完整命令行
			cmdStr := "ffmpeg"
			for _, arg := range execCmd.Args {
				cmdStr += " " + arg
			}
			callback([]byte("执行命令: " + cmdStr))

			execCmd.Stdout = os.Stdout
			execCmd.Stderr = os.Stderr

			if err := execCmd.Run(); err != nil {
				mu.Lock()
				failedCount++
				failedFiles = append(failedFiles, fname+" (处理失败: "+err.Error()+")")
				mu.Unlock()
				callback([]byte(fmt.Sprintf("❌ [%d/%d] 处理失败: %s", idx+1, len(req.FileNames), fname)))
				return
			}

			// 删除原文件并重命名，任何一步失败都不能算成功，
			// 否则原文件被删、结果留在 xxx_cut.mp4 却被计为成功
			if err := os.Remove(filePath); err != nil {
				mu.Lock()
				failedCount++
				failedFiles = append(failedFiles, fname+" (删除原文件失败: "+err.Error()+")")
				mu.Unlock()
				callback([]byte(fmt.Sprintf("❌ [%d/%d] 删除原文件失败: %s", idx+1, len(req.FileNames), fname)))
				return
			}
			if err := os.Rename(outputPath, filePath); err != nil {
				mu.Lock()
				failedCount++
				failedFiles = append(failedFiles, fname+" (重命名失败: "+err.Error()+")")
				mu.Unlock()
				callback([]byte(fmt.Sprintf("❌ [%d/%d] 重命名失败: %s", idx+1, len(req.FileNames), fname)))
				return
			}

			mu.Lock()
			successCount++
			mu.Unlock()
			callback([]byte(fmt.Sprintf("✅ [%d/%d] 完成: %s", idx+1, len(req.FileNames), fname)))
		}(i, fileName)
	}

	wg.Wait()

	elapsed := time.Since(startTime)
	callback([]byte(fmt.Sprintf("处理完成！成功: %d，失败: %d，总耗时: %s", successCount, failedCount, elapsed.String())))

	return TrimVideoEndByFilesRes{
		Success:      failedCount == 0,
		Message:      fmt.Sprintf("成功: %d，失败: %d", successCount, failedCount),
		TotalCount:   len(req.FileNames),
		SuccessCount: successCount,
		FailedCount:  failedCount,
		FailedFiles:  failedFiles,
		TotalCost:    elapsed.String(),
	}
}

// TrimVideoStartByFilesReq 批量去除片头请求
type TrimVideoStartByFilesReq struct {
	DirPath     string  `json:"dirPath"`     // 视频目录路径
	FileNames   []string `json:"fileNames"` // 要处理的文件名列表
	Duration    float64  `json:"duration"`   // 要切除的片头时长（秒）
	ThreadCount int     `json:"threadCount"` // 并行线程数，默认 4
}

// TrimVideoStartByFilesRes 批量去除片头结果
type TrimVideoStartByFilesRes struct {
	Success      bool     `json:"success"`
	Message      string   `json:"message"`
	TotalCount   int      `json:"totalCount"`
	SuccessCount int      `json:"successCount"`
	FailedCount  int      `json:"failedCount"`
	FailedFiles  []string `json:"failedFiles"`
	TotalCost    string   `json:"totalCost"`
}

// TrimVideoStartByFiles 批量去除视频片头（并发执行）
func TrimVideoStartByFiles(req TrimVideoStartByFilesReq, callback func([]byte)) TrimVideoStartByFilesRes {
	startTime := time.Now()
	callback([]byte(fmt.Sprintf("开始批量去除片头，共 %d 个视频，时长 %.2f 秒", len(req.FileNames), req.Duration)))

	if req.DirPath == "" {
		return TrimVideoStartByFilesRes{Success: false, Message: "请选择视频目录"}
	}

	if !core.FileExist(req.DirPath) {
		return TrimVideoStartByFilesRes{Success: false, Message: "目录不存在：" + req.DirPath}
	}

	if len(req.FileNames) == 0 {
		return TrimVideoStartByFilesRes{Success: false, Message: "没有选择要处理的文件"}
	}

	if req.Duration <= 0 {
		return TrimVideoStartByFilesRes{Success: false, Message: "片头时长必须大于0"}
	}

	threadCount := req.ThreadCount
	if threadCount <= 0 {
		threadCount = 4
	}
	callback([]byte(fmt.Sprintf("并行线程数: %d", threadCount)))

	var mu sync.Mutex
	var successCount, failedCount int
	var failedFiles []string
	var wg sync.WaitGroup

	// 控制并发数
	semaphore := make(chan struct{}, threadCount)

	for i, fileName := range req.FileNames {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(idx int, fname string) {
			defer wg.Done()
			defer func() { <-semaphore }()

			callback([]byte(fmt.Sprintf("[%d/%d] 正在处理: %s", idx+1, len(req.FileNames), fname)))

			filePath := filepath.Join(req.DirPath, fname)

			// 获取视频总时长
			duration, err := getVideoDuration(filePath)
			if err != nil {
				mu.Lock()
				failedCount++
				failedFiles = append(failedFiles, fname+" (获取时长失败)")
				mu.Unlock()
				callback([]byte(fmt.Sprintf("❌ [%d/%d] 获取时长失败: %s", idx+1, len(req.FileNames), fname)))
				return
			}

			if req.Duration >= duration {
				mu.Lock()
				failedCount++
				failedFiles = append(failedFiles, fname+" (片头时长大于视频总时长)")
				mu.Unlock()
				callback([]byte(fmt.Sprintf("❌ [%d/%d] 片头时长大于视频总时长: %s", idx+1, len(req.FileNames), fname)))
				return
			}

			ext := strings.ToLower(filepath.Ext(filePath))
			basePath := filePath[:len(filePath)-len(ext)]
			outputPath := fmt.Sprintf("%s_cut%s", basePath, ext)

			// 无损切割：-ss 放在 -i 之前会自动从最近关键帧开始
			execCmd := core.Command("ffmpeg",
				"-ss", fmt.Sprintf("%.3f", req.Duration),
				"-i", filePath,
				"-c", "copy",
				"-avoid_negative_ts", "make_zero",
				"-y",
				outputPath)

			// 记录完整命令行
			cmdStr := "ffmpeg"
			for _, arg := range execCmd.Args {
				cmdStr += " " + arg
			}
			callback([]byte("执行命令: " + cmdStr))

			execCmd.Stdout = os.Stdout
			execCmd.Stderr = os.Stderr

			if err := execCmd.Run(); err != nil {
				mu.Lock()
				failedCount++
				failedFiles = append(failedFiles, fname+" (处理失败: "+err.Error()+")")
				mu.Unlock()
				callback([]byte(fmt.Sprintf("❌ [%d/%d] 处理失败: %s", idx+1, len(req.FileNames), fname)))
				return
			}

			// 删除原文件并重命名，任何一步失败都不能算成功，
			// 否则原文件被删、结果留在 xxx_cut.mp4 却被计为成功
			if err := os.Remove(filePath); err != nil {
				mu.Lock()
				failedCount++
				failedFiles = append(failedFiles, fname+" (删除原文件失败: "+err.Error()+")")
				mu.Unlock()
				callback([]byte(fmt.Sprintf("❌ [%d/%d] 删除原文件失败: %s", idx+1, len(req.FileNames), fname)))
				return
			}
			if err := os.Rename(outputPath, filePath); err != nil {
				mu.Lock()
				failedCount++
				failedFiles = append(failedFiles, fname+" (重命名失败: "+err.Error()+")")
				mu.Unlock()
				callback([]byte(fmt.Sprintf("❌ [%d/%d] 重命名失败: %s", idx+1, len(req.FileNames), fname)))
				return
			}

			mu.Lock()
			successCount++
			mu.Unlock()
			callback([]byte(fmt.Sprintf("✅ [%d/%d] 完成: %s", idx+1, len(req.FileNames), fname)))
		}(i, fileName)
	}

	wg.Wait()

	elapsed := time.Since(startTime)
	callback([]byte(fmt.Sprintf("处理完成！成功: %d，失败: %d，总耗时: %s", successCount, failedCount, elapsed.String())))

	return TrimVideoStartByFilesRes{
		Success:      failedCount == 0,
		Message:      fmt.Sprintf("成功: %d，失败: %d", successCount, failedCount),
		TotalCount:   len(req.FileNames),
		SuccessCount: successCount,
		FailedCount:  failedCount,
		FailedFiles:  failedFiles,
		TotalCost:    elapsed.String(),
	}
}

// VideoExtInfo 视频扩展名统计信息
type VideoExtInfo struct {
	Ext   string `json:"ext"`   // 扩展名（如 .mp4）
	Count int    `json:"count"` // 数量
}

// ScanVideoDirReq 扫描视频目录请求
type ScanVideoDirReq struct {
	DirPath string `json:"dirPath"` // 要扫描的目录路径
}

// ScanVideoDirRes 扫描视频目录结果
type ScanVideoDirRes struct {
	Success    bool           `json:"success"`
	Message    string         `json:"message"`
	TotalCount int            `json:"totalCount"` // 视频总数
	ExtInfos   []VideoExtInfo `json:"extInfos"`   // 各类型统计
}

// ScanVideoDir 扫描视频目录，统计视频个数和各类型个数
func ScanVideoDir(req ScanVideoDirReq, callback func([]byte)) ScanVideoDirRes {
	callback([]byte("开始扫描目录：" + req.DirPath))

	if req.DirPath == "" {
		return ScanVideoDirRes{Success: false, Message: "请选择目录"}
	}

	if !core.FileExist(req.DirPath) {
		return ScanVideoDirRes{Success: false, Message: "目录不存在：" + req.DirPath}
	}

	// 常见视频格式扩展名
	videoExts := map[string]bool{
		".mp4":  true,
		".avi":  true,
		".mkv":  true,
		".mov":  true,
		".wmv":  true,
		".flv":  true,
		".m4v":  true,
		".ts":   true,
		".rmvb": true,
		".webm": true,
	}

	entries, err := os.ReadDir(req.DirPath)
	if err != nil {
		return ScanVideoDirRes{Success: false, Message: "无法读取目录：" + err.Error()}
	}

	extCountMap := make(map[string]int)
	totalCount := 0

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		fileName := entry.Name()
		ext := strings.ToLower(filepath.Ext(fileName))
		if !videoExts[ext] {
			continue
		}
		extCountMap[ext]++
		totalCount++
	}

	var extInfos []VideoExtInfo
	for ext, count := range extCountMap {
		extInfos = append(extInfos, VideoExtInfo{Ext: ext, Count: count})
	}

	// 按数量降序排序
	sort.Slice(extInfos, func(i, j int) bool {
		return extInfos[i].Count > extInfos[j].Count
	})

	callback([]byte(fmt.Sprintf("扫描完成，共找到 %d 个视频文件", totalCount)))

	return ScanVideoDirRes{
		Success:    true,
		Message:    fmt.Sprintf("共 %d 个视频", totalCount),
		TotalCount: totalCount,
		ExtInfos:   extInfos,
	}
}

// MergeVideosReq 合并视频请求
type MergeVideosReq struct {
	DirPath    string `json:"dirPath"`    // 视频目录路径
	IgnoreSort bool   `json:"ignoreSort"` // 是否忽略排序
}

// MergeVideosRes 合并视频结果
type MergeVideosRes struct {
	Success      bool     `json:"success"`
	Message      string   `json:"message"`
	TotalCount   int      `json:"totalCount"`   // 总视频组数
	SuccessCount int      `json:"successCount"` // 成功数
	FailedCount  int      `json:"failedCount"`  // 失败数
	FailedDirs   []string `json:"failedDirs"`   // 失败目录列表
	TotalCost    string   `json:"totalCost"`     // 总耗时
}

// MergeVideos 将目录下的子目录中的视频分别合并
func MergeVideos(req MergeVideosReq, callback func([]byte)) MergeVideosRes {
	startTime := time.Now()
	callback([]byte("开始合并视频..."))

	if req.DirPath == "" {
		return MergeVideosRes{Success: false, Message: "请选择视频目录"}
	}

	if !core.FileExist(req.DirPath) {
		return MergeVideosRes{Success: false, Message: "目录不存在：" + req.DirPath}
	}

	callback([]byte(fmt.Sprintf("工作目录: %s", req.DirPath)))

	// 读取目录下所有子目录
	entries, err := os.ReadDir(req.DirPath)
	if err != nil {
		return MergeVideosRes{Success: false, Message: "无法读取目录：" + err.Error()}
	}

	var subDirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			subDirs = append(subDirs, entry.Name())
		}
	}

	if len(subDirs) == 0 {
		return MergeVideosRes{Success: false, Message: "目录下没有子目录，每个子目录将被视为一组要合并的视频"}
	}

	callback([]byte(fmt.Sprintf("找到 %d 个子目录", len(subDirs))))

	var mu sync.Mutex
	var successCount, failedCount int
	var failedDirs []string

	// 按名称排序，保证处理顺序一致
	sort.Strings(subDirs)

	for i, subDir := range subDirs {
		callback([]byte(fmt.Sprintf("[%d/%d] 处理目录: %s", i+1, len(subDirs), subDir)))

		subDirPath := filepath.Join(req.DirPath, subDir)
		success, reason := mergeSingleDir(subDirPath, req.IgnoreSort, callback)
		if success {
			mu.Lock()
			successCount++
			mu.Unlock()
			callback([]byte(fmt.Sprintf("✅ [%d/%d] 完成: %s", i+1, len(subDirs), subDir)))
		} else {
			mu.Lock()
			failedCount++
			failedDirs = append(failedDirs, subDir+" ("+reason+")")
			mu.Unlock()
			callback([]byte(fmt.Sprintf("❌ [%d/%d] 失败: %s - %s", i+1, len(subDirs), subDir, reason)))
		}
	}

	elapsed := time.Since(startTime)
	totalCost := fmt.Sprintf("%.1f秒", elapsed.Seconds())

	callback([]byte(fmt.Sprintf("合并完成！成功: %d，失败: %d，总耗时: %s", successCount, failedCount, totalCost)))

	return MergeVideosRes{
		Success:      failedCount == 0,
		Message:      fmt.Sprintf("合并完成，成功: %d，失败: %d", successCount, failedCount),
		TotalCount:   len(subDirs),
		SuccessCount: successCount,
		FailedCount:  failedCount,
		FailedDirs:   failedDirs,
		TotalCost:    totalCost,
	}
}

// mergeSingleDir 合并单个目录下的所有视频
func mergeSingleDir(dirPath string, ignoreSort bool, callback func([]byte)) (bool, string) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return false, "无法读取目录"
	}

	// 收集所有 mp4 文件
	var nameList []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext == ".mp4" {
			nameList = append(nameList, entry.Name())
		}
	}

	if len(nameList) == 0 {
		return false, "没有找到 MP4 文件"
	}

	// 按文件名数字排序
	if !ignoreSort {
		sort.Slice(nameList, func(i, j int) bool {
			name1 := strings.Split(nameList[i], ".")[0]
			name2 := strings.Split(nameList[j], ".")[0]
			return core.StrToInt(name1) < core.StrToInt(name2)
		})
	}

	callback([]byte(fmt.Sprintf("  找到 %d 个视频文件", len(nameList))))

	// 生成文件列表内容
	var listBuilder strings.Builder
	for _, v := range nameList {
		listBuilder.WriteString(fmt.Sprintf("file '%s'\n", filepath.Join(dirPath, v)))
	}

	// 写入临时文件列表
	listFilePath := filepath.Join(dirPath, "merge_list.txt")
	if err := os.WriteFile(listFilePath, []byte(listBuilder.String()), 0644); err != nil {
		return false, "无法创建文件列表"
	}

	// 输出文件列表内容到前端
	callback([]byte(fmt.Sprintf("  文件列表内容:")))
	for i, v := range nameList {
		callback([]byte(fmt.Sprintf("    %d. %s", i+1, v)))
	}

	// 输出文件路径
	outputPath := dirPath + ".mp4"

	// 输出完整命令到前端
	ffmpegCmd := fmt.Sprintf("ffmpeg -f concat -safe 0 -i \"%s\" -c copy -y \"%s\"", listFilePath, outputPath)
	callback([]byte(fmt.Sprintf("执行命令: %s", ffmpegCmd)))

	// 执行 ffmpeg 合并
	cmd := core.Command("ffmpeg",
		"-f", "concat",
		"-safe", "0",
		"-i", listFilePath,
		"-c", "copy",
		"-y",
		outputPath)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		os.Remove(listFilePath)
		return false, "ffmpeg 执行失败: " + err.Error()
	}

	// 删除临时文件列表
	os.Remove(listFilePath)

	return true, ""
}

// MergeVideosByFilesReq 按文件列表合并视频请求
type MergeVideosByFilesReq struct {
	DirPath    string   `json:"dirPath"`    // 视频目录路径
	FileNames  []string `json:"fileNames"`  // 要合并的文件名列表
	OutputName string   `json:"outputName"` // 输出文件名（不含扩展名）
}

// MergeVideosByFilesRes 按文件列表合并视频结果
type MergeVideosByFilesRes struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	OutputPath string `json:"outputPath"`
	TotalCost  string `json:"totalCost"`
}

// MergeVideosByFiles 合并指定的视频文件列表
func MergeVideosByFiles(req MergeVideosByFilesReq, callback func([]byte)) MergeVideosByFilesRes {
	startTime := time.Now()
	callback([]byte("开始合并视频..."))

	if req.DirPath == "" {
		return MergeVideosByFilesRes{Success: false, Message: "请选择视频目录"}
	}

	if len(req.FileNames) == 0 {
		return MergeVideosByFilesRes{Success: false, Message: "没有选择要合并的文件"}
	}

	callback([]byte(fmt.Sprintf("要合并 %d 个视频文件", len(req.FileNames))))

	// 生成临时文件列表
	var listBuilder strings.Builder
	for _, fileName := range req.FileNames {
		filePath := filepath.Join(req.DirPath, fileName)
		listBuilder.WriteString(fmt.Sprintf("file '%s'\n", filePath))
	}

	listFilePath := filepath.Join(req.DirPath, "merge_list_"+strconv.FormatInt(time.Now().UnixNano(), 10)+".txt")
	if err := os.WriteFile(listFilePath, []byte(listBuilder.String()), 0644); err != nil {
		return MergeVideosByFilesRes{Success: false, Message: "无法创建文件列表"}
	}

	// 输出文件列表内容到前端
	callback([]byte(fmt.Sprintf("  文件列表内容:")))
	for i, fileName := range req.FileNames {
		callback([]byte(fmt.Sprintf("    %d. %s", i+1, fileName)))
	}

	// 确定输出路径
	outputName := req.OutputName
	if outputName == "" {
		outputName = "merged"
	}
	outputPath := filepath.Join(req.DirPath, outputName+".mp4")

	// 输出完整命令到前端
	ffmpegCmd := fmt.Sprintf("ffmpeg -f concat -safe 0 -i \"%s\" -c copy -y \"%s\"", listFilePath, outputPath)
	callback([]byte(fmt.Sprintf("执行命令: %s", ffmpegCmd)))

	// 执行 ffmpeg 合并
	cmd := core.Command("ffmpeg",
		"-f", "concat",
		"-safe", "0",
		"-i", listFilePath,
		"-c", "copy",
		"-y",
		outputPath)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	callback([]byte("执行 ffmpeg 合并..."))

	if err := cmd.Run(); err != nil {
		os.Remove(listFilePath)
		return MergeVideosByFilesRes{Success: false, Message: "合并失败: " + err.Error()}
	}

	// 删除临时文件列表
	os.Remove(listFilePath)

	elapsed := time.Since(startTime)
	totalCost := fmt.Sprintf("%.1f秒", elapsed.Seconds())

	callback([]byte(fmt.Sprintf("合并完成，输出: %s，耗时: %s", outputPath, totalCost)))

	return MergeVideosByFilesRes{
		Success:    true,
		Message:    "视频合并完成",
		OutputPath: outputPath,
		TotalCost:  totalCost,
	}
}

// GetKeyframesReq 获取关键帧请求
type GetKeyframesReq struct {
	DirPath     string  `json:"dirPath"`     // 视频目录路径
	FileName    string  `json:"fileName"`    // 文件名
	MaxDuration float64 `json:"maxDuration"` // 最大获取时长（秒），默认10秒
}

// GetKeyframesRes 获取关键帧结果
type GetKeyframesRes struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	FileName  string      `json:"fileName"`
	Keyframes []KeyframeInfo `json:"keyframes"` // 关键帧列表
}

// KeyframeInfo 关键帧信息
type KeyframeInfo struct {
	Index     int     `json:"index"`      // 帧序号
	Time      float64 `json:"time"`        // 时间位置（秒）
	TimeStr   string  `json:"timeStr"`     // 时间字符串（分:秒.毫秒）
}

// GetKeyframes 获取视频关键帧位置信息
func GetKeyframes(req GetKeyframesReq, callback func([]byte)) GetKeyframesRes {
	if req.DirPath == "" {
		return GetKeyframesRes{Success: false, Message: "请选择视频目录"}
	}
	if req.FileName == "" {
		return GetKeyframesRes{Success: false, Message: "请选择视频文件"}
	}

	maxDuration := req.MaxDuration
	if maxDuration <= 0 {
		maxDuration = 10.0
	}

	filePath := filepath.Join(req.DirPath, req.FileName)

	cmd := core.Command("ffprobe",
		"-select_streams", "v:0",
		"-show_frames",
		"-show_entries", "frame=pts_time,pict_type",
		"-of", "csv",
		"-read_intervals", fmt.Sprintf("0%%%.0f", maxDuration),
		filePath)


	// 输出完整命令行
	callback([]byte("执行命令: " + strings.Join(cmd.Args, " ")))

	output, err := cmd.Output()
	if err != nil {
		return GetKeyframesRes{Success: false, Message: "获取关键帧失败: " + err.Error()}
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	keyframes := []KeyframeInfo{}
	frameIndex := 0

	for _, line := range lines {
		if line == "" {
			continue
		}
		// CSV 格式: frame,pts_time,pict_type[,extra]
		parts := strings.Split(line, ",")
		if len(parts) < 3 {
			continue
		}
		// pict_type 是第3列 (index 2)
		if parts[2] != "I" {
			continue
		}

		ptsTime, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			continue
		}

		frameIndex++
		minutes := int(ptsTime) / 60
		seconds := int(ptsTime) % 60
		millis := int((ptsTime - float64(int(ptsTime))) * 1000)
		timeStr := fmt.Sprintf("%02d:%02d.%03d", minutes, seconds, millis)

		keyframes = append(keyframes, KeyframeInfo{
			Index:   frameIndex,
			Time:    ptsTime,
			TimeStr: timeStr,
		})
	}

	return GetKeyframesRes{
		Success:   true,
		Message:   fmt.Sprintf("共找到 %d 个关键帧", len(keyframes)),
		FileName:  req.FileName,
		Keyframes: keyframes,
	}
}

// ClassifyVideosByResolutionReq 按分辨率分类视频请求
type ClassifyVideosByResolutionReq struct {
	DirPath string `json:"dirPath"` // 视频目录路径
}

// ClassifyVideosByResolutionRes 按分辨率分类视频结果
type ClassifyVideosByResolutionRes struct {
	Success        bool     `json:"success"`
	Message       string   `json:"message"`
	TotalCount     int      `json:"totalCount"`     // 总视频数
	SuccessCount   int      `json:"successCount"`   // 成功分类数
	FailedCount    int      `json:"failedCount"`    // 失败数
	ResolutionMap  map[string]int `json:"resolutionMap"` // 分辨率 -> 数量
	FailedFiles    []string `json:"failedFiles"`     // 失败文件列表
	TotalCost      string   `json:"totalCost"`       // 总耗时
}

// ClassifyVideosByResolution 将目录下的视频按分辨率分类到不同文件夹
func ClassifyVideosByResolution(req ClassifyVideosByResolutionReq, callback func([]byte)) ClassifyVideosByResolutionRes {
	startTime := time.Now()
	callback([]byte("开始按分辨率分类视频..."))

	if req.DirPath == "" {
		return ClassifyVideosByResolutionRes{Success: false, Message: "请选择视频目录"}
	}

	if !core.FileExist(req.DirPath) {
		return ClassifyVideosByResolutionRes{Success: false, Message: "目录不存在：" + req.DirPath}
	}

	// 常见视频格式扩展名
	videoExts := map[string]bool{
		".mp4":  true, ".avi": true, ".mkv": true, ".mov": true,
		".wmv":  true, ".flv": true, ".m4v": true, ".ts": true,
		".rmvb": true, ".webm": true,
	}

	// 扫描目录下所有视频文件（不递归）
	entries, err := os.ReadDir(req.DirPath)
	if err != nil {
		return ClassifyVideosByResolutionRes{Success: false, Message: "无法读取目录：" + err.Error()}
	}

	var videoFiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		fileName := entry.Name()
		ext := strings.ToLower(filepath.Ext(fileName))
		if !videoExts[ext] {
			continue
		}
		videoFiles = append(videoFiles, fileName)
	}

	if len(videoFiles) == 0 {
		return ClassifyVideosByResolutionRes{Success: false, Message: "目录中没有找到视频文件"}
	}

	callback([]byte(fmt.Sprintf("共找到 %d 个视频文件", len(videoFiles))))

	// 用于统计各分辨率数量
	resolutionMap := make(map[string]int)
	var mu sync.Mutex
	var successCount, failedCount int
	var failedFiles []string

	for i, fileName := range videoFiles {
		callback([]byte(fmt.Sprintf("[%d/%d] 处理中: %s", i+1, len(videoFiles), fileName)))

		filePath := filepath.Join(req.DirPath, fileName)

		// 获取视频分辨率
		detail, err := getVideoDetail(filePath)
		if err != nil {
			mu.Lock()
			failedCount++
			failedFiles = append(failedFiles, fileName+" (获取分辨率失败)")
			mu.Unlock()
			callback([]byte(fmt.Sprintf("❌ [%d/%d] 获取分辨率失败: %s", i+1, len(videoFiles), fileName)))
			continue
		}

		resolution := fmt.Sprintf("%dx%d", detail.Width, detail.Height)
		if resolution == "0x0" {
			resolution = "未知"
		}

		// 创建分辨率文件夹
		resDir := filepath.Join(req.DirPath, resolution)
		if err := os.MkdirAll(resDir, 0755); err != nil {
			mu.Lock()
			failedCount++
			failedFiles = append(failedFiles, fileName+" (创建目录失败)")
			mu.Unlock()
			callback([]byte(fmt.Sprintf("❌ [%d/%d] 创建目录失败: %s", i+1, len(videoFiles), fileName)))
			continue
		}

		// 目标文件路径
		targetPath := filepath.Join(resDir, fileName)

		// 如果目标文件已存在，跳过
		if _, err := os.Stat(targetPath); err == nil {
			callback([]byte(fmt.Sprintf("⏭️ [%d/%d] 跳过（已存在）: %s -> %s", i+1, len(videoFiles), fileName, resolution)))
			mu.Lock()
			resolutionMap[resolution]++
			successCount++
			mu.Unlock()
			continue
		}

		// 移动文件到分辨率文件夹
		if err := os.Rename(filePath, targetPath); err != nil {
			mu.Lock()
			failedCount++
			failedFiles = append(failedFiles, fileName+" (移动失败: "+err.Error()+")")
			mu.Unlock()
			callback([]byte(fmt.Sprintf("❌ [%d/%d] 移动失败: %s", i+1, len(videoFiles), fileName)))
			continue
		}

		mu.Lock()
		resolutionMap[resolution]++
		successCount++
		mu.Unlock()
		callback([]byte(fmt.Sprintf("✅ [%d/%d] 完成: %s -> %s", i+1, len(videoFiles), fileName, resolution)))
	}

	elapsed := time.Since(startTime)
	totalCost := fmt.Sprintf("%.1f秒", elapsed.Seconds())

	callback([]byte(fmt.Sprintf("分类完成！成功: %d，失败: %d，总耗时: %s", successCount, failedCount, totalCost)))
	callback([]byte(""))
	callback([]byte("=== 分辨率统计 ==="))
	for res, count := range resolutionMap {
		callback([]byte(fmt.Sprintf("  %s: %d个", res, count)))
	}

	return ClassifyVideosByResolutionRes{
		Success:       failedCount == 0,
		Message:       fmt.Sprintf("分类完成，成功: %d，失败: %d", successCount, failedCount),
		TotalCount:    len(videoFiles),
		SuccessCount:  successCount,
		FailedCount:   failedCount,
		ResolutionMap: resolutionMap,
		FailedFiles:   failedFiles,
		TotalCost:     totalCost,
	}
}