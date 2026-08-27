package main

import (
	"embed"
	"log"
	"os"
	"path/filepath"
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

// 日志目录：所有 log.Printf / fmt.Println / Wails 内部 stderr 输出统一写入此目录下的 wcj-go-text.log
const logDir = `C:\Users\wangchaojun\AppData\Local\wtools`

// init 注入嵌入式资源到 app 包，并保留 BuildTime 符号；
// 同时记录启动起点，使 GetStartupTime 能覆盖 Go runtime + 全部包 init + Wails 启动 + 前端加载耗时。
func init() {
	app.Assets = frontendDist
	app.TplFS = tplDist
	app.KeepBuildTimeAlive()
	app.SetStartupStart(time.Now().UnixNano())
	setupFileLogging()
}

// setupFileLogging 将全部日志（标准 log、fmt.Println、os.Stderr）重定向到日志文件。
// 失败时静默回退到 stderr，不影响主流程。
func setupFileLogging() {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return
	}
	logPath := filepath.Join(logDir, "wcj-go-text.log")
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	log.SetOutput(f)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
	log.SetPrefix("")
	os.Stdout = f
	os.Stderr = f
	log.Printf("=== session start %s ===", time.Now().Format(time.RFC3339))
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
		HideWindowOnClose: false,
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
