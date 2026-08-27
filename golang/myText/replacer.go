package myText

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
	"github.com/wyzzgzhdcxy/wcj-go-common/core"
)

type Replacer struct {
	Template string
	exps     []string
}

/*
*
从字符串中提取要替换的表达式
*/
func (rc *Replacer) collectExpression(data string) {
	reg1 := regexp.MustCompile(`\$\{[^\}]+\}`)
	if reg1 == nil {
		fmt.Println("模板解析异常")
	}
	//根据规则提取关键信息
	result1 := reg1.FindAllStringSubmatch(data, -1)
	//r := make([]string, 0)
	for _, v := range result1 {
		rc.exps = append(rc.exps, v[0][2:len(v[0])-1])
	}
}

func (rc *Replacer) RepalceText() string {
	rc.collectExpression(rc.Template)
	for _, flag := range rc.exps {
		rexArr := strings.Split(flag, ":")
		if len(rexArr) >= 2 {
			value := strings.Join(rexArr[1:], ":")
			switch rexArr[0] {
			case "str":
				size, _ := strconv.Atoi(value)
				rc.Template = strings.Replace(rc.Template, "${"+flag+"}", core.RandomInt(size), 1)
			case "file":
				rc.fileFlag(flag, value)
			}
		}
	}
	return rc.Template
}

// file标记 替换为指定的文件内容
func (rc *Replacer) fileFlag(flag string, value string) {
	text, err := ioutil.ReadFile(value)
	if err != nil {
		text = []byte("文件不存在：" + value)
	} else {
		rc.Template = strings.Replace(rc.Template, "${"+flag+"}", string(text), 1)
	}
}
