package main

import (
	"encoding/json"
	"fmt"
	"time"

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

// Ping IP 列表
func (a *App) Ping(ips []string) []string {
	var results []string
	for _, ip := range ips {
		results = append(results, fmt.Sprintf("%s: ping", ip))
	}
	return results
}

// SalarySummaryVo 工资计算结果
type SalarySummaryVo = utils.SalarySummaryVo
