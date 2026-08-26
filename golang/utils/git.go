package utils

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
)

func GetRsaPrivateKeyPath() string {
	userDir, _ := os.UserHomeDir()
	return userDir + "\\.ssh\\id_rsa"
}

func hideWindow(cmd *exec.Cmd) {
	if runtime.GOOS == "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	}
}

// GitClone 克隆仓库（使用命令行 git）
func GitClone(repoURL, targetDir string) error {
	cmd := exec.Command("git", "clone", repoURL, targetDir)
	cmd.Env = append(os.Environ(), "GIT_SSH_COMMAND=ssh -i "+GetRsaPrivateKeyPath())
	hideWindow(cmd)
	return cmd.Run()
}

// GitPush 推送（使用命令行 git）
func GitPush(repoPath string) error {
	cmd := exec.Command("git", "push")
	cmd.Dir = repoPath
	cmd.Env = append(os.Environ(), "GIT_SSH_COMMAND=ssh -i "+GetRsaPrivateKeyPath())
	hideWindow(cmd)
	return cmd.Run()
}

// GitPull 拉取（使用命令行 git）
func GitPull(repoPath string) string {
	cmd := exec.Command("git", "pull")
	cmd.Dir = repoPath
	cmd.Env = append(os.Environ(), "GIT_SSH_COMMAND=ssh -i "+GetRsaPrivateKeyPath())
	hideWindow(cmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return strings.TrimSpace(string(output))
	}
	return ""
}

// GitAdd 添加文件（使用命令行 git）
func GitAdd(repoPath string) error {
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = repoPath
	hideWindow(cmd)
	return cmd.Run()
}

// AddCommitPush 提交并推送（使用命令行 git）
func AddCommitPush(repoPath string) string {
	// git add .
	if err := GitAdd(repoPath); err != nil {
		return fmt.Sprintf("添加文件失败: %v", err)
	}

	// git commit -m "wtools update"
	cmd := exec.Command("git", "commit", "-m", "wtools update")
	cmd.Dir = repoPath
	hideWindow(cmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return strings.TrimSpace(string(output))
	}

	// git push
	cmd = exec.Command("git", "push")
	cmd.Dir = repoPath
	cmd.Env = append(os.Environ(), "GIT_SSH_COMMAND=ssh -i "+GetRsaPrivateKeyPath())
	hideWindow(cmd)
	output, err = cmd.CombinedOutput()
	if err != nil {
		return strings.TrimSpace(string(output))
	}

	return ""
}

// PullAddCommitPush 拉取后提交推送
func PullAddCommitPush(repoPath string) string {
	if len(GitPull(repoPath)) == 0 {
		return AddCommitPush(repoPath)
	}
	return "拉取仓库失败, repoPath:" + repoPath
}