#!/bin/bash
# Build macOS .pkg installer for UteamUP CLI
set -euo pipefail

VERSION="${1:-dev}"
ARCH="${2:-$(uname -m)}"
BINARY="bin/uteamup"
PKG_ID="com.uteamup.cli"
INSTALL_DIR="/usr/local/bin"

if [ ! -f "$BINARY" ]; then
    echo "Error: Binary not found at $BINARY"
    echo "Run 'make build' first."
    exit 1
fi

# Create staging directory
STAGING=$(mktemp -d)
trap "rm -rf $STAGING" EXIT

mkdir -p "$STAGING/payload${INSTALL_DIR}"
cp "$BINARY" "$STAGING/payload${INSTALL_DIR}/uteamup"
chmod 755 "$STAGING/payload${INSTALL_DIR}/uteamup"

# Create ut symlink
ln -sf uteamup "$STAGING/payload${INSTALL_DIR}/ut"

# Create postinstall script
mkdir -p "$STAGING/scripts"
cat > "$STAGING/scripts/postinstall" << 'SCRIPT'
#!/bin/bash
ln -sf /usr/local/bin/uteamup /usr/local/bin/ut
echo "UteamUP CLI installed. Run 'uteamup login' or 'ut login' to get started."
SCRIPT
chmod 755 "$STAGING/scripts/postinstall"

# Build component package
pkgbuild \
    --root "$STAGING/payload" \
    --identifier "$PKG_ID" \
    --version "$VERSION" \
    --scripts "$STAGING/scripts" \
    --install-location "/" \
    "$STAGING/uteamup-component.pkg"

# Build product archive (adds welcome/license screens)
mkdir -p dist
productbuild \
    --package "$STAGING/uteamup-component.pkg" \
    "dist/uteamup-${VERSION}-macos-${ARCH}.pkg"

echo "Created dist/uteamup-${VERSION}-macos-${ARCH}.pkg"
