package models

import "time"

// ImageAnalysisResult is the complete analysis result for a single entity found in an image.
type ImageAnalysisResult struct {
	ImagePath        string                 `json:"image_path"`
	OriginalFilename string                 `json:"original_filename"`
	FileHashSHA256   string                 `json:"file_hash_sha256"`
	PerceptualHash   string                 `json:"perceptual_hash"`
	Classification   ClassificationResult   `json:"classification"`
	ExtractedData    ExtractedData          `json:"extracted_data"`
	EXIFMetadata     map[string]interface{} `json:"exif_metadata"`
	FlaggedForReview bool                   `json:"flagged_for_review"`
	ReviewReason     string                 `json:"review_reason"`
	ProcessedAt      time.Time              `json:"processed_at"`
	PairedImages     []string               `json:"paired_images"`
	RelatedTo        string                 `json:"related_to"`
}

// ImageGroup represents a group of images depicting the same physical item.
type ImageGroup struct {
	Primary         ImageAnalysisResult   `json:"primary"`
	Members         []ImageAnalysisResult `json:"members"`
	GroupConfidence float64               `json:"group_confidence"`
}

// AllImagePaths collects all image paths from Primary, Members, and their PairedImages.
func (g *ImageGroup) AllImagePaths() []string {
	paths := []string{g.Primary.ImagePath}
	for _, m := range g.Members {
		paths = append(paths, m.ImagePath)
		paths = append(paths, m.PairedImages...)
	}
	paths = append(paths, g.Primary.PairedImages...)
	return paths
}

// AllOriginalFilenames collects all original filenames from Primary and Members.
func (g *ImageGroup) AllOriginalFilenames() []string {
	names := []string{g.Primary.OriginalFilename}
	for _, m := range g.Members {
		names = append(names, m.OriginalFilename)
	}
	return names
}
