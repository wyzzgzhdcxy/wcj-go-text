package myText

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"github.com/wyzzgzhdcxy/wcj-go-common/core"
)

type TplData struct {
	Tpl  string
	Data [][]string
}

type TplResult struct {
	Result string
	Line   int
}

func TplReplace(tplData TplData) TplResult {
	tplStr := strings.ReplaceAll(tplData.Tpl, "{{", "{{index . ")
	var wg sync.WaitGroup
	size := len(tplData.Data)
	page := size/100 + 1
	var tmp = make([][]string, page)
	wg.Add(page)
	for i := 0; i < size; i = i + 100 {
		go func(start int) {
			defer wg.Done()
			end := start + 100
			if end > size {
				end = size
			}
			tmpArr := tplData.Data[start:end]
			tmp[start/100] = convertData(&tmpArr, tplStr)
		}(i)
	}
	wg.Wait()
	strResult := core.DimensionalToStringResult(&tmp)
	tplResult := TplResult{
		Result: strResult.Result,
		Line:   strResult.Count,
	}
	return tplResult
}

func convertData(data *[][]string, tplStr string) []string {
	var result []string
	for _, arr := range *data {
		buffer := convert(tplStr, &arr)
		result = append(result, buffer.String())
	}
	return result
}

func convert(tplStr string, arr *[]string) bytes.Buffer {
	var buffer bytes.Buffer
	tmpl, err := template.New("myTpl").Parse(tplStr)
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(&buffer, *arr)
	return buffer
}

// Text2Json ParseCSVLikeData 解析类似CSV格式的数据
// jsonStrReq 是一个包含内容和分隔符的结构体
func Text2Json(content string, splitChar string) []map[string]string {
	// 处理不同换行符
	lines1 := strings.Split(content, "\r\n")
	lines2 := strings.Split(content, "\n")

	var lines []string
	if len(lines2) > len(lines1) {
		lines = lines2
	} else {
		lines = lines1
	}

	var list []map[string]string

	if len(lines) == 0 {
		return list
	}

	// 获取表头
	headers := strings.Split(lines[0], splitChar)
	fmt.Println(headers)
	fmt.Println(splitChar)
	fmt.Printf(strconv.Itoa(len(headers)))

	// 处理数据行
	for i := 1; i < len(lines); i++ {
		row := make(map[string]string)
		columns := strings.Split(lines[i], splitChar)

		for j := 0; j < len(columns); j++ {
			if j < len(headers) {
				row[headers[j]] = columns[j]
			}
		}

		list = append(list, row)
	}

	return list
}
