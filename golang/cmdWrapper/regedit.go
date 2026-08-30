package cmdWrapper

import (
	"fmt"
	"strings"

	"github.com/wyzzgzhdcxy/wcj-go-common/core"
)

// regeditAppletsKey regedit 用 LastKey 值记忆下次打开时定位的注册表项
const regeditAppletsKey = `HKCU\Software\Microsoft\Windows\CurrentVersion\Applets\Regedit`

func OpenRegistryStartup() error {
	fmt.Println("打开注册表启动项位置")

	switch core.GOOS() {
	case "windows":
		return openRegeditAt(`HKEY_CURRENT_USER\Software\Microsoft\Windows\CurrentVersion\Run`)
	case "darwin":
		return fmt.Errorf("macOS 暂不支持")
	case "linux":
		return fmt.Errorf("Linux 暂不支持")
	default:
		return fmt.Errorf("不支持的平台: %s", core.GOOS())
	}
}

// openRegeditAt 打开注册表编辑器并定位到 targetKey。
// 不能用 regjump 跳转：regjump 被 HideWindow 隐藏启动后，会把 SW_HIDE 传播给
// 它拉起的 regedit，导致 regedit 进程在后台隐身运行，窗口永远不显示。
// 这里用 regedit 原生机制：先把目标键写入 LastKey（regedit 启动时定位到该键），
// 再以可见方式启动 regedit。
func openRegeditAt(targetKey string) error {
	// regedit 是单实例：已运行时再启动只会激活旧窗口、不会重新读 LastKey，先结束它
	_ = core.Command("taskkill", "/F", "/IM", "regedit.exe").Run()

	cmd := core.Command("reg", "add", regeditAppletsKey, "/v", "LastKey", "/d", targetKey, "/f")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("设置注册表定位失败: %v, %s", err, string(output))
	}

	editCmd := core.CommandVisible("regedit")
	if err := editCmd.Start(); err != nil {
		return fmt.Errorf("打开注册表编辑器失败: %v", err)
	}
	return nil
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
