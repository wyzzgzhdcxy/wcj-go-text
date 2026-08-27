package app

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/wyzzgzhdcxy/wcj-go-common/core"
)

// Assets 持有嵌入式资源（来自 main.go），用于读取 system.config.json

// settingsDb 图像/emoji 设置数据库（image_tools.db）
var settingsDb *sql.DB

const emojiTmpDir = `C:\Users\wangchaojun\AppData\Local\wtools\tmp\emoji`

// initImageSettingsDb 初始化配置数据库（sqlite）
func (a *App) initImageSettingsDb() error {
	dbPath := filepath.Join(core.GetTempDir(), "data", "image_tools.db")
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("创建数据库目录失败: %v", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("打开数据库失败: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS settings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			key TEXT NOT NULL UNIQUE,
			value TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		db.Close()
		return fmt.Errorf("创建表失败: %v", err)
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_settings_key ON settings(key)`)
	if err != nil {
		db.Close()
		return fmt.Errorf("创建索引失败: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS image_prompts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			prompt TEXT NOT NULL UNIQUE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		db.Close()
		return fmt.Errorf("创建图片提示词表失败: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS emoji_images (
			id TEXT PRIMARY KEY,
			png_data BLOB NOT NULL,
			ico_data BLOB NOT NULL,
			emoji TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		db.Close()
		return fmt.Errorf("创建 emoji 图片表失败: %v", err)
	}

	settingsDb = db
	log.Printf("图片配置数据库初始化成功: %s", dbPath)
	return nil
}

func saveSetting(key, value string) {
	if settingsDb == nil {
		return
	}
	_, err := settingsDb.Exec(`
		INSERT INTO settings (key, value, updated_at)
		VALUES (?, ?, datetime('now'))
		ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at
	`, key, value)
	if err != nil {
		log.Printf("保存配置失败: %v", err)
	}
}

func getSetting(key string) string {
	if settingsDb == nil {
		return ""
	}
	var value string
	err := settingsDb.QueryRow("SELECT value FROM settings WHERE key = ?", key).Scan(&value)
	if err != nil {
		return ""
	}
	return value
}

// WindowState 窗口状态
type WindowState struct {
	X      int
	Y      int
	Width  int
	Height int
}

func (a *App) SaveWindowState(x, y, width, height int) {
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if width < 600 {
		width = 600
	}
	if height < 400 {
		height = 400
	}
	saveSetting("window.x", fmt.Sprintf("%d", x))
	saveSetting("window.y", fmt.Sprintf("%d", y))
	saveSetting("window.width", fmt.Sprintf("%d", width))
	saveSetting("window.height", fmt.Sprintf("%d", height))
	saveSetting("window.x_saved", "1")
}

func (a *App) GetWindowState() WindowState {
	ws := WindowState{}
	if settingsDb == nil {
		return ws
	}
	xSaved := getSetting("window.x_saved")
	if xSaved != "1" {
		return ws
	}
	x := getSetting("window.x")
	if x == "" {
		return ws
	}
	ws.X, _ = strconv.Atoi(x)
	ws.Y, _ = strconv.Atoi(getSetting("window.y"))
	ws.Width, _ = strconv.Atoi(getSetting("window.width"))
	ws.Height, _ = strconv.Atoi(getSetting("window.height"))
	if ws.X < 0 {
		ws.X = 0
	}
	if ws.Y < 0 {
		ws.Y = 0
	}
	if ws.Width < 600 {
		ws.Width = 600
	}
	if ws.Height < 400 {
		ws.Height = 400
	}
	return ws
}

// AddConfigValue 添加配置
func (a *App) AddConfigValue(key, value string) string {
	if settingsDb == nil {
		return "数据库未初始化"
	}
	_, err := settingsDb.Exec(`
		INSERT INTO settings (key, value, updated_at)
		VALUES (?, ?, datetime('now'))
		ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at
	`, key, value)
	if err != nil {
		log.Printf("保存配置失败: %v", err)
		return "Failed: " + err.Error()
	}
	return "OK"
}

// DeleteConfig 删除配置
func (a *App) DeleteConfig(key string) {
	if settingsDb == nil {
		return
	}
	_, err := settingsDb.Exec("DELETE FROM settings WHERE key = ?", key)
	if err != nil {
		log.Printf("删除配置失败: %v", err)
	}
}

// ListConfig 获取配置列表
func (a *App) ListConfig(prefix string) map[string]string {
	tmpMap := make(map[string]string)
	if settingsDb == nil {
		return tmpMap
	}

	query := "SELECT key, value FROM settings WHERE key LIKE ? ORDER BY key LIMIT 100"
	rows, err := settingsDb.Query(query, prefix+"%")
	if err != nil {
		log.Printf("查询配置失败: %v", err)
		return tmpMap
	}
	defer rows.Close()

	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			continue
		}
		tmpMap[key] = value
	}
	return tmpMap
}

// GetConfigKey 获取配置（先查 DB，再读 assets/config/system.config.json）
func (a *App) GetConfigKey(key string) string {
	if settingsDb != nil {
		var value string
		err := settingsDb.QueryRow("SELECT value FROM settings WHERE key = ?", "system.setting."+key).Scan(&value)
		if err == nil && value != "" {
			return value
		}
	}
	if Assets != (embed.FS{}) {
		content, _ := Assets.ReadFile("config/system.config.json")
		var myMap map[string]string
		contentByte := []byte(content)
		core.JsonToObject(&contentByte, &myMap)
		return myMap[key]
	}
	return ""
}
