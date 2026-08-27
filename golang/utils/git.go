package utils

import (
	"fmt"
	"os"
	"strings"

	"wcj-go-text/golang/cmdWrapper"
)

func GetRsaPrivateKeyPath() string {
	userDir, _ := os.UserHomeDir()
	return userDir + "\\.ssh\\id_rsa"
}

// GitClone 克隆仓库（使用命令行 git）
func GitClone(repoURL, targetDir string) error {
	cmd := cmdWrapper.Command("git", "clone", repoURL, targetDir)
	cmd.Env = append(os.Environ(), "GIT_SSH_COMMAND=ssh -i "+GetRsaPrivateKeyPath())
	return cmd.Run()
}

// GitPush 推送（使用命令行 git）
func GitPush(repoPath string) error {
	cmd := cmdWrapper.Command("git", "push")
	cmd.Dir = repoPath
	cmd.Env = append(os.Environ(), "GIT_SSH_COMMAND=ssh -i "+GetRsaPrivateKeyPath())
	return cmd.Run()
}

// GitPull 拉取（使用命令行 git）
func GitPull(repoPath string) string {
	cmd := cmdWrapper.Command("git", "pull")
	cmd.Dir = repoPath
	cmd.Env = append(os.Environ(), "GIT_SSH_COMMAND=ssh -i "+GetRsaPrivateKeyPath())
	output, err := cmd.CombinedOutput()
	if err != nil {
		return strings.TrimSpace(string(output))
	}
	return ""
}

// GitAdd 添加文件（使用命令行 git）
func GitAdd(repoPath string) error {
	cmd := cmdWrapper.Command("git", "add", ".")
	cmd.Dir = repoPath
	return cmd.Run()
}

// AddCommitPush 提交并推送（使用命令行 git）
func AddCommitPush(repoPath string) string {
	// git add .
	if err := GitAdd(repoPath); err != nil {
		return fmt.Sprintf("添加文件失败: %v", err)
	}

	// git commit -m "wtools update"
	cmd := cmdWrapper.Command("git", "commit", "-m", "wtools update")
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return strings.TrimSpace(string(output))
	}

	// git push
	cmd = cmdWrapper.Command("git", "push")
	cmd.Dir = repoPath
	cmd.Env = append(os.Environ(), "GIT_SSH_COMMAND=ssh -i "+GetRsaPrivateKeyPath())
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