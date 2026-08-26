package golang

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/wyzzgzhdcxy/wcj-go-common/core"
)

// YtDlpVideo video information from yt-dlp --dump-json
type YtDlpVideo struct {
	ID               string                     `json:"id"`
	Title            string                     `json:"title"`
	Thumbnail        string                     `json:"thumbnail"`
	Description      string                     `json:"description"`
	Duration         float64                    `json:"duration"`
	DurationString   string                     `json:"duration_string"`
	Formats          []YtDlpFormat              `json:"formats"`
	Thumbnails       []YtDlpThumbnail           `json:"thumbnails"`
	Subtitles        map[string][]YtDlpSubtitle `json:"subtitles"`
	RequestedFormats []YtDlpFormat              `json:"requested_formats"`
	Filename         string                     `json:"_filename"`
	UploadDate       string                     `json:"upload_date"`
	IsLive           bool                       `json:"is_live"`
	Chapters         []YtDlpChapter             `json:"chapters"`
}

// YtDlpFormat format information
type YtDlpFormat struct {
	FormatID     string  `json:"format_id"`
	FormatNote   string  `json:"format_note"`
	Ext          string  `json:"ext"`
	Width        int     `json:"width"`
	Height       int     `json:"height"`
	Resolution   string  `json:"resolution"`
	FPS          float64 `json:"fps"`
	VCodec       string  `json:"vcodec"`
	ACodec       string  `json:"acodec"`
	Filesize     int64   `json:"filesize"`
	FormatType   string  `json:"format"` // "video", "audio", "video+audio"
	DynamicRange string  `json:"dynamic_range"`
	ASR          int     `json:"asr"` // audio sample rate
	TBR          float64 `json:"tbr"`
}

// YtDlpThumbnail thumbnail information
type YtDlpThumbnail struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// YtDlpSubtitle subtitle information
type YtDlpSubtitle struct {
	Ext string `json:"ext"`
	URL string `json:"url"`
}

// YtDlpChapter chapter information
type YtDlpChapter struct {
	Title     string  `json:"title"`
	StartTime float64 `json:"start_time"`
	EndTime   float64 `json:"end_time"`
}

// YtDlpOptions download options
type YtDlpOptions struct {
	FormatID       string
	OutputPath     string
	Proxy          string
	ProxyEnabled   bool
	CookieType     string // "chrome", "edge", "firefox", etc.
	UseCookie      bool
	EmbedThumbnail bool
	EmbedChapters  bool
	EmbedSubtitles bool
	LimitRate      string
	TimeRange      string
	ExtractAudio   bool
	AudioFormat    string
}

// ProgressInfo progress information
type ProgressInfo struct {
	Percent    float64 `json:"percent"`
	Speed      string  `json:"speed"`
	ETA        string  `json:"eta"`
	Downloaded int64   `json:"downloaded"`
	Total      int64   `json:"total"`
	Status     string  `json:"status"` // "downloading", "finished", "error"
}

var (
	ytdlpMutex     sync.Mutex
	currentProcess *exec.Cmd
	isDownloading  bool
)

// DownloadTask 下载任务结构
type DownloadTask struct {
	ID           string       `json:"id"`           // 任务ID
	URL          string       `json:"url"`          // 视频URL
	FormatID     string       `json:"format_id"`     // 格式ID
	OutputPath   string       `json:"output_path"`   // 输出路径
	Status       string       `json:"status"`       // pending, downloading, completed, failed, cancelled
	Progress     float64      `json:"progress"`     // 下载进度 0-100
	Speed        string       `json:"speed"`        // 下载速度
	ETA          string       `json:"eta"`          // 预计剩余时间
	ErrorMsg     string       `json:"error_msg"`    // 错误信息
	Platform     string       `json:"platform"`     // youtube, bilibili
	CreateTime   time.Time    `json:"create_time"`  // 创建时间
}

// DownloadTaskManager 下载任务管理器
type DownloadTaskManager struct {
	tasks   map[string]*DownloadTask
	mutex   sync.RWMutex
}

var taskManager *DownloadTaskManager

func init() {
	taskManager = &DownloadTaskManager{
		tasks: make(map[string]*DownloadTask),
	}
}

// GetTaskManager 获取任务管理器实例
func GetTaskManager() *DownloadTaskManager {
	return taskManager
}

// AddTask 添加下载任务
func (m *DownloadTaskManager) AddTask(task *DownloadTask) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	task.CreateTime = time.Now()
	if task.Status == "" {
		task.Status = "pending"
	}
	m.tasks[task.ID] = task
}

