# UteamUP CLI

Command-line interface for the [UteamUP](https://uteamup.com) platform. Manage assets, work orders, users, and more from any terminal. Includes AI-powered image analysis for CMMS inventory onboarding.

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
sudo dpkg -i uteamup_0.3.0_amd64.deb
```

**Fedora / RHEL / CentOS (.rpm)**:
```bash
sudo rpm -i uteamup-0.3.0.x86_64.rpm
```

**Manual (tar.gz)**:
```bash
tar xzf uteamup_0.3.0_linux_amd64.tar.gz
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

AI-powered image analysis for bulk CMMS inventory onboarding. Analyzes photos of equipment, tools, parts, and chemicals using Google Gemini Vision AI and exports structured CSV data ready for import.

### Setup

```bash
# 1. Set your Gemini API key
ut config apikey YOUR_GEMINI_API_KEY

# 2. Choose a model (optional — defaults to gemini-3.1-flash-lite-preview)
ut config model list                         # See available models
ut config model gemini-3.1-pro-preview       # Use pro for higher accuracy
```

### Usage

```bash
ut image analyze ./photos                    # Analyze all images in folder
ut image analyze ./photos --dry-run          # Estimate cost first
ut image analyze ./photos --output ./results # Custom output directory
ut img analyze ./photos --model gemini-pro-latest --verbose
ut img analyze ./photos --no-rename          # Keep original filenames
```

### What It Does

1. **Scans** folder for images (JPG, PNG, HEIC, WebP, TIFF, BMP)
2. **Detects** iPhone edit pairs and duplicates automatically
3. **Classifies** each image as asset, tool, part, or chemical using Gemini AI
4. **Extracts** CMMS fields: name, serial number, model, manufacturer, condition, etc.
5. **Detects multiple entities** per image (e.g., a machine with visible parts)
6. **Links relationships** — parts/tools/chemicals linked to parent assets via `related_to`
7. **Groups** duplicate images of the same item across photos
8. **Exports** per-entity-type CSVs (`assets.csv`, `tools.csv`, `parts.csv`, `chemicals.csv`)
9. **Renames** images with descriptive filenames (e.g., `asset_air_compressor_001_20260322.HEIC`)

### Available Models

| Model | Type | Best For |
|-------|------|----------|
| `gemini-pro-latest` | Pro | Always newest, may be unstable |
| `gemini-3.1-pro-preview` | Pro | Highest accuracy |
| `gemini-3.1-flash-lite-preview` | Flash Lite | **Default** — fastest, cheapest |
| `gemini-3-flash-preview` | Flash | Previous gen |
| `gemini-2.5-pro` | Pro | Stable |
| `gemini-2.5-flash` | Flash | Stable |

### Requirements

- [UteamUP Image Analyzer](https://github.com/UteamUP/ImageAnalyzer) Python tool installed
- Python 3.10+ with virtual environment
- Google Gemini API key ([Get one here](https://aistudio.google.com/apikey))
- Must be logged in (`uteamup login`)

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
      "geminiApiKey": "<Google Gemini API key>",
      "geminiModel": "gemini-3.1-flash-lite-preview"
    },
    "development": {
      "name": "Development",
      "apiKey": "<dev API key>",
      "secret": "<dev secret>",
      "baseUrl": "https://localhost:5002",
      "logLevel": "DEBUG",
      "requestTimeout": 30000,
      "maxRetries": 1,
      "geminiApiKey": "<Google Gemini API key>",
      "geminiModel": "gemini-3.1-pro-preview"
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
| `GEMINI_API_KEY` | Google Gemini API key for image analysis |
| `GEMINI_MODEL` | Default Gemini model name |

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
  Expires:     2026-03-29 13:00:00 UTC
  Status:      Valid (6d23h remaining)
```

### `uteamup config`

Manage CLI configuration.

| Subcommand | Description |
|------------|-------------|
| `config init` | Create config file interactively |
| `config show` | Display current config (secrets redacted) |
| `config set <key> <value>` | Set a value in the active profile |
| `config profile <name>` | Switch active profile |
| `config apikey [key]` | Get or set the Gemini API key |
| `config model [name]` | Get or set the default Gemini model |
| `config model list` | List available Gemini models |

**Valid keys for `config set`:** `baseUrl`, `apiKey`, `secret`, `logLevel`, `requestTimeout`, `maxRetries`, `name`, `geminiApiKey`, `geminiModel`, `exportJson`, `exportDir`

**Examples:**
```bash
uteamup config init
ut config show
ut config set baseUrl https://localhost:5002
ut config set logLevel DEBUG
ut config profile development
ut config apikey AIzaSy...
ut config model gemini-3.1-pro-preview
ut config model list
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

**Flags for `analyze`:**

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--output` | `-o` | Output folder for CSVs | `./Output` |
| `--model` | | Gemini model override | From config |
| `--api-key` | | Gemini API key override | From config |
| `--dry-run` | | Estimate cost only | `false` |
| `--no-rename` | | Skip image renaming | `false` |
| `--config` | | Config YAML override | |
| `--verbose` | `-V` | Enable verbose output | `false` |

**Examples:**
```bash
ut image analyze ./photos
ut image analyze ./photos --dry-run
ut img analyze ./photos --model gemini-3.1-pro-preview -o ./results
ut img analyze ./photos --no-rename --verbose
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

```
uteamup_cli/
├── main.go                 # Entry point
├── cmd/                    # Cobra commands (root, login, logout, auth, config, version, completion)
├── internal/
│   ├── auth/               # OAuth 2.0 + PKCE, login, token cache
│   ├── client/             # HTTP client, retry, SSE parser
│   ├── config/             # Config loading, profiles, validation
│   ├── registry/           # Domain registry + command builder
│   ├── output/             # Table / JSON / YAML formatters
│   ├── logging/            # Structured logging with redaction
│   └── errors/             # Typed error hierarchy
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
