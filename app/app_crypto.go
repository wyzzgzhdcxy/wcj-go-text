package app

import (
	"wcj-go-text/golang/myCrypto"
)

// GenerateKey 生成 RSA 密钥对
func (a *App) GenerateKey(keyBits int) bool {
	if keyBits <= 0 {
		keyBits = 2048
	}
	myCrypto.GenerateKey(keyBits)
	return true
}

// RsaCryptoStr RSA 加密
func (a *App) RsaCryptoStr(pkPath string, ori string) string {
	return myCrypto.RsaCryptoStr(pkPath, ori)
}

// RsaDeCryptoStr RSA 解密
func (a *App) RsaDeCryptoStr(pkPath string, ori string) string {
	return myCrypto.RsaDeCryptoStr(pkPath, ori)
}
