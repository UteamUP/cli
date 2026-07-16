# UteamUP CLI

Command-line interface for the [UteamUP](https://uteamup.com) platform. Manage assets, work orders, users, and more from any terminal. Includes AI-powered image and video analysis for CMMS inventory onboarding.

Both `uteamup` and `ut` (shortname) are installed — they are identical.

## Installation

### macOS

**Homebrew** (recommended):
```bash
brew tap uteamup/tap
brew install uteamup
```

**macOS Installer (.pkg)**:
Download the `.pkg` file from [Releases](https://github.com/uteamup/cli/releases) and run it.

### Linux

**Debian / Ubuntu (.deb)**:
```bash
sudo dpkg -i uteamup_0.6.2_amd64.deb
```

**Fedora / RHEL / CentOS (.rpm)**:
```bash
sudo rpm -i uteamup-0.6.2.x86_64.rpm
```

**Manual (tar.gz)**:
```bash
tar xzf uteamup_0.6.2_linux_amd64.tar.gz
sudo mv uteamup /usr/local/bin/
sudo ln -sf /usr/local/bin/uteamup /usr/local/bin/ut
```

### Windows

**MSI Installer** (recommended):
Download the `.msi` file from [Releases](https://github.com/uteamup/cli/releases) and run the installer wizard. It adds `uteamup.exe` and `ut.exe` to your PATH.

**Manual (zip)**:
Extract and add the folder to your system PATH.

### From Source

```bash
go install github.com/uteamup/cli@latest
```

Requires Go 1.22+.

---

## Quick Start

```bash
# 1. Create config file
uteamup config init

# 2. Authenticate
uteamup login                    # interactive (email/password)
# or
ut login --api-key=KEY --api-secret=SECRET   # API key

# 3. Use it
ut asset list
ut workorder get 123 -o json
ut user list --page 1 --page-size 10
```

---

## Image Analysis

AI-powered image analysis for bulk CMMS inventory onboarding. The CLI sends
validated media to UteamUP's authenticated AI gateway and exports structured CSV
data ready for human review and import. The backend owns the task and model route,
so managed AI or Tenant BYOK is applied consistently.

### Setup

```bash
ut login                   # Authenticate first
ut tenant select           # Select a tenant if your account has several
```

Provider credentials and model selection are never stored in the CLI. Tenant
admins configure BYOK in UteamUP's web administration surface.

### Usage

```bash
ut image analyze ./photos                    # Analyze all images in folder
ut image analyze ./photos --dry-run          # Validate and show upload scope
ut image analyze ./photos --output ./results # Custom output directory
ut img analyze ./photos --no-rename          # Keep original filenames
ut img analyze ./photos --resume             # Resume interrupted analysis
ut img analyze ./photos --timeout 10m         # Per-request timeout
```

### What It Does

1. **Scans** folder for images (JPG, PNG, HEIC, WebP, TIFF, BMP)
2. **Detects** iPhone edit pairs and duplicates automatically
3. **Uploads safely** through an authenticated, HTTPS-only UteamUP endpoint
4. **Classifies** each image as asset, tool, part, chemical, or unclassified
5. **Extracts** CMMS fields: name, serial number, model, manufacturer, condition, etc.
6. **Detects multiple entities** per image (e.g., a machine with visible parts)
7. **Links relationships** — parts/tools/chemicals linked to parent assets via `related_to`
8. **Preserves local GPS** coordinates from EXIF without calling mapping services
9. **Groups** duplicate images of the same item across photos
10. **Exports** review CSVs and a summary report
11. **Renames** images with descriptive filenames when enabled

### CSV Output

| File | Contents |
|------|----------|
| `assets.csv` | Equipment, machinery, vehicles with vendor/location links |
| `tools.csv` | Handheld tools and instruments |
| `parts.csv` | Spare parts and components |
| `chemicals.csv` | Chemical products with GHS/safety data |
| `vendors.csv` | Deduplicated manufacturers detected in governed analysis results |
| `locations.csv` | Locally extracted GPS and model-suggested location data |
| `summary_report.md` | Analysis summary with counts and timing |

### Requirements

- Must be logged in (`uteamup login`)
- An authenticated tenant with permission to use the inventory AI task
- HTTPS backend URL, except where local-development policy explicitly permits otherwise
- Images no larger than 15 MB each

---

## Video Analysis

AI-powered video analysis for CMMS inventory. Walk through a facility recording video, and the analyzer identifies equipment, tools, parts, and chemicals with timestamps.

### Usage

```bash
ut video analyze ./videos                    # Analyze all videos in folder
ut video analyze ./recording.mp4             # Analyze a single video
ut video analyze ./videos --dry-run          # Validate and show upload scope
ut vid analyze ./videos --timeout 10m         # Per-request timeout
```

### Supported Formats

| Format | MIME Type | Action |
|--------|-----------|--------|
| MP4 | video/mp4 | Analyzed by video pipeline |
| MOV | video/quicktime | Analyzed by video pipeline |
| GIF | image/gif | Routed to image analyzer |

### What It Does

1. **Validates** files via magic byte MIME detection (not extensions)
2. **Streams** video to the authenticated UteamUP AI gateway
3. **Analyzes** video frames for CMMS entities with timestamp detection (MM:SS)
4. **Extracts GPS** from MP4/MOV container metadata (©xyz and ISO 6709 atoms)
5. **Deduplicates** entities across frames (temporal, 30s window) and across videos (grouping)
6. **Exports** the same review CSV format as the image analyzer

### Authentication & Tenant Required

Video analysis requires UteamUP login and an active tenant subscription plan:

```bash
ut login                   # Authenticate first
ut tenant select           # Select which tenant to use (if you have multiple)
ut video analyze ./videos  # Tenant + plan validated automatically
```

### Video Analysis Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--output` | `-o` | Output folder for CSVs | `./Output` |
| `--dry-run` | | Validate files and show upload scope only | `false` |
| `--config` | | Config YAML override | |
| `--similarity-threshold` | | Grouping threshold (0.0-1.0) | `0.75` |
| `--confidence-threshold` | | Min classification confidence | `0.5` |
| `--timeout` | | Maximum time for each backend request | `10m` |

### Advanced Flags Explained

#### `--resume` — Continue from Checkpoint

Resumes a previously interrupted analysis after a failure, cancellation, or
process restart. Already processed image hashes are skipped.

```bash
# Day 1: Processing is interrupted
ut image analyze ./photos
#   Checkpoint saved

# Day 2: Resume where you left off
ut image analyze ./photos --resume
#   Already processed images are skipped
```

Budget enforcement and model eligibility are server-side. The CLI reports the
usage receipt returned by UteamUP; it never guesses provider pricing. A dry run
therefore reports validated upload scope, not a fabricated cost estimate.

#### `--similarity-threshold` — Grouping Sensitivity

Controls how aggressively the deduplication groups similar entities. Lower values = more aggressive merging, higher values = more separate entries.

```bash
# Default (0.75): balanced — groups obvious duplicates
ut image analyze ./photos
#   3 photos of the same pump → 1 group
#   2 similar-looking valves → 2 separate entries (not similar enough)
#   assets.csv: 45 rows

# Lower threshold (0.5): aggressive — groups loosely similar items
ut image analyze ./photos --similarity-threshold 0.5
#   3 photos of the same pump → 1 group
#   2 similar-looking valves → 1 group (merged!)
#   assets.csv: 38 rows (fewer, more merged)

# Higher threshold (0.9): conservative — only exact matches
ut image analyze ./photos --similarity-threshold 0.9
#   3 photos of the same pump → maybe 2 groups (if names differ slightly)
#   assets.csv: 52 rows (more, less merging)
```

**How similarity is calculated (6 weighted signals):**

| Signal | Weight | Match Type |
|--------|--------|------------|
| Serial number | 0.40 | Exact match |
| Model number | 0.20 | Exact match |
| Name | 0.20 | Fuzzy (Levenshtein ratio) |
| Description | 0.10 | Fuzzy (Levenshtein ratio) |
| Perceptual hash | 0.05 | Visual similarity |
| Brand | 0.05 | Exact (case-insensitive) |

Two items with matching serial + model = 0.60 score → grouped at default 0.75? No. Add fuzzy name match (0.18) = 0.78 → grouped.

#### `--confidence-threshold` — Classification Filtering

Sets the minimum AI confidence score (0.0-1.0) required to classify an entity. Below this threshold, entities are marked as "unclassified" and exported to a separate review file.

```bash
# Default (0.5): accepts most classifications
ut image analyze ./photos
#   "Industrial Pump" (confidence: 0.95) → assets.csv ✓
#   "Metal Object" (confidence: 0.45) → unclassifieds.csv (flagged for review)
#   "Pipe Fitting" (confidence: 0.72) → parts.csv ✓
#   assets.csv: 40 rows, unclassifieds.csv: 3 rows

# Higher threshold (0.8): only high-confidence results
ut image analyze ./photos --confidence-threshold 0.8
#   "Industrial Pump" (confidence: 0.95) → assets.csv ✓
#   "Metal Object" (confidence: 0.45) → unclassifieds.csv
#   "Pipe Fitting" (confidence: 0.72) → unclassifieds.csv (now below threshold!)
#   assets.csv: 30 rows, unclassifieds.csv: 13 rows

# Lower threshold (0.3): accept everything the AI suggests
ut image analyze ./photos --confidence-threshold 0.3
#   Almost nothing goes to unclassified
#   assets.csv: 43 rows, unclassifieds.csv: 0 rows
#   ⚠ May include some misclassifications
```

**Recommendation:** Start with the default (0.5), review `unclassifieds.csv`, then adjust up if you're getting too many false positives or down if too many items are flagged.

---

## Tenant Management

Manage the tenants you have access to. Required for video analysis and other tenant-scoped features.

```bash
ut tenant show             # List all tenants (name, GUID, plan, status)
ut tenant list             # Same as show (alias)
ut tenant ls               # Same as show (alias)
ut tenant select           # Interactive picker — saves to config + updates active token
```

When you have multiple tenants, `ut tenant select` presents a numbered list:

```
Select a tenant for demo@iteggs.com:

  * 1. Acme Corp [Professional]
    2. Test Org [Starter]

Select tenant (1-2): 1

Active tenant set to: Acme Corp (abc123-def456...)
```

The selected tenant is saved to your config profile (`tenantGuid`) and immediately reflected in `ut auth status`.
Use `ut tenant select` so membership is checked and the cached authenticated
tenant is updated atomically.

---

## Configuration

Config is stored at `~/.uteamup/config.json`. Supports multiple profiles for different environments.

### Config File Structure

```json
{
  "activeProfile": "production",
  "profiles": {
    "production": {
      "name": "Production",
      "apiKey": "<32-char API key>",
      "secret": "<64+ char secret>",
      "baseUrl": "https://api.uteamup.com",
      "logLevel": "INFO",
      "requestTimeout": 30000,
      "maxRetries": 3,
      "tenantGuid": "<optional tenant GUID override>"
    },
    "development": {
      "name": "Development",
      "apiKey": "<dev API key>",
      "secret": "<dev secret>",
      "baseUrl": "https://localhost:5002",
      "logLevel": "DEBUG",
      "requestTimeout": 30000,
      "maxRetries": 1,
      "tenantGuid": "<optional tenant GUID override>"
    }
  }
}
```

### Environment Variable Overrides

Environment variables override config file values (same names as the MCP server):

| Variable | Description |
|----------|-------------|
| `UTEAMUP_API_KEY` | API key (32 characters) |
| `UTEAMUP_SECRET` | API secret (64+ characters) |
| `UTEAMUP_API_BASE_URL` | API endpoint URL |
| `UTEAMUP_LOG_LEVEL` | Log level: TRACE, DEBUG, INFO, WARN, ERROR |
| `UTEAMUP_TENANT_GUID` | Override tenant GUID |

Legacy AI-provider keys are removed from the config file automatically when it
is loaded. Provider credentials belong in UteamUP web administration, not on a
CLI workstation.

### Profile Management

```bash
uteamup config init              # Create config interactively
uteamup config show              # Display current config (secrets redacted)
uteamup config set baseUrl https://localhost:5002   # Change a value
uteamup config profile development   # Switch active profile
```

---

## Authentication

You must authenticate before running any command (except `login`, `version`, `config`, `completion`).

### Method 1: Interactive Login (email/password)

```bash
uteamup login
# or
ut login
```

Prompts for email and password. Calls the backend login endpoint and receives a JWT token.

### Method 2: API Key Auth (OAuth 2.0 + PKCE)

```bash
uteamup login --api-key=KEY --api-secret=SECRET
# or
ut login --api-key=KEY --api-secret=SECRET
```

Uses the same OAuth 2.0 + PKCE flow as the MCP server. Requires an API key (32 chars) and secret (64+ chars) from the UteamUP backend.

If you omit `--api-secret`, you'll be prompted interactively.

### Token Management

- Tokens are cached at `~/.uteamup/token.json` (file permissions: 0600)
- Tokens auto-refresh before expiry
- Run `uteamup auth status` to check current auth state
- Run `uteamup logout` to clear the cached token

---

## Command Reference

### Global Flags

These flags apply to all commands:

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--output` | `-o` | Output format: `table`, `json`, `yaml` | `table` |
| `--profile` | `-P` | Config profile to use | active profile |
| `--verbose` | `-v` | Enable debug logging | `false` |
| `--insecure` | | Skip TLS verification (dev) | `false` |
| `--help` | `-h` | Show help | |

### `uteamup login`

Authenticate with UteamUP.

```
Usage: uteamup login [flags]

Flags:
  --api-key string      API key (32 characters) for OAuth 2.0 + PKCE auth
  --api-secret string   API secret (64+ characters) for OAuth 2.0 + PKCE auth
```

**Examples:**
```bash
uteamup login                              # Interactive email/password
ut login --api-key=KEY --api-secret=SECRET  # API key auth
ut login --api-key=KEY                      # API key (prompted for secret)
```

### `uteamup logout`

Clear the cached authentication token.

```
Usage: uteamup logout
```

### `uteamup auth status`

Show current authentication status — who is logged in, auth method, token expiry.

```
Usage: uteamup auth status
```

**Example output:**
```
Authentication Status
---------------------
  User:        demo@iteggs.com
  Method:      login
  Profile:     production
  Tenant:      Acme Corp
  Tenant GUID: abc123-def456-789...
  Expires:     2026-03-29 13:00:00 UTC
  Status:      Valid (6d23h remaining)
```

### `uteamup tenant` (aliases: `tenants`)

Manage tenants.

| Subcommand | Description |
|------------|-------------|
| `tenant show` | List all tenants with name, GUID, plan, and status |
| `tenant list` | Same as `show` |
| `tenant ls` | Same as `show` |
| `tenant select` | Interactive picker — saves selection to config and updates active token |

**Examples:**
```bash
ut tenant show              # List all your tenants
ut tenant select            # Pick a tenant interactively
ut tenant select            # Select through authenticated memberships
```

### `uteamup config`

Manage CLI configuration.

| Subcommand | Description |
|------------|-------------|
| `config init` | Create config file interactively |
| `config show` | Display current config (secrets redacted) |
| `config set <key> <value>` | Set a value in the active profile |
| `config profile <name>` | Switch active profile |

**Valid keys for `config set`:** `baseUrl`, `apiKey`, `secret`, `logLevel`,
`requestTimeout`, `maxRetries`, `name`, `exportJson`, `exportDir`

**Examples:**
```bash
uteamup config init
ut config show
ut config set baseUrl https://localhost:5002
ut config set logLevel DEBUG
ut config profile development
```

### `uteamup version`

Print version, commit, build date, Go version, and OS/architecture.

```
Usage: uteamup version
```

### `uteamup completion`

Generate shell completion scripts.

```
Usage: uteamup completion [bash|zsh|fish|powershell]
```

**Setup:**
```bash
# Bash
echo 'source <(uteamup completion bash)' >> ~/.bashrc

# Zsh
echo 'source <(uteamup completion zsh)' >> ~/.zshrc

# Fish
uteamup completion fish > ~/.config/fish/completions/uteamup.fish

# PowerShell
uteamup completion powershell | Out-String | Invoke-Expression
```

---

### Image Commands

#### `uteamup image` (aliases: `img`, `images`)

AI-powered image analysis for CMMS inventory.

| Action | Usage | Description |
|--------|-------|-------------|
| `analyze` | `ut image analyze <path> [flags]` | Analyze images in a folder |

| Action | Usage | Description |
|--------|-------|-------------|
| `analyze` | `ut image analyze <path> [flags]` | Analyze images in a folder |
| `status` | `ut image status` | Show checkpoint progress |

**Flags for `analyze`:**

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--output` | `-o` | Output folder for CSVs | `./Output` |
| `--dry-run` | | Validate files and show upload scope only | `false` |
| `--no-rename` | | Skip image renaming | `false` |
| `--resume` | | Resume from checkpoint | `false` |
| `--similarity-threshold` | | Grouping similarity (0.0-1.0) | `0.75` |
| `--confidence-threshold` | | Min confidence to classify (0.0-1.0) | `0.5` |
| `--config` | | Config YAML override | |
| `--timeout` | | Maximum time for each backend request | `5m` |

**Examples:**
```bash
ut image analyze ./photos
ut image analyze ./photos --dry-run
ut img analyze ./photos -o ./results
ut img analyze ./photos --resume --timeout 10m
ut image status                              # Check analysis progress
```

---

### Video Commands

#### `uteamup video` (aliases: `vid`, `videos`)

AI-powered video analysis for CMMS inventory.

| Action | Usage | Description |
|--------|-------|-------------|
| `analyze` | `ut video analyze <path> [flags]` | Analyze videos in a folder or single file |

MP4 and MOV are supported up to 100 MB per file. Use `--dry-run` to validate
files and show upload scope. GIFs are reported for processing by the image
analyzer.

---

### Tenant Commands

#### `uteamup tenant` (aliases: `tenants`)

Manage tenant selection for multi-tenant environments.

| Action | Usage | Description |
|--------|-------|-------------|
| `show` | `ut tenant show` | List all tenants with name, GUID, plan, status |
| `select` | `ut tenant select` | Interactive tenant picker |

**Examples:**
```bash
ut tenant show                               # List all tenants
ut tenant select                             # Pick tenant interactively
```

---

### Domain Commands

Domain commands follow the pattern: `uteamup <domain> <action> [args] [flags]`

#### `uteamup asset` (aliases: `assets`)

Manage assets and equipment inventory.

| Action | Usage | Description |
|--------|-------|-------------|
| `list` | `ut asset list [flags]` | List assets with pagination |
| `get` | `ut asset get <id>` | Get asset by ID |
| `create` | `ut asset create --name NAME [flags]` | Create a new asset |
| `update` | `ut asset update <id> [flags]` | Update an asset |
| `delete` | `ut asset delete <id>` | Delete an asset |
| `search` | `ut asset search <query> [flags]` | Search by name or serial |

**Flags for `list` and `search`:**

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--page` | `-p` | Page number | `1` |
| `--page-size` | `-s` | Items per page | `25` |
| `--filter` | `-f` | Filter by name | |
| `--sort-by` | | Sort field | `Name` |
| `--sort-order` | | Sort direction (`asc`/`desc`) | `asc` |

**Flags for `create`:**

| Flag | Description | Required |
|------|-------------|----------|
| `--name` | Asset name | Yes |
| `--serial` | Serial number | No |
| `--asset-type-id` | Asset type ID | No |
| `--location-id` | Location ID | No |
| `--from-json` | JSON file with asset data | No |

**Examples:**
```bash
ut asset list
ut asset list -p 2 -s 50 --sort-by CreatedAt --sort-order desc
ut asset get 42
ut asset get 42 -o json
ut asset create --name "HVAC Unit A" --serial "HV-2024-001"
ut asset create --from-json new-asset.json
ut asset search "conveyor"
ut asset delete 42
```

#### `uteamup workorder` (aliases: `wo`, `workorders`)

Manage work orders.

| Action | Usage | Description |
|--------|-------|-------------|
| `list` | `ut wo list [flags]` | List work orders |
| `get` | `ut wo get <id>` | Get work order by ID |
| `create` | `ut wo create --title TITLE [flags]` | Create a work order |
| `update` | `ut wo update <id> [flags]` | Update a work order |
| `delete` | `ut wo delete <id>` | Delete a work order |
| `search` | `ut wo search <query> [flags]` | Search work orders |

**Flags for `list`:**

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--page` | `-p` | Page number | `1` |
| `--page-size` | `-s` | Items per page | `25` |
| `--status` | | Filter by status | |
| `--priority` | | Filter by priority | |
| `--sort-by` | | Sort field | `CreatedAt` |
| `--sort-order` | | Sort direction | `desc` |

**Flags for `create`:**

| Flag | Description | Required |
|------|-------------|----------|
| `--title` | Work order title | Yes |
| `--description` | Description | No |
| `--priority` | Priority (Low, Medium, High, Critical) | No (default: Medium) |
| `--asset-id` | Associated asset ID | No |
| `--assigned-to` | Assigned user ID | No |
| `--from-json` | JSON file with data | No |

**Examples:**
```bash
ut wo list
ut wo list --status Open --priority High
ut wo get 123
ut wo get 123 -o yaml
ut wo create --title "Fix pump seal" --priority High --asset-id 42
ut wo create --from-json work-order.json
ut wo search "pump"
ut wo update 123 --status Completed
```

#### `uteamup user` (aliases: `users`)

Manage users.

| Action | Usage | Description |
|--------|-------|-------------|
| `list` | `ut user list [flags]` | List users |
| `get` | `ut user get <id>` | Get user by ID |

**Flags for `list`:**

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--page` | `-p` | Page number | `1` |
| `--page-size` | `-s` | Items per page | `25` |
| `--filter` | `-f` | Filter by name or email | |

**Examples:**
```bash
ut user list
ut user list -f "john"
ut user get abc123-def456
ut user get abc123 -o json
```

---

## Output Formats

All domain commands support three output formats via `--output` / `-o`:

**Table** (default) — human-readable columns:
```bash
ut asset list
```
```
ID    NAME              STATUS    TYPE
--    ----              ------    ----
1     HVAC Unit A       Active    Equipment
2     Conveyor Belt B   Active    Machine
```

**JSON** — structured data:
```bash
ut asset get 1 -o json
```
```json
{
  "id": 1,
  "name": "HVAC Unit A",
  "status": "Active"
}
```

**YAML** — configuration-friendly:
```bash
ut asset get 1 -o yaml
```
```yaml
id: 1
name: HVAC Unit A
status: Active
```

---

## Version Management

### Check Current Version

```bash
ut version
```

### Upgrade

```bash
brew update && brew upgrade uteamup
```

### Downgrade to a Specific Version

```bash
# 1. Roll back the Homebrew formula to the desired version
cd $(brew --repo UteamUP/tap)
git log --oneline Formula/uteamup.rb       # find the commit for the version you want
git checkout <commit-hash> -- Formula/uteamup.rb
brew reinstall uteamup
brew pin uteamup                            # prevent auto-upgrade

# To unpin later:
brew unpin uteamup
```

Or download a specific release binary directly:

```bash
# macOS Apple Silicon example (change version and arch as needed)
brew uninstall uteamup
curl -L https://github.com/UteamUP/cli/releases/download/v0.1.0/uteamup_0.1.0_darwin_arm64.tar.gz | tar xz
sudo mv uteamup /usr/local/bin/
sudo ln -sf /usr/local/bin/uteamup /usr/local/bin/ut
```

### Pin Version (Prevent Auto-upgrade)

```bash
brew pin uteamup       # stay on current version
brew unpin uteamup     # allow upgrades again
```

All releases are available at: https://github.com/UteamUP/cli/releases

---

## Development

### Prerequisites

- Go 1.22+
- [golangci-lint](https://golangci-lint.run/) (for linting)
- [GoReleaser](https://goreleaser.com/) (for releases)

### Build

```bash
make build        # Build bin/uteamup + bin/ut
make test         # Run all tests with race detection
make lint         # Run golangci-lint
make check        # fmt + vet + lint + test + build
make install      # Rebuild + install uteamup & ut to /usr/local/bin (adds PATH to .zshrc)
make uninstall    # Remove uteamup & ut from /usr/local/bin
```

### Release

```bash
make snapshot     # Build all platforms locally (no publish)
make release      # Full GoReleaser release (tags + publishes)
```

### Project Structure

```text
uteamup_cli/
├── main.go                 # Entry point
├── cmd/                    # Cobra commands (root, login, logout, auth, config, image, video, tenant, version)
├── internal/
│   ├── auth/               # OAuth 2.0 + PKCE, login, token cache
│   ├── client/             # Authenticated HTTP client, retry, SSE and bounded multipart upload
│   ├── config/             # Config loading, profiles, validation
│   ├── mediaanalyzer/      # Governed backend media contract and response validation
│   ├── registry/           # Domain registry + command builder
│   ├── output/             # Table / JSON / YAML formatters
│   ├── logging/            # Structured logging with redaction
│   ├── errors/             # Typed error hierarchy
│   ├── imageanalyzer/      # Local image preparation, grouping and export
│   │   ├── models/         # Entity types, extracted data, CSV columns
│   │   ├── config/         # Provider-neutral YAML and functional options
│   │   ├── scanner/        # Folder walk, size checks, hashing, EXIF/GPS, duplicates
│   │   ├── grouper/        # Similarity scoring, clustering, dedup
│   │   ├── exporter/       # CSV export, image renaming, summary report
│   │   ├── pipeline/       # Backend analysis orchestration
│   │   ├── imageutil/      # Image resize, normalization and validation
│   │   └── checkpoint/     # JSON checkpoint, file locking, resume
│   └── videoanalyzer/      # Local video validation, GPS, grouping and export
│       ├── config/         # Provider-neutral video config
│       ├── fileutil/       # MIME detection, bounded file scanning
│       ├── gps/            # GPS extraction from MP4/MOV metadata
│       └── pipeline/       # Governed backend analysis orchestration
├── packaging/              # Installer configs (MSI, pkg, deb, rpm, Homebrew)
└── docs/commands/          # Auto-generated command docs
```

### Adding a New Domain

1. Create `internal/registry/domains_<name>.go`
2. Register actions with tool names matching the backend MCP tools
3. Build and test

Example:
```go
package registry

func init() {
    Register(&Domain{
        Name:        "vendor",
        Aliases:     []string{"vendors"},
        Description: "Manage vendors",
        Actions: []Action{
            {
                Name:     "list",
                ToolName: "UteamupVendorList",
                Flags:    []FlagDef{{Name: "page", Short: "p", Default: 1, Type: "int"}},
            },
        },
    })
}
```

---

## License

MIT
