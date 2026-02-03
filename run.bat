@echo off
setlocal

:: Set Root Dir
set "ROOT_DIR=%~dp0"

:: 1. Attempt to add existing local tools to PATH (Run 1)
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

:: 4. Re-apply path updates (Run 2)
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

:: Parse Arguments
set SKIP_BUILD=0
set BUILD_BINARY=0
set "ARGS="

:parse_args
if "%~1"=="" goto end_parse_args
if "%~1"=="--skip-build" (
    set SKIP_BUILD=1
) else if "%~1"=="--build" (
    set BUILD_BINARY=1
) else (
    set "ARGS=%ARGS% %1"
)
shift
goto parse_args
:end_parse_args

if %SKIP_BUILD% equ 0 (
    echo Checking dependencies...
    if not exist "node_modules\" (
        echo Installing frontend dependencies...
        call npm install
    )

    echo Building frontend...
    call npm run build
) else (
    echo Skipping frontend build...
)

echo Starting JJMC...
if %BUILD_BINARY% equ 1 (
    go build -o bin\jjmc.exe main.go
    bin\jjmc.exe %ARGS%
) else (
    go run main.go %ARGS%
)

if %errorlevel% neq 0 pause
