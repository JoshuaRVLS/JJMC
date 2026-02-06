param (
    [switch]$SkipBuild,
    [switch]$Build,
    [Parameter(ValueFromRemainingArguments=$true)]
    [string[]]$RemainingArgs
)

# Colors
function Write-Color($text, $color) {
    Write-Host $text -ForegroundColor $color
}

# Prepend local tools to PATH if they exist
$rootDir = $PSScriptRoot
if (Test-Path "$rootDir\.tools\go\bin") { $env:PATH = "$rootDir\.tools\go\bin;$env:PATH" }
if (Test-Path "$rootDir\.tools\node") { $env:PATH = "$rootDir\.tools\node;$env:PATH" }

# Check if tools are missing and install if needed
$npmMissing = -not (Get-Command npm -ErrorAction SilentlyContinue)
$goMissing = -not (Get-Command go -ErrorAction SilentlyContinue)

if ($npmMissing -or $goMissing) {
    Write-Color "Developer tools missing. Attempting to install portable toolchain..." "Cyan"
    if (Test-Path "$rootDir\tools\setup.bat") {
        # Run setup.bat - it downloads tools to .tools/
        Start-Process -FilePath "cmd.exe" -ArgumentList "/c", "`"$rootDir\tools\setup.bat`"" -Wait -NoNewWindow
        
        # Re-prepend local tools to PATH
        if (Test-Path "$rootDir\.tools\go\bin") { $env:PATH = "$rootDir\.tools\go\bin;$env:PATH" }
        if (Test-Path "$rootDir\.tools\node") { $env:PATH = "$rootDir\.tools\node;$env:PATH" }
    } else {
        Write-Color "Error: tools\setup.bat not found." "Red"
        Exit 1
    }
}

# Final check
if (-not (Get-Command npm -ErrorAction SilentlyContinue) -or -not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Color "Error: Failed to find or install required tools (Go/Node)." "Red"
    Exit 1
}

if (-not $SkipBuild) {
    Write-Color "Checking dependencies..." "Cyan"
    if (-not (Test-Path "frontend/node_modules")) {
        Write-Color "Installing frontend dependencies..." "Yellow"
        Push-Location frontend
        npm install
        Pop-Location
    }

    Write-Color "Building frontend..." "Cyan"
    Push-Location frontend
    npm run build
    Pop-Location
} else {
    Write-Color "Skipping frontend build..." "Yellow"
}

Write-Color "Starting JJMC..." "Green"
if ($Build) {
    go build -o bin/jjmc.exe cmd/jjmc/main.go
    .\bin\jjmc.exe $RemainingArgs
} else {
    go run cmd/jjmc/main.go $RemainingArgs
}
