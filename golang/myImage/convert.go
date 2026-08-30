package myImage

import (
	"fmt"
	"image/png"
	"os"
	"strings"

	ico "github.com/Kodeworks/golang-image-ico"
	"github.com/wyzzgzhdcxy/wcj-go-common/core"
)

func PngToIcon(pngPath string) error {
	gifPath := strings.TrimSuffix(pngPath, "png") + "gif"
	bmpPath := strings.TrimSuffix(pngPath, "png") + "bmp"
	if err := PngToIcon1(pngPath); err != nil {
		return err
	}
	if err := ConvertImage(pngPath, gifPath); err != nil {
		return err
	}
	return ConvertImage(pngPath, bmpPath)
}

func PngToIcon1(pngPath string) error {
	pngFile, err := os.Open(pngPath)
	if err != nil {
		return fmt.Errorf("打开PNG文件失败: %v", err)
	}
	defer core.Close(pngFile)
	img, err := png.Decode(pngFile)
	if err != nil {
		return fmt.Errorf("解码PNG失败: %v", err)
	}
	icoPath := strings.TrimSuffix(pngPath, "png") + "ico"
	icoFile, err := os.Create(icoPath)
	if err != nil {
		return fmt.Errorf("创建ICO文件失败: %v", err)
	}
	defer core.Close(icoFile)
	if err = ico.Encode(icoFile, img); err != nil {
		return fmt.Errorf("转换ICO失败: %v", err)
	}
	return nil
}

func PngToJpg(pngPath string) error {
	jpgPath := strings.TrimSuffix(pngPath, "png") + "jpg"
	return ConvertImage(pngPath, jpgPath)
}

func Font2Image(content string, size int, pngPath string) {
	TextToPng(pngPath, content, size)
}