// GetTask 获取任务
func (m *DownloadTaskManager) GetTask(id string) *DownloadTask {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.tasks[id]
}

// GetAllTasks 获取所有任务
func (m *DownloadTaskManager) GetAllTasks() []*DownloadTask {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	tasks := make([]*DownloadTask, 0, len(m.tasks))
	for _, task := range m.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}

// UpdateTask 更新任务状态
func (m *DownloadTaskManager) UpdateTask(id string, updateFunc func(*DownloadTask)) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if task, ok := m.tasks[id]; ok {
		updateFunc(task)
	}
}

// RemoveTask 移除任务
func (m *DownloadTaskManager) RemoveTask(id string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.tasks, id)
}

// GetPendingTasks 获取所有待执行的任务
func (m *DownloadTaskManager) GetPendingTasks() []*DownloadTask {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	tasks := make([]*DownloadTask, 0)
	for _, task := range m.tasks {
		if task.Status == "pending" || task.Status == "downloading" {
			tasks = append(tasks, task)
		}
	}
	return tasks
}

// AddDownloadTask 添加下载任务（前端调用的接口）
func AddDownloadTask(url string, formatID string, outputPath string, platform string) string {
	task := &DownloadTask{
		ID:         uuid.New().String(),
		URL:        url,
		FormatID:   formatID,
		OutputPath: outputPath,
		Platform:   platform,
		Status:     "pending",
	}
	GetTaskManager().AddTask(task)
	// 异步开始下载
	go startDownloadTask(task.ID)
	return task.ID
}

// startDownloadTask 开始下载任务
func startDownloadTask(taskID string) {
	task := GetTaskManager().GetTask(taskID)
	if task == nil {
		return
	}

	GetTaskManager().UpdateTask(taskID, func(t *DownloadTask) {
		t.Status = "downloading"
	})

	// 实际执行下载
	executeDownload(task.URL, task.FormatID, task.OutputPath, task.Platform, func(progress ProgressInfo, logMsg string) {
		if logMsg != "" {
			log.Printf("[Task %s] %s", taskID, logMsg)
		}
		GetTaskManager().UpdateTask(taskID, func(t *DownloadTask) {
			if progress.Percent > 0 {
				t.Progress = progress.Percent
			}
			if progress.Speed != "" {
				t.Speed = progress.Speed
			}
			if progress.ETA != "" {
				t.ETA = progress.ETA
			}
			if progress.Status != "" {
				t.Status = progress.Status
			}
		})
	})
}

// executeDownload 执行下载
func executeDownload(url string, formatID string, outputPath string, platform string, progressCallback func(ProgressInfo, string)) error {
	ytdlpMutex.Lock()
	if isDownloading {
		ytdlpMutex.Unlock()
		return fmt.Errorf("download already in progress")
	}
	isDownloading = true
	ytdlpMutex.Unlock()

	defer func() {
		ytdlpMutex.Lock()
		isDownloading = false
		ytdlpMutex.Unlock()
	}()

	ytdlpPath := FindYtDlp()
	log.Printf("[executeDownload] yt-dlp path: %s", ytdlpPath)
	log.Printf("[executeDownload] url: %s", url)
	log.Printf("[executeDownload] output: %s", outputPath)

	args := []string{
		"--progress-template",
		"[yt-dlp]%(progress._percent_str)s,%(progress._eta_str)s,%(progress.downloaded_bytes)s,%(progress.total_bytes)s,%(progress.speed)s,%(progress.eta)s",
	}

	if platform == "youtube" {
		args = append(args, "--cookies", getYoutubeCookiePath())
		args = append(args, "--proxy", "socks5://127.0.0.1:10808")
		args = append(args, "--js-runtimes", "deno:E:\\application\\我的工具箱\\deno.exe")
		args = append(args, "--merge-output-format", "mp4")
	} else {
		args = append(args, "--cookies", "D:\\cookies.txt")
		args = append(args, "-N", "4")
		args = append(args, "--throttled-rate", "100K")
		args = append(args, "--user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	}

	args = append(args, "-f", formatID)
	args = append(args, "-o", outputPath)
	args = append(args, url)

	fullCmd := ytdlpPath + " " + strings.Join(args, " ")
	log.Printf("[executeDownload] cmd: %s", fullCmd)
	progressCallback(ProgressInfo{}, fmt.Sprintf("执行命令: %s\n", fullCmd))

	cmd := exec.Command(ytdlpPath, args...)
	currentProcess = cmd
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Printf("[executeDownload] stderr pipe error: %v", err)
		return fmt.Errorf("failed to create stderr pipe: %v", err)
	}

	err = cmd.Start()
	if err != nil {
		log.Printf("[executeDownload] start error: %v", err)
		return fmt.Errorf("failed to start yt-dlp: %v", err)
	}

	scanner := bufio.NewScanner(stderr)
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			log.Printf("[yt-dlp] %s", line)
			progressCallback(ProgressInfo{}, line)
			progress := parseProgressLine(line)
			if progress != nil {
				progressCallback(*progress, "")
			}
		}
	}()

	err = cmd.Wait()
	currentProcess = nil

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == -1 {
				return fmt.Errorf("download cancelled")
			}
		}
		return fmt.Errorf("download failed: %v", err)
	}

	progressCallback(ProgressInfo{
		Status:  "finished",
		Percent: 100,
	}, "")

	return nil
}

