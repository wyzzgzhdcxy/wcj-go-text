# 一键构建脚本（PowerShell 版本）：自动注入构建时间到 main.BuildTime，并将产物复制到工具箱目录
# 用法：在仓库根目录执行 script/build-local.ps1，或 powershell -ExecutionPolicy Bypass -File script/build-local.ps1

$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Definition
$repoRoot = Resolve-Path (Join-Path $scriptDir "..")
Set-Location $repoRoot

$timestamp = Get-Date -Format "yyyy-MM-ddTHH:mm:ss"
Write-Host "[build] BuildTime = $timestamp"

wails build -ldflags "-X wcj-go-text/app.BuildTime=$timestamp"
if ($LASTEXITCODE -ne 0) {
    Write-Host "[build] wails build failed, aborting."
    exit 1
}

$src = "build\bin\文本工具.exe"
$dst = "E:\application\我的工具箱\文本工具.exe"
if (Test-Path -LiteralPath $src) {
    New-Item -ItemType Directory -Path (Split-Path $dst) -Force | Out-Null
    Copy-Item -LiteralPath $src -Destination $dst -Force
    Write-Host "[build] Copied"
} else {
    Write-Host "[build] Source exe not found"
}
