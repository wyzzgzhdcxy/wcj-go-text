package app

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/wyzzgzhdcxy/wcj-go-common/core"
	"github.com/wyzzgzhdcxy/wcj-go-common/utils"

	"wcj-go-text/golang"
	"wcj-go-text/golang/cmdWrapper"
	"wcj-go-text/golang/sqllite"
)

// ---------------- 注册电影相关事件 ----------------

func (a *App) registerMovieEvents() {
}

// ---------------- 选择目录/文件/打开目录 ----------------

func (a *App) SelectDirectory() string {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		DefaultDirectory: a.downloadsDir(),
	})
	if err != nil {
		return ""
	}
	return dir
}

func (a *App) OpenExplorer(path string) {
	if path == "." || path == "" {
		path, _ = os.Getwd()
	}
	path = filepath.FromSlash(path)
	cmdWrapper.CommandVisible("explorer", path).Start()
}

func (a *App) GetTempDir() string {
	return core.GetTempDir()
}

// ListConfigLegacy 已废弃（占位符，保留仅为兼容旧绑定）
func (a *App) ListConfigLegacy() string {
	return "{}"
}

// ---------------- M3U8 提取 ----------------

type ExtractM3u8LinksReq struct {
	URL string `json:"url"`
}

type ExtractM3u8LinksRes struct {
	Success bool       `json:"success"`
	Links   []M3u8Link `json:"links"`
	Message string     `json:"message"`
}

type M3u8Link struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

func (a *App) ExtractM3u8Links(req ExtractM3u8LinksReq) ExtractM3u8LinksRes {
	if req.URL == "" {
		return ExtractM3u8LinksRes{
			Success: false,
			Message: "URL不能为空",
		}
	}

	content := utils.HttpGetContent(req.URL)
	if content == "" {
		return ExtractM3u8LinksRes{
			Success: false,
			Message: "无法获取网页内容",
		}
	}

	var links []M3u8Link

	parts := strings.Split(content, "m3u8")
	for i := 0; i < len(parts)-1; i++ {
		start := strings.LastIndex(parts[i], "https:")
		if start == -1 {
			start = strings.LastIndex(parts[i], "http:")
		}
		if start != -1 {
			u := strings.ReplaceAll(parts[i][start:]+"m3u8", "\\/", "/")
			if qIdx := strings.Index(u, "?"); qIdx != -1 {
				u = u[:qIdx]
			}
			if !strings.Contains(u, ".m3u8") {
				u += ".m3u8"
			}
			links = append(links, M3u8Link{
				Title: fmt.Sprintf("视频_%d", i+1),
				URL:   u,
			})
		}
	}

	seen := make(map[string]bool)
	uniqueLinks := []M3u8Link{}
	for _, link := range links {
		if !seen[link.URL] {
			seen[link.URL] = true
			uniqueLinks = append(uniqueLinks, link)
		}
	}

	if len(uniqueLinks) == 0 {
		return ExtractM3u8LinksRes{
			Success: false,
			Message: "未找到m3u8链接",
		}
	}

	return ExtractM3u8LinksRes{
		Success: true,
		Links:   uniqueLinks,
		Message: fmt.Sprintf("成功提取%d个链接", len(uniqueLinks)),
	}
}

// ---------------- M3U8 下载 ----------------

type DownloadM3u8Req struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	OutputDir   string `json:"outputDir"`
	ThreadCount int    `json:"threadCount"`
	TaskId      string `json:"taskId"`
}

type DownloadM3u8Res struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Output  string `json:"output"`
}

