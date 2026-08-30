package myCrypto

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"os"
	"strings"

	"github.com/tjfoc/gmsm/sm2"
	gmx509 "github.com/tjfoc/gmsm/x509"
)

// SM2Util SM2工具结构体（国密非对称算法）
type SM2Util struct {
	privateKey *sm2.PrivateKey
	publicKey  *sm2.PublicKey
}

// NewSM2Util 创建SM2工具实例
func NewSM2Util() *SM2Util {
	return &SM2Util{}
}

// GenerateKey 生成SM2密钥对（固定256位）
func (s *SM2Util) GenerateKey() error {
	privateKey, err := sm2.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}
	s.privateKey = privateKey
	s.publicKey = &privateKey.PublicKey
	return nil
}

// LoadPrivateKeyFromPem 从PEM格式加载私钥
func (s *SM2Util) LoadPrivateKeyFromPem(privateKeyPem []byte) error {
	privateKey, err := gmx509.ReadPrivateKeyFromPem(privateKeyPem, nil)
	if err != nil {
		return err
	}
	s.privateKey = privateKey
	s.publicKey = &privateKey.PublicKey
	return nil
}

// LoadPublicKeyFromPem 从PEM格式加载公钥
func (s *SM2Util) LoadPublicKeyFromPem(publicKeyPem []byte) error {
	publicKey, err := gmx509.ReadPublicKeyFromPem(publicKeyPem)
	if err != nil {
		return err
	}
	s.publicKey = publicKey
	return nil
}

// SavePrivateKeyToPemFile 将私钥保存到PEM文件（PKCS8/未加密）
func (s *SM2Util) SavePrivateKeyToPemFile(filename string) error {
	if s.privateKey == nil {
		return errors.New("private key is nil")
	}
	pemBytes, err := gmx509.WritePrivateKeyToPem(s.privateKey, nil)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, pemBytes, 0644)
}

// SavePublicKeyToPemFile 将公钥保存到PEM文件
func (s *SM2Util) SavePublicKeyToPemFile(filename string) error {
	if s.publicKey == nil {
		return errors.New("public key is nil")
	}
	pemBytes, err := gmx509.WritePublicKeyToPem(s.publicKey)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, pemBytes, 0644)
}

// EncryptBase64 使用公钥加密，输出标准 SM2 密文（ASN.1/DER）的 Base64
func (s *SM2Util) EncryptBase64(plaintext []byte) (string, error) {
	if s.publicKey == nil {
		return "", errors.New("public key is nil")
	}
	ciphertext, err := sm2.EncryptAsn1(s.publicKey, plaintext, rand.Reader)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptBase64 解密 Base64 编码的 SM2 密文（ASN.1/DER）
func (s *SM2Util) DecryptBase64(ciphertextBase64 string) ([]byte, error) {
	if s.privateKey == nil {
		return nil, errors.New("private key is nil")
	}
	ciphertext, err := base64.StdEncoding.DecodeString(strings.TrimSpace(ciphertextBase64))
	if err != nil {
		return nil, err
	}
	return sm2.DecryptAsn1(s.privateKey, ciphertext)
}

// SignBase64 使用私钥签名（SM3 摘要 + 默认 UID），返回 Base64 编码的 ASN.1(DER) 签名
func (s *SM2Util) SignBase64(data []byte) (string, error) {
	if s.privateKey == nil {
		return "", errors.New("private key is nil")
	}
	sig, err := s.privateKey.Sign(rand.Reader, data, nil)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(sig), nil
}

// VerifyBase64 使用公钥验证 Base64 编码的签名，返回验签是否通过
func (s *SM2Util) VerifyBase64(data []byte, sigBase64 string) (bool, error) {
	if s.publicKey == nil {
		return false, errors.New("public key is nil")
	}
	sig, err := base64.StdEncoding.DecodeString(strings.TrimSpace(sigBase64))
	if err != nil {
		return false, errors.New("签名不是合法的 Base64")
	}
	return s.publicKey.Verify(data, sig), nil
}
