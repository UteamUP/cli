# Changelog

All notable changes to the UteamUP CLI will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **`ut workorder quick-close`** — new action on the existing `workorder` domain that mirrors the backend `UteamupWorkorderQuickClose` MCP tool (atomic create + close from a pre-approved template). Required flags: `--template <guid>`, `--asset <guid>`, `--note <text>`. Optional flags: `--idempotency-key <guid>` (CLI generates one per invocation when omitted), `--industry-code <guid>` (informational), `--performed-at <ISO-8601>` (clamped to ±0/−30 days server-side). Falls under the stricter automation rate-limit tier (5/min, 50/day). Test file `internal/registry/domains_workorder_test.go` verifies the action is registered with `ToolName: UteamupWorkorderQuickClose`, that the three required flags are marked `Required: true`, that the three optional flags stay optional, and that the action takes zero positional args (all identifiers are GUIDs that would be painful to position-order). 3/3 passing under `go test -race`.
- **`bugsandfeatures` domain.** New `internal/registry/domains_bugsandfeatures.go` registers four actions that mirror the new MCP tools: `list` (global-admin; filters by type/status/severity/tenant/submitter, pagination, default hides Rejected/Confirmed), `get <externalGuid>` (global-admin), `create` (any authenticated user; requires `--title`, `--description`, `--idempotency-key`), and `update-status <externalGuid> <toStatus>` (global-admin; `--note` required on reject/reopen, `--resolution-reference` required on Fixed — enforced server-side). Aliases: `bugs`, `features`, `baf`.

### Changed
- `ut workorder list` / `create` / `update` priority flag descriptions now enumerate the canonical tiers `1=Low, 2=Medium, 3=High, 4=Urgent, 5=Critical` to match the backend `WorkorderPriority` enum and the new Critical dropdown option in the frontend. Registry metadata only — no behavioral change; `go vet` clean.

### Added
- `ut asset get-by-guid <guid>` (also `uteamup asset get-by-guid ...`) — new subcommand mirroring the backend `uteamup_asset_get_by_guid` MCP tool / `GET /api/asset/by-guid/{guid}`. Fetches an asset using its stable `ExternalGuid` (survives migrations and reseeds) rather than the integer id, making CLI invocations safe to copy between environments. Registered in `internal/registry/domains_asset.go` with ToolName `UteamupAssetGetByGuid`, and `domains_asset` test expects the new `get-by-guid` action (alongside `list`, `get`, `create`, `update`, `delete`, `search`).
- Four new domain registries mirroring the new image/document import MCP tools: `domains_document_import.go` (`document-import get`), `domains_logbook_import.go` (`logbook-import get`), `domains_document_review.go` (`document-review queue`, `document-review acknowledge`), `domains_ai_usage.go` (`ai-usage summary`). Read-only + acknowledge surface only; multipart upload and batch commit stay HTTP-only by design.
- `ut project my-projects` (also `uteamup project my-projects`) — new subcommand mirroring the backend `GET /api/project/my-projects` endpoint. Lists projects that contain workorders assigned to the current user (primary or secondary). Registered in `internal/registry/domains_project.go` with ToolName `UteamupProjectMyProjects`; the backend MediatR handler was added in the same PR (MCP `UteamupProjectMyProjects` tool).
- `internal/registry/domains_project_test.go` — unit tests covering the project domain registration, `projects` alias, `search` / `my-projects` action/ToolName mapping, and a regression guard asserting `my-projects` takes zero args and zero flags (user identity must come from the API key). Follows the existing `domains_journal_test.go` pattern.

### Tests
- Verified `go fmt`, `go vet`, `go test ./... -race`, and `make build` all pass. Note: `make lint` (golangci-lint) continues to fail on pre-existing issues in `internal/imageanalyzer/analyzer/json_fix.go` and `internal/videoanalyzer/gps/mp4meta.go` (commits `0d539bd` / `26b6eed`, March 2026) — unrelated to this change.

## [0.7.1] — 2026-03-28

