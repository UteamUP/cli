# Changelog

All notable changes to the UteamUP CLI will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
