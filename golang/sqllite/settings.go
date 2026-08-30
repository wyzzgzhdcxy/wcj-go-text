package sqllite

import (
	"database/sql"
	"fmt"
	"sync"
)

// settingsDB 单例句柄，对应 tools.db 中的 settings / image_prompts / emoji_images。
var (
	settingsDB   *sql.DB
	settingsOnce sync.Once
)

// getSettingsDB 返回 settingsDB，第一次调用时建表并缓存（并发安全）。
func getSettingsDB() *sql.DB {
	settingsOnce.Do(func() {
		db := getOrOpen(mainDBFile)
		if db == nil {
			return
		}
		execSchema(db, settingsSchema, "创建 settings 表失败:")
		settingsDB = db
	})
	return settingsDB
}

var settingsSchema = []string{
	`CREATE TABLE IF NOT EXISTS settings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		key TEXT NOT NULL UNIQUE,
		value TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE INDEX IF NOT EXISTS idx_settings_key ON settings(key)`,
	`CREATE TABLE IF NOT EXISTS image_prompts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		prompt TEXT NOT NULL UNIQUE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE TABLE IF NOT EXISTS emoji_images (
		id TEXT PRIMARY KEY,
		png_data BLOB NOT NULL,
		ico_data BLOB NOT NULL,
		emoji TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`,
}

// SaveSetting 写入/更新一个 KV 配置项。
func SaveSetting(key, value string) error {
	db := getSettingsDB()
	if db == nil {
		return fmt.Errorf("数据库未初始化")
	}
	_, err := db.Exec(`
		INSERT INTO settings (key, value, updated_at)
		VALUES (?, ?, datetime('now'))
		ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at
	`, key, value)
	return err
}

// GetSetting 读取一个 KV 配置项；不存在返回空串。
func GetSetting(key string) string {
	db := getSettingsDB()
	if db == nil {
		return ""
	}
	var v string
	if err := db.QueryRow("SELECT value FROM settings WHERE key = ?", key).Scan(&v); err != nil {
		return ""
	}
	return v
}

// GetSettingWithPrefix 按前缀查询配置项（最多 100 条）。
func GetSettingWithPrefix(prefix string) map[string]string {
	out := make(map[string]string)
	db := getSettingsDB()
	if db == nil {
		return out
	}
	rows, err := db.Query("SELECT key, value FROM settings WHERE key LIKE ? ORDER BY key LIMIT 100", prefix+"%")
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err == nil {
			out[k] = v
		}
	}
	return out
}

// DeleteSetting 删除一个 KV 配置项。
func DeleteSetting(key string) error {
	db := getSettingsDB()
	if db == nil {
		return fmt.Errorf("数据库未初始化")
	}
	_, err := db.Exec("DELETE FROM settings WHERE key = ?", key)
	return err
}

// GetSystemSetting 读取以 system.setting. 为前缀的配置项（用于兼容旧 GetConfigKey）。
func GetSystemSetting(key string) string {
	db := getSettingsDB()
	if db == nil {
		return ""
	}
	var v string
	if err := db.QueryRow("SELECT value FROM settings WHERE key = ?", "system.setting."+key).Scan(&v); err != nil {
		return ""
	}
	return v
}

// SaveImagePrompt 保存一条图片生成提示词（重复则忽略）。
func SaveImagePrompt(prompt string) error {
	if prompt == "" {
		return nil
	}
	db := getSettingsDB()
	if db == nil {
		return fmt.Errorf("数据库未初始化")
	}
	_, err := db.Exec(`INSERT OR IGNORE INTO image_prompts(prompt) VALUES(?)`, prompt)
	return err
}

// GetImagePrompts 倒序返回最近 50 条提示词。
func GetImagePrompts() ([]string, error) {
	db := getSettingsDB()
	if db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}
	rows, err := db.Query(`SELECT prompt FROM image_prompts ORDER BY created_at DESC LIMIT 50`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err == nil {
			out = append(out, p)
		}
	}
	return out, nil
}

// GetEmojiPngData 按 id 读取 emoji 的 png 字节，未命中返回 (nil, sql.ErrNoRows)。
func GetEmojiPngData(id string) ([]byte, error) {
	db := getSettingsDB()
	if db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}
	var data []byte
	err := db.QueryRow("SELECT png_data FROM emoji_images WHERE id = ?", id).Scan(&data)
	return data, err
}

// GetEmojiIcoData 按 id 读取 emoji 的 ico 字节，未命中返回 (nil, sql.ErrNoRows)。
func GetEmojiIcoData(id string) ([]byte, error) {
	db := getSettingsDB()
	if db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}
	var data []byte
	err := db.QueryRow("SELECT ico_data FROM emoji_images WHERE id = ?", id).Scan(&data)
	return data, err
}

// SaveEmojiImage 写入 emoji 的 png/ico 字节，按 id 覆盖。
func SaveEmojiImage(id string, png, ico []byte, emoji string) error {
	db := getSettingsDB()
	if db == nil {
		return fmt.Errorf("数据库未初始化")
	}
	_, err := db.Exec(
		"INSERT OR REPLACE INTO emoji_images (id, png_data, ico_data, emoji) VALUES (?, ?, ?, ?)",
		id, png, ico, emoji,
	)
	return err
}

// EnsureSettingsDB 触发 settingsDB 懒加载（在需要立即初始化的场景可主动调用）。
func EnsureSettingsDB() { getSettingsDB() }
