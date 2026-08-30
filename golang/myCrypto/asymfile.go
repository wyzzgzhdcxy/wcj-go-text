package myCrypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"os"

	"github.com/tjfoc/gmsm/sm2"
)

// asymMagic 混合加密文件头标识
const asymMagic = "WCJASYM1"

// 算法标识
const (
	asymAlgoRSA byte = 0x01 // RSA-OAEP(SHA256) 加密会话密钥
	asymAlgoSM2 byte = 0x02 // SM2(ASN.1) 加密会话密钥
)

// 加密文件格式：
// [8字节 magic][1字节算法][2字节会话密钥密文长度][会话密钥密文][AES-256-GCM 密文(Nonce+密文+Tag)]
// 即：随机 AES-256 会话密钥加密文件内容，非对称公钥加密会话密钥（数字信封）

// EncryptFileWithPublicKey 混合加密文件（algo: rsa | sm2）
func EncryptFileWithPublicKey(algo string, publicKeyPath, inPath, outPath string) error {
	plain, err := os.ReadFile(inPath)
	if err != nil {
		return err
	}

	// 生成随机 AES-256 会话密钥并加密文件内容
	sessionKey := make([]byte, 32)
	if _, err := rand.Read(sessionKey); err != nil {
		return err
	}
	gcm, err := newGCM(sessionKey)
	if err != nil {
		return err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return err
	}
	sealed := gcm.Seal(nonce, nonce, plain, nil)

	// 用非对称公钥加密会话密钥
	var algoByte byte
	var encKey []byte
	switch algo {
	case "rsa", "rsa2048", "rsa4096":
		keyPem, err := ReadFileAll(publicKeyPath)
		if err != nil {
			return errors.New("读取公钥文件失败: " + err.Error())
		}
		rsaUtil := NewRSAUtil()
		if err := rsaUtil.LoadPublicKeyFromPem(keyPem); err != nil {
			return errors.New("公钥文件格式错误: " + err.Error())
		}
		if rsaUtil.publicKey == nil {
			return errors.New("public key is nil")
		}
		encKey, err = rsa.EncryptOAEP(sha256.New(), rand.Reader, rsaUtil.publicKey, sessionKey, nil)
		algoByte = asymAlgoRSA
	case "sm2":
		keyPem, err := ReadFileAll(publicKeyPath)
		if err != nil {
			return errors.New("读取公钥文件失败: " + err.Error())
		}
		sm2Util := NewSM2Util()
		if err := sm2Util.LoadPublicKeyFromPem(keyPem); err != nil {
			return errors.New("公钥文件格式错误: " + err.Error())
		}
		if sm2Util.publicKey == nil {
			return errors.New("public key is nil")
		}
		encKey, err = sm2.EncryptAsn1(sm2Util.publicKey, sessionKey, rand.Reader)
		algoByte = asymAlgoSM2
	default:
		return errors.New("不支持的算法: " + algo)
	}
	if err != nil {
		return err
	}

	// 组装文件
	out := make([]byte, 0, len(asymMagic)+3+len(encKey)+len(sealed))
	out = append(out, []byte(asymMagic)...)
	out = append(out, algoByte)
	lenBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(lenBuf, uint16(len(encKey)))
	out = append(out, lenBuf...)
	out = append(out, encKey...)
	out = append(out, sealed...)

	return os.WriteFile(outPath, out, 0644)
}

// DecryptFileWithPrivateKey 解密混合加密文件（算法从文件头自动识别）
func DecryptFileWithPrivateKey(privateKeyPath, inPath, outPath string) error {
	data, err := os.ReadFile(inPath)
	if err != nil {
		return err
	}
	headerLen := len(asymMagic) + 3
	if len(data) < headerLen || string(data[:len(asymMagic)]) != asymMagic {
		return errors.New("不是本工具生成的加密文件")
	}
	algo := data[len(asymMagic)]
	keyLen := int(binary.BigEndian.Uint16(data[len(asymMagic)+1 : len(asymMagic)+3]))
	if len(data) < headerLen+keyLen {
		return errors.New("加密文件已损坏")
	}
	encKey := data[headerLen : headerLen+keyLen]
	sealed := data[headerLen+keyLen:]

	// 解密会话密钥
	var sessionKey []byte
	switch algo {
	case asymAlgoRSA:
		keyPem, err := ReadFileAll(privateKeyPath)
		if err != nil {
			return errors.New("读取私钥文件失败: " + err.Error())
		}
		rsaUtil := NewRSAUtil()
		if err := rsaUtil.LoadPrivateKeyFromPem(keyPem); err != nil {
			return errors.New("私钥文件格式错误: " + err.Error())
		}
		if rsaUtil.privateKey == nil {
			return errors.New("private key is nil")
		}
		sessionKey, err = rsa.DecryptOAEP(sha256.New(), rand.Reader, rsaUtil.privateKey, encKey, nil)
	case asymAlgoSM2:
		keyPem, err := ReadFileAll(privateKeyPath)
		if err != nil {
			return errors.New("读取私钥文件失败: " + err.Error())
		}
		sm2Util := NewSM2Util()
		if err := sm2Util.LoadPrivateKeyFromPem(keyPem); err != nil {
			return errors.New("私钥文件格式错误: " + err.Error())
		}
		if sm2Util.privateKey == nil {
			return errors.New("private key is nil")
		}
		sessionKey, err = sm2.DecryptAsn1(sm2Util.privateKey, encKey)
	default:
		return errors.New("未知的加密算法标识")
	}
	if err != nil {
		return errors.New("会话密钥解密失败（私钥不匹配）: " + err.Error())
	}

	gcm, err := newGCM(sessionKey)
	if err != nil {
		return err
	}
	plain, err := gcm.Open(nil, sealed[:gcm.NonceSize()], sealed[gcm.NonceSize():], nil)
	if err != nil {
		return errors.New("文件内容解密失败: " + err.Error())
	}
	return os.WriteFile(outPath, plain, 0644)
}

func newGCM(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

// ReadFileAll 读取文件内容（包内辅助）
func ReadFileAll(path string) ([]byte, error) {
	return os.ReadFile(path)
}
