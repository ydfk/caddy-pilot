[CmdletBinding()]
param()

. (Join-Path $PSScriptRoot "common.ps1")

$repoRoot = Get-CaddyPilotRoot
$backendDir = Join-Path $repoRoot "backend"
$runtimeDir = Join-Path $repoRoot "data\runtime"

$null = Resolve-RequiredExecutable "go"
$air = Resolve-RequiredExecutable "air"
Assert-CgoCompiler

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
    & $air -c .air.toml
}
finally {
    Pop-Location
}
