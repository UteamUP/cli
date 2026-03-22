# Tasks

## Phase 1: Foundation (models, config, imageutil)

- [ ] Task 1.1: Create `internal/imageanalyzer/models/types.go` — Define `EntityType` enum (asset, tool, part, chemical, unclassified), `ClassificationResult` struct (PrimaryType, Confidence, SecondaryType, Reasoning), `ImageInfo` struct (Path, Filename, Extension, FileSizeBytes, SHA256Hash, PerceptualHash, EXIFMetadata, IsIPhoneEdit, PairedWith). Reference: Python `models.py` lines 10-33, `scanner.py` lines 21-32.

- [ ] Task 1.2: Create `internal/imageanalyzer/models/extracted.go` — Define `ExtractedAssetData`, `ExtractedToolData`, `ExtractedPartData`, `ExtractedChemicalData` structs with JSON tags matching Python field names. All fields use `*string`/`*float64` for nullable types. Reference: Python `models.py` lines 28-130.

- [ ] Task 1.3: Create `internal/imageanalyzer/models/results.go` — Define `ImageAnalysisResult` struct (ImagePath, OriginalFilename, FileHashSHA256, PerceptualHash, Classification, ExtractedData interface{}, EXIFMetadata, FlaggedForReview, ReviewReason, ProcessedAt, PairedImages, RelatedTo) and `ImageGroup` struct with `AllImagePaths()` and `AllOriginalFilenames()` methods. Reference: Python `models.py` lines 135-178.

- [ ] Task 1.4: Create `internal/imageanalyzer/models/csv_columns.go` — Define `AssetCSVColumns`, `ToolCSVColumns`, `PartCSVColumns`, `ChemicalCSVColumns`, `UnclassifiedCSVColumns` as `[]string` slices, and `CSVColumnsByType` map. Reference: Python `models.py` lines 186-233. Exact column lists must match.

- [ ] Task 1.5: Create `internal/imageanalyzer/config/config.go` — Define `AppConfig`, `GeminiConfig`, `ScanConfig`, `ProcessingConfig` structs with YAML tags. Implement `LoadConfig()` that reads YAML file, applies env var overrides (`GEMINI_API_KEY`, `GEMINI_MODEL`, `GEMINI_TEMPERATURE`, `GEMINI_REQUESTS_PER_MINUTE`, `GEMINI_MAX_OUTPUT_TOKENS`, `IMAGE_FOLDER`, `OUTPUT_FOLDER`, `RENAMED_IMAGES_FOLDER`), applies CLI flag overrides, and validates. Go dependency: `gopkg.in/yaml.v3`. Reference: Python `config.py` lines 12-153.

- [ ] Task 1.6: Create `internal/imageanalyzer/imageutil/resize.go` — Implement `ResizeImage(imageBytes []byte, maxDimension int) ([]byte, error)` that decodes any supported format, resizes if exceeding maxDimension preserving aspect ratio using Lanczos, re-encodes as JPEG quality 90. Go dependency: `golang.org/x/image/draw` for Lanczos resampling. Reference: Python `utils/image_utils.py` lines 23-46.

- [ ] Task 1.7: Create `internal/imageanalyzer/imageutil/heic.go` — Implement `ConvertHEICToJPEG(filePath string) ([]byte, error)` and `IsHEIFAvailable() bool`. Go dependency: `github.com/jdeng/goheif` (or build-tag gated). Reference: Python `utils/image_utils.py` lines 49-65.

- [ ] Task 1.8: Create `internal/imageanalyzer/imageutil/validate.go` — Implement `IsValidImage(filePath string) bool` that attempts to decode the image header; for HEIC, checks goheif availability. Reference: Python `utils/image_utils.py` lines 68-90.

- [ ] Task 1.9: Create `internal/imageanalyzer/imageutil/sanitize.go` — Implement `SanitizeFilename(name string) string` (lowercase, spaces to underscores, strip non-`[a-z0-9_-]`, collapse duplicates, trim). Reference: Python `utils/image_utils.py` lines 93-122.

- [ ] Task 1.10: Create `internal/imageanalyzer/imageutil/load.go` — Implement `LoadImageBytes(filePath string, maxDimension int) ([]byte, error)` that detects HEIC/HEIF and routes through conversion, then resizes. Reference: Python `utils/image_utils.py` lines 125-141.

