#!/bin/bash
set -e

echo "Starting installation of developer-help-tool-cli..."

# Check OS and Architecture
OS="$(uname -s)"
ARCH="$(uname -m)"

if [ "$OS" != "Darwin" ] || [ "$ARCH" != "arm64" ]; then
    echo "Error: This installation script currently only supports macOS (arm64)."
    echo "Detected OS: $OS, Architecture: $ARCH"
    exit 1
fi

# Configuration
BINARY_URL="https://github.com/eno314/developer-help-tool-cli/releases/latest/download/developer-help-tool-cli-darwin-arm64"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="developer-help-tool-cli"

# Check if we have write permission to INSTALL_DIR
if [ ! -w "$INSTALL_DIR" ]; then
    echo "Error: No write permission to $INSTALL_DIR."
    echo "Please run this script with sudo, e.g.:"
    echo "wget -qO- https://raw.githubusercontent.com/eno314/developer-help-tool-cli/main/install.sh | sudo bash"
    exit 1
fi

echo "Downloading $BINARY_NAME for macOS arm64..."

# Create a temporary file
TMP_FILE=$(mktemp)

# Download the binary
if command -v wget >/dev/null 2>&1; then
    wget -qO "$TMP_FILE" "$BINARY_URL"
elif command -v curl >/dev/null 2>&1; then
    curl -sL -o "$TMP_FILE" "$BINARY_URL"
else
    echo "Error: wget or curl is required to download the binary."
    rm -f "$TMP_FILE"
    exit 1
fi

echo "Installing to $INSTALL_DIR/$BINARY_NAME..."
chmod +x "$TMP_FILE"

# Move the binary to the install directory
if mv "$TMP_FILE" "$INSTALL_DIR/$BINARY_NAME"; then
    echo "Installation successfully completed!"
    echo "You can now run '$BINARY_NAME' from your terminal."
else
    echo "Error: Failed to move the binary to $INSTALL_DIR."
    rm -f "$TMP_FILE"
    exit 1
fi
