#!/bin/sh
set -e

REPO="nmashchenko/aegis-cli"
BINARY="aegis"
INSTALL_DIR="/usr/local/bin"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  arm64)   ARCH="arm64" ;;
  *)       echo "Unsupported architecture: $ARCH" && exit 1 ;;
esac

case "$OS" in
  darwin|linux) ;;
  *) echo "Unsupported OS: $OS" && exit 1 ;;
esac

LATEST=$(curl -sI "https://github.com/$REPO/releases/latest" | grep -i "^location:" | sed 's/.*tag\///' | tr -d '\r\n')

if [ -z "$LATEST" ]; then
  echo "Failed to fetch latest release"
  exit 1
fi

URL="https://github.com/$REPO/releases/download/${LATEST}/${BINARY}_${OS}_${ARCH}.tar.gz"

echo "Installing $BINARY $LATEST ($OS/$ARCH)..."

TMP=$(mktemp -d)
curl -sL "$URL" | tar xz -C "$TMP"

if [ -w "$INSTALL_DIR" ]; then
  mv "$TMP/$BINARY" "$INSTALL_DIR/$BINARY"
else
  sudo mv "$TMP/$BINARY" "$INSTALL_DIR/$BINARY"
fi

rm -rf "$TMP"

echo "Installed $BINARY to $INSTALL_DIR/$BINARY"
