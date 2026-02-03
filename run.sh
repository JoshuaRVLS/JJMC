#!/bin/bash

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Prepend local tools to PATH if they exist
if [ -d ".tools/go/bin" ]; then
    export PATH="$(pwd)/.tools/go/bin:$PATH"
fi
if [ -d ".tools/node/bin" ]; then
    export PATH="$(pwd)/.tools/node/bin:$PATH"
fi

# Check if tools are missing and install if needed
if ! command -v npm &> /dev/null || ! command -v go &> /dev/null; then
    echo -e "${YELLOW}Developer tools missing. Attempting to install portable toolchain...${NC}"
    if [ -f "tools/setup.sh" ]; then
        bash tools/setup.sh
        # Re-prepend local tools to PATH after setup
        if [ -d ".tools/go/bin" ]; then export PATH="$(pwd)/.tools/go/bin:$PATH"; fi
        if [ -d ".tools/node/bin" ]; then export PATH="$(pwd)/.tools/node/bin:$PATH"; fi
    else
        echo -e "${RED}Error: tools/setup.sh not found.${NC}"
        exit 1
    fi
fi

# Final check
if ! command -v npm &> /dev/null || ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Failed to find or install required tools (Go/Node).${NC}"
    exit 1
fi

SKIP_BUILD=false
BUILD_BINARY=false
ARGS=()

# Parse arguments
for arg in "$@"; do
    case $arg in
        --skip-build)
            SKIP_BUILD=true
            shift
            ;;
        --build)
            BUILD_BINARY=true
            shift
            ;;
        *)
            ARGS+=("$arg")
            ;;
    esac
done

if [ "$SKIP_BUILD" = false ]; then
    echo -e "${BLUE}Checking dependencies...${NC}"
    if [ ! -d "node_modules" ]; then
        echo -e "${YELLOW}Installing frontend dependencies...${NC}"
        npm install
    fi

    echo -e "${BLUE}Building frontend...${NC}"
    npm run build
else
    echo -e "${YELLOW}Skipping frontend build...${NC}"
fi

echo -e "${GREEN}Starting JJMC...${NC}"
if [ "$BUILD_BINARY" = true ]; then
    go build -o bin/jjmc main.go
    ./bin/jjmc "${ARGS[@]}"
else
    go run main.go "${ARGS[@]}"
fi
