# Pipeline Orchestration Specification

Replaces Python `pipeline.py`, `utils/checkpoint.py`, `utils/rate_limiter.py`, and `config.py` with native Go implementation.

## ADDED Requirements

### Requirement: Four-phase pipeline execution

Orchestrate the full pipeline: Scan -> Analyze -> Group -> Export.

#### Scenario: Normal full run
- **WHEN** `Pipeline.Run()` is called with a valid config
- **THEN** Phase 1 (scan) discovers images and detects duplicates/iPhone pairs, Phase 2 (analyze) sends images to Gemini, Phase 3 (group) clusters results, Phase 4 (export) writes CSVs and renames images

#### Scenario: No images found
- **WHEN** the scan phase returns zero images
- **THEN** a warning is logged and the pipeline returns early (no error)

### Requirement: Duplicate and iPhone pair filtering before analysis

Skip duplicates and iPhone edit variants from analysis to reduce API calls.

#### Scenario: Duplicates detected
- **WHEN** 5 duplicate images are found from 50 total
- **THEN** only 45 unique images proceed to analysis

#### Scenario: iPhone edit pairs detected
- **WHEN** `IMG_E1234.jpg` is paired with `IMG_1234.jpg`
- **THEN** only the original (`IMG_1234.jpg`) is analyzed; the edit variant path is attached to the result's `PairedImages` list after analysis

### Requirement: Dry-run mode

Estimate cost without making any Gemini API calls.

#### Scenario: Dry run
- **WHEN** `processing.dry_run` is true
- **THEN** the pipeline prints: images to analyze, duplicates skipped, iPhone edit pairs, model name, estimated input/output tokens, estimated total cost in USD, estimated time (images / requests_per_minute), and budget cap if set; then returns without calling Gemini

### Requirement: Budget cap enforcement

Stop processing when estimated cost exceeds the configured `max_cost`.

#### Scenario: Budget limit reached
- **WHEN** cumulative spent cost plus the cost of one more image exceeds `max_cost`
- **THEN** processing stops with a warning log: "Budget limit reached, spent=$X.XXXX, limit=$Y.YY"

#### Scenario: No budget cap
- **WHEN** `max_cost` is nil/zero
- **THEN** all images are processed regardless of cost

### Requirement: Confidence threshold enforcement

Reclassify low-confidence results as unclassified.

#### Scenario: Result below confidence threshold
- **WHEN** an analysis result has confidence < `confidence_threshold` (default 0.5)
- **THEN** its `primary_type` is changed to `unclassified`, `FlaggedForReview` is set to true, and `ReviewReason` is `"Low confidence: 0.XX"`

### Requirement: Progress reporting

Display progress during the analysis phase.

#### Scenario: Processing images
- **WHEN** images are being analyzed
- **THEN** a progress indicator shows current image filename (truncated to 30 chars), count, and unit ("img")

### Requirement: Checkpoint persistence for resume capability

Persist processing state to a JSON checkpoint file for resume-after-interruption.

#### Scenario: Checkpoint save after each image
- **WHEN** an image is successfully analyzed
- **THEN** its SHA-256 hash and result JSON are written atomically to the checkpoint file (write to temp file, then rename)

#### Scenario: Resume from checkpoint
- **WHEN** a checkpoint file exists from a previous interrupted run
- **THEN** previously processed results are loaded, and images whose SHA-256 hash is already in the checkpoint are skipped

#### Scenario: Checkpoint cleanup on success
- **WHEN** the pipeline completes successfully
- **THEN** both the checkpoint file and its lock file are deleted

#### Scenario: Corrupted checkpoint entry
- **WHEN** a checkpoint entry cannot be deserialized
- **THEN** the corrupted entry is skipped silently

### Requirement: Checkpoint lock file

Prevent concurrent pipeline runs on the same checkpoint.

#### Scenario: Lock file from another running process
- **WHEN** a lock file exists and the PID in it is still running
- **THEN** the pipeline returns an error: "Another process (PID X) is using this checkpoint"

#### Scenario: Stale lock file
- **WHEN** a lock file exists but the PID is not running (stale)
- **THEN** the stale lock is removed with a warning log, and the pipeline proceeds

