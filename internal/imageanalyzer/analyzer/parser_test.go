package analyzer

import (
	"testing"

	"github.com/uteamup/cli/internal/imageanalyzer/models"
)

func TestParseMultiEntity(t *testing.T) {
	input := `{
		"entities": [
			{
				"classification": {
					"primary_type": "asset",
					"confidence": 0.95,
					"reasoning": "Industrial pump visible"
				},
				"related_to": null,
				"extracted_data": {
					"name": "Centrifugal Pump",
					"description": "Industrial centrifugal pump"
				}
			},
			{
				"classification": {
					"primary_type": "part",
					"confidence": 0.85,
					"reasoning": "Filter attached to pump"
				},
				"related_to": "Centrifugal Pump",
				"extracted_data": {
					"name": "Oil Filter",
					"description": "Replacement oil filter"
				}
			}
		]
	}`

	results := parseMultiEntityResponse(input, "/images/pump.jpg")

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// First entity: asset.
	if results[0].Classification.PrimaryType != models.EntityTypeAsset {
		t.Errorf("expected asset, got %s", results[0].Classification.PrimaryType)
	}
	if results[0].Classification.Confidence != 0.95 {
		t.Errorf("expected confidence 0.95, got %f", results[0].Classification.Confidence)
	}
	if results[0].ExtractedData.Asset == nil {
		t.Fatal("expected asset extracted data, got nil")
	}
	if results[0].ExtractedData.Asset.Name != "Centrifugal Pump" {
		t.Errorf("expected 'Centrifugal Pump', got %q", results[0].ExtractedData.Asset.Name)
	}
	if results[0].RelatedTo != "" {
		t.Errorf("expected empty related_to for primary, got %q", results[0].RelatedTo)
	}

	// Second entity: part.
	if results[1].Classification.PrimaryType != models.EntityTypePart {
		t.Errorf("expected part, got %s", results[1].Classification.PrimaryType)
	}
	if results[1].ExtractedData.Part == nil {
		t.Fatal("expected part extracted data, got nil")
	}
	if results[1].ExtractedData.Part.Name != "Oil Filter" {
		t.Errorf("expected 'Oil Filter', got %q", results[1].ExtractedData.Part.Name)
	}
	if results[1].RelatedTo != "Centrifugal Pump" {
		t.Errorf("expected related_to 'Centrifugal Pump', got %q", results[1].RelatedTo)
	}
}

func TestParseLegacySingle(t *testing.T) {
	input := `{
		"classification": {
			"primary_type": "tool",
			"confidence": 0.90,
			"reasoning": "Wrench visible"
		},
		"extracted_data": {
			"name": "Adjustable Wrench",
			"description": "12-inch adjustable wrench"
		}
	}`

	results := parseMultiEntityResponse(input, "/images/wrench.jpg")

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if results[0].Classification.PrimaryType != models.EntityTypeTool {
		t.Errorf("expected tool, got %s", results[0].Classification.PrimaryType)
	}
	if results[0].ExtractedData.Tool == nil {
		t.Fatal("expected tool extracted data, got nil")
	}
	if results[0].ExtractedData.Tool.Name != "Adjustable Wrench" {
		t.Errorf("expected 'Adjustable Wrench', got %q", results[0].ExtractedData.Tool.Name)
	}
}

func TestParseMarkdownFences(t *testing.T) {
	input := "```json\n" + `{
		"entities": [
			{
				"classification": {
					"primary_type": "chemical",
					"confidence": 0.88,
					"reasoning": "Lubricant bottle"
				},
				"related_to": null,
				"extracted_data": {
					"name": "WD-40",
					"description": "Multi-use lubricant spray"
				}
			}
		]
	}` + "\n```"

	results := parseMultiEntityResponse(input, "/images/wd40.jpg")

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if results[0].Classification.PrimaryType != models.EntityTypeChemical {
		t.Errorf("expected chemical, got %s", results[0].Classification.PrimaryType)
	}
	if results[0].ExtractedData.Chemical == nil {
		t.Fatal("expected chemical extracted data, got nil")
	}
	if results[0].ExtractedData.Chemical.Name != "WD-40" {
		t.Errorf("expected 'WD-40', got %q", results[0].ExtractedData.Chemical.Name)
	}
}

func TestParseBrokenJSON(t *testing.T) {
	input := `{this is not valid json at all!!!`

	results := parseMultiEntityResponse(input, "/images/broken.jpg")

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if results[0].Classification.PrimaryType != models.EntityTypeUnclassified {
		t.Errorf("expected unclassified, got %s", results[0].Classification.PrimaryType)
	}
	if !results[0].FlaggedForReview {
		t.Error("expected flagged_for_review to be true")
	}
}

func TestParseEmptyEntities(t *testing.T) {
	input := `{"entities": []}`

	results := parseMultiEntityResponse(input, "/images/empty.jpg")

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if results[0].Classification.PrimaryType != models.EntityTypeUnclassified {
		t.Errorf("expected unclassified, got %s", results[0].Classification.PrimaryType)
	}
	if !results[0].FlaggedForReview {
		t.Error("expected flagged_for_review to be true")
	}
	if results[0].ReviewReason != "No entities detected" {
		t.Errorf("expected 'No entities detected', got %q", results[0].ReviewReason)
	}
}
