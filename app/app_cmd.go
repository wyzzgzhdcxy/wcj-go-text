package app

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/wyzzgzhdcxy/wcj-go-common/core"
	"golang.org/x/sys/windows/registry"
	"wcj-go-text/golang/cmdWrapper"
	"wcj-go-text/golang/sqllite"
)

type EnvCheckResult struct {
	Name    string `json:"name"`
	Found   bool   `json:"found"`
	Version string `json:"version"`
	Path    string `json:"path"`
}

func (a *App) ExecCommand(command string, dir string) string {
	cmd := core.Command("cmd", "/c", command)
	if dir != "" {
		cmd.Dir = dir
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Error: %v\n%s", err, string(output))
	}
	return string(output)
}

func (a *App) StartEnvironment() {
	cmdWrapper.OenEnvironmentWindows()
}

func (a *App) CpuZ() {
	cmdWrapper.TopWindows("CPU-Z")
}

func (a *App) CheckEnvironment() []EnvCheckResult {
	results := []EnvCheckResult{}
	checkers := []struct {
		name string
		cmd  string
		args []string
	}{
		{"Java", "java", []string{"-version"}},
		{"Maven", "mvn", []string{"--version"}},
		{"Gradle", "gradle", []string{"--version"}},
		{"Python", "python", []string{"--version"}},
		{"NodeJS", "node", []string{"--version"}},
		{"Golang", "go", []string{"version"}},
	}
	for _, c := range checkers {
		result := EnvCheckResult{Name: c.name}
		cmd := core.Command(c.cmd, c.args...)
		output, err := cmd.CombinedOutput()
		if err == nil {
			result.Found = true
			result.Version = strings.TrimSpace(strings.Split(string(output), "\n")[0])
			result.Path, _ = exec.LookPath(c.cmd)
		}
		results = append(results, result)
	}
	return results
}

func (a *App) ShutdownAfterSeconds(seconds int) string {
	cmd := core.Command("shutdown", "/s", "/t", strconv.Itoa(seconds))
	err := cmd.Run()
	if err != nil {
		return "关机设置失败: " + err.Error()
	}
	return fmt.Sprintf("系统将在 %d 秒后关机", seconds)
}

func (a *App) ShutdownAt(timeStr string) string {
	cmd := core.Command("shutdown", "/s", "/t", timeStr)
	err := cmd.Run()
	if err != nil {
		return "关机设置失败: " + err.Error()
	}
	return "系统将在 " + timeStr + " 秒后关机"
}

func (a *App) CancelShutdown() string {
	var out bytes.Buffer
	cmd := core.Command("shutdown", "/a")
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			code := exitErr.ExitCode()
			msg := strings.TrimSpace(out.String())
			if code == 1110 || strings.Contains(msg, "1110") ||
				strings.Contains(msg, "没有进行关机") ||
				strings.Contains(msg, "没有关机") ||
				strings.Contains(msg, "No shutdown was in progress") {
				return "没有定时关机任务"
			}
			if msg != "" {
				return "取消关机失败: " + msg
			}
		}
		return "取消关机失败: " + err.Error()
	}
	return "取消成功"
}

func (a *App) RunCmdWithDir(dir, name string, args []string) string {
	output, err := core.RunWithDirAndOutput(dir, name, args...)
	if err != nil {
		return fmt.Sprintf("Error: %v\n%s", err, output)
	}
	return output
}

func (a *App) OpenRegistryStartup() error {
	return cmdWrapper.OpenRegistryStartup()
}

func (a *App) GetEnvVar(name string) string {
	return os.Getenv(name)
}

func (a *App) SetEnvVar(name, value string) error {
	return os.Setenv(name, value)
}

func (a *App) CmdRecord(command string) {
	_ = sqllite.RecordCmdHistory(command)
}

func (a *App) GetCmdHistory() []string {
	cmds, err := sqllite.GetCmdHistory()
	if err != nil {
		return nil
	}
	return cmds
}

func (a *App) RecordToolUsage(link string) {
	_ = sqllite.RecordToolUsage(link)
}

func (a *App) GetToolUsageStats() map[string]int {
	stats, err := sqllite.GetToolUsageStats()
	if err != nil {
		return nil
	}
	return stats
}

