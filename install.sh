#!/bin/sh
# =============================================================================
# go-qo Installation Script
#
# Usage:
#   curl ... | sh                     -> Installs latest version
#   curl ... | VERSION=v1.0.0 sh      -> Installs specific version
#   curl ... | BINDIR=./custom/bin sh -> Installs specific bin directory
# =============================================================================

set -e

OWNER=kiki-ki
REPO=go-qo
BINARY=qo
FORMAT=tar.gz
BINDIR=${BINDIR:-./bin}
VERSION=${VERSION:-}

uname_os() {
  os=$(uname -s | tr '[:upper:]' '[:lower:]')
  case "$os" in
    cygwin_nt*) os="windows" ;;
    mingw*)     os="windows" ;;
    msys_nt*)   os="windows" ;;
  esac
  echo "$os"
}

uname_arch() {
  arch=$(uname -m)
  case $arch in
    x86_64) arch="amd64" ;;
    x86)    arch="i386" ;;
    i686)   arch="i386" ;;
    aarch64) arch="arm64" ;;
    armv8*)  arch="arm64" ;;
    arm64)   arch="arm64" ;;
  esac
  echo ${arch}
}

OS=$(uname_os)
ARCH=$(uname_arch)

if [ -n "$VERSION" ]; then
  # if VERSION is set to a specific version
  echo "Looking for version: ${VERSION} for ${OS}/${ARCH}..."
  API_URL="https://api.github.com/repos/${OWNER}/${REPO}/releases/tags/${VERSION}"
else
  # if VERSION is not set, get the latest release
  echo "Looking for latest release for ${OS}/${ARCH}..."
  API_URL="https://api.github.com/repos/${OWNER}/${REPO}/releases/latest"
fi

# Get download URL from GitHub API
DOWNLOAD_URL=$(curl -s "${API_URL}" \
  | grep "browser_download_url" \
  | grep -i "${OS}_${ARCH}.${FORMAT}" \
  | cut -d '"' -f 4)

if [ -z "$DOWNLOAD_URL" ]; then
  echo "Error: Could not find download URL for ${OS}/${ARCH}"
  echo "    Target: ${API_URL}"
  exit 1
fi

echo "Downloading ${DOWNLOAD_URL}..."

TMPDIR=$(mktemp -d)
curl -sL "$DOWNLOAD_URL" | tar xz -C "$TMPDIR"

if [ ! -d "$BINDIR" ]; then
  mkdir -p "$BINDIR"
fi

echo "Installing ${BINARY} to ${BINDIR}..."

mv "$TMPDIR/$BINARY" "$BINDIR/"
chmod +x "$BINDIR/$BINARY"

echo "Successfully installed ${BINARY} to ${BINDIR}"
rm -rf "$TMPDIR"
