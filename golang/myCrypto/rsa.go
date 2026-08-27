package myCrypto

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"os"
)

// RSAUtil RSA工具结构体
type RSAUtil struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

// NewRSAUtil 创建RSA工具实例
func NewRSAUtil() *RSAUtil {
	return &RSAUtil{}
}

// GenerateKey 生成RSA密钥对
// bits: 密钥长度，推荐2048或以上
func (r *RSAUtil) GenerateKey(bits int) error {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return err
	}
	r.privateKey = privateKey
	r.publicKey = &privateKey.PublicKey
	return nil
}

// LoadPrivateKeyFromPem 从PEM格式加载私钥
func (r *RSAUtil) LoadPrivateKeyFromPem(privateKeyPem []byte) error {
	block, _ := pem.Decode(privateKeyPem)
	if block == nil {
		return errors.New("failed to parse PEM block containing the private key")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return err
	}

	r.privateKey = privateKey
	r.publicKey = &privateKey.PublicKey
	return nil
}

// LoadPublicKeyFromPem 从PEM格式加载公钥
func (r *RSAUtil) LoadPublicKeyFromPem(publicKeyPem []byte) error {
	block, _ := pem.Decode(publicKeyPem)
	if block == nil {
		return errors.New("failed to parse PEM block containing the public key")
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return err
	}

	switch pub := publicKey.(type) {
	case *rsa.PublicKey:
		r.publicKey = pub
		return nil
	default:
		return errors.New("key type is not RSA")
	}
}

// SavePrivateKeyToPemFile 将私钥保存到PEM文件
func (r *RSAUtil) SavePrivateKeyToPemFile(filename string) error {
	if r.privateKey == nil {
		return errors.New("private key is nil")
	}

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(r.privateKey)
	privateKeyPem := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return pem.Encode(file, privateKeyPem)
}

// SavePublicKeyToPemFile 将公钥保存到PEM文件
func (r *RSAUtil) SavePublicKeyToPemFile(filename string) error {
	if r.publicKey == nil {
		return errors.New("public key is nil")
	}

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(r.publicKey)
	if err != nil {
		return err
	}

	publicKeyPem := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return pem.Encode(file, publicKeyPem)
}

// Encrypt 使用公钥加密
func (r *RSAUtil) Encrypt(plaintext []byte) ([]byte, error) {
	if r.publicKey == nil {
		return nil, errors.New("public key is nil")
	}

	ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, r.publicKey, plaintext)
	if err != nil {
		return nil, err
	}

	return ciphertext, nil
}

// Decrypt 使用私钥解密
func (r *RSAUtil) Decrypt(ciphertext []byte) ([]byte, error) {
	if r.privateKey == nil {
		return nil, errors.New("private key is nil")
	}

	plaintext, err := rsa.DecryptPKCS1v15(rand.Reader, r.privateKey, ciphertext)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// Sign 使用私钥签名
func (r *RSAUtil) Sign(data []byte) ([]byte, error) {
	if r.privateKey == nil {
		return nil, errors.New("private key is nil")
	}

	hashed := sha256.Sum256(data)
	signature, err := rsa.SignPKCS1v15(rand.Reader, r.privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return nil, err
	}

	return signature, nil
}

// Verify 使用公钥验证签名
func (r *RSAUtil) Verify(data []byte, signature []byte) error {
	if r.publicKey == nil {
		return errors.New("public key is nil")
	}

	hashed := sha256.Sum256(data)
	return rsa.VerifyPKCS1v15(r.publicKey, crypto.SHA256, hashed[:], signature)
}

// EncryptBase64 加密并返回Base64编码结果
func (r *RSAUtil) EncryptBase64(plaintext []byte) (string, error) {
	ciphertext, err := r.Encrypt(plaintext)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptBase64 解密Base64编码的密文
func (r *RSAUtil) DecryptBase64(ciphertextBase64 string) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return nil, err
	}
	return r.Decrypt(ciphertext)
}

// SignBase64 签名并返回Base64编码结果
func (r *RSAUtil) SignBase64(data []byte) (string, error) {
	signature, err := r.Sign(data)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}

// VerifyBase64 验证Base64编码的签名
func (r *RSAUtil) VerifyBase64(data []byte, signatureBase64 string) error {
	signature, err := base64.StdEncoding.DecodeString(signatureBase64)
	if err != nil {
		return err
	}
	return r.Verify(data, signature)
}
