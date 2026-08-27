package myImage

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	ico "github.com/Kodeworks/golang-image-ico"
	"github.com/disintegration/letteravatar"
	"github.com/fogleman/gg"
	"github.com/go-text/render"
	gotextfont "github.com/go-text/typesetting/font"
	"github.com/golang/freetype"
	"github.com/wyzzgzhdcxy/wcj-go-common/core"
	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

func sha256Hash(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))[:16]
}

func TextToPng(pngPath string, content string, size int) {
	firstLetter, _ := utf8.DecodeRuneInString(content)
	fontPath := fmt.Sprintf("%s/font_lib/方正粗黑宋简体.ttf", core.GetTempDir())
	fontFile, _ := os.ReadFile(fontPath)
	font, _ := freetype.ParseFont(fontFile)
	options := &letteravatar.Options{
		Font: font,
	}
	img, err := letteravatar.Draw(size, firstLetter, options)
	if err != nil {
		log.Printf("%v", err)
	}
	file, err := os.Create(pngPath)
	if err != nil {
		log.Printf("%v", err)
	}
	err = png.Encode(file, img)
	if err != nil {
		log.Printf("%v", err)
	}
}

func GetImageResolution(path string) (int, int) {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return 0, 0
	}
	defer core.Close(file)
	img, err := jpeg.Decode(file)
	if err != nil {
		fmt.Println("Error decoding JPEG:", err)
		return 0, 0
	}
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	return width, height
}

type FileRes struct {
	MimeType string `json:"mimeType"`
	Encoded  string `json:"encoded"`
}

func ReadFileWithBase64Encode(path string) FileRes {
	fileData, _ := os.ReadFile(path)
	mimeType := http.DetectContentType(fileData)
	encoded := base64.StdEncoding.EncodeToString(fileData)
	return FileRes{
		MimeType: mimeType,
		Encoded:  encoded,
	}
}

type ImageRes struct {
	PngMimeType string `json:"pngMimeType"`
	PngEncoded  string `json:"pngEncoded"`
	PngUrl      string `json:"pngUrl"`
	JpgMimeType string `json:"jpgMimeType"`
	JpgEncoded  string `json:"jpgEncoded"`
	JpgUrl      string `json:"jpgUrl"`
	IcoMimeType string `json:"icoMimeType"`
	IcoEncoded  string `json:"icoEncoded"`
	IcoUrl      string `json:"icoUrl"`
}

