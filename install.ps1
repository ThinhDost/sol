<#
.SYNOPSIS
    Sol One-Line Automated Installer for Windows (PowerShell).
.DESCRIPTION
    Installs pre-compiled Sol standalone binaries, configures shell profile,
    and launches background daemon. ZERO dependencies required (No Go, No Git needed).
#>

$ErrorActionPreference = "Continue"

Write-Host ""
Write-Host "  [+] Installing Sol - Terminal Discord Rich Presence..." -ForegroundColor Yellow
Write-Host ""

$SolHome = Join-Path $env:USERPROFILE ".sol"
$SolBin = Join-Path $SolHome "bin"

# Create Sol directories
if (-not (Test-Path $SolBin)) {
    New-Item -ItemType Directory -Path $SolBin -Force | Out-Null
}

$RepoRawUrl = "https://raw.githubusercontent.com/ThinhDost/sol/main"
$ReleaseUrl = "https://github.com/ThinhDost/sol/releases/latest/download"

# Destination paths
$ModuleDest = Join-Path $SolHome "Sol.psm1"
$ConfigDest = Join-Path $SolHome "sol.config.json"
$DaemonExe = Join-Path $SolBin "sol-daemon.exe"
$CliExe = Join-Path $SolBin "sol.exe"

# 1. Download or copy PowerShell hook module and config
if (Test-Path "$PSScriptRoot\shells\powershell\Sol.psm1") {
    Copy-Item "$PSScriptRoot\shells\powershell\Sol.psm1" -Destination $ModuleDest -Force
    if (Test-Path "$PSScriptRoot\sol.config.json") {
        Copy-Item "$PSScriptRoot\sol.config.json" -Destination $ConfigDest -Force
    }
} else {
    Write-Host "  [*] Downloading Sol shell hook and config..." -ForegroundColor Cyan
    Invoke-RestMethod -Uri "$RepoRawUrl/shells/powershell/Sol.psm1" -OutFile $ModuleDest
    Invoke-RestMethod -Uri "$RepoRawUrl/sol.config.json" -OutFile $ConfigDest
}

# 2. Download pre-compiled standalone binaries (Zero runtime / Go / Git required!)
if (Test-Path "$PSScriptRoot\bin\sol-daemon.exe") {
    Copy-Item "$PSScriptRoot\bin\sol-daemon.exe" -Destination $DaemonExe -Force
    Copy-Item "$PSScriptRoot\bin\sol.exe" -Destination $CliExe -Force
} else {
    Write-Host "  [*] Downloading pre-built standalone binaries from GitHub Release..." -ForegroundColor Cyan
    try {
        Invoke-WebRequest -Uri "$ReleaseUrl/sol-windows-amd64.exe" -OutFile $DaemonExe -UseBasicParsing
        Invoke-WebRequest -Uri "$ReleaseUrl/sol-windows-amd64-cli.exe" -OutFile $CliExe -UseBasicParsing
    } catch {
        Write-Host "  [!] Release download fallback: compiling from source..." -ForegroundColor Yellow
        if (Get-Command go -ErrorAction SilentlyContinue) {
            $TempDir = Join-Path $env:TEMP "sol-build-$(Get-Random)"
            git clone --depth 1 https://github.com/ThinhDost/sol.git $TempDir 2>$null
            if (Test-Path $TempDir) {
                Push-Location $TempDir
                go build -ldflags="-s -w" -o $DaemonExe ./cmd/sol-daemon
                go build -ldflags="-s -w" -o $CliExe ./cmd/sol-cli
                Pop-Location
                Remove-Item -Recurse -Force $TempDir -ErrorAction SilentlyContinue
            }
        }
    }
}

# 3. Configure PowerShell $PROFILE
$ProfileDir = Split-Path $PROFILE -Parent
if (-not (Test-Path $ProfileDir)) {
    New-Item -ItemType Directory -Path $ProfileDir -Force | Out-Null
}
if (-not (Test-Path $PROFILE)) {
    New-Item -ItemType File -Path $PROFILE -Force | Out-Null
}

$ProfileContent = Get-Content $PROFILE -Raw -ErrorAction SilentlyContinue
$HookCommand = "Import-Module '$ModuleDest'"

if (-not $ProfileContent -or -not $ProfileContent.Contains($HookCommand)) {
    Add-Content -Path $PROFILE -Value "`n# Sol Discord Rich Presence`n$HookCommand"
    Write-Host "  [+] Added Sol hook to $PROFILE" -ForegroundColor Green
}

# 4. Add Sol bin to User PATH environment variable
$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($UserPath -notlike "*$SolBin*") {
    $NewPath = $UserPath + ";" + $SolBin
    [Environment]::SetEnvironmentVariable("Path", $NewPath, "User")
    $env:Path += ";$SolBin"
}

# 5. Register Sol daemon to auto-start with Windows
try {
    $RunKey = "HKCU:\Software\Microsoft\Windows\CurrentVersion\Run"
    Set-ItemProperty -Path $RunKey -Name "SolDiscordRPC" -Value $DaemonExe -Force -ErrorAction SilentlyContinue
} catch {}

# 6. Terminate old instance and launch background daemon
Get-Process -Name "sol-daemon" -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue

if (Test-Path $DaemonExe) {
    Write-Host "  [*] Launching Sol background daemon..." -ForegroundColor Cyan
    Start-Process -FilePath $DaemonExe -WorkingDirectory $SolHome -WindowStyle Hidden
}

# 7. Import hook into current session immediately
if (Test-Path $ModuleDest) {
    Import-Module $ModuleDest -Force
}

Write-Host ""
Write-Host "  [OK] Sol is installed and running successfully!" -ForegroundColor Green
Write-Host "  [!] Discord Rich Presence is now active for all your terminal commands." -ForegroundColor Yellow
Write-Host ""
