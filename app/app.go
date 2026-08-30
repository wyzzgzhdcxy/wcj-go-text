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
	"encoding/base32"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"html"
	"io"
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
	"unicode/utf16"
	"unicode/utf8"

	"github.com/atotto/clipboard"
	"github.com/tjfoc/gmsm/sm3"
	"github.com/wailsapp/wails/v2/pkg/options"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/wyzzgzhdcxy/wcj-go-common/core"
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

// Startup Wails 启动回调，仅保存 ctx；DB 初始化已改为懒加载（sync.Once），
// 模板解压已改为懒加载，在 BuildProject 首次调用时执行。
// 导出供 main.go 在 options.App.OnStartup 中绑定。
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	a.registerMovieEvents()
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

// TextEncode 文本编码解码
func (a *App) TextEncode(plainText string, opType string) string {
	if len(plainText) == 0 {
		return plainText
	}

	opType = strings.ToLower(strings.TrimSpace(opType))

	switch opType {
	case "url编码", "url encode":
		return urlEscape(plainText)
	case "url解码", "url decode":
		decoded, err := url.QueryUnescape(plainText)
		if err != nil {
			return "URL解码错误: " + err.Error()
		}
		return decoded
	case "base64编码", "base64 encode":
		return base64.StdEncoding.EncodeToString([]byte(plainText))
	case "base64解码", "base64 decode":
		decoded, err := decodeBase64Lines(plainText)
		if err != nil {
			return "Base64解码错误: " + err.Error()
		}
		return string(decoded)
	case "base32编码", "base32 encode":
		return base32.StdEncoding.EncodeToString([]byte(plainText))
	case "base32解码", "base32 decode":
		decoded, err := decodeBase32Flexible(plainText)
		if err != nil {
			return "Base32解码错误: " + err.Error()
		}
		return string(decoded)
	case "hex编码", "hex encode":
		return hex.EncodeToString([]byte(plainText))
	case "hex解码", "hex decode":
		decoded, err := hex.DecodeString(stripHexNoise(plainText))
		if err != nil {
			return "Hex解码错误: " + err.Error()
		}
		return string(decoded)
	case "unicode编码", "unicode encode":
		return unicodeEncode(plainText)
	case "unicode解码", "unicode decode":
		decoded, err := unicodeDecode(plainText)
		if err != nil {
			return "Unicode解码错误: " + err.Error()
		}
		return decoded
	case "html编码", "html encode":
		return html.EscapeString(plainText)
	case "html解码", "html decode":
		return html.UnescapeString(plainText)
	case "二进制编码", "binary encode":
		parts := make([]string, 0, len(plainText))
		for _, b := range []byte(plainText) {
			parts = append(parts, fmt.Sprintf("%08b", b))
		}
		return strings.Join(parts, " ")
	case "二进制解码", "binary decode":
		decoded, err := binaryDecode(plainText)
		if err != nil {
			return "二进制解码错误: " + err.Error()
		}
		return decoded
	case "md5":
		return fmt.Sprintf("%x", md5.Sum([]byte(plainText)))
	case "sha1":
		hash := sha1.Sum([]byte(plainText))
		return hex.EncodeToString(hash[:])
	case "sha224":
		hash := sha256.Sum224([]byte(plainText))
		return hex.EncodeToString(hash[:])
	case "sha256":
		hash := sha256.Sum256([]byte(plainText))
		return hex.EncodeToString(hash[:])
	case "sha384":
		hash := sha512.Sum384([]byte(plainText))
		return hex.EncodeToString(hash[:])
	case "sha512":
		hash := sha512.Sum512([]byte(plainText))
		return hex.EncodeToString(hash[:])
	case "sm3":
		h := sm3.New()
		h.Write([]byte(plainText))
		return hex.EncodeToString(h.Sum(nil))
	case "crc32":
		return fmt.Sprintf("%08x", crc32.ChecksumIEEE([]byte(plainText)))
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
			if err != nil || code < 0 || code > unicode.MaxRune || !utf8.ValidRune(rune(code)) {
				return "ASCII解码错误: 非法码值 " + part
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
	case "去除空格", "压缩空格", "trim space":
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

// camelToSnake 驼峰转下划线：小写/数字后接大写、或大写缩略词接小写单词时插入下划线
// （JSONData → json_data、HTTPServer → http_server、fooBar → foo_bar）
func camelToSnake(s string) string {
	runes := []rune(s)
	var result []rune
	for i, r := range runes {
		if unicode.IsUpper(r) && i > 0 {
			prevLowerOrDigit := unicode.IsLower(runes[i-1]) || unicode.IsDigit(runes[i-1])
			prevUpperAndNextLower := unicode.IsUpper(runes[i-1]) && i+1 < len(runes) && unicode.IsLower(runes[i+1])
			if prevLowerOrDigit || prevUpperAndNextLower {
				result = append(result, '_')
			}
			result = append(result, unicode.ToLower(r))
		} else if unicode.IsUpper(r) {
			result = append(result, unicode.ToLower(r))
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

// urlEscape 与 JavaScript 的 encodeURIComponent 行为一致：
// 仅保留 A-Za-z0-9-_.~，空格编码为 %20（修正 QueryEscape 将空格编为 "+" 的问题），十六进制大写
func urlEscape(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') ||
			c == '-' || c == '_' || c == '.' || c == '~' {
			b.WriteByte(c)
		} else {
			fmt.Fprintf(&b, "%%%02X", c)
		}
	}
	return b.String()
}

// stripWhitespace 去除字符串中的所有空白字符
func stripWhitespace(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, s)
}

// stripHexNoise 去除 HEX 文本中的常见分隔符（空白/冒号/逗号/连字符/0x 前缀），
// 兼容 "68 65 6c"、"68:65:6c" 等粘贴格式
func stripHexNoise(s string) string {
	s = strings.ReplaceAll(s, "0x", "")
	s = strings.ReplaceAll(s, "0X", "")
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) || r == ':' || r == ',' || r == '-' {
			return -1
		}
		return r
	}, s)
}

// decodeBase64Flexible 解码 Base64：忽略空白，兼容 URL-safe 与无填充变体
func decodeBase64Flexible(s string) ([]byte, error) {
	s = stripWhitespace(s)
	for _, enc := range []*base64.Encoding{
		base64.StdEncoding, base64.RawStdEncoding,
		base64.URLEncoding, base64.RawURLEncoding,
	} {
		if b, err := enc.DecodeString(s); err == nil {
			return b, nil
		}
	}
	return nil, fmt.Errorf("不是合法的 Base64 字符串")
}

// decodeBase64Lines 解码 Base64：
// 优先整体解码（兼容单个长值被换行拆开的粘贴格式）；
// 整体失败时按行逐条解码（每行一个独立值，结果按行拼接），某行非法则报错定位到行
func decodeBase64Lines(s string) ([]byte, error) {
	if b, err := decodeBase64Flexible(s); err == nil {
		return b, nil
	}
	var nonEmpty []string
	for _, l := range strings.Split(s, "\n") {
		if t := strings.TrimSpace(l); t != "" {
			nonEmpty = append(nonEmpty, t)
		}
	}
	if len(nonEmpty) <= 1 {
		return nil, fmt.Errorf("不是合法的 Base64 字符串")
	}
	var out []byte
	for i, l := range nonEmpty {
		b, err := decodeBase64Flexible(l)
		if err != nil {
			return nil, fmt.Errorf("第 %d 行不是合法的 Base64", i+1)
		}
		out = append(out, b...)
		if i < len(nonEmpty)-1 {
			out = append(out, '\n')
		}
	}
	return out, nil
}

// decodeBase32Flexible 解码 Base32：忽略空白、大小写不敏感、兼容无填充变体
func decodeBase32Flexible(s string) ([]byte, error) {
	s = strings.ToUpper(stripWhitespace(s))
	noPad := base32.StdEncoding.WithPadding(base32.NoPadding)
	for _, enc := range []*base32.Encoding{
		base32.StdEncoding, noPad,
	} {
		if b, err := enc.DecodeString(s); err == nil {
			return b, nil
		}
	}
	return nil, fmt.Errorf("不是合法的 Base32 字符串")
}

// unicodeEncode 编码为 \uXXXX 转义序列，超出 BMP 的字符按 UTF-16 代理对展开（与 JS JSON.stringify 一致）
func unicodeEncode(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r > 0xFFFF {
			hi, lo := utf16.EncodeRune(r)
			fmt.Fprintf(&b, "\\u%04x\\u%04x", hi, lo)
		} else {
			fmt.Fprintf(&b, "\\u%04x", r)
		}
	}
	return b.String()
}

