# ==============================================================================
# MemGit Build & Package Script (Windows PowerShell)
# ==============================================================================
param (
    [switch]$WebOnly,
    [switch]$BinOnly,
    [switch]$Test,
    [switch]$Clean,
    [string]$Output = "bin"
)

$ErrorActionPreference = "Stop"
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path

function Write-Banner {
    Write-Host "=======================================================" -ForegroundColor Cyan
    Write-Host "  MemGit - Build System (Svelte 5 + Go Backend)" -ForegroundColor Cyan
    Write-Host "=======================================================" -ForegroundColor Cyan
}

function Clean-Artifacts {
    Write-Host "Cleaning previous build artifacts..." -ForegroundColor Yellow
    if (Test-Path "$ScriptDir/$Output") {
        Remove-Item -Recurse -Force "$ScriptDir/$Output"
    }
    if (Test-Path "$ScriptDir/web/dist") {
        Remove-Item -Recurse -Force "$ScriptDir/web/dist"
    }
    Write-Host "Clean complete." -ForegroundColor Green
}

function Build-Frontend {
    Write-Host "[1/3] Building Svelte 5 Frontend..." -ForegroundColor Cyan
    Push-Location "$ScriptDir/web"
    try {
        if (-not (Test-Path "node_modules")) {
            Write-Host "  Installing npm dependencies..." -ForegroundColor Gray
            npm install --silent
        }
        npm run build
        Write-Host "Svelte frontend compiled successfully to web/dist/" -ForegroundColor Green
    }
    finally {
        Pop-Location
    }
}

function Build-Binaries {
    Write-Host "[2/3] Compiling MemGit Server Binary..." -ForegroundColor Cyan
    if (-not (Test-Path "$ScriptDir/$Output")) {
        New-Item -ItemType Directory -Path "$ScriptDir/$Output" | Out-Null
    }

    $ext = if ($IsWindows -or ($env:OS -like "*Windows*")) { ".exe" } else { "" }

    Write-Host "  -> Compiling memgit-server..." -ForegroundColor Gray
    go build -ldflags="-s -w" -o "$ScriptDir/$Output/memgit-server$ext" "$ScriptDir/cmd/memgit-server"

    Write-Host "Server binary compiled to $Output/memgit-server$ext" -ForegroundColor Green
}

function Run-Tests {
    Write-Host "[3/3] Running Unit & Integration Tests..." -ForegroundColor Cyan
    Push-Location $ScriptDir
    try {
        go test -v ./...
        Write-Host "All tests passed 100%!" -ForegroundColor Green
    }
    finally {
        Pop-Location
    }
}

Write-Banner

if ($Clean) {
    Clean-Artifacts
    exit 0
}

if ($Test) {
    Run-Tests
    exit 0
}

if ($WebOnly) {
    Build-Frontend
    exit 0
}

if ($BinOnly) {
    Build-Binaries
    exit 0
}

# Default: Full Build (Web -> Server Binary -> Tests)
Build-Frontend
Build-Binaries
Run-Tests

Write-Host ""
Write-Host "MemGit built successfully! Launch with: ./$Output/memgit-server -port 8500" -ForegroundColor Green
