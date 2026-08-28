package myText

import (
	"encoding/json"
	"fmt"
	"strings"
)

func GenerateStruct(jsonStr, structName string) (string, error) {
	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return "", err
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("type %s struct {\n", structName))
	generateFields(&builder, data, 1)
	builder.WriteString("}\n")
	return builder.String(), nil
}

func generateFields(builder *strings.Builder, data interface{}, indent int) {
	indentStr := strings.Repeat("\t", indent)

	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			fieldName := toCamelCase(key)
			fieldType := getType(value)

			builder.WriteString(fmt.Sprintf("%s%s %s `json:\"%s\"`",
				indentStr, fieldName, fieldType, key))

			// 处理嵌套结构
			if nested, ok := value.(map[string]interface{}); ok {
				builder.WriteString(" {\n")
				generateFields(builder, nested, indent+1)
				builder.WriteString(fmt.Sprintf("%s}", indentStr))
			}

			builder.WriteString("\n")
		}
	}
}

func getType(value interface{}) string {
	switch v := value.(type) {
	case float64:
		if v == float64(int64(v)) {
			return "int"
		}
		return "float64"
	case bool:
		return "bool"
	case string:
		return "string"
	case []interface{}:
		if len(v) > 0 {
			return "[]" + getType(v[0])
		}
		return "[]interface{}"
	case map[string]interface{}:
		return "struct"
	case nil:
		return "interface{}"
	default:
		return "interface{}"
	}
}

func toCamelCase(s string) string {
	parts := strings.Split(s, "_")
	for i := range parts {
		if i > 0 {
			parts[i] = strings.Title(parts[i])
		}
	}
	return strings.Join(parts, "")
}
