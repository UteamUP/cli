# UteamUP CLI — Development & Release Guidelines

## Overview

The UteamUP CLI (`uteamup` / `ut`) is a Go-based command-line tool that mirrors MCP server capabilities as terminal commands. It uses OAuth 2.0 + PKCE authentication and communicates with the UteamUP backend via JSON-RPC 2.0.

---

## Build & Development

### Prerequisites

- Go 1.22+
- golangci-lint (for linting)
- GoReleaser v2 (for releases)
- WiX Toolset v4+ (for Windows MSI, optional)

### Common Commands

```bash
make build          # Build bin/uteamup + bin/ut (current platform)
make test           # Run all tests with race detection
make lint           # Run golangci-lint
make check          # fmt → vet → lint → test → build (pre-commit)
make install        # Build + install to /usr/local/bin (adds PATH to .zshrc)
make uninstall      # Remove from /usr/local/bin
make snapshot       # GoReleaser snapshot (all platforms, no publish)
make release        # Full GoReleaser release (requires tag)
```

### Version Injection

Version is injected at build time via LDFLAGS:

```go
// main.go
var (
    version = "dev"
    commit  = "none"
    date    = "unknown"
)
```

`make build` uses `git describe --tags --always --dirty` for the version string.

---

## Configuration

### Config File Location

`~/.uteamup/config.json` — multi-profile JSON config.

### Setup

```bash
uteamup config init                              # Interactive setup
ut config set baseUrl https://localhost:5002      # Set a value
ut config profile dev                             # Switch profile
ut config show                                    # Display config (redacted)
```

### Gemini AI Configuration (Image Analysis)

```bash
ut config apikey AIzaSy...                        # Set Gemini API key
ut config apikey                                  # Show current key (redacted)
ut config model gemini-3.1-pro-preview            # Set default model
ut config model                                   # Show current model
ut config model list                              # List available models
```

**Available Gemini models:**

| Model | Type | Notes |
|-------|------|-------|
| `gemini-pro-latest` | Pro | Rolling alias, always newest |
| `gemini-3.1-pro-preview` | Pro | Latest explicit pro |
| `gemini-3.1-flash-lite-preview` | Flash Lite | Default — fastest, cheapest |
| `gemini-3-pro-preview` | Pro | Previous gen |
| `gemini-3-flash-preview` | Flash | Previous gen |
| `gemini-2.5-pro` | Pro | Stable |
| `gemini-2.5-flash` | Flash | Stable |

### Environment Variable Overrides

Env vars override config file values:

| Variable | Overrides |
|----------|-----------|
| `UTEAMUP_API_KEY` | `apiKey` |
| `UTEAMUP_SECRET` | `secret` |
| `UTEAMUP_API_BASE_URL` | `baseUrl` |
| `UTEAMUP_LOG_LEVEL` | `logLevel` |
| `GEMINI_API_KEY` | `geminiApiKey` |
| `GEMINI_MODEL` | `geminiModel` |

---

## Image Analysis

### Usage

```bash
uteamup image analyze ./photos                              # Analyze images
uteamup image analyze ./photos --dry-run                     # Cost estimate only
uteamup image analyze ./photos --model gemini-3.1-pro-preview  # Override model
uteamup image analyze ./photos --output ./results            # Custom output dir
ut img analyze ./photos --no-rename --verbose                # Skip renaming
```

### Requirements

- UteamUP Image Analyzer Python tool must be installed
- Python 3.10+ with virtual environment
- User must be logged in (`uteamup login`)

### Analyzer Discovery

The CLI locates the Image Analyzer in this order:
1. `UTEAMUP_IMAGE_ANALYZER_PATH` environment variable
2. Sibling directory `../UteamUP_ImageAnalyzer` relative to the CLI binary
3. `~/UteamUP_ImageAnalyzer`

### Installing the Analyzer

```bash
git clone https://github.com/UteamUP/ImageAnalyzer ~/UteamUP_ImageAnalyzer
cd ~/UteamUP_ImageAnalyzer
python3 -m venv .venv
.venv/bin/pip install -r requirements.txt
cp .env.example .env  # Add your GEMINI_API_KEY
```

---

## Release Process

> **IMPORTANT**: Every release MUST follow these steps. Skipping any step will result in users not getting updates via `brew upgrade`, missing binaries, or stale Homebrew formulas.

### Prerequisites (One-Time Setup)

1. **GITHUB_TOKEN** — Required for GoReleaser to create releases and push to Homebrew tap.
   ```bash
   # Create a GitHub Personal Access Token with `repo` scope at:
   # https://github.com/settings/tokens
   # Then export it:
   export GITHUB_TOKEN=ghp_your_token_here

   # Optionally add to shell profile for persistence:
   echo 'export GITHUB_TOKEN=ghp_your_token_here' >> ~/.zshrc
   ```