// CancelTask 取消任务
func (m *DownloadTaskManager) CancelTask(id string) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if task, ok := m.tasks[id]; ok {
		if task.Status == "downloading" {
			if currentProcess != nil {
				currentProcess.Process.Kill()
			}
		}
		task.Status = "cancelled"
		return true
	}
	return false
}

// FindYtDlp finds yt-dlp executable
func FindYtDlp() string {
	// Check common locations
	locations := []string{
		"yt-dlp",
		"yt-dlp.exe",
		filepath.Join(core.ExecPath(), "yt-dlp.exe"),
		filepath.Join(core.ExecPath(), "yt-dlp"),
	}

	for _, loc := range locations {
		if strings.HasPrefix(loc, "/") || strings.HasPrefix(loc, "C:") || strings.HasPrefix(loc, "E:") {
			if core.FileExist(loc) {
				return loc
			}
		} else {
			// Try to find in PATH
			cmd := exec.Command("where", "yt-dlp")
			cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
			output, err := cmd.Output()
			if err == nil && len(output) > 0 {
				lines := strings.Split(strings.TrimSpace(string(output)), "\n")
				if len(lines) > 0 {
					return strings.TrimSpace(lines[0])
				}
			}
		}
	}
	return "yt-dlp" // Fallback to just "yt-dlp" and hope it's in PATH
}

// GetVideoInfo retrieves video information using yt-dlp --dump-json
func GetVideoInfo(url string) (string, error) {
	ytdlpPath := FindYtDlp()

	args := []string{
		"--dump-json",
		"--no-playlist",
		"--flat-playlist",
		url,
	}

	cmd := exec.Command(ytdlpPath, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Stderr = os.Stderr

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get video info: %v", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// ParseVideoInfo parses JSON output from yt-dlp
func ParseVideoInfo(jsonStr string) (*YtDlpVideo, error) {
	var video YtDlpVideo
	err := json.Unmarshal([]byte(jsonStr), &video)
	if err != nil {
		return nil, fmt.Errorf("failed to parse video info: %v", err)
	}
	return &video, nil
}

// DownloadVideo downloads a video using yt-dlp with progress reporting
func DownloadVideo(url string, options YtDlpOptions, progressCallback func(ProgressInfo)) error {
	ytdlpMutex.Lock()
	if isDownloading {
		ytdlpMutex.Unlock()
		return fmt.Errorf("download already in progress")
	}
	isDownloading = true
	ytdlpMutex.Unlock()

	defer func() {
		ytdlpMutex.Lock()
		isDownloading = false
		ytdlpMutex.Unlock()
	}()

	ytdlpPath := FindYtDlp()

	args := []string{
		"--progress-template",
		"[yt-dlp]%(progress._percent_str)s,%(progress._eta_str)s,%(progress.downloaded_bytes)s,%(progress.total_bytes)s,%(progress.speed)s,%(progress.eta)s",
		"--no-playlist",
		"-f", options.FormatID,
		"-o", options.OutputPath,
	}

	// Add optional arguments
	if options.ProxyEnabled && options.Proxy != "" {
		args = append(args, "--proxy", options.Proxy)
	}

	if options.UseCookie && options.CookieType != "" {
		args = append(args, "--cookies-from-browser", options.CookieType)
	}

	if options.EmbedThumbnail {
		args = append(args, "--embed-thumbnail")
	}

	if options.EmbedChapters {
		args = append(args, "--embed-chapters")
	}

	if options.EmbedSubtitles {
		args = append(args, "--embed-subs")
	}

	if options.LimitRate != "" {
		args = append(args, "--limit-rate", options.LimitRate)
	}

	if options.TimeRange != "" {
		args = append(args, "--download-sections", options.TimeRange)
	}

	if options.ExtractAudio {
		args = append(args, "-x", "--audio-format", options.AudioFormat)
	}

	args = append(args, url)

	cmd := exec.Command(ytdlpPath, args...)
	currentProcess = cmd
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	// Create a pipe to read stderr (yt-dlp outputs progress to stderr)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %v", err)
	}

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start yt-dlp: %v", err)
	}

	// Read stderr line by line and parse progress
	scanner := bufio.NewScanner(stderr)
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			progress := parseProgressLine(line)
			if progress != nil {
				progressCallback(*progress)
			}
		}
	}()

	err = cmd.Wait()
	currentProcess = nil

	if err != nil {
		// Check if it was cancelled
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == -1 {
				return fmt.Errorf("download cancelled")
			}
		}
		return fmt.Errorf("download failed: %v", err)
	}

	progressCallback(ProgressInfo{
		Status:  "finished",
		Percent: 100,
	})

	return nil
}