## Phase 2: Infrastructure (rate limiter, checkpoint, retry)

- [ ] Task 2.1: Create `internal/imageanalyzer/ratelimiter/token_bucket.go` — Implement `TokenBucketRateLimiter` with `NewTokenBucket(requestsPerMinute int)` and `Acquire()` that blocks until a token is available. Refill rate = RPM/60 tokens/second. Uses `time.Now()` for monotonic timing and `time.Sleep()` for blocking. Reference: Python `utils/rate_limiter.py` lines 11-49.

- [ ] Task 2.2: Create `internal/imageanalyzer/retry/handler.go` — Implement `RetryHandler` with `NewRetryHandler(maxRetries int)` and `Execute(fn func() (interface{}, error)) (interface{}, error)`. Retry on network/connection/timeout errors and HTTP 429/5xx. Fail fast on 4xx (except 429). Backoff: `2^attempt + random(0, 2^attempt * 0.5)`. Include `extractStatusCode(err error) *int` helper. Reference: Python `utils/rate_limiter.py` lines 52-146.

- [ ] Task 2.3: Create `internal/imageanalyzer/checkpoint/checkpoint.go` — Implement `Checkpoint` struct with `Load(path string)`, `AcquireLock()`, `ReleaseLock()`, `IsProcessed(fileHash string) bool`, `AddResult(fileHash string, resultJSON []byte)`, `GetResults() []json.RawMessage`, `Delete()`, `GetStatus()`. Atomic save via temp file + rename. Lock file with PID check via `os.FindProcess`. Reference: Python `utils/checkpoint.py` lines 1-140.

- [ ] Task 2.4: Create `internal/imageanalyzer/checkpoint/checkpoint_test.go` — Unit tests: load/save round-trip, lock acquisition/release, stale lock cleanup, is_processed check, atomic save, delete cleanup. No external dependencies.

- [ ] Task 2.5: Create `internal/imageanalyzer/ratelimiter/token_bucket_test.go` — Unit tests: immediate acquire when tokens available, blocking when exhausted, refill timing. Use short RPM values for fast tests.

- [ ] Task 2.6: Create `internal/imageanalyzer/retry/handler_test.go` — Unit tests: immediate success, retry on transient error, fail fast on 4xx, exhausted retries, backoff timing. Use mock functions.

## Phase 3: Core Logic (scanner, analyzer, prompts)

- [ ] Task 3.1: Create `internal/imageanalyzer/scanner/scanner.go` — Implement `ImageScanner` with `NewScanner(config ScanConfig)` and `ScanFolder() ([]ImageInfo, error)`. Walk folder (recursive or not), filter by supported extensions, validate via `imageutil.IsValidImage`, compute SHA-256 (8 KiB chunked) and perceptual hash (resize to 8x8 grayscale, compare to mean), extract EXIF. Return sorted by path. Reference: Python `scanner.py` lines 40-98.

- [ ] Task 3.2: Create `internal/imageanalyzer/scanner/hashing.go` — Implement `ComputeHashes(filePath string) (sha256Hex string, perceptualHash string, err error)`. SHA-256 via `crypto/sha256` with 8 KiB streaming. Perceptual hash: decode image, resize to 8x8, convert to grayscale, compute mean pixel value, generate 64-bit hash comparing each pixel to mean, encode as hex. Go dependency: `golang.org/x/image/draw`. Reference: Python `scanner.py` lines 135-157.

- [ ] Task 3.3: Create `internal/imageanalyzer/scanner/exif.go` — Implement `ExtractEXIF(filePath string) map[string]interface{}`. Extract DateTimeOriginal (tag 36867 or 306), Make (271), Model (272), GPS IFD (0x8825). Go dependency: `github.com/rwcarlsen/goexif/exif`. Reference: Python `scanner.py` lines 100-133.

- [ ] Task 3.4: Create `internal/imageanalyzer/scanner/duplicates.go` — Implement `DetectDuplicates(images []ImageInfo) (unique []ImageInfo, duplicatePairs [][2]string)` using SHA-256 map. First-seen-by-path wins. Reference: Python `scanner.py` lines 159-191.

