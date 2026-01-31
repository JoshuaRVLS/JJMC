@echo off
setlocal

:: Set Root Dir
set "ROOT_DIR=%~dp0"

:: 1. Attempt to add existing local tools to PATH (Run 1)
:: We use set "VAR=VAL" to safely handle spaces and special characters
if exist "%ROOT_DIR%.tools\go\bin" set "PATH=%ROOT_DIR%.tools\go\bin;%PATH%"
if exist "%ROOT_DIR%.tools\node" set "PATH=%ROOT_DIR%.tools\node;%PATH%"

:: 2. Check if tools are missing
set TOOLS_MISSING=0
where npm >nul 2>nul
if %errorlevel% neq 0 set TOOLS_MISSING=1
where go >nul 2>nul
if %errorlevel% neq 0 set TOOLS_MISSING=1

:: 3. Install if missing
if %TOOLS_MISSING% equ 1 (
    echo Developer tools missing. Attempting to install portable toolchain...
    if exist "tools\setup.bat" (
        call tools\setup.bat
    ) else (
        echo Error: tools\setup.bat not found.
        pause
        exit /b 1
    )
)

:: 4. Re-apply path updates (Run 2) - This must be outside the if block above
:: This ensures that if we just installed them, they are added to PATH now.
:: Doing this outside the block prevents syntax errors if PATH contains parenthesis (e.g. "Program Files (x86)")
if exist "%ROOT_DIR%.tools\go\bin" set "PATH=%ROOT_DIR%.tools\go\bin;%PATH%"
if exist "%ROOT_DIR%.tools\node" set "PATH=%ROOT_DIR%.tools\node;%PATH%"

:: 5. Final Check
where npm >nul 2>nul
if %errorlevel% neq 0 (
    echo Error: Failed to find or install required tools (npm).
    exit /b 1
)
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo Error: Failed to find or install required tools (go).
    pause
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
    shift
    bin\jjmc.exe %*
) else (
    go run main.go %*
)

if %errorlevel% neq 0 pause

