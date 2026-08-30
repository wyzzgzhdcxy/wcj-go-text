// Package app 对称加解密（AES/DES/3DES/SM4）的前端绑定方法。
// 布局与"常用编码"页面配套：由 frontend/src/pages/symmetricCrypto.vue 调用。
package app

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"math/big"
	"strings"

	"github.com/tjfoc/gmsm/sm4"
)

// SymKeyIV 随机密钥生成结果
type SymKeyIV struct {
	Key string `json:"key"`
	IV  string `json:"iv"`
}

// symAlgoSpec 各算法的密钥长度与分组长度（字节）
var symAlgoSpec = map[string]struct{ keyLen, blockLen int }{
	"aes128": {16, 16},
	"aes192": {24, 16},
	"aes256": {32, 16},
	"des":    {8, 8},
	"3des":   {24, 8},
	"sm4":    {16, 16},
}

// symRandCharset 生成"文本"格式随机密钥时使用的可打印字符集
const symRandCharset = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// SymCrypto 对称加解密
//
//	algo:       aes128 | aes192 | aes256 | des | 3des | sm4
//	mode:       ECB | CBC | CTR | GCM | CFB | OFB
//	key:        密钥文本，按 keyEncoding 解析为字节，不足自动补 0，超出自动截断
//	iv:         IV 文本（ECB 不需要；GCM 为 12 字节 Nonce，其余为分组长度）
//	keyEncoding: text | hex | base64，密钥/IV 的编码方式
//	outEncoding: base64 | hex，密文的输出（解密时输入）格式
//	opType:     加密 | 解密
func (a *App) SymCrypto(plainText string, algo string, mode string, key string, iv string, keyEncoding string, outEncoding string, opType string) (string, error) {
	spec, ok := symAlgoSpec[strings.ToLower(strings.TrimSpace(algo))]
	if !ok {
		return "", errors.New("不支持的算法: " + algo)
	}

	mode = strings.ToUpper(strings.TrimSpace(mode))
	keyBytes, err := symKeyBytes(key, keyEncoding, spec.keyLen, "密钥")
	if err != nil {
		return "", err
	}

	block, err := symNewBlock(strings.ToLower(strings.TrimSpace(algo)), keyBytes)
	if err != nil {
		return "", err
	}

	encrypt := strings.TrimSpace(opType) == "加密" || strings.EqualFold(opType, "encrypt")

	if mode == "ECB" {
		return symECB(block, []byte(plainText), encrypt, outEncoding)
	}

	ivLen := spec.blockLen
	if mode == "GCM" {
		ivLen = 12
	}
	ivBytes, err := symKeyBytes(iv, keyEncoding, ivLen, "IV")
	if err != nil {
		return "", err
	}

	switch mode {
	case "CBC":
		return symCBC(block, ivBytes, []byte(plainText), encrypt, outEncoding)
	case "CTR":
		return symStream(cipher.NewCTR(block, ivBytes), []byte(plainText), encrypt, outEncoding)
	case "GCM":
		if block.BlockSize() != 16 {
			return "", errors.New(mode + "模式要求16字节分组，" + strings.ToUpper(strings.TrimSpace(algo)) + "不支持")
		}
		return symGCM(block, ivBytes, []byte(plainText), encrypt, outEncoding)
	case "CFB":
		return symCFB(block, ivBytes, []byte(plainText), encrypt, outEncoding)
	case "OFB":
		return symStream(cipher.NewOFB(block, ivBytes), []byte(plainText), encrypt, outEncoding)
	default:
		return "", errors.New("不支持的模式: " + mode)
	}
}

// SymRandomKey 生成随机密钥和 IV
//
//	algo: 算法名；mode: 模式名（决定 IV 长度，ECB 返回空 IV）；
//	encoding: text | hex | base64（text 生成等长可打印字符串）
func (a *App) SymRandomKey(algo string, mode string, encoding string) SymKeyIV {
	spec, ok := symAlgoSpec[strings.ToLower(strings.TrimSpace(algo))]
	if !ok {
		return SymKeyIV{}
	}

	mode = strings.ToUpper(strings.TrimSpace(mode))
	encoding = strings.ToLower(strings.TrimSpace(encoding))

	ivLen := 0
	if mode != "ECB" {
		ivLen = spec.blockLen
		if mode == "GCM" {
			ivLen = 12
		}
	}

	return SymKeyIV{
		Key: symRandomText(spec.keyLen, encoding),
		IV:  symRandomText(ivLen, encoding),
	}
}

