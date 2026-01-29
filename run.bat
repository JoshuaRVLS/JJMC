@echo off
setlocal

:: Prepend local tools to PATH if they exist
set ROOT_DIR=%~dp0
if exist "%ROOT_DIR%.tools\go\bin" set PATH=%ROOT_DIR%.tools\go\bin;%PATH%
if exist "%ROOT_DIR%.tools\node" set PATH=%ROOT_DIR%.tools\node;%PATH%

:: Check if tools are missing and install if needed
where npm >nul 2>nul
set NPM_MISSING=%errorlevel%
where go >nul 2>nul
set GO_MISSING=%errorlevel%

if %NPM_MISSING% neq 0 (
    set TOOLS_MISSING=1
) else if %GO_MISSING% neq 0 (
    set TOOLS_MISSING=1
) else (
    set TOOLS_MISSING=0
)

if %TOOLS_MISSING% equ 1 (
    echo Developer tools missing. Attempting to install portable toolchain...
    if exist "tools\setup.bat" (
        call tools\setup.bat
        :: Re-check path
        if exist "%ROOT_DIR%.tools\go\bin" set PATH=%ROOT_DIR%.tools\go\bin;%PATH%
        if exist "%ROOT_DIR%.tools\node" set PATH=%ROOT_DIR%.tools\node;%PATH%
    ) else (
        echo Error: tools\setup.bat not found.
        exit /b 1
    )
)

:: Final check
where npm >nul 2>nul
if %errorlevel% neq 0 (
    echo Error: Failed to find or install required tools (Go/Node).
    exit /b 1
)
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo Error: Failed to find or install required tools (Go/Node).
    exit /b 1
)

echo Checking dependencies...
if not exist "node_modules\" (
    echo Installing frontend dependencies...
    call npm install
)

echo Building frontend...
call npm run build

echo Starting JJMC...
if "%1"=="--build" (
    go build -o bin\jjmc.exe main.go
    bin\jjmc.exe
) else (
    go run main.go
)
