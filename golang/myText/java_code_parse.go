package myText

import (
	"encoding/json"
	"regexp"
	"strings"
)

// ParseJavaClassToJson 将 Java 类代码解析为 JSON 结构
func ParseJavaClassToJson(javaCode string) (string, error) {
	fields := extractJavaFields(javaCode)
	if len(fields) == 0 {
		return "{}", nil
	}

	jsonMap := make(map[string]interface{})
	for _, field := range fields {
		jsonMap[field.Name] = getDefaultValue(field.Type)
	}

	result, err := json.MarshalIndent(jsonMap, "", "  ")
	if err != nil {
		return "", err
	}
	return string(result), nil
}

// javaField Java 字段结构
type javaField struct {
	Type     string // 字段类型
	Name     string // 字段名
	FullDeco string // 完整声明
}

// extractJavaFields 从 Java 代码中提取所有字段
func extractJavaFields(javaCode string) []javaField {
	var fields []javaField

	// 移除注释
	code := removeComments(javaCode)

	// 匹配字段声明正则 (private/protected/public 字段)
	fieldPattern := regexp.MustCompile(`(?:private|protected|public)\s+([\w<>?\[\]]+)\s+(\w+)\s*;`)
	matches := fieldPattern.FindAllStringSubmatch(code, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			fields = append(fields, javaField{
				Type: match[1],
				Name: match[2],
			})
		}
	}

	return fields
}

// removeComments 移除 Java 代码中的注释
func removeComments(code string) string {
	// 移除单行注释 //
	code = regexp.MustCompile(`//.*?$`).ReplaceAllString(code, "")
	// 移除多行注释 /* */
	code = regexp.MustCompile(`/\*.*?\*/`).ReplaceAllString(code, "")
	return code
}

// getDefaultValue 根据 Java 类型返回默认值
func getDefaultValue(javaType string) interface{} {
	javaType = strings.TrimPrefix(javaType, "java.lang.")

	switch javaType {
	case "byte":
		return int8(0)
	case "short":
		return int16(0)
	case "int", "Integer":
		return 0
	case "long":
		return int64(0)
	case "float":
		return float32(0)
	case "double":
		return float64(0)
	case "boolean", "Boolean":
		return false
	case "char", "Character":
		return ""
	case "String":
		return ""
	default:
		// 处理数组类型
		if strings.HasSuffix(javaType, "[]") {
			return []interface{}{}
		}
		// 处理泛型如 List<String>, Map<String, Object>
		if strings.HasPrefix(javaType, "List<") || strings.HasPrefix(javaType, "ArrayList<") {
			return []interface{}{}
		}
		if strings.HasPrefix(javaType, "Map<") || strings.HasPrefix(javaType, "HashMap<") {
			return map[string]interface{}{}
		}
		return nil
	}
}
