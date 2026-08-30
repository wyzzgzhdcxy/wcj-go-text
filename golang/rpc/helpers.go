package rpc

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
)

// decodeHex 解码 hex 字符串；失败时回退到 base64。
func decodeHex(s string) ([]byte, error) {
	if b, err := hex.DecodeString(s); err == nil {
		return b, nil
	}
	if b, err := base64.StdEncoding.DecodeString(s); err == nil {
		return b, nil
	}
	return nil, fmt.Errorf("非 hex/base64 编码")
}

// detectAudioFormat 根据字节头识别常见音频格式。
func detectAudioFormat(data []byte) string {
	if len(data) < 4 {
		return ".mp3"
	}
	// MP3: FF FB / FF F3 / ID3 等
	if data[0] == 0xFF && (data[1]&0xE0) == 0xE0 {
		return ".mp3"
	}
	// ID3 标签
	if len(data) >= 3 && string(data[0:3]) == "ID3" {
		return ".mp3"
	}
	// WAV
	if len(data) >= 12 && string(data[0:4]) == "RIFF" && string(data[8:12]) == "WAVE" {
		return ".wav"
	}
	// OGG
	if string(data[0:4]) == "OggS" {
		return ".ogg"
	}
	// FLAC
	if string(data[0:4]) == "fLaC" {
		return ".flac"
	}
	// M4A/AAC
	if len(data) >= 8 && string(data[4:8]) == "ftyp" {
		return ".m4a"
	}
	return ".mp3"
}

// encodeImageAsDataURL 读取本地图片文件并以 data URL 形式返回。
func encodeImageAsDataURL(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	mime := "image/jpeg"
	if strings.HasSuffix(strings.ToLower(path), ".png") {
		mime = "image/png"
	}
	return fmt.Sprintf("data:%s;base64,%s", mime, base64.StdEncoding.EncodeToString(data)), nil
}