### Fixed
- Fixed panic when running geofence commands: `interface conversion: interface {} is int, not float64`
- Made float flag default handling in domain registry defensive against int/float64 type mismatch

## [0.7.0] — 2026-03-27

### Added
- `report-analytics` domain (alias: `report-stats`) — view aggregated report analytics with cost trends, top assets by maintenance cost, and completion metrics
- `asset-reports` domain — view paginated reports for a specific asset with summary statistics
- Enriched `report` domain description with cost breakdown, checklists, meter readings, labour, and tool usage details

## [0.6.3] — 2026-03-23

### Added
- Dual progress bars for image and video analysis: per-item steps (load/upload/analyze/save) + overall 0%-100% progress
- File size display in video analysis per-video headers
- Per-item entity count summary after each image/video is analyzed

## [0.6.2] — 2026-03-23

### Added
- `ut tenant show` command (aliases: `list`, `ls`) — lists all tenants with name, GUID, plan, and status
- `ut tenant select` command — interactive tenant picker that saves selection to config and updates active token
- Tenant selection updates both config profile (`tenantGuid`) and cached token so `ut auth status` reflects the change immediately

## [0.6.1] — 2026-03-23

### Added
- Video analyzer requires UteamUP authentication (login) and active tenant subscription plan
- Interactive multi-tenant selector when user has access to multiple tenants and no `tenantGuid` is configured
- `tenantGuid` field in CLI profile config (`~/.uteamup/config.json`) for tenant override
- `UTEAMUP_TENANT_GUID` environment variable override
- Tenant mismatch detection: re-authentication required when config tenant differs from logged-in tenant
- Plan name and tenant name displayed in video analyzer banner

## [0.6.0] — 2026-03-23

### Added
- `uteamup video analyze <path>` command for AI-powered CMMS video analysis
- Video file validation via magic byte MIME detection (MP4, MOV supported; GIF routed to image analyzer)
- Gemini File API integration with async upload, processing poll with terminal spinner, and automatic cleanup
- Video-specific CMMS entity extraction prompt with timestamp detection (MM:SS format)
- GPS coordinate extraction from MP4/MOV container metadata (©xyz and ISO 6709 atoms)
- Vendor enrichment via follow-up Gemini lookup (website, full name, business category)
- Temporal deduplication to merge same-entity detections across video frames
- Cross-video deduplication using existing grouping algorithm
- Consistent CSV output (assets, tools, parts, chemicals, vendors, locations) matching image analyzer format
- Dry-run mode for video cost estimation
- Video Analysis section in CLIGuidelines.md

## [0.3.0] — 2026-03-22

### Added
- `uteamup image analyze <path>` command for AI-powered CMMS image analysis
- Gemini AI configuration in CLI profiles (`geminiApiKey`, `geminiModel`)
- `ut config apikey [key]` shortcut to get/set Gemini API key
- `ut config model [name]` shortcut to get/set default Gemini model
- `ut config model list` to display all available Gemini models
- Pre-processing status banner showing image count, model, and output path
- Config init prompts for Gemini settings with model selection
- Support for `=` syntax in config commands (`ut config apikey=xyz`)
- Image analyze requires authentication (login required)
- CLIGuidelines.md with full release, packaging, and Homebrew documentation

## [0.1.0] — 2026-03-22

### Added
- Initial project scaffold with Cobra CLI framework
- Dual authentication: interactive login (email/password) and API key auth (OAuth 2.0 + PKCE)
- `ut` shortname alias for `uteamup` binary
- JSON config file with multi-profile support (~/.uteamup/config.json)
- Domain registry pattern for declarative command definitions
- Starter domains: Asset, WorkOrder, User
- HTTP client with exponential backoff retry and SSE parsing
- Output formatters: table (default), JSON, YAML
- Auth gate requiring login before any command
- Cross-platform installers: MSI (Windows), .pkg + Homebrew (macOS), .deb + .rpm (Linux)
- Shell completions for bash, zsh, fish, powershell
- Structured logging with sensitive data redaction
