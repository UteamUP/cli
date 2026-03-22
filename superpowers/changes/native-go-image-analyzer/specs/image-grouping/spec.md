# Image Grouping Specification

Replaces Python `grouper.py` with native Go implementation.

## ADDED Requirements

### Requirement: Entity type partitioning

Partition analysis results by entity type before grouping. Unclassified items are never grouped.

#### Scenario: Mixed entity types
- **WHEN** results contain assets, tools, parts, chemicals, and unclassified items
- **THEN** each entity type is clustered independently; unclassified items become solo groups (one image per group)

#### Scenario: All unclassified
- **WHEN** all results are unclassified
- **THEN** each result is its own group with `GroupConfidence` equal to its classification confidence

### Requirement: iPhone edit pair pre-merging

Merge iPhone `IMG_E*` edit pairs before clustering so they don't form separate groups.

#### Scenario: Edit pair with original
- **WHEN** an analysis result has `PairedImages` referencing another result's path
- **THEN** the paired image is absorbed into the primary result and removed from the clustering input

### Requirement: Serial number exact matching

Group items that share the same non-empty serial number (strongest grouping signal).

#### Scenario: Two images with same serial number
- **WHEN** two results in the same entity-type partition have identical non-empty `serial_number` values
- **THEN** they are placed in the same group

#### Scenario: Empty serial numbers
- **WHEN** results have no serial number
- **THEN** they fall through to name matching or similarity clustering

### Requirement: Name exact matching (case-insensitive)

Group items that share the exact same name within the same entity type.

#### Scenario: Same name, different case
- **WHEN** two results have names "Hydraulic Pump" and "hydraulic pump"
- **THEN** they are grouped together (case-insensitive, trimmed comparison)

#### Scenario: Single item with unique name
- **WHEN** only one result has a particular name
- **THEN** it falls through to similarity clustering (not grouped by name alone)

### Requirement: Agglomerative similarity clustering

Single-linkage agglomerative clustering on remaining ungrouped items using weighted multi-signal similarity.

#### Scenario: Two similar items above threshold
- **WHEN** the pairwise similarity score between two results exceeds the `similarity_threshold` (default 0.75)
- **THEN** their clusters are merged

#### Scenario: Similarity below threshold
- **WHEN** no pair across two clusters exceeds the threshold
- **THEN** the clusters remain separate

### Requirement: Weighted multi-signal similarity computation

Compute pairwise similarity using six weighted signals.

#### Scenario: Similarity weight breakdown
- **WHEN** comparing two `ImageAnalysisResult` objects of the same entity type
- **THEN** the similarity score is computed as:
  - Serial number exact match: 0.40
  - Model number exact match: 0.20
  - Name fuzzy match (Levenshtein ratio): 0.20
  - Description fuzzy match (Levenshtein ratio): 0.10
  - Perceptual hash similarity (normalized Hamming distance): 0.05
  - Manufacturer/brand exact match (case-insensitive): 0.05

#### Scenario: Different entity types
- **WHEN** two results have different `primary_type` values
- **THEN** similarity is 0.0 (never grouped across entity types)

### Requirement: Perceptual hash similarity

Compare two hex-encoded perceptual hashes via normalized Hamming distance.

#### Scenario: Identical perceptual hashes
- **WHEN** two results have the same perceptual hash hex string
- **THEN** perceptual hash similarity is 1.0

#### Scenario: Completely different hashes
- **WHEN** the Hamming distance is maximal
- **THEN** perceptual hash similarity approaches 0.0

### Requirement: Representative selection

Select the highest-confidence result as the representative for each group.

#### Scenario: Group with varying confidence
- **WHEN** a group contains results with confidences [0.92, 0.85, 0.78]
- **THEN** the result with confidence 0.92 is selected as the primary/representative

### Requirement: Extracted data merging

Fill null fields on the representative from member data, preferring higher-confidence members.

#### Scenario: Representative missing serial number, member has it
- **WHEN** the representative's `serial_number` is nil but a member (confidence 0.85) has `serial_number = "SN-1234"`
- **THEN** the representative's `serial_number` is set to `"SN-1234"`

#### Scenario: Representative already has a value
- **WHEN** the representative's `name` is already set
- **THEN** it is not overwritten by member data (only nil fields are filled)

### Requirement: Fuzzy string matching

Use Levenshtein distance ratio for name and description comparison (Go equivalent of Python `thefuzz.fuzz.ratio`).

#### Scenario: Similar strings
- **WHEN** comparing "Hydraulic Pump Model A" and "Hydraulic Pump Model B"
- **THEN** the ratio is high (close to 1.0) since most characters match

#### Scenario: Completely different strings
- **WHEN** comparing "Wrench" and "Lubricant"
- **THEN** the ratio is low (close to 0.0)
