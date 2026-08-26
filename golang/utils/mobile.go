package utils

import (
	"fmt"
	"math/rand"
	"time"
)

// 运营商类型
const (
	CarrierMobile  = "mobile"  // 中国移动
	CarrierUnicom  = "unicom"  // 中国联通
	CarrierTelecom = "telecom" // 中国电信
	CarrierVirtual = "virtual" // 虚拟运营商
)

// 号段配置结构
type PrefixConfig struct {
	Mobile  []string `json:"mobile"`
	Unicom  []string `json:"unicom"`
	Telecom []string `json:"telecom"`
	Virtual []string `json:"virtual"`
}

var prefixConfig PrefixConfig

// 生成手机号选项
type Options struct {
	Carrier string // 指定运营商
}

// Generate 生成随机手机号
func Generate(opt Options) (string, error) {
	rand.Seed(time.Now().UnixNano())

	var prefixes []string

	switch opt.Carrier {
	case CarrierMobile:
		prefixes = prefixConfig.Mobile
	case CarrierUnicom:
		prefixes = prefixConfig.Unicom
	case CarrierTelecom:
		prefixes = prefixConfig.Telecom
	case CarrierVirtual:
		prefixes = prefixConfig.Virtual
	case "": // 不指定运营商则随机选择
		allPrefixes := append(prefixConfig.Mobile, prefixConfig.Unicom...)
		allPrefixes = append(allPrefixes, prefixConfig.Telecom...)
		prefixes = allPrefixes
	default:
		return "", fmt.Errorf("不支持的运营商类型: %s", opt.Carrier)
	}

	// 随机选择号段
	prefix := prefixes[rand.Intn(len(prefixes))]

	// 生成后8位
	suffix := fmt.Sprintf("%08d", rand.Intn(100000000))

	return prefix + suffix, nil
}

// 验证手机号有效性
func Validate(number string) bool {
	if len(number) != 11 {
		return false
	}

	// 检查所有运营商号段
	checkPrefix := func(prefixes []string) bool {
		for _, prefix := range prefixes {
			if len(number) >= len(prefix) && number[:len(prefix)] == prefix {
				return true
			}
		}
		return false
	}

	return checkPrefix(prefixConfig.Mobile) ||
		checkPrefix(prefixConfig.Unicom) ||
		checkPrefix(prefixConfig.Telecom) ||
		checkPrefix(prefixConfig.Virtual)
}
