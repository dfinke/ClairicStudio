#Requires -Version 7.2
[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'
Push-Location $PSScriptRoot
try {
    foreach ($tool in 'go', 'npm') {
        if (-not (Get-Command $tool -ErrorAction SilentlyContinue)) {
            throw "Install $tool and add it to PATH before building."
        }
    }

    $wailsCommand = Get-Command wails -ErrorAction SilentlyContinue
    if ($wailsCommand) {
        $wailsPath = $wailsCommand.Source
    } else {
        $goBin = go env GOBIN
        if ($LASTEXITCODE) {
            throw 'Could not locate the Go binary directory.'
        }
        if (-not $goBin) {
            $goPaths = (go env GOPATH) -split [IO.Path]::PathSeparator
            if ($LASTEXITCODE) {
                throw 'Could not locate GOPATH.'
            }
            $goBin = Join-Path $goPaths[0] 'bin'
        }
        $wailsPath = Join-Path $goBin $(if ($IsWindows) { 'wails.exe' } else { 'wails' })
        if (-not (Test-Path -LiteralPath $wailsPath)) {
            throw 'Install Wails: go install github.com/wailsapp/wails/v2/cmd/wails@v2.10.2'
        }
    }

    Push-Location frontend
    try {
        npm ci
        if ($LASTEXITCODE) {
            throw 'Frontend dependency installation failed.'
        }
        npm test
        if ($LASTEXITCODE) {
            throw 'Frontend tests failed.'
        }
        npm run build
        if ($LASTEXITCODE) {
            throw 'Frontend build failed.'
        }
    } finally {
        Pop-Location
    }

    go test ./...
    if ($LASTEXITCODE) {
        throw 'Go tests failed.'
    }
    go vet ./...
    if ($LASTEXITCODE) {
        throw 'Go analysis failed.'
    }
    & $wailsPath build
    if ($LASTEXITCODE) {
        throw 'Desktop build failed.'
    }
    Write-Host "Built $PSScriptRoot/build/bin/ClairicStudio.exe"
} finally {
    Pop-Location
}