func (a *App) NowTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// OpenEnvironmentEditor 打开系统环境变量编辑器
func (a *App) OpenEnvironmentEditor() {
	cmd := core.Command("rundll32", "sysdm.cpl,EditEnvironmentVariables")
	cmd.Run()
}

// EnvVar 环境变量
type EnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// GetUserEnvVars 获取用户环境变量（直接读注册表 HKCU\Environment，与系统
// 环境变量编辑器一致。cmd set 返回的是进程合并后的环境，会混入系统变量和
// 进程自身的变量，导致列表比真实的用户变量多很多）
func (a *App) GetUserEnvVars() []EnvVar {
	return queryRegistryEnvVars(`HKCU\Environment`)
}

// GetProcessEnvVars 获取进程环境变量
func (a *App) GetProcessEnvVars() []EnvVar {
	var result []EnvVar
	for _, e := range os.Environ() {
		if idx := strings.Index(e, "="); idx > 0 {
			result = append(result, EnvVar{Name: e[:idx], Value: e[idx+1:]})
		}
	}
	return result
}

// GetSystemEnvVars 获取系统环境变量（直接读注册表 HKLM\...\Environment，
// 不依赖 cmd set 的合并结果，避免原实现中 seen 全命中导致返回空列表的 bug）。
func (a *App) GetSystemEnvVars() []EnvVar {
	return queryRegistryEnvVars(`HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Environment`)
}

// queryRegistryEnvVars 通过注册表 API 直接读取项下所有字符串值
// （REG_SZ / REG_EXPAND_SZ，保持未展开形式）。
// 不用 reg query 子进程：其输出为控制台 OEM 代码页（中文系统为 GBK），
// 直接按 UTF-8 解析会导致含中文的变量名/值显示乱码。
func queryRegistryEnvVars(regPath string) []EnvVar {
	root, sub := splitRegPath(regPath)
	if root == 0 {
		return nil
	}
	key, err := registry.OpenKey(root, sub, registry.QUERY_VALUE)
	if err != nil {
		return nil
	}
	defer key.Close()
	names, err := key.ReadValueNames(-1)
	if err != nil {
		return nil
	}
	var result []EnvVar
	for _, name := range names {
		if name == "" {
			continue
		}
		val, _, err := key.GetStringValue(name)
		if err != nil {
			continue // 跳过非字符串类型的值
		}
		result = append(result, EnvVar{Name: name, Value: val})
	}
	return result
}

// splitRegPath 解析 "HKCU\Environment" 形式的注册表路径为根键 + 子键
func splitRegPath(regPath string) (registry.Key, string) {
	i := strings.Index(regPath, "\\")
	if i < 0 {
		return 0, ""
	}
	sub := regPath[i+1:]
	switch strings.ToUpper(regPath[:i]) {
	case "HKCU", "HKEY_CURRENT_USER":
		return registry.CURRENT_USER, sub
	case "HKLM", "HKEY_LOCAL_MACHINE":
		return registry.LOCAL_MACHINE, sub
	}
	return 0, ""
}

// DeleteUserEnvVar 删除用户环境变量
func (a *App) DeleteUserEnvVar(name string) error {
	cmd := core.Command("reg", "delete", `HKCU\Environment`, "/v", name, "/f")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("删除失败: %s", string(output))
	}
	return nil
}

// DeleteSystemEnvVar 删除系统环境变量
func (a *App) DeleteSystemEnvVar(name string) error {
	cmd := core.Command("reg", "delete", `HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Environment`, "/v", name, "/f")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("删除失败: %s", string(output))
	}
	return nil
}

// PathEntry PATH 环境变量条目
type PathEntry struct {
	Path        string `json:"path"`
	Source      string `json:"source"`
	IsDuplicate bool   `json:"isDuplicate"`
}

