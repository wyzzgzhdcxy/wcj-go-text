package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"strings"
)

type SalaryReq struct {
	SalaryAmount       string
	ExemptionAmount    string
	SocialSecurityBase string
	ExtraSalary        string // 新增字段
	// 其他字段...
}

type SalaryVo struct {
	Month           string     // 月份
	Salary          *big.Float // 税前收入
	Yanglaobx       *big.Float // 养老保险
	Yilbx           *big.Float // 医疗保险
	Gjj             *big.Float // 公积金保险
	Shengyubx       *big.Float // 生育保险
	Shiyebx         *big.Float // 失业保险
	SocialBx        *big.Float // 社保合计
	Tax             *big.Float // 当月个税
	TaxAmount       *big.Float // 全年纳税额
	ExemptionAmount *big.Float // 免征额
	ReadyTax        *big.Float // 已缴税额
	RealSalary      *big.Float // 税后收入
}

type SalarySummaryVo struct {
	List            []SalaryVo
	TaxRuleList     interface{} // 税率规则表
	TotalSalary     *big.Float  // 新增：总工资
	SocialBxTotal   *big.Float  // 新增：社保公积金总额
	TotalRealSalary *big.Float  // 新增：实际到手工资总额
}

func SalaryList(request SalaryReq) SalarySummaryVo {
	salaryAmount := new(big.Float)
	salaryAmount.SetString(request.SalaryAmount)

	exemptionAmount := new(big.Float)
	exemptionAmount.SetString(request.ExemptionAmount)

	socialSecurityBase := new(big.Float)
	socialSecurityBase.SetString(request.SocialSecurityBase)

	// 每月的额外收入
	mapMonthExtraSalary := getMonthExtraSalaryMap(request)
	mapJSON, _ := json.Marshal(mapMonthExtraSalary)
	log.Printf("每月额外收入:%s", string(mapJSON))

	taxAmount := new(big.Float).SetInt64(0)
	readyTax := new(big.Float).SetInt64(0)
	var salaryVos []SalaryVo

	for i := 1; i <= 12; i++ {
		monthExtraSalary := mapMonthExtraSalary[i]
		var salaryVo SalaryVo

		monthStr := fmt.Sprintf("%02d月", i)
		salaryVo.Month = monthStr

		currentSalary := new(big.Float).Set(salaryAmount)
		if monthExtraSalary != nil {
			currentSalary.Add(currentSalary, monthExtraSalary)
		}
		salaryVo.Salary = currentSalary

		// 设置免税额
		salaryVo.ExemptionAmount = roundTo2(exemptionAmount)

		// 社保计算
		computerSocialBx(socialSecurityBase, &salaryVo)

		// 扣除社保，公积金的金额
		taxRealSalary := new(big.Float).Sub(salaryVo.Salary, salaryVo.SocialBx)

		if taxRealSalary.Cmp(salaryVo.ExemptionAmount) <= 0 {
			// 扣除社保后小于免税额不要交税
			salaryVo.Tax = new(big.Float).SetInt64(0)
			salaryVo.TaxAmount = roundTo2(taxAmount)
		} else {
			temp := new(big.Float).Sub(taxRealSalary, salaryVo.ExemptionAmount)
			taxAmount.Add(taxAmount, temp)

			salaryVo.TaxAmount = roundTo2(taxAmount)

			taxValue := getTax(taxAmount)
			taxValue.Sub(taxValue, readyTax)
			salaryVo.Tax = roundTo2(taxValue)

			readyTax.Add(readyTax, salaryVo.Tax)
		}

		salaryVo.ReadyTax = roundTo2(readyTax)

		salaryVo.RealSalary = roundTo2(new(big.Float).Sub(taxRealSalary, salaryVo.Tax))

		salaryVos = append(salaryVos, salaryVo)
	}

	salaryInfoVo := SalarySummaryVo{
		List: salaryVos,
	}
	summaryInfo(&salaryInfoVo)
	salaryInfoVo.TaxRuleList = taxRuleList // 假设taxRuleList已定义

	return salaryInfoVo
}

// 辅助函数
func getMonthExtraSalaryMap(request SalaryReq) map[int]*big.Float {
	result := make(map[int]*big.Float)

	if request.ExtraSalary != "" {
		// 分割字符串
		parts := strings.Split(request.ExtraSalary, ",")

		// 遍历分割后的字符串
		for i, str := range parts {
			month := i + 1 // 月份从1开始
			str = strings.TrimSpace(str)

			if str != "" {
				// 转换为big.Float
				amount := new(big.Float)
				_, success := amount.SetString(str)
				if success {
					result[month] = amount
				} else {
					log.Printf("无效的额外薪资格式: %s", str)
				}
			}
		}
	}

	return result
}

