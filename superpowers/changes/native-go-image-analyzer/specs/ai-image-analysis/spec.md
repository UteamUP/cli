# AI Image Analysis Specification

Replaces Python `analyzer.py` + `prompts.py` with native Go Gemini client.

## ADDED Requirements

### Requirement: Gemini API client initialization

Initialize a Gemini generative AI client with configurable model, temperature, and max output tokens.

#### Scenario: Valid API key and model
- **WHEN** a valid `GeminiConfig` is provided with `APIKey`, `Model` (default `gemini-3.1-flash`), `Temperature` (default 0.1), `MaxOutputTokens` (default 4096)
- **THEN** a `GeminiAnalyzer` is created with a configured generative model, rate limiter, and retry handler

#### Scenario: Missing API key in non-dry-run mode
- **WHEN** `APIKey` is empty and the pipeline is not in dry-run mode
- **THEN** initialization returns an error: "GEMINI_API_KEY is required"

### Requirement: Single-image multi-entity analysis

Send an image to Gemini with the unified analysis prompt and parse multiple entities from the response.

#### Scenario: Image with single entity
- **WHEN** an image of one asset is analyzed
- **THEN** the response JSON `{"entities": [{"classification": ..., "extracted_data": ..., "related_to": null}]}` is parsed into one `ImageAnalysisResult`

#### Scenario: Image with multiple entities
- **WHEN** an image contains an asset with visible tools and parts
- **THEN** the response contains multiple entities, each with its own classification, extracted data, and `related_to` field linking child entities to the primary entity name

#### Scenario: Unclassified image
- **WHEN** Gemini cannot determine the entity type
- **THEN** `primary_type` is `"unclassified"`, confidence is low, and `extracted_data` is nil

### Requirement: Entity type classification

Classify each entity into one of five types: asset, tool, part, chemical, unclassified.

#### Scenario: Asset classification
- **WHEN** an image shows fixed or movable equipment (pumps, generators, forklifts, HVAC, vehicles)
- **THEN** classification `primary_type` is `"asset"` with entity-specific fields: name, description, serial_number, reference_number, model_number, upc_number, manufacturer_brand, visible_condition, is_vehicle, vehicle_type, license_plate, suggested_category, suggested_vendor, suggested_location, icon_name, additional_info, notes

#### Scenario: Tool classification
- **WHEN** an image shows handheld or portable instruments
- **THEN** classification `primary_type` is `"tool"` with fields: name, description, dimensions (width/height/length/depth/weight), value, barcode_number, serial_number, reference_number, model_number, tool_number, manufacturer_brand, suggested_category, suggested_vendor

#### Scenario: Part classification
- **WHEN** an image shows spare parts, components, or consumables
- **THEN** classification `primary_type` is `"part"` with fields: name, description, serial_number, reference_number, model_number, part_number, value, manufacturer_brand, suggested_category, suggested_vendor

#### Scenario: Chemical classification
- **WHEN** an image shows chemical products or hazardous materials
- **THEN** classification `primary_type` is `"chemical"` with fields: name, description, chemical_formula, cas_number, ec_number, un_number, ghs_hazard_class, signal_word, physical_state, color, ph, flash_point, storage_class, storage_requirements, PPE fields (respiratory/hand/eye/skin protection), first_aid_measures, firefighting_measures, spill_leak_procedures, disposal_considerations, unit_of_measure, hazard_statements (list), precautionary_statements (list), manufacturer_name, suggested_category, suggested_vendor

### Requirement: Confidence scoring and review flagging

Flag low-confidence results for human review.

#### Scenario: Confidence below threshold (0.5)
- **WHEN** Gemini returns a classification with confidence < 0.5
- **THEN** `FlaggedForReview` is true and `ReviewReason` is `"Low confidence: 0.XX"`

#### Scenario: Confidence at or above threshold
- **WHEN** confidence >= 0.5
- **THEN** `FlaggedForReview` is false (unless other error conditions apply)

### Requirement: JSON response parsing with markdown fence stripping

Parse Gemini's JSON response, handling markdown fences and both multi-entity and legacy single-entity formats.

#### Scenario: Response wrapped in markdown fences
- **WHEN** Gemini returns `` ```json\n{...}\n``` ``
- **THEN** the fences are stripped before JSON parsing

#### Scenario: Multi-entity format
- **WHEN** response contains `{"entities": [...]}`
- **THEN** each element in the array is parsed as a separate entity

#### Scenario: Legacy single-entity format
- **WHEN** response contains `{"classification": ..., "extracted_data": ...}` without `entities` wrapper
- **THEN** it is treated as a single-entity response

### Requirement: JSON repair via Gemini

When initial JSON parsing fails, send the broken text back to Gemini for correction.

#### Scenario: Broken JSON response
- **WHEN** the initial response fails JSON parsing
- **THEN** a JSON fix prompt is sent to Gemini: "Fix this broken JSON and return ONLY valid JSON"
- **THEN** the corrected response is parsed; if it also fails, the image is flagged as unclassified with review reason "JSON parse failure"

### Requirement: Cost estimation

Estimate API cost based on Gemini 2.0 Flash pricing before processing.

#### Scenario: Cost estimate for N images
- **WHEN** `EstimateCost(imageCount)` is called
- **THEN** returns estimated input tokens (258 per image + 1500 prompt overhead), output tokens (500 per response), input cost ($0.10/1M tokens), output cost ($0.40/1M tokens), and total cost in USD

### Requirement: Running cost tracking

Track cumulative token usage and cost across all analyzed images.

#### Scenario: After analyzing 10 images
- **WHEN** 10 images have been processed
- **THEN** `TotalCost()` returns cumulative input/output tokens and USD cost based on per-image estimates

### Requirement: Rate limiting integration

Acquire a rate limiter token before each Gemini API call.

#### Scenario: Within rate limit
- **WHEN** tokens are available in the bucket
- **THEN** the API call proceeds immediately

#### Scenario: Rate limit exhausted
- **WHEN** no tokens are available
- **THEN** the caller blocks until a token refills (based on requests_per_minute)

### Requirement: Retry handling integration

Retry transient Gemini API errors with exponential backoff.

#### Scenario: Transient 429 or 5xx error
- **WHEN** Gemini returns HTTP 429, 500, 502, 503, or 504
- **THEN** the request is retried up to `max_retries` times with exponential backoff + jitter

#### Scenario: Permanent 4xx error (not 429)
- **WHEN** Gemini returns HTTP 400, 401, 403, etc.
- **THEN** the error is returned immediately (no retry)

#### Scenario: API error fallback result
- **WHEN** all retries are exhausted or a non-retryable error occurs
- **THEN** an `ImageAnalysisResult` with `primary_type=unclassified`, confidence=0.0, `FlaggedForReview=true`, and `ReviewReason="API error: <message>"` is returned

### Requirement: Unified analysis prompt

Use the exact CMMS-focused prompt that instructs Gemini to identify ALL distinct entities visible in an image.

#### Scenario: Prompt content
- **WHEN** the analyzer sends a request to Gemini
- **THEN** the prompt includes: entity type definitions (asset/tool/part/chemical/unclassified), relationship instructions (`related_to` field), critical rules (only visually present data, `[?]` for partially readable text, null for missing fields), and entity-specific JSON field schemas for all four types
