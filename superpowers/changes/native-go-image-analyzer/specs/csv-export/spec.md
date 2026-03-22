# CSV Export Specification

Replaces Python `exporter.py` with native Go implementation.

## ADDED Requirements

### Requirement: Per-entity-type CSV export

Write one CSV file per entity type, using entity-specific column schemas.

#### Scenario: Mixed entity types after grouping
- **WHEN** groups contain assets, tools, parts, chemicals, and unclassified items
- **THEN** five CSV files are written: `assets.csv`, `tools.csv`, `parts.csv`, `chemicals.csv`, `unclassifieds.csv` (only for types that have data)

#### Scenario: No groups for a type
- **WHEN** no groups exist for a particular entity type
- **THEN** no CSV file is created for that type

### Requirement: Asset CSV columns

Asset CSV must contain these columns in order:
`name, description, serial_number, reference_number, model_number, upc_number, additional_info, notes, check_in_procedure, check_out_procedure, icon_name, suggested_vendor, suggested_category, suggested_location, manufacturer_brand, visible_condition, is_vehicle, vehicle_type, license_plate, related_to, image_paths, original_filenames, confidence_score, flagged_for_review, review_reason`

#### Scenario: Asset row population
- **WHEN** an asset group is exported
- **THEN** all extracted data fields are mapped to their columns, `image_paths` and `original_filenames` are semicolon-separated lists of all images in the group, and `related_to` contains the parent entity name or empty string

### Requirement: Tool CSV columns

Tool CSV must contain these columns in order:
`name, description, width, height, length, depth, weight, value, barcode_number, serial_number, reference_number, model_number, tool_number, additional_info, notes, suggested_vendor, suggested_category, manufacturer_brand, related_to, image_paths, original_filenames, confidence_score, flagged_for_review, review_reason`

#### Scenario: Tool row population
- **WHEN** a tool group is exported
- **THEN** numeric fields (width, height, length, depth, weight, value) are formatted as numbers; nil values are empty strings

### Requirement: Part CSV columns

Part CSV must contain these columns in order:
`name, description, serial_number, reference_number, model_number, part_number, additional_info, notes, value, suggested_vendor, suggested_category, manufacturer_brand, related_to, image_paths, original_filenames, confidence_score, flagged_for_review, review_reason`

#### Scenario: Part row population
- **WHEN** a part group is exported
- **THEN** all fields are mapped; `value` is formatted as a number or empty

### Requirement: Chemical CSV columns

Chemical CSV must contain these columns in order:
`name, description, chemical_formula, cas_number, ec_number, un_number, ghs_hazard_class, signal_word, physical_state, color, odor, ph, melting_point, boiling_point, flash_point, solubility, storage_class, storage_requirements, respiratory_protection, hand_protection, eye_protection, skin_protection, first_aid_measures, firefighting_measures, spill_leak_procedures, disposal_considerations, unit_of_measure, hazard_statements, precautionary_statements, manufacturer_name, suggested_vendor, suggested_category, related_to, image_paths, original_filenames, confidence_score, flagged_for_review, review_reason`

#### Scenario: Chemical list fields
- **WHEN** a chemical has `hazard_statements: ["H302: Harmful if swallowed", "H315: Causes skin irritation"]`
- **THEN** the CSV cell contains `"H302: Harmful if swallowed; H315: Causes skin irritation"` (semicolon-separated)

### Requirement: Unclassified CSV columns

Unclassified CSV must contain these columns in order:
`original_filename, image_path, confidence_score, flagged_for_review, review_reason, classification_reasoning, related_to`

#### Scenario: Unclassified items
- **WHEN** unclassified results exist
- **THEN** each is written as its own row with the reasoning from classification

### Requirement: Image renaming (copy with descriptive names)

Copy images to a designated folder with descriptive filenames based on entity type, name, sequence, and date.

#### Scenario: Rename enabled
- **WHEN** `rename_images` is true
- **THEN** each image is copied (not moved) to the renamed images folder with pattern `{entity_type}_{sanitized_name}_{seq:03d}_{YYYYMMDD}.{ext}`

#### Scenario: Filename collision
- **WHEN** the target filename already exists
- **THEN** the sequence number is incremented until a unique name is found

#### Scenario: Rename disabled
- **WHEN** `rename_images` is false
- **THEN** no images are copied/renamed

#### Scenario: Source image missing
- **WHEN** the source image file no longer exists at export time
- **THEN** a warning is logged and the image is skipped

### Requirement: Summary report generation

Generate a Markdown summary report with processing statistics.

#### Scenario: After successful pipeline run
- **WHEN** export is complete
- **THEN** a `summary_report.md` is written to the output folder containing: date, total images processed, groups formed, flagged for review count, duplicates found, processing duration, and per-entity-type breakdown table

### Requirement: Output folder creation

Create output and renamed images folders if they don't exist.

#### Scenario: Output folder does not exist
- **WHEN** the configured output folder path does not exist
- **THEN** it is created with `MkdirAll` (parents included)
