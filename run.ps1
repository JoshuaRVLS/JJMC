# Prepend local tools to PATH if they exist
$rootDir = $PSScriptRoot
if (Test-Path "$rootDir\.tools\go\bin") { $env:PATH = "$rootDir\.tools\go\bin;$env:PATH" }
if (Test-Path "$rootDir\.tools\node") { $env:PATH = "$rootDir\.tools\node;$env:PATH" }

# Check if tools are missing and install if needed
$npmMissing = -not (Get-Command npm -ErrorAction SilentlyContinue)
$goMissing = -not (Get-Command go -ErrorAction SilentlyContinue)

if ($npmMissing -or $goMissing) {
    Write-Host "Developer tools missing. Attempting to install portable toolchain..." -ForegroundColor Cyan
    if (Test-Path "$rootDir\tools\setup.bat") {
        # Run setup.bat - it downloads tools to .tools/
        Start-Process -FilePath "cmd.exe" -ArgumentList "/c", "`"$rootDir\tools\setup.bat`"" -Wait -NoNewWindow
        
        # Re-prepend local tools to PATH
        if (Test-Path "$rootDir\.tools\go\bin") { $env:PATH = "$rootDir\.tools\go\bin;$env:PATH" }
        if (Test-Path "$rootDir\.tools\node") { $env:PATH = "$rootDir\.tools\node;$env:PATH" }
    } else {
        Write-Host "Error: tools\setup.bat not found." -ForegroundColor Red
        Exit 1
    }
}

# Final check
if (-not (Get-Command npm -ErrorAction SilentlyContinue) -or -not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Host "Error: Failed to find or install required tools (Go/Node)." -ForegroundColor Red
    Exit 1
}

Write-Host "Checking dependencies..." -ForegroundColor Cyan
if (-not (Test-Path "node_modules")) {
    Write-Host "Installing frontend dependencies..." -ForegroundColor Yellow
    npm install
}

Write-Host "Building frontend..." -ForegroundColor Cyan
npm run build

Write-Host "Starting JJMC..." -ForegroundColor Green
if ($args[0] -eq "--build") {
    go build -o bin/jjmc.exe main.go
    .\bin\jjmc.exe
} else {
    go run main.go
}
