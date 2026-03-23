package pipeline

import (
	"testing"

	"github.com/uteamup/cli/internal/imageanalyzer/models"
)

func makeResult(videoPath, name string, entityType models.EntityType, timestamp string, confidence float64) models.ImageAnalysisResult {
	result := models.ImageAnalysisResult{
		ImagePath: videoPath,
		Classification: models.ClassificationResult{
			PrimaryType: entityType,
			Confidence:  confidence,
		},
		EXIFMetadata: map[string]interface{}{
			"video_timestamp": timestamp,
		},
	}
	// Set name based on entity type.
	switch entityType {
	case models.EntityTypeAsset:
		result.ExtractedData.Asset = &models.ExtractedAssetData{Name: name}
	case models.EntityTypeTool:
		result.ExtractedData.Tool = &models.ExtractedToolData{Name: name}
	case models.EntityTypePart:
		result.ExtractedData.Part = &models.ExtractedPartData{Name: name}
	}
	return result
}

func TestTemporalDedup_SameNameSameVideo(t *testing.T) {
	results := []models.ImageAnalysisResult{
		makeResult("/videos/a.mp4", "Compressor", models.EntityTypeAsset, "00:15", 0.90),
		makeResult("/videos/a.mp4", "Compressor", models.EntityTypeAsset, "00:30", 0.85),
	}

	deduped := TemporalDedup(results, 30)
	if len(deduped) != 1 {
		t.Fatalf("expected 1 result after dedup, got %d", len(deduped))
	}
	if deduped[0].ExtractedData.GetName() != "Compressor" {
		t.Errorf("expected name %q, got %q", "Compressor", deduped[0].ExtractedData.GetName())
	}
}

func TestTemporalDedup_SameNameDifferentTimestamps(t *testing.T) {
	results := []models.ImageAnalysisResult{
		makeResult("/videos/a.mp4", "Compressor", models.EntityTypeAsset, "00:15", 0.90),
		makeResult("/videos/a.mp4", "Compressor", models.EntityTypeAsset, "02:00", 0.88),
	}

	deduped := TemporalDedup(results, 30)
	if len(deduped) != 2 {
		t.Fatalf("expected 2 results (timestamps too far apart), got %d", len(deduped))
	}
}

func TestTemporalDedup_DifferentTypes(t *testing.T) {
	results := []models.ImageAnalysisResult{
		makeResult("/videos/a.mp4", "Widget", models.EntityTypeAsset, "00:15", 0.90),
		makeResult("/videos/a.mp4", "Widget", models.EntityTypeTool, "00:20", 0.85),
	}

	deduped := TemporalDedup(results, 30)
	if len(deduped) != 2 {
		t.Fatalf("expected 2 results (different types), got %d", len(deduped))
	}
}

func TestTemporalDedup_DifferentVideos(t *testing.T) {
	results := []models.ImageAnalysisResult{
		makeResult("/videos/a.mp4", "Compressor", models.EntityTypeAsset, "00:15", 0.90),
		makeResult("/videos/b.mp4", "Compressor", models.EntityTypeAsset, "00:20", 0.88),
	}

	deduped := TemporalDedup(results, 30)
	if len(deduped) != 2 {
		t.Fatalf("expected 2 results (different videos), got %d", len(deduped))
	}
}

func TestTemporalDedup_SingleResult(t *testing.T) {
	results := []models.ImageAnalysisResult{
		makeResult("/videos/a.mp4", "Drill", models.EntityTypeTool, "00:30", 0.95),
	}

	deduped := TemporalDedup(results, 30)
	if len(deduped) != 1 {
		t.Fatalf("expected 1 result, got %d", len(deduped))
	}
}

func TestTemporalDedup_EmptyResults(t *testing.T) {
	var results []models.ImageAnalysisResult

	deduped := TemporalDedup(results, 30)
	if len(deduped) != 0 {
		t.Fatalf("expected 0 results, got %d", len(deduped))
	}
}

func TestTemporalDedup_KeepsHigherConfidence(t *testing.T) {
	results := []models.ImageAnalysisResult{
		makeResult("/videos/a.mp4", "Compressor", models.EntityTypeAsset, "00:15", 0.70),
		makeResult("/videos/a.mp4", "Compressor", models.EntityTypeAsset, "00:25", 0.95),
	}

	deduped := TemporalDedup(results, 30)
	if len(deduped) != 1 {
		t.Fatalf("expected 1 result after merge, got %d", len(deduped))
	}
	if deduped[0].Classification.Confidence != 0.95 {
		t.Errorf("expected confidence 0.95 (higher), got %f", deduped[0].Classification.Confidence)
	}
}

func TestParseTimestampSec(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"00:15", 15},
		{"01:30", 90},
		{"10:05", 605},
		{"", -1},
		{"invalid", -1},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := parseTimestampSec(tc.input)
			if got != tc.expected {
				t.Errorf("parseTimestampSec(%q) = %d, want %d", tc.input, got, tc.expected)
			}
		})
	}
}
