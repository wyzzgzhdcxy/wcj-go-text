package myImage

import (
	"fmt"
	"image/png"
	"os"
	"strings"

	ico "github.com/Kodeworks/golang-image-ico"
	"github.com/wyzzgzhdcxy/wcj-go-common/core"
)

func PngToIcon(pngPath string) {
	gifPath := strings.TrimSuffix(pngPath, "png") + "gif"
	bmpPath := strings.TrimSuffix(pngPath, "png") + "bmp"
	webpPath := strings.TrimSuffix(pngPath, "png") + "webp"
	PngToIcon1(pngPath)
	_ = ConvertImage(pngPath, gifPath)
	_ = ConvertImage(pngPath, bmpPath)
	_ = ConvertImage(pngPath, webpPath)
}

func PngToIcon1(pngPath string) {
	pngFile, err := os.Open(pngPath)
	if err != nil {
		fmt.Println("打开PNG文件失败:", err)
		return
	}
	defer core.Close(pngFile)
	img, err := png.Decode(pngFile)
	if err != nil {
		fmt.Println("解码PNG失败:", err)
		return
	}
	icoPath := strings.TrimSuffix(pngPath, "png") + "ico"
	icoFile, err := os.Create(icoPath)
	if err != nil {
		fmt.Println("创建ICO文件失败:", err)
		return
	}
	defer core.Close(icoFile)
	err = ico.Encode(icoFile, img)
	if err != nil {
		fmt.Println("转换ICO失败:", err)
		return
	}
	fmt.Println("成功将PNG转换为ICO:", icoPath)
}

func PngToJpg(pngPath string) {
	jpgPath := strings.TrimSuffix(pngPath, "png") + "jpg"
	_ = ConvertImage(pngPath, jpgPath)
}

func Font2Image(content string, size int, pngPath string) {
	TextToPng(pngPath, content, size)
}
