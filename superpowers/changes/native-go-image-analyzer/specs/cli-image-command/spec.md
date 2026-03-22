# CLI Image Command Specification

Rewrite `cmd/image.go` to use the native Go pipeline instead of shelling out to Python.

## ADDED Requirements

### Requirement: Remove Python dependency

Replace the `exec.Command` Python subprocess call with direct invocation of the native Go pipeline.

#### Scenario: Running image analyze
- **WHEN** `uteamup image analyze <path>` is executed
- **THEN** the native Go pipeline is invoked directly (no Python, no venv, no `findAnalyzerDir()`)

### Requirement: Preserve existing CLI flags

All current flags must be preserved with identical names and behavior.

#### Scenario: Flag compatibility
- **WHEN** the user runs `uteamup image analyze ./photos --output ./results --dry-run --model gemini-2.5-pro --api-key AIza... --no-rename --config config.yaml --verbose`
- **THEN** all flags are parsed and passed to the native Go pipeline config

### Requirement: Flag definitions

The following flags must be supported on the `image analyze` subcommand:

#### Scenario: --output / -o flag
- **WHEN** `--output ./results` is provided
- **THEN** the pipeline's `output_folder` is set to the absolute path of `./results` (default: `./Output`)

#### Scenario: --model flag
- **WHEN** `--model gemini-2.5-pro` is provided
- **THEN** the pipeline's `gemini.model` is set to `gemini-2.5-pro` (default: empty, falls back to config/env/`gemini-3.1-flash`)

#### Scenario: --api-key flag
- **WHEN** `--api-key AIza...` is provided
- **THEN** the pipeline's `gemini.api_key` is set (overrides `GEMINI_API_KEY` env var and config file)

#### Scenario: --dry-run flag
- **WHEN** `--dry-run` is provided
- **THEN** the pipeline runs in dry-run mode (cost estimate only, no API calls)

#### Scenario: --no-rename flag
- **WHEN** `--no-rename` is provided
- **THEN** the pipeline skips image renaming after analysis

#### Scenario: --config flag
- **WHEN** `--config custom.yaml` is provided
- **THEN** the pipeline loads configuration from `custom.yaml` instead of the default `config.yaml`

#### Scenario: --verbose / -V flag
- **WHEN** `--verbose` is provided
- **THEN** debug-level logging is enabled

### Requirement: Config profile integration

Load Gemini API key and model from the CLI config profile as defaults.

#### Scenario: Profile has GeminiAPIKey
- **WHEN** the active CLI profile has `GeminiAPIKey` set and no `--api-key` flag is provided
- **THEN** the profile's `GeminiAPIKey` is used as the API key

#### Scenario: Profile has GeminiModel
- **WHEN** the active CLI profile has `GeminiModel` set and no `--model` flag is provided
- **THEN** the profile's `GeminiModel` is used as the model name

#### Scenario: Flag overrides profile
- **WHEN** both the profile and the `--api-key` flag provide a value
- **THEN** the flag value takes precedence

### Requirement: Path validation

Validate that the provided image path exists and is a directory.

#### Scenario: Valid directory path
- **WHEN** the provided path exists and is a directory
- **THEN** the pipeline proceeds

#### Scenario: Path does not exist
- **WHEN** the provided path does not exist
- **THEN** an error is returned: `image path "<path>" does not exist`

#### Scenario: Path is a file, not directory
- **WHEN** the provided path exists but is a file
- **THEN** an error is returned: `image path "<path>" is not a directory`

### Requirement: Pre-run banner

Display a summary banner before starting the pipeline.

#### Scenario: Banner output
- **WHEN** the pipeline is about to start
- **THEN** a banner is printed showing: source path, output path, image count, model name, and mode (DRY RUN if applicable)

### Requirement: Command aliases

Support `image`, `img`, and `images` as command aliases.

#### Scenario: Alias usage
- **WHEN** the user runs `ut img analyze ./photos` or `uteamup images analyze ./photos`
- **THEN** the command executes identically to `uteamup image analyze ./photos`

### Requirement: Exactly one positional argument

The `analyze` subcommand requires exactly one positional argument: the image folder path.

#### Scenario: No path provided
- **WHEN** the user runs `uteamup image analyze` without a path
- **THEN** Cobra returns an error indicating exactly 1 argument is required

#### Scenario: Multiple paths provided
- **WHEN** the user runs `uteamup image analyze ./a ./b`
- **THEN** Cobra returns an error indicating exactly 1 argument is required

### Requirement: New flags for native pipeline

Add new flags that expose native pipeline capabilities not available in the Python wrapper.

#### Scenario: --max-cost flag
- **WHEN** `--max-cost 1.50` is provided
- **THEN** the pipeline stops processing when estimated cost exceeds $1.50

#### Scenario: --resume flag
- **WHEN** `--resume` is provided
- **THEN** the pipeline loads an existing checkpoint file and resumes from where it left off

#### Scenario: --similarity-threshold flag
- **WHEN** `--similarity-threshold 0.80` is provided
- **THEN** the grouper uses 0.80 instead of the default 0.75

#### Scenario: --confidence-threshold flag
- **WHEN** `--confidence-threshold 0.6` is provided
- **THEN** results below 0.6 confidence are reclassified as unclassified
