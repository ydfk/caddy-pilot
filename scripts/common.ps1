$ErrorActionPreference = "Stop"

function Get-CaddyPilotRoot {
    return Split-Path -Parent $PSScriptRoot
}

function Resolve-RequiredExecutable([string]$Name) {
    $command = Get-Command $Name -ErrorAction SilentlyContinue
    if (-not $command) {
        throw "未找到 $Name，请安装后加入 PATH。"
    }
    return $command.Source
}

function Resolve-PnpmCommand {
    $command = Get-Command "pnpm" -ErrorAction SilentlyContinue
    if ($command) {
        $previousPreference = $ErrorActionPreference
        $ErrorActionPreference = "Continue"
        & $command.Source --version *> $null
        $exitCode = $LASTEXITCODE
        $ErrorActionPreference = $previousPreference
        if ($exitCode -eq 0) {
            return @{ File = $command.Source; Prefix = @() }
        }
    }

    $npx = Resolve-RequiredExecutable "npx"
    return @{ File = $npx; Prefix = @("--yes", "pnpm@10.6.2") }
}

function Invoke-PnpmCommand {
    param(
        [hashtable]$Pnpm,
        [string]$WorkingDirectory,
        [string[]]$Arguments
    )

    Push-Location $WorkingDirectory
    try {
        & $Pnpm.File @($Pnpm.Prefix) @Arguments
        if ($LASTEXITCODE -ne 0) {
            throw "pnpm 命令执行失败。"
        }
    }
    finally {
        Pop-Location
    }
}

function Assert-CgoCompiler {
    if (-not (Get-Command "gcc" -ErrorAction SilentlyContinue)) {
        throw "SQLite 驱动需要 CGO 和 C 编译器，请安装 MinGW-w64 并将 gcc.exe 加入 PATH。"
    }
}
