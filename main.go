package main

import (
	"embed"
	"time"

	"wcj-go-text/app"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var frontendDist embed.FS

//go:embed all:tpl
var tplDist embed.FS

// init 注入嵌入式资源到 app 包，并保留 BuildTime 符号；
// 同时记录启动起点，使 GetStartupTime 能覆盖 Go runtime + 全部包 init + Wails 启动 + 前端加载耗时。
func init() {
	app.Assets = frontendDist
	app.TplFS = tplDist
	app.KeepBuildTimeAlive()
	app.SetStartupStart(time.Now().UnixNano())
}

func main() {
	a := app.NewApp()

	err := wails.Run(&options.App{
		Title:  "文本工具箱",
		Width:  1180,
		Height: 860,
		AssetServer: &assetserver.Options{
			Assets: frontendDist,
		},
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId:               "wcj-go-text-e3f2a1b4-8c9d-4e5f-a6b7",
			OnSecondInstanceLaunch: a.OnSecondInstanceLaunch,
		},
		OnStartup: a.Startup,
		Bind: []interface{}{
			a,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
