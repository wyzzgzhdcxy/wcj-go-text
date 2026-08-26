package main

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
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unicode"

	"github.com/atotto/clipboard"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:tpl
var templateFS embed.FS

// App struct
type App struct {
	ctx          context.Context
	startupTime  int64
	templateDir  string
}

// BuildTime 由 wails build 通过 -ldflags 在编译期注入，例如：
//   wails build -ldflags "-X main.BuildTime=2026-08-26 11:00:00"
var BuildTime string

// GetBuildTime 返回构建时间（注入自 -ldflags）。
//
//go:noinline
func (a *App) GetBuildTime() string {
	return BuildTime
}

// ProjectConfig holds all build configuration
type ProjectConfig struct {
	AppName     string `json:"appName"`
	Title       string `json:"title"`
	IconPath    string `json:"iconPath"`
	OutputDir   string `json:"outputDir"`
	RedirectURL string `json:"redirectURL"`
}

// BuildResult holds the result of a build operation
type BuildResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		startupTime: time.Now().UnixNano(),
	}
}

// startup is called when the app starts. The context is saved
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.extractTemplate("")
	a.registerMovieEvents()
}

func (a *App) onSecondInstanceLaunch(_ options.SecondInstanceData) {
	runtime.WindowUnminimise(a.ctx)
	runtime.WindowShow(a.ctx)
}

// SetTitle 更新窗口标题
func (a *App) SetTitle(title string) {
	runtime.WindowSetTitle(a.ctx, title)
}

// GetStartupTime 获取启动耗时(毫秒)
func (a *App) GetStartupTime() int64 {
	return (time.Now().UnixNano() - a.startupTime) / 1000000
}

// SelectFile 打开文件选择对话框
func (a *App) SelectFile() (string, error) {
	selection, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择文件",
	})
	if err != nil {
		return "", err
	}
	return selection, nil
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

// extractTemplate extracts the embedded zip template to targetDir.
// If targetDir is empty, it falls back to %LOCALAPPDATA%\WailsTemplate\my_deepseek.
func (a *App) extractTemplate(targetDir string) {
	if targetDir == "" {
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			localAppData = os.TempDir()
		}
		targetDir = filepath.Join(localAppData, "WailsTemplate", "my_deepseek")
	}

	os.RemoveAll(targetDir)
	os.MkdirAll(targetDir, 0755)

	zipData, err := templateFS.ReadFile("tpl/my_deepseek.zip")
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

// GetTemplateDir returns the path to the extracted template directory
func (a *App) GetTemplateDir() string {
	return a.templateDir
}

// SelectOutputDir opens a directory dialog for selecting the output directory
func (a *App) SelectOutputDir() string {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择打包输出目录",
	})
	if err != nil {
		return ""
	}
	return dir
}

// SelectIconFile opens a file dialog for selecting an icon
func (a *App) SelectIconFile() string {
	file, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择图标文件 (ICO/PNG)",
		Filters: []runtime.FileFilter{
			{DisplayName: "图标文件 (*.ico;*.png)", Pattern: "*.ico;*.png"},
		},
	})
	if err != nil {
		return ""
	}
	return file
}

// BuildProject performs the full build pipeline
func (a *App) BuildProject(config ProjectConfig) BuildResult {
	log := func(msg string) {
		fmt.Fprintln(os.Stdout, msg)
	}

	if config.AppName == "" {
		return BuildResult{false, "请填写应用名称"}
	}

	log("📦 开始打包...")
	a.extractTemplate("")

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
	cmd := exec.Command("wails", "build")
	cmd.Dir = a.templateDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
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