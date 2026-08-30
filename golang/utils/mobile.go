package utils

import (
	"fmt"
	"math/rand"
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

// 国内运营商号段（用作默认数据，避免未加载配置时生成号码崩溃）
var prefixConfig = PrefixConfig{
	Mobile:  []string{"134", "135", "136", "137", "138", "139", "147", "148", "150", "151", "152", "157", "158", "159", "178", "182", "183", "184", "187", "188", "195", "197", "198"},
	Unicom:  []string{"130", "131", "132", "145", "146", "155", "156", "166", "175", "176", "185", "186", "196"},
	Telecom: []string{"133", "149", "153", "173", "177", "180", "181", "189", "190", "191", "193", "199"},
	Virtual: []string{"162", "165", "167", "170", "171"},
}

// 生成手机号选项
type Options struct {
	Carrier string // 指定运营商
}

// Generate 生成随机手机号
func Generate(opt Options) (string, error) {
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
		allPrefixes := make([]string, 0, len(prefixConfig.Mobile)+len(prefixConfig.Unicom)+len(prefixConfig.Telecom))
		allPrefixes = append(allPrefixes, prefixConfig.Mobile...)
		allPrefixes = append(allPrefixes, prefixConfig.Unicom...)
		allPrefixes = append(allPrefixes, prefixConfig.Telecom...)
		prefixes = allPrefixes
	default:
		return "", fmt.Errorf("不支持的运营商类型: %s", opt.Carrier)
	}

	if len(prefixes) == 0 {
		return "", fmt.Errorf("运营商 %s 没有可用的号段", opt.Carrier)
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
