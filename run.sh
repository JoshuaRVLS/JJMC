#!/bin/bash

# Prepend local tools to PATH if they exist
if [ -d ".tools/go/bin" ]; then
    export PATH="$(pwd)/.tools/go/bin:$PATH"
fi
if [ -d ".tools/node/bin" ]; then
    export PATH="$(pwd)/.tools/node/bin:$PATH"
fi

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

# Final check
if ! command -v npm &> /dev/null || ! command -v go &> /dev/null; then
    echo "Error: Failed to find or install required tools (Go/Node)."
    exit 1
fi

echo "Checking dependencies..."
if [ ! -d "node_modules" ]; then
    echo "Installing frontend dependencies..."
    npm install
fi

echo "Building frontend..."
npm run build

echo "Starting JJMC..."
if [ "$1" == "--build" ]; then
    go build -o bin/jjmc main.go
    ./bin/jjmc
else
    go run main.go
fi
