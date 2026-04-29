# Changelog

All notable changes to the UteamUP CLI will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.3.0](https://github.com/UteamUP/cli/compare/1.2.0...1.3.0) (2026-04-29)


### Features

* **attachments:** add commands for managing bug attachments (list, upload, download, delete) ([8a518fe](https://github.com/UteamUP/cli/commit/8a518fed9d3b6b68692ae7f9e28358468a75a1d8))
* **comments:** add commands for listing and adding comments on bugs ([ff7878b](https://github.com/UteamUP/cli/commit/ff7878bc00de9e2dcce6650ca3bb1cd6f6d4bda3))


### Documentation

* update release process guidelines for automation and remove manual steps ([cefd11a](https://github.com/UteamUP/cli/commit/cefd11a3d2ac57ef5da8818d959fddc8e3e98ef4))

## [1.2.0](https://github.com/UteamUP/cli/compare/1.1.0...1.2.0) (2026-04-27)


### Features

* **bugs:** add performance auto-monitoring to validated --source flag values ([57d6b9b](https://github.com/UteamUP/cli/commit/57d6b9b18d493e461e60f4e095dfe3babc80ea25))
* **industry-coding:** CLI domain for hotspot CRUD (Task 7.2) ([3d9959d](https://github.com/UteamUP/cli/commit/3d9959d01f88a8fcb3b887efdaa71e2f888e6e49))
* **registry:** add assign-asset command for code-catalog entry assignment with audit log preservation ([198e8fb](https://github.com/UteamUP/cli/commit/198e8fb75359e3c6c5a70ea476257500129aeedb))
* **registry:** add search parameter for free-text search in bugs and features ([e2a8a4f](https://github.com/UteamUP/cli/commit/e2a8a4fd53897ac5d51c19ee15249948f69108b0))
* **registry:** add source filter for bug and feature queries ([862e9b3](https://github.com/UteamUP/cli/commit/862e9b335e87c484c888d0111b02b8966fa97164))
* **registry:** add update-notes command for admin notes management and enhance REST path handling for update sub-routes ([6f61c0c](https://github.com/UteamUP/cli/commit/6f61c0c16e9ff3fcf76bc5c6d8d3a7c566fb1700))
* **registry:** add update-type command for converting submissions between Bug and Feature with audit history ([274050f](https://github.com/UteamUP/cli/commit/274050fe47dd6a4ecf3a7c854659e37667e7c0f5))


### Code Refactoring

* **registry:** remove unused helper functions ([f75a71b](https://github.com/UteamUP/cli/commit/f75a71b0241e20ba2132846fb4ee682d6784433a))

## [1.1.0](https://github.com/UteamUP/cli/compare/1.0.0...1.1.0) (2026-04-24)


### Features

* Add all MCP domains, install to /usr/local/bin, .zshrc PATH setup ([1f4958c](https://github.com/UteamUP/cli/commit/1f4958ca81f5f930703a4b03fc5bedb72ca9fdcc))
* add asset-type-meter domain with actions for managing meter definitions ([6d832c5](https://github.com/UteamUP/cli/commit/6d832c5297627e4de76fbd6918b20b78edf7b008))
* add auth, plan validation, and tenant override to video analyzer ([52e58d4](https://github.com/UteamUP/cli/commit/52e58d499df0acc51b0d6eeaa6fbc88c737d4460))
* add domain for managing subscription plans with list and get actions ([43c7e52](https://github.com/UteamUP/cli/commit/43c7e521eeace57d0d8a820ab5d85999eb116de2))
* add dual progress bars to image and video analyzers ([6788145](https://github.com/UteamUP/cli/commit/678814576e5cee335fa0f27b7d7c4de9671c613b))
* Add JSON export config option for CLI responses ([1e59cf0](https://github.com/UteamUP/cli/commit/1e59cf04ac656600884ca4950d3d5ab9ed8faeff))
* add multi-tenant selector, plan validation, and auth status tenant info ([736f562](https://github.com/UteamUP/cli/commit/736f562c3488388b199d1b6a79b0fc2d4aaf8d74))
* add new domains for condition, criticality, geofence, improvement, meter schedule, and sales booking management ([3277fa0](https://github.com/UteamUP/cli/commit/3277fa0ac873f4a03d4acd5808b5aa160a8865c4))
* add report-analytics and asset-reports domains ([1adb97f](https://github.com/UteamUP/cli/commit/1adb97f6cbf637d6c0a07beb7345e54edd859dd7))
* Add REST API support for email/password login auth ([1203c37](https://github.com/UteamUP/cli/commit/1203c3768ae94dcd3f7063e134ff2a4edac765df))
* add tenant show and tenant select commands ([876ef8f](https://github.com/UteamUP/cli/commit/876ef8f0831e4006d18d3c5f94e7b60ff6c9d82f))
* add video analyzer command for CMMS inventory extraction from MP4/MOV videos ([26b6eed](https://github.com/UteamUP/cli/commit/26b6eedf523a0ad3a7fed885c538eb4047f0dea0))
* **ai:** Add BYOK AI provider CLI domain registry ([d15e90c](https://github.com/UteamUP/cli/commit/d15e90cacea79730aec5177d151a5c52c63c0351))
* **asset:** add `ut asset get-by-guid <guid>` subcommand ([6a8426d](https://github.com/UteamUP/cli/commit/6a8426d40b2b13c1fd00cca2b4e7fedec00e4035))
* **asset:** Multi-type flags and get-specs subcommand for ut asset ([544cc79](https://github.com/UteamUP/cli/commit/544cc79fb422d86a897f36969325eb790d1a4967))
* **bank-transfer:** add CLI domain registry for bank transfer billing ([f669649](https://github.com/UteamUP/cli/commit/f66964928e44b67aaf52c5e4c8d325630a0fdcce))
* **bugs-and-features:** add delete action for global-admin to permanently remove submissions ([2333f73](https://github.com/UteamUP/cli/commit/2333f73a00d82350e50fbbf85a50f64256627b34))
* **bugsandfeatures:** add bugsandfeatures CLI domain ([fd4727a](https://github.com/UteamUP/cli/commit/fd4727ad29165fe62b7d53bc03e412c7d08b98cf))
* **cli:** add codecatalog update-by-guid / deactivate-by-guid / remove-asset-assignment ([b3d18ed](https://github.com/UteamUP/cli/commit/b3d18ed19bd041d5cda2bfecd61c4474bbee608d))
* **cli:** add config apikey and config model shortcut commands ([6fc6cb2](https://github.com/UteamUP/cli/commit/6fc6cb27c78a12481a9e589a79a8ce595cce85dc))
* **cli:** add document-import, logbook-import, document-review, ai-usage domains ([b5aa10d](https://github.com/UteamUP/cli/commit/b5aa10d129d616ba343b02dfd58f6a48498b2218))
* **cli:** add image analyze command with Gemini config integration ([3337106](https://github.com/UteamUP/cli/commit/33371066014fe9a881b1b1d06fbe3ec64cf915d1))
* **cli:** add user-ui-state domain registry ([74119d2](https://github.com/UteamUP/cli/commit/74119d23cd8afa430db69e49c71e6c11d699a8c6))
* **dedicated-instance:** Add dedicated instance integration with domain registry and client URL resolution ([b6fb5ef](https://github.com/UteamUP/cli/commit/b6fb5ef03c8b6c73ba46e596de87b6e182b78910))
* **document:** Add version and archive actions to document domain ([a6f7237](https://github.com/UteamUP/cli/commit/a6f7237b60c4ef7a82d71f1e0e6306b9c5fff07f))
* Initial UteamUP CLI project — Go CLI mirroring MCP server ([8f280cb](https://github.com/UteamUP/cli/commit/8f280cbdf560ad3fa37be6d6ce9533cd8da4fb53))
* **journal-code-linking:** Add journal and codecatalog CLI domains ([79e607b](https://github.com/UteamUP/cli/commit/79e607bf81cb8c3ed38e8d4c8531dfa88b46fa33))
* **journal:** add import, create-from-image, and mention search domains ([4e39d05](https://github.com/UteamUP/cli/commit/4e39d057d4bae556469053288946ff3f7a1e955f))
* **meter-reading:** add CLI domain for GUID-based meter-reading commands ([c0ac999](https://github.com/UteamUP/cli/commit/c0ac999453794bbfdc87f9cecc08a29260624a60))
* native Go image analyzer — remove Python dependency entirely ([0d539bd](https://github.com/UteamUP/cli/commit/0d539bdcfc2c80c98120f7f6e3b4100e82029d2a))
* **output:** enhance `bugs get` command to display full status history in a dedicated block ([981c52f](https://github.com/UteamUP/cli/commit/981c52fb0f6e7b925e6ecddb8508066ab2d8b6af))
* **project:** add my-projects subcommand ([ed79e47](https://github.com/UteamUP/cli/commit/ed79e47d176c019470f8709d6edb33b0080d9a02))
* **registry:** add support for update-status action and GUID-based identifier handling ([fb83a56](https://github.com/UteamUP/cli/commit/fb83a560b2517d9a321a4c95b99376523250bd6f))
* **registry:** implement admin-billing-gateway commands for tenant billing management ([1a60625](https://github.com/UteamUP/cli/commit/1a606252f5f3f4c270a2713df478b7c7e87ad207))
* **tenant:** add invite-defaults get/set commands ([2d1d0e8](https://github.com/UteamUP/cli/commit/2d1d0e829ab76542e96b5b1ace5971f7fbf2032c))
* **ux-simplification:** Add quick-report CLI domain ([69a5407](https://github.com/UteamUP/cli/commit/69a5407f35fb5b380f10a6d2baf4d1ab2f99bfc3))
* v0.3.0 — add CLIGuidelines.md and update changelog ([0decf59](https://github.com/UteamUP/cli/commit/0decf591f35e54bbf94545be7d3448c34d78502f))
* vendor/location enrichment with GPS geocoding and online lookup ([0714680](https://github.com/UteamUP/cli/commit/0714680b5668e03e0e16af871d0c4d1c128981e3))
* **workorder:** add `ut workorder quick-close` action with required and optional flags ([3ca51ab](https://github.com/UteamUP/cli/commit/3ca51abdeb85d9477623b3e40ee616a922a09a67))
* **workorder:** add quick-close action with required and optional flags ([490633e](https://github.com/UteamUP/cli/commit/490633ee815d92c6b893547a0af716cafe301e35))


### Bug Fixes

* Add tenant headers (X-Tenant-ID, X-Tenant-Guid) to all API requests ([3256adc](https://github.com/UteamUP/cli/commit/3256adc0799cd4f145571931eb8f119132f48b95))
* **ci:** correct release-please-action SHA to v4.4.0 ([520ff24](https://github.com/UteamUP/cli/commit/520ff24a684607706e746f803b2a1c83c1fed0d0))
* **cli:** honor UTEAMUP_API_BASE_URL when no config file exists ([86a9f3e](https://github.com/UteamUP/cli/commit/86a9f3ec47b6486bfd84342ad97c6a82d9830f14))
* **cli:** print image analyzer status banner to stdout for clean display ([d7b332b](https://github.com/UteamUP/cli/commit/d7b332b502a9e46a9b4d8ab3399e89b0d1b68fe1))
* **cli:** show errors instead of silent exit, expand analyzer search paths ([67ef32d](https://github.com/UteamUP/cli/commit/67ef32dd74f0feb2a669adf1b1b59aa4eae577e2))
* correct Gemini MIME type — ImageData() prepends 'image/' automatically ([fd24466](https://github.com/UteamUP/cli/commit/fd244664845d4e9764917c18c2e19b4d35a2d6a3))
* move checkpoint to ~/.uteamup/, add superpowers to .gitignore ([c52f4b9](https://github.com/UteamUP/cli/commit/c52f4b90511062a7291d2c04278f4065eca880dc))
* resolve panic on float flag with int default in domain registry ([10405e1](https://github.com/UteamUP/cli/commit/10405e1bbb80933c1ba7dc1221bc7ea1d2bac92d))


### Miscellaneous

* add firebase-debug.log to gitignore ([71d78fd](https://github.com/UteamUP/cli/commit/71d78fd9a2f0ad3eb02ea8ed5d5462e3efe969c1))
* Add MIT LICENSE file ([1f52848](https://github.com/UteamUP/cli/commit/1f5284866163c26098fc178a87add4becdcc0615))
* update go.mod/go.sum, ignore Images/ directory ([5f92665](https://github.com/UteamUP/cli/commit/5f9266547f33d68ea002de9286be98a490386453))
* Update GoReleaser GitHub owner to UteamUP ([cc9f490](https://github.com/UteamUP/cli/commit/cc9f4907ddd24dd8c1cebcb45f831a36aabea14d))


### Documentation

* add detailed flag documentation with before/after examples ([6bda188](https://github.com/UteamUP/cli/commit/6bda1886de865efff662e50645e370ff50cbb6d6))
* Add version management section (upgrade, downgrade, pin) ([204fb08](https://github.com/UteamUP/cli/commit/204fb083912cd6aac40de627f954920b4bfbc8dd))
* comprehensive release process documentation in CLIGuidelines.md ([f23c0f7](https://github.com/UteamUP/cli/commit/f23c0f78aa498973585bf21f4ef8aaad42ef00f4))
* **guidelines:** document REST-routing and CSRF/auth rules for domain commands ([8902781](https://github.com/UteamUP/cli/commit/8902781304c3e19c78d0302b0160c4eaf207860d))
* update changelog for v0.10.0 ([2d2d251](https://github.com/UteamUP/cli/commit/2d2d251e47cdeb7fd30d0cfa9e8786d35cebb61f))
* update changelog version to v0.6.0 ([808438a](https://github.com/UteamUP/cli/commit/808438a162140c89637df65b5d95a743c707c557))
* update changelog with image analyzer CLI features ([0a33c31](https://github.com/UteamUP/cli/commit/0a33c3160d888a75e015dadacdf882e35c930795))
* update README with image analysis, Gemini config, and v0.3.0 references ([bc2adc6](https://github.com/UteamUP/cli/commit/bc2adc6a7428d1b3c4e7761db437f17543b8e264))
* update README with video analysis, tenant management, vendor/location enrichment ([bb19e8f](https://github.com/UteamUP/cli/commit/bb19e8f724e66eab045c5e670436276061a3aad6))
* **workorder:** document canonical priority tiers on CLI flags ([8479de5](https://github.com/UteamUP/cli/commit/8479de535c8bb861e5ff092ba1e429c2de754a8e))


### Tests

* **project:** add domains_project_test.go and verify CLI pipeline ([2b1695c](https://github.com/UteamUP/cli/commit/2b1695c584bd1812097569779674a672cdb03cfa))


### CI/CD

* add CI ([db51ec9](https://github.com/UteamUP/cli/commit/db51ec95244359ab714641f1fc8e075195acfc75))
* add CODEOWNERS ([76d395d](https://github.com/UteamUP/cli/commit/76d395d73f5e268dad50befedf73cb7b51eb7ac7))
* add Dependabot ([0c8dbaa](https://github.com/UteamUP/cli/commit/0c8dbaaa5b20e9f932e4fe6a5a8d4a1b59b5f88c))
* add release ([90772f2](https://github.com/UteamUP/cli/commit/90772f277673fdbff4b6994b60bb43f5c9370de8))
* add Release Please automated versioning ([06f7362](https://github.com/UteamUP/cli/commit/06f736281a7103163b6ecd8c6be04d88eefe30ac))

## [Unreleased]

### Added
- **Performance Auto-Monitoring (CLI).** Added `performance-auto` to the validated `--source` flag values for the `uteamup bugs list` command in `domains_bugs.go`. Added corresponding test coverage.

### Changed
- **`uteamup bugs get <externalGuid>` default human output now surfaces the full `statusHistory[]` as a `History:` block.** `-o json` and `-o yaml` were already pass-through and already carried the array; the gap was the default table renderer, which flattened the nested array into a 60-char-truncated JSON blob. `internal/output/table.go::printObjectTable` now skips the `statusHistory` key from the key/value section and, after the main object, prints a chronological `History:` block — one line per entry with ISO-8601 timestamp, `from -> to` transition, author (`changedByUserEmail` falling back to `changedByUserId` so the `system:auto-ingest` sentinel remains visible), and the note truncated to fit the terminal width (`$COLUMNS` with a 160-col fallback). `bugs list` output is intentionally unchanged — per-row history would turn a one-screen list into screens of noise. Six new tests in `internal/output/table_test.go` cover: multi-entry chronological ordering, `system:auto-ingest` author visibility, long `[auto-reopen]` note truncation with `...`, empty history renders `History: (none)` rather than crashing, list output does NOT expand history per row, `fromStatus -> toStatus` arrow marker. `go vet ./... && go test ./... -race && make build` all clean.

## [0.10.0] — 2026-04-22

### Fixed
- **`cmd/root.go` + `cmd/login.go` — honor `UTEAMUP_API_BASE_URL` when no `~/.uteamup/config.json` exists.** Both entry points previously applied the env var override only *inside* `config.Load()` (which returns an error when there is no config file), then silently fell back to the hardcoded production URL `https://api.uteamup.com`. The uteamup-debug Claude skill, CI, and anyone else following the README's "set `UTEAMUP_API_BASE_URL`" instructions were unknowingly hitting prod. `runLogin` and `registerDomainCommands` now consult the env var before the hardcoded fallback. An active profile's `BaseURL` still wins when a config file is present, so existing workflows are unchanged.
- **`internal/client/client.go::CallREST` — unconditionally set `X-Requested-With: XMLHttpRequest`.** Mutating endpoints (POST/PUT/PATCH/DELETE) on routes like `/api/bugsandfeatures` reject requests without the marker with HTTP 400 `"Missing required X-Requested-With header."`. The frontend `apiCall()` has always sent it; the CLI did not, which made `uteamup bugs update-status … Fixed` fail on its PATCH step.
- **`internal/registry/registry.go::buildRESTPath` — route `update-status` and GUID-keyed actions correctly.** The `HTTPMethod` map gained `"update-status": "PATCH"`, and `buildRESTPath` now accepts `args["externalGuid"]` as the positional identifier in addition to `args["id"]`, so GUID-first domains get `PATCH /api/<domain>/{guid}/status` and `GET /api/<domain>/{guid}` instead of the list endpoint as a fallback. Unblocks `uteamup bugs update-status <externalGuid> Fixed --resolution-reference <sha>`.

### Docs
- **`CLIGuidelines.md`** — added three subsections under "Architecture Quick Reference" documenting the above changes: (1) REST routing table (action → HTTP method → URL pattern), (2) CSRF header requirement on mutating calls, and (3) the `LocalOrAzureAdPolicy` pattern that new CLI-facing backend controllers should prefer over the legacy stacked `[Authorize(Policy="AzureAdPolicy"), Authorize(Policy="LocalPolicy")]` (a third scheme like Google listed in a policy triggers a Google JWT Bearer challenge that short-circuits to 401 even when Local validates).

### Added
- **`tenant` domain.** New `internal/registry/domains_tenant.go` registers two actions mirroring the new backend MCP tools: `invite-defaults-get <tenantGuid>` and `invite-defaults-set <tenantGuid>` with flags `--auto-license`, `--license-type 0|1` (0=Regular, 1=Helpdesk), `--auto-role`, `--role-id <guid>`. These let operators configure a tenant so every new invite automatically assigns a license + role. `go vet` clean, registry tests pass under `-race`, `make build` produces `bin/uteamup` and `bin/ut`.

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
