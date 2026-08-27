// Package app 包含 Wails 应用的全部 UI 绑定方法、App 生命周期、构建管线等。
// 由 main.go 通过 wcj-go-text/app 引入并绑定到前端。
package app

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"embed"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	goruntime "runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode"

	"wcj-go-text/golang"
	"wcj-go-text/golang/cmdWrapper"

	"github.com/atotto/clipboard"
	"github.com/wailsapp/wails/v2/pkg/options"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// Assets 嵌入式前端资源（frontend/dist），由 main.go 在 init() 中注入。
// app_settings.go / app_file.go / ReadAssetFile 都通过它读取内嵌文件。
var Assets embed.FS

// TplFS 嵌入式模板资源（tpl 目录），由 main.go 在 init() 中注入，
// 用于应用生成器（BuildProject）展开 my_deepseek.zip 模板。
var TplFS embed.FS

// App Wails 绑定的应用对象。所有以 (a *App) 为接收者的导出方法都会暴露给前端。
type App struct {
	ctx          context.Context
	templateDir  string
	templateOnce sync.Once
}

// BuildTime 由 wails build 通过 -ldflags 在编译期注入，例如：
//
//	wails build -ldflags "-X wcj-go-text/app.BuildTime=2026-08-26 11:00:00"
var BuildTime string

// startupStart 进程启动起点（包级 var 初始化时打点），由 SetStartupStart 注入。
// GetStartupTime 用它计算"Go runtime init + 依赖包 init + Wails 启动 + 前端加载"全链路耗时。
var startupStart int64

// SetStartupStart 由 main.go 在最早时机（init 中）调用一次，
// 记录从进程启动到 Wails 启动回调开始之间的总耗时起点。
func SetStartupStart(ns int64) {
	atomic.StoreInt64(&startupStart, ns)
}

// KeepBuildTimeAlive 用 SetFinalizer 强制持有 BuildTime 指针，
// 防止编译器/链接器对 -X wcj-go-text/app.BuildTime 注入的符号进行 dead-code 优化。
// SetFinalizer 会让编译器无法证明该值无副作用，从而保留符号。
// 由 main.go 在 init() 中调用。
func KeepBuildTimeAlive() {
	v := BuildTime
	goruntime.SetFinalizer(&v, func(p *string) {})
}

// GetBuildTime 返回构建时间（注入自 -ldflags）。
//
//go:noinline
func (a *App) GetBuildTime() string {
	return BuildTime
}

// ProjectConfig 应用生成器配置
type ProjectConfig struct {
	AppName     string `json:"appName"`
	Title       string `json:"title"`
	IconPath    string `json:"iconPath"`
	OutputDir   string `json:"outputDir"`
	RedirectURL string `json:"redirectURL"`
}

// BuildResult 应用生成器结果
type BuildResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// NewApp 创建新的 App 实例。Wails 在启动时调用。
func NewApp() *App {
	return &App{}
}

// Startup Wails 启动回调，保存 ctx 并初始化数据库。
// 模板解压已改为懒加载，在 BuildProject 首次调用时执行。
// 导出供 main.go 在 options.App.OnStartup 中绑定。
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	a.registerMovieEvents()
	if err := a.initImageSettingsDb(); err != nil {
		log.Printf("初始化图片配置数据库失败: %v", err)
	}
	golang.GetToolsDB()
}

// OnSecondInstanceLaunch 第二次启动时把已有窗口拉到前台。
// 导出供 main.go 在 SingleInstanceLock.OnSecondInstanceLaunch 中绑定。
func (a *App) OnSecondInstanceLaunch(_ options.SecondInstanceData) {
	wailsruntime.WindowUnminimise(a.ctx)
	wailsruntime.WindowShow(a.ctx)
}

// SetTitle 更新窗口标题
func (a *App) SetTitle(title string) {
	wailsruntime.WindowSetTitle(a.ctx, title)
}

// GetStartupTime 返回从 main.init() 打点到当前调用的总耗时(毫秒)，
// 覆盖 Go runtime 初始化、依赖包 init、Wails 启动、前端加载完整链路。
func (a *App) GetStartupTime() int64 {
	start := atomic.LoadInt64(&startupStart)
	if start == 0 {
		return 0
	}
	return (time.Now().UnixNano() - start) / 1000000
}

// SelectFile 打开文件选择对话框
func (a *App) SelectFile() (string, error) {
	selection, err := wailsruntime.OpenFileDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title: "选择文件",
	})
	if err != nil {
		return "", err
	}
	return selection, nil
}

