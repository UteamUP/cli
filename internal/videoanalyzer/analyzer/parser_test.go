package analyzer

import (
	"testing"

	"github.com/uteamup/cli/internal/imageanalyzer/models"
)

func TestParseVideoResponse_SingleAsset(t *testing.T) {
	jsonText := `{"entities": [{"type": "asset", "timestamp": "01:30", "confidence": 0.92, "reasoning": "Industrial compressor visible", "flagged_for_review": false, "review_reason": null, "related_to": null, "extracted_data": {"name": "Air Compressor Unit", "description": "Large industrial air compressor", "serial_number": "AC-2024-001", "reference_number": null, "model_number": "XR-500", "upc_number": null, "additional_info": null, "notes": "Good condition", "check_in_procedure": null, "check_out_procedure": null, "icon_name": "settings", "suggested_vendor": "Atlas Copco", "suggested_category": "Manufacturing", "suggested_location": null, "manufacturer_brand": "Atlas Copco", "visible_condition": "Good", "is_vehicle": false, "vehicle_type": null, "license_plate": null, "asset_category_group": "Manufacturing"}}]}`

	results, err := ParseVideoResponse(jsonText, "/videos/test.mp4", "test.mp4")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	r := results[0]
	if r.Classification.PrimaryType != models.EntityTypeAsset {
		t.Errorf("expected primary type %q, got %q", models.EntityTypeAsset, r.Classification.PrimaryType)
	}
	if r.Classification.Confidence != 0.92 {
		t.Errorf("expected confidence 0.92, got %f", r.Classification.Confidence)
	}
	if r.ExtractedData.Asset == nil {
		t.Fatal("expected ExtractedData.Asset to be non-nil")
	}
	if r.ExtractedData.Asset.Name != "Air Compressor Unit" {
		t.Errorf("expected asset name %q, got %q", "Air Compressor Unit", r.ExtractedData.Asset.Name)
	}
	if r.ImagePath != "/videos/test.mp4" {
		t.Errorf("expected image path %q, got %q", "/videos/test.mp4", r.ImagePath)
	}
	if r.OriginalFilename != "test.mp4" {
		t.Errorf("expected original filename %q, got %q", "test.mp4", r.OriginalFilename)
	}

	// Verify timestamp stored in EXIFMetadata.
	ts, ok := r.EXIFMetadata["video_timestamp"]
	if !ok {
		t.Fatal("expected video_timestamp in EXIFMetadata")
	}
	if ts != "01:30" {
		t.Errorf("expected timestamp %q, got %v", "01:30", ts)
	}
}

func TestParseVideoResponse_MultipleEntities(t *testing.T) {
	jsonText := `{"entities": [
		{"type": "asset", "timestamp": "00:10", "confidence": 0.90, "reasoning": "Machine visible", "flagged_for_review": false, "review_reason": null, "related_to": null, "extracted_data": {"name": "CNC Machine", "description": "CNC milling machine"}},
		{"type": "tool", "timestamp": "00:45", "confidence": 0.85, "reasoning": "Hand tool visible", "flagged_for_review": false, "review_reason": null, "related_to": null, "extracted_data": {"name": "Torque Wrench"}},
		{"type": "part", "timestamp": "01:20", "confidence": 0.88, "reasoning": "Spare part visible", "flagged_for_review": false, "review_reason": null, "related_to": null, "extracted_data": {"name": "Bearing Assembly"}}
	]}`

	results, err := ParseVideoResponse(jsonText, "/videos/multi.mp4", "multi.mp4")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	expected := []struct {
		entityType models.EntityType
		name       string
	}{
		{models.EntityTypeAsset, "CNC Machine"},
		{models.EntityTypeTool, "Torque Wrench"},
		{models.EntityTypePart, "Bearing Assembly"},
	}

	for i, exp := range expected {
		if results[i].Classification.PrimaryType != exp.entityType {
			t.Errorf("result[%d]: expected type %q, got %q", i, exp.entityType, results[i].Classification.PrimaryType)
		}
		if results[i].ExtractedData.GetName() != exp.name {
			t.Errorf("result[%d]: expected name %q, got %q", i, exp.name, results[i].ExtractedData.GetName())
		}
	}
}

func TestParseVideoResponse_EmptyEntities(t *testing.T) {
	jsonText := `{"entities": []}`

	results, err := ParseVideoResponse(jsonText, "/videos/empty.mp4", "empty.mp4")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results != nil {
		t.Errorf("expected nil results for empty entities, got %d results", len(results))
	}
}

