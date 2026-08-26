package mysqlUtils

import (
	"database/sql"
	"fmt"
	"github.com/wyzzgzhdcxy/wcj-go-common/core"

	_ "github.com/go-sql-driver/mysql"
)

func DoQuery(db *sql.DB, sqlInfo string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := db.Query(sqlInfo, args...)
	if err != nil {
		return nil, err
	}
	columns, _ := rows.Columns()
	columnLength := len(columns)
	cache := make([]interface{}, columnLength) //临时存储每行数据
	for index, _ := range cache {              //为每一列初始化一个指针
		var a interface{}
		cache[index] = &a
	}
	var list []map[string]interface{} //返回的切片
	for rows.Next() {
		_ = rows.Scan(cache...)

		item := make(map[string]interface{})
		for i, data := range cache {
			item[columns[i]] = *data.(*interface{}) //取实际类型
		}
		list = append(list, item)
	}
	_ = rows.Close()
	return list, nil
}

func GetConn(url string) *sql.DB {
	db, err := sql.Open("mysql", url)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	return db
}

func CloseDB(db *sql.DB) {
	err := db.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
}

func AllTableCreateSql(url string) (error, string) {
	db, err := sql.Open("mysql", url)
	if err != nil {
		return err, ""
	}
	defer CloseDB(db)
	rows, err := db.Query("show tables")
	if err != nil {
		return err, ""
	}
	var tableNames []string
	//循环读取结果
	for rows.Next() {
		var tableName string
		rows.Scan(&tableName)
		tableNames = append(tableNames, tableName)
	}
	var ddlSql string
	for _, tableName := range tableNames {
		rows, err := db.Query("show create table " + tableName)
		//将每一行的结果都赋值到一个user对象中
		for rows.Next() {
			var tableName1 string
			var createSql string
			rows.Scan(&tableName1, &createSql)
			if err != nil {
				fmt.Println("rows fail")
			}
			ddlSql = ddlSql + createSql + ";\n"
		}
	}
	return nil, ddlSql
}

func GetTableData(connUrl string, tableName string) (error, [][]string) {
	db := GetConn(connUrl)
	defer CloseDB(db)
	rows, err := db.Query("select * from " + tableName)
	if err != nil {
		return err, nil
	}
	columns, _ := rows.Columns()
	columnLength := len(columns)

	var list [][]interface{} //返回的切片
	for rows.Next() {
		cache := rowsToArray(columnLength, rows)
		list = append(list, cache)
	}
	_ = rows.Close()

	resultArr := core.DimensionalInterface2String(list)
	return nil, resultArr
}

func rowsToArray(columnLength int, rows *sql.Rows) []interface{} {
	cache := make([]interface{}, columnLength) //临时存储每行数据
	for index, _ := range cache {              //为每一列初始化一个指针
		var a interface{}
		cache[index] = &a
	}
	_ = rows.Scan(cache...)
	return cache
}

func GetTableNames(connUrl string) (error, []string) {
	db := GetConn(connUrl)
	defer CloseDB(db)
	rows, err := db.Query("show tables")
	if err != nil {
		return err, nil
	}
	var tableNames []string
	//循环读取结果
	for rows.Next() {
		var tableName string
		rows.Scan(&tableName)
		tableNames = append(tableNames, tableName)
	}
	return nil, tableNames
}

func GetTableDataMap(connUrl string, tableName string) (error, []map[string]interface{}) {
	db := GetConn(connUrl)
	defer CloseDB(db)
	rows, err := db.Query("select * from " + tableName)
	if err != nil {
		return err, nil
	}
	columns, _ := rows.Columns()
	columnLength := len(columns)
	var list []map[string]interface{} //返回的切片
	for rows.Next() {
		cache := rowsToArray(columnLength, rows)
		item := make(map[string]interface{})
		for i, data := range cache {
			item[columns[i]] = *data.(*interface{}) //取实际类型
		}
		list = append(list, item)
	}
	_ = rows.Close()
	return nil, list
}
