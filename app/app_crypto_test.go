package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// RSA 文本加解密回环
func TestRsaTextRoundTrip(t *testing.T) {
	a := &App{}
	if _, err := a.GenerateKey(1024); err != nil {
		t.Fatalf("RSA生成密钥失败: %v", err)
	}
	dir := a.GetKeysDir()
	ct, err := a.RsaCryptoStr(filepath.Join(dir, "rsa_public.pem"), "RSA测试文本123")
	if err != nil {
		t.Fatalf("RSA加密失败: %v", err)
	}
	pt, err := a.RsaDeCryptoStr(filepath.Join(dir, "rsa_private.pem"), ct)
	if err != nil {
		t.Fatalf("RSA解密失败: %v", err)
	}
	if pt != "RSA测试文本123" {
		t.Fatalf("RSA回解结果不符: %q", pt)
	}
}

// SM2 文本加解密回环
func TestSm2TextRoundTrip(t *testing.T) {
	a := &App{}
	dir, err := a.Sm2GenerateKey()
	if err != nil {
		t.Fatalf("SM2生成密钥失败: %v", err)
	}
	ct, err := a.Sm2CryptoStr(filepath.Join(dir, "sm2_public.pem"), "SM2国密测试文本😀")
	if err != nil {
		t.Fatalf("SM2加密失败: %v", err)
	}
	pt, err := a.Sm2DeCryptoStr(filepath.Join(dir, "sm2_private.pem"), ct)
	if err != nil {
		t.Fatalf("SM2解密失败: %v", err)
	}
	if pt != "SM2国密测试文本😀" {
		t.Fatalf("SM2回解结果不符: %q", pt)
	}
	// SM2 密文应为 ASN.1 序列
	if !strings.HasPrefix(ct, "MF") && !strings.HasPrefix(ct, "MI") {
		t.Logf("SM2密文Base64前缀: %q（非典型ASN.1前缀，仅提示）", ct[:4])
	}
}

// 混合加密文件回环（RSA / SM2）
func TestAsymFileRoundTrip(t *testing.T) {
	a := &App{}
	dir := a.GetKeysDir()
	if _, err := a.GenerateKey(1024); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Sm2GenerateKey(); err != nil {
		t.Fatal(err)
	}

	src := filepath.Join(t.TempDir(), "plain.txt")
	content := strings.Repeat("混合加密文件内容测试 hello 你好 1234567890\n", 100)
	if err := os.WriteFile(src, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	t.Run("rsa", func(t *testing.T) {
		encPath, err := a.AsymEncryptFile("rsa", filepath.Join(dir, "rsa_public.pem"), src)
		if err != nil {
			t.Fatalf("RSA文件加密失败: %v", err)
		}
		if !strings.HasSuffix(encPath, ".enc") {
			t.Fatalf("输出路径应以.enc结尾: %q", encPath)
		}
		decPath, err := a.AsymDecryptFile(filepath.Join(dir, "rsa_private.pem"), encPath)
		if err != nil {
			t.Fatalf("RSA文件解密失败: %v", err)
		}
		got, _ := os.ReadFile(decPath)
		if string(got) != content {
			t.Fatalf("RSA文件回解内容不符")
		}
	})

	t.Run("sm2", func(t *testing.T) {
		encPath, err := a.AsymEncryptFile("sm2", filepath.Join(dir, "sm2_public.pem"), src)
		if err != nil {
			t.Fatalf("SM2文件加密失败: %v", err)
		}
		decPath, err := a.AsymDecryptFile(filepath.Join(dir, "sm2_private.pem"), encPath)
		if err != nil {
			t.Fatalf("SM2文件解密失败: %v", err)
		}
		got, _ := os.ReadFile(decPath)
		if string(got) != content {
			t.Fatalf("SM2文件回解内容不符")
		}
	})

	t.Run("wrongKey", func(t *testing.T) {
		encPath, err := a.AsymEncryptFile("sm2", filepath.Join(dir, "sm2_public.pem"), src)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := a.AsymDecryptFile(filepath.Join(dir, "rsa_private.pem"), encPath); err == nil {
			t.Fatal("用错误私钥解密应当失败")
		}
	})
}

// RSA 签名/验签
func TestRsaSignVerify(t *testing.T) {
	a := &App{}
	dir := a.GetKeysDir()
	priPath := filepath.Join(dir, "rsa_private.pem")
	pubPath := filepath.Join(dir, "rsa_public.pem")

	sig, err := a.RsaSignStr(priPath, "待签名数据 RSA")
	if err != nil {
		t.Fatalf("RSA签名失败: %v", err)
	}
	ok, err := a.RsaVerifyStr(pubPath, "待签名数据 RSA", sig)
	if err != nil || !ok {
		t.Fatalf("RSA验签应通过: ok=%v err=%v", ok, err)
	}
	// 篡改数据后验签应失败
	ok, err = a.RsaVerifyStr(pubPath, "待签名数据 RSA 篡改", sig)
	if err != nil || ok {
		t.Fatalf("RSA篡改数据验签应失败: ok=%v err=%v", ok, err)
	}
	// 篡改签名后验签应失败
	ok, err = a.RsaVerifyStr(pubPath, "待签名数据 RSA", sig[:len(sig)-4]+"AAAA")
	if err != nil || ok {
		t.Fatalf("RSA篡改签名验签应失败: ok=%v err=%v", ok, err)
	}
}

// SM2 签名/验签
func TestSm2SignVerify(t *testing.T) {
	a := &App{}
	dir := a.GetKeysDir()
	priPath := filepath.Join(dir, "sm2_private.pem")
	pubPath := filepath.Join(dir, "sm2_public.pem")

	sig, err := a.Sm2SignStr(priPath, "待签名数据 SM2 国密")
	if err != nil {
		t.Fatalf("SM2签名失败: %v", err)
	}
	ok, err := a.Sm2VerifyStr(pubPath, "待签名数据 SM2 国密", sig)
	if err != nil || !ok {
		t.Fatalf("SM2验签应通过: ok=%v err=%v", ok, err)
	}
	// 篡改数据后验签应失败
	ok, err = a.Sm2VerifyStr(pubPath, "待签名数据 SM2 篡改", sig)
	if err != nil || ok {
		t.Fatalf("SM2篡改数据验签应失败: ok=%v err=%v", ok, err)
	}
	// 签名值应为 ASN.1 序列（SEQUENCE 头 0x30）
	if !strings.HasPrefix(sig, "ME") && !strings.HasPrefix(sig, "MI") {
		t.Logf("SM2签名Base64前缀: %q（非典型ASN.1前缀，仅提示）", sig[:4])
	}
}
