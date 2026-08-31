package app

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"image/gif"
	"image/jpeg"

	"github.com/boombuler/barcode"
	bcaztec "github.com/boombuler/barcode/aztec"
	bccodabar "github.com/boombuler/barcode/codabar"
	bc128 "github.com/boombuler/barcode/code128"
	bc39 "github.com/boombuler/barcode/code39"
	bc93 "github.com/boombuler/barcode/code93"
	bcdm "github.com/boombuler/barcode/datamatrix"
	bcean "github.com/boombuler/barcode/ean"
	bcqr "github.com/boombuler/barcode/qr"
	bcitf "github.com/boombuler/barcode/twooffive"
	"github.com/disintegration/imaging"
	"github.com/kbinani/screenshot"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/aztec"
	"github.com/makiuchi-d/gozxing/datamatrix"
	multiqr "github.com/makiuchi-d/gozxing/multi/qrcode"
	"github.com/makiuchi-d/gozxing/oned"
	"github.com/makiuchi-d/gozxing/qrcode"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.design/x/clipboard"
	"golang.org/x/image/bmp"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
	"golang.org/x/image/tiff"
	"golang.org/x/image/webp"
)

// ============================================================================
// 公共类型
// ============================================================================

// BarcodePoint 码在图中的定位点，前端可用来画高亮框
type BarcodePoint struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// BarcodeDecodeResult 单条识别结果
type BarcodeDecodeResult struct {
	Text   string         `json:"text"`
	Format string         `json:"format"`
	Type   string         `json:"type"` // qrcode / barcode / unknown
	Points []BarcodePoint `json:"points"`
}

// BarcodeGenerateRequest 生成参数
type BarcodeGenerateRequest struct {
	Content  string `json:"content"`
	Format   string `json:"format"`
	Size     int    `json:"size"`     // 二维码边长 / 一维码宽度（像素）
	Height   int    `json:"height"`   // 一维码高度（像素）
	Level    string `json:"level"`    // 二维码纠错等级 L/M/Q/H
	FGColor  string `json:"fgColor"`
	BGColor  string `json:"bgColor"`
	ShowText bool   `json:"showText"` // 一维码下方显示内容
	Margin   int    `json:"margin"`   // 二维码：静区模块数（0-8）；一维码：边距像素（0-64）
	Logo     string `json:"logo"`     // 二维码中心图标（data URL），仅 QR 支持
}

