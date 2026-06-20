[CmdletBinding()]
param(
    [switch]$SkipInstall,
    [switch]$Check
)

$ErrorActionPreference = "Stop"
$repoRoot = Split-Path -Parent $PSScriptRoot
$backendDir = Join-Path $repoRoot "backend"
$frontendDir = Join-Path $repoRoot "frontend"

function Resolve-Executable([string]$Name) {
    $command = Get-Command $Name -ErrorAction SilentlyContinue
    if (-not $command) {
        throw "$Name was not found. Install it and add it to PATH."
    }
    return $command.Source
}

function Resolve-Pnpm {
    $command = Get-Command "pnpm" -ErrorAction SilentlyContinue
    if ($command) {
        $previousPreference = $ErrorActionPreference
        $ErrorActionPreference = "Continue"
        & $command.Source --version *> $null
        $pnpmExitCode = $LASTEXITCODE
        $ErrorActionPreference = $previousPreference
        if ($pnpmExitCode -eq 0) {
            return @{ File = $command.Source; Prefix = @() }
        }
    }

    $npx = Resolve-Executable "npx"
    return @{ File = $npx; Prefix = @("--yes", "pnpm@10.6.2") }
}

$go = Resolve-Executable "go"
$pnpm = Resolve-Pnpm

if ($Check) {
    Write-Host "Local development prerequisites are ready." -ForegroundColor Green
    Write-Host "Go: $go"
    Write-Host "pnpm launcher: $($pnpm.File) $($pnpm.Prefix -join ' ')"
    exit 0
}

if (-not $SkipInstall) {
    Push-Location $frontendDir
    try {
        & $pnpm.File @($pnpm.Prefix) install
        if ($LASTEXITCODE -ne 0) {
            throw "Frontend dependency installation failed."
        }
    }
    finally {
        Pop-Location
    }
}

$backendJob = Start-Job -Name "CaddyPilot-Backend" -ArgumentList $backendDir, $go -ScriptBlock {
    param($WorkingDirectory, $GoExecutable)
    Set-Location $WorkingDirectory
    $env:CADDYPILOT_BACKEND_ADDR = "127.0.0.1:25610"
    & $GoExecutable run ./cmd
}

$pnpmPrefix = $pnpm.Prefix -join "|"
$frontendJob = Start-Job -Name "CaddyPilot-Frontend" -ArgumentList $frontendDir, $pnpm.File, $pnpmPrefix -ScriptBlock {
    param($WorkingDirectory, $PnpmExecutable, $SerializedPrefix)
    Set-Location $WorkingDirectory
    $PnpmPrefix = if ($SerializedPrefix) { $SerializedPrefix -split "\|" } else { @() }
    & $PnpmExecutable @($PnpmPrefix) dev -- --host 127.0.0.1
}

$jobs = @($backendJob, $frontendJob)
Write-Host "Starting CaddyPilot local development..." -ForegroundColor Cyan
Write-Host "Web UI: http://localhost:3000"
Write-Host "Backend API: http://127.0.0.1:25610"
Write-Host "Press Ctrl+C to stop both processes."

try {
    while (($jobs | Where-Object State -eq "Running").Count -eq $jobs.Count) {
        $jobs | Receive-Job
        Start-Sleep -Milliseconds 300
    }

    $jobs | Receive-Job
    $failedJob = $jobs | Where-Object State -eq "Failed" | Select-Object -First 1
    if ($failedJob) {
        throw "$($failedJob.Name) failed to start."
    }
}
finally {
    $jobs | Stop-Job -ErrorAction SilentlyContinue
    $jobs | Remove-Job -Force -ErrorAction SilentlyContinue
}