// CancelDownload cancels the current download
func CancelDownload() {
	ytdlpMutex.Lock()
	defer ytdlpMutex.Unlock()

	if currentProcess != nil && isDownloading {
		currentProcess.Process.Kill()
		isDownloading = false
	}
}

// parseProgressLine parses a progress line from yt-dlp output
func parseProgressLine(line string) *ProgressInfo {
	// Format: [yt-dlp]10.5%,00:30,100MB,1GB,1.5MB/s,ETA
	re := regexp.MustCompile(`\[yt-dlp\]([^,]*),([^,]*),([^,]*),([^,]*),([^,]*),(\S+)`)
	matches := re.FindStringSubmatch(line)
	if matches == nil {
		return nil
	}

	percentStr := matches[1]
	percent, _ := strconv.ParseFloat(strings.TrimSuffix(percentStr, "%"), 64)

	downloadedStr := matches[3]
	totalStr := matches[4]
	speed := matches[5]
	eta := matches[6]

	downloaded := parseSize(downloadedStr)
	total := parseSize(totalStr)

	status := "downloading"
	if percent >= 100 || strings.Contains(strings.ToLower(line), "finished") {
		status = "finished"
	}

	return &ProgressInfo{
		Percent:    percent,
		Speed:      speed,
		ETA:        eta,
		Downloaded: downloaded,
		Total:      total,
		Status:     status,
	}
}

// parseSize parses size strings like "100MB", "1.5GB" to bytes
func parseSize(s string) int64 {
	s = strings.TrimSpace(s)
	if s == "" || s == "N/A" {
		return 0
	}

	re := regexp.MustCompile(`([\d.]+)\s*([KMGT]?B?)`)
	matches := re.FindStringSubmatch(s)
	if matches == nil {
		return 0
	}

	value, _ := strconv.ParseFloat(matches[1], 64)
	unit := strings.ToUpper(matches[2])

	var multiplier int64 = 1
	switch unit {
	case "KB", "K":
		multiplier = 1024
	case "MB", "M":
		multiplier = 1024 * 1024
	case "GB", "G":
		multiplier = 1024 * 1024 * 1024
	case "TB", "T":
		multiplier = 1024 * 1024 * 1024 * 1024
	}

	return int64(value * float64(multiplier))
}

// ListFormats returns formatted list of available formats for a video
func ListFormats(url string) (string, error) {
	ytdlpPath := FindYtDlp()

	args := []string{
		"--list-formats",
		"--no-playlist",
		url,
	}

	cmd := exec.Command(ytdlpPath, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to list formats: %v", err)
	}

	return string(output), nil
}

// CheckDependencies checks if yt-dlp and ffmpeg are available
func CheckDependencies() map[string]bool {
	result := make(map[string]bool)

	// Check yt-dlp
	ytdlpPath := FindYtDlp()
	cmd := exec.Command(ytdlpPath, "--version")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	err := cmd.Run()
	result["yt-dlp"] = err == nil

	// Check ffmpeg
	ffmpegCmd := exec.Command("ffmpeg", "-version")
	ffmpegCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	err = ffmpegCmd.Run()
	result["ffmpeg"] = err == nil

	return result
}

// GetDownloadPath returns the default download path
func GetDownloadPath() string {
	// 优先使用D盘
	if _, err := os.Stat("D:\\"); err == nil {
		return "D:\\"
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return core.GetTempDir()
	}
	return filepath.Join(homeDir, "Downloads")
}

