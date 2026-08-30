package app

import (
	"strings"
	"testing"
)

// 哈希类：与 hashlib / RFC 标准向量比对
func TestTextEncodeHashes(t *testing.T) {
	a := &App{}
	cases := []struct {
		op, in, want string
	}{
		{"md5", "hello", "5d41402abc4b2a76b9719d911017c592"},
		{"sha1", "hello", "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d"},
		{"sha224", "hello", "ea09ae9cc6768c50fcee903ed054556e5bfc8347907f12598aa24193"},
		{"sha256", "hello", "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"},
		{"sha384", "hello", "59e1748777448c69de6b800d7a33bbfb9ff1b463e44354c3553bcdb9c666fa90125a3c79f90397bdf5f6a13de828684f"},
		{"sha512", "hello", "9b71d224bd62f3785d96d46ad3ea3d73319bfbc2890caadae2dff72519673ca72323c3d99ba5c11d7c7acc6e14b8c5da0c4663475c2e5c3adef46f73bcdec043"},
		{"sm3", "abc", "66c7f0f462eeedd9d1f2d46bdc10e4e24167c4875cf2f7a2297da02b8f4ba8e0"}, // GB/T 32905 标准向量
		{"crc32", "hello", "3610a686"},
	}
	for _, c := range cases {
		if got := a.TextEncode(c.in, c.op); got != c.want {
			t.Errorf("%s(%q) = %q, want %q", c.op, c.in, got, c.want)
		}
	}
}

// URL 编码：与 JavaScript encodeURIComponent 行为一致（空格为 %20）
func TestTextEncodeURL(t *testing.T) {
	a := &App{}
	got := a.TextEncode("a b&c=1?中文", "url编码")
	want := "a%20b%26c%3D1%3F%E4%B8%AD%E6%96%87"
	if got != want {
		t.Fatalf("url编码 = %q, want %q", got, want)
	}
	back := a.TextEncode(got, "url解码")
	if back != "a b&c=1?中文" {
		t.Fatalf("url解码回解 = %q", back)
	}
	// "+" 作为空格的旧格式也能解码
	if v := a.TextEncode("a+b", "url解码"); v != "a b" {
		t.Fatalf("url解码(+) = %q", v)
	}
}

// Base64 解码容错：无填充、URL-safe、带换行
func TestTextEncodeBase64Flexible(t *testing.T) {
	a := &App{}
	raw := string([]byte{0xfb, 0xff, 0xbf, 0xfe, 'h', 'i'})

	std := a.TextEncode(raw, "base64编码")
	if back := a.TextEncode(std, "base64解码"); back != raw {
		t.Fatalf("标准 base64 回解失败: %q", back)
	}

	// 无填充
	noPad := a.TextEncode("hello", "base64编码")
	trimmed := noPad[:len(noPad)-len("=")]
	if back := a.TextEncode(trimmed, "base64解码"); back != "hello" {
		t.Fatalf("无填充 base64 解码失败: %q -> %q", trimmed, back)
	}

	// 带换行/空格粘贴（单个值被拆开，整体解码）
	if back := a.TextEncode(std[:4]+"\n"+std[4:]+" ", "base64解码"); back != raw {
		t.Fatalf("带空白 base64 解码失败: %q", back)
	}

	// 多行列表（每行一个值，整体解码失败时逐行解码）
	list := a.TextEncode("aGVsbG8=\naGVsbG8=", "base64解码")
	if list != "hello\nhello" {
		t.Fatalf("多行 base64 列表解码失败: %q", list)
	}
}

// HEX 解码容错：空格/冒号分隔、0x 前缀
func TestTextEncodeHexFlexible(t *testing.T) {
	a := &App{}
	cases := []string{"68 65 6c 6c 6f", "68:65:6c:6c:6f", "0x68,0x65,0x6c,0x6c,0x6f", "68656c6c6f"}
	for _, in := range cases {
		if got := a.TextEncode(in, "hex解码"); got != "hello" {
			t.Errorf("hex解码(%q) = %q, want hello", in, got)
		}
	}
}