#### Scenario: Corrupt lock file
- **WHEN** the lock file contains invalid JSON
- **THEN** the corrupt lock is removed with a warning log, and the pipeline proceeds

#### Scenario: Lock release
- **WHEN** analysis phase completes (success or error)
- **THEN** the lock file is released via `defer`

### Requirement: Token bucket rate limiter

Control Gemini API request rate using a token bucket algorithm.

#### Scenario: Requests within limit
- **WHEN** requests are made at or below `requests_per_minute`
- **THEN** tokens are consumed immediately with no delay

#### Scenario: Burst exceeds limit
- **WHEN** requests exceed the bucket capacity
- **THEN** `Acquire()` blocks until tokens refill (refill rate = RPM / 60 tokens per second)

### Requirement: Exponential backoff retry handler

Retry transient errors with exponential backoff and jitter.

#### Scenario: Transient network/connection/timeout error
- **WHEN** a network error, connection error, or timeout occurs
- **THEN** the request is retried up to `max_retries` times with delay = 2^attempt + random(0, 2^attempt * 0.5)

#### Scenario: Transient HTTP status (429, 500, 502, 503, 504)
- **WHEN** the API returns one of these status codes
- **THEN** the request is retried with backoff

#### Scenario: Non-transient HTTP error (400, 401, 403)
- **WHEN** the API returns a 4xx error other than 429
- **THEN** the error is returned immediately (no retry)

#### Scenario: All retries exhausted
- **WHEN** `max_retries` attempts have failed
- **THEN** the last error is returned

### Requirement: YAML + environment variable configuration

Load configuration from YAML file with environment variable overrides and CLI flag overrides.

#### Scenario: YAML config file exists
- **WHEN** `config.yaml` exists with `gemini`, `scan`, and `processing` sections
- **THEN** values are loaded from YAML

#### Scenario: Environment variable override
- **WHEN** `GEMINI_API_KEY`, `GEMINI_MODEL`, `GEMINI_TEMPERATURE`, `GEMINI_REQUESTS_PER_MINUTE`, `GEMINI_MAX_OUTPUT_TOKENS`, `IMAGE_FOLDER`, `OUTPUT_FOLDER`, or `RENAMED_IMAGES_FOLDER` is set
- **THEN** the environment variable value overrides the YAML value

#### Scenario: CLI flag override
- **WHEN** `--folder`, `--output`, `--dry-run`, `--no-rename`, or `--max-cost` flags are provided
- **THEN** the flag value overrides both YAML and environment variable values

#### Scenario: Config validation
- **WHEN** config is loaded
- **THEN** validation checks: API key required (unless dry-run), image folder must exist, temperature must be 0-2, requests_per_minute must be 1-1000

### Requirement: Default configuration values

Provide sensible defaults when no config file or env vars are present.

#### Scenario: No config file
- **WHEN** no `config.yaml` exists and no env vars are set
- **THEN** defaults are: model=`gemini-3.1-flash`, temperature=0.1, max_output_tokens=4096, RPM=15, max_retries=3, timeout=60s, image_folder=`./Images/Original`, output_folder=`./Output`, renamed_images_folder=`./Images/Updated`, recursive=true, supported_formats=[.jpg,.jpeg,.png,.webp,.heic,.heif,.tiff,.bmp], max_image_dimension=2048, max_file_size_mb=20, confidence_threshold=0.5, grouping_similarity_threshold=0.75, rename_pattern=`{entity_type}_{name}_{seq}_{date}.{ext}`

### Requirement: Analysis failure handling

Create unclassified results for images that fail during analysis.

#### Scenario: Image analysis throws an exception
- **WHEN** an image fails to load, resize, or analyze
- **THEN** an `ImageAnalysisResult` with `primary_type=unclassified`, confidence=0.0, `FlaggedForReview=true`, and `ReviewReason="Analysis error: <message>"` is created and checkpointed

### Requirement: Classified vs unclassified separation before grouping

Separate classified and unclassified results before the grouping phase.

#### Scenario: Mixed results
- **WHEN** analysis produces 40 classified and 10 unclassified results
- **THEN** only the 40 classified results are passed to the grouper; the 10 unclassified are passed directly to the exporter
