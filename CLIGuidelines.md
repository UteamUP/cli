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

### AI Media Analysis

The CLI never stores provider credentials or selects provider models. Image and video
commands upload media to authenticated, server-owned UteamUP endpoints. The backend resolves
the registered task route, Global Admin policy, Tenant BYOK connection, budget, fallback,
usage ledger, and model disclosure for the authenticated tenant.

Legacy `geminiApiKey`, `geminiModel`, and `googleMapsApiKey` fields are removed from
`~/.uteamup/config.json` the next time the CLI loads that profile. Legacy provider sections
in image/video YAML overrides are rejected instead of silently accepting unused secrets.

### Environment Variable Overrides

Env vars override config file values:

| Variable | Overrides |
|----------|-----------|
| `UTEAMUP_API_KEY` | `apiKey` |
| `UTEAMUP_SECRET` | `secret` |
| `UTEAMUP_API_BASE_URL` | `baseUrl` |
| `UTEAMUP_LOG_LEVEL` | `logLevel` |

---

## Image Analysis

### Usage

```bash
uteamup image analyze ./photos                              # Analyze images
uteamup image analyze ./photos --dry-run                     # Upload scope; no AI call
uteamup image analyze ./photos --output ./results            # Custom output dir
uteamup image analyze ./photos --timeout 5m                  # Finite backend timeout
ut img analyze ./photos --no-rename --verbose                # Skip renaming
```

### Requirements and boundary

- User must be logged in (`uteamup login`) to the tenant being analyzed.
- The active profile base URL must be an HTTPS origin without credentials, path, query, or
  fragment. Local self-signed TLS remains available through the existing `--insecure` flag.
- Photos are capped at 15 MB before decoding and normalized to JPEG before upload.
- The CLI sends no tenant integer ID, provider, model, task key, or provider credential.

---

## Video Analysis

### Usage

```bash
uteamup video analyze ./videos                               # Analyze all videos in folder
uteamup video analyze ./recording.mp4                         # Analyze a single video file
uteamup video analyze ./videos --dry-run                      # Upload scope; no AI call
ut vid analyze ./videos -o ./results --verbose                # Custom output, verbose
ut video analyze ./walkthrough.mov --timeout 10m              # Finite backend timeout
```

### Supported Formats

| Format | MIME Type | Action |
|--------|-----------|--------|
| MP4 | video/mp4 | Analyzed by video pipeline |
| MOV | video/quicktime | Analyzed by video pipeline |
| GIF | image/gif | Routed to image analyzer |
| Other | — | Skipped with warning |

File format detection uses magic bytes (file header), not file extensions.

### Requirements

- An authenticated UteamUP tenant session with the required AI and inventory permissions.
- No provider SDK, provider key, provider model, or external analyzer is required by the CLI.
- Videos are capped at 100 MB and must pass MP4/MOV magic-byte validation.

### How It Works

1. **Validate** — Scan input path, detect MIME types via magic bytes, route GIFs to image analyzer
2. **Upload + Analyze** — Stream each video to the authenticated UteamUP backend; the server executes the governed task route
3. **Deduplicate** — Merge duplicate entities within same video (temporal dedup) and across videos (grouping)
4. **Export** — Write CSVs: assets.csv, tools.csv, parts.csv, chemicals.csv, vendors.csv, locations.csv

### GPS and Location

Videos from mobile devices often contain GPS coordinates in container metadata. The video analyzer:
- Extracts GPS from MP4/MOV metadata atoms (©xyz, ISO 6709)
- Keeps detected coordinates local; the CLI does not fetch arbitrary or third-party URLs
- Assigns detected entities to their GPS-derived locations in the CSV output

### Vendor Enrichment

Vendor/manufacturer values returned by the governed backend analysis are deduplicated into
`vendors.csv`. The CLI performs no direct provider or web lookup.

### Cost Estimation

`--dry-run` reports deterministic file counts and bytes without making an AI call. Provider
cost cannot be inferred locally because routes and BYOK pricing vary; the backend response
returns the authoritative usage receipt and Tenant Admin budgets remain the hard ceiling.

---

## Release Process

> **IMPORTANT**: Releases are **fully automated** via Release-Please + GoReleaser GitHub Actions. The human action is **merging the Release-Please PR**, nothing more. Do NOT run `make release`, `goreleaser release`, or hand-create tags from your laptop. Doing so will collide with the automated tagger and produce duplicate releases.