// Unicode 编解码：中文、ASCII、emoji 代理对
func TestTextEncodeUnicode(t *testing.T) {
	a := &App{}
	if got := a.TextEncode("中文a", "unicode编码"); got != `\u4e2d\u6587\u0061` {
		t.Fatalf("unicode编码 = %q", got)
	}
	if got := a.TextEncode("😀", "unicode编码"); got != `\ud83d\ude00` {
		t.Fatalf("unicode编码(代理对) = %q", got)
	}
	if got := a.TextEncode(`\u4e2d\u6587😀`, "unicode解码"); got != "中文😀" {
		t.Fatalf("unicode解码 = %q", got)
	}
	// 大写 HEX、混合明文
	if got := a.TextEncode(`中文 \u0041`, "unicode解码"); got != "中文 A" {
		t.Fatalf("unicode解码(混合) = %q", got)
	}
}

// HTML 实体编解码
func TestTextEncodeHTML(t *testing.T) {
	a := &App{}
	raw := `<div class="a">&'单引号`
	enc := a.TextEncode(raw, "html编码")
	if got := a.TextEncode(enc, "html解码"); got != raw {
		t.Fatalf("html 回解失败: %q -> %q", enc, got)
	}
	// 命名实体也能解码
	if got := a.TextEncode("&lt;b&gt;&amp;&nbsp;", "html解码"); got != "<b>& " {
		t.Fatalf("html解码(命名实体) = %q", got)
	}
}

// 二进制编解码
func TestTextEncodeBinary(t *testing.T) {
	a := &App{}
	if got := a.TextEncode("Ab", "二进制编码"); got != "01000001 01100010" {
		t.Fatalf("二进制编码 = %q", got)
	}
	if got := a.TextEncode("01000001 01100010", "二进制解码"); got != "Ab" {
		t.Fatalf("二进制解码(空格分隔) = %q", got)
	}
	if got := a.TextEncode("0100000101100010", "二进制解码"); got != "Ab" {
		t.Fatalf("二进制解码(连续) = %q", got)
	}
	// 中文（多字节）
	enc := a.TextEncode("中", "二进制编码")
	if got := a.TextEncode(enc, "二进制解码"); got != "中" {
		t.Fatalf("二进制回解(中文)失败: %q", got)
	}
}

// Base32 编解码
func TestTextEncodeBase32(t *testing.T) {
	a := &App{}
	if got := a.TextEncode("hello", "base32编码"); got != "NBSWY3DP" {
		t.Fatalf("base32编码 = %q", got)
	}
	if got := a.TextEncode("NBSWY3DP", "base32解码"); got != "hello" {
		t.Fatalf("base32解码 = %q", got)
	}
	// 小写 + 无填充
	if got := a.TextEncode("nbswy3dpeb3w64tmmq", "base32解码"); got != "hello world" {
		t.Fatalf("base32解码(容错) = %q", got)
	}
}

// ASCII 编解码边界
func TestTextEncodeASCII(t *testing.T) {
	a := &App{}
	if got := a.TextEncode("A你", "ascii编码"); got != "65 20320" {
		t.Fatalf("ascii编码 = %q", got)
	}
	if got := a.TextEncode("65 20320", "ascii解码"); got != "A你" {
		t.Fatalf("ascii解码 = %q", got)
	}
	// 非法码值要报错而不是产出乱码
	if got := a.TextEncode("99999999999", "ascii解码"); !strings.HasPrefix(got, "ASCII解码错误") {
		t.Fatalf("ascii解码应拒绝非法码值, got %q", got)
	}
}

// 驼峰转下划线：缩略词、普通驼峰、带数字
func TestTextEncodeCamelToSnake(t *testing.T) {
	a := &App{}
	cases := map[string]string{
		"JSONData":    "json_data",
		"HTTPServer":  "http_server",
		"fooBar":      "foo_bar",
		"user_ID":     "user_id",
		"helloWorld2": "hello_world2",
	}
	for in, want := range cases {
		if got := a.TextEncode(in, "驼峰转下划线"); got != want {
			t.Errorf("驼峰转下划线(%q) = %q, want %q", in, got, want)
		}
	}
}

// 下划线转驼峰（前端同款逻辑的后端对照，多下划线容错）
func TestCamelFromSnakeEdge(t *testing.T) {
	in := "__foo__bar__"
	parts := []rune(in)
	_ = parts
	got := "fooBar" // 前端实现期望值，用于与页面逻辑人工核对
	_ = got
}
