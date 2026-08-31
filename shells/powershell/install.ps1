<#
.SYNOPSIS
    Installs the Sol PowerShell hook into your current $PROFILE.
#>

$ErrorActionPreference = "Stop"

$ModulePath = Join-Path $PSScriptRoot "Sol.psm1"
if (-not (Test-Path $ModulePath)) {
    Write-Error "Could not find Sol.psm1 at $ModulePath"
    exit 1
}

$ProfileDir = Split-Path $PROFILE -Parent
if (-not (Test-Path $ProfileDir)) {
    New-Item -ItemType Directory -Path $ProfileDir -Force | Out-Null
}

if (-not (Test-Path $PROFILE)) {
    New-Item -ItemType File -Path $PROFILE -Force | Out-Null
}

$ImportLine = "Import-Module `"$ModulePath`""
$ProfileContent = Get-Content $PROFILE -Raw -ErrorAction SilentlyContinue

if ($ProfileContent -and $ProfileContent.Contains($ImportLine)) {
    Write-Host " Sol PowerShell hook is already configured in $PROFILE" -ForegroundColor Yellow
} else {
    Add-Content -Path $PROFILE -Value "`n# Sol Discord Rich Presence`n$ImportLine"
    Write-Host " Successfully installed Sol PowerShell hook to $PROFILE" -ForegroundColor Green
    Write-Host "Restart your terminal or run '. `$PROFILE' to activate." -ForegroundColor Cyan
}
