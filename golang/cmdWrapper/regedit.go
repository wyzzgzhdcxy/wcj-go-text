package cmdWrapper

import (
	"fmt"
	"strings"

	"github.com/wyzzgzhdcxy/wcj-go-common/core"
)

func OpenRegistryStartup() error {
	fmt.Println("打开注册表启动项位置")

	switch core.GOOS() {
	case "windows":
		// 使用 regjump 打开注册表并定位到指定项
		cmd := core.Command("regjump", `HKEY_CURRENT_USER\Software\Microsoft\Windows\CurrentVersion\Run`)
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
		return fmt.Errorf("不支持的平台: %s", core.GOOS())
	}
}

// OpenExplorer 用系统文件管理器打开目录。
// Windows 传入的 path 要反斜杠（C:\\tmp\\wtools\\png），其余平台保持原样。
func OpenExplorer(path string) error {
	if core.IsWindows() {
		path = strings.ReplaceAll(path, "/", "\\")
	}
	fmt.Println("open explorer", path)
	switch core.GOOS() {
	case "windows":
		fmt.Println("打开资源管理器" + path)
		return execCommandVisible("explorer", path)
	case "darwin":
		return execCommandVisible("open", path)
	case "linux":
		return execCommandVisible("xdg-open", path)
	default:
		return fmt.Errorf("unsupported platform")
	}
}

// execCommandVisible 启动一个允许显示窗口的子进程。
// 用于 explorer 等 GUI 程序：被 HideWindow 的 explorer 会静默退出而不弹出窗口。
func execCommandVisible(cmdstr, args string) error {
	cmd := core.CommandVisible(cmdstr, args)
	cmd.Stdout = nil
	cmd.Stderr = nil
	err := cmd.Start()
	return err
}
