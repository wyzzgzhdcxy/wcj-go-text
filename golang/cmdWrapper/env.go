package cmdWrapper

import "github.com/wyzzgzhdcxy/wcj-go-common/core"

// OenEnvironmentWindows 打开环境变量窗口
func OenEnvironmentWindows() {
	core.ExecuteCommand("rundll32", "sysdm.cpl,EditEnvironmentVariables")
}

// TopWindows 置顶指定名称的窗口
func TopWindows(name string) {
	core.ExecuteCommand("TopMost_x64", "/S", name)
}
