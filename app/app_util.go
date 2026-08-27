package app

import (
	"encoding/json"
	"fmt"
	"time"

	"wcj-go-text/golang/myNet"
	"wcj-go-text/golang/utils"
)

func jsonMarshal(v interface{}) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// TimeConvert 时间戳/日期格式转换
func (a *App) TimeConvert(input string, inputFormat string, outputFormat string) string {
	t, err := time.Parse(inputFormat, input)
	if err != nil {
		return fmt.Sprintf("解析错误: %v", err)
	}
	return t.Format(outputFormat)
}

// SalaryList 工资计算
func (a *App) SalaryList(salaryJSON string) string {
	var req utils.SalaryReq
	if err := json.Unmarshal([]byte(salaryJSON), &req); err != nil {
		return fmt.Sprintf("参数错误: %v", err)
	}
	summary := utils.SalaryList(req)
	data, _ := jsonMarshal(summary)
	return data
}

// NextTimeList cron 下次执行时间
func (a *App) NextTimeList(cronExpress string, startDate string, count int) []utils.CronRes {
	return utils.NextTimeList(cronExpress, startDate, count)
}

// IpParse IP 地址解析
func (a *App) IpParse(ip string) string {
	return utils.IpParse(ip)
}

// Ping IP 列表（基于 ICMP，需管理员权限）
func (a *App) Ping(ips []string) []string {
	var results []string
	for _, ip := range ips {
		if myNet.Ping(ip) {
			results = append(results, fmt.Sprintf("%s: 通", ip))
		} else {
			results = append(results, fmt.Sprintf("%s: 不通", ip))
		}
	}
	return results
}

// SalarySummaryVo 工资计算结果
type SalarySummaryVo = utils.SalarySummaryVo
