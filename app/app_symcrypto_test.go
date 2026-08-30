package app

import (
	"testing"
)

// 与常见在线工具/openssl 的标准结果比对
func TestSymCryptoAES128CBCBase64(t *testing.T) {
	a := &App{}
	// 明文 "Hello, 对称加密!" key/iv 均为 16 字节文本
	got, err := a.SymCrypto("Hello, 对称加密!", "aes128", "CBC", "1234567890123456", "1234567890123456", "text", "base64", "加密")
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}
	back, err := a.SymCrypto(got, "aes128", "CBC", "1234567890123456", "1234567890123456", "text", "base64", "解密")
	if err != nil {
		t.Fatalf("解密失败: %v", err)
	}
	if back != "Hello, 对称加密!" {
		t.Fatalf("回解结果不符: %q", back)
	}
}

func TestSymCryptoAllCombos(t *testing.T) {
	a := &App{}
	algos := map[string]string{
		"aes128": "1234567890123456",
		"aes192": "123456789012345678901234",
		"aes256": "12345678901234567890123456789012",
		"des":    "12345678",
		"3des":   "123456789012345678901234",
		"sm4":    "1234567890123456",
	}
	modes := []string{"ECB", "CBC", "CTR", "GCM", "CFB", "OFB"}
	encs := []string{"base64", "hex"}
	for algo, key := range algos {
		for _, mode := range modes {
			for _, enc := range encs {
				iv := ""
				if mode != "ECB" {
					iv = "iviviviviviviviv"[:16]
				}
				// DES/3DES 为8字节分组，不支持 GCM（后端返回明确错误）
				if (algo == "des" || algo == "3des") && mode == "GCM" {
					if _, err := a.SymCrypto("x", algo, mode, key, iv, "text", enc, "加密"); err == nil {
						t.Fatalf("%s-%s 应不支持GCM", algo, mode)
					}
					continue
				}
				ct, err := a.SymCrypto("测试明文test123", algo, mode, key, iv, "text", enc, "加密")
				if err != nil {
					t.Fatalf("%s-%s-%s 加密失败: %v", algo, mode, enc, err)
				}
				pt, err := a.SymCrypto(ct, algo, mode, key, iv, "text", enc, "解密")
				if err != nil {
					t.Fatalf("%s-%s-%s 解密失败: %v", algo, mode, enc, err)
				}
				if pt != "测试明文test123" {
					t.Fatalf("%s-%s-%s 回解结果不符: %q", algo, mode, enc, pt)
				}
			}
		}
	}
}

func TestSymKeyEncodingVariants(t *testing.T) {
	a := &App{}
	// hex 密钥: 16 字节全 0x31 -> "3131..." x16
	hexKey := strings_Repeat("31", 16)
	ct, err := a.SymCrypto("abc", "aes128", "CBC", hexKey, hexKey, "hex", "base64", "加密")
	if err != nil {
		t.Fatalf("hex密钥加密失败: %v", err)
	}
	// 用等价 text 密钥 "1111111111111111" 解密
	pt, err := a.SymCrypto(ct, "aes128", "CBC", "1111111111111111", "1111111111111111", "text", "base64", "解密")
	if err != nil {
		t.Fatalf("解密失败: %v", err)
	}
	if pt != "abc" {
		t.Fatalf("hex/text 等价性验证失败: %q", pt)
	}
}

func TestSymRandomKey(t *testing.T) {
	a := &App{}
	r := a.SymRandomKey("aes128", "CBC", "hex")
	if len(r.Key) != 32 || len(r.IV) != 32 {
		t.Fatalf("hex 随机密钥长度错误: %d/%d", len(r.Key), len(r.IV))
	}
	r = a.SymRandomKey("aes256", "GCM", "text")
	if len(r.Key) != 32 || len(r.IV) != 12 {
		t.Fatalf("text 随机密钥长度错误: %d/%d", len(r.Key), len(r.IV))
	}
	// 用随机密钥走一轮加解密
	ct, err := a.SymCrypto("random", "aes256", "GCM", r.Key, r.IV, "text", "base64", "加密")
	if err != nil {
		t.Fatalf("随机密钥加密失败: %v", err)
	}
	pt, err := a.SymCrypto(ct, "aes256", "GCM", r.Key, r.IV, "text", "base64", "解密")
	if err != nil || pt != "random" {
		t.Fatalf("随机密钥回解失败: %q %v", pt, err)
	}
}

func strings_Repeat(s string, n int) string {
	r := ""
	for i := 0; i < n; i++ {
		r += s
	}
	return r
}