func (a *App) DownloadM3u8(req DownloadM3u8Req) DownloadM3u8Res {
	fmt.Printf("DownloadM3u8 req: url=%s title=%s outputDir=%s\n", req.URL, req.Title, req.OutputDir)
	if req.URL == "" {
		return DownloadM3u8Res{
			Success: false,
			Message: "URL不能为空",
		}
	}

	if req.OutputDir == "" {
		req.OutputDir = a.downloadsDir()
	}

	if req.ThreadCount <= 0 {
		req.ThreadCount = 16
	}

	title := req.Title
	if title == "" {
		title = fmt.Sprintf("video_%s", uuid.New().String()[:8])
	}

	runtime.EventsEmit(a.ctx, "m3u8_progress", map[string]any{
		"status":  "starting",
		"message": fmt.Sprintf("开始下载: %s", title),
		"percent": 0,
		"taskId":  req.TaskId,
	})

	go golang.DownloadM3u8WithProgress(req.URL, title, req.OutputDir, req.ThreadCount, func(status string, percent float64, speed string) {
		runtime.EventsEmit(a.ctx, "m3u8_progress", map[string]any{
			"status":  status,
			"percent": percent,
			"speed":   speed,
			"message": speed,
			"taskId":  req.TaskId,
		})
	})

	return DownloadM3u8Res{
		Success: true,
		Message: fmt.Sprintf("下载已开始: %s", title),
		Output:  filepath.Join(req.OutputDir, title+".mp4"),
	}
}

// ---------------- yt-dlp ----------------

type YtDlpOptions struct {
	FormatID       string `json:"format_id"`
	OutputPath     string `json:"output_path"`
	Proxy          string `json:"proxy"`
	ProxyEnabled   bool   `json:"proxy_enabled"`
	CookieType     string `json:"cookie_type"`
	UseCookie      bool   `json:"use_cookie"`
	EmbedThumbnail bool   `json:"embed_thumbnail"`
	EmbedChapters  bool   `json:"embed_chapters"`
	EmbedSubtitles bool   `json:"embed_subtitles"`
	LimitRate      string `json:"limit_rate"`
	TimeRange      string `json:"time_range"`
	ExtractAudio   bool   `json:"extract_audio"`
	AudioFormat    string `json:"audio_format"`
}

func (a *App) GetYtDlpVideoInfo(videoURL string) string {
	info, err := golang.GetVideoInfo(videoURL)
	if err != nil {
		return ""
	}
	return info
}

func (a *App) DownloadWithYtDlp(videoURL string, opts YtDlpOptions) {
	ytdlpOpts := golang.YtDlpOptions{
		FormatID:       opts.FormatID,
		OutputPath:     opts.OutputPath,
		Proxy:          opts.Proxy,
		ProxyEnabled:   opts.ProxyEnabled,
		CookieType:     opts.CookieType,
		UseCookie:      opts.UseCookie,
		EmbedThumbnail: opts.EmbedThumbnail,
		EmbedChapters:  opts.EmbedChapters,
		EmbedSubtitles: opts.EmbedSubtitles,
		LimitRate:      opts.LimitRate,
		TimeRange:      opts.TimeRange,
		ExtractAudio:   opts.ExtractAudio,
		AudioFormat:    opts.AudioFormat,
	}

	golang.DownloadVideo(videoURL, ytdlpOpts, func(progress golang.ProgressInfo) {
		runtime.EventsEmit(a.ctx, "ytdlp_progress", map[string]any{
			"percent": progress.Percent,
			"speed":   progress.Speed,
			"eta":     progress.ETA,
			"status":  progress.Status,
		})
	})
}

func (a *App) CancelYtDlpDownload() {
	golang.CancelDownload()
}

func (a *App) GetYtDlpDownloadPath() string {
	return golang.GetDownloadPath()
}

func (a *App) CheckYtDlpDeps() map[string]bool {
	return golang.CheckDependencies()
}

// ---------------- B站 / YouTube 下载 ----------------

func (a *App) BilibiliListFormats(videoURL string) string {
	result, err := golang.BilibiliListFormats(videoURL)
	if err != nil {
		return ""
	}
	return result
}

func (a *App) BilibiliDownload(videoURL string, formatID string, outputPath string) {
	platform := "bilibili"
	if strings.Contains(videoURL, "youtube.com") || strings.Contains(videoURL, "youtu.be") {
		platform = "youtube"
	}
	golang.AddDownloadTask(videoURL, formatID, outputPath, platform)
}