2. **HOMEBREW_TAP_GITHUB_TOKEN** — Used by GoReleaser to push the formula to the tap.
   This can be the same as `GITHUB_TOKEN` if it has `repo` scope on `UteamUP/homebrew-tap`.
   ```bash
   export HOMEBREW_TAP_GITHUB_TOKEN=$GITHUB_TOKEN
   ```

3. **GoReleaser** — Must be installed:
   ```bash
   brew install goreleaser   # macOS
   ```

### Step-by-Step Release Checklist

**Every time a new version is released, ALL of these steps must be completed:**

```bash
# === Step 1: Update CHANGELOG.md ===
# Move [Unreleased] items to a new version section
# Example: ## [0.4.0] — 2026-04-01

# === Step 2: Commit the changelog and any final changes ===
git add CHANGELOG.md
git commit -m "docs: update changelog for vX.Y.Z"

# === Step 3: Push code to BOTH remotes ===
git push origin main
git push github main

# === Step 4: Create and push the version tag ===
git tag -a vX.Y.Z -m "Release vX.Y.Z — brief description"
git push origin vX.Y.Z
git push github vX.Y.Z

# === Step 5: Run GoReleaser (requires GITHUB_TOKEN) ===
export GITHUB_TOKEN=ghp_your_token_here
export HOMEBREW_TAP_GITHUB_TOKEN=$GITHUB_TOKEN
make release
# OR: goreleaser release --clean

# === Step 6 (Optional): Build MSI for Windows ===
# Requires WiX Toolset v4+ on Windows/CI
wix build packaging/msi/uteamup.wxs -o dist/uteamup.msi -arch x64
# Then upload MSI to the GitHub release manually:
gh release upload vX.Y.Z dist/uteamup.msi --repo UteamUP/cli
```

### What GoReleaser Does Automatically (Step 5)

When you run `make release`, GoReleaser:

1. **Builds** all 6 platform binaries (darwin/linux/windows × amd64/arm64)
2. **Creates archives** (tar.gz for Unix, zip for Windows)
3. **Generates** SHA256 checksums (`checksums.txt`)
4. **Creates** GitHub release at https://github.com/UteamUP/cli/releases with all artifacts
5. **Updates Homebrew formula** at https://github.com/UteamUP/homebrew-tap/blob/main/Formula/uteamup.rb
   - Auto-fills version, download URLs, and SHA256 hashes
   - Users get the update via `brew update && brew upgrade uteamup`
6. **Generates** .deb and .rpm Linux packages

### Homebrew Tap Details

- **Tap repository**: https://github.com/UteamUP/homebrew-tap
- **Formula**: `Formula/uteamup.rb`
- **GoReleaser config**: `.goreleaser.yml` → `brews` section
- **Template**: `packaging/homebrew/uteamup.rb.tmpl` (reference only — GoReleaser generates the actual formula)

**How `brew upgrade` works:**
1. GoReleaser pushes updated `uteamup.rb` to `UteamUP/homebrew-tap` with new version + SHA256s
2. User runs `brew update` → pulls latest formula from the tap
3. User runs `brew upgrade uteamup` → downloads new binary from GitHub release

**If Homebrew auto-update fails**, manually update the formula:
```bash
git clone git@github.com:UteamUP/homebrew-tap.git
cd homebrew-tap
# Edit Formula/uteamup.rb — update version, URLs, and SHA256s
# Get SHA256s from the release checksums.txt:
curl -sL https://github.com/UteamUP/cli/releases/download/vX.Y.Z/checksums.txt
git add Formula/uteamup.rb
git commit -m "Update uteamup to vX.Y.Z"
git push
```

### Snapshot Build (Testing)

To test the release pipeline without publishing:

```bash
make snapshot
```

This creates all binaries in `dist/` without pushing to GitHub or Homebrew.

### Troubleshooting Releases

| Problem | Cause | Fix |
|---------|-------|-----|
| `missing GITHUB_TOKEN` | Token not exported | `export GITHUB_TOKEN=ghp_...` |
| Homebrew formula not updated | `HOMEBREW_TAP_GITHUB_TOKEN` missing or no repo scope | Set token with `repo` scope on `UteamUP/homebrew-tap` |
| `brew upgrade` shows "already up-to-date" | Formula not pushed or `brew update` not run | Run `brew update` first; check tap formula version |
| Tag already exists | Trying to re-release | Delete tag: `git tag -d vX.Y.Z && git push origin :refs/tags/vX.Y.Z` |
| GoReleaser fails on build | Code doesn't compile | Run `make check` before releasing |

