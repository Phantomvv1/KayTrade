#! /bin/bash
set -e

REPO="Phantomvv1/KayTrade"
BINARY="kaytrade"
INSTALL_DIR="/usr/local/bin"

LATEST_TAG="$(curl -s https://api.github.com/repos/$REPO/releases/latest \
  | grep '"tag_name"' \
  | cut -d '"' -f 4)"

URL="https://github.com/$REPO/releases/download/$LATEST_TAG/$BINARY"

echo "📦 Installing $BINARY $LATEST_TAG"
echo "⬇️  Downloading $URL"

if [ -w "$INSTALL_DIR" ]; then
  curl -fL "$URL" -o "$INSTALL_DIR/$BINARY"
else
  echo "🔐 Installing to $INSTALL_DIR (sudo required)"
  curl -fL "$URL" -o "/tmp/$BINARY"
  chmod +x "/tmp/$BINARY"
  sudo mv "/tmp/$BINARY" "$INSTALL_DIR/$BINARY"
fi

chmod +x "$INSTALL_DIR/$BINARY"

echo "✅ Installed successfully!"
echo "📍 Binary location: $INSTALL_DIR/$BINARY"
echo "▶️  Run: $BINARY --help"
