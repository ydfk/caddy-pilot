[CmdletBinding()]
param(
    [switch]$SkipInstall
)

$ErrorActionPreference = "Stop"
$repoRoot = Split-Path -Parent $PSScriptRoot
$frontendDir = Join-Path $repoRoot "frontend"

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

    $npx = Get-Command "npx" -ErrorAction Stop
    return @{ File = $npx.Source; Prefix = @("--yes", "pnpm@10.6.2") }
}

$pnpm = Resolve-Pnpm

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

Write-Host "Starting frontend dev server..." -ForegroundColor Cyan
Write-Host "Vite dev URL: http://127.0.0.1:3000"

Push-Location $frontendDir
try {
    & $pnpm.File @($pnpm.Prefix) dev -- --host 127.0.0.1
}
finally {
    Pop-Location
}
