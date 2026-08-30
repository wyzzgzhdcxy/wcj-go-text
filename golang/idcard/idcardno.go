package idcard

import (
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
	"github.com/wyzzgzhdcxy/wcj-go-common/core"
)

// 权重
var idNoWeightArray = []int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}

// 身份证号码
var idNoCheckCode = "10X98765432"

// 身份证号码正则匹配
var idCardNoRegexpPattern = "^([1-9]\\d{7}((0\\d)|(1[0-2]))(([012]\\d)|3[0-1])\\d{3})|([1-9]\\d{5}[1-9]\\d{3}((0\\d)|(1[0-2]))(([012]\\d)|3[0-1])((\\d{4})|\\d{3}[Xx]))$"

// IdCardNoInfo 身份证信息
type IdCardNoInfo struct {
	IdCardNo    string // 身份证号码
	AreaCode    string // 地区编号
	AreaName    string // 地区名称
	BirthDayYMD string // 年月日生日，20060102
	Age         int    // 年龄
	Sex         int    // 性别，女0，男1
}

// Validate18IdCardNo 校验18位身份证号码有效性
func Validate18IdCardNo(idNo string) bool {
	if len(idNo) != 18 {
		return false
	}
	isMatch, _ := regexp.MatchString(idCardNoRegexpPattern, idNo)
	if !isMatch {
		return false
	}
	return getCheckDigit(idNo) == idNo[17:18]
}

var areaCodes []string

func GetAreaCodeList() []string {
	if len(areaCodes) == 0 {
		areaPath := core.GetTempDir() + "/app_data/area.txt"
		areaCodes = *core.ReadLines(areaPath)
	}
	return areaCodes
}

// AutoCreate18IdCardNo 自动生成18位身份证号码
func AutoCreate18IdCardNo(ran *rand.Rand) string {
	idNo := ""
	// 随机数种子
	// 6位地区编码
	idNo += GetAreaCodeList()[ran.Intn(len(GetAreaCodeList()))]
	// 8位年月日生日
	idNo += RandBirthDay(ran).Format("20060102")
	// 2位顺序码
	idNo += strconv.Itoa(ran.Intn(9)+1) + strconv.Itoa(ran.Intn(10))
	// 1位性别，女双数，男单数
	idNo += strconv.Itoa(ran.Intn(10))
	// 1位校验位
	idNo += computerCheckDigit(idNo)
	fmt.Println("---" + idNo)
	return idNo
}

func Create18IdCardNo(num int) []string {
	var idNoList []string
	ran := rand.New(rand.NewSource(time.Now().Unix()))
	for i := 0; i < num; i++ {
		idNoList = append(idNoList, AutoCreate18IdCardNo(ran))
	}
	return idNoList
}

// 获取18位身份证号码信息
func Parse18IdCardNoInfo(idNo string) (bool, *IdCardNoInfo) {
	isIdCardNo := Validate18IdCardNo(idNo)
	if !isIdCardNo {
		return false, nil
	}
	return true, &IdCardNoInfo{
		IdCardNo:    idNo,
		AreaCode:    getAreaCode(idNo),
		AreaName:    getAreaName(idNo),
		BirthDayYMD: getBirthDayYMD(idNo),
		Age:         getAge(idNo),
		Sex:         getSex(idNo),
	}
}

// 获取校验位
func getCheckDigit(idNo string) string {
	data := idNo[0:17]
	s := 0
	for i, _ := range data {
		n, _ := strconv.Atoi(string(data[i]))
		s += n * idNoWeightArray[i]
	}
	y := s % 11
	return idNoCheckCode[y : y+1]
}

// 生成校验位
func computerCheckDigit(idNo string) string {
	checkSum := 0
	for i := 0; i < 17; i++ {
		n, _ := strconv.Atoi(string(idNo[i]))
		checkSum += ((1 << uint(17-i)) % 11) * n
	}
	checkDigit := (12 - (checkSum % 11)) % 11
	if checkDigit >= 10 {
		return "X"
	} else {
		return strconv.Itoa(checkDigit)
	}
}

func getAreaCode(idNo string) string {
	return idNo[0:6]
}

var areaCodeNameMap map[string]string

func GetAreaCodeNameMap() map[string]string {
	if len(areaCodeNameMap) == 0 {
		areaCodeNameMap = make(map[string]string)
		areaPath := core.GetTempDir() + "/app_data/area_name.txt"
		areaCodes = *core.ReadLines(areaPath)
		for _, v := range areaCodes {
			arr := strings.Split(v, ",")
			if len(arr) != 2 {
				log.Printf("GetAreaCodeNameMap生成codenameMap 出现错误！")
				continue
			}
			areaCodeNameMap[arr[0]] = arr[1]
		}
	}
	return areaCodeNameMap
}

func getAreaName(idNo string) string {
	return GetAreaCodeNameMap()[idNo[0:6]]
}

func getBirthDayYMD(idNo string) string {
	return idNo[6:14]
}

func getAge(idNo string) int {
	return GetAgeByBirthDayYMD(idNo[6:14])
}

func getSex(idNo string) int {
	sex, _ := strconv.Atoi(idNo[16:17])
	return sex % 2
}
