@echo off
setlocal

:: JJMC Portable Toolchain Setup (Windows)
set TOOLS_DIR=%~dp0..\.tools
mkdir "%TOOLS_DIR%" 2>nul

:: 1. Download Go
set GO_VERSION=1.23.5
echo Downloading Go %GO_VERSION%...
powershell -Command "Invoke-WebRequest -Uri 'https://go.dev/dl/go%GO_VERSION%.windows-amd64.zip' -OutFile '%TOOLS_DIR%\go.zip'"
echo Extracting Go...
powershell -Command "Expand-Archive -Path '%TOOLS_DIR%\go.zip' -DestinationPath '%TOOLS_DIR%' -Force"
del "%TOOLS_DIR%\go.zip"

:: 2. Download Node.js
set NODE_VERSION=v22.13.1
echo Downloading Node.js %NODE_VERSION%...
powershell -Command "Invoke-WebRequest -Uri 'https://nodejs.org/dist/%NODE_VERSION%/node-%NODE_VERSION%-win-x64.zip' -OutFile '%TOOLS_DIR%\node.zip'"
echo Extracting Node.js...
powershell -Command "Expand-Archive -Path '%TOOLS_DIR%\node.zip' -DestinationPath '%TOOLS_DIR%' -Force"
:: Node extracts to a folder like node-v22.13.1-win-x64, rename it for consistency
for /d %%i in ("%TOOLS_DIR%\node-*") do move "%%i" "%TOOLS_DIR%\node"
del "%TOOLS_DIR%\node.zip"

echo.
echo Setup complete! Tools installed in %TOOLS_DIR%
echo You can now run run.bat
pause