func (a *App) GetDownloadTasks() string {
	tasks := golang.GetTaskManager().GetAllTasks()
	data, _ := json.Marshal(tasks)
	return string(data)
}

func (a *App) CancelDownloadTask(taskID string) bool {
	return golang.GetTaskManager().CancelTask(taskID)
}

func (a *App) RemoveDownloadTask(taskID string) {
	golang.GetTaskManager().RemoveTask(taskID)
}

// ---------------- 通用消息 ----------------

func (a *App) SendToFrontend(message string) {
	runtime.EventsEmit(a.ctx, "back_msg", map[string]any{
		"time": getNowDateStr(),
		"msg":  message,
	})
}

// ---------------- 浏览器 ----------------

func (a *App) OpenUrl(openURL string) error {
	cmd := cmdWrapper.CommandVisible("cmd", "/c", "start", "", openURL)
	return cmd.Start()
}

func (a *App) OpenBrowser(openURL string) error {
	cmd := cmdWrapper.CommandVisible("rundll32", "url.dll,FileProtocolHandler", openURL)
	return cmd.Start()
}

// ---------------- 音乐搜索 ----------------

type MusicSearchResult struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Artist   string `json:"artist"`
	Album    string `json:"album"`
	Duration int    `json:"duration"`
	Source   string `json:"source"`
}

func (a *App) SearchMusic(keyword string) []MusicSearchResult {
	if keyword == "" {
		return nil
	}

	if resultsJSON, hit := sqllite.SearchMusicCache(keyword); hit {
		fmt.Println("从缓存获取搜索结果:", keyword)
		var results []MusicSearchResult
		if err := json.Unmarshal([]byte(resultsJSON), &results); err == nil {
			return results
		}
	}

	results := searchNetease(keyword)

	if len(results) > 0 {
		go func() {
			if resultsJSON, err := json.Marshal(results); err == nil {
				if err := sqllite.SaveSearchResults(keyword, string(resultsJSON)); err != nil {
					fmt.Println("保存搜索结果失败:", err)
				}
			}
		}()
	}

	return results
}

func (a *App) SaveMusicAudioSource(songID int64, source, audioURL, picURL, lrcURL string, fileSize int64) error {
	return sqllite.SaveAudioSource(songID, source, audioURL, picURL, lrcURL, fileSize)
}

