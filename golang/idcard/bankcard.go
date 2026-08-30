package idcard

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// 银行BIN号范围定义
var bankBinRanges = map[string][]struct {
	Min, Max int64
	Name     string
	CardType string // DC:借记卡, CC:信用卡, SCC:准贷记卡, PC:预付费卡
}{
	"中国工商银行": {
		{Min: 620058, Max: 620058, Name: "中国工商银行", CardType: "PC"}, // 工行预付卡
		{Min: 622200, Max: 622599, Name: "中国工商银行", CardType: "DC"},
		{Min: 955880, Max: 955881, Name: "中国工商银行", CardType: "DC"},
	},
	"中国建设银行": {
		{Min: 436700, Max: 436799, Name: "中国建设银行", CardType: "DC"},
		{Min: 622700, Max: 622799, Name: "中国建设银行", CardType: "DC"},
	},
	"中国银行": {
		{Min: 621660, Max: 621669, Name: "中国银行", CardType: "DC"},
		{Min: 456351, Max: 456351, Name: "中国银行", CardType: "CC"},
	},
	"中国农业银行": {
		{Min: 622820, Max: 622829, Name: "中国农业银行", CardType: "DC"},
		{Min: 955950, Max: 955959, Name: "中国农业银行", CardType: "DC"},
	},
	// 更多银行可以继续添加...
}

// BankCardInfo 银行卡信息
type BankCardInfo struct {
	BankName  string // 银行名称
	CardType  string // 卡类型: DC-借记卡, CC-信用卡, SCC-准贷记卡, PC-预付费卡
	Valid     bool   // 是否有效卡号
	Formatted string // 格式化后的卡号
}

// ParseBankCard 解析银行卡号
func ParseBankCard(cardNo string) (*BankCardInfo, error) {
	// 清理卡号中的非数字字符
	cleaned := cleanCardNumber(cardNo)

	// 验证卡号长度
	if len(cleaned) < 12 || len(cleaned) > 19 {
		return nil, fmt.Errorf("银行卡号长度不正确")
	}

	// 验证卡号有效性(Luhn算法)
	isValid := validateLuhn(cleaned)

	// 获取银行信息
	bankName, cardType := getBankInfo(cleaned)

	// 格式化卡号
	formatted := formatCardNumber(cleaned)

	return &BankCardInfo{
		BankName:  bankName,
		CardType:  cardType,
		Valid:     isValid,
		Formatted: formatted,
	}, nil
}

// cleanCardNumber 清理卡号中的非数字字符
func cleanCardNumber(cardNo string) string {
	reg := regexp.MustCompile(`[^0-9]`)
	return reg.ReplaceAllString(cardNo, "")
}

// validateLuhn 使用Luhn算法验证卡号有效性
func validateLuhn(cardNo string) bool {
	sum := 0
	alternate := false

	for i := len(cardNo) - 1; i >= 0; i-- {
		num, err := strconv.Atoi(string(cardNo[i]))
		if err != nil {
			return false
		}

		if alternate {
			num *= 2
			if num > 9 {
				num = (num % 10) + 1
			}
		}

		sum += num
		alternate = !alternate
	}

	return sum%10 == 0
}

// getBankInfo 根据卡号获取银行信息
func getBankInfo(cardNo string) (string, string) {
	// 取前6位作为BIN号
	if len(cardNo) < 6 {
		return "未知", ""
	}

	bin, err := strconv.ParseInt(cardNo[:6], 10, 64)
	if err != nil {
		return "未知", ""
	}

	for bankName, ranges := range bankBinRanges {
		for _, r := range ranges {
			if bin >= r.Min && bin <= r.Max {
				return bankName, r.CardType
			}
		}
	}

	return "未知", ""
}

// formatCardNumber 格式化卡号显示
func formatCardNumber(cardNo string) string {
	// 每4位加一个空格
	var builder strings.Builder
	for i, r := range cardNo {
		if i > 0 && i%4 == 0 {
			builder.WriteString(" ")
		}
		builder.WriteRune(r)
	}
	return builder.String()
}

// GenerateBankCardNo 生成指定数量的银行卡号
func GenerateBankCardNo(count int) []string {
	var results []string
	for i := 0; i < count; i++ {
		// 随机选择一个银行和卡类型
		bankNames := make([]string, 0, len(bankBinRanges))
		for name := range bankBinRanges {
			bankNames = append(bankNames, name)
		}
		if len(bankNames) == 0 {
			continue
		}
		// 这里简化处理，生成一个符合Luhn算法的随机卡号
		cardNo := generateValidCardNumber()
		results = append(results, cardNo)
	}
	return results
}

// generateValidCardNumber 生成一个符合Luhn算法的随机卡号
func generateValidCardNumber() string {
	// 随机选择6位BIN
	bins := []string{"620058", "622200", "436700", "622700", "621660", "622820"}
	bin := bins[time.Now().UnixNano()%int64(len(bins))]
	// 生成9位随机数
	randStr := fmt.Sprintf("%09d", time.Now().UnixNano()%1000000000)
	partial := bin + randStr
	// 计算校验位
	checkDigit := calculateLuhnCheckDigit(partial)
	return partial + checkDigit
}

// calculateLuhnCheckDigit 计算Luhn校验位
func calculateLuhnCheckDigit(cardNo string) string {
	sum := 0
	alternate := true
	for i := len(cardNo) - 1; i >= 0; i-- {
		num, _ := strconv.Atoi(string(cardNo[i]))
		if alternate {
			num *= 2
			if num > 9 {
				num = (num % 10) + 1
			}
		}
		sum += num
		alternate = !alternate
	}
	checkDigit := (10 - (sum % 10)) % 10
	return strconv.Itoa(checkDigit)
}
