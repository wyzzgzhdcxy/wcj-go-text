package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

// init 引用 BuildTime，防止链接器 dead-code 优化丢弃 -X 注入的值。
// （通过 runtime.SetFinalizer 引用，使符号必须保留）
func init() {
	runtimeSetFinalizerForBuildTime()
	assetsFS = assets
}

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:  "文本工具箱",
		Width:  1180,
		Height: 860,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId:               "wcj-go-text-e3f2a1b4-8c9d-4e5f-a6b7",
			OnSecondInstanceLaunch: app.onSecondInstanceLaunch,
		},
		OnStartup: app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
