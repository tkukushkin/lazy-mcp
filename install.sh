#!/bin/sh
set -e

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64)  ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
esac

URL="https://github.com/tkukushkin/lazy-mcp/releases/latest/download/lazy-mcp-${OS}-${ARCH}"
DEST="/usr/local/bin/lazy-mcp"

echo "Downloading lazy-mcp for ${OS}/${ARCH}..."
curl -fsSL "$URL" -o "$DEST"
chmod +x "$DEST"
echo "Installed lazy-mcp to ${DEST}"
