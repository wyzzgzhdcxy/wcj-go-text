package app

import (
	"embed"
	"fmt"
	"strconv"

	"wcj-go-text/golang/sqllite"

	"github.com/wyzzgzhdcxy/wcj-go-common/core"
)

// Assets 持有嵌入式资源（来自 main.go），用于读取 system.config.json

// WindowState 窗口状态
type WindowState struct {
	X      int
	Y      int
	Width  int
	Height int
}

func saveSetting(key, value string) {
	_ = sqllite.SaveSetting(key, value)
}

func getSetting(key string) string {
	return sqllite.GetSetting(key)
}

// SaveWindowState 保存窗口位置与尺寸。
func (a *App) SaveWindowState(x, y, width, height int) {
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if width < 600 {
		width = 600
	}
	if height < 400 {
		height = 400
	}
	saveSetting("window.x", fmt.Sprintf("%d", x))
	saveSetting("window.y", fmt.Sprintf("%d", y))
	saveSetting("window.width", fmt.Sprintf("%d", width))
	saveSetting("window.height", fmt.Sprintf("%d", height))
	saveSetting("window.x_saved", "1")
}

// GetWindowState 读取上次的窗口位置与尺寸。
func (a *App) GetWindowState() WindowState {
	ws := WindowState{}
	xSaved := getSetting("window.x_saved")
	if xSaved != "1" {
		return ws
	}
	x := getSetting("window.x")
	if x == "" {
		return ws
	}
	ws.X, _ = strconv.Atoi(x)
	ws.Y, _ = strconv.Atoi(getSetting("window.y"))
	ws.Width, _ = strconv.Atoi(getSetting("window.width"))
	ws.Height, _ = strconv.Atoi(getSetting("window.height"))
	if ws.X < 0 {
		ws.X = 0
	}
	if ws.Y < 0 {
		ws.Y = 0
	}
	if ws.Width < 600 {
		ws.Width = 600
	}
	if ws.Height < 400 {
		ws.Height = 400
	}
	return ws
}

// AddConfigValue 添加配置
func (a *App) AddConfigValue(key, value string) string {
	if err := sqllite.SaveSetting(key, value); err != nil {
		return "Failed: " + err.Error()
	}
	return "OK"
}

// DeleteConfig 删除配置
func (a *App) DeleteConfig(key string) {
	_ = sqllite.DeleteSetting(key)
}

// ListConfig 获取配置列表
func (a *App) ListConfig(prefix string) map[string]string {
	return sqllite.GetSettingWithPrefix(prefix)
}

// GetConfigKey 获取配置（先查 DB，再读 assets/config/system.config.json）
func (a *App) GetConfigKey(key string) string {
	if v := sqllite.GetSystemSetting(key); v != "" {
		return v
	}
	if Assets != (embed.FS{}) {
		content, _ := Assets.ReadFile("config/system.config.json")
		var myMap map[string]string
		contentByte := []byte(content)
		core.JsonToObject(&contentByte, &myMap)
		return myMap[key]
	}
	return ""
}