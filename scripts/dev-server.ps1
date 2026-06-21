[CmdletBinding()]
param()

$ErrorActionPreference = "Stop"
$repoRoot = Split-Path -Parent $PSScriptRoot
$backendDir = Join-Path $repoRoot "backend"
$runtimeDir = Join-Path $repoRoot "data\runtime"

$go = Get-Command "go" -ErrorAction Stop
$air = Get-Command "air" -ErrorAction Stop

# Check CGO availability (required by gorm sqlite driver)
$gcc = Get-Command "gcc" -ErrorAction SilentlyContinue
if (-not $gcc) {
    throw "gorm.io/driver/sqlite requires CGO and a C compiler. Install MinGW-w64 and add gcc.exe to PATH."
}

if (Get-NetTCPConnection -LocalPort 8080 -State Listen -ErrorAction SilentlyContinue) {
    throw "Port 8080 is already in use. Stop the conflicting service before starting the backend."
}

$env:CADDYPILOT_BACKEND_ADDR = "127.0.0.1:25610"
$env:CADDYPILOT_FRONTEND_PROXY = "http://127.0.0.1:3000"
$env:CADDYPILOT_RUNTIME_DIR = $runtimeDir
$env:CADDYPILOT_MANAGE_ADDR = ":8080"

Write-Host "Starting backend server (air hot-reload)..." -ForegroundColor Cyan
Write-Host "Web UI: http://localhost:8080"
Write-Host "Backend API: http://127.0.0.1:25610"
Write-Host "Caddy Admin API: http://127.0.0.1:2019"

Push-Location $backendDir
try {
    & $air.Source -c .air.toml
}
finally {
    Pop-Location
}
