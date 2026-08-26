package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image/png"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// 内嵌的模板文件内容
const templateConstantsGo = `package main

// 目标网站 URL - 在这里修改要打开的网站
const TargetURL = "{{TargetURL}}"
const Version = "0.1.0"
const title = "{{title}}"
`

const templateWailsJSON = `{
  "$schema": "https://wails.io/schemas/config.v2.json",

 "frontend": {
    "build": "npm run build",
    "dev": "npm run dev"
  },

  "author": {
    "name": "",
    "email": ""
  },
  "frontend:build": "npm run build",
  "frontend:install": "npm install",
  "name": "{{appName}}",
  "outputfilename": "{{appName}}"
}
`

// updateTemplateFiles generates and writes constants.go and wails.json to projectDir
func updateTemplateFiles(projectDir, appName, title, redirectURL string) error {
	// Write constants.go
	constantsContent := strings.ReplaceAll(templateConstantsGo, "{{TargetURL}}", redirectURL)
	constantsContent = strings.ReplaceAll(constantsContent, "{{title}}", title)
	if err := os.WriteFile(filepath.Join(projectDir, "constants.go"), []byte(constantsContent), 0644); err != nil {
		return err
	}

	// Write wails.json
	wailsContent := strings.ReplaceAll(templateWailsJSON, "{{appName}}", appName)
	return os.WriteFile(filepath.Join(projectDir, "wails.json"), []byte(wailsContent), 0644)
}

// fetchFavicon tries multiple paths to get the website icon, returns the first successful one
func fetchFavicon(redirectURL string) (string, error) {
	u, err := url.Parse(redirectURL)
	if err != nil || u.Host == "" {
		return "", fmt.Errorf("invalid URL: %s", redirectURL)
	}

	// Try multiple icon paths in order of preference
	iconPaths := []string{
		"/apple-touch-icon.png",
		"/apple-touch-icon-precomposed.png",
		"/favicon.png",
		"/favicon.ico",
		"/favicon-192x192.png",
		"/favicon-256x256.png",
		"/favicon-512x512.png",
		"/icons/icon.png",
		"/images/icon.png",
		"/img/favicon.ico",
	}

	client := &http.Client{Timeout: 1 * time.Second}

	for _, iconPath := range iconPaths {
		iconURL := fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, iconPath)

		resp, err := client.Get(iconURL)
		if err != nil {
			continue
		}
		resp.Body.Close()

		if resp.StatusCode != 200 {
			continue
		}

		// Download
		resp2, err := client.Get(iconURL)
		if err != nil {
			continue
		}
		data, err := io.ReadAll(resp2.Body)
		resp2.Body.Close()
		if err != nil || len(data) == 0 {
			continue
		}

		// Validate it's an image
		isICO := len(data) > 4 && data[0] == 0x00 && data[1] == 0x00 && data[2] == 0x01 && data[3] == 0x00
		isPNG := len(data) > 4 && data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47
		isJPEG := len(data) > 2 && data[0] == 0xFF && data[1] == 0xD8

		if !isICO && !isPNG && !isJPEG {
			continue
		}

		fmt.Fprintf(os.Stdout, "  ✅ 找到图标: %s (%d bytes)\n", iconPath, len(data))

		// Save to temp file
		ext := ".ico"
		if isPNG {
			ext = ".png"
		} else if isJPEG {
			ext = ".jpg"
		}
		tmpFile := filepath.Join(os.TempDir(), "wails_icon"+ext)
		if err := os.WriteFile(tmpFile, data, 0644); err != nil {
			fmt.Fprintf(os.Stdout, "  ❌ 保存失败: %v\n", err)
			continue
		}

		fmt.Fprintf(os.Stdout, "  ✅ 已保存到: %s\n", tmpFile)
		return tmpFile, nil
	}

	return "", fmt.Errorf("未找到网站图标")
}

// copyIconFiles copies the icon to build/windows/icon.ico
func copyIconFiles(projectDir, iconPath string) error {
	data, err := os.ReadFile(iconPath)
	if err != nil {
		return fmt.Errorf("read icon file: %w", err)
	}

	isICO := len(data) > 4 && data[0] == 0x00 && data[1] == 0x00 && data[2] == 0x01 && data[3] == 0x00

	icoDst := filepath.Join(projectDir, "build", "windows", "icon.ico")

	if isICO {
		// ICO file - copy directly
		if err := os.WriteFile(icoDst, data, 0644); err != nil {
			return fmt.Errorf("write icon.ico: %w", err)
		}
	} else {
		// PNG or other - convert to ICO
		if err := convertPNGToICO(iconPath, icoDst); err != nil {
			return fmt.Errorf("convert to icon.ico: %w", err)
		}
	}

	return nil
}

// convertPNGToICO embeds a PNG image inside a .ico file (Windows Vista+ compatible)
func convertPNGToICO(pngPath, icoPath string) error {
	// Read and validate PNG
	f, err := os.Open(pngPath)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err = png.Decode(f); err != nil {
		return fmt.Errorf("invalid PNG: %w", err)
	}

	// Re-read raw bytes
	if _, err = f.Seek(0, io.SeekStart); err != nil {
		return err
	}
	pngData, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	// Build ICO in a buffer
	var buf bytes.Buffer
	le := binary.LittleEndian

	// ICONDIR header (6 bytes)
	binary.Write(&buf, le, uint16(0)) // reserved
	binary.Write(&buf, le, uint16(1)) // type = 1 (icon)
	binary.Write(&buf, le, uint16(1)) // count = 1

	// ICONDIRENTRY (16 bytes)
	imageOffset := uint32(6 + 16) // header + one entry
	binary.Write(&buf, le, uint8(0))              // width  (0 = 256)
	binary.Write(&buf, le, uint8(0))              // height (0 = 256)
	binary.Write(&buf, le, uint8(0))              // color count
	binary.Write(&buf, le, uint8(0))              // reserved
	binary.Write(&buf, le, uint16(1))             // planes
	binary.Write(&buf, le, uint16(32))            // bit count
	binary.Write(&buf, le, uint32(len(pngData)))  // bytes in resource
	binary.Write(&buf, le, imageOffset)           // offset to image data

	// PNG data
	buf.Write(pngData)

	return os.WriteFile(icoPath, buf.Bytes(), 0644)
}
