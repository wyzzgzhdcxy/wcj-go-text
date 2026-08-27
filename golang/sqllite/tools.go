package sqllite

import (
	"database/sql"
	"fmt"
)

// toolsDB 单例句柄，对应 tools_settings.db。
var toolsDB *sql.DB

func getToolsDB() *sql.DB {
	if toolsDB != nil {
		return toolsDB
	}
	db := getOrOpen("tools_settings.db")
	if db == nil {
		return nil
	}
	execSchema(db, toolsSchema, "创建 tools 表失败:")
	toolsDB = db
	return toolsDB
}

var toolsSchema = []string{
	`CREATE TABLE IF NOT EXISTS settings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		key TEXT NOT NULL UNIQUE,
		value TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE INDEX IF NOT EXISTS idx_settings_key ON settings(key)`,
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
	`CREATE TABLE IF NOT EXISTS cmd_commands (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		command TEXT NOT NULL,
		description TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
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

// CmdCommand 命令行管理记录
type CmdCommand struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Command     string `json:"command"`
	Description string `json:"description"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// ListCmdCommands 倒序列出全部命令管理记录。
func ListCmdCommands() ([]CmdCommand, error) {
	db := getToolsDB()
	if db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}
	rows, err := db.Query("SELECT id, name, command, description, created_at, updated_at FROM cmd_commands ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []CmdCommand
	for rows.Next() {
		var c CmdCommand
		if err := rows.Scan(&c.ID, &c.Name, &c.Command, &c.Description, &c.CreatedAt, &c.UpdatedAt); err == nil {
			out = append(out, c)
		}
	}
	return out, nil
}

// AddCmdCommand 插入一条命令管理记录，返回新 id。
func AddCmdCommand(c CmdCommand) (int, error) {
	db := getToolsDB()
	if db == nil {
		return 0, fmt.Errorf("数据库未初始化")
	}
	res, err := db.Exec("INSERT INTO cmd_commands(name, command, description) VALUES(?, ?, ?)",
		c.Name, c.Command, c.Description)
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return int(id), nil
}

// UpdateCmdCommand 按 id 更新命令管理记录。
func UpdateCmdCommand(c CmdCommand) error {
	db := getToolsDB()
	if db == nil {
		return fmt.Errorf("数据库未初始化")
	}
	_, err := db.Exec("UPDATE cmd_commands SET name=?, command=?, description=?, updated_at=datetime('now') WHERE id=?",
		c.Name, c.Command, c.Description, c.ID)
	return err
}

// DeleteCmdCommand 按 id 删除命令管理记录。
func DeleteCmdCommand(id int) error {
	db := getToolsDB()
	if db == nil {
		return fmt.Errorf("数据库未初始化")
	}
	_, err := db.Exec("DELETE FROM cmd_commands WHERE id=?", id)
	return err
}

// GetCmdCommandByID 按 id 读取命令管理记录。
func GetCmdCommandByID(id int) (string, error) {
	db := getToolsDB()
	if db == nil {
		return "", fmt.Errorf("数据库未初始化")
	}
	var stored string
	err := db.QueryRow("SELECT command FROM cmd_commands WHERE id = ?", id).Scan(&stored)
	return stored, err
}