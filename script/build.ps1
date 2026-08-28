# 一键构建脚本（PowerShell 版本）：自动注入构建时间到 main.BuildTime
# 用法：在仓库根目录执行 script/build.ps1，或 powershell -ExecutionPolicy Bypass -File script/build.ps1

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Definition
$repoRoot = Resolve-Path (Join-Path $scriptDir "..")
Set-Location $repoRoot

$timestamp = Get-Date -Format "yyyy-MM-ddTHH:mm:ss"
Write-Host "[build] BuildTime = $timestamp"
wails build -ldflags "-X wcj-go-text/app.BuildTime=$timestamp"