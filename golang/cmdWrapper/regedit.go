package cmdWrapper

import (
	"fmt"
	"os/exec"
	rt "runtime"
	"strings"
)

func OpenRegistryStartup() error {
	fmt.Println("打开注册表启动项位置")

	switch rt.GOOS {
	case "windows":
		// 使用 regjump 打开注册表并定位到指定项
		cmd := exec.Command("regjump", `HKEY_CURRENT_USER\Software\Microsoft\Windows\CurrentVersion\Run`)
		cmd.Stdout = nil
		cmd.Stderr = nil

		if err := cmd.Start(); err != nil {
			fmt.Printf("打开注册表失败: %v\n", err)
			return fmt.Errorf("打开注册表编辑器失败: %v", err)
		}

		return nil

	case "darwin":
		return fmt.Errorf("macOS 暂不支持")
	case "linux":
		return fmt.Errorf("Linux 暂不支持")
	default:
		return fmt.Errorf("不支持的平台: %s", rt.GOOS)
	}
}

// OpenExplorer C:\\tmp\\wtools\\png 其中的path要反斜杠
func OpenExplorer(path string) error {
	path = strings.ReplaceAll(path, "/", "\\")
	fmt.Println("open explorer", path)
	switch rt.GOOS {
	case "windows":
		fmt.Println("打开资源管理器" + path)
		return execCommand("explorer", path)
	case "darwin":
		return execCommand("open", path)
	case "linux":
		return execCommand("xdg-open", path)
	default:
		return fmt.Errorf("unsupported platform")
	}
}

func execCommand(cmdstr, args string) error {
	cmd := exec.Command(cmdstr, args)
	// 不设置任何输出
	cmd.Stdout = nil
	cmd.Stderr = nil
	err := cmd.Start()
	return err
}
