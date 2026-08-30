// Package app 非对称加解密（RSA / SM2）的前端绑定方法。
// 由 frontend/src/pages/crypto_encryption.vue（非对称加解密页面）调用。
package app

import (
	"encoding/base64"
	"errors"
	"path/filepath"
	"strings"

	"github.com/wyzzgzhdcxy/wcj-go-common/core"
	"wcj-go-text/golang/myCrypto"
)

// GetKeysDir 返回密钥保存目录（%LOCALAPPDATA% 等临时目录下的 keys 目录）
func (a *App) GetKeysDir() string {
	return core.GetTempDir() + "/keys"
}

// GenerateKey 生成 RSA 密钥对并保存到密钥目录，返回目录路径
func (a *App) GenerateKey(keyBits int) (string, error) {
	if keyBits <= 0 {
		keyBits = 2048
	}
	rsaUtil := myCrypto.NewRSAUtil()
	if err := rsaUtil.GenerateKey(keyBits); err != nil {
		return "", err
	}
	dir := a.GetKeysDir()
	core.MkDirALl0755(dir)
	if err := rsaUtil.SavePrivateKeyToPemFile(filepath.Join(dir, "rsa_private.pem")); err != nil {
		return "", err
	}
	if err := rsaUtil.SavePublicKeyToPemFile(filepath.Join(dir, "rsa_public.pem")); err != nil {
		return "", err
	}
	return dir, nil
}

// Sm2GenerateKey 生成 SM2 密钥对并保存到密钥目录，返回目录路径
func (a *App) Sm2GenerateKey() (string, error) {
	sm2Util := myCrypto.NewSM2Util()
	if err := sm2Util.GenerateKey(); err != nil {
		return "", err
	}
	dir := a.GetKeysDir()
	core.MkDirALl0755(dir)
	if err := sm2Util.SavePrivateKeyToPemFile(filepath.Join(dir, "sm2_private.pem")); err != nil {
		return "", err
	}
	if err := sm2Util.SavePublicKeyToPemFile(filepath.Join(dir, "sm2_public.pem")); err != nil {
		return "", err
	}
	return dir, nil
}

// RsaCryptoStr 使用 RSA 公钥加密文本，输出 Base64
func (a *App) RsaCryptoStr(pkPath string, ori string) (string, error) {
	rsaUtil := myCrypto.NewRSAUtil()
	if err := rsaUtil.LoadPublicKeyFromPem(core.ReadFileToByte(pkPath)); err != nil {
		return "", err
	}
	return rsaUtil.EncryptBase64([]byte(ori))
}

// RsaDeCryptoStr 使用 RSA 私钥解密 Base64 密文
func (a *App) RsaDeCryptoStr(pkPath string, ori string) (string, error) {
	rsaUtil := myCrypto.NewRSAUtil()
	if err := rsaUtil.LoadPrivateKeyFromPem(core.ReadFileToByte(pkPath)); err != nil {
		return "", err
	}
	plain, err := rsaUtil.DecryptBase64(ori)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

// Sm2CryptoStr 使用 SM2 公钥加密文本，输出 Base64（标准 SM2 ASN.1 密文）
func (a *App) Sm2CryptoStr(pkPath string, ori string) (string, error) {
	sm2Util := myCrypto.NewSM2Util()
	if err := sm2Util.LoadPublicKeyFromPem(core.ReadFileToByte(pkPath)); err != nil {
		return "", err
	}
	return sm2Util.EncryptBase64([]byte(ori))
}

// Sm2DeCryptoStr 使用 SM2 私钥解密 Base64 密文
func (a *App) Sm2DeCryptoStr(pkPath string, ori string) (string, error) {
	sm2Util := myCrypto.NewSM2Util()
	if err := sm2Util.LoadPrivateKeyFromPem(core.ReadFileToByte(pkPath)); err != nil {
		return "", err
	}
	plain, err := sm2Util.DecryptBase64(ori)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

// AsymEncryptFile 混合加密文件（数字信封：AES-256-GCM 加密内容，公钥加密会话密钥），
// 输出文件为原文件名 + ".enc"，返回输出路径
func (a *App) AsymEncryptFile(algo string, keyPath string, inFile string) (string, error) {
	if strings.TrimSpace(inFile) == "" {
		return "", errors.New("请先选择待加密文件")
	}
	outPath := inFile + ".enc"
	if err := myCrypto.EncryptFileWithPublicKey(strings.ToLower(strings.TrimSpace(algo)), keyPath, inFile, outPath); err != nil {
		return "", err
	}
	return outPath, nil
}

// AsymDecryptFile 解密混合加密文件（算法从文件头自动识别），
// 输出文件去掉 ".enc" 后缀（无后缀则追加 ".dec"），返回输出路径
func (a *App) AsymDecryptFile(keyPath string, inFile string) (string, error) {
	if strings.TrimSpace(inFile) == "" {
		return "", errors.New("请先选择待解密文件")
	}
	var outPath string
	if strings.HasSuffix(strings.ToLower(inFile), ".enc") {
		outPath = inFile[:len(inFile)-4]
	} else {
		outPath = inFile + ".dec"
	}
	if err := myCrypto.DecryptFileWithPrivateKey(keyPath, inFile, outPath); err != nil {
		return "", err
	}
	return outPath, nil
}

// RsaSignStr 使用 RSA 私钥签名（SHA256 摘要 + PKCS1v15），返回 Base64 签名
func (a *App) RsaSignStr(priPath string, data string) (string, error) {
	rsaUtil := myCrypto.NewRSAUtil()
	if err := rsaUtil.LoadPrivateKeyFromPem(core.ReadFileToByte(priPath)); err != nil {
		return "", err
	}
	return rsaUtil.SignBase64([]byte(data))
}

// RsaVerifyStr 使用 RSA 公钥验签，返回验签是否通过（签名不匹配返回 false 而非错误）
func (a *App) RsaVerifyStr(pkPath string, data string, signBase64 string) (bool, error) {
	rsaUtil := myCrypto.NewRSAUtil()
	if err := rsaUtil.LoadPublicKeyFromPem(core.ReadFileToByte(pkPath)); err != nil {
		return false, err
	}
	sig, err := base64.StdEncoding.DecodeString(strings.TrimSpace(signBase64))
	if err != nil {
		return false, errors.New("签名不是合法的 Base64")
	}
	if err := rsaUtil.Verify([]byte(data), sig); err != nil {
		return false, nil
	}
	return true, nil
}

// Sm2SignStr 使用 SM2 私钥签名（SM3 摘要 + 默认 UID），返回 Base64 编码的 ASN.1 签名
func (a *App) Sm2SignStr(priPath string, data string) (string, error) {
	sm2Util := myCrypto.NewSM2Util()
	if err := sm2Util.LoadPrivateKeyFromPem(core.ReadFileToByte(priPath)); err != nil {
		return "", err
	}
	return sm2Util.SignBase64([]byte(data))
}

// Sm2VerifyStr 使用 SM2 公钥验签，返回验签是否通过（签名不匹配返回 false 而非错误）
func (a *App) Sm2VerifyStr(pkPath string, data string, signBase64 string) (bool, error) {
	sm2Util := myCrypto.NewSM2Util()
	if err := sm2Util.LoadPublicKeyFromPem(core.ReadFileToByte(pkPath)); err != nil {
		return false, err
	}
	return sm2Util.VerifyBase64([]byte(data), signBase64)
}
