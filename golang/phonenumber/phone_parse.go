package phonenumber

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"github.com/wyzzgzhdcxy/wcj-go-common/core"
)

// 号段信息结构
type PrefixInfo struct {
	Prefix   string `json:"prefix"`
	Province string `json:"province"`
	City     string `json:"city"`
	ISP      string `json:"isp"`
}

// 省份信息结构
type ProvinceInfo struct {
	Code   string `json:"code"`
	Region string `json:"region"`
}

// 解析结果结构
type PhoneInfo struct {
	Number       string `json:"number"`
	Prefix       string `json:"prefix"`
	Province     string `json:"province"`
	City         string `json:"city"`
	ISP          string `json:"isp"`
	ProvinceCode string `json:"province_code"`
	Region       string `json:"region"`
	CardType     string `json:"card_type"`
}

var (
	prefixConfig   map[string][]PrefixInfo
	provinceConfig map[string]ProvinceInfo
)

func Init() {
	loadConfigs()
}

func loadConfigs() {
	// 加载号段配置
	prefixData, err := os.ReadFile(core.GetTempDir() + "/app_data/phone_prefixes.json")
	if err != nil {
		panic(fmt.Sprintf("加载号段配置失败: %v", err))
	}
	if err := json.Unmarshal(prefixData, &prefixConfig); err != nil {
		panic(fmt.Sprintf("解析号段配置失败: %v", err))
	}

	// 加载省份配置
	provinceData, err := os.ReadFile(core.GetTempDir() + "/app_data/province_codes.json")
	if err != nil {
		panic(fmt.Sprintf("加载省份配置失败: %v", err))
	}
	if err := json.Unmarshal(provinceData, &provinceConfig); err != nil {
		panic(fmt.Sprintf("解析省份配置失败: %v", err))
	}
}

// Parse 解析手机号信息
func Parse(phone string) (*PhoneInfo, error) {
	if len(phone) != 11 {
		return nil, fmt.Errorf("无效的手机号长度")
	}

	// 查找最匹配的号段
	var bestMatch *PrefixInfo
	for _, prefixes := range prefixConfig {
		for _, p := range prefixes {
			if strings.HasPrefix(phone, p.Prefix) {
				// 选择最长匹配的号段(如1349优先于134)
				if bestMatch == nil || len(p.Prefix) > len(bestMatch.Prefix) {
					bestMatch = &p
				}
			}
		}
	}

	if bestMatch == nil {
		return nil, fmt.Errorf("无法识别的手机号号段")
	}

	// 获取省份信息
	provinceInfo, ok := provinceConfig[bestMatch.Province]
	if !ok {
		provinceInfo = ProvinceInfo{Code: "UNKNOWN", Region: "未知"}
	}

	// 判断卡类型
	cardType := "实体卡"
	if strings.HasPrefix(phone, "170") || strings.HasPrefix(phone, "171") {
		cardType = "虚拟运营商"
	}

	return &PhoneInfo{
		Number:       phone,
		Prefix:       bestMatch.Prefix,
		Province:     bestMatch.Province,
		City:         bestMatch.City,
		ISP:          bestMatch.ISP,
		ProvinceCode: provinceInfo.Code,
		Region:       provinceInfo.Region,
		CardType:     cardType,
	}, nil
}