// BarcodeImageResult 生成/截图结果
type BarcodeImageResult struct {
	DataURL string `json:"dataURL"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
}

// BarcodeFormatOption 前端下拉用的格式描述
type BarcodeFormatOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
	Kind  string `json:"kind"` // qr(二维) / barcode(一维)
	Hint  string `json:"hint"`
}

// pngDataURLPrefix 标准 PNG data URL 前缀
const pngDataURLPrefix = "data:image/png;base64,"

// ============================================================================
// 识别（解码）
// ============================================================================

// twoDFormats 矩阵式二维码（gozxing 无 PDF417 reader，故不提供生成）
var twoDFormats = map[string]bool{
	"qr":         true,
	"datamatrix": true,
	"aztec":      true,
}

// ListBarcodeFormats 返回支持的码制
func (a *App) ListBarcodeFormats() []BarcodeFormatOption {
	return []BarcodeFormatOption{
		{Value: "qr", Label: "二维码 QR Code", Kind: "qr", Hint: "文本、网址、中文都支持，可加中心图标"},
		{Value: "datamatrix", Label: "Data Matrix", Kind: "qr", Hint: "零件追溯常用，小尺寸高密度，建议英文/数字"},
		{Value: "aztec", Label: "Aztec", Kind: "qr", Hint: "火车票 / 支付码常用，中心有定位靶，不支持图标"},
		{Value: "code128", Label: "Code 128", Kind: "barcode", Hint: "最通用的一维码，支持全部 ASCII 字符"},
		{Value: "code39", Label: "Code 39", Kind: "barcode", Hint: "数字、大写字母和 - . $ / + %，小写自动转大写"},
		{Value: "code93", Label: "Code 93", Kind: "barcode", Hint: "Code 39 的增强版，密度更高"},
		{Value: "ean13", Label: "EAN-13", Kind: "barcode", Hint: "13 位商品条码，12 位可自动补校验位"},
		{Value: "ean8", Label: "EAN-8", Kind: "barcode", Hint: "8 位商品条码，7 位可自动补校验位"},
		{Value: "upca", Label: "UPC-A", Kind: "barcode", Hint: "12 位北美商品条码，11 位可自动补校验位"},
		{Value: "itf", Label: "ITF 交叉二五", Kind: "barcode", Hint: "仅数字，且位数必须为偶数"},
		{Value: "codabar", Label: "Codabar", Kind: "barcode", Hint: "数字与 - $ : / . +，首尾码自动补 A"},
	}
}

// decodeImage 对一张图做多轮尝试，返回所有能认出来的码。
// 部分国产二维码用 GBK/GB2312 编码中文，UTF-8 强解会出乱码（U+FFFD），
// 检测到乱码时自动用 GB18030 重试。
func decodeImage(img image.Image) []BarcodeDecodeResult {
	results := decodeWithCharset(img, "UTF-8")
	if hasGarbledText(results) {
		if retry := decodeWithCharset(img, "GB18030"); len(retry) > 0 && !hasGarbledText(retry) {
			return filterRedundant(retry)
		}
	}
	return filterRedundant(results)
}

// decodeWithCharset 按指定字符集跑完整的识别流水线：
// 原图 → 增强变体（放大/对比度/锐化/缩小）→ 旋转 90/180/270 兜底
func decodeWithCharset(img image.Image, charset string) []BarcodeDecodeResult {
	seen := make(map[string]struct{})
	results := make([]BarcodeDecodeResult, 0, 4)

	collect := func(rs []BarcodeDecodeResult) bool {
		for _, r := range rs {
			key := r.Format + "\x00" + r.Text
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			results = append(results, r)
		}
		return len(results) > 0
	}

	// 第一轮：原图
	if collect(tryReaders(img, charset)) {
		return results
	}

	// 第二轮：图像增强变体
	for _, variant := range enhanceVariants(img) {
		if collect(tryReaders(variant, charset)) {
			return results
		}
	}

	// 第三轮：旋转兜底，处理条码倾斜或竖排的情况
	for _, rotated := range []image.Image{
		imaging.Rotate90(img),
		imaging.Rotate180(img),
		imaging.Rotate270(img),
	} {
		if collect(tryReaders(rotated, charset)) {
			return results
		}
	}

	return results
}

// hasGarbledText 结果里是否含替换字符（编码不对的典型表现）
func hasGarbledText(results []BarcodeDecodeResult) bool {
	for _, r := range results {
		if strings.ContainsRune(r.Text, '\uFFFD') {
			return true
		}
	}
	return false
}

// filterRedundant 过滤冗余结果：
// UPC-A 码本质是首位为 0 的 EAN-13，两种识别器会各报一条，保留 EAN-13 即可
func filterRedundant(results []BarcodeDecodeResult) []BarcodeDecodeResult {
	eanTexts := make(map[string]struct{}, len(results))
	for _, r := range results {
		if r.Format == "EAN_13" {
			eanTexts[r.Text] = struct{}{}
		}
	}
	out := results[:0]
	for _, r := range results {
		if r.Format == "UPC_A" && len(r.Text) == 12 {
			if _, ok := eanTexts["0"+r.Text]; ok {
				continue
			}
		}
		out = append(out, r)
	}
	return out
}

// enhanceVariants 生成一组增强后的候选图，逐个送进识别器
func enhanceVariants(img image.Image) []image.Image {
	b := img.Bounds()
	variants := make([]image.Image, 0, 5)

	// 小图放大：模块太少时采样不到，放大后能救回来不少
	if b.Dx() < 1000 {
		variants = append(variants, imaging.Resize(img, b.Dx()*2, 0, imaging.Lanczos))
	}
	if b.Dx() < 500 {
		variants = append(variants, imaging.Resize(img, b.Dx()*4, 0, imaging.Lanczos))
	}
	// 大图缩小：噪点和摩尔纹会干扰二值化
	if b.Dx() > 2000 {
		variants = append(variants, imaging.Resize(img, 1600, 0, imaging.Lanczos))
	}

	gray := imaging.Grayscale(img)
	variants = append(variants, imaging.AdjustContrast(gray, 40))
	variants = append(variants, imaging.Sharpen(gray, 1.2))

	return variants
}

// tryReaders 依次用不同识别器读同一张图，汇总所有结果
func tryReaders(img image.Image, charset string) []BarcodeDecodeResult {
	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return nil
	}

	hints := map[gozxing.DecodeHintType]interface{}{
		gozxing.DecodeHintType_TRY_HARDER:    true,
		gozxing.DecodeHintType_CHARACTER_SET: charset,
	}

	var out []BarcodeDecodeResult

	// 1) 一图多码：一次性把图里所有二维码捞出来
	if rs, err := multiqr.NewQRCodeMultiReader().DecodeMultiple(bmp, hints); err == nil {
		for _, r := range rs {
			if r != nil {
				out = append(out, convertBarcodeResult(r))
			}
		}
	}

	// 2) 单码二维码，多码判定失败时走这条路
	if r, err := qrcode.NewQRCodeReader().Decode(bmp, hints); err == nil && r != nil {
		out = append(out, convertBarcodeResult(r))
	}

	// 3) 其他矩阵式二维码（gozxing v0.1.1 没有 PDF417 reader，识别侧暂缺该码制）
	for _, rd := range []gozxing.Reader{
		datamatrix.NewDataMatrixReader(),
		aztec.NewAztecReader(),
	} {
		if r, err := rd.Decode(bmp, hints); err == nil && r != nil {
			out = append(out, convertBarcodeResult(r))
		}
	}

	// 4) 一维码逐格式精读（gozxing 没提供 1D 的聚合 reader，只能挨个试）
	for _, rd := range oneDReaders() {
		if r, err := rd.Decode(bmp, hints); err == nil && r != nil {
			out = append(out, convertBarcodeResult(r))
		}
	}

	return out
}

// oneDReaders 常用一维码识别器集合
func oneDReaders() []gozxing.Reader {
	return []gozxing.Reader{
		oned.NewCode128Reader(),
		oned.NewCode39Reader(),
		oned.NewCode93Reader(),
		oned.NewEAN13Reader(),
		oned.NewEAN8Reader(),
		oned.NewUPCAReader(),
		oned.NewUPCEReader(),
		oned.NewITFReader(),
		oned.NewCodaBarReader(),
	}
}

// convertBarcodeResult 把 gozxing 的结果转成前端结构
func convertBarcodeResult(r *gozxing.Result) BarcodeDecodeResult {
	format := r.GetBarcodeFormat().String()

	typ := "barcode"
	switch r.GetBarcodeFormat() {
	case gozxing.BarcodeFormat_QR_CODE,
		gozxing.BarcodeFormat_DATA_MATRIX,
		gozxing.BarcodeFormat_AZTEC,
		gozxing.BarcodeFormat_PDF_417:
		typ = "qrcode"
	}

	pts := make([]BarcodePoint, 0, 4)
	for _, p := range r.GetResultPoints() {
		pts = append(pts, BarcodePoint{X: int(p.GetX()), Y: int(p.GetY())})
	}

	return BarcodeDecodeResult{
		Text:   r.GetText(),
		Format: format,
		Type:   typ,
		Points: pts,
	}
}

// DecodeBarcodeImageData 识别图片数据（data URL 或裸 base64）
func (a *App) DecodeBarcodeImageData(dataURL string) ([]BarcodeDecodeResult, error) {
	raw, err := parseBarcodeDataURL(dataURL)
	if err != nil {
		return nil, fmt.Errorf("解析图片数据失败: %w", err)
	}
	img, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("无法解码该图片（仅支持 png/jpg/gif/bmp/tiff/webp）: %w", err)
	}

	results := decodeImage(img)
	if len(results) == 0 {
		return nil, fmt.Errorf("未在图中识别到二维码或条形码")
	}
	return results, nil
}

// DecodeBarcodeImageRegion 从图片上裁一块区域再识别（截图框选用）
func (a *App) DecodeBarcodeImageRegion(dataURL string, x, y, w, h int) ([]BarcodeDecodeResult, error) {
	raw, err := parseBarcodeDataURL(dataURL)
	if err != nil {
		return nil, fmt.Errorf("解析图片数据失败: %w", err)
	}
	img, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("无法解码该图片: %w", err)
	}

	// 选区往四周扩 8%，容差框选时没框全的码
	expandX, expandY := w*8/100, h*8/100
	rect := image.Rect(x-expandX, y-expandY, x+w+expandX, y+h+expandY).Intersect(img.Bounds())
	if rect.Empty() || rect.Dx() < 8 || rect.Dy() < 8 {
		return nil, fmt.Errorf("选区太小或无效，请重新框选")
	}

	cropped := imaging.Crop(img, rect)
	results := decodeImage(cropped)
	if len(results) == 0 {
		return nil, fmt.Errorf("选区内未识别到二维码或条形码")
	}
	return results, nil
}

// DecodeBarcodeImageFile 识别本地图片文件
func (a *App) DecodeBarcodeImageFile(path string) ([]BarcodeDecodeResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("无法解码该文件（仅支持 png/jpg/gif/bmp/tiff/webp）: %w", err)
	}

	results := decodeImage(img)
	if len(results) == 0 {
		return nil, fmt.Errorf("未在文件中识别到二维码或条形码")
	}
	return results, nil
}

// ============================================================================
// 生成（编码）
// ============================================================================

// GenerateBarcode 按请求生成码图，返回 PNG data URL
func (a *App) GenerateBarcode(req BarcodeGenerateRequest) (BarcodeImageResult, error) {
	content := strings.TrimSpace(req.Content)
	if content == "" {
		return BarcodeImageResult{}, fmt.Errorf("内容不能为空")
	}

	format := strings.ToLower(strings.TrimSpace(req.Format))
	if format == "" {
		format = "qr"
	}

	fg := color.RGBA{R: 0, G: 0, B: 0, A: 255}
	if req.FGColor != "" {
		c, err := parseBarcodeColor(req.FGColor)
		if err != nil {
			return BarcodeImageResult{}, err
		}
		fg = c
	}

	bg := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	if req.BGColor != "" {
		c, err := parseBarcodeColor(req.BGColor)
		if err != nil {
			return BarcodeImageResult{}, err
		}
		bg = c
	}
	if fg.R == bg.R && fg.G == bg.G && fg.B == bg.B {
		return BarcodeImageResult{}, fmt.Errorf("前景色和背景色不能相同，否则无法扫码")
	}

	var out image.Image
	var err error
	if twoDFormats[format] {
		out, err = build2DBarcode(content, format, req, fg, bg)
	} else {
		out, err = build1DBarcode(content, format, req, fg, bg)
	}
	if err != nil {
		return BarcodeImageResult{}, err
	}

	dataURL, err := barcodeImageToDataURL(out)
	if err != nil {
		return BarcodeImageResult{}, fmt.Errorf("编码 PNG 失败: %w", err)
	}

	b := out.Bounds()
	return BarcodeImageResult{DataURL: dataURL, Width: b.Dx(), Height: b.Dy()}, nil
}

// build2DBarcode 生成矩阵式二维码（QR / DataMatrix / Aztec / PDF417）
// 编码器输出的图每个模块 1px，这里统一放大到目标边长，静区按模块数计算
func build2DBarcode(content, format string, req BarcodeGenerateRequest, fg, bg color.RGBA) (image.Image, error) {
	var code barcode.Barcode
	var err error

	switch format {
	case "qr":
		code, err = bcqr.Encode(content, barcodeQRLevel(req.Level), bcqr.Auto)
	case "datamatrix":
		code, err = bcdm.Encode(content)
	case "aztec":
		code, err = bcaztec.Encode([]byte(content), 33, 0)
	default:
		return nil, fmt.Errorf("不支持的二维码格式: %s", format)
	}
	if err != nil {
		return nil, fmt.Errorf("%s 编码失败: %w", format, err)
	}

	size := req.Size
	if size < 64 {
		size = 320
	}
	if size > 4096 {
		size = 4096
	}

	mb := code.Bounds()
	modW, modH := mb.Dx(), mb.Dy()

	// 目标边长不够放下原始模块时保持 1:1，避免无意义的缩小导致编码失败
	scale := float64(size) / float64(max(modW, modH))
	if scale < 1 {
		scale = 1
	}
	w := max(modW, int(float64(modW)*scale+0.5))
	h := max(modH, int(float64(modH)*scale+0.5))
	code, err = barcode.Scale(code, w, h)
	if err != nil {
		return nil, fmt.Errorf("二维码缩放失败: %w", err)
	}

	// 静区以模块为单位（QR 规范推荐 4 个模块）
	mods := clampBarcodeMarginModules(req.Margin)
	quiet := int(float64(mods) * scale)
	canvasW := code.Bounds().Dx() + quiet*2
	canvasH := code.Bounds().Dy() + quiet*2

	canvas := barcodeNewCanvas(canvasW, canvasH, bg)
	cb := code.Bounds()
	draw.Draw(canvas, image.Rect(quiet, quiet, quiet+cb.Dx(), quiet+cb.Dy()),
		code, cb.Min, draw.Over)

	colored := recolorBarcodeImage(canvas, fg, bg)

	// 中心图标只支持 QR：Aztec 中心是定位靶，DataMatrix/PDF417 容错较低
	if format == "qr" && req.Logo != "" {
		if err := overlayBarcodeLogo(colored, req.Logo, bg); err != nil {
			return nil, fmt.Errorf("叠加图标失败: %w", err)
		}
	}

	return colored, nil
}

func build1DBarcode(content, format string, req BarcodeGenerateRequest, fg, bg color.RGBA) (image.Image, error) {
	// 数字类码制统一去掉空格，容错用户复制来的格式化数字
	// 注意 codabar 的 - 是合法字符，不能清洗
	numeric := map[string]bool{"ean13": true, "ean8": true, "upca": true, "itf": true}
	if numeric[format] {
		content = strings.NewReplacer(" ", "", "-", "").Replace(content)
	}

	var code barcode.Barcode
	var err error
	displayText := content

	switch format {
	case "code128":
		code, err = bc128.Encode(content)
	case "code39":
		// 基础 Code 39 只有数字和大写字母，小写自动转大写
		code, err = bc39.Encode(strings.ToUpper(content), false, false)
	case "code93":
		// addChecksum=true：必须带校验位，否则 gozxing 拒绝识别
		code, err = bc93.Encode(content, true, false)
	case "ean13":
		code, err = bcean.Encode(content)
		if err == nil {
			displayText = code.Content() // 含自动补出的校验位
		}
	case "ean8":
		code, err = bcean.Encode(content)
		if err == nil {
			displayText = code.Content()
		}
	case "upca":
		code, err = encodeBarcodeUPCA(content)
		if err == nil {
			displayText = code.Content() // 13 位 EAN-13 形式，含前导 0 和校验位
		}
	case "itf":
		// 交叉二五要求数字位数为偶数，奇数位补前导 0（业内惯例）
		if len(content)%2 == 1 {
			content = "0" + content
		}
		code, err = bcitf.Encode(content, true)
		if err == nil {
			displayText = content
		}
	case "codabar":
		// 库要求首尾必须是 A-D 起止符，没有就自动补上
		if !isCodabarFramed(content) {
			content = "A" + strings.ToUpper(content) + "A"
		}
		code, err = bccodabar.Encode(strings.ToUpper(content))
		if err == nil {
			displayText = content
		}
	default:
		return nil, fmt.Errorf("不支持的条码格式: %s", format)
	}
	if err != nil {
		return nil, fmt.Errorf("%s 编码失败: %w", format, err)
	}

	mb := code.Bounds()
	width := req.Size
	if width < 60 {
		width = 300
	}
	if width > 4096 {
		width = 4096
	}
	// 内容太长时一维码模块数会超过请求宽度，只能按原始宽度出图
	if width < mb.Dx() {
		width = mb.Dx()
	}
	height := req.Height
	if height < 20 {
		height = 100
	}
	if height > 2048 {
		height = 2048
	}
	if height < mb.Dy() {
		height = mb.Dy()
	}

	code, err = barcode.Scale(code, width, height)
	if err != nil {
		return nil, fmt.Errorf("条码缩放失败: %w", err)
	}

	quiet := clampBarcodeMarginPx(req.Margin)
	textH := 0
	if req.ShowText {
		textH = barcodeTextAreaHeight(height)
	}

	canvas := barcodeNewCanvas(code.Bounds().Dx()+quiet*2, code.Bounds().Dy()+quiet*2+textH, bg)
	cb := code.Bounds()
	draw.Draw(canvas, image.Rect(quiet, quiet, quiet+cb.Dx(), quiet+cb.Dy()),
		code, cb.Min, draw.Over)

	colored := recolorBarcodeImage(canvas, fg, bg)

	if req.ShowText {
		drawBarcodeCenteredText(colored, displayText, fg, textH)
	}

	return colored, nil
}

// encodeBarcodeUPCA 把 UPC-A 内容归一化后按 EAN-13 编码（UPC-A 即首位为 0 的 EAN-13）。
// 11 位 = 数据位，自动算校验位；12 位 = 已含校验位，校验后整体前补 0。
// 之前直接调 ean.Encode 会把 12 位 UPC-A 当成"缺校验位的 EAN-13"再补一位，编码出错误内容。
func encodeBarcodeUPCA(content string) (barcode.Barcode, error) {
	if len(content) == 0 {
		return nil, fmt.Errorf("内容不能为空")
	}
	for _, r := range content {
		if r < '0' || r > '9' {
			return nil, fmt.Errorf("UPC-A 只能是数字，收到: %s", content)
		}
	}

	switch len(content) {
	case 11:
		content += string(barcodeUPCCheckDigit(content))
	case 12:
		if barcodeUPCCheckDigit(content[:11]) != rune(content[11]) {
			return nil, fmt.Errorf("校验位不正确：应为 %c%s", barcodeUPCCheckDigit(content[:11]), content)
		}
	default:
		return nil, fmt.Errorf("UPC-A 需要 11 位数据或 12 位含校验位的数字，收到 %d 位", len(content))
	}

	return bcean.Encode("0" + content)
}

// barcodeUPCCheckDigit 计算 UPC-A / EAN 的校验位（从左起奇数位×3）
func barcodeUPCCheckDigit(digits string) rune {
	sum := 0
	for i, r := range digits {
		v := int(r - '0')
		if i%2 == 0 {
			sum += v * 3
		} else {
			sum += v
		}
	}
	return rune((10 - sum%10) % 10 + '0')
}

func isCodabarFramed(s string) bool {
	if len(s) < 2 {
		return false
	}
	first, last := barcodeUpperByte(s[0]), barcodeUpperByte(s[len(s)-1])
	return (first == 'A' || first == 'B' || first == 'C' || first == 'D') &&
		(last == 'A' || last == 'B' || last == 'C' || last == 'D')
}

func barcodeUpperByte(b byte) byte {
	if b >= 'a' && b <= 'z' {
		return b - 32
	}
	return b
}

func barcodeQRLevel(level string) bcqr.ErrorCorrectionLevel {
	switch strings.ToUpper(strings.TrimSpace(level)) {
	case "L":
		return bcqr.L
	case "Q":
		return bcqr.Q
	case "H":
		return bcqr.H
	default:
		return bcqr.M
	}
}

// ============================================================================
// 截图
// ============================================================================

// CaptureBarcodeScreen 截屏并返回 PNG data URL
// mode: "primary" 只截主屏，"all" 把所有显示器拼成一张
// 截图前会临时隐藏本窗口，防止把工具自己拍进去
func (a *App) CaptureBarcodeScreen(mode string) (BarcodeImageResult, error) {
	if a.ctx != nil {
		wailsruntime.WindowHide(a.ctx)
		defer wailsruntime.WindowShow(a.ctx)
		// 留足窗口真正消失的时间，否则会截到自己的残影
		time.Sleep(300 * time.Millisecond)
	}

	var img *image.RGBA
	var err error
	if mode == "all" {
		img, err = captureAllBarcodeDisplays()
	} else {
		img, err = screenshot.CaptureDisplay(0)
	}
	if err != nil {
		return BarcodeImageResult{}, fmt.Errorf("截图失败: %w", err)
	}
	if img == nil {
		return BarcodeImageResult{}, fmt.Errorf("截图结果为空")
	}

	dataURL, err := barcodeImageToDataURL(img)
	if err != nil {
		return BarcodeImageResult{}, fmt.Errorf("截图编码失败: %w", err)
	}

	b := img.Bounds()
	return BarcodeImageResult{DataURL: dataURL, Width: b.Dx(), Height: b.Dy()}, nil
}

// DecodeBarcodeScreenshot 截屏后直接整屏识别，适合屏幕上只有一个码的场景
func (a *App) DecodeBarcodeScreenshot(mode string) ([]BarcodeDecodeResult, error) {
	shot, err := a.CaptureBarcodeScreen(mode)
	if err != nil {
		return nil, err
	}
	return a.DecodeBarcodeImageData(shot.DataURL)
}

// captureAllBarcodeDisplays 按各显示器的相对位置拼接成一张大图
func captureAllBarcodeDisplays() (*image.RGBA, error) {
	n := screenshot.NumActiveDisplays()
	if n <= 0 {
		return nil, fmt.Errorf("未检测到可用显示器")
	}

	rects := make([]image.Rectangle, 0, n)
	minX, minY, maxX, maxY := 0, 0, 0, 0

	for i := 0; i < n; i++ {
		b := screenshot.GetDisplayBounds(i)
		rects = append(rects, b)
		if i == 0 {
			minX, minY, maxX, maxY = b.Min.X, b.Min.Y, b.Max.X, b.Max.Y
			continue
		}
		if b.Min.X < minX {
			minX = b.Min.X
		}
		if b.Min.Y < minY {
			minY = b.Min.Y
		}
		if b.Max.X > maxX {
			maxX = b.Max.X
		}
		if b.Max.Y > maxY {
			maxY = b.Max.Y
		}
	}

	canvas := image.NewRGBA(image.Rect(0, 0, maxX-minX, maxY-minY))
	for _, b := range rects {
		img, err := screenshot.CaptureRect(b)
		if err != nil {
			continue
		}
		target := image.Rect(b.Min.X-minX, b.Min.Y-minY, b.Max.X-minX, b.Max.Y-minY)
		draw.Draw(canvas, target, img, img.Bounds().Min, draw.Src)
	}

	return canvas, nil
}

// ============================================================================
// 文件 / 剪贴板 / 系统集成
// ============================================================================

// SelectBarcodeImageFile 弹出文件选择框（带图片过滤），返回图片路径；取消时返回空串
func (a *App) SelectBarcodeImageFile() (string, error) {
	path, err := wailsruntime.OpenFileDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title: "选择要识别的图片",
		Filters: []wailsruntime.FileFilter{
			{DisplayName: "图片文件", Pattern: "*.png;*.jpg;*.jpeg;*.gif;*.bmp;*.tiff;*.tif;*.webp"},
			{DisplayName: "所有文件", Pattern: "*.*"},
		},
	})
	if err != nil {
		return "", err
	}
	return path, nil
}

// SelectBarcodeLogoFile 选择二维码中心图标，直接返回统一后的 PNG data URL；取消时返回空串
func (a *App) SelectBarcodeLogoFile() (string, error) {
	path, err := wailsruntime.OpenFileDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title: "选择中心图标",
		Filters: []wailsruntime.FileFilter{
			{DisplayName: "图片文件", Pattern: "*.png;*.jpg;*.jpeg;*.gif;*.bmp"},
		},
	})
	if err != nil {
		return "", err
	}
	if path == "" {
		return "", nil
	}
	return a.ReadBarcodeImageFile(path)
}

// ReadBarcodeImageFile 读取本地图片为 PNG data URL，供前端预览
func (a *App) ReadBarcodeImageFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("读取文件失败: %w", err)
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("不支持的图片格式: %w", err)
	}
	return barcodeImageToDataURL(img)
}

// SaveBarcodeImage 弹出保存对话框并写入 PNG；取消时返回空串
func (a *App) SaveBarcodeImage(dataURL, defaultName string) (string, error) {
	raw, err := parseBarcodeDataURL(dataURL)
	if err != nil {
		return "", fmt.Errorf("解析图片失败: %w", err)
	}
	if strings.TrimSpace(defaultName) == "" {
		defaultName = "code.png"
	}
	if !strings.HasSuffix(strings.ToLower(defaultName), ".png") {
		defaultName += ".png"
	}

	path, err := wailsruntime.SaveFileDialog(a.ctx, wailsruntime.SaveDialogOptions{
		Title:           "保存图片",
		DefaultFilename: defaultName,
		Filters: []wailsruntime.FileFilter{
			{DisplayName: "PNG 图片", Pattern: "*.png"},
		},
	})
	if err != nil {
		return "", err
	}
	if path == "" {
		return "", nil
	}
	if !strings.HasSuffix(strings.ToLower(path), ".png") {
		path += ".png"
	}

	// 统一转成 PNG 再落盘，避免把原始 jpg 字节写成 .png
	pngBytes, err := toBarcodePNGBytes(raw)
	if err != nil {
		return "", fmt.Errorf("转换 PNG 失败: %w", err)
	}
	if err := os.WriteFile(path, pngBytes, 0644); err != nil {
		return "", fmt.Errorf("保存失败: %w", err)
	}
	return path, nil
}

// QuickSaveBarcodeToDesktop 一键保存到桌面（含 OneDrive 桌面重定向的机器）
func (a *App) QuickSaveBarcodeToDesktop(dataURL, fileName string) (string, error) {
	raw, err := parseBarcodeDataURL(dataURL)
	if err != nil {
		return "", fmt.Errorf("解析图片失败: %w", err)
	}
	if strings.TrimSpace(fileName) == "" {
		fileName = "code.png"
	}
	if !strings.HasSuffix(strings.ToLower(fileName), ".png") {
		fileName += ".png"
	}

	desktop, err := barcodeDesktopDir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(desktop, fileName)
	pngBytes, err := toBarcodePNGBytes(raw)
	if err != nil {
		return "", fmt.Errorf("转换 PNG 失败: %w", err)
	}
	if err := os.WriteFile(path, pngBytes, 0644); err != nil {
		return "", fmt.Errorf("保存失败: %w", err)
	}
	return path, nil
}

// barcodeDesktopDir 尝试定位真实的桌面目录：
// 部分 Windows 机器把桌面重定向到了 OneDrive，优先探测
func barcodeDesktopDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("无法定位用户目录: %w", err)
	}
	candidates := []string{
		filepath.Join(home, "Desktop"),
		filepath.Join(home, "OneDrive", "Desktop"),
		filepath.Join(home, "OneDrive", "桌面"),
	}
	for _, dir := range candidates {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir, nil
		}
	}
	// 都不存在就建一个标准桌面目录
	if err := os.MkdirAll(candidates[0], 0755); err != nil {
		return "", fmt.Errorf("无法创建桌面目录: %w", err)
	}
	return candidates[0], nil
}

// ShowBarcodeInFolder 在资源管理器中定位到文件
func (a *App) ShowBarcodeInFolder(path string) error {
	if path == "" {
		return fmt.Errorf("路径为空")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	return exec.Command("explorer", "/select,", abs).Start()
}

// OpenBarcodeURL 用系统默认浏览器打开链接
func (a *App) OpenBarcodeURL(link string) error {
	link = strings.TrimSpace(link)
	u, err := url.Parse(link)
	if err != nil || u.Scheme == "" {
		return fmt.Errorf("不是有效的链接: %s", link)
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("仅支持打开 http/https 链接: %s", link)
	}
	wailsruntime.BrowserOpenURL(a.ctx, link)
	return nil
}

// ============================================================================
// 剪贴板（带图片通道，golang.design/x/clipboard 与 atotto/clipboard 并存）
// ============================================================================

var (
	barcodeClipOnce sync.Once
	barcodeClipErr  error
)

func ensureBarcodeClipboard() error {
	barcodeClipOnce.Do(func() {
		barcodeClipErr = clipboard.Init()
	})
	return barcodeClipErr
}

func barcodeClipboardCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 3*time.Second)
}

// ReadBarcodeClipboardText 读取剪贴板里的文本
func (a *App) ReadBarcodeClipboardText() (string, error) {
	if err := ensureBarcodeClipboard(); err != nil {
		return "", fmt.Errorf("剪贴板初始化失败: %w", err)
	}
	ctx, cancel := barcodeClipboardCtx()
	defer cancel()
	data, err := clipboard.Read(ctx, clipboard.FmtText)
	if err != nil {
		return "", fmt.Errorf("读取剪贴板失败: %w", err)
	}
	return string(data), nil
}

// WriteBarcodeClipboardText 写入文本到剪贴板
func (a *App) WriteBarcodeClipboardText(text string) error {
	if err := ensureBarcodeClipboard(); err != nil {
		return fmt.Errorf("剪贴板初始化失败: %w", err)
	}
	ctx, cancel := barcodeClipboardCtx()
	defer cancel()
	if _, err := clipboard.Write(ctx, clipboard.FmtText, []byte(text)); err != nil {
		return fmt.Errorf("写入剪贴板失败: %w", err)
	}
	return nil
}

// ReadBarcodeClipboardImage 读取剪贴板里的图片，返回 PNG data URL
func (a *App) ReadBarcodeClipboardImage() (string, error) {
	if err := ensureBarcodeClipboard(); err != nil {
		return "", fmt.Errorf("剪贴板初始化失败: %w", err)
	}
	ctx, cancel := barcodeClipboardCtx()
	defer cancel()
	data, err := clipboard.Read(ctx, clipboard.FmtImage)
	if err != nil {
		return "", fmt.Errorf("读取剪贴板失败: %w", err)
	}
	if len(data) == 0 {
		return "", fmt.Errorf("剪贴板里没有图片，先截图或复制一张图片再来")
	}
	return pngDataURLPrefix + base64.StdEncoding.EncodeToString(data), nil
}

// WriteBarcodeClipboardImage 把图片写入剪贴板（非 PNG 会自动转码）
func (a *App) WriteBarcodeClipboardImage(dataURL string) error {
	if err := ensureBarcodeClipboard(); err != nil {
		return fmt.Errorf("剪贴板初始化失败: %w", err)
	}
	raw, err := parseBarcodeDataURL(dataURL)
	if err != nil {
		return fmt.Errorf("解析图片失败: %w", err)
	}
	pngBytes, err := toBarcodePNGBytes(raw)
	if err != nil {
		return fmt.Errorf("转换 PNG 失败: %w", err)
	}
	ctx, cancel := barcodeClipboardCtx()
	defer cancel()
	if _, err := clipboard.Write(ctx, clipboard.FmtImage, pngBytes); err != nil {
		return fmt.Errorf("写入剪贴板失败: %w", err)
	}
	return nil
}

// DecodeBarcodeClipboard 直接识别剪贴板里的图片
func (a *App) DecodeBarcodeClipboard() ([]BarcodeDecodeResult, error) {
	dataURL, err := a.ReadBarcodeClipboardImage()
	if err != nil {
		return nil, err
	}
	return a.DecodeBarcodeImageData(dataURL)
}

// ============================================================================
// 内部工具函数
// ============================================================================

// barcodeNewCanvas 铺好背景色的空白画布
func barcodeNewCanvas(w, h int, bg color.RGBA) *image.RGBA {
	canvas := image.NewRGBA(image.Rect(0, 0, w, h))
	if bg.A > 0 {
		draw.Draw(canvas, canvas.Bounds(), image.NewUniform(bg), image.Point{}, draw.Src)
	}
	return canvas
}

// clampBarcodeMarginModules 二维码静区模块数，限制 0-8
func clampBarcodeMarginModules(m int) int {
	if m < 0 {
		return 0
	}
	if m > 8 {
		return 8
	}
	return m
}

// clampBarcodeMarginPx 一维码边距像素，限制 0-64
func clampBarcodeMarginPx(m int) int {
	if m < 0 {
		return 0
	}
	if m > 64 {
		return 64
	}
	return m
}

// barcodeTextAreaHeight 根据条码高度给出合适的文字区高度
func barcodeTextAreaHeight(barHeight int) int {
	size := barHeight / 8
	if size < 12 {
		size = 12
	}
	if size > 24 {
		size = 24
	}
	return size + 10
}

// overlayBarcodeLogo 在二维码正中叠一个图标，带背景色底衬
func overlayBarcodeLogo(dst *image.RGBA, logoData string, bg color.RGBA) error {
	raw, err := parseBarcodeDataURL(logoData)
	if err != nil {
		return err
	}
	logo, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return err
	}

	b := dst.Bounds()
	maxLogo := b.Dx() / 5
	if maxLogo < 24 {
		maxLogo = 24
	}
	// Fit 保持宽高比，之前 Resize 会把非方形图标拉变形
	logo = imaging.Fit(logo, maxLogo, maxLogo, imaging.Lanczos)
	lb := logo.Bounds()

	offset := image.Point{
		X: (b.Dx() - lb.Dx()) / 2,
		Y: (b.Dy() - lb.Dy()) / 2,
	}
	rect := image.Rect(offset.X, offset.Y, offset.X+lb.Dx(), offset.Y+lb.Dy())

	// 底衬用不透明的背景色，避免 Logo 压在码点上影响识别
	opaque := color.RGBA{R: bg.R, G: bg.G, B: bg.B, A: 255}
	draw.Draw(dst, rect.Inset(-4), image.NewUniform(opaque), image.Point{}, draw.Over)
	draw.Draw(dst, rect, logo, lb.Min, draw.Over)
	return nil
}

// drawBarcodeCenteredText 在图像底部居中绘制内容文字（一维码的人眼可读部分）。
// 使用系统字体以支持中文，字号随条码高度自适应，超宽时自动缩小。
func drawBarcodeCenteredText(dst *image.RGBA, text string, col color.RGBA, areaH int) {
	if areaH <= 0 || text == "" {
		return
	}
	b := dst.Bounds()
	maxW := b.Dx() - 8
	if maxW <= 0 {
		return
	}

	fontSize := areaH - 10
	if fontSize < 9 {
		fontSize = 9
	}
	var face font.Face
	var d *font.Drawer
	for {
		face = newBarcodeTextFace(float64(fontSize))
		d = &font.Drawer{Face: face, Src: image.NewUniform(col)}
		if d.MeasureString(text).Ceil() <= maxW || fontSize <= 9 {
			break
		}
		face.Close()
		fontSize -= 2
	}
	defer face.Close()

	d.Dst = dst

	m := face.Metrics()
	asc := m.Ascent.Ceil()
	desc := m.Descent.Ceil()
	gap := (areaH - asc - desc) / 2
	if gap < 0 {
		gap = 0
	}

	width := d.MeasureString(text).Ceil()
	x := (b.Dx() - width) / 2
	if x < 4 {
		x = 4
	}
	d.Dot = fixed.P(x, b.Dy()-areaH+gap+asc)
	d.DrawString(text)
}

// parseBarcodeDataURL 从 data URL（或裸 base64）中取出原始字节
func parseBarcodeDataURL(s string) ([]byte, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, fmt.Errorf("图片数据为空")
	}
	if strings.HasPrefix(s, "data:") {
		if i := strings.Index(s, ","); i >= 0 {
			s = s[i+1:]
		}
	}
	return base64.StdEncoding.DecodeString(s)
}

// barcodeImageToDataURL 把图像编码为 PNG 格式的 data URL
func barcodeImageToDataURL(img image.Image) (string, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", err
	}
	return pngDataURLPrefix + base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// parseBarcodeColor 支持 #RGB / #RRGGBB / #RRGGBBAA，也支持 "transparent"
func parseBarcodeColor(s string) (color.RGBA, error) {
	s = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(s), "#"))
	if strings.EqualFold(s, "transparent") {
		return color.RGBA{R: 0, G: 0, B: 0, A: 0}, nil
	}
	var r, g, b, a uint8
	a = 255
	var err error
	switch len(s) {
	case 3:
		_, err = fmt.Sscanf(s, "%1x%1x%1x", &r, &g, &b)
		r, g, b = r*17, g*17, b*17
	case 6:
		_, err = fmt.Sscanf(s, "%02x%02x%02x", &r, &g, &b)
	case 8:
		_, err = fmt.Sscanf(s, "%02x%02x%02x%02x", &r, &g, &b, &a)
	default:
		return color.RGBA{}, fmt.Errorf("不支持的颜色格式: %s", s)
	}
	if err != nil {
		return color.RGBA{}, err
	}
	return color.RGBA{R: r, G: g, B: b, A: a}, nil
}

// recolorBarcodeImage 把黑白码图按亮度重新着色：暗部→前景色，亮部→背景色
func recolorBarcodeImage(src image.Image, fg, bg color.RGBA) *image.RGBA {
	b := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			r, g, bb, _ := src.At(b.Min.X+x, b.Min.Y+y).RGBA()
			// RGBA() 返回 0..65535，加权后满量程为 1000*65535，取 50% 作为黑白分界
			lum := 299*uint32(r) + 587*uint32(g) + 114*uint32(bb)
			if lum < 1000*65535/2 {
				dst.SetRGBA(x, y, fg)
			} else {
				dst.SetRGBA(x, y, bg)
			}
		}
	}
	return dst
}

// toBarcodePNGBytes 把任意格式的图片字节转成 PNG 字节
func toBarcodePNGBytes(raw []byte) ([]byte, error) {
	img, format, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	if format == "png" {
		return raw, nil
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ============================================================================
// 中文字体加载（用于一维码底部文字）
// ============================================================================

var (
	barcodeFontOnce   sync.Once
	barcodeCachedFont *opentype.Font
)

// loadBarcodeSystemFont 加载一个支持中文的系统字体，找不到时返回 nil（回退 basicfont）
func loadBarcodeSystemFont() *opentype.Font {
	barcodeFontOnce.Do(func() {
		windir := os.Getenv("WINDIR")
		if windir == "" {
			windir = `C:\Windows`
		}
		candidates := []string{
			filepath.Join(windir, "Fonts", "msyh.ttc"),   // 微软雅黑
			filepath.Join(windir, "Fonts", "simhei.ttf"), // 黑体
			filepath.Join(windir, "Fonts", "simsun.ttc"), // 宋体
			filepath.Join(windir, "Fonts", "arial.ttf"),  // ASCII 兜底
		}
		for _, p := range candidates {
			data, err := os.ReadFile(p)
			if err != nil {
				continue
			}
			if strings.HasSuffix(strings.ToLower(p), ".ttc") {
				col, err := opentype.ParseCollection(data)
				if err != nil {
					continue
				}
				f, err := col.Font(0)
				if err != nil {
					continue
				}
				barcodeCachedFont = f
				return
			}
			f, err := opentype.Parse(data)
			if err != nil {
				continue
			}
			barcodeCachedFont = f
			return
		}
	})
	return barcodeCachedFont
}

// newBarcodeTextFace 按字号创建可绘制中文的 Face，失败时回退到内置 7x13 点阵字体
func newBarcodeTextFace(sizePx float64) font.Face {
	if f := loadBarcodeSystemFont(); f != nil {
		face, err := opentype.NewFace(f, &opentype.FaceOptions{
			Size: sizePx,
			DPI:  72,
		})
		if err == nil {
			return face
		}
	}
	return basicfont.Face7x13
}

// 引用标准库以触发对应解码器注册
var (
	_ = jpeg.Decode
	_ = gif.Decode
	_ = bmp.Decode
	_ = tiff.Decode
	_ = webp.Decode
)
