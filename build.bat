@echo off
REM Build script: inject current timestamp into main.BuildTime, then run wails build.
REM Usage: double-click build.bat or run .\build.bat from cmd.

setlocal
for /f "delims=" %%I in ('powershell -NoProfile -Command "Get-Date -Format 'yyyy-MM-ddTHH:mm:ss'"') do set BUILD_TIME=%%I

echo [build] BuildTime = %BUILD_TIME%
wails build -ldflags "-X main.BuildTime=%BUILD_TIME%"
endlocal