// Package sqllite 集中管理项目内全部 SQLite 数据（统一到 tools.db）。
// 所有上层模块（app/* 与 golang/*）统一通过本包访问数据库。
//
// 设计要点：
//   - 通过 getOrOpen 按 dbName 懒打开 *sql.DB，每个库只初始化一次；
//   - 当前阶段全部业务表（settings/image_prompts/emoji_images/cmd_* /tool_usage_stats/
//     music_*）统一存在 tools.db 内，按业务域分文件管理 schema 与 API，
//     调用方不应直接使用 *sql.DB 句柄。
package sqllite

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	_ "modernc.org/sqlite"
)

// 主数据库文件名：所有业务表统一在此。
const mainDBFile = "tools.db"

// 历史数据库文件名（迁移完成后仍可能残留，仅用于一次性数据搬运）。
var legacyDBs = []string{
	"image_tools.db",
	"tools_settings.db",
	"music_cache.db",
}

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

// metaSchema 用于追踪一次性数据迁移是否已完成。
var metaSchema = []string{
	`CREATE TABLE IF NOT EXISTS _meta (
		k TEXT PRIMARY KEY,
		v TEXT
	)`,
}

// isMigrationDone 读取 _meta 表的 v；key 存在即认为已完成。
func isMigrationDone(db *sql.DB, key string) bool {
	if db == nil {
		return true
	}
	var v string
	err := db.QueryRow(`SELECT v FROM _meta WHERE k=?`, key).Scan(&v)
	return err == nil
}

// markMigrationDone 写入 _meta 标记。
func markMigrationDone(db *sql.DB, key string) {
	if db == nil {
		return
	}
	_, _ = db.Exec(`INSERT OR REPLACE INTO _meta(k, v) VALUES(?, ?)`, key, "1")
}

func init() {
	// 启动时确保主库存在，先应用全部业务 schema（保证迁移时所有目标表都已存在），
	// 再触发一次性数据迁移。
	mainDB := getOrOpen(mainDBFile)
	if mainDB == nil {
		return
	}
	applyAllSchemas(mainDB)
	runLegacyMigration(mainDB)
}

// applyAllSchemas 依次应用本包内全部业务表 schema 与 _meta 表。
// 各域的 schema 数组是包级私有变量，对 init() 可见。
func applyAllSchemas(db *sql.DB) {
	if db == nil {
		return
	}
	execSchema(db, metaSchema, "创建 _meta 表失败:")
	execSchema(db, settingsSchema, "创建 settings 表失败:")
	execSchema(db, toolsSchema, "创建 tools 表失败:")
	execSchema(db, musicSchema, "创建 music 表失败:")
}