// 修复后的计算函数
func computerSocialBx(socialSecurityBase *big.Float, salaryVo *SalaryVo) {
	// 防御性检查（关键修复点1）
	if socialSecurityBase == nil {
		log.Println("社保基数不能为nil")
		return
	}
	if salaryVo == nil {
		log.Println("SalaryVo对象不能为nil")
		return
	}

	// 初始化所有金额字段（关键修复点2）
	if salaryVo.Yanglaobx == nil {
		salaryVo.Yanglaobx = new(big.Float)
	}
	if salaryVo.Yilbx == nil {
		salaryVo.Yilbx = new(big.Float)
	}
	if salaryVo.Shiyebx == nil {
		salaryVo.Shiyebx = new(big.Float)
	}
	if salaryVo.Shengyubx == nil {
		salaryVo.Shengyubx = new(big.Float)
	}
	if salaryVo.Gjj == nil {
		salaryVo.Gjj = new(big.Float)
	}
	if salaryVo.SocialBx == nil {
		salaryVo.SocialBx = new(big.Float)
	}

	// 定义费率
	rates := map[string]float64{
		"yanglao":   0.08,  // 养老
		"yiliao":    0.02,  // 医疗
		"shiye":     0.005, // 失业
		"gongjijin": 0.07,  // 公积金
	}

	// 计算各项金额（关键修复点3：使用新变量存储计算结果）
	yanglao := new(big.Float).Mul(socialSecurityBase, new(big.Float).SetFloat64(rates["yanglao"]))
	yiliao := new(big.Float).Mul(socialSecurityBase, new(big.Float).SetFloat64(rates["yiliao"]))
	shiye := new(big.Float).Mul(socialSecurityBase, new(big.Float).SetFloat64(rates["shiye"]))
	gongjijin := new(big.Float).Mul(socialSecurityBase, new(big.Float).SetFloat64(rates["gongjijin"]))

	// 设置精度（修复点4：保留2位小数用四舍五入，而不是 SetPrec 的二进制有效位）
	salaryVo.Yanglaobx = roundTo2(yanglao)
	salaryVo.Yilbx = roundTo2(yiliao)
	salaryVo.Shiyebx = roundTo2(shiye)
	salaryVo.Shengyubx = new(big.Float).SetFloat64(0) // 生育保险固定0
	salaryVo.Gjj = roundTo(gongjijin, 0)              // 公积金取整

	// 计算社保总额（修复点6：重新计算避免累积错误）
	socialBx := new(big.Float).Set(salaryVo.Yanglaobx)
	socialBx.Add(socialBx, salaryVo.Yilbx)
	socialBx.Add(socialBx, salaryVo.Shiyebx)
	socialBx.Add(socialBx, salaryVo.Shengyubx)
	socialBx.Add(socialBx, salaryVo.Gjj)
	salaryVo.SocialBx = roundTo2(socialBx)

	// 调试日志（生产环境可移除）
	log.Printf("计算结果: 养老=%.2f 医疗=%.2f 失业=%.2f 公积金=%.0f 社保总额=%.2f",
		salaryVo.Yanglaobx, salaryVo.Yilbx, salaryVo.Shiyebx, salaryVo.Gjj, salaryVo.SocialBx)
}

func getTax(taxAmount *big.Float) *big.Float {
	// 检查税额是否为负数
	if taxAmount.Cmp(big.NewFloat(0)) < 0 {
		return big.NewFloat(0)
	}

	// 遍历税率规则表
	for _, taxRule := range taxRuleList {
		end := big.NewFloat(taxRule.End)
		if taxAmount.Cmp(end) < 0 {
			// 计算税额：taxAmount * taxRate - deduct
			taxRate := big.NewFloat(taxRule.TaxRate)
			deduct := big.NewFloat(taxRule.Deduct)

			result := new(big.Float).Mul(taxAmount, taxRate)
			result.Sub(result, deduct)
			return result
		}
	}

	// 默认最高税率计算
	taxRate := big.NewFloat(0.45)
	deduct := big.NewFloat(181920)
	result := new(big.Float).Mul(taxAmount, taxRate)
	result.Sub(result, deduct)
	return result
}

type TaxRule struct {
	End     float64 // 区间上限
	TaxRate float64 // 税率
	Deduct  float64 // 速算扣除数
}

// 假设的税率规则表（需要根据实际业务设置）
var taxRuleList = []TaxRule{
	{36000, 0.03, 0},
	{144000, 0.1, 2520},
	{300000, 0.2, 16920},
	{420000, 0.25, 31920},
	{660000, 0.3, 52920},
	{960000, 0.35, 85920},
	// 超过960000的部分使用默认的0.45税率
}

func summaryInfo(salaryInfoVo *SalarySummaryVo) {
	// 初始化汇总值为0
	totalSalary := new(big.Float).SetFloat64(0)
	socialBxTotal := new(big.Float).SetFloat64(0)
	totalRealSalary := new(big.Float).SetFloat64(0)

	// 遍历所有月份的工资数据
	for _, salaryVo := range salaryInfoVo.List {
		// 累加各项金额
		totalSalary.Add(totalSalary, salaryVo.Salary)
		socialBxTotal.Add(socialBxTotal, salaryVo.SocialBx)
		totalRealSalary.Add(totalRealSalary, salaryVo.RealSalary)
	}

	// 更新到汇总对象
	salaryInfoVo.TotalSalary = roundTo2(totalSalary)
	salaryInfoVo.SocialBxTotal = roundTo2(socialBxTotal)
	salaryInfoVo.TotalRealSalary = roundTo2(totalRealSalary)
}

// roundTo 把金额四舍五入保留 decimals 位小数。
// 注意不能用 big.Float.SetPrec：它设置的是有效二进制位而非小数位数，
// SetPrec(2) 会把 5000 这类数值破坏成 2 位有效位的近似值。
func roundTo(f *big.Float, decimals int) *big.Float {
	if f == nil {
		return nil
	}
	pow := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil))
	scaled := new(big.Float).Mul(f, pow)
	if scaled.Sign() < 0 {
		scaled.Sub(scaled, big.NewFloat(0.5))
	} else {
		scaled.Add(scaled, big.NewFloat(0.5))
	}
	i, _ := scaled.Int(nil)
	result := new(big.Float).SetInt(i)
	return result.Quo(result, pow)
}

// roundTo2 四舍五入保留2位小数
func roundTo2(f *big.Float) *big.Float {
	return roundTo(f, 2)
}
