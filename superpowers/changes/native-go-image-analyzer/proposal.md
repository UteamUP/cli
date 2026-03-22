## Why

The CLI currently shells out to a Python tool (UteamUP_ImageAnalyzer) for image analysis, requiring Python 3.10+, a virtual environment, and pip dependencies. This creates installation friction, version mismatches, and makes the CLI non-self-contained. Rewriting as native Go eliminates the Python dependency and makes `uteamup image analyze` work out of the box after `brew install uteamup`.

## What Changes

- **BREAKING**: Remove Python shell-out from `cmd/image.go` — the `UTEAMUP_IMAGE_ANALYZER_PATH` env var and Python venv detection are removed
- Add native Go image analysis pipeline under `internal/imageanalyzer/`
- Add Go dependencies: Gemini AI SDK, image hashing, HEIC decoding, progress bar, fuzzy matching
- Rewrite `cmd/image.go` to call native Go pipeline directly
- All existing CLI flags (`--model`, `--api-key`, `--dry-run`, `--no-rename`, `--output`, `--verbose`, `--config`) remain unchanged
- All output formats (CSV per entity type, image renaming, summary report) remain identical

## Capabilities

### New Capabilities
- `image-scanning`: Folder walking, HEIC/HEIF conversion, SHA-256 + perceptual hashing, duplicate detection, iPhone edit pair detection
- `ai-image-analysis`: Gemini Vision API integration, multi-entity detection per image, relationship linking, JSON response parsing with retry
- `image-grouping`: Cross-image deduplication via exact name match, serial number match, weighted similarity scoring, agglomerative clustering, field merging
- `csv-export`: Per-entity-type CSV export (assets, tools, parts, chemicals), image renaming with descriptive filenames, Markdown summary report
- `pipeline-orchestration`: 4-phase pipeline (scan → analyze → group → export), checkpoint/resume, rate limiting, cost estimation, progress display

### Modified Capabilities
- `cli-image-command`: Rewrite to call native Go pipeline instead of Python subprocess

## Scope

- **In scope**: Full feature parity with Python UteamUP_ImageAnalyzer v0.2.0
- **Out of scope**: New features beyond what Python already does; backend API integration for uploading results

## Risks

- HEIC decoding in pure Go may have limited format support compared to Python's pillow-heif
- Gemini Go SDK API differences may require prompt adjustments
- Fuzzy string matching quality may differ slightly from Python's thefuzz library
