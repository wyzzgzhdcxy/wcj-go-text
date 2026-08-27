# 一键构建脚本（PowerShell 版本）：自动注入构建时间到 main.BuildTime
# 用法：右键 PowerShell 执行 ./build.ps1，或 powershell -File build.ps1

$timestamp = Get-Date -Format "yyyy-MM-ddTHH:mm:ss"
Write-Host "[build] BuildTime = $timestamp"
wails build -ldflags "-X wcj-go-text/app.BuildTime=$timestamp"