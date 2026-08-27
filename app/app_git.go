package app

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"wcj-go-text/golang/cmdWrapper"

	"github.com/wyzzgzhdcxy/wcj-go-common/core"
)

var (
	gitDb     *sql.DB
	gitDbOnce sync.Once
)

// initGitDb 懒加载初始化 Git 工具的 sqlite 数据库（git_sync.db）。
// 表：git_repos（仓库列表）、sync_logs（同步日志）。
// 失败时仅记录日志，调用方做 nil 兜底。
func (a *App) initGitDb() {
	gitDbOnce.Do(func() {
		dbPath := filepath.Join(core.GetTempDir(), "data", "sync_list.db")
		if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
			log.Printf("创建 git 同步数据库目录失败: %v", err)
			return
		}

		db, err := sql.Open("sqlite", dbPath)
		if err != nil {
			log.Printf("打开 git 同步数据库失败: %v", err)
			return
		}

		_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS git_repos (
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
			)
		`)
		if err != nil {
			db.Close()
			log.Printf("创建 git_repos 表失败: %v", err)
			return
		}

		_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_git_repos_path ON git_repos(path)`)
		if err != nil {
			db.Close()
			log.Printf("创建 git_repos 索引失败: %v", err)
			return
		}

		_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS sync_logs (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				repo_name TEXT NOT NULL,
				repo_path TEXT NOT NULL,
				time TEXT,
				success INTEGER NOT NULL DEFAULT 0,
				message TEXT,
				commit_log TEXT,
				pull_log TEXT,
				push_log TEXT
			)
		`)
		if err != nil {
			db.Close()
			log.Printf("创建 sync_logs 表失败: %v", err)
			return
		}

		_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_sync_logs_repo_path ON sync_logs(repo_path)`)
		if err != nil {
			db.Close()
			log.Printf("创建 sync_logs 索引失败: %v", err)
			return
		}

		gitDb = db
		log.Printf("Git 同步数据库初始化成功: %s", dbPath)
	})
}

// ensureGitDb 兜底触发懒加载。
func (a *App) ensureGitDb() {
	if gitDb == nil {
		a.initGitDb()
	}
}

// ---------- 类型定义 ----------

// GitRepo Git仓库信息
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

// GitSyncReq Git同步请求
type GitSyncReq struct {
	Repos []GitRepo `json:"repos"`
}

// GitSyncResult 单个仓库同步结果
type GitSyncResult struct {
	Path      string `json:"path"`
	Name      string `json:"name"`
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	PullLog   string `json:"pullLog"`
	PushLog   string `json:"pushLog"`
	CommitLog string `json:"commitLog"`
	Committed bool   `json:"committed"`
	Pushed    bool   `json:"pushed"`
}

// GitSyncRes Git同步结果
type GitSyncRes struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Results []GitSyncResult `json:"results"`
}

// GetGitRepoInfoReq 获取Git仓库信息请求
type GetGitRepoInfoReq struct {
	Path string `json:"path"`
}

// GetGitRepoInfoRes 获取仓库信息结果
type GetGitRepoInfoRes struct {
	Success   bool     `json:"success"`
	Message   string   `json:"message"`
	Repo      *GitRepo `json:"repo"`
	IsGitRepo bool     `json:"isGitRepo"`
}

// GitRepoListReq 仓库列表请求
type GitRepoListReq struct {
	Repos []GitRepo `json:"repos"`
}

// GitRepoListRes 仓库列表结果
type GitRepoListRes struct {
	Success bool      `json:"success"`
	Message string    `json:"message"`
	Repos   []GitRepo `json:"repos"`
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

// GetSyncLogsReq 获取同步日志请求
type GetSyncLogsReq struct {
	RepoPath string `json:"repoPath"`
	Limit    int    `json:"limit"`
}

// GetSyncLogsRes 获取同步日志结果
type GetSyncLogsRes struct {
	Success bool         `json:"success"`
	Message string       `json:"message"`
	Logs    []GitSyncLog `json:"logs"`
}

// ResetReq 重置请求
type ResetReq struct {
	Path string `json:"path"`
}

// ResetResult 重置结果
type ResetResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Output  string `json:"output"`
}

