package models

// EntityType represents the classification type for an analyzed image.
type EntityType string

const (
	EntityTypeAsset        EntityType = "asset"
	EntityTypeTool         EntityType = "tool"
	EntityTypePart         EntityType = "part"
	EntityTypeChemical     EntityType = "chemical"
	EntityTypeUnclassified EntityType = "unclassified"
)

// ClassificationResult holds the AI classification output for an image.
type ClassificationResult struct {
	PrimaryType   EntityType `json:"primary_type"`
	Confidence    float64    `json:"confidence"`
	SecondaryType *string    `json:"secondary_type,omitempty"`
	Reasoning     string     `json:"reasoning"`
}

// ImageInfo holds metadata about an image file.
type ImageInfo struct {
	Path           string                 `json:"path"`
	Filename       string                 `json:"filename"`
	Extension      string                 `json:"extension"`
	FileSizeBytes  int64                  `json:"file_size_bytes"`
	SHA256Hash     string                 `json:"sha256_hash"`
	PerceptualHash string                 `json:"perceptual_hash"`
	EXIFMetadata   map[string]interface{} `json:"exif_metadata"`
	IsIPhoneEdit   bool                   `json:"is_iphone_edit"`
	PairedWith     string                 `json:"paired_with"`
}
