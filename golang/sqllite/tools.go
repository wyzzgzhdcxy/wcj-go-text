package sqllite

import (
	"database/sql"
	"fmt"
	"sync"
)

// toolsDB 单例句柄，对应 tools.db 中的 tool_usage_stats / cmd_history
// （settings 表由 settings.go 统一管理）。
var (
	toolsDB   *sql.DB
	toolsOnce sync.Once
)

func getToolsDB() *sql.DB {
	toolsOnce.Do(func() {
		db := getOrOpen(mainDBFile)
		if db == nil {
			return
		}
		execSchema(db, toolsSchema, "创建 tools 表失败:")
		toolsDB = db
	})
	return toolsDB
}

var toolsSchema = []string{
	`CREATE TABLE IF NOT EXISTS tool_usage_stats (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		link TEXT NOT NULL,
		use_count INTEGER DEFAULT 1,
		last_used DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(link)
	)`,
	`CREATE TABLE IF NOT EXISTS cmd_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		command TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`,
}

// RecordCmdHistory 追加一条命令执行历史。
func RecordCmdHistory(command string) error {
	db := getToolsDB()
	if db == nil {
		return fmt.Errorf("数据库未初始化")
	}
	_, err := db.Exec("INSERT INTO cmd_history(command) VALUES(?)", command)
	return err
}

// GetCmdHistory 倒序返回最近 50 条命令历史。
func GetCmdHistory() ([]string, error) {
	db := getToolsDB()
	if db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}
	rows, err := db.Query("SELECT command FROM cmd_history ORDER BY created_at DESC LIMIT 50")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err == nil {
			out = append(out, c)
		}
	}
	return out, nil
}

// RecordToolUsage 工具使用次数 +1（首次插入）。
func RecordToolUsage(link string) error {
	db := getToolsDB()
	if db == nil {
		return fmt.Errorf("数据库未初始化")
	}
	_, err := db.Exec(`INSERT INTO tool_usage_stats(link, use_count, last_used)
		VALUES(?, 1, CURRENT_TIMESTAMP)
		ON CONFLICT(link) DO UPDATE SET use_count = use_count + 1, last_used = CURRENT_TIMESTAMP`, link)
	return err
}

// GetToolUsageStats 按使用次数倒序返回前 20 个 link -> count。
func GetToolUsageStats() (map[string]int, error) {
	db := getToolsDB()
	if db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}
	rows, err := db.Query("SELECT link, use_count FROM tool_usage_stats ORDER BY use_count DESC LIMIT 20")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[string]int)
	for rows.Next() {
		var link string
		var count int
		if err := rows.Scan(&link, &count); err == nil {
			out[link] = count
		}
	}
	return out, nil
}