// symNewBlock 根据算法名构造 cipher.Block
func symNewBlock(algo string, key []byte) (cipher.Block, error) {
	switch algo {
	case "aes128", "aes192", "aes256":
		return aes.NewCipher(key)
	case "des":
		return des.NewCipher(key)
	case "3des":
		return des.NewTripleDESCipher(key)
	case "sm4":
		return sm4.NewCipher(key)
	default:
		return nil, errors.New("不支持的算法: " + algo)
	}
}

// symKeyBytes 将用户输入的密钥/IV 文本解析为指定长度的字节：
// 按 encoding 解码后不足补 0、超出截断，保证长度固定，与常见在线工具行为一致。
func symKeyBytes(s string, encoding string, size int, name string) ([]byte, error) {
	encoding = strings.ToLower(strings.TrimSpace(encoding))
	var b []byte
	var err error

	switch encoding {
	case "hex":
		b, err = hex.DecodeString(strings.TrimSpace(s))
		if err != nil {
			return nil, errors.New(name + "不是合法的 HEX: " + err.Error())
		}
	case "base64":
		b, err = symBase64Decode(strings.TrimSpace(s))
		if err != nil {
			return nil, errors.New(name + "不是合法的 Base64: " + err.Error())
		}
	default:
		b = []byte(s)
	}

	if len(b) == 0 {
		return nil, errors.New(name + "不能为空")
	}
	if len(b) > size {
		return b[:size], nil
	}
	return append(b, make([]byte, size-len(b))...), nil
}

// symBase64Decode 兼容 Std/RawStd/URL/RawURL 四种 Base64 变体
func symBase64Decode(s string) ([]byte, error) {
	for _, enc := range []*base64.Encoding{
		base64.StdEncoding, base64.RawStdEncoding,
		base64.URLEncoding, base64.RawURLEncoding,
	} {
		if b, err := enc.DecodeString(s); err == nil {
			return b, nil
		}
	}
	return nil, errors.New("无法解码")
}

// symEncodeCipher 按输出格式编码密文
func symEncodeCipher(data []byte, encoding string) string {
	if strings.EqualFold(strings.TrimSpace(encoding), "hex") {
		return hex.EncodeToString(data)
	}
	return base64.StdEncoding.EncodeToString(data)
}

// symDecodeCipher 解码输入的密文
func symDecodeCipher(s string, encoding string) ([]byte, error) {
	if strings.EqualFold(strings.TrimSpace(encoding), "hex") {
		return hex.DecodeString(strings.TrimSpace(s))
	}
	return symBase64Decode(strings.TrimSpace(s))
}

