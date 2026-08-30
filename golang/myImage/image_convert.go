package myImage

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/image/bmp"
)

// 不含 webp：项目没有 webp 编码器，之前在已创建输出文件后才报错，
// 会留下空的 webp 文件
var supportedFormats = map[string]bool{
	"jpg":  true,
	"jpeg": true,
	"png":  true,
	"gif":  true,
	"bmp":  true,
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
		// JPG 不支持透明通道，先把透明区域铺成白色，避免输出黑底/黑角
		if hasAlpha(img) {
			img = compositeOnWhite(img)
		}
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

// hasAlpha 判断图像是否含非完全不透明像素
func hasAlpha(img image.Image) bool {
	switch m := img.(type) {
	case *image.NRGBA:
		for y := m.Rect.Min.Y; y < m.Rect.Max.Y; y++ {
			row := m.Pix[(y-m.Rect.Min.Y)*m.Stride:]
			for x := 0; x < m.Rect.Dx(); x++ {
				if row[x*4+3] != 255 {
					return true
				}
			}
		}
	case *image.RGBA:
		for y := m.Rect.Min.Y; y < m.Rect.Max.Y; y++ {
			row := m.Pix[(y-m.Rect.Min.Y)*m.Stride:]
			for x := 0; x < m.Rect.Dx(); x++ {
				if row[x*4+3] != 255 {
					return true
				}
			}
		}
	default:
		return true
	}
	return false
}

// compositeOnWhite 把图像的透明区域铺成白色
func compositeOnWhite(src image.Image) image.Image {
	b := src.Bounds()
	dst := image.NewRGBA(b)
	draw.Draw(dst, b, image.NewUniform(color.White), image.Point{}, draw.Src)
	draw.Draw(dst, b, src, b.Min, draw.Over)
	return dst
}