// unicodeDecode 解析 \uXXXX 转义（自动合并代理对），其余字符原样保留
func unicodeDecode(s string) (string, error) {
	var b strings.Builder
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		if runes[i] != '\\' || i+5 >= len(runes) || (runes[i+1] != 'u' && runes[i+1] != 'U') {
			b.WriteRune(runes[i])
			continue
		}
		v, err := strconv.ParseUint(string(runes[i+2:i+6]), 16, 32)
		if err != nil {
			return "", fmt.Errorf("非法的 Unicode 转义")
		}
		code := rune(v)
		// 高位代理后面紧跟低位代理时合并为完整字符（如 😀 → \ud83d\ude00）
		if utf16.IsSurrogate(code) && i+11 < len(runes) && runes[i+6] == '\\' && (runes[i+7] == 'u' || runes[i+7] == 'U') {
			v2, err := strconv.ParseUint(string(runes[i+8:i+12]), 16, 32)
			if err == nil && utf16.IsSurrogate(rune(v2)) {
				if r := utf16.DecodeRune(code, rune(v2)); r != unicode.ReplacementChar {
					b.WriteRune(r)
					i += 11
					continue
				}
			}
		}
		if utf16.IsSurrogate(code) {
			b.WriteRune(unicode.ReplacementChar)
		} else {
			b.WriteRune(code)
		}
		i += 5
	}
	return b.String(), nil
}

// binaryDecode 解码二进制文本：支持空格分隔的任意宽度二进制分组，
// 也支持连续的 8 位二进制串（如 0100000101100010）
func binaryDecode(s string) (string, error) {
	tokens := strings.Fields(s)
	if len(tokens) == 0 {
		return "", nil
	}
	// 无分隔的连续二进制串按 8 位一组切分
	if len(tokens) == 1 && len(tokens[0])%8 == 0 && isBinaryDigits(tokens[0]) {
		t := tokens[0]
		out := make([]byte, 0, len(t)/8)
		for i := 0; i < len(t); i += 8 {
			v, err := strconv.ParseUint(t[i:i+8], 2, 32)
			if err != nil {
				return "", fmt.Errorf("非法的二进制字符")
			}
			out = append(out, byte(v))
		}
		return string(out), nil
	}
	var out []byte
	for _, t := range tokens {
		v, err := strconv.ParseUint(t, 2, 32)
		if err != nil || v > 255 {
			return "", fmt.Errorf("非法的二进制分组 %s", t)
		}
		out = append(out, byte(v))
	}
	return string(out), nil
}

func isBinaryDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] != '0' && s[i] != '1' {
			return false
		}
	}
	return true
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
	cmd := core.Command("wails", "build")
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
