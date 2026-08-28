package sqllite

import (
	"database/sql"
	"fmt"
	"time"
)

// GitRepo Git 仓库信息（与前端 JSON 字段保持一致）
type GitRepo struct {
	Path            string `json:"path"`
	Name            string `json:"name"`
	Branch          string `json:"branch"`
	Remote          string `json:"remote"`
	RemoteUrl       string `json:"remoteUrl"`
	LastSyncTime    string `json:"lastSyncTime"`
	Status          string `json:"status"`
	Enabled         bool   `json:"enabled"`
	AutoSync        bool   `json:"autoSync"`
	CommitOnly      bool   `json:"commitOnly"`
	LastSyncSuccess int    `json:"lastSyncSuccess"` // -1=从未同步, 0=失败, 1=成功
}

// GitSyncLog 同步日志
type GitSyncLog struct {
	ID        int    `json:"id"`
	RepoName  string `json:"repoName"`
	RepoPath  string `json:"repoPath"`
	Time      string `json:"time"`
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	CommitLog string `json:"commitLog"`
	PullLog   string `json:"pullLog"`
	PushLog   string `json:"pushLog"`
}

// gitDB 单例句柄，对应 sync_list.db 中的 git_repos / sync_logs。
var gitDB *sql.DB

func getGitDB() *sql.DB {
	if gitDB != nil {
		return gitDB
	}
	db := getOrOpen(mainDBFile)
	if db == nil {
		return nil
	}
	execSchema(db, gitSchema, "创建 git 表失败:")
	gitDB = db
	return gitDB
}

var gitSchema = []string{
	`CREATE TABLE IF NOT EXISTS git_repos (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		path TEXT NOT NULL UNIQUE,
		name TEXT NOT NULL,
		branch TEXT,
		remote TEXT,
		remote_url TEXT,
		last_sync_time TEXT,
		status TEXT,
		enabled INTEGER NOT NULL DEFAULT 1,
		auto_sync INTEGER NOT NULL DEFAULT 0,
		commit_only INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE INDEX IF NOT EXISTS idx_git_repos_path ON git_repos(path)`,
	`CREATE TABLE IF NOT EXISTS sync_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		repo_name TEXT NOT NULL,
		repo_path TEXT NOT NULL,
		time TEXT,
		success INTEGER NOT NULL DEFAULT 0,
		message TEXT,
		commit_log TEXT,
		pull_log TEXT,
		push_log TEXT
	)`,
	`CREATE INDEX IF NOT EXISTS idx_sync_logs_repo_path ON sync_logs(repo_path)`,
}

// ClearGitRepos 清空仓库列表（用于 SaveGitRepoList 整体覆盖前的预备步骤）。
func ClearGitRepos() error {
	db := getGitDB()
	if db == nil {
		return fmt.Errorf("数据库未初始化")
	}
	_, err := db.Exec("DELETE FROM git_repos")
	return err
}

// InsertGitRepo 在事务内插入一条仓库记录，调用方需自行管理事务。
func InsertGitRepo(tx *sql.Tx, repo GitRepo) error {
	autoSync, enabled, commitOnly := 0, 0, 0
	if repo.AutoSync {
		autoSync = 1
	}
	if repo.Enabled {
		enabled = 1
	}
	if repo.CommitOnly {
		commitOnly = 1
	}
	_, err := tx.Exec(`
		INSERT INTO git_repos (path, name, branch, remote, remote_url, last_sync_time, status, enabled, auto_sync, commit_only)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, repo.Path, repo.Name, repo.Branch, repo.Remote, repo.RemoteUrl, repo.LastSyncTime, repo.Status, enabled, autoSync, commitOnly)
	return err
}

// BeginGitTx 开启仓库写入事务（用于整体覆盖保存）。
func BeginGitTx() (*sql.Tx, error) {
	db := getGitDB()
	if db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}
	return db.Begin()
}

// LoadGitRepos 加载仓库列表，并附带每个仓库最近一次同步状态。
func LoadGitRepos() ([]GitRepo, error) {
	db := getGitDB()
	if db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}
	rows, err := db.Query(`
		SELECT path, name, branch, remote, remote_url, last_sync_time, status, enabled, auto_sync, commit_only,
			COALESCE((SELECT success FROM sync_logs WHERE repo_path = git_repos.path ORDER BY id DESC LIMIT 1), -1) AS last_sync_success
		FROM git_repos ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []GitRepo
	for rows.Next() {
		var repo GitRepo
		var enabled, autoSync, commitOnly, lastSyncSuccess int
		if err := rows.Scan(&repo.Path, &repo.Name, &repo.Branch, &repo.Remote, &repo.RemoteUrl,
			&repo.LastSyncTime, &repo.Status, &enabled, &autoSync, &commitOnly, &lastSyncSuccess); err != nil {
			continue
		}
		repo.Enabled = enabled == 1
		repo.AutoSync = autoSync == 1
		repo.CommitOnly = commitOnly == 1
		repo.LastSyncSuccess = lastSyncSuccess
		out = append(out, repo)
	}
	return out, nil
}

// RecordSyncLog 写入一条 sync_log 并刷新仓库 last_sync_time。
func RecordSyncLog(repoName, repoPath, message string, success bool, commitLog, pullLog, pushLog string) {
	db := getGitDB()
	if db == nil {
		return
	}
	successInt := 0
	if success {
		successInt = 1
	}
	now := time.Now().Format("2006-01-02 15:04:05")
	if _, err := db.Exec(`
		INSERT INTO sync_logs (repo_name, repo_path, time, success, message, commit_log, pull_log, push_log)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, repoName, repoPath, now, successInt, message, commitLog, pullLog, pushLog); err != nil {
		fmt.Println("写入 sync_log 失败:", err)
	}
	if _, err := db.Exec(`UPDATE git_repos SET last_sync_time = ?, updated_at = CURRENT_TIMESTAMP WHERE path = ?`,
		now, repoPath); err != nil {
		fmt.Println("更新 git_repos.last_sync_time 失败:", err)
	}
}

// GetSyncLogs 拉取同步日志，limit<=0 或 >100 时使用默认值 20。
func GetSyncLogs(repoPath string, limit int) ([]GitSyncLog, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	db := getGitDB()
	if db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}
	query := "SELECT id, repo_name, repo_path, time, success, message, commit_log, pull_log, push_log FROM sync_logs"
	args := []interface{}{}
	if repoPath != "" {
		query += " WHERE repo_path = ?"
		args = append(args, repoPath)
	}
	query += " ORDER BY time DESC LIMIT ?"
	args = append(args, limit)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []GitSyncLog
	for rows.Next() {
		var l GitSyncLog
		var success int
		if err := rows.Scan(&l.ID, &l.RepoName, &l.RepoPath, &l.Time, &success,
			&l.Message, &l.CommitLog, &l.PullLog, &l.PushLog); err == nil {
			l.Success = success == 1
			out = append(out, l)
		}
	}
	return out, nil
}