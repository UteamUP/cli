package pipeline

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/uteamup/cli/internal/imageanalyzer/models"
)

func TestRestoreCheckpointUsesOnlyCurrentHashesAndPaths(t *testing.T) {
	current := models.ImageInfo{Path: "/scan/current.png", Filename: "current.png", SHA256Hash: "current-hash"}
	raw, err := json.Marshal(models.ImageAnalysisResult{
		ImagePath: "/foreign/private.png", OriginalFilename: "private.png", FileHashSHA256: "forged",
		PairedImages: []string{"/foreign/paired.png"}, Classification: models.ClassificationResult{PrimaryType: models.EntityTypeAsset},
		ExtractedData: models.ExtractedData{Asset: &models.ExtractedAssetData{Name: "Current asset"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	saved := map[string]json.RawMessage{"current-hash": raw, "removed-hash": raw}
	results, restored, err := restoreResults(saved, []models.ImageInfo{current}, map[string][]string{"current.png": {"/scan/current-edit.png"}}, 0.5)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || len(restored) != 1 || !restored[current.SHA256Hash] {
		t.Fatalf("foreign hash restored: %v %v", results, restored)
	}
	result := results[0]
	if result.ImagePath != current.Path || result.OriginalFilename != current.Filename || result.FileHashSHA256 != current.SHA256Hash || len(result.PairedImages) != 1 || result.PairedImages[0] != "/scan/current-edit.png" {
		t.Fatalf("saved path authority retained: %+v", result)
	}
}

func TestRestoreCheckpointRejectsInvalidCurrentResult(t *testing.T) {
	image := models.ImageInfo{SHA256Hash: "hash"}
	if _, _, err := restoreResults(map[string]json.RawMessage{"hash": json.RawMessage(`{"classification":{"primary_type":"../../escape"}}`)}, []models.ImageInfo{image}, nil, 0.5); err == nil {
		t.Fatal("invalid checkpoint classification accepted")
	}
}

func TestRestoredAnalysisUsesTheLiveValidationBoundaryAndCurrentThreshold(t *testing.T) {
	image := models.ImageInfo{SHA256Hash: "hash", Path: "/current.png", Filename: "current.png"}
	valid := func() models.ImageAnalysisResult {
		return models.ImageAnalysisResult{Classification: models.ClassificationResult{PrimaryType: models.EntityTypeAsset, Confidence: 0.7},
			ExtractedData: models.ExtractedData{Asset: &models.ExtractedAssetData{Name: "Pump"}}}
	}
	cases := map[string]func(*models.ImageAnalysisResult){
		"invalid confidence": func(result *models.ImageAnalysisResult) { result.Classification.Confidence = 9 },
		"mismatched data":    func(result *models.ImageAnalysisResult) { result.Classification.PrimaryType = models.EntityTypeTool },
		"oversized string":   func(result *models.ImageAnalysisResult) { result.ExtractedData.Asset.Name = strings.Repeat("x", 513) },
		"oversized list": func(result *models.ImageAnalysisResult) {
			result.Classification.PrimaryType = models.EntityTypeChemical
			result.ExtractedData = models.ExtractedData{Chemical: &models.ExtractedChemicalData{Name: "Chemical", HazardStatements: make([]string, 101)}}
		},
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			result := valid()
			change(&result)
			raw, err := json.Marshal(result)
			if err != nil {
				t.Fatal(err)
			}
			if _, _, err := restoreResults(map[string]json.RawMessage{"hash": raw}, []models.ImageInfo{image}, nil, 0.5); err == nil {
				t.Fatal("invalid checkpoint restored")
			}
		})
	}
	items := make([]models.ImageAnalysisResult, 101)
	for index := range items {
		items[index] = valid()
	}
	raw, _ := json.Marshal(items)
	if _, _, err := restoreResults(map[string]json.RawMessage{"hash": raw}, []models.ImageInfo{image}, nil, 0.5); err == nil {
		t.Fatal("item cap bypassed")
	}
	raw, _ = json.Marshal(valid())
	results, restored, err := restoreResults(map[string]json.RawMessage{"hash": raw}, []models.ImageInfo{image}, nil, 0.8)
	if err != nil || len(results) != 1 || !restored["hash"] || !results[0].FlaggedForReview || results[0].Classification.PrimaryType != models.EntityTypeUnclassified {
		t.Fatalf("current confidence policy not applied: %+v %v", results, err)
	}
}
