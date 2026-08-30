#Requires -Version 5.1
$ErrorActionPreference = 'Stop'

# Change to repo root (script directory's parent)
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location (Join-Path $ScriptDir '..')

# Ensure git is available
if (-not (Get-Command git -ErrorAction SilentlyContinue)) {
    Write-Host '[ERROR] git not found. Install Git first.'
    exit 1
}

# Find the highest v* tag (version-sorted, descending)
# 注意：不要对 git 原生命令做 2>$null 重定向——$ErrorActionPreference='Stop' 时，
# PS 5.1 会把 stderr 重定向内容包装成 ErrorRecord 并直接抛出 NativeCommandError
$latest = git tag --list 'v*' --sort=-version:refname | Select-Object -First 1
if (-not $latest) { $latest = 'v0.0.0' }

# Parse major.minor.patch from the latest tag
$ver = $latest -replace '^v', ''
$parts = $ver.Split('.')
$maj = 0; $min = 0; $pat = 0
if ($parts.Count -ge 1) { [int]::TryParse($parts[0], [ref]$maj) | Out-Null }
if ($parts.Count -ge 2) { [int]::TryParse($parts[1], [ref]$min) | Out-Null }
if ($parts.Count -ge 3) { [int]::TryParse($parts[2], [ref]$pat) | Out-Null }

# Bump the patch version
$pat++
$newtag = "v$maj.$min.$pat"

# Guard against an already-existing tag
# 不用 rev-parse --verify：tag 不存在时它本就会向 stderr 报 fatal 并返回非零，
# 在 PS 5.1 + Stop 偏好下会触发 NativeCommandError。tag --list 只走 stdout，
# 无匹配时输出为空，天然避开该坑
if (git tag --list $newtag) {
    Write-Host "[ERROR] tag $newtag already exists."
    exit 1
}

Write-Host '============================================'
Write-Host "  Latest tag : $latest"
Write-Host "  New release: $newtag"
Write-Host '============================================'
Write-Host ''

git tag $newtag
if ($LASTEXITCODE -ne 0) {
    Write-Host "[ERROR] failed to create tag $newtag"
    exit 1
}

git push origin $newtag
if ($LASTEXITCODE -ne 0) {
    Write-Host "[ERROR] failed to push tag $newtag"
    exit 1
}

Write-Host ''
Write-Host "[OK] Tag $newtag pushed. GitHub Actions is building and publishing now."
Write-Host '      Progress: https://github.com/wyzzgzhdcxy/wcj-go-text/actions'

Read-Host 'Press Enter to exit'
