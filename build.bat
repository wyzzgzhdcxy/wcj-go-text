@echo off
REM 一键构建脚本：自动把当前时间注入到 main.BuildTime，再调用 wails build
REM 用法：双击 build.bat 或在命令行执行 .\build.bat

setlocal
for /f "tokens=2 delims==" %%I in ('wmic os get localdatetime /value ^| findstr /r "="') do set DATETIME=%%I
set BUILD_TIME=%DATETIME:~0,4%-%DATETIME:~4,2%-%DATETIME:~6,2%T%DATETIME:~8,2%:%DATETIME:~10,2%:%DATETIME:~12,2%

echo [build] BuildTime = %BUILD_TIME%
wails build -ldflags "-X main.BuildTime=%BUILD_TIME%"
endlocal