---

## Installation Methods

### macOS — Homebrew (Recommended)

```bash
brew tap uteamup/tap
brew install uteamup
```

**Upgrade:**
```bash
brew update && brew upgrade uteamup
```

**Downgrade to specific version:**
```bash
brew uninstall uteamup
brew install uteamup@0.2.0  # Or download specific release from GitHub
```

**Uninstall:**
```bash
brew uninstall uteamup
```

### macOS — .pkg Installer

Download the `.pkg` from the [GitHub releases page](https://github.com/UteamUP/cli/releases).

**Build .pkg locally:**
```bash
make build
./packaging/macos/build-pkg.sh
```

Installs to `/usr/local/bin/uteamup` with `ut` symlink. Includes welcome screen and license.

### Linux — apt (Debian/Ubuntu)

```bash
# Download .deb from GitHub releases
sudo dpkg -i uteamup_0.3.0_linux_amd64.deb
```

**Upgrade:**
```bash
sudo dpkg -i uteamup_0.3.0_linux_amd64.deb  # Install newer .deb
```

### Linux — rpm (RHEL/Fedora)

```bash
sudo rpm -i uteamup_0.3.0_linux_amd64.rpm
```

**Upgrade:**
```bash
sudo rpm -U uteamup_0.3.0_linux_amd64.rpm
```

### Linux — Manual (tar.gz)

```bash
tar -xzf uteamup_0.3.0_linux_amd64.tar.gz
sudo mv uteamup /usr/local/bin/
sudo ln -sf /usr/local/bin/uteamup /usr/local/bin/ut
```

### Windows — MSI Installer

Download the `.msi` from [GitHub releases](https://github.com/UteamUP/cli/releases).

- Installs to `Program Files/UteamUP CLI/`
- Automatically adds to PATH
- Installs both `uteamup.exe` and `ut.exe`
- Creates Start Menu shortcuts

**Upgrade:** Run the newer MSI — it auto-detects and upgrades the previous version.

**Uninstall:** Control Panel → Programs → UteamUP CLI → Uninstall

### All Platforms — From Source

```bash
git clone https://github.com/UteamUP/cli.git
cd cli
make install
```

---

## Homebrew Tap Management

The Homebrew formula is at `uteamup/homebrew-tap` on GitHub.

### How It Works

1. GoReleaser uses the template at `packaging/homebrew/uteamup.rb.tmpl`
2. On release, GoReleaser:
   - Fills in version, download URLs, SHA256 checksums
   - Pushes the rendered formula to `uteamup/homebrew-tap`
3. Users get updates via `brew update && brew upgrade uteamup`

### Required Token

GoReleaser needs `HOMEBREW_TAP_GITHUB_TOKEN` set to push to the tap repo. This is a GitHub PAT with `repo` scope for the `UteamUP/homebrew-tap` repository.

### Manual Formula Update

If GoReleaser doesn't update the tap automatically:

1. Clone the tap: `git clone git@github.com:UteamUP/homebrew-tap.git`
2. Edit `Formula/uteamup.rb`:
   - Update `version`
   - Update download URLs to point to new release
   - Update SHA256 checksums (from `checksums.txt` in the release)
3. Commit and push

### Testing the Formula

```bash
brew install --build-from-source ./Formula/uteamup.rb
brew test uteamup
```

---

## Architecture Quick Reference

### Adding a New Command

**MCP-backed commands** (via backend): Add a domain registry file:
```
internal/registry/domains_<name>.go
```

**Local commands** (no backend): Add a Cobra command file:
```
cmd/<name>.go
```
Register in `cmd/root.go` `init()` function.

### Auth Exemption

Commands that don't require login are listed in `cmd/root.go`:
```go
var commandsExemptFromAuth = map[string]bool{
    "login":      true,
    "logout":     true,
    "version":    true,
    "completion": true,
    "config":     true,
    "help":       true,
}
```

### Pre-commit Checklist

```bash
go vet ./...            # No vet errors
go test ./... -race     # All tests pass
make build              # Build succeeds
```

---

## Git Remotes

| Remote | URL | Purpose |
|--------|-----|---------|
| `origin` | `ssh.dev.azure.com:v3/UteamUP/UteamUP_CLI/UteamUP_CLI` | Primary (Azure DevOps) |
| `github` | `github.com:UteamUP/cli.git` | Public releases, Homebrew tap |

Always push to both:
```bash
git push origin main --tags
git push github main --tags
```
