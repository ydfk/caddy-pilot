[CmdletBinding()]
param(
    [switch]$SkipInstall
)

. (Join-Path $PSScriptRoot "common.ps1")

$repoRoot = Get-CaddyPilotRoot
$frontendDir = Join-Path $repoRoot "frontend"
$pnpm = Resolve-PnpmCommand

if (-not $SkipInstall) {
    Invoke-PnpmCommand -Pnpm $pnpm -WorkingDirectory $frontendDir -Arguments @("install")
}

Write-Host "Starting frontend dev server..." -ForegroundColor Cyan
Write-Host "Vite dev URL: http://127.0.0.1:3000"

Push-Location $frontendDir
try {
    & $pnpm.File @($pnpm.Prefix) dev -- --host 127.0.0.1
}
finally {
    Pop-Location
}