// ReadAssetFile 读取嵌入资源文件
func (a *App) ReadAssetFile(path string) (string, error) {
	data, err := Assets.ReadFile(strings.TrimPrefix(path, "/"))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// CopyToClipboard 复制文本到剪贴板
func (a *App) CopyToClipboard(text string) error {
	return clipboard.WriteAll(text)
}

// FileEncode 文件编码处理
func (a *App) FileEncode(plainText string, opType string) string {
	if len(plainText) == 0 {
		return plainText
	}

	opType = strings.ToLower(strings.TrimSpace(opType))

	switch opType {
	case "md5":
		return fmt.Sprintf("%x", md5.Sum([]byte(plainText)))
	case "sha1":
		hash := sha1.Sum([]byte(plainText))
		return hex.EncodeToString(hash[:])
	case "sha256":
		hash := sha256.Sum256([]byte(plainText))
		return hex.EncodeToString(hash[:])
	case "sha512":
		hash := sha512.Sum512([]byte(plainText))
		return hex.EncodeToString(hash[:])
	case "去除空格", "trim space":
		return strings.Join(strings.Fields(plainText), " ")
	default:
		return plainText
	}
}

// TextEncode 文本编码解码
func (a *App) TextEncode(plainText string, opType string) string {
	if len(plainText) == 0 {
		return plainText
	}

	opType = strings.ToLower(strings.TrimSpace(opType))

	switch opType {
	case "url编码", "url encode":
		return url.QueryEscape(plainText)
	case "url解码", "url decode":
		decoded, err := url.QueryUnescape(plainText)
		if err != nil {
			return "URL解码错误: " + err.Error()
		}
		return decoded
	case "base64编码", "base64 encode":
		return base64.StdEncoding.EncodeToString([]byte(plainText))
	case "base64解码", "base64 decode":
		decoded, err := base64.StdEncoding.DecodeString(plainText)
		if err != nil {
			return "Base64解码错误: " + err.Error()
		}
		return string(decoded)
	case "hex编码", "hex encode":
		return hex.EncodeToString([]byte(plainText))
	case "hex解码", "hex decode":
		decoded, err := hex.DecodeString(plainText)
		if err != nil {
			return "Hex解码错误: " + err.Error()
		}
		return string(decoded)
	case "md5":
		return fmt.Sprintf("%x", md5.Sum([]byte(plainText)))
	case "sha1":
		hash := sha1.Sum([]byte(plainText))
		return hex.EncodeToString(hash[:])
	case "sha256":
		hash := sha256.Sum256([]byte(plainText))
		return hex.EncodeToString(hash[:])
	case "sha512":
		hash := sha512.Sum512([]byte(plainText))
		return hex.EncodeToString(hash[:])
	case "ascii编码", "ascii encode":
		var result strings.Builder
		for _, r := range plainText {
			result.WriteString(fmt.Sprintf("%d ", r))
		}
		return strings.TrimSpace(result.String())
	case "ascii解码", "ascii decode":
		parts := strings.Fields(plainText)
		var result strings.Builder
		for _, part := range parts {
			code, err := strconv.Atoi(part)
			if err != nil {
				return "ASCII解码错误: " + err.Error()
			}
			result.WriteRune(rune(code))
		}
		return result.String()
	case "驼峰转下划线", "camel to snake":
		return camelToSnake(plainText)
	case "全大写", "upper case":
		return strings.ToUpper(plainText)
	case "全小写", "lower case":
		return strings.ToLower(plainText)
	case "反转字符串", "reverse":
		return reverseString(plainText)
	case "去除空格", "trim space":
		return strings.Join(strings.Fields(plainText), " ")
	default:
		return plainText
	}
}

// Text2Json 解析类似CSV格式的数据
func (a *App) Text2Json(content string, splitChar string) []map[string]string {
	lines1 := strings.Split(content, "\r\n")
	lines2 := strings.Split(content, "\n")

	var lines []string
	if len(lines2) > len(lines1) {
		lines = lines2
	} else {
		lines = lines1
	}

	var list []map[string]string

	if len(lines) == 0 {
		return list
	}

	headers := strings.Split(lines[0], splitChar)

	for i := 1; i < len(lines); i++ {
		row := make(map[string]string)
		columns := strings.Split(lines[i], splitChar)

		for j := 0; j < len(columns); j++ {
			if j < len(headers) {
				row[headers[j]] = columns[j]
			}
		}

		list = append(list, row)
	}

	return list
}

func camelToSnake(s string) string {
	var result []rune
	for i, r := range s {
		if unicode.IsUpper(r) && i > 0 {
			result = append(result, '_')
		}
		result = append(result, unicode.ToLower(r))
	}
	return string(result)
}

func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// ---------- 应用生成器 ----------

// extractTemplate 将内嵌的 my_deepseek.zip 模板解压到 targetDir。
// targetDir 为空时回退到 %LOCALAPPDATA%\WailsTemplate\my_deepseek。
// 若目标目录已存在且非空则跳过解压（懒加载：避免每次启动都重写磁盘）。
func (a *App) extractTemplate(targetDir string) {
	if targetDir == "" {
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			localAppData = os.TempDir()
		}
		targetDir = filepath.Join(localAppData, "WailsTemplate", "my_deepseek")
	}

	// 已存在且非空：复用现有目录，避免每次启动都重写整个 wails 脚手架
	if entries, err := os.ReadDir(targetDir); err == nil && len(entries) > 0 {
		a.templateDir = targetDir
		return
	}

	os.RemoveAll(targetDir)
	os.MkdirAll(targetDir, 0755)

	zipData, err := TplFS.ReadFile("tpl/my_deepseek.zip")
	if err != nil {
		return
	}

	zr, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return
	}

	for _, f := range zr.File {
		name := strings.TrimPrefix(f.Name, "my_deepseek/")
		if name == "" {
			continue
		}
		outPath := filepath.Join(targetDir, name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(outPath, 0755)
			continue
		}
		os.MkdirAll(filepath.Dir(outPath), 0755)

		rc, err := f.Open()
		if err != nil {
			continue
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			continue
		}
		os.WriteFile(outPath, data, 0644)
	}

	a.templateDir = targetDir
}