### How a release happens

1. **Land conventional commits on `main`.** Prefixes drive the version bump:
   - `feat:` → minor bump
   - `fix:`, `chore:`, `refactor:`, `docs:`, `test:`, `perf:`, `ci:`, `build:`, `style:`, `remove:` → patch bump
   - `feat!:` / `fix!:` / `BREAKING CHANGE:` in body → major bump
2. **Release-Please opens (or updates) a "Release PR".** It bumps the version in `.release-please-manifest.json`, rewrites `CHANGELOG.md`, and titles the PR `chore(main): release X.Y.Z`.
3. **Merge the Release PR.** This is the only manual step.
4. **On merge, `.github/workflows/release-please.yml` does the rest:**
   - Creates the version tag (e.g. `1.2.0` — **no `v` prefix**, because `release-please-config.json` sets `"include-v-in-tag": false`).
   - Runs the `goreleaser` job: builds darwin/linux/windows × amd64/arm64, generates `checksums.txt`, creates `.deb` + `.rpm`, publishes the GitHub Release at `https://github.com/UteamUP/cli/releases/tag/<version>`.
   - As part of that same `goreleaser` run, GoReleaser's `brews:` block pushes the regenerated `Formula/uteamup.rb` (new version + SHA256s) straight to `UteamUP/homebrew-tap` using `HOMEBREW_TAP_GITHUB_TOKEN` (commits land as `goreleaserbot`). This is the **sole** tap-update mechanism — there is no `repository_dispatch` follow-up and the tap repo has no workflow listening for one.
5. **Verify the local Homebrew install (REQUIRED).** Always run this after a release:
   ```bash
   brew update
   brew upgrade uteamup
   uteamup version           # must show the new X.Y.Z, NOT --version (that flag does not exist)
   ```
   If `uteamup version` still reports the old version, the tap may not have refreshed yet — run `brew update` once more, then `brew upgrade uteamup`. If the formula at `https://github.com/UteamUP/homebrew-tap/blob/main/Formula/uteamup.rb` already shows the new version but `brew upgrade` won't pick it up, run `brew untap uteamup/tap && brew tap uteamup/tap` to force a fresh fetch.

### What you do NOT do

- ❌ `make release` — owned by the GitHub Action, never run locally.
- ❌ `git tag -a vX.Y.Z` / `git push origin vX.Y.Z` — Release-Please creates the tag.
- ❌ `git push github main` — there is no `github` remote. `origin` is already `github.com:UteamUP/cli.git`.
- ❌ Hand-edit `Formula/uteamup.rb` in the tap — GoReleaser overwrites it.
- ❌ Add a `v` prefix when referencing post-1.0 tags (`1.2.0` not `v1.2.0`). Older `v0.x` tags kept the prefix; new tags do not.

### Snapshot Build (testing GoReleaser locally without publishing)

```bash
make snapshot
```
Builds all platform artifacts into `dist/` without touching GitHub or the tap. Use to test `.goreleaser.yml` changes before merging.

### Manual Formula Update (only if automation is broken)

If GoReleaser's `brews:` push to `UteamUP/homebrew-tap` fails (e.g. `HOMEBREW_TAP_GITHUB_TOKEN` expired) and the tap is stuck on an old version, fall back to hand-editing the formula. Get user permission first.

```bash
git clone git@github.com:UteamUP/homebrew-tap.git
cd homebrew-tap
# Pull the new SHA256s from the release
curl -sL https://github.com/UteamUP/cli/releases/download/X.Y.Z/checksums.txt
# Edit Formula/uteamup.rb — bump version, swap URLs, swap sha256 lines for darwin amd64/arm64 + linux amd64/arm64
git add Formula/uteamup.rb
git commit -m "Update uteamup to X.Y.Z"
git push
```

### Optional: MSI for Windows

The standard release does not produce an MSI. If a user explicitly needs one, build it on a Windows runner with WiX Toolset v4+ and upload manually:

```bash
wix build packaging/msi/uteamup.wxs -o dist/uteamup.msi -arch x64
gh release upload X.Y.Z dist/uteamup.msi --repo UteamUP/cli
```

### Homebrew Tap Details

