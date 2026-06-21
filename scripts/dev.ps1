[CmdletBinding()]
param(
    [switch]$SkipInstall,
    [switch]$Check
)

$ErrorActionPreference = "Stop"
$repoRoot = Split-Path -Parent $PSScriptRoot
$backendDir = Join-Path $repoRoot "backend"
$frontendDir = Join-Path $repoRoot "frontend"
$runtimeDir = Join-Path $repoRoot "data\runtime"

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
    Write-Host "Local prerequisites are ready. Caddy will be managed by the backend." -ForegroundColor Green
    Write-Host "Go: $go"
    Write-Host "pnpm launcher: $($pnpm.File) $($pnpm.Prefix -join ' ')"
    exit 0
}

if (Get-NetTCPConnection -LocalPort 8080 -State Listen -ErrorAction SilentlyContinue) {
    throw "Port 8080 is already in use. Stop the Docker deployment before local development."
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

$backendJob = Start-Job -Name "CaddyPilot-Backend" -ArgumentList $backendDir, $go, $runtimeDir -ScriptBlock {
    param($WorkingDirectory, $GoExecutable, $RuntimeDirectory)
    Set-Location $WorkingDirectory
    $env:CADDYPILOT_BACKEND_ADDR = "127.0.0.1:25610"
    $env:CADDYPILOT_FRONTEND_PROXY = "http://127.0.0.1:3000"
    $env:CADDYPILOT_RUNTIME_DIR = $RuntimeDirectory
    $env:CADDYPILOT_MANAGE_ADDR = ":8080"
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
Write-Host "Starting backend-managed CaddyPilot..." -ForegroundColor Cyan
Write-Host "Web UI: http://localhost:8080"
Write-Host "Vite internal URL: http://127.0.0.1:3000"
Write-Host "Backend API: http://127.0.0.1:25610"
Write-Host "Caddy Admin API: http://127.0.0.1:2019"
Write-Host "Press Ctrl+C to stop Caddy, backend, and frontend together."

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
    try {
        Invoke-WebRequest -Method Post -Uri "http://127.0.0.1:2019/stop" -TimeoutSec 3 -UseBasicParsing | Out-Null
    }
    catch {
    }
    $jobs | Stop-Job -ErrorAction SilentlyContinue
    $jobs | Remove-Job -Force -ErrorAction SilentlyContinue
}
