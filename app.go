package main

import (
	"context"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"net/url"
	"strconv"
	"strings"
	"unicode"

	"github.com/atotto/clipboard"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// SetTitle 更新窗口标题
func (a *App) SetTitle(title string) {
	runtime.WindowSetTitle(a.ctx, title)
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