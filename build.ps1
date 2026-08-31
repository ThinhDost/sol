<#
.SYNOPSIS
    Builds optimized Sol binaries for Windows.
#>

$ErrorActionPreference = "Stop"

Write-Host "Building Sol Daemon (sol-daemon.exe)..." -ForegroundColor Cyan
go build -ldflags="-s -w" -o bin/sol-daemon.exe ./cmd/sol-daemon

Write-Host "Building Sol CLI (sol.exe)..." -ForegroundColor Cyan
go build -ldflags="-s -w" -o bin/sol.exe ./cmd/sol-cli

Write-Host "Build complete! Output binaries located in 'bin/' directory:" -ForegroundColor Green
Get-ChildItem -Path bin/ | Select-Object Name, Length, LastWriteTime | Format-Table -AutoSize
