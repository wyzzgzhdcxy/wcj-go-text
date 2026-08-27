package app

import (
	"strings"

	"wcj-go-text/golang/myText"
)

func splitLines(text string) []string {
	if text == "" {
		return nil
	}
	lines := strings.Split(text, "\n")
	result := make([]string, 0, len(lines))
	for _, l := range lines {
		l = strings.TrimRight(l, "\r")
		if strings.TrimSpace(l) != "" {
			result = append(result, l)
		}
	}
	return result
}

func joinLines(lines []string) string {
	return strings.Join(lines, "\n")
}

// TextUnion 文本并集
func (a *App) TextUnion(text1, text2 string) string {
	return joinLines(myText.Union(splitLines(text1), splitLines(text2)))
}

// TextIntersection 文本交集
func (a *App) TextIntersection(text1, text2 string) string {
	return joinLines(myText.Intersection(splitLines(text1), splitLines(text2)))
}

// TextDifference 文本差集
func (a *App) TextDifference(text1, text2 string) string {
	return joinLines(myText.Difference(splitLines(text1), splitLines(text2)))
}

// TextRemoveDuplicate 文本去重
func (a *App) TextRemoveDuplicate(text string) string {
	seen := make(map[string]bool)
	var result []string
	for _, l := range splitLines(text) {
		if !seen[l] {
			seen[l] = true
			result = append(result, l)
		}
	}
	return joinLines(result)
}

// TextReplace 文本替换
func (a *App) TextReplace(text, oldStr, newStr string) string {
	return strings.ReplaceAll(text, oldStr, newStr)
}

// TplReplace 模板替换
func (a *App) TplReplace(tpl string, data [][]string) myText.TplResult {
	return myText.TplReplace(myText.TplData{Tpl: tpl, Data: data})
}

// GenGoCodeByJsonStr JSON 转 Go 结构体
func (a *App) GenGoCodeByJsonStr(jsonStr string) string {
	result, err := myText.GenerateStruct(jsonStr, "MyStruct")
	if err != nil {
		return "Error: " + err.Error()
	}
	return result
}

// GenJavaCodeByJsonStr JSON 转 Java 类
func (a *App) GenJavaCodeByJsonStr(jsonStr string) string {
	result, err := myText.GenerateJavaClass(jsonStr, "MyClass")
	if err != nil {
		return "Error: " + err.Error()
	}
	return result
}

// JavaCodeToJson Java 代码转 JSON（简化实现）
func (a *App) JavaCodeToJson(javaCode string) string {
	return "{\"note\":\"Java code parsing not yet implemented\"}"
}

// GenJsonStringSplitChar JSON 分割符列表
func (a *App) GenJsonStringSplitChar() []string {
	return []string{",", "\t", "|", ";", " "}
}

// Text2JsonCSV 文本转 JSON（CSV 格式）
func (a *App) Text2JsonCSV(content string, splitChar string) []map[string]string {
	return myText.Text2Json(content, splitChar)
}

// ReadDemoFile 读取 demo 文件
func (a *App) ReadDemoFile(path string) string {
	content, _ := a.ReadAssetFile("/demo/" + path)
	return content
}

// MoveFilesReq 移动文件请求
type MoveFilesReq struct {
	SrcDir     string   `json:"srcDir"`
	DstDir     string   `json:"dstDir"`
	FileNames  []string `json:"fileNames"`
}

// MoveFilesRes 移动文件结果
type MoveFilesRes struct {
	Success int      `json:"success"`
	Failed  int      `json:"failed"`
	Errors  []string `json:"errors"`
}

// FileSplitReq 文件分割请求
type FileSplitReq struct {
	FilePath  string `json:"filePath"`
	LineCount int    `json:"lineCount"`
}

// FileSplitRes 文件分割结果
type FileSplitRes struct {
	Success bool     `json:"success"`
	Message string   `json:"message"`
	Files   []string `json:"files"`
}

// FileMergeReq 文件合并请求
type FileMergeReq struct {
	FilePaths []string `json:"filePaths"`
	Output    string   `json:"output"`
}

// FileMergeRes 文件合并结果
type FileMergeRes struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
