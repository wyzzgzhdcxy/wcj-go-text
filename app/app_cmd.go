package app

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"wcj-go-text/golang/cmdWrapper"
	"wcj-go-text/golang/sqllite"
	"wcj-go-text/golang/utils"
)

type EnvCheckResult struct {
	Name    string `json:"name"`
	Found   bool   `json:"found"`
	Version string `json:"version"`
	Path    string `json:"path"`
}

func (a *App) ExecCommand(command string, dir string) string {
	cmd := cmdWrapper.Command("cmd", "/c", command)
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
		cmd := cmdWrapper.Command(c.cmd, c.args...)
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
	cmd := cmdWrapper.Command("shutdown", "/s", "/t", strconv.Itoa(seconds))
	err := cmd.Run()
	if err != nil {
		return "关机设置失败: " + err.Error()
	}
	return fmt.Sprintf("系统将在 %d 秒后关机", seconds)
}

func (a *App) ShutdownAt(timeStr string) string {
	cmd := cmdWrapper.Command("shutdown", "/s", "/t", timeStr)
	err := cmd.Run()
	if err != nil {
		return "关机设置失败: " + err.Error()
	}
	return "系统将在 " + timeStr + " 秒后关机"
}

func (a *App) CancelShutdown() string {
	var out bytes.Buffer
	cmd := cmdWrapper.Command("shutdown", "/a")
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
	output, err := cmdWrapper.RunWithDirAndOutput(dir, name, args...)
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

// CmdCommand 命令行管理（与 sqllite.CmdCommand 字段对齐，供 Wails 绑定）
type CmdCommand = sqllite.CmdCommand

func (a *App) ListCmdCommands() []CmdCommand {
	cmds, err := sqllite.ListCmdCommands()
	if err != nil {
		return []CmdCommand{}
	}
	return cmds
}

func (a *App) AddCmdCommand(cmd CmdCommand) (int, error) {
	return sqllite.AddCmdCommand(cmd)
}

func (a *App) UpdateCmdCommand(cmd CmdCommand) error {
	return sqllite.UpdateCmdCommand(cmd)
}

func (a *App) DeleteCmdCommand(id int) error {
	return sqllite.DeleteCmdCommand(id)
}

// ExecuteCmdCommand 执行命令行
func (a *App) ExecuteCmdCommand(id int, command string) (string, error) {
	storedCommand, err := sqllite.GetCmdCommandByID(id)
	if err != nil {
		return "", fmt.Errorf("查询命令行失败: %v", err)
	}
	cmdStr := storedCommand
	if command != "" {
		cmdStr = command
	}
	if cmdStr == "" {
		return "", fmt.Errorf("命令为空")
	}
	cmd := cmdWrapper.Command("cmd", "/C", cmdStr)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("执行命令失败: %v", err)
	}
	return string(output), nil
}

// ExecuteRawCommand 执行原始命令
func (a *App) ExecuteRawCommand(command string) (string, error) {
	cmd := cmdWrapper.Command("cmd", "/C", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("执行命令失败: %v", err)
	}
	return string(output), nil
}

// OpenEnvironmentEditor 打开系统环境变量编辑器
func (a *App) OpenEnvironmentEditor() {
	cmd := cmdWrapper.Command("rundll32", "sysdm.cpl,EditEnvironmentVariables")
	cmd.Run()
}

// EnvVar 环境变量
type EnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// GetUserEnvVars 获取用户环境变量
func (a *App) GetUserEnvVars() []EnvVar {
	cmd := cmdWrapper.Command("cmd", "/C", "set")
	output, _ := cmd.Output()
	var result []EnvVar
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if idx := strings.Index(line, "="); idx > 0 {
			result = append(result, EnvVar{Name: line[:idx], Value: line[idx+1:]})
		}
	}
	return result
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

// queryRegistryEnvVars 通过 reg query 读取注册表项下的所有 (默认) 值与显式值，
// 返回 name=value 形式的 EnvVar 列表；PATH/Path 等 REG_EXPAND_SZ 自动展开。
func queryRegistryEnvVars(regPath string) []EnvVar {
	cmd := cmdWrapper.Command("reg", "query", regPath)
	output, err := cmd.Output()
	if err != nil {
		return nil
	}
	var result []EnvVar
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		idx := strings.Index(line, "REG_")
		if idx <= 0 {
			continue
		}
		kv := strings.TrimSpace(line[idx:])
		eq := strings.Index(kv, "=")
		if eq <= 0 {
			continue
		}
		name := strings.TrimSpace(kv[:eq])
		value := strings.TrimSpace(kv[eq+1:])
		if name == "" {
			continue
		}
		result = append(result, EnvVar{Name: name, Value: value})
	}
	return result
}

// DeleteUserEnvVar 删除用户环境变量
func (a *App) DeleteUserEnvVar(name string) error {
	cmd := cmdWrapper.Command("reg", "delete", `HKCU\Environment`, "/v", name, "/f")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("删除失败: %s", string(output))
	}
	return nil
}

// DeleteSystemEnvVar 删除系统环境变量
func (a *App) DeleteSystemEnvVar(name string) error {
	cmd := cmdWrapper.Command("reg", "delete", `HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Environment`, "/v", name, "/f")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("删除失败: %s", string(output))
	}
	return nil
}

// GetPathInfo 获取 PATH 环境变量信息
func (a *App) GetPathInfo() map[string]string {
	cmd := cmdWrapper.Command("cmd", "/C", "echo %PATH%")
	output, _ := cmd.Output()
	path := strings.TrimSpace(string(output))
	return map[string]string{"path": path}
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

// GetEnvBackups 获取环境变量备份列表
func (a *App) GetEnvBackups() []string {
	return utils.GetBackupFileList()
}

// RestoreEnvBackup 恢复环境变量备份
func (a *App) RestoreEnvBackup(backupFile string) error {
	if msg := utils.RestoreEnvVars(backupFile); msg != "" {
		return fmt.Errorf("%s", msg)
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