- **Tap repository**: https://github.com/UteamUP/homebrew-tap
- **Formula**: `Formula/uteamup.rb`
- **GoReleaser config**: `.goreleaser.yml` → `brews` section
- **Template**: `packaging/homebrew/uteamup.rb.tmpl` (reference only — GoReleaser generates the actual formula)
- **Update mechanism**: the `goreleaser` job in `release-please.yml` runs `goreleaser release`, whose `brews:` block pushes the regenerated `Formula/uteamup.rb` directly to the tap (authenticated with `HOMEBREW_TAP_GITHUB_TOKEN`, committed by `goreleaserbot`). There is no `repository_dispatch` step and the tap has no workflow listening for one.

### Troubleshooting Releases

| Problem | Likely cause | Fix |
|---------|--------------|-----|
| Release-Please PR didn't open | Latest commits don't use a recognized conventional prefix | Reword the commit (`git commit --amend` or land a new `chore:` commit) |
| `goreleaser` job failed in Actions | Build error / dirty tree on the runner | Re-run the workflow from the Actions tab; if persistent, run `make snapshot` locally to reproduce |
| Tap not updated after release | GoReleaser `brews:` push failed — usually an expired/invalid `HOMEBREW_TAP_GITHUB_TOKEN` (see the `goreleaser` job logs) | Re-run the `goreleaser` job after refreshing the token secret, or fall back to "Manual Formula Update" above |
| `brew upgrade` says "already up-to-date" but `uteamup version` is stale | Tap cache not refreshed | `brew update`, then `brew upgrade uteamup`. If still stale: `brew untap uteamup/tap && brew tap uteamup/tap && brew install uteamup` |
| Tag already exists | Tried to re-release the same version | Land a new commit so Release-Please proposes the next version. Do NOT delete an existing public tag |

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

### REST routing for domain actions

Although domain actions declare an MCP `ToolName`, the runtime path in
`internal/registry/registry.go::runCommand` calls `apiClient.CallREST(...)` —
so the HTTP method and URL are built from the action's `Name`, the domain's
`APIPath` (or auto-derived from `Name`), and the positional args via
`buildRESTPath` + the `HTTPMethod` map.

| Action name     | HTTP method | URL pattern                                       |
|-----------------|-------------|---------------------------------------------------|
| `list`          | `GET`       | `{basePath}[?query...]` (or `{basePath}/{RESTPath}`) |
| `get`           | `GET`       | `{basePath}/{id|externalGuid}`                   |
| `create`        | `POST`      | `{basePath}` (body = args)                        |
| `update`        | `PUT`       | `{basePath}/{id|externalGuid}`                   |
| `update-status` | `PATCH`     | `{basePath}/{id|externalGuid}/status`            |
| `delete`        | `DELETE`    | `{basePath}/{id|externalGuid}`                   |
| `search`        | `GET`       | `{basePath}/search` (or `{basePath}/{RESTPath}`)  |

GUID-first domains (every new domain, per `Guidelines/ApiGuidelines.md`)
should declare their positional arg as `externalGuid`; legacy integer-id domains
can keep `id`. `buildRESTPath` accepts either.

### CSRF header on mutating calls

`CallREST` sets `X-Requested-With: XMLHttpRequest` on every outgoing request.
Backend `[Authorize(Policy = "BugsAndFeaturesCreate")]`-style policies (and the
bug-create CSRF guard) reject POST/PUT/PATCH/DELETE without it. Do NOT strip
the header in a new adapter or middleware.

### Backend auth policy for debug-user access

The email/password login on `POST /api/auth/login` issues a JWT validated only
by the `"Local"` auth scheme. Controllers that historically stacked
`[Authorize(Policy="AzureAdPolicy"), Authorize(Policy="LocalPolicy")]` work
fine for admin-UI Entra users but fail for the debug service account when a
third scheme (e.g. Google) is listed on the same policy — Google's JWT Bearer
challenge short-circuits the chain with 401. If you add a new CLI-facing
controller, prefer `[Authorize(Policy = "LocalOrAzureAdPolicy")]` (single
policy, both schemes declared on it) over stacked attributes.

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
| `origin` | `https://github.com/UteamUP/cli.git` | Sole remote — public releases, Homebrew tap, CI |

There is **no separate `github` remote** anymore. Push normally:
```bash
git push origin main
```

Tags are created by the Release-Please GitHub Action — do not push tags from your laptop.