- [ ] Task 3.5: Create `internal/imageanalyzer/scanner/iphone_pairs.go` — Implement `DetectIPhonePairs(images []ImageInfo) map[string][]string`. Regex `^IMG_E(\d{4})` and `^IMG_(\d{4})` (case-insensitive). Set `IsIPhoneEdit` and `PairedWith` on matching ImageInfo. Reference: Python `scanner.py` lines 193-242.

- [ ] Task 3.6: Create `internal/imageanalyzer/analyzer/prompts.go` — Define `UnifiedAnalysisPrompt` and `JSONFixPrompt` as Go string constants. Exact text from Python `prompts.py`. Reference: Python `prompts.py` lines 3-187.

- [ ] Task 3.7: Create `internal/imageanalyzer/analyzer/analyzer.go` — Implement `GeminiAnalyzer` with `NewAnalyzer(config GeminiConfig)`, `AnalyzeImage(imagePath string, imageBytes []byte) ([]ImageAnalysisResult, error)`. Uses Google Generative AI Go SDK. Integrates rate limiter and retry handler. Tracks token costs. Go dependency: `github.com/google/generative-ai-go/genai`, `google.golang.org/api/option`. Reference: Python `analyzer.py` lines 44-136.

- [ ] Task 3.8: Create `internal/imageanalyzer/analyzer/parser.go` — Implement `parseMultiEntityResponse(responseText string, imagePath string) []ImageAnalysisResult`. Strip markdown fences, parse JSON, handle both `{"entities":[...]}` and legacy `{"classification":...}` formats. Map entity type to correct extracted data struct. Flag low confidence < 0.5. Reference: Python `analyzer.py` lines 149-302.

- [ ] Task 3.9: Create `internal/imageanalyzer/analyzer/cost.go` — Implement `EstimateCost(imageCount int) CostEstimate` and `TotalCost() CostEstimate`. Constants: input $0.10/1M tokens, output $0.40/1M tokens, 258 tokens/image, 1500 prompt tokens, 500 output tokens. Reference: Python `analyzer.py` lines 36-42, 338-380.

- [ ] Task 3.10: Create `internal/imageanalyzer/analyzer/json_fix.go` — Implement `attemptJSONFix(brokenText string) (map[string]interface{}, error)`. Send broken JSON to Gemini with the JSON fix prompt, parse the response. Reference: Python `analyzer.py` lines 323-336.

- [ ] Task 3.11: Create `internal/imageanalyzer/scanner/scanner_test.go` — Unit tests: scan with mixed file types, empty folder, non-existent folder, duplicate detection, iPhone pair detection. Use temp directories with test image files.

- [ ] Task 3.12: Create `internal/imageanalyzer/analyzer/parser_test.go` — Unit tests: parse multi-entity JSON, parse legacy single-entity JSON, handle markdown fences, handle broken JSON, handle unknown format, handle empty entities array. Use JSON string fixtures.

## Phase 4: Processing (grouper, exporter)

- [ ] Task 4.1: Create `internal/imageanalyzer/grouper/grouper.go` — Implement `ImageGrouper` with `NewGrouper(similarityThreshold float64)` and `GroupImages(results []ImageAnalysisResult) []ImageGroup`. Five-stage algorithm: partition by type, pre-merge iPhone pairs, group by serial number, group by name (case-insensitive), agglomerative cluster remainder. Reference: Python `grouper.py` lines 13-100.

- [ ] Task 4.2: Create `internal/imageanalyzer/grouper/similarity.go` — Implement `computeSimilarity(a, b ImageAnalysisResult) float64` with weights: serial_number=0.40, model_number=0.20, name_fuzzy=0.20, description_fuzzy=0.10, perceptual_hash=0.05, manufacturer_brand=0.05. Implement `levenshteinRatio(a, b string) float64` for fuzzy matching. Implement `phashSimilarity(hashA, hashB string) float64` via normalized Hamming distance. Go dependency: `github.com/texttheater/golang-levenshtein/levenshtein` (or inline implementation). Reference: Python `grouper.py` lines 205-300, 346-361.

- [ ] Task 4.3: Create `internal/imageanalyzer/grouper/merge.go` — Implement `selectRepresentative(group []ImageAnalysisResult) ImageAnalysisResult` (highest confidence) and `mergeExtractedData(representative *ImageAnalysisResult, members []ImageAnalysisResult)` (fill nil fields from members sorted by descending confidence). Use reflection or type-switch to iterate struct fields. Reference: Python `grouper.py` lines 171-203, 302-306.

