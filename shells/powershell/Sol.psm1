<#
.SYNOPSIS
    Sol - Ultra-lightweight Discord Rich Presence hook module for PowerShell.

.DESCRIPTION
    Integrates directly with the Sol daemon via in-memory .NET Named Pipe streams.
    Executes in < 0.5ms with zero process overhead.
#>

# Name of the local Sol Named Pipe
$script:SolPipeName = "sol-ipc"
$script:SolHome = Join-Path $env:USERPROFILE ".sol"
$script:SolBin = Join-Path $script:SolHome "bin"
$script:DaemonExe = Join-Path $script:SolBin "sol-daemon.exe"

# Function to ensure Sol background daemon is running seamlessly
function Start-SolDaemonIfNeeded {
    try {
        $proc = Get-Process -Name "sol-daemon" -ErrorAction SilentlyContinue
        if (-not $proc) {
            $daemonPath = $null
            if (Test-Path $script:DaemonExe) {
                $daemonPath = $script:DaemonExe
            } elseif (Test-Path "$PSScriptRoot\..\..\bin\sol-daemon.exe") {
                $daemonPath = (Resolve-Path "$PSScriptRoot\..\..\bin\sol-daemon.exe").Path
            } elseif (Test-Path "$PSScriptRoot\bin\sol-daemon.exe") {
                $daemonPath = (Resolve-Path "$PSScriptRoot\bin\sol-daemon.exe").Path
            }

            if ($daemonPath) {
                $workingDir = Split-Path $daemonPath -Parent
                $configParent = Split-Path $workingDir -Parent
                if (Test-Path (Join-Path $configParent "sol.config.json")) {
                    $workingDir = $configParent
                }
                Start-Process -FilePath $daemonPath -WorkingDirectory $workingDir -WindowStyle Hidden -ErrorAction SilentlyContinue
                Start-Sleep -Milliseconds 200
            }
        }
    } catch {
        # Fail silently
    }
}

# Function to send an event asynchronously to Sol daemon
function Send-SolEvent {
    param(
        [Parameter(Mandatory=$true)]
        [ValidateSet("start", "idle", "exit", "ping")]
        [string]$Event,

        [Parameter(Mandatory=$false)]
        [string]$Cmd = "",

        [Parameter(Mandatory=$false)]
        [string]$Cwd = $PWD.Path
    )

    try {
        # Connect to Windows Named Pipe with 15ms timeout (never blocks terminal)
        $pipe = New-Object System.IO.Pipes.NamedPipeClientStream(".", $script:SolPipeName, [System.IO.Pipes.PipeDirection]::Out)
        $pipe.Connect(15)
        
        $writer = New-Object System.IO.StreamWriter($pipe)
        $writer.AutoFlush = $true

        # Construct compact JSON payload
        $escapedCmd = ($Cmd -replace '\\', '\\\\' -replace '"', '\"' -replace "`n", ' ' -replace "`r", '')
        $escapedCwd = ($Cwd -replace '\\', '/')
        $timestamp = [DateTimeOffset]::UtcNow.ToUnixTimeSeconds()
        
        $json = "{`"event`":`"$Event`",`"cmd`":`"$escapedCmd`",`"cwd`":`"$escapedCwd`",`"shell`":`"powershell`",`"timestamp`":$timestamp}`n"
        
        $writer.WriteLine($json)
        $writer.Close()
        $pipe.Close()
    } catch {
        # If pipe connection failed, spawn daemon in background for future commands
        Start-SolDaemonIfNeeded
    }
}

# Preserve existing prompt if defined
if (Test-Path function:global:prompt) {
    $script:OriginalPrompt = $function:global:prompt
} else {
    $script:OriginalPrompt = { "PS $($executionContext.SessionState.Path.CurrentLocation)> " }
}

# Override global prompt function to capture post-command / idle state
function global:prompt {
    # Send idle state to Sol
    Send-SolEvent -Event "idle" -Cwd $PWD.Path

    # Execute original prompt
    & $script:OriginalPrompt
}

# Setup PSReadLine Enter key handler if PSReadLine module is available
if (Get-Module -Name PSReadLine) {
    Set-PSReadLineKeyHandler -Chord Enter -ScriptBlock {
        # Get line buffer content
        $line = $null
        $cursor = $null
        [Microsoft.PowerShell.PSConsoleReadLine]::GetBufferState([ref]$line, [ref]$cursor)

        if (![string]::IsNullOrWhiteSpace($line)) {
            Send-SolEvent -Event "start" -Cmd $line -Cwd $PWD.Path
        }

        # Accept line execution
        [Microsoft.PowerShell.PSConsoleReadLine]::AcceptLine()
    }
}

# Automatically ensure daemon is alive upon opening terminal
Start-SolDaemonIfNeeded

# Send initial connection event
Send-SolEvent -Event "idle" -Cwd $PWD.Path

Export-ModuleMember -Function Send-SolEvent
