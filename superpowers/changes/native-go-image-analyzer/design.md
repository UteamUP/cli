## Context

The UteamUP CLI (`uteamup`/`ut`) is a Go binary distributed via Homebrew, .deb, .rpm, and MSI. The `image analyze` command currently shells out to a Python tool requiring Python 3.10+, venv, and pip dependencies. This design replaces the Python dependency with native Go code under `internal/imageanalyzer/`.

The Python reference implementation is at `UteamUP_ImageAnalyzer/src/image_analyzer/` and provides: folder scanning, HEIC conversion, Gemini AI analysis (multi-entity), similarity-based grouping, CSV export, image renaming, checkpoint/resume, and rate limiting.

## Goals / Non-Goals

**Goals:**
- Full feature parity with Python UteamUP_ImageAnalyzer v0.2.0
- Zero external runtime dependencies (no Python, no venv)
- Same CLI flags and output format (CSVs, renamed images, summary report)
- Same Gemini prompt producing identical entity detection
- Works on macOS (arm64/amd64), Linux (amd64/arm64), Windows (amd64/arm64)

**Non-Goals:**
- New features beyond Python parity
- Backend API integration for uploading CSV results
- GUI or web interface
- Support for non-Gemini AI providers

## Decisions

### 1. Package Structure: `internal/imageanalyzer/` with subpackages

**Decision**: Create 10 subpackages mirroring the Python module structure.

```
internal/imageanalyzer/
  models/         -- Data structures, entity types, CSV columns
  config/         -- YAML config loading, validation
  imageutil/      -- Image resize, HEIC convert, sanitize filename
  ratelimiter/    -- Token bucket + exponential backoff retry
  checkpoint/     -- JSON checkpoint persistence, file locking
  scanner/        -- Folder walk, hashing, duplicate/pair detection
  analyzer/       -- Gemini API client, prompt, response parsing
  grouper/        -- Similarity scoring, clustering, dedup, field merge
  exporter/       -- CSV writing, image renaming, summary report
  pipeline/       -- 4-phase orchestration
```

**Rationale**: Maps 1:1 to Python modules for easy reference. Each package has a single responsibility and clear dependency direction (models at bottom, pipeline at top).

### 2. Gemini SDK: `github.com/google/generative-ai-go`

**Decision**: Use Google's official Go SDK for Gemini.

**Rationale**: Official SDK, actively maintained, supports `genai.ImageData()` for inline image upload. Simpler than Python — no PIL Image conversion needed, just pass raw JPEG bytes.

**Alternative considered**: Raw HTTP calls to Gemini REST API — rejected due to auth complexity and missing retry support.

### 3. ExtractedData Union Type: Wrapper struct with pointer fields

**Decision**: Use a struct with four optional pointer fields:
```go
type ExtractedData struct {
    Asset    *ExtractedAssetData
    Tool     *ExtractedToolData
    Part     *ExtractedPartData
    Chemical *ExtractedChemicalData
}
```

**Rationale**: Go lacks union types. This approach provides compile-time type safety and avoids `interface{}` type assertions. Only one pointer is non-nil at a time.

**Alternative considered**: `interface{}` with type switches — rejected as error-prone.

### 4. HEIC Decoding: `github.com/adrium/goheif`

**Decision**: Use `adrium/goheif` (maintained fork of `jdeng/goheif`), pure Go, no CGo.

**Rationale**: No C library dependency means cross-compilation works. The fork is actively maintained. Read-only is sufficient (we only decode HEIC → JPEG).

**Alternative considered**: `strukturag/libheif` (CGo, requires libheif C library) — rejected for cross-platform build complexity.

### 5. Fuzzy Matching: Levenshtein ratio calculation

**Decision**: Compute fuzzy ratio as `1.0 - float64(distance) / float64(max(len(a), len(b)))` using `github.com/agnivade/levenshtein`.

**Rationale**: Matches Python's `thefuzz.fuzz.ratio` behavior (Levenshtein-based). Simple, no large dependency.

### 6. Progress Bar: `github.com/schollz/progressbar/v3`

**Decision**: Use schollz/progressbar for terminal progress display.

**Rationale**: Widely used, supports ETA, custom descriptions, works on all platforms.

### 7. Config Loading: Merge CLI config + YAML + env vars

**Decision**: Image analyzer config comes from three sources (priority order):
1. CLI flags (`--model`, `--api-key`)
2. CLI profile config (`~/.uteamup/config.json` → `geminiApiKey`, `geminiModel`)
3. Optional `config.yaml` in project directory
4. Env vars (`GEMINI_API_KEY`, `GEMINI_MODEL`)

**Rationale**: Matches current behavior. Users who set `ut config apikey` and `ut config model` don't need any additional config files.

## Risks / Trade-offs

- **HEIC quality**: Pure Go HEIC decoder may not support all HEIC variants (e.g., HEIC with depth maps) → Mitigation: gracefully skip unsupported files, log warning
- **Gemini SDK API changes**: Go SDK is newer than Python SDK → Mitigation: pin SDK version in go.mod
- **Binary size increase**: Adding image processing and AI SDK increases binary by ~5-10MB → Acceptable trade-off for zero runtime dependencies
- **Fuzzy match parity**: Levenshtein ratio may differ slightly from Python's SequenceMatcher → Mitigation: same threshold (0.75) provides sufficient tolerance

## Migration Plan

1. Implement all packages under `internal/imageanalyzer/`
2. Rewrite `cmd/image.go` to call native pipeline
3. Remove Python shell-out code (`findAnalyzerDir`, `isAnalyzerDir`, `countImages`)
4. Test with same image set used for Python validation
5. Tag release, GoReleaser builds + updates Homebrew
6. Users get native Go via `brew upgrade uteamup` — no Python needed
7. UteamUP_ImageAnalyzer repo can be archived
