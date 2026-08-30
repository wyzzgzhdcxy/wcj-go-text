package myText

import (
	"bytes"
	"fmt"
	"github.com/wyzzgzhdcxy/wcj-go-common/core"
	"strings"
	"sync"
	"text/template"
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
	tmpl, err := template.New("myTpl").Parse(tplStr)
	if err != nil {
		return TplResult{
			Result: fmt.Sprintf("模板解析失败: %v", err),
			Line:   0,
		}
	}
	var wg sync.WaitGroup
	size := len(tplData.Data)
	page := (size + 99) / 100
	var tmp = make([][]string, page)
	wg.Add(page)
	for i := 0; i < size; i += 100 {
		go func(start int) {
			defer wg.Done()
			end := start + 100
			if end > size {
				end = size
			}
			tmpArr := tplData.Data[start:end]
			tmp[start/100] = convertData(tmpl, &tmpArr)
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

func convertData(tmpl *template.Template, data *[][]string) []string {
	var result []string
	for _, arr := range *data {
		buffer := convert(tmpl, &arr)
		result = append(result, buffer.String())
	}
	return result
}

func convert(tmpl *template.Template, arr *[]string) bytes.Buffer {
	var buffer bytes.Buffer
	// text/template 的 Execute 支持并发调用
	_ = tmpl.Execute(&buffer, *arr)
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
