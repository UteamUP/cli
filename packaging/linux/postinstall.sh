#!/bin/sh
# Post-install script for UteamUP CLI (.deb / .rpm)
set -e

# Create ut symlink
ln -sf /usr/bin/uteamup /usr/bin/ut

# Generate shell completions if shells are available
if command -v bash >/dev/null 2>&1; then
    mkdir -p /usr/share/bash-completion/completions
    uteamup completion bash > /usr/share/bash-completion/completions/uteamup 2>/dev/null || true
    cp /usr/share/bash-completion/completions/uteamup /usr/share/bash-completion/completions/ut 2>/dev/null || true
fi

if command -v zsh >/dev/null 2>&1; then
    mkdir -p /usr/share/zsh/vendor-completions
    uteamup completion zsh > /usr/share/zsh/vendor-completions/_uteamup 2>/dev/null || true
fi

if command -v fish >/dev/null 2>&1; then
    mkdir -p /usr/share/fish/vendor_completions.d
    uteamup completion fish > /usr/share/fish/vendor_completions.d/uteamup.fish 2>/dev/null || true
fi

echo "UteamUP CLI installed successfully."
echo "Run 'uteamup login' or 'ut login' to get started."