// GetTemplateDir 返回已解压模板目录的路径
func (a *App) GetTemplateDir() string {
	return a.templateDir
}

// SelectOutputDir 打开目录选择对话框，用于选择打包输出目录
func (a *App) SelectOutputDir() string {
	dir, err := wailsruntime.OpenDirectoryDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title: "选择打包输出目录",
	})
	if err != nil {
		return ""
	}
	return dir
}

// SelectIconFile 打开图标文件选择对话框
func (a *App) SelectIconFile() string {
	file, err := wailsruntime.OpenFileDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title: "选择图标文件 (ICO/PNG)",
		Filters: []wailsruntime.FileFilter{
			{DisplayName: "图标文件 (*.ico;*.png)", Pattern: "*.ico;*.png"},
		},
	})
	if err != nil {
		return ""
	}
	return file
}

// BuildProject 执行完整打包流水线：生成配置 -> 图标处理 -> wails build -> 导出
func (a *App) BuildProject(config ProjectConfig) BuildResult {
	log := func(msg string) {
		fmt.Fprintln(os.Stdout, msg)
	}

	if config.AppName == "" {
		return BuildResult{false, "请填写应用名称"}
	}

	// 懒加载：首次调用时才解压模板到磁盘（已存在则复用）
	a.templateOnce.Do(func() {
		a.extractTemplate("")
	})

	log("📦 开始打包...")

	log("⚙ 生成配置...")
	if err := updateTemplateFiles(a.templateDir, config.AppName, config.Title, config.RedirectURL); err != nil {
		return BuildResult{false, fmt.Sprintf("生成配置失败: %v", err)}
	}

	if config.IconPath != "" {
		if err := copyIconFiles(a.templateDir, config.IconPath); err != nil {
			log(fmt.Sprintf("⚠ 图标处理失败: %v", err))
		}
	} else if config.RedirectURL != "" {
		if iconPath, err := fetchFavicon(config.RedirectURL); err == nil && iconPath != "" {
			config.IconPath = iconPath
			copyIconFiles(a.templateDir, config.IconPath)
		}
	}

	log("🔨 编译中...")
	cmd := cmdWrapper.Command("wails", "build")
	cmd.Dir = a.templateDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return BuildResult{false, fmt.Sprintf("编译失败: %v", err)}
	}

	log("📤 导出中...")
	outputDir := config.OutputDir
	if outputDir == "" {
		outputDir = "D:\\"
	}
	srcExe := filepath.Join(a.templateDir, "build", "bin", config.AppName+".exe")
	if _, err := os.Stat(srcExe); os.IsNotExist(err) {
		srcExe = filepath.Join(a.templateDir, "build", "bin", config.AppName)
	}
	dstExe := filepath.Join(outputDir, filepath.Base(srcExe))
	if err := moveFile(srcExe, dstExe); err != nil {
		log(fmt.Sprintf("❌ 移动文件失败: %v", err))
		return BuildResult{false, fmt.Sprintf("移动文件失败: %v", err)}
	}
	log(fmt.Sprintf("✓ 已导出到: %s", dstExe))
	return BuildResult{true, "打包完成！"}
}

// copyFile 复制单个文件
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func moveFile(src, dst string) error {
	if err := copyFile(src, dst); err != nil {
		return err
	}
	return os.Remove(src)
}
