package app

import (
	"encoding/json"
	"fmt"

	"wcj-go-text/golang/idcard"
	"wcj-go-text/golang/phonenumber"
	"wcj-go-text/golang/utils"
)

// GenIdNo 生成身份证号
func (a *App) GenIdNo(num int) []string {
	return idcard.Create18IdCardNo(num)
}

// ParseIdNo 解析身份证号
func (a *App) ParseIdNo(idNo string) string {
	ok, info := idcard.Parse18IdCardNoInfo(idNo)
	if !ok {
		return fmt.Sprintf("身份证号无效: %s", idNo)
	}
	data, _ := json.Marshal(info)
	return string(data)
}

// GenerateBankCardNo 生成银行卡号
func (a *App) GenerateBankCardNo(count int) []string {
	return idcard.GenerateBankCardNo(count)
}

// ParseBankCard 解析银行卡号
func (a *App) ParseBankCard(cardNo string) string {
	info, err := idcard.ParseBankCard(cardNo)
	if err != nil {
		return fmt.Sprintf("银行卡号无效: %v", err)
	}
	data, _ := json.Marshal(info)
	return string(data)
}

// ParsePhone 解析手机号
func (a *App) ParsePhone(phone string) string {
	info, err := phonenumber.Parse(phone)
	if err != nil {
		return fmt.Sprintf("手机号无效: %v", err)
	}
	data, _ := json.Marshal(info)
	return string(data)
}

// GeneratePhone 生成虚拟手机号
func (a *App) GeneratePhone(count int) []string {
	var phones []string
	for i := 0; i < count; i++ {
		phone, _ := utils.Generate(utils.Options{})
		phones = append(phones, phone)
	}
	return phones
}

// ValidateIdCard 校验身份证号
func (a *App) ValidateIdCard(idNo string) bool {
	return idcard.Validate18IdCardNo(idNo)
}
