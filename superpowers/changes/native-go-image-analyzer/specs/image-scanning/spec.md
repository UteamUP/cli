# Image Scanning Specification

Replaces Python `scanner.py` + `utils/image_utils.py` with native Go implementation.

## ADDED Requirements

### Requirement: Recursive folder walking

Walk a configured image folder and discover all valid image files, respecting a recursive flag.

#### Scenario: Recursive scan of nested directories
- **WHEN** `recursive` is true and the image folder contains subdirectories with images
- **THEN** all images in all subdirectories are discovered and returned sorted by path

#### Scenario: Non-recursive scan
- **WHEN** `recursive` is false
- **THEN** only images in the top-level image folder are discovered (no subdirectory traversal)

#### Scenario: Image folder does not exist
- **WHEN** the configured image folder path does not exist or is not a directory
- **THEN** an error is logged and an empty list is returned

### Requirement: Supported format filtering

Only process files whose extension is in the configured `supported_formats` list.

#### Scenario: Supported image formats
- **WHEN** a file has extension `.jpg`, `.jpeg`, `.png`, `.webp`, `.heic`, `.heif`, `.tiff`, or `.bmp`
- **THEN** the file is included in scan results (assuming it passes validation)

#### Scenario: Unsupported file format
- **WHEN** a file has an extension not in the supported formats list (e.g., `.txt`, `.pdf`, `.gif`)
- **THEN** the file is skipped with a debug log entry

### Requirement: Image validation

Verify that each discovered file is a readable, non-corrupted image before including it.

#### Scenario: Valid image file
- **WHEN** a file can be decoded by the Go image package (or goheif for HEIC)
- **THEN** the file is included in scan results

#### Scenario: Corrupted or unreadable image
- **WHEN** a file has a supported extension but cannot be decoded as an image
- **THEN** the file is skipped with a warning log entry

#### Scenario: HEIC/HEIF without library support
- **WHEN** HEIC/HEIF decoding is not available (build tag or missing dependency)
- **THEN** HEIC/HEIF files are skipped with a warning log entry

### Requirement: SHA-256 hashing

Compute a SHA-256 hash of the raw file bytes for every valid image.

#### Scenario: Hash computation
- **WHEN** a valid image is discovered
- **THEN** its SHA-256 hash is computed by streaming 8 KiB chunks and stored as a hex string in `ImageInfo.SHA256Hash`

### Requirement: Perceptual hashing

Compute an average perceptual hash for visual similarity detection.

#### Scenario: Perceptual hash computation
- **WHEN** a valid image is decoded
- **THEN** an average hash is computed (resize to 8x8 grayscale, compare each pixel to mean) and stored as a hex string in `ImageInfo.PerceptualHash`

#### Scenario: Perceptual hash failure
- **WHEN** an image cannot be decoded for perceptual hashing (but file bytes are valid)
- **THEN** `PerceptualHash` is set to empty string and a debug log is emitted

### Requirement: EXIF metadata extraction

Extract useful EXIF fields: date_taken, camera_model, GPS coordinates.

#### Scenario: Image with EXIF data
- **WHEN** an image contains EXIF tags for DateTimeOriginal (tag 36867), Make (271), Model (272), or GPS IFD (0x8825)
- **THEN** the corresponding fields are populated in `ImageInfo.EXIFMetadata` map

#### Scenario: Image without EXIF data
- **WHEN** an image has no EXIF data or EXIF extraction fails
- **THEN** `EXIFMetadata` is an empty map (no error)

### Requirement: Duplicate detection by SHA-256

Detect exact-duplicate images by comparing SHA-256 hashes.

#### Scenario: Two files with identical content
- **WHEN** two image files produce the same SHA-256 hash
- **THEN** the first one seen (by sorted path order) is kept, the second is logged as a duplicate, and only unique images are returned

#### Scenario: No duplicates
- **WHEN** all images have unique SHA-256 hashes
- **THEN** all images are returned and duplicate count is zero

### Requirement: iPhone edit pair detection

Detect iPhone edited variants (`IMG_EXXXX`) paired with originals (`IMG_XXXX`).

#### Scenario: Matching iPhone edit pair
- **WHEN** `IMG_E1234.jpg` and `IMG_1234.jpg` both exist in the scan results
- **THEN** `IMG_E1234.jpg` is marked as `IsIPhoneEdit=true` with `PairedWith="IMG_1234.jpg"`, and a pair mapping is returned (`IMG_1234.jpg` -> `[IMG_E1234.jpg path]`)

#### Scenario: Edit variant without original
- **WHEN** `IMG_E5678.jpg` exists but `IMG_5678.jpg` does not
- **THEN** `IMG_E5678.jpg` is treated as a regular image (no pairing)

#### Scenario: Case-insensitive matching
- **WHEN** filenames use mixed case (e.g., `img_e1234.JPG`, `IMG_1234.jpg`)
- **THEN** the regex matching is case-insensitive and pairs are still detected

### Requirement: Image resizing for API submission

Resize images that exceed a configurable max dimension, preserving aspect ratio.

#### Scenario: Image exceeds max dimension
- **WHEN** an image has width or height greater than `max_image_dimension` (default 2048)
- **THEN** the image is resized proportionally so the largest dimension equals `max_image_dimension`, re-encoded as JPEG at quality 90

#### Scenario: Image within bounds
- **WHEN** an image is already within the max dimension
- **THEN** the image is re-encoded as JPEG at quality 90 without upscaling

### Requirement: HEIC to JPEG conversion

Convert HEIC/HEIF files to JPEG bytes for API submission.

#### Scenario: HEIC file conversion
- **WHEN** a file has `.heic` or `.heif` extension
- **THEN** it is decoded via goheif, converted to RGB, and encoded as JPEG bytes

### Requirement: Filename sanitization

Sanitize filenames to contain only `[a-z0-9_-]` characters.

#### Scenario: Filename with special characters
- **WHEN** a name contains spaces, unicode, or special characters (e.g., `"My Asset (v2) #3!"`)
- **THEN** the sanitized result is `my_asset_v2_3` (lowercase, spaces to underscores, special chars stripped, collapsed duplicates)
