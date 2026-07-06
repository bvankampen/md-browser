#!/bin/sh
set -e

echo "=== Markdown Browser Installer ==="

# 1. Resolve Target Installation Directory
INSTALL_DIR=""

# Check env GO_PATH, GOPATH, or go env GOPATH
GP=""
if [ -n "$GO_PATH" ]; then
    GP="$GO_PATH"
elif [ -n "$GOPATH" ]; then
    GP="$GOPATH"
else
    GP="$(go env GOPATH 2>/dev/null || echo "")"
fi

if [ -n "$GP" ] && [ -d "$GP/bin" ]; then
    INSTALL_DIR="$GP/bin"
elif [ -n "$GP" ] && [ -d "$GP" ]; then
    # If GOPATH exists but bin subdirectory doesn't, try to create it
    mkdir -p "$GP/bin" 2>/dev/null && INSTALL_DIR="$GP/bin"
fi

# Fallback to /usr/local/bin if GOPATH/bin is unavailable or doesn't exist
if [ -z "$INSTALL_DIR" ] || [ ! -d "$INSTALL_DIR" ]; then
    INSTALL_DIR="/usr/local/bin"
fi

echo "Target installation directory: $INSTALL_DIR"

# 2. Build the application binary
echo "Building binary..."
go build -o md-browser ./cmd/md-browser

# 3. Copy binary with proper permissions
echo "Installing binary to $INSTALL_DIR..."
if [ -w "$INSTALL_DIR" ]; then
    cp md-browser "$INSTALL_DIR/md-browser"
else
    echo "Requires administrative privileges to write to $INSTALL_DIR. Prompting for sudo..."
    sudo cp md-browser "$INSTALL_DIR/md-browser"
fi

# Clean up local build artifact
rm md-browser

echo "======================================================"
echo "Installation completed successfully!"
echo "You can now run the browser with: md-browser"
echo "======================================================"
