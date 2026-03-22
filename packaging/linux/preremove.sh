#!/bin/sh
# Pre-remove script for UteamUP CLI (.deb / .rpm)

# Remove ut symlink
rm -f /usr/bin/ut

# Remove shell completions
rm -f /usr/share/bash-completion/completions/uteamup
rm -f /usr/share/bash-completion/completions/ut
rm -f /usr/share/zsh/vendor-completions/_uteamup
rm -f /usr/share/fish/vendor_completions.d/uteamup.fish

echo "UteamUP CLI removed."
