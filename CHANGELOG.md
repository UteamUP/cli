# Changelog

All notable changes to the UteamUP CLI will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
