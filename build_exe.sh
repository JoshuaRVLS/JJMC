#!/bin/bash

# Check if tools are missing and install if needed
if ! command -v npm &> /dev/null || ! command -v go &> /dev/null; then
    echo "Developer tools missing. Attempting to install portable toolchain..."
    if [ -f "tools/setup.sh" ]; then
        bash tools/setup.sh
        # Re-prepend local tools to PATH after setup
        if [ -d ".tools/go/bin" ]; then export PATH="$(pwd)/.tools/go/bin:$PATH"; fi
        if [ -d ".tools/node/bin" ]; then export PATH="$(pwd)/.tools/node/bin:$PATH"; fi
    else
        echo "Error: tools/setup.sh not found."
        exit 1
    fi
fi

# This script cross-compiles JJMC for Windows
echo "Building frontend..."
npm run build

echo "Building Windows executable (bin/jjmc.exe)..."
GOOS=windows GOARCH=amd64 go build -o bin/jjmc.exe main.go

echo "Done! bin/jjmc.exe is ready."
