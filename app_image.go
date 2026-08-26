package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"strings"

	"wcj-go-text/golang/myImage"

	"github.com/wyzzgzhdcxy/wcj-go-common/core"
)

func imgSha256Hash(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))[:16]
}

// Font2Image 生成字体图片（PNG/JPG/ICO）
func (a *App) Font2Image(text string, size int, cornerRadius int) myImage.ImageRes {
	output := core.GetTempDir() + "/png"
	core.MkDirALl0755(output)
	hash := fmt.Sprintf("%d_%d_%x", size, cornerRadius, imgSha256Hash(text))
	pngPath := output + "/" + hash + ".png"
	if cornerRadius == 0 {
		myImage.Font2Image(text, size, pngPath)
	} else {
		myImage.DrawStringAnchoredWithFilename(text, size, float64(cornerRadius), hash)
	}
	myImage.PngToIcon(pngPath)
	myImage.PngToJpg(pngPath)
	pngFullPath := output + "/" + hash + ".png"
	jpgFullPath := output + "/" + hash + ".jpg"
	icoFullPath := output + "/" + hash + ".ico"
	fs1 := myImage.ReadFileWithBase64Encode(pngFullPath)
	fs2 := myImage.ReadFileWithBase64Encode(jpgFullPath)
	fs3 := myImage.ReadFileWithBase64Encode(icoFullPath)
	return myImage.ImageRes{
		PngMimeType: fs1.MimeType,
		PngEncoded:  fs1.Encoded,
		PngUrl:      a.GetFileUrl(pngFullPath),
		JpgMimeType: fs2.MimeType,
		JpgEncoded:  fs2.Encoded,
		JpgUrl:      a.GetFileUrl(jpgFullPath),
		IcoMimeType: fs3.MimeType,
		IcoEncoded:  fs3.Encoded,
		IcoUrl:      a.GetFileUrl(icoFullPath),
	}
}

// GetFileUrl 通过本地 HTTP 服务获取文件的 URL
func (a *App) GetFileUrl(filePath string) string {
	filePath = strings.ReplaceAll(filePath, "\\", "/")
	encodedPath := url.QueryEscape(filePath)
	return fmt.Sprintf("http://localhost:45670/file?path=%s", encodedPath)
}

// PngToIcon PNG转ICO（含其他格式）
func (a *App) PngToIcon(pngPath string) error {
	myImage.PngToIcon(pngPath)
	return nil
}

// OpenImgExplorer 打开图片文件夹
func (a *App) OpenImgExplorer() error {
	pngDir := core.GetTempDir() + "/png"
	os.MkdirAll(pngDir, 0755)
	a.OpenExplorer(pngDir)
	return nil
}
