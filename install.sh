#!/bin/sh
set -e

# gilo installer
# Usage: curl -fsSL https://raw.githubusercontent.com/andreasbaumgartner/gilo/main/install.sh | sh

REPO="andreasbaumgartner/gilo"
INSTALL_DIR="${GILO_INSTALL_DIR:-/usr/local/bin}"

# Detect OS
OS="$(uname -s)"
case "$OS" in
  Linux)  OS="linux" ;;
  Darwin) OS="darwin" ;;
  *)
    echo "Error: unsupported operating system: $OS" >&2
    exit 1
    ;;
esac

# Detect architecture
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64|amd64)  ARCH="amd64" ;;
  aarch64|arm64)  ARCH="arm64" ;;
  *)
    echo "Error: unsupported architecture: $ARCH" >&2
    exit 1
    ;;
esac

# Get latest version from GitHub API
echo "Fetching latest version..."
VERSION="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')"

if [ -z "$VERSION" ]; then
  echo "Error: could not determine latest version" >&2
  exit 1
fi

echo "Installing gilo ${VERSION} (${OS}/${ARCH})..."

# Download binary
BINARY_NAME="gilo_${OS}_${ARCH}"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY_NAME}"

TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

curl -fsSL -o "${TMPDIR}/gilo" "$DOWNLOAD_URL"

# Verify checksum
CHECKSUM_URL="https://github.com/${REPO}/releases/download/${VERSION}/checksums.txt"
curl -fsSL -o "${TMPDIR}/checksums.txt" "$CHECKSUM_URL"
EXPECTED="$(grep "${BINARY_NAME}" "${TMPDIR}/checksums.txt" | awk '{print $1}')"
if command -v sha256sum >/dev/null 2>&1; then
  ACTUAL="$(sha256sum "${TMPDIR}/gilo" | awk '{print $1}')"
else
  ACTUAL="$(shasum -a 256 "${TMPDIR}/gilo" | awk '{print $1}')"
fi
if [ "$EXPECTED" != "$ACTUAL" ]; then
  echo "Error: checksum mismatch (expected ${EXPECTED}, got ${ACTUAL})" >&2
  exit 1
fi

chmod +x "${TMPDIR}/gilo"

# Install
if [ -w "$INSTALL_DIR" ]; then
  mv "${TMPDIR}/gilo" "${INSTALL_DIR}/gilo"
else
  echo "Installing to ${INSTALL_DIR} (requires sudo)..."
  sudo mv "${TMPDIR}/gilo" "${INSTALL_DIR}/gilo"
fi

echo "gilo ${VERSION} installed to ${INSTALL_DIR}/gilo"
