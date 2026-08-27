package mysqlUtils

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"
	"unicode"
	"github.com/wyzzgzhdcxy/wcj-go-common/core"

	_ "github.com/go-sql-driver/mysql"
)

// 定义模板
const htmlTemplate = `<!DOCTYPE html>
<html>
<head>
    <title>数据库文档 - {{.DatabaseName}}</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        h1 { color: #333; }
        h2 { color: #444; margin-top: 30px; border-bottom: 1px solid #eee; padding-bottom: 5px; }
        table { width: 100%; border-collapse: collapse; margin-bottom: 20px; }
        th { background-color: #f5f5f5; text-align: left; padding: 8px; }
        td { padding: 8px; border-bottom: 1px solid #ddd; }
        tr:nth-child(even) { background-color: #f9f9f9; }
        .timestamp { color: #999; font-size: 0.9em; margin-bottom: 20px; }
    </style>
</head>
<body>
    <h1>数据库文档 - {{.DatabaseName}}</h1>
    <div class="timestamp">生成时间: {{.Timestamp}}</div>
    
    {{range .Tables}}
    <h2>表: {{.Name}}</h2>
    <table>
        <thead>
            <tr>
                <th>字段名</th>
                <th>类型</th>
                <th>必填</th>
                <th>键</th>
                <th>默认值</th>
                <th>额外</th>
                <th>注释</th>
            </tr>
        </thead>
        <tbody>
            {{range .Columns}}
            <tr>
                <td>{{.Name}}</td>
                <td>{{.Type}}</td>
                <td>{{.Required}}</td>
                <td>{{.Key}}</td>
                <td>{{.Default}}</td>
                <td>{{.Extra}}</td>
                <td>{{.Comment}}</td>
            </tr>
            {{end}}
        </tbody>
    </table>
    {{end}}
</body>
</html>`

const clazzTemplate = `

    {{range .Tables}}
package com.itiaoling.ef.lite.cw.sku.web.repo.model;

import com.baomidou.mybatisplus.annotation.TableField;
import com.baomidou.mybatisplus.annotation.TableName;
import com.itiaoling.spring.mytabisplus.InputIdBaseEntity;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.math.BigDecimal;

/**
 * @author: wtools Generator
 * @date {{$.Timestamp}}
 */
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
@TableName("{{.Name}}")
public class {{.ClazzName}} extends InputIdBaseEntity {
{{range .Columns}}
     // {{.Type}} 
     @TableField("{{.Name}}")
     private {{.JavaType}} {{.FieldName}};
    {{end}}

}
{{end}}
`

// 数据库表结构
type Table struct {
	Name      string
	ClazzName string
	Columns   []Column
}

// 表列结构
type Column struct {
	Name      string
	FieldName string
	Type      string
	JavaType  string
	Nullable  string
	Required  string
	Key       string
	Default   string
	Extra     string
	Comment   string
}

// 文档数据
type DocData struct {
	DatabaseName string
	Timestamp    string
	Tables       []Table
}

func CLoseRows(rows *sql.Rows) {
	err := rows.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
}

func GetDbData(connUrl string) *DocData {
	// 连接数据库
	db, err := sql.Open("mysql", connUrl)
	if err != nil {
		panic(err.Error())
	}
	defer CloseDB(db)

	// 获取数据库名称
	var dbName string
	err = db.QueryRow("SELECT DATABASE()").Scan(&dbName)
	if err != nil {
		panic(err.Error())
	}

	// 获取表列表
	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		panic(err.Error())
	}
	defer CLoseRows(rows)

	var tables []Table

	// 遍历所有表
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			panic(err.Error())
		}

		// 获取表结构
		query := fmt.Sprintf(`
			SELECT 
				COLUMN_NAME, 
				COLUMN_TYPE, 
				IS_NULLABLE, 
				COLUMN_KEY, 
				IFNULL(COLUMN_DEFAULT, 'NULL'), 
				EXTRA,
				IFNULL(COLUMN_COMMENT, '')
			FROM 
				INFORMATION_SCHEMA.COLUMNS 
			WHERE 
				TABLE_NAME = '%s' 
				AND TABLE_SCHEMA = '%s'
			ORDER BY ORDINAL_POSITION`, tableName, dbName)

		descRows, err := db.Query(query)
		if err != nil {
			panic(err.Error())
		}

		var columns []Column
		for descRows.Next() {
			var col Column
			if err := descRows.Scan(
				&col.Name,
				&col.Type,
				&col.Nullable,
				&col.Key,
				&col.Default,
				&col.Extra,
				&col.Comment,
			); err != nil {
				panic(err.Error())
			}

			if strings.HasPrefix(col.Type, "decimal") || strings.HasPrefix(col.Type, "number") {
				col.JavaType = "BigDecimal"
			} else if strings.HasPrefix(col.Type, "datetime") || strings.HasPrefix(col.Type, "timestamp") {
				col.JavaType = "Date"
			} else {
				col.JavaType = "String"
			}

			col.FieldName = UnderScoreToCamelCase(col.Name)
			if col.Nullable == "NO" {
				col.Required = "Y"
			} else {
				col.Required = "N"
			}
			columns = append(columns, col)
		}
		descRows.Close()

		tables = append(tables, Table{
			Name:      tableName,
			ClazzName: UnderScoreToPascalCase(tableName),
			Columns:   columns,
		})
	}

	// 准备文档数据
	data := DocData{
		DatabaseName: dbName,
		Timestamp:    time.Now().Format("2006-01-02 15:04:05"),
		Tables:       tables,
	}

	return &data
}

func GenDoc(connUrl string) {
	// 准备文档数据
	data := GetDbData(connUrl)
	// 创建模板
	tmpl, err := template.New("dbdoc").Parse(htmlTemplate)
	if err != nil {
		panic(err.Error())
	}

	// 创建输出文件
	file, err := os.Create("database_documentation.html")
	if err != nil {
		panic(err.Error())
	}
	defer core.Close(file)

	// 执行模板
	if err := tmpl.Execute(file, data); err != nil {
		panic(err.Error())
	}
	fmt.Println("数据库文档已生成: database_documentation.html")
}

func GenEntityCode(connUrl string) string {
	// 准备文档数据
	data := GetDbData(connUrl)

	// 创建模板
	tmpl, err := template.New("entityCode").Parse(clazzTemplate)
	if err != nil {
		panic(err.Error())
	}

	// 执行模板

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		panic(err.Error())
	}

	result := buf.String()
	println(result) // 输出: Hello, World!
	return result
}

// UnderScoreToCamelCase 下划线转驼峰
func UnderScoreToCamelCase(s string) string {
	words := strings.Split(s, "_")
	for i := 1; i < len(words); i++ {
		words[i] = strings.Title(words[i])
	}
	return strings.Join(words, "")
}

func UnderScoreToPascalCase(s string) string {
	var result strings.Builder
	nextUpper := true

	for _, r := range s {
		if r == '_' {
			nextUpper = true
			continue
		}

		if nextUpper {
			result.WriteRune(unicode.ToUpper(r))
			nextUpper = false
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}
