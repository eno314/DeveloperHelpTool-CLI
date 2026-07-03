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
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="developer-help-tool-cli"

# Determine version (default to latest)
VERSION=${1:-latest}

if [ "$VERSION" = "latest" ]; then
    if command -v curl >/dev/null 2>&1; then
        LATEST_TAG=$(curl -sL https://api.github.com/repos/eno314/DeveloperHelpTool-CLI/releases/latest | grep '"tag_name":' | cut -d '"' -f 4)
    elif command -v wget >/dev/null 2>&1; then
        LATEST_TAG=$(wget -qO- https://api.github.com/repos/eno314/DeveloperHelpTool-CLI/releases/latest | grep '"tag_name":' | cut -d '"' -f 4)
    else
        echo "Error: wget or curl is required."
        exit 1
    fi

    if [ -z "$LATEST_TAG" ]; then
        echo "Error: Could not determine the latest release tag."
        exit 1
    fi

    BINARY_URL="https://github.com/eno314/DeveloperHelpTool-CLI/releases/download/$LATEST_TAG/developer-help-tool-cli-$LATEST_TAG-darwin-arm64"
    echo "Downloading $BINARY_NAME ($LATEST_TAG) for macOS arm64..."
else
    # Ensure version starts with 'v' if it's a typical semver
    if [[ ! "$VERSION" =~ ^v ]]; then
        VERSION="v$VERSION"
    fi
    BINARY_URL="https://github.com/eno314/DeveloperHelpTool-CLI/releases/download/$VERSION/developer-help-tool-cli-$VERSION-darwin-arm64"
    echo "Downloading $BINARY_NAME (version $VERSION) for macOS arm64..."
fi

# Check if we have write permission to INSTALL_DIR
if [ ! -w "$INSTALL_DIR" ]; then
    echo "Error: No write permission to $INSTALL_DIR."
    echo "Please run this script with sudo, e.g.:"
    echo "wget -qO- https://raw.githubusercontent.com/eno314/DeveloperHelpTool-CLI/main/install.sh | sudo bash"
    exit 1
fi

# Create a temporary file
TMP_FILE=$(mktemp)

# Download the binary
if command -v wget >/dev/null 2>&1; then
    wget -qO "$TMP_FILE" "$BINARY_URL"
elif command -v curl >/dev/null 2>&1; then
    curl -sLfL -o "$TMP_FILE" "$BINARY_URL"
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
