#!/bin/bash

set -e

REPO="Phantomvv1/KayTrade"
BINARY="kaytrade"
INSTALL_DIR="/usr/local/bin"

URL="https://github.com/$REPO/releases/download/v0.1.5/$BINARY"

echo "📦 Installing $BINARY (latest release)"
echo "⬇️  Downloading from:"
echo "   $URL"
echo

if [ -w "$INSTALL_DIR" ]; then
  curl -fL "$URL" -o "$INSTALL_DIR/$BINARY"
else
  echo "🔐 Installing to $INSTALL_DIR (sudo required)"
  curl -fL "$URL" -o "/tmp/$BINARY"
  chmod +x "/tmp/$BINARY"
  sudo mv "/tmp/$BINARY" "$INSTALL_DIR/$BINARY"
fi

chmod +x "$INSTALL_DIR/$BINARY"

echo
echo "✅ Installation complete!"
echo "📍 Binary installed at: $INSTALL_DIR/$BINARY"
echo "▶️  Run: $BINARY --version"