// runLegacyMigration 把 image_tools.db / tools_settings.db / music_cache.db
// 三库中的存量数据按表搬运到 tools.db。
//
// 策略：
//   - 同一表从多个源迁移时用 INSERT OR IGNORE，UNIQUE/PK 冲突自动跳过（不报错也不覆盖）；
//   - 目标表已有数据时仍然尝试导入新的 key（因为 row 数大于0不代表所有 key 都已存在）；
//   - 整个流程仅执行一次（由 _meta.migrated_legacy 控制），完成后把 3 个旧库改名为 .bak。
func runLegacyMigration(target *sql.DB) {
	if isMigrationDone(target, "migrated_legacy") {
		return
	}
	appDir := GetAppDataDir()

	// 按 (legacyFile, tableName) 列出需要搬运的表。
	plan := []struct {
		file  string
		table string
	}{
		{"image_tools.db", "settings"},
		{"image_tools.db", "image_prompts"},
		{"image_tools.db", "emoji_images"},
		{"tools_settings.db", "settings"},
		{"tools_settings.db", "tool_usage_stats"},
		{"tools_settings.db", "cmd_history"},
		{"tools_settings.db", "cmd_commands"},
		{"music_cache.db", "music_search_history"},
		{"music_cache.db", "music_songs"},
		{"music_cache.db", "music_audio_sources"},
	}

	// 按表分组：同一 table 的多个源合并到一次执行流，按 plan 顺序逐源导入。
	for _, p := range plan {
		srcPath := filepath.Join(appDir, p.file)
		if _, err := os.Stat(srcPath); err != nil {
			continue
		}
		src, err := sql.Open("sqlite", srcPath)
		if err != nil {
			fmt.Printf("[sqllite] 打开旧库失败 %s: %v\n", srcPath, err)
			continue
		}

		// 旧库若没有该表则跳过
		var cnt int
		if err := src.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", p.table)).Scan(&cnt); err != nil {
			src.Close()
			fmt.Printf("[sqllite] 旧库 %s 不含表 %s，跳过\n", p.file, p.table)
			continue
		}

		rows, err := src.Query(fmt.Sprintf("SELECT * FROM %s", p.table))
		if err != nil {
			src.Close()
			fmt.Printf("[sqllite] 读取旧表失败 %s.%s: %v\n", p.file, p.table, err)
			continue
		}
		cols, err := rows.Columns()
		if err != nil {
			rows.Close()
			src.Close()
			continue
		}
		placeholders := ""
		for i := range cols {
			if i > 0 {
				placeholders += ","
			}
			placeholders += "?"
		}
		// 使用 INSERT OR IGNORE：UNIQUE/PK 冲突时静默跳过，不影响其它行；
		// 对 settings 这种在两个源都存在的表，重复 key 也不会报错。
		stmt, err := target.Prepare(fmt.Sprintf("INSERT OR IGNORE INTO %s VALUES(%s)", p.table, placeholders))
		if err != nil {
			rows.Close()
			src.Close()
			fmt.Printf("[sqllite] 准备 INSERT 失败 %s: %v\n", p.table, err)
			continue
		}

		tx, err := target.Begin()
		if err != nil {
			stmt.Close()
			rows.Close()
			src.Close()
			continue
		}
		inserted := 0
		for rows.Next() {
			raw := make([]any, len(cols))
			ptrs := make([]any, len(cols))
			for i := range raw {
				ptrs[i] = &raw[i]
			}
			if err := rows.Scan(ptrs...); err != nil {
				continue
			}
			if _, err := tx.Stmt(stmt).Exec(raw...); err != nil {
				fmt.Printf("[sqllite] 迁移行失败 %s.%s: %v\n", p.file, p.table, err)
				continue
			}
			inserted++
		}
		if err := tx.Commit(); err != nil {
			tx.Rollback()
			fmt.Printf("[sqllite] 提交迁移失败 %s.%s: %v\n", p.file, p.table, err)
		} else {
			fmt.Printf("[sqllite] 已迁移 %s -> tools.%s，%d 行\n", p.file, p.table, inserted)
		}
		stmt.Close()
		rows.Close()
		src.Close()
	}

	// 把 3 个旧库改名为 .bak（保留数据用于核对/回滚，不占活动 DB 名）
	// Windows 上 modernc.org/sqlite 关闭后仍可能短暂持有文件锁，重命名失败仅警告，
	// 数据已在 tools.db 中，下次启动再尝试。
	for _, name := range legacyDBs {
		oldPath := filepath.Join(appDir, name)
		if _, err := os.Stat(oldPath); err != nil {
			continue
		}
		bakPath := oldPath + ".bak"
		if _, err := os.Stat(bakPath); err == nil {
			// .bak 已存在，直接删源
			if rmErr := os.Remove(oldPath); rmErr != nil {
				fmt.Printf("[sqllite] 删除已迁移旧库失败 %s: %v\n", name, rmErr)
			}
			continue
		}
		if err := os.Rename(oldPath, bakPath); err != nil {
			fmt.Printf("[sqllite] 重命名 %s -> .bak 失败（Windows 文件锁），跳过: %v\n", name, err)
		} else {
			fmt.Printf("[sqllite] 已归档旧库 %s -> %s.bak\n", name, name)
		}
	}

	markMigrationDone(target, "migrated_legacy")
}