// symECB 手工实现 ECB（标准库未提供），分组模式带 PKCS7 填充
func symECB(block cipher.Block, data []byte, encrypt bool, outEncoding string) (string, error) {
	bs := block.BlockSize()
	if encrypt {
		data = symPkcs7Pad(data, bs)
		out := make([]byte, len(data))
		for i := 0; i < len(data); i += bs {
			block.Encrypt(out[i:i+bs], data[i:i+bs])
		}
		return symEncodeCipher(out, outEncoding), nil
	}

	data, err := symDecodeCipher(string(data), outEncoding)
	if err != nil {
		return "", errors.New("密文解码错误: " + err.Error())
	}
	if len(data) == 0 || len(data)%bs != 0 {
		return "", errors.New("密文长度不是分组长度的整数倍")
	}
	out := make([]byte, len(data))
	for i := 0; i < len(data); i += bs {
		block.Decrypt(out[i:i+bs], data[i:i+bs])
	}
	out, err = symPkcs7Unpad(out, bs)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// symCBC CBC 模式，PKCS7 填充
func symCBC(block cipher.Block, iv []byte, data []byte, encrypt bool, outEncoding string) (string, error) {
	bs := block.BlockSize()
	if encrypt {
		data = symPkcs7Pad(data, bs)
		out := make([]byte, len(data))
		cipher.NewCBCEncrypter(block, iv).CryptBlocks(out, data)
		return symEncodeCipher(out, outEncoding), nil
	}

	data, err := symDecodeCipher(string(data), outEncoding)
	if err != nil {
		return "", errors.New("密文解码错误: " + err.Error())
	}
	if len(data) == 0 || len(data)%bs != 0 {
		return "", errors.New("密文长度不是分组长度的整数倍")
	}
	out := make([]byte, len(data))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(out, data)
	out, err = symPkcs7Unpad(out, bs)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// symGCM GCM 模式（认证加密，密文末尾附带 16 字节认证标签）
func symGCM(block cipher.Block, nonce []byte, data []byte, encrypt bool, outEncoding string) (string, error) {
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if encrypt {
		return symEncodeCipher(gcm.Seal(nil, nonce, data, nil), outEncoding), nil
	}

	data, err = symDecodeCipher(string(data), outEncoding)
	if err != nil {
		return "", errors.New("密文解码错误: " + err.Error())
	}
	out, err := gcm.Open(nil, nonce, data, nil)
	if err != nil {
		return "", errors.New("GCM解密失败（密文或密钥/Nonce错误）")
	}
	return string(out), nil
}

// symCFB CFB 模式
func symCFB(block cipher.Block, iv []byte, data []byte, encrypt bool, outEncoding string) (string, error) {
	if encrypt {
		out := make([]byte, len(data))
		cipher.NewCFBEncrypter(block, iv).XORKeyStream(out, data)
		return symEncodeCipher(out, outEncoding), nil
	}

	data, err := symDecodeCipher(string(data), outEncoding)
	if err != nil {
		return "", errors.New("密文解码错误: " + err.Error())
	}
	out := make([]byte, len(data))
	cipher.NewCFBDecrypter(block, iv).XORKeyStream(out, data)
	return string(out), nil
}

// symStream 流模式（CTR/OFB，加解密同一条路径）
func symStream(stream cipher.Stream, data []byte, encrypt bool, outEncoding string) (string, error) {
	if !encrypt {
		var err error
		data, err = symDecodeCipher(string(data), outEncoding)
		if err != nil {
			return "", errors.New("密文解码错误: " + err.Error())
		}
	}
	out := make([]byte, len(data))
	stream.XORKeyStream(out, data)
	if encrypt {
		return symEncodeCipher(out, outEncoding), nil
	}
	return string(out), nil
}

// symPkcs7Pad PKCS7 填充
func symPkcs7Pad(data []byte, blockSize int) []byte {
	pad := blockSize - len(data)%blockSize
	return append(data, bytes.Repeat([]byte{byte(pad)}, pad)...)
}

// symPkcs7Unpad 去除 PKCS7 填充
func symPkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("解密失败：数据为空")
	}
	pad := int(data[len(data)-1])
	if pad == 0 || pad > blockSize || pad > len(data) {
		return nil, errors.New("解密失败：填充数据无效，请检查密钥是否正确")
	}
	for _, b := range data[len(data)-pad:] {
		if int(b) != pad {
			return nil, errors.New("解密失败：填充数据无效，请检查密钥是否正确")
		}
	}
	return data[:len(data)-pad], nil
}

// symRandomText 生成随机密钥/IV 文本：text 生成等长可打印串，其余生成随机字节后编码
func symRandomText(size int, encoding string) string {
	if size <= 0 {
		return ""
	}
	if encoding == "text" {
		out := make([]byte, size)
		max := big.NewInt(int64(len(symRandCharset)))
		for i := range out {
			n, _ := rand.Int(rand.Reader, max)
			out[i] = symRandCharset[n.Int64()]
		}
		return string(out)
	}

	raw := make([]byte, size)
	rand.Read(raw)
	if encoding == "hex" {
		return hex.EncodeToString(raw)
	}
	return base64.StdEncoding.EncodeToString(raw)
}
