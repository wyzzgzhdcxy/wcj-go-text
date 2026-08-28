@echo off
setlocal

REM 切到仓库根目录（脚本所在目录的上级）
cd /d "%~dp0\.."

REM 注入构建时间到 main.BuildTime，然后执行 wails build
for /f "usebackq delims=" %%I in (`powershell -NoProfile -Command "Get-Date -Format 'yyyy-MM-ddTHH:mm:ss'"`) do set BUILD_TIME=%%I

echo [build] BuildTime = %BUILD_TIME%
wails build -ldflags "-X wcj-go-text/app.BuildTime=%BUILD_TIME%"
if errorlevel 1 (
    echo [build] wails build failed, aborting.
    exit /b 1
)

REM 复制 exe 到 E:\application\我的工具箱（用 PowerShell 处理中文路径）
powershell -NoProfile -Command "$src = 'build\bin\文本工具.exe'; $dst = 'E:\application\我的工具箱\文本工具.exe'; if (Test-Path $src) { New-Item -ItemType Directory -Path (Split-Path $dst) -Force | Out-Null; Copy-Item -LiteralPath $src -Destination $dst -Force; Write-Host '[build] Copied' } else { Write-Host '[build] Source exe not found' }"

endlocal
