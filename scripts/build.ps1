[CmdletBinding()]
param(
    [switch]$SkipInstall
)

. (Join-Path $PSScriptRoot "common.ps1")

$repoRoot = Get-CaddyPilotRoot
$backendDir = Join-Path $repoRoot "backend"
$frontendDir = Join-Path $repoRoot "frontend"
$backendOutput = Join-Path $backendDir "bin\caddypilot.exe"
$go = Resolve-RequiredExecutable "go"
$pnpm = Resolve-PnpmCommand

Assert-CgoCompiler
if (-not $SkipInstall) {
    Invoke-PnpmCommand -Pnpm $pnpm -WorkingDirectory $frontendDir -Arguments @("install", "--frozen-lockfile")
}
Invoke-PnpmCommand -Pnpm $pnpm -WorkingDirectory $frontendDir -Arguments @("build")

New-Item -ItemType Directory -Force -Path (Split-Path -Parent $backendOutput) | Out-Null
Push-Location $backendDir
try {
    & $go build -o $backendOutput ./cmd
    if ($LASTEXITCODE -ne 0) {
        throw "后端构建失败。"
    }
}
finally {
    Pop-Location
}

Write-Host "构建完成：frontend/dist 与 backend/bin/caddypilot.exe" -ForegroundColor Green
