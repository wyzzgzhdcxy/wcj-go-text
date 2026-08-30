package app

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
	// 统一走完整文字渲染（圆角为 0 即直角），原来 cornerRadius=0 时走 letteravatar 只画第一个字符
	myImage.DrawStringAnchoredWithFilename(text, size, float64(cornerRadius), hash)
	if err := myImage.PngToJpg(pngPath); err != nil {
		fmt.Printf("生成JPG失败: %v\n", err)
	}
	pngFullPath := output + "/" + hash + ".png"
	jpgFullPath := output + "/" + hash + ".jpg"
	icoFullPath := output + "/" + hash + ".ico"
	fs1 := myImage.ReadFileWithBase64Encode(pngFullPath)
	fs2 := myImage.ReadFileWithBase64Encode(jpgFullPath)
	fs3 := myImage.ReadFileWithBase64Encode(icoFullPath)
	// 预览必须用 data URL：应用内没有本地 HTTP 文件服务，
	// 之前拼的 http://localhost:45670/file?... 地址无法加载，页面上三张图一直是裂的
	return myImage.ImageRes{
		PngMimeType: fs1.MimeType,
		PngEncoded:  fs1.Encoded,
		PngUrl:      myImage.DataUrl(fs1),
		JpgMimeType: fs2.MimeType,
		JpgEncoded:  fs2.Encoded,
		JpgUrl:      myImage.DataUrl(fs2),
		IcoMimeType: fs3.MimeType,
		IcoEncoded:  fs3.Encoded,
		// ICO 魔数 sniff 不到标准图片类型，直接显式指定
		IcoUrl: "data:image/x-icon;base64," + fs3.Encoded,
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
	if strings.TrimSpace(pngPath) == "" {
		return fmt.Errorf("请先导入 PNG 图片")
	}
	return myImage.PngToIcon(pngPath)
}

// OpenImgExplorer 打开图片文件夹
func (a *App) OpenImgExplorer() error {
	pngDir := core.GetTempDir() + "/png"
	os.MkdirAll(pngDir, 0755)
	a.OpenExplorer(pngDir)
	return nil
}