func searchNetease(keyword string) []MusicSearchResult {
	apiUrl := "http://music.163.com/api/search/get/"
	data := "s=" + url.QueryEscape(keyword) + "&type=1&limit=20&offset=0"

	req, err := http.NewRequest("POST", apiUrl, strings.NewReader(data))
	if err != nil {
		return nil
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Referer", "http://music.163.com/")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	var result struct {
		Code   int `json:"code"`
		Result struct {
			Songs []struct {
				ID      int64  `json:"id"`
				Name    string `json:"name"`
				Artists []struct {
					Name string `json:"name"`
				} `json:"artists"`
				Album struct {
					Name string `json:"album"`
				} `json:"album"`
				Duration int `json:"duration"`
			} `json:"songs"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil
	}

	var songs []MusicSearchResult
	for _, s := range result.Result.Songs {
		artist := ""
		for i, ar := range s.Artists {
			if i > 0 {
				artist += "/"
			}
			artist += ar.Name
		}
		songs = append(songs, MusicSearchResult{
			ID:       s.ID,
			Name:     s.Name,
			Artist:   artist,
			Album:    s.Album.Name,
			Duration: s.Duration,
			Source:   "netease",
		})
	}
	return songs
}

// ---------------- 音乐解析 ----------------

func (a *App) GetRedirectUrl(redirectURL string) string {
	resp, err := http.Head(redirectURL)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	return resp.Request.URL.String()
}

type AudioFileInfo struct {
	URL         string `json:"url"`
	Size        int64  `json:"size"`
	ContentType string `json:"contentType"`
	Filename    string `json:"filename"`
}

func (a *App) GetFileInfo(fileURL string) string {
	resp, err := http.Head(fileURL)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	info := AudioFileInfo{
		URL:         resp.Request.URL.String(),
		Size:        resp.ContentLength,
		ContentType: resp.Header.Get("Content-Type"),
	}
	cd := resp.Header.Get("Content-Disposition")
	if cd != "" && strings.Contains(cd, "filename=") {
		parts := strings.Split(cd, "filename=")
		if len(parts) > 1 {
			info.Filename = strings.Trim(strings.Split(parts[1], ";")[0], "\" ")
		}
	}
	if info.Filename == "" {
		pathParts := strings.Split(resp.Request.URL.Path, "/")
		if len(pathParts) > 0 {
			info.Filename = pathParts[len(pathParts)-1]
		}
	}
	data, _ := json.Marshal(info)
	return string(data)
}

func (a *App) DownloadFile(fileURL string, fileType string, customFilename string, outputDir string) error {
	fmt.Printf("[DownloadFile] 开始下载 - url: %s, fileType: %s, customFilename: %s, outputDir: %s\n", fileURL, fileType, customFilename, outputDir)

	resp, err := http.Get(fileURL)
	if err != nil {
		return fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("下载失败，状态码: %d", resp.StatusCode)
	}

	musicDir := outputDir
	if musicDir == "" {
		musicDir = "D:\\音乐"
	}
	if err := os.MkdirAll(musicDir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	contentType := resp.Header.Get("Content-Type")
	var ext string
	switch fileType {
	case "pic":
		if strings.Contains(contentType, "png") {
			ext = ".png"
		} else if strings.Contains(contentType, "webp") {
			ext = ".webp"
		} else if strings.Contains(contentType, "gif") {
			ext = ".gif"
		} else {
			ext = ".jpg"
		}
	case "lrc":
		ext = ".lrc"
	default:
		if strings.Contains(contentType, "flac") {
			ext = ".flac"
		} else if strings.Contains(contentType, "wav") {
			ext = ".wav"
		} else if strings.Contains(contentType, "ogg") {
			ext = ".ogg"
		} else {
			ext = ".mp3"
		}
	}

	var filename string
	if customFilename != "" {
		safeFilename := strings.NewReplacer("/", "-", "\\", "-", ":", "-", "*", "-", "?", "-", "\"", "-", "<", "-", ">", "-", "|", "-").Replace(customFilename)
		filename = safeFilename + ext
	} else {
		filename = "download"
		if cd := resp.Header.Get("Content-Disposition"); cd != "" && strings.Contains(cd, "filename=") {
			parts := strings.Split(cd, "filename=")
			if len(parts) > 1 {
				filename = strings.Trim(strings.Split(parts[1], ";")[0], "\" ")
			}
		}
		if filename == "download" || filename == "" {
			pathParts := strings.Split(fileURL, "/")
			if len(pathParts) > 0 {
				filename = pathParts[len(pathParts)-1]
				if strings.Contains(filename, "?") {
					filename = strings.Split(filename, "?")[0]
				}
			}
			if filename == "" {
				filename = "download" + ext
			}
		}
	}

	outputPath := filepath.Join(musicDir, filename)
	fmt.Printf("[DownloadFile] 保存到: %s\n", outputPath)

	out, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %v", err)
	}
	defer out.Close()

	if _, err = io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}
	fmt.Printf("[DownloadFile] 下载完成: %s\n", outputPath)
	return nil
}

// ---------------- 内部工具 ----------------

func (a *App) downloadsDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return core.GetTempDir()
	}
	downloadDir := filepath.Join(homeDir, "Downloads")
	if _, err := os.Stat(downloadDir); os.IsNotExist(err) {
		return homeDir
	}
	return downloadDir
}

func getNowDateStr() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
