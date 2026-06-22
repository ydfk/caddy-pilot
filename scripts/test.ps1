[CmdletBinding()]
param()

. (Join-Path $PSScriptRoot "common.ps1")

$repoRoot = Get-CaddyPilotRoot
$backendDir = Join-Path $repoRoot "backend"
$frontendDir = Join-Path $repoRoot "frontend"
$go = Resolve-RequiredExecutable "go"
$pnpm = Resolve-PnpmCommand

Assert-CgoCompiler
Push-Location $backendDir
try {
    & $go test ./...
    if ($LASTEXITCODE -ne 0) {
        throw "后端测试失败。"
    }
}
finally {
    Pop-Location
}

Invoke-PnpmCommand -Pnpm $pnpm -WorkingDirectory $frontendDir -Arguments @("test:run")
