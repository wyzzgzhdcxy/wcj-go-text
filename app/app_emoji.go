package app

import (
	"fmt"
	"os"
	"path/filepath"

	"wcj-go-text/golang/myImage"
	"wcj-go-text/golang/sqllite"
)

// emojiCacheDir emoji 图片本地缓存目录
var emojiCacheDir = `C:\Users\wangchaojun\AppData\Local\wtools\tmp\emoji`

// SaveEmojiToCache 保存 Emoji 到缓存
func (a *App) SaveEmojiToCache(emoji string) string {
	hash := imgSha256Hash(emoji)
	id := hash[:16]

	if png, err := sqllite.GetEmojiPngData(id); err == nil && len(png) > 0 {
		fmt.Printf("[SaveEmojiToCache] DB hit for id=%s\n", id)
		return id
	}

	fmt.Printf("[SaveEmojiToCache] DB miss for emoji=%s, calling external API\n", emoji)
	pngPath, icoPath := myImage.SaveEmojiToCacheWithDir(emoji, emojiCacheDir)
	if pngPath == "" || icoPath == "" {
		fmt.Printf("[SaveEmojiToCache] External API failed, pngPath=%s, icoPath=%s\n", pngPath, icoPath)
		return ""
	}

	pngData, err := os.ReadFile(pngPath)
	if err != nil {
		fmt.Printf("[SaveEmojiToCache] Read png failed: %v\n", err)
		return ""
	}
	icoData, err := os.ReadFile(icoPath)
	if err != nil {
		fmt.Printf("[SaveEmojiToCache] Read ico failed: %v\n", err)
		return ""
	}

	if err := sqllite.SaveEmojiImage(id, pngData, icoData, emoji); err != nil {
		fmt.Printf("[SaveEmojiToCache] DB insert failed: %v\n", err)
	}

	fmt.Printf("[SaveEmojiToCache] Success, id=%s, pngSize=%d, icoSize=%d\n", id, len(pngData), len(icoData))
	return id
}

// GetEmojiImageUrl 获取 Emoji 图片的 HTTP URL（DB 兜底）
func (a *App) GetEmojiImageUrl(emoji string, imgType string) string {
	hash := imgSha256Hash(emoji)
	id := hash[:16]
	pngDir := emojiCacheDir
	os.MkdirAll(pngDir, 0755)
	var filePath string
	if imgType == "ico" {
		filePath = filepath.Join(pngDir, hash+".ico")
	} else {
		filePath = filepath.Join(pngDir, hash+".png")
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		var data []byte
		var readErr error
		if imgType == "ico" {
			data, readErr = sqllite.GetEmojiIcoData(id)
		} else {
			data, readErr = sqllite.GetEmojiPngData(id)
		}
		if readErr == nil && len(data) > 0 {
			os.WriteFile(filePath, data, 0644)
		} else {
			a.SaveEmojiToCache(emoji)
		}
	}

	return a.GetFileUrl(filePath)
}

// OpenEmojiFolder 打开 emoji 图片文件夹
func (a *App) OpenEmojiFolder() error {
	os.MkdirAll(emojiCacheDir, 0755)
	a.OpenExplorer(emojiCacheDir)
	return nil
}