// UpdateYtDlp updates yt-dlp to the latest version
func UpdateYtDlp() (string, error) {
	ytdlpPath := FindYtDlp()

	cmd := exec.Command(ytdlpPath, "-U")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("update failed: %v", err)
	}

	return string(output), nil
}

// isYoutube checks if the URL is from YouTube
func isYoutube(url string) bool {
	return strings.Contains(url, "youtube.com") || strings.Contains(url, "youtu.be")
}

// getYoutubeCookiePath returns the YouTube cookie file path for today
func getYoutubeCookiePath() string {
	today := time.Now().Format("2006-01-02")
	return fmt.Sprintf("C:\\Users\\wangchaojun\\Downloads\\cookies_www.youtube.com_%s.txt", today)
}

// BilibiliListFormats 获取B站或YouTube视频格式列表
func BilibiliListFormats(url string) (string, error) {
	ytdlpPath := FindYtDlp()

	args := []string{}

	if isYoutube(url) {
		// YouTube: 使用YouTube的cookies和代理
		args = []string{
			"--cookies", getYoutubeCookiePath(),
			"--proxy", "socks5://127.0.0.1:10808",
			"--js-runtimes", "deno:E:\\application\\我的工具箱\\deno.exe",
			"-F",
			url,
		}
	} else {
		// B站
		args = []string{
			"--cookies", "D:\\cookies.txt",
			"-F",
			url,
		}
	}

	cmd := exec.Command(ytdlpPath, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to list formats: %v, output: %s", err, string(output))
	}

	return string(output), nil
}

// BilibiliDownload 下载B站或YouTube视频
func BilibiliDownload(url string, formatID string, outputPath string, progressCallback func(ProgressInfo, string)) error {
	ytdlpMutex.Lock()
	if isDownloading {
		ytdlpMutex.Unlock()
		return fmt.Errorf("download already in progress")
	}
	isDownloading = true
	ytdlpMutex.Unlock()

	defer func() {
		ytdlpMutex.Lock()
		isDownloading = false
		ytdlpMutex.Unlock()
	}()

	ytdlpPath := FindYtDlp()
	log.Printf("[BilibiliDownload] yt-dlp path: %s", ytdlpPath)
	log.Printf("[BilibiliDownload] url: %s", url)
	log.Printf("[BilibiliDownload] output: %s", outputPath)

	args := []string{
		"--progress-template",
		"[yt-dlp]%(progress._percent_str)s,%(progress._eta_str)s,%(progress.downloaded_bytes)s,%(progress.total_bytes)s,%(progress.speed)s,%(progress.eta)s",
	}

	if isYoutube(url) {
		// YouTube: 使用YouTube的cookies和代理
		args = append(args, "--cookies", getYoutubeCookiePath())
		args = append(args, "--proxy", "socks5://127.0.0.1:10808")
		args = append(args, "--js-runtimes", "deno:E:\\application\\我的工具箱\\deno.exe")
		args = append(args, "--merge-output-format", "mp4")
	} else {
		// B站
		args = append(args, "--cookies", "D:\\cookies.txt")
		args = append(args, "-N", "4")
		args = append(args, "--throttled-rate", "100K")
		args = append(args, "--user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	}

	args = append(args, "-f", formatID)
	args = append(args, "-o", outputPath)
	args = append(args, url)

	fullCmd := ytdlpPath + " " + strings.Join(args, " ")
	log.Printf("[BilibiliDownload] cmd: %s", fullCmd)
	progressCallback(ProgressInfo{}, fmt.Sprintf("执行命令: %s\n", fullCmd))

	cmd := exec.Command(ytdlpPath, args...)
	currentProcess = cmd
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Printf("[BilibiliDownload] stderr pipe error: %v", err)
		return fmt.Errorf("failed to create stderr pipe: %v", err)
	}

	err = cmd.Start()
	if err != nil {
		log.Printf("[BilibiliDownload] start error: %v", err)
		return fmt.Errorf("failed to start yt-dlp: %v", err)
	}

	scanner := bufio.NewScanner(stderr)
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			log.Printf("[yt-dlp] %s", line)
			progressCallback(ProgressInfo{}, line)
			progress := parseProgressLine(line)
			if progress != nil {
				progressCallback(*progress, "")
			}
		}
	}()

	err = cmd.Wait()
	currentProcess = nil

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == -1 {
				return fmt.Errorf("download cancelled")
			}
		}
		return fmt.Errorf("download failed: %v", err)
	}

	progressCallback(ProgressInfo{
		Status:  "finished",
		Percent: 100,
	}, "")

	return nil
}
