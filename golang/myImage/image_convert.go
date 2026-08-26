package myImage

import (
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/image/bmp"
)

var supportedFormats = map[string]bool{
	"jpg":  true,
	"jpeg": true,
	"png":  true,
	"gif":  true,
	"bmp":  true,
	"webp": true,
}

func ConvertImage(inputPath, outputPath string) error {
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return fmt.Errorf("输入文件不存在: %s", inputPath)
	}

	outputExt := strings.ToLower(filepath.Ext(outputPath))
	if len(outputExt) == 0 {
		return errors.New("输出文件缺少扩展名")
	}
	outputExt = outputExt[1:]

	if !supportedFormats[outputExt] {
		return fmt.Errorf("不支持的输出格式: %s", outputExt)
	}

	inputFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("无法打开输入文件: %v", err)
	}
	defer inputFile.Close()

	img, format, err := image.Decode(inputFile)
	if err != nil {
		return fmt.Errorf("无法解码图片: %v", err)
	}

	if !supportedFormats[format] {
		return fmt.Errorf("不支持的输入格式: %s", format)
	}

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("无法创建输出文件: %v", err)
	}
	defer outputFile.Close()

	switch outputExt {
	case "jpeg", "jpg":
		err = jpeg.Encode(outputFile, img, &jpeg.Options{Quality: 90})
	case "png":
		err = png.Encode(outputFile, img)
	case "gif":
		err = gif.Encode(outputFile, img, &gif.Options{})
	case "bmp":
		err = bmp.Encode(outputFile, img)
	case "webp":
		return errors.New("WEBP编码需要第三方库支持")
	default:
		err = fmt.Errorf("不支持的输出格式: %s", outputExt)
	}

	if err != nil {
		return fmt.Errorf("编码图片失败: %v", err)
	}

	return nil
}

func IsFormatSupported(format string) bool {
	return supportedFormats[strings.ToLower(format)]
}

func GetSupportedFormats() []string {
	formats := make([]string, 0, len(supportedFormats))
	for f := range supportedFormats {
		formats = append(formats, f)
	}
	return formats
}