- [ ] Task 4.4: Create `internal/imageanalyzer/exporter/csv_exporter.go` — Implement `CSVExporter` with `NewExporter(outputFolder, renamedImagesFolder string, renameImages bool, renamePattern string)` and `ExportCSVs(groups []ImageGroup, unclassified []ImageAnalysisResult) (map[string]string, error)`. Write one CSV per entity type. Use `encoding/csv` with `DictWriter`-style approach. Reference: Python `exporter.py` lines 29-87, 200-257.

- [ ] Task 4.5: Create `internal/imageanalyzer/exporter/renamer.go` — Implement `RenameImages(groups []ImageGroup) (map[string]string, error)`. Copy files with pattern `{entity_type}_{sanitized_name}_{seq:03d}_{YYYYMMDD}.{ext}`. Handle collisions by incrementing seq. Use `io.Copy` for file copying. Reference: Python `exporter.py` lines 89-133.

- [ ] Task 4.6: Create `internal/imageanalyzer/exporter/report.go` — Implement `GenerateSummaryReport(groups []ImageGroup, unclassified []ImageAnalysisResult, durationSeconds float64, duplicatesFound int) string`. Write `summary_report.md` to output folder. Reference: Python `exporter.py` lines 135-195.

- [ ] Task 4.7: Create `internal/imageanalyzer/grouper/grouper_test.go` — Unit tests: partition by type, serial number grouping, name grouping, similarity clustering, representative selection, data merging, unclassified isolation. Use fixture ImageAnalysisResult slices.

- [ ] Task 4.8: Create `internal/imageanalyzer/exporter/csv_exporter_test.go` — Unit tests: CSV column order for each entity type, chemical list field semicolon joining, unclassified rows, empty groups, output folder creation. Use temp directories.

## Phase 5: Integration (pipeline, cmd rewrite)

- [ ] Task 5.1: Create `internal/imageanalyzer/pipeline/pipeline.go` — Implement `Pipeline` with `NewPipeline(config AppConfig)` and `Run() error`. Wire all four phases: scanner -> analyzer -> grouper -> exporter. Handle dry-run, budget cap, confidence threshold, progress reporting, checkpoint integration, error handling for failed images. Reference: Python `pipeline.py` lines 15-240.

- [ ] Task 5.2: Rewrite `cmd/image.go` — Remove all Python/subprocess code (`findAnalyzerDir`, `isAnalyzerDir`, `exec.Command`). Replace with direct `pipeline.NewPipeline(config).Run()`. Keep existing flags, add new flags: `--max-cost`, `--resume`, `--similarity-threshold`, `--confidence-threshold`. Load config via `imageanalyzer/config.LoadConfig()` with CLI flag overrides. Keep profile integration for `GeminiAPIKey` and `GeminiModel`. Keep pre-run banner. Reference: current `cmd/image.go` for flag definitions and banner format.

- [ ] Task 5.3: Update `go.mod` — Add new dependencies: `gopkg.in/yaml.v3`, `github.com/google/generative-ai-go/genai`, `google.golang.org/api/option`, `github.com/rwcarlsen/goexif/exif`, `github.com/jdeng/goheif` (or alternative), `golang.org/x/image/draw`. Run `go mod tidy`.

- [ ] Task 5.4: Create `internal/imageanalyzer/pipeline/pipeline_test.go` — Integration tests: dry-run mode (no API calls), empty folder early return, config validation errors, budget cap enforcement. Use mock analyzer (interface) to avoid real API calls.

- [ ] Task 5.5: Create `internal/imageanalyzer/analyzer/analyzer_interface.go` — Define `Analyzer` interface with `AnalyzeImage(imagePath string, imageBytes []byte) ([]ImageAnalysisResult, error)`, `EstimateCost(count int) CostEstimate`, `TotalCost() CostEstimate`. Used by pipeline for testability. `GeminiAnalyzer` implements this interface.

- [ ] Task 5.6: End-to-end smoke test — Create `internal/imageanalyzer/pipeline/e2e_test.go` with a test that uses a mock analyzer, scans a temp directory with 3 test images (2 unique + 1 duplicate), runs the full pipeline, and verifies: CSV files created with correct columns, duplicate detected, summary report generated.
