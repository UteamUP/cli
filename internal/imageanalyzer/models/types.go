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
	GPSLatitude    float64                `json:"gps_latitude"`
	GPSLongitude   float64                `json:"gps_longitude"`
	HasGPS         bool                   `json:"has_gps"`
}

// DetectedVendor represents a vendor detected across multiple entities.
type DetectedVendor struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Email       string `json:"email"`
	Website     string `json:"website"`
	PhoneNumber string `json:"phone_number"`
	// Linking info
	EntityNames []string `json:"entity_names"`
	EntityTypes []string `json:"entity_types"`
	ImagePaths  []string `json:"image_paths"`
	Count       int      `json:"count"`
}

// DetectedLocation represents a location detected from GPS data or AI analysis.
type DetectedLocation struct {
	Name             string  `json:"name"`
	Description      string  `json:"description"`
	Street           string  `json:"street"`
	City             string  `json:"city"`
	State            string  `json:"state"`
	ZipCode          string  `json:"zip_code"`
	PostalCode       string  `json:"postal_code"`
	Country          string  `json:"country"`
	GooglePlaceId    string  `json:"google_place_id"`
	FormattedAddress string  `json:"formatted_address"`
	Latitude         float64 `json:"latitude"`
	Longitude        float64 `json:"longitude"`
	GoogleMapsUrl    string  `json:"google_maps_url"`
	HasGPS           bool    `json:"has_gps"`
	Source           string  `json:"source"` // "gps_exif", "gemini_suggested", "reverse_geocoded"
	// Linking info
	EntityNames []string `json:"entity_names"`
	EntityTypes []string `json:"entity_types"`
	ImagePaths  []string `json:"image_paths"`
	Count       int      `json:"count"`
}