func LoadFont(fontPath string, size float64) (font.Face, error) {
	fontFile, _ := os.ReadFile(fontPath)
	fontParsed, err := opentype.Parse(fontFile)
	if err != nil {
		return nil, err
	}
	face, err := opentype.NewFace(fontParsed, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	return face, err
}

func DrawStringAnchored(text string, imageSize int, cornerRadius float64) {
	hash := fmt.Sprintf("%d_%x", imageSize, sha256Hash(text))
	DrawStringAnchoredWithFilename(text, imageSize, cornerRadius, hash)
}

func DrawStringAnchoredWithFilename(text string, imageSize int, cornerRadius float64, filename string) {
	fmt.Printf("=== DrawStringAnchoredWithFilename: text=%s, size=%d ===\n", text, imageSize)

	width, height := imageSize, imageSize

	emojiFontPath := fmt.Sprintf("%s/font_lib/EmojiOneColor.otf", core.GetTempDir())
	if _, err := os.Stat(emojiFontPath); err != nil {
		emojiFontPath = "C:/Windows/Fonts/seguiemj.ttf"
	}
	fmt.Printf("使用字体: %s\n", emojiFontPath)

	fontFile, err := os.ReadFile(emojiFontPath)
	if err != nil {
		fmt.Printf("字体文件读取失败: %v\n", err)
		return
	}

	fnt, err := gotextfont.ParseTTF(bytes.NewReader(fontFile))
	if err != nil {
		fmt.Printf("字体解析失败: %v\n", err)
		return
	}
	fmt.Println("字体解析成功")

	img := image.NewNRGBA(image.Rect(0, 0, width, height))

	r := &render.Renderer{
		FontSize: float32(height - 40),
		PixScale: 1.0,
		Color:    color.White,
	}

	x := (width - int(r.FontSize)) / 2
	y := height/2 + int(r.FontSize)/3
	endX := r.DrawStringAt(text, img, x, y, fnt)
	fmt.Printf("文字绘制完成, endX=%d, img已绑定\n", endX)

	dc := gg.NewContext(width, height)
	bgColor := color.RGBA{R: 70, G: 130, B: 180, A: 255}
	dc.SetColor(bgColor)
	dc.DrawRoundedRectangle(0, 0, float64(width), float64(height), cornerRadius)
	dc.Fill()

	dc.DrawImage(img, 0, 0)

	save2PngWithFilename(dc, filename)
	save2IcoWithFilename(dc, filename)
}

func downloadAndCompositeEmoji(text string, dc *gg.Context, width, height int) {
	emojiURL := getEmojiImageURL(text)
	fmt.Printf("emoji URL: %s\n", emojiURL)
	if emojiURL == "" {
		fmt.Println("无法获取emoji图片URL")
		return
	}

	img, err := downloadImage(emojiURL)
	if err != nil {
		fmt.Println("下载emoji图片失败:", err)
		return
	}

	emojiSize := width * 3 / 4
	resized := resizeImage(img, emojiSize, emojiSize)
	x := (width - emojiSize) / 2
	y := (height - emojiSize) / 2

	dcImg := dc.Image().(*image.RGBA)
	draw.Draw(dcImg, image.Rect(x, y, x+emojiSize, y+emojiSize), resized, resized.Bounds().Min, draw.Over)
}

func getEmojiImageURL(emoji string) string {
	var codes []string
	for _, r := range emoji {
		codes = append(codes, fmt.Sprintf("%x", r))
	}
	codeStr := strings.Join(codes, "-")
	return fmt.Sprintf("https://cdn.jsdelivr.net/gh/twitter/twemoji@14.0.2/assets/256x256/%s.png", codeStr)
}

func downloadImage(imgURL string) (image.Image, error) {
	resp, err := http.Get(imgURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	img, _, err := image.Decode(resp.Body)
	return img, err
}

func resizeImage(img image.Image, width, height int) image.Image {
	newImg := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.CatmullRom.Scale(newImg, newImg.Rect, img, img.Bounds(), draw.Src, nil)
	return newImg
}

func save2PngWithFilename(dc *gg.Context, filename string) {
	err := dc.SavePNG(core.GetTempDir() + "/png/" + filename + ".png")
	if err != nil {
		fmt.Println("保存文件出错:", err)
		return
	}
	fmt.Println("图标已生成:", filename+".png")
}

func save2IcoWithFilename(dc *gg.Context, filename string) {
	outputFile, err := os.Create(core.GetTempDir() + "/png/" + filename + ".ico")
	if err != nil {
		fmt.Println("创建文件失败:", err)
		return
	}
	defer core.Close(outputFile)

	err = ico.Encode(outputFile, dc.Image())
	if err != nil {
		fmt.Println("转换ICO失败:", err)
		return
	}
	fmt.Println("ICO图标已生成:", filename+".ico")
}

func SaveEmojiToCache(emoji string) (string, string) {
	return SaveEmojiToCacheWithDir(emoji, filepath.Join(core.GetTempDir(), "png"))
}

func SaveEmojiToCacheWithDir(emoji string, saveDir string) (string, string) {
	emojiURL := fmt.Sprintf("https://emoji-route.deno.dev/png/%s", url.PathEscape(emoji))
	fmt.Printf("emoji URL: %s\n", emojiURL)

	resp, err := http.Get(emojiURL)
	if err != nil {
		fmt.Println("下载emoji图片失败:", err)
		return "", ""
	}
	defer resp.Body.Close()

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		fmt.Println("解码emoji图片失败:", err)
		return "", ""
	}

	hash := sha256Hash(emoji)
	os.MkdirAll(saveDir, 0755)
	pngPath := filepath.Join(saveDir, hash+".png")
	icoPath := filepath.Join(saveDir, hash+".ico")

	pngFile, err := os.Create(pngPath)
	if err != nil {
		fmt.Println("创建PNG文件失败:", err)
		return "", ""
	}
	err = png.Encode(pngFile, img)
	if err != nil {
		fmt.Println("保存PNG失败:", err)
		return "", ""
	}
	core.Close(pngFile)

	icoFile, err := os.Create(icoPath)
	if err != nil {
		fmt.Println("创建ICO文件失败:", err)
		return "", ""
	}
	err = ico.Encode(icoFile, img)
	if err != nil {
		fmt.Println("转换ICO失败:", err)
		return "", ""
	}
	core.Close(icoFile)

	return pngPath, icoPath
}
