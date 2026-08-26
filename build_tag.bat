@echo off
setlocal enabledelayedexpansion

REM Change to script directory (repo root)
cd /d "%~dp0"

REM Ensure git is available
where git >nul 2>nul
if errorlevel 1 (
    echo [ERROR] git not found. Install Git first.
    exit /b 1
)

REM Find the highest v* tag (version-sorted, descending)
set "latest="
for /f "delims=" %%i in ('git tag --list "v*" --sort=-version:refname 2^>nul') do (
    set "latest=%%i"
    goto :found
)

:found
if not defined latest set "latest=v0.0.0"

REM Parse major.minor.patch from the latest tag
set "ver=%latest:v=%"
set "maj=0"
set "min=0"
set "pat=0"
for /f "tokens=1,2,3 delims=." %%a in ("%ver%") do (
    set "maj=%%a"
    set "min=%%b"
    set "pat=%%c"
)

REM Bump the patch version
set /a pat+=1
set "newtag=v%maj%.%min%.%pat%"

REM Guard against an already-existing tag
git rev-parse --verify refs/tags/%newtag% >nul 2>nul
if not errorlevel 1 (
    echo [ERROR] tag %newtag% already exists.
    exit /b 1
)

echo ============================================
echo   Latest tag : %latest%
echo   New release: %newtag%
echo ============================================
echo.

git tag %newtag%
if errorlevel 1 (
    echo [ERROR] failed to create tag %newtag%
    exit /b 1
)

git push origin %newtag%
if errorlevel 1 (
    echo [ERROR] failed to push tag %newtag%
    exit /b 1
)

echo.
echo [OK] Tag %newtag% pushed. GitHub Actions is building and publishing now.
echo      Progress: https://github.com/wyzzgzhdcxy/wcj-go-text/actions

endlocal
pause
