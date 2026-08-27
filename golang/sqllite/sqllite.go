// Package sqllite 集中管理项目内全部 SQLite 数据库（懒加载、按库名复用单例），
// 所有上层模块（app/* 与 golang/*）统一通过本包访问数据库。
//
// 设计要点：
//   - 通过 getOrOpen 按 dbName（文件名）懒打开 *sql.DB，每个库只初始化一次；
//   - 每个业务库在自己的文件里提供高层 API（SearchMusicCache、SaveSetting 等），
//     调用方不应直接使用 *sql.DB 句柄，避免再次散落到业务文件；
//   - 旧代码里的 settingsDb / gitDb / musicDB / toolsDB / imageDB 全在此包内
//     统一收口，调用方只依赖包级函数。
package sqllite

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	_ "modernc.org/sqlite"
)

// 全局并发保护：dbName -> 单例句柄
var (
	dbMu    sync.Mutex
	dbCache = map[string]*sql.DB{}
)

// GetAppDataDir 返回 %LOCALAPPDATA%/wtools/data，等价于 core.GetTempDir()+"/data"。
// 所有业务库文件均存放在此目录下，保持与历史实现一致。
func GetAppDataDir() string {
	dir := os.Getenv("LOCALAPPDATA")
	if dir == "" {
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			dir = filepath.Join(home, "AppData", "Local")
		} else {
			dir = os.TempDir()
		}
	}
	return filepath.Join(dir, "wtools", "data")
}

// getOrOpen 按文件名（相对 GetAppDataDir）懒打开并缓存 *sql.DB。
// 同一 fileName 多次调用只会 Open / Ping 一次。
func getOrOpen(fileName string) *sql.DB {
	dbMu.Lock()
	defer dbMu.Unlock()
	if db, ok := dbCache[fileName]; ok && db != nil {
		return db
	}
	appDir := GetAppDataDir()
	if err := os.MkdirAll(appDir, 0755); err != nil {
		fmt.Println("创建数据库目录失败:", err)
		return nil
	}
	dbPath := filepath.Join(appDir, fileName)
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		fmt.Println("打开数据库失败:", err, "path:", dbPath)
		return nil
	}
	if err = db.Ping(); err != nil {
		fmt.Println("数据库连接失败:", err, "path:", dbPath)
		return nil
	}
	dbCache[fileName] = db
	return db
}

// execSchema 在 db 上依次执行若干 CREATE TABLE / CREATE INDEX 语句；
// 任一语句失败仅打印日志，不中断后续建表（与历史实现一致）。
func execSchema(db *sql.DB, stmts []string, errPrefix string) {
	if db == nil {
		return
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			fmt.Println(errPrefix, err)
		}
	}
}

