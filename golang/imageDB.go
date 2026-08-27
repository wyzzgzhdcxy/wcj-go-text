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
	imageDB   *sql.DB
	imageOnce sync.Once
)

func GetImageDB() *sql.DB {
	imageOnce.Do(func() {
		initImageDB()
	})
	return imageDB
}

func initImageDB() {
	appDir := getAppDataDir()
	dbPath := filepath.Join(appDir, "image_tools.db")
	os.MkdirAll(appDir, 0755)

	var err error
	imageDB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		fmt.Println("打开image数据库失败:", err)
		return
	}

	if err = imageDB.Ping(); err != nil {
		fmt.Println("image数据库连接失败:", err)
		return
	}

	createImageTables()
}

func createImageTables() {
	_, err := imageDB.Exec(`
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

	_, err = imageDB.Exec(`CREATE INDEX IF NOT EXISTS idx_settings_key ON settings(key)`)
	if err != nil {
		fmt.Println("创建配置索引失败:", err)
	}

	_, err = imageDB.Exec(`
		CREATE TABLE IF NOT EXISTS image_prompts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			prompt TEXT NOT NULL UNIQUE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		fmt.Println("创建图片提示词表失败:", err)
	}

	_, err = imageDB.Exec(`
		CREATE TABLE IF NOT EXISTS emoji_images (
			id TEXT PRIMARY KEY,
			png_data BLOB NOT NULL,
			ico_data BLOB NOT NULL,
			emoji TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		fmt.Println("创建emoji图片表失败:", err)
	}
}