func TestParseVideoResponse_FlaggedForReview(t *testing.T) {
	jsonText := `{"entities": [{"type": "asset", "timestamp": "02:00", "confidence": 0.45, "reasoning": "Blurry image", "flagged_for_review": true, "review_reason": "Low confidence due to motion blur", "related_to": null, "extracted_data": {"name": "Unknown Equipment"}}]}`

	results, err := ParseVideoResponse(jsonText, "/videos/blurry.mp4", "blurry.mp4")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	r := results[0]
	if !r.FlaggedForReview {
		t.Error("expected FlaggedForReview to be true")
	}
	if r.ReviewReason != "Low confidence due to motion blur" {
		t.Errorf("expected review reason %q, got %q", "Low confidence due to motion blur", r.ReviewReason)
	}
}

func TestParseVideoResponse_UnclassifiedEntity(t *testing.T) {
	jsonText := `{"entities": [{"type": "unclassified", "timestamp": "03:15", "confidence": 0.30, "reasoning": "Cannot determine type", "flagged_for_review": true, "review_reason": "Unidentifiable object", "related_to": null, "extracted_data": null}]}`

	results, err := ParseVideoResponse(jsonText, "/videos/unknown.mp4", "unknown.mp4")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	r := results[0]
	if r.Classification.PrimaryType != models.EntityTypeUnclassified {
		t.Errorf("expected type %q, got %q", models.EntityTypeUnclassified, r.Classification.PrimaryType)
	}
	// ExtractedData should have no populated fields.
	if r.ExtractedData.Asset != nil || r.ExtractedData.Tool != nil || r.ExtractedData.Part != nil || r.ExtractedData.Chemical != nil {
		t.Error("expected all ExtractedData fields to be nil for unclassified entity")
	}
}

func TestParseVideoResponse_InvalidJSON(t *testing.T) {
	jsonText := `{"entities": [INVALID`

	_, err := ParseVideoResponse(jsonText, "/videos/bad.mp4", "bad.mp4")
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestParseVideoResponse_MarkdownFences(t *testing.T) {
	jsonText := "```json\n" + `{"entities": [{"type": "tool", "timestamp": "00:05", "confidence": 0.91, "reasoning": "Wrench visible", "flagged_for_review": false, "review_reason": null, "related_to": null, "extracted_data": {"name": "Socket Wrench Set"}}]}` + "\n```"

	results, err := ParseVideoResponse(jsonText, "/videos/fenced.mp4", "fenced.mp4")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].ExtractedData.Tool == nil {
		t.Fatal("expected ExtractedData.Tool to be non-nil")
	}
	if results[0].ExtractedData.Tool.Name != "Socket Wrench Set" {
		t.Errorf("expected tool name %q, got %q", "Socket Wrench Set", results[0].ExtractedData.Tool.Name)
	}
}

func TestParseVideoResponse_RelatedTo(t *testing.T) {
	jsonText := `{"entities": [{"type": "part", "timestamp": "01:00", "confidence": 0.87, "reasoning": "Part visible near machine", "flagged_for_review": false, "review_reason": null, "related_to": "CNC Machine", "extracted_data": {"name": "Spindle Motor"}}]}`

	results, err := ParseVideoResponse(jsonText, "/videos/related.mp4", "related.mp4")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].RelatedTo != "CNC Machine" {
		t.Errorf("expected RelatedTo %q, got %q", "CNC Machine", results[0].RelatedTo)
	}
}

func TestGetTimestamp(t *testing.T) {
	t.Run("with timestamp", func(t *testing.T) {
		r := &models.ImageAnalysisResult{
			EXIFMetadata: map[string]interface{}{
				"video_timestamp": "01:30",
			},
		}
		got := GetTimestamp(r)
		if got != "01:30" {
			t.Errorf("expected %q, got %q", "01:30", got)
		}
	})

	t.Run("without EXIFMetadata", func(t *testing.T) {
		r := &models.ImageAnalysisResult{}
		got := GetTimestamp(r)
		if got != "" {
			t.Errorf("expected empty string, got %q", got)
		}
	})

	t.Run("nil EXIFMetadata", func(t *testing.T) {
		r := &models.ImageAnalysisResult{
			EXIFMetadata: nil,
		}
		got := GetTimestamp(r)
		if got != "" {
			t.Errorf("expected empty string, got %q", got)
		}
	})
}

func TestMapEntityType(t *testing.T) {
	tests := []struct {
		input    string
		expected models.EntityType
	}{
		{"asset", models.EntityTypeAsset},
		{"TOOL", models.EntityTypeTool},
		{"Chemical", models.EntityTypeChemical},
		{"part", models.EntityTypePart},
		{"unknown", models.EntityTypeUnclassified},
		{"", models.EntityTypeUnclassified},
		{"  asset  ", models.EntityTypeAsset},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := mapEntityType(tc.input)
			if got != tc.expected {
				t.Errorf("mapEntityType(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}