// SoftResetReq 软重置请求
type SoftResetReq struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}

// SoftResetResult 软重置结果
type SoftResetResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Output  string `json:"output"`
}

// PackageReq 打包请求
type PackageReq struct {
	Path string `json:"path"`
}

// PackageResult 打包结果
type PackageResult struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Output    string `json:"output"`
	OutputDir string `json:"outputDir"`
}

// ---------- Git 工具 ----------

// runGitDir 在指定目录执行 git 命令，返回去除尾部空格的 combined 输出。
func runGitDir(dir string, args ...string) (string, error) {
	cmd := cmdWrapper.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// recordSyncLog 写入 sync_logs 表一条记录。
func recordSyncLog(repoName, repoPath, message string, success bool, commitLog, pullLog, pushLog string) {
	if gitDb == nil {
		return
	}
	successInt := 0
	if success {
		successInt = 1
	}
	_, err := gitDb.Exec(`
		INSERT INTO sync_logs (repo_name, repo_path, time, success, message, commit_log, pull_log, push_log)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, repoName, repoPath, time.Now().Format("2006-01-02 15:04:05"), successInt, message, commitLog, pullLog, pushLog)
	if err != nil {
		log.Printf("写入 sync_log 失败: %v", err)
	}

	// 更新仓库的 last_sync_time
	_, _ = gitDb.Exec(`UPDATE git_repos SET last_sync_time = ?, updated_at = CURRENT_TIMESTAMP WHERE path = ?`,
		time.Now().Format("2006-01-02 15:04:05"), repoPath)
}

// GetGitRepoInfo 获取 Git 仓库信息
func (a *App) GetGitRepoInfo(req GetGitRepoInfoReq) GetGitRepoInfoRes {
	if req.Path == "" {
		return GetGitRepoInfoRes{Success: false, Message: "请输入仓库路径"}
	}
	if _, err := os.Stat(req.Path); os.IsNotExist(err) {
		return GetGitRepoInfoRes{Success: false, Message: "目录不存在"}
	}
	if _, err := os.Stat(filepath.Join(req.Path, ".git")); err != nil {
		return GetGitRepoInfoRes{Success: false, Message: "不是 Git 仓库", IsGitRepo: false}
	}

	branchOutput, _ := runGitDir(req.Path, "rev-parse", "--abbrev-ref", "HEAD")
	branch := strings.TrimSpace(branchOutput)

	remoteOutput, _ := runGitDir(req.Path, "remote", "get-url", "origin")
	remoteUrl := strings.TrimSpace(remoteOutput)

	repoName := filepath.Base(req.Path)
	repo := &GitRepo{
		Path:      req.Path,
		Name:      repoName,
		Branch:    branch,
		Remote:    "origin",
		RemoteUrl: remoteUrl,
		Status:    "就绪",
		Enabled:   true,
	}
	return GetGitRepoInfoRes{Success: true, Message: "获取成功", Repo: repo, IsGitRepo: true}
}

// SaveGitRepoList 保存仓库列表到 SQLite（整体覆盖：先删后插）。
func (a *App) SaveGitRepoList(req GitRepoListReq) GitRepoListRes {
	a.ensureGitDb()
	if gitDb == nil {
		return GitRepoListRes{Success: false, Message: "数据库未初始化"}
	}
	if _, err := gitDb.Exec("DELETE FROM git_repos"); err != nil {
		return GitRepoListRes{Success: false, Message: "清空仓库列表失败: " + err.Error()}
	}
	tx, err := gitDb.Begin()
	if err != nil {
		return GitRepoListRes{Success: false, Message: "开始事务失败: " + err.Error()}
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO git_repos (path, name, branch, remote, remote_url, last_sync_time, status, enabled, auto_sync, commit_only)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return GitRepoListRes{Success: false, Message: "预处理失败: " + err.Error()}
	}
	defer stmt.Close()

	for _, repo := range req.Repos {
		autoSync := 0
		if repo.AutoSync {
			autoSync = 1
		}
		enabled := 0
		if repo.Enabled {
			enabled = 1
		}
		commitOnly := 0
		if repo.CommitOnly {
			commitOnly = 1
		}
		if _, err := stmt.Exec(repo.Path, repo.Name, repo.Branch, repo.Remote, repo.RemoteUrl, repo.LastSyncTime, repo.Status, enabled, autoSync, commitOnly); err != nil {
			return GitRepoListRes{Success: false, Message: "保存仓库失败: " + err.Error()}
		}
	}
	if err := tx.Commit(); err != nil {
		return GitRepoListRes{Success: false, Message: "提交事务失败: " + err.Error()}
	}
	return GitRepoListRes{Success: true, Message: fmt.Sprintf("保存了 %d 个仓库", len(req.Repos)), Repos: req.Repos}
}

// LoadGitRepoList 从 SQLite 加载仓库列表，并附带最近一次同步状态。
func (a *App) LoadGitRepoList() GitRepoListRes {
	a.ensureGitDb()
	if gitDb == nil {
		return GitRepoListRes{Success: false, Message: "数据库未初始化", Repos: []GitRepo{}}
	}
	rows, err := gitDb.Query(`
		SELECT path, name, branch, remote, remote_url, last_sync_time, status, enabled, auto_sync, commit_only,
			COALESCE((SELECT success FROM sync_logs WHERE repo_path = git_repos.path ORDER BY id DESC LIMIT 1), -1) AS last_sync_success
		FROM git_repos ORDER BY id
	`)
	if err != nil {
		return GitRepoListRes{Success: false, Message: "查询失败: " + err.Error(), Repos: []GitRepo{}}
	}
	defer rows.Close()

	repos := []GitRepo{}
	for rows.Next() {
		var repo GitRepo
		var enabled, autoSync, commitOnly int
		var lastSyncSuccess int
		if err := rows.Scan(&repo.Path, &repo.Name, &repo.Branch, &repo.Remote, &repo.RemoteUrl, &repo.LastSyncTime, &repo.Status, &enabled, &autoSync, &commitOnly, &lastSyncSuccess); err != nil {
			continue
		}
		repo.Enabled = enabled == 1
		repo.AutoSync = autoSync == 1
		repo.CommitOnly = commitOnly == 1
		repo.LastSyncSuccess = lastSyncSuccess
		repos = append(repos, repo)
	}
	return GitRepoListRes{Success: true, Message: fmt.Sprintf("加载了 %d 个仓库", len(repos)), Repos: repos}
}

// GitSync 同步 Git 仓库：add → commit → pull → push（commitOnly 时跳过 push）。
// 原 wcj-go-git 通过 HTTP 调用外部 sync 服务，这里改为内联执行以保持自包含。
func (a *App) GitSync(req GitSyncReq) GitSyncRes {
	a.ensureGitDb()
	results := make([]GitSyncResult, 0, len(req.Repos))
	if len(req.Repos) == 0 {
		return GitSyncRes{Success: true, Message: "没有需要同步的仓库", Results: results}
	}

	for _, repo := range req.Repos {
		res := syncOneRepo(repo)
		results = append(results, res)
		recordSyncLog(repo.Name, repo.Path, res.Message, res.Success, res.CommitLog, res.PullLog, res.PushLog)
	}

	overallSuccess := true
	for _, r := range results {
		if !r.Success {
			overallSuccess = false
			break
		}
	}
	return GitSyncRes{
		Success: overallSuccess,
		Message: fmt.Sprintf("同步完成：%d 个仓库", len(results)),
		Results: results,
	}
}

// syncOneRepo 执行单个仓库的同步流程。
func syncOneRepo(repo GitRepo) GitSyncResult {
	res := GitSyncResult{Path: repo.Path, Name: repo.Name}
	if repo.Path == "" {
		res.Message = "路径为空"
		return res
	}
	if _, err := os.Stat(filepath.Join(repo.Path, ".git")); err != nil {
		res.Message = "不是 Git 仓库"
		return res
	}

	// 1. add .
	_, _ = runGitDir(repo.Path, "add", ".")

	// 2. 仅在没有未提交内容时跳过 commit
	statusOut, _ := runGitDir(repo.Path, "status", "--porcelain")
	if strings.TrimSpace(statusOut) != "" {
		commitOut, commitErr := runGitDir(repo.Path, "commit", "-m", "wtools update")
		res.CommitLog = commitOut
		if commitErr != nil {
			res.Message = "提交失败: " + commitErr.Error()
			res.Success = false
			return res
		}
		res.Committed = true
	} else {
		res.CommitLog = "无变更需要提交"
	}

	// 3. pull
	pullOut, pullErr := runGitDir(repo.Path, "pull", "--rebase", "--autostash")
	res.PullLog = pullOut
	if pullErr != nil && !strings.Contains(pullOut, "Already up to date") {
		res.Message = "拉取失败: " + pullErr.Error()
		res.Success = false
		return res
	}

	// 4. push（commitOnly 模式跳过）
	if !repo.CommitOnly {
		pushOut, pushErr := runGitDir(repo.Path, "push")
		res.PushLog = pushOut
		if pushErr != nil {
			res.Message = "推送失败: " + pushErr.Error()
			res.Success = false
			return res
		}
		res.Pushed = true
	} else {
		res.PushLog = "仅提交模式，跳过推送"
	}

	res.Success = true
	if res.Committed {
		res.Message = "已提交并同步"
	} else if !repo.CommitOnly {
		res.Message = "已拉取并推送"
	} else {
		res.Message = "已拉取（仅提交）"
	}
	return res
}

// GetSyncLogs 获取同步日志
func (a *App) GetSyncLogs(req GetSyncLogsReq) GetSyncLogsRes {
	a.ensureGitDb()
	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if gitDb == nil {
		return GetSyncLogsRes{Success: false, Message: "数据库未初始化", Logs: []GitSyncLog{}}
	}

	query := "SELECT id, repo_name, repo_path, time, success, message, commit_log, pull_log, push_log FROM sync_logs"
	args := []interface{}{}
	if req.RepoPath != "" {
		query += " WHERE repo_path = ?"
		args = append(args, req.RepoPath)
	}
	query += " ORDER BY time DESC LIMIT ?"
	args = append(args, limit)

	rows, err := gitDb.Query(query, args...)
	if err != nil {
		return GetSyncLogsRes{Success: false, Message: "查询失败: " + err.Error(), Logs: []GitSyncLog{}}
	}
	defer rows.Close()

	logs := []GitSyncLog{}
	for rows.Next() {
		var syncLog GitSyncLog
		var success int
		if err := rows.Scan(&syncLog.ID, &syncLog.RepoName, &syncLog.RepoPath, &syncLog.Time, &success, &syncLog.Message, &syncLog.CommitLog, &syncLog.PullLog, &syncLog.PushLog); err != nil {
			log.Printf("扫描同步日志失败: %v", err)
			continue
		}
		syncLog.Success = success == 1
		logs = append(logs, syncLog)
	}
	return GetSyncLogsRes{Success: true, Message: fmt.Sprintf("共 %d 条日志", len(logs)), Logs: logs}
}

// ResetProject 重置项目：删除 .git → 重新初始化 → 提交 → 强制推送。
func (a *App) ResetProject(req ResetReq) ResetResult {
	projectDir := req.Path
	log.Printf("开始重置项目, 目录: %s", projectDir)
	if projectDir == "" {
		return ResetResult{Success: false, Message: "请提供项目路径"}
	}

	var output string
	runStep := func(name string, fn func() error) {
		output += name + "\n"
		if err := fn(); err != nil {
			output += "失败: " + err.Error() + "\n"
		} else {
			output += "成功\n"
		}
	}

	branchCmd := cmdWrapper.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	branchCmd.Dir = projectDir
	branchOut, branchErr := branchCmd.CombinedOutput()
	branch := strings.TrimSpace(string(branchOut))
	if branchErr != nil || branch == "" {
		return ResetResult{Success: false, Message: "未检测到分支名，请确保是有效的 Git 仓库", Output: string(branchOut)}
	}

	remoteCmd := cmdWrapper.Command("git", "remote", "get-url", "origin")
	remoteCmd.Dir = projectDir
	remoteOut, remoteErr := remoteCmd.CombinedOutput()
	remoteURL := strings.TrimSpace(string(remoteOut))
	if remoteErr != nil || remoteURL == "" {
		return ResetResult{Success: false, Message: "未检测到远程地址，请确保仓库已配置 remote origin", Output: string(remoteOut)}
	}

	gitDir := filepath.Join(projectDir, ".git")
	runStep("rm -rf .git", func() error { return os.RemoveAll(gitDir) })

	initCmd := cmdWrapper.Command("git", "init", "-b", branch)
	initCmd.Dir = projectDir
	initOut, initErr := initCmd.CombinedOutput()
	output += fmt.Sprintf("git init -b %s\n%s\n", branch, string(initOut))
	if initErr != nil {
		log.Printf("git init 失败: %v", initErr)
	}

	addCmd := cmdWrapper.Command("git", "add", ".")
	addCmd.Dir = projectDir
	addOut, addErr := addCmd.CombinedOutput()
	output += "git add .\n" + string(addOut) + "\n"
	if addErr != nil {
		log.Printf("git add 失败: %v", addErr)
	}

	commitMsg := "基本功能实现V1.0"
	commitCmd := cmdWrapper.Command("git", "commit", "-m", commitMsg)
	commitCmd.Dir = projectDir
	commitOut, commitErr := commitCmd.CombinedOutput()
	output += fmt.Sprintf("git commit -m \"%s\"\n%s\n", commitMsg, string(commitOut))
	if commitErr != nil {
		log.Printf("git commit 失败: %v", commitErr)
	}

	remoteAddCmd := cmdWrapper.Command("git", "remote", "add", "origin", remoteURL)
	remoteAddCmd.Dir = projectDir
	if _, err := remoteAddCmd.CombinedOutput(); err != nil {
		setUrlCmd := cmdWrapper.Command("git", "remote", "set-url", "origin", remoteURL)
		setUrlCmd.Dir = projectDir
		setUrlOut, setUrlErr := setUrlCmd.CombinedOutput()
		output += fmt.Sprintf("git remote set-url origin %s\n%s\n", remoteURL, string(setUrlOut))
		if setUrlErr != nil {
			log.Printf("git remote set-url 失败: %v", setUrlErr)
		}
	} else {
		output += fmt.Sprintf("git remote add origin %s\n成功\n", remoteURL)
	}

	pushCmd := cmdWrapper.Command("git", "push", "-f", "-u", "origin", branch)
	pushCmd.Dir = projectDir
	pushOut, pushErr := pushCmd.CombinedOutput()
	output += fmt.Sprintf("git push -f -u origin %s\n%s\n", branch, string(pushOut))
	if pushErr != nil {
		log.Printf("git push 失败: %v", pushErr)
	}

	log.Printf("重置完成, 输出:\n%s", output)
	return ResetResult{Success: true, Message: "重置完成", Output: output}
}

// SoftReset 软重置：把本地未推送到远程的提交合并为一次提交。
func (a *App) SoftReset(req SoftResetReq) SoftResetResult {
	projectDir := req.Path
	commitMsg := strings.TrimSpace(req.Message)
	if commitMsg == "" {
		commitMsg = "合并本地未推送的提交"
	}
	log.Printf("开始软重置, 目录: %s, 提交信息: %s", projectDir, commitMsg)

	var output string
	runGit := func(args ...string) (string, error) {
		c := cmdWrapper.Command("git", args...)
		c.Dir = projectDir
		out, err := c.CombinedOutput()
		return strings.TrimSpace(string(out)), err
	}

	branchOut, err := runGit("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil || branchOut == "" || branchOut == "HEAD" {
		return SoftResetResult{Success: false, Message: "未检测到当前分支，请确保是有效的 Git 仓库", Output: branchOut}
	}
	branch := branchOut
	output += fmt.Sprintf("当前分支: %s\n", branch)

	fetchOut, fetchErr := runGit("fetch", "origin", branch)
	output += fmt.Sprintf("git fetch origin %s\n%s\n", branch, fetchOut)
	if fetchErr != nil {
		log.Printf("git fetch 失败: %v", fetchErr)
	}

	remoteRef := fmt.Sprintf("origin/%s", branch)
	if existsOut, _ := runGit("rev-parse", "--verify", "--quiet", remoteRef); existsOut == "" {
		return SoftResetResult{Success: false, Message: fmt.Sprintf("未找到远程分支 %s，无法软重置", remoteRef), Output: output}
	}

	resetOut, resetErr := runGit("reset", "--soft", remoteRef)
	output += fmt.Sprintf("git reset --soft %s\n%s\n", remoteRef, resetOut)
	if resetErr != nil {
		log.Printf("git reset --soft 失败: %v", resetErr)
		return SoftResetResult{Success: false, Message: "软重置失败，请确认本地有未推送的提交", Output: output}
	}

	commitOut, commitErr := runGit("commit", "-m", commitMsg)
	output += fmt.Sprintf("git commit -m \"%s\"\n%s\n", commitMsg, commitOut)
	if commitErr != nil {
		log.Printf("git commit 失败: %v", commitErr)
		return SoftResetResult{Success: false, Message: "提交失败，可能没有可提交的内容", Output: output}
	}

	log.Printf("软重置完成, 输出:\n%s", output)
	return SoftResetResult{Success: true, Message: "合并成功", Output: output}
}

// PackageProject 在指定目录执行 wails build，把产物复制到目标目录。
// 目标目录硬编码为 E:\application\我的工具箱（沿用原 wcj-go-git 行为，可后续调整）。
func (a *App) PackageProject(req PackageReq) PackageResult {
	projectDir := req.Path
	log.Printf("开始打包项目, 目录: %s", projectDir)

	wailsConfig := filepath.Join(projectDir, "wails.json")
	if _, err := os.Stat(wailsConfig); os.IsNotExist(err) {
		return PackageResult{Success: false, Message: "不是 Wails 项目目录"}
	}

	var stdout, stderr bytes.Buffer
	cmd := cmdWrapper.Command("wails", "build")
	cmd.Dir = projectDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	log.Printf("执行命令: wails build, 工作目录: %s", projectDir)
	err := cmd.Run()
	output := stdout.String() + stderr.String()
	log.Printf("打包输出:\n%s", output)
	if err != nil {
		log.Printf("打包失败: %v", err)
		return PackageResult{Success: false, Message: "打包失败: " + err.Error(), Output: output}
	}

	outputDir := filepath.Join(projectDir, "build", "bin")
	var exeFile string
	if entries, err := os.ReadDir(outputDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".exe") {
				exeFile = entry.Name()
				break
			}
		}
	}

	if exeFile == "" {
		return PackageResult{Success: true, Message: "打包成功，但未找到 exe 文件", Output: output, OutputDir: outputDir}
	}

	targetDir := `E:\application\我的工具箱`
	targetPath := filepath.Join(targetDir, exeFile)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return PackageResult{Success: false, Message: "打包成功但创建目标目录失败: " + err.Error(), Output: output, OutputDir: outputDir}
	}
	sourcePath := filepath.Join(outputDir, exeFile)
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		return PackageResult{Success: false, Message: "打包成功但读取 exe 失败: " + err.Error(), Output: output, OutputDir: outputDir}
	}
	if err := os.WriteFile(targetPath, data, 0755); err != nil {
		return PackageResult{Success: false, Message: "打包成功但复制到目标目录失败: " + err.Error(), Output: output, OutputDir: outputDir}
	}

	log.Printf("打包成功, 已复制到: %s", targetPath)
	return PackageResult{Success: true, Message: "打包成功，已复制到: " + targetPath, Output: output, OutputDir: targetPath}
}
