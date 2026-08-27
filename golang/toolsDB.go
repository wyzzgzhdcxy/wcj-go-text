package golang

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	_ "modernc.org/sqlite"
)

var (
	toolsDB   *sql.DB
	toolsOnce sync.Once
)

func GetToolsDB() *sql.DB {
	toolsOnce.Do(func() {
		initToolsDB()
	})
	return toolsDB
}

func initToolsDB() {
	appDir := getAppDataDir()
	dbPath := filepath.Join(appDir, "tools_settings.db")
	os.MkdirAll(appDir, 0755)

	var err error
	toolsDB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		fmt.Println("打开tools数据库失败:", err)
		return
	}

	if err = toolsDB.Ping(); err != nil {
		fmt.Println("tools数据库连接失败:", err)
		return
	}

	createToolsTables()
}

func createToolsTables() {
	_, err := toolsDB.Exec(`
		CREATE TABLE IF NOT EXISTS settings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			key TEXT NOT NULL UNIQUE,
			value TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		fmt.Println("创建配置表失败:", err)
	}

	_, err = toolsDB.Exec(`CREATE INDEX IF NOT EXISTS idx_settings_key ON settings(key)`)
	if err != nil {
		fmt.Println("创建配置索引失败:", err)
	}

	_, err = toolsDB.Exec(`
		CREATE TABLE IF NOT EXISTS tool_usage_stats (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			link TEXT NOT NULL,
			use_count INTEGER DEFAULT 1,
			last_used DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(link)
		)
	`)
	if err != nil {
		fmt.Println("创建工具使用统计表失败:", err)
	}

	_, err = toolsDB.Exec(`
		CREATE TABLE IF NOT EXISTS cmd_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			command TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		fmt.Println("创建命令历史表失败:", err)
	}

	_, err = toolsDB.Exec(`
		CREATE TABLE IF NOT EXISTS cmd_commands (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			command TEXT NOT NULL,
			description TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		fmt.Println("创建命令管理表失败:", err)
	}
}
