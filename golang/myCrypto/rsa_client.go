package myCrypto

import (
	"fmt"
	"log"
	"github.com/wyzzgzhdcxy/wcj-go-common/core"
)

func GenerateKey(keyBits int) {
	// 创建RSA工具实例
	rsaUtil := NewRSAUtil()
	// 生成2048位的RSA密钥对
	err := rsaUtil.GenerateKey(keyBits)
	if err != nil {
		log.Printf("生成密钥失败:%v", err)
	}

	keyDir := core.GetTempDir() + "/keys"
	core.MkDirALl0755(keyDir)
	// 保存密钥到文件
	err = rsaUtil.SavePrivateKeyToPemFile(keyDir + "/private.pem")
	if err != nil {
		log.Printf("保存私钥失败:%v", err)
	}

	err = rsaUtil.SavePublicKeyToPemFile(keyDir + "/public.pem")
	if err != nil {
		log.Printf("保存公钥失败:%v", err)
	}
}

func RsaCryptoStr(path string, plainText string) string {
	// 创建RSA工具实例
	rsaUtil := NewRSAUtil()
	err := rsaUtil.LoadPublicKeyFromPem(core.ReadFileToByte(path))
	if err != nil {
		return ""
	}

	// 加密
	ciphertext, err := rsaUtil.EncryptBase64([]byte(plainText))
	if err != nil {
		log.Printf("加密失败:%v", err)
	}
	fmt.Printf("加密结果: %x\n", ciphertext)
	return ciphertext
}

func RsaDeCryptoStr(path string, plainText string) string {
	// 创建RSA工具实例
	rsaUtil := NewRSAUtil()
	err := rsaUtil.LoadPrivateKeyFromPem(core.ReadFileToByte(path))
	if err != nil {
		return ""
	}

	// 加密
	ciphertext, err := rsaUtil.DecryptBase64(plainText)
	if err != nil {
		log.Printf("解密失败:%v", err)
	}
	fmt.Printf("解密结果: %x\n", ciphertext)
	return string(ciphertext)
}