// GetPathInfo 获取 PATH 条目列表（含来源与重复标记），供前端 PATH 变量页展示
func (a *App) GetPathInfo() []PathEntry {
	userPath := registryEnvValue(`HKCU\Environment`, "Path")
	systemPath := registryEnvValue(`HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Environment`, "Path")
	processPath := os.Getenv("PATH")

	splitPath := func(raw string) []string {
		var out []string
		for _, p := range strings.Split(raw, ";") {
			if p = strings.TrimSpace(p); p != "" {
				out = append(out, p)
			}
		}
		return out
	}
	// 以进程 PATH（合并后的实际生效值）统计各目录出现次数，出现多次即视为重复
	counts := make(map[string]int)
	norm := func(p string) string {
		return strings.ToLower(strings.TrimSuffix(strings.TrimSuffix(p, "\\"), "/"))
	}
	for _, p := range splitPath(processPath) {
		counts[norm(p)]++
	}

	var result []PathEntry
	appendEntries := func(raw, source string) {
		for _, p := range splitPath(raw) {
			result = append(result, PathEntry{Path: p, Source: source, IsDuplicate: counts[norm(p)] > 1})
		}
	}
	appendEntries(userPath, "User")
	appendEntries(systemPath, "System")
	appendEntries(processPath, "Process")
	return result
}

// registryEnvValue 读取注册表环境变量项下指定名称的值
func registryEnvValue(regPath, name string) string {
	for _, v := range queryRegistryEnvVars(regPath) {
		if strings.EqualFold(v.Name, name) {
			return v.Value
		}
	}
	return ""
}

// SetEnvVars 批量设置环境变量
func (a *App) SetEnvVars(vars []EnvVar) error {
	for _, v := range vars {
		if err := os.Setenv(v.Name, v.Value); err != nil {
			return err
		}
	}
	return nil
}

// EnvBackupInfo 环境变量备份信息
type EnvBackupInfo struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Modified string `json:"modified"`
	Size     int64  `json:"size"`
}

// envBackupRoot 环境变量备份根目录：%LOCALAPPDATA%/wtools/data/env_backups
func envBackupRoot() string {
	return filepath.Join(sqllite.GetAppDataDir(), "env_backups")
}

// BackupUserEnvVars 备份用户环境变量（reg export HKCU\Environment），返回备份目录
func (a *App) BackupUserEnvVars() (string, error) {
	dir := filepath.Join(envBackupRoot(), "user_env_"+time.Now().Format("20060102_150405"))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	regFile := filepath.Join(dir, "user_env_backup.reg")
	cmd := core.Command("reg", "export", `HKEY_CURRENT_USER\Environment`, regFile, "/y")
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("备份失败: %s", strings.TrimSpace(string(output)))
	}
	return dir, nil
}

// GetEnvBackups 获取环境变量备份列表
func (a *App) GetEnvBackups() []EnvBackupInfo {
	entries, err := os.ReadDir(envBackupRoot())
	if err != nil {
		return nil
	}
	var result []EnvBackupInfo
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		info := EnvBackupInfo{Name: e.Name(), Path: filepath.Join(envBackupRoot(), e.Name())}
		if fi, err := e.Info(); err == nil {
			info.Modified = fi.ModTime().Format("2006-01-02 15:04:05")
		}
		if files, err := os.ReadDir(info.Path); err == nil {
			for _, f := range files {
				if fi, err := f.Info(); err == nil {
					info.Size += fi.Size()
				}
			}
		}
		result = append(result, info)
	}
	return result
}

// RestoreEnvBackup 恢复用户环境变量（reg import 备份目录中的 user_env_backup.reg）
func (a *App) RestoreEnvBackup(backupPath string) error {
	regFile := filepath.Join(backupPath, "user_env_backup.reg")
	if _, err := os.Stat(regFile); err != nil {
		return fmt.Errorf("备份文件不存在: %s", regFile)
	}
	cmd := core.Command("reg", "import", regFile)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("恢复失败: %s", strings.TrimSpace(string(output)))
	}
	return nil
}

// RemoveEnvVars 删除环境变量
func (a *App) RemoveEnvVars(names []string) error {
	for _, name := range names {
		if err := os.Unsetenv(name); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) BackUpEnvFiles() string {
	return ""
}

func (a *App) RestoreEnvFiles() string {
	return ""
}

func (a *App) RemoveEnvConfigs(envNames []string) error {
	return nil
}

func (a *App) BackupSystemEnv(backupFile string) error {
	return nil
}
