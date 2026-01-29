#!/bin/bash
set -e

# JJMC Portable Toolchain Setup
# This script downloads Go and Node.js to .tools/ for local use.

TOOLS_DIR="$(pwd)/.tools"
mkdir -p "$TOOLS_DIR"

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

# 1. Download Go
GO_VERSION="1.23.5"
GO_URL=""

if [ "$OS" == "linux" ]; then
    GO_URL="https://go.dev/dl/go${GO_VERSION}.linux-${ARCH}.tar.gz"
elif [ "$OS" == "darwin" ]; then
    GO_URL="https://go.dev/dl/go${GO_VERSION}.darwin-${ARCH}.tar.gz"
fi

if [ -n "$GO_URL" ]; then
    echo "Downloading Go ${GO_VERSION} for ${OS}/${ARCH}..."
    curl -L "$GO_URL" -o "$TOOLS_DIR/go.tar.gz"
    echo "Extracting Go..."
    rm -rf "$TOOLS_DIR/go"
    tar -C "$TOOLS_DIR" -xzf "$TOOLS_DIR/go.tar.gz"
    rm "$TOOLS_DIR/go.tar.gz"
else
    echo "Unsupported OS for Go: $OS"
fi

# 2. Download Node.js (LTS)
NODE_VERSION="v22.13.1"
NODE_URL=""

if [ "$OS" == "linux" ]; then
    NODE_URL="https://nodejs.org/dist/${NODE_VERSION}/node-${NODE_VERSION}-linux-${ARCH}.tar.gz"
elif [ "$OS" == "darwin" ]; then
    NODE_URL="https://nodejs.org/dist/${NODE_VERSION}/node-${NODE_VERSION}-darwin-${ARCH}.tar.gz"
fi

if [ -n "$NODE_URL" ]; then
    echo "Downloading Node.js ${NODE_VERSION} for ${OS}/${ARCH}..."
    curl -L "$NODE_URL" -o "$TOOLS_DIR/node.tar.gz"
    echo "Extracting Node.js..."
    rm -rf "$TOOLS_DIR/node"
    mkdir -p "$TOOLS_DIR/node_tmp"
    tar -C "$TOOLS_DIR/node_tmp" -xzf "$TOOLS_DIR/node.tar.gz" --strip-components=1
    mv "$TOOLS_DIR/node_tmp" "$TOOLS_DIR/node"
    rm "$TOOLS_DIR/node.tar.gz"
else
    echo "Unsupported OS for Node.js: $OS"
fi

echo ""
echo "Setup complete! Tools installed in $TOOLS_DIR"
echo "You can now run ./run.sh"
