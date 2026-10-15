#!/bin/sh
# Installs the latest sutra CLI release for Linux or macOS.
#   curl -fsSL https://raw.githubusercontent.com/iCoderabhishek/Sutra-AI/HEAD/scripts/install.sh | sh
# Set SUTRA_INSTALL_DIR to change the install location (default: ~/.local/bin).
set -eu

REPO="iCoderabhishek/Sutra-AI"
INSTALL_DIR="${SUTRA_INSTALL_DIR:-$HOME/.local/bin}"

case "$(uname -s)" in
  Linux) os=linux ;;
  Darwin) os=darwin ;;
  *) echo "Unsupported OS: $(uname -s). On Windows use install.ps1." >&2; exit 1 ;;
esac

case "$(uname -m)" in
  x86_64 | amd64) arch=amd64 ;;
  arm64 | aarch64) arch=arm64 ;;
  *) echo "Unsupported architecture: $(uname -m)" >&2; exit 1 ;;
esac

# Resolve the latest tag from the redirect, so no API token or jq is needed.
tag=$(curl -fsSLI -o /dev/null -w '%{url_effective}' "https://github.com/$REPO/releases/latest")
tag=${tag##*/}
case "$tag" in
  v*) ;;
  *) echo "Could not find the latest release." >&2; exit 1 ;;
esac
version=${tag#v}

archive="sutra_${version}_${os}_${arch}.tar.gz"
base="https://github.com/$REPO/releases/download/$tag"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

echo "Downloading sutra $tag ($os/$arch)..."
curl -fsSL "$base/$archive" -o "$tmp/$archive"
curl -fsSL "$base/checksums.txt" -o "$tmp/checksums.txt"

# Verify the archive against the published checksums.
expected=$(grep " $archive\$" "$tmp/checksums.txt" | cut -d' ' -f1)
if command -v sha256sum >/dev/null 2>&1; then
  actual=$(sha256sum "$tmp/$archive" | cut -d' ' -f1)
else
  actual=$(shasum -a 256 "$tmp/$archive" | cut -d' ' -f1)
fi
if [ -z "$expected" ] || [ "$expected" != "$actual" ]; then
  echo "Checksum mismatch for $archive" >&2
  exit 1
fi

tar -xzf "$tmp/$archive" -C "$tmp"
mkdir -p "$INSTALL_DIR"
install -m 755 "$tmp/sutra" "$INSTALL_DIR/sutra"

echo "Installed sutra $tag to $INSTALL_DIR/sutra"
case ":$PATH:" in
  *":$INSTALL_DIR:"*) echo "Run: sutra" ;;
  *) echo "Add it to your PATH:  export PATH=\"$INSTALL_DIR:\$PATH\"" ;;
esac
