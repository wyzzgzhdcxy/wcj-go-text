package myText

import (
	"encoding/json"
	"fmt"
	"strings"
)

func GenerateJavaClass(jsonStr, className string) (string, error) {
	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return "", err
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("public class %s {\n", className))
	generateJavaFields(&builder, data, 1)
	builder.WriteString("}\n")
	return builder.String(), nil
}

func generateJavaFields(builder *strings.Builder, data interface{}, indent int) {
	_ = strings.Repeat("\t", indent) // indent保留以便后续扩展

	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			fieldName := toJavaFieldName(key)
			fieldType := getJavaType(value)

			builder.WriteString(fmt.Sprintf("\tprivate %s %s;\n", fieldType, fieldName))

			// 处理嵌套结构
			if nested, ok := value.(map[string]interface{}); ok {
				nestedClassName := toJavaClassName(key)
				builder.WriteString(fmt.Sprintf("\n\t// 内部类: %s\n", nestedClassName))
				builder.WriteString(fmt.Sprintf("\tpublic static class %s {\n", nestedClassName))
				generateJavaFields(builder, nested, 2)
				builder.WriteString("\t}\n")
			}

			// 处理数组类型
			if arr, ok := value.([]interface{}); ok && len(arr) > 0 {
				if nested, ok := arr[0].(map[string]interface{}); ok {
					nestedClassName := toJavaClassName(key)
					builder.WriteString(fmt.Sprintf("\n\t// 内部类: %s\n", nestedClassName))
					builder.WriteString(fmt.Sprintf("\tpublic static class %s {\n", nestedClassName))
					generateJavaFields(builder, nested, 2)
					builder.WriteString("\t}\n")
				}
			}
		}
	case []interface{}:
		if len(v) > 0 {
			generateJavaFields(builder, v[0], indent)
		}
	}
}

func getJavaType(value interface{}) string {
	switch v := value.(type) {
	case float64:
		if v == float64(int64(v)) {
			return "int"
		}
		return "double"
	case bool:
		return "boolean"
	case string:
		return "String"
	case []interface{}:
		if len(v) > 0 {
			elemType := getJavaType(v[0])
			return elemType + "[]"
		}
		return "Object[]"
	case map[string]interface{}:
		return "Object"
	case nil:
		return "Object"
	default:
		return "Object"
	}
}

// toJavaFieldName 将 JSON key 转换为 Java 字段名（驼峰命名）
func toJavaFieldName(s string) string {
	parts := strings.Split(s, "_")
	for i := range parts {
		if i == 0 {
			parts[i] = strings.ToLower(parts[i])
		} else {
			parts[i] = strings.ToUpper(string(parts[i][0])) + strings.ToLower(parts[i][1:])
		}
	}
	return strings.Join(parts, "")
}

// toJavaClassName 将 JSON key 转换为 Java 类名（首字母大写）
func toJavaClassName(s string) string {
	name := toJavaFieldName(s)
	if len(name) == 0 {
		return name
	}
	return strings.ToUpper(string(name[0])) + name[1:]
}
