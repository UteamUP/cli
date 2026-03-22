package grouper

import (
	"testing"

	"github.com/uteamup/cli/internal/imageanalyzer/models"
)

// --- helpers ---

func strPtr(s string) *string { return &s }

func makeResult(path string, etype models.EntityType, confidence float64, asset *models.ExtractedAssetData, tool *models.ExtractedToolData) models.ImageAnalysisResult {
	r := models.ImageAnalysisResult{
		ImagePath: path,
		Classification: models.ClassificationResult{
			PrimaryType: etype,
			Confidence:  confidence,
		},
	}
	if asset != nil {
		r.ExtractedData.Asset = asset
	}
	if tool != nil {
		r.ExtractedData.Tool = tool
	}
	return r
}

// --- tests ---

func TestPartitionByType(t *testing.T) {
	g := NewGrouper(0.75)

	results := []models.ImageAnalysisResult{
		makeResult("a.jpg", models.EntityTypeAsset, 0.9, &models.ExtractedAssetData{Name: "Pump A"}, nil),
		makeResult("b.jpg", models.EntityTypeTool, 0.8, nil, &models.ExtractedToolData{Name: "Wrench"}),
		makeResult("c.jpg", models.EntityTypeAsset, 0.7, &models.ExtractedAssetData{Name: "Pump B"}, nil),
	}

	groups := g.GroupImages(results)

	// Assets and tools must not be in the same group.
	for _, grp := range groups {
		primaryType := grp.Primary.Classification.PrimaryType
		for _, m := range grp.Members {
			if m.Classification.PrimaryType != primaryType {
				t.Fatalf("mixed entity types in group: primary=%s, member=%s", primaryType, m.Classification.PrimaryType)
			}
		}
	}
}

func TestGroupBySerial(t *testing.T) {
	g := NewGrouper(0.75)

	sn := "SN-12345"
	results := []models.ImageAnalysisResult{
		makeResult("a.jpg", models.EntityTypeAsset, 0.9, &models.ExtractedAssetData{Name: "Pump", SerialNumber: &sn}, nil),
		makeResult("b.jpg", models.EntityTypeAsset, 0.8, &models.ExtractedAssetData{Name: "Pump Photo 2", SerialNumber: &sn}, nil),
		makeResult("c.jpg", models.EntityTypeAsset, 0.7, &models.ExtractedAssetData{Name: "Different Asset"}, nil),
	}

	groups := g.GroupImages(results)

	// The two items with the same serial should be in one group.
	var serialGroup *models.ImageGroup
	for i := range groups {
		if groups[i].Primary.ExtractedData.GetSerialNumber() == sn {
			serialGroup = &groups[i]
			break
		}
	}

	if serialGroup == nil {
		t.Fatal("expected a group with serial number SN-12345")
	}
	// primary + 1 member = 2 items total
	if len(serialGroup.Members) != 1 {
		t.Fatalf("expected 1 member in serial group, got %d", len(serialGroup.Members))
	}
}

func TestGroupByName(t *testing.T) {
	g := NewGrouper(0.75)

	results := []models.ImageAnalysisResult{
		makeResult("a.jpg", models.EntityTypeAsset, 0.9, &models.ExtractedAssetData{Name: "  Hydraulic Pump  "}, nil),
		makeResult("b.jpg", models.EntityTypeAsset, 0.7, &models.ExtractedAssetData{Name: "hydraulic pump"}, nil),
		makeResult("c.jpg", models.EntityTypeAsset, 0.6, &models.ExtractedAssetData{Name: "Compressor"}, nil),
	}

	groups := g.GroupImages(results)

	// The two "Hydraulic Pump" items (case-insensitive, trimmed) should be grouped.
	found := false
	for _, grp := range groups {
		total := 1 + len(grp.Members)
		nameKey := grp.Primary.ExtractedData.GetName()
		if total == 2 && (nameKey == "  Hydraulic Pump  " || nameKey == "hydraulic pump") {
			found = true
		}
	}
	if !found {
		t.Fatal("expected a group of 2 items with name 'hydraulic pump'")
	}
}

func TestSimilaritySameType(t *testing.T) {
	a := makeResult("a.jpg", models.EntityTypeAsset, 0.9, &models.ExtractedAssetData{
		Name:              "CAT 320 Excavator",
		SerialNumber:      strPtr("SN-999"),
		ModelNumber:       strPtr("320F"),
		ManufacturerBrand: strPtr("Caterpillar"),
	}, nil)
	b := makeResult("b.jpg", models.EntityTypeAsset, 0.8, &models.ExtractedAssetData{
		Name:              "CAT 320 Excavator",
		SerialNumber:      strPtr("SN-999"),
		ModelNumber:       strPtr("320F"),
		ManufacturerBrand: strPtr("caterpillar"),
	}, nil)

	sim := computeSimilarity(a, b)
	// serial(0.40) + model(0.20) + name(0.20) + brand(0.05) = 0.85+
	if sim < 0.80 {
		t.Fatalf("expected high similarity, got %f", sim)
	}
}

func TestSimilarityDifferentType(t *testing.T) {
	a := makeResult("a.jpg", models.EntityTypeAsset, 0.9, &models.ExtractedAssetData{Name: "Pump", SerialNumber: strPtr("SN-1")}, nil)
	b := makeResult("b.jpg", models.EntityTypeTool, 0.9, nil, &models.ExtractedToolData{Name: "Pump", SerialNumber: strPtr("SN-1")})

	sim := computeSimilarity(a, b)
	if sim != 0.0 {
		t.Fatalf("expected 0.0 for different types, got %f", sim)
	}
}

func TestRepresentativeSelection(t *testing.T) {
	group := []models.ImageAnalysisResult{
		makeResult("low.jpg", models.EntityTypeAsset, 0.5, &models.ExtractedAssetData{Name: "Low"}, nil),
		makeResult("high.jpg", models.EntityTypeAsset, 0.95, &models.ExtractedAssetData{Name: "High"}, nil),
		makeResult("mid.jpg", models.EntityTypeAsset, 0.7, &models.ExtractedAssetData{Name: "Mid"}, nil),
	}

	rep := selectRepresentative(group)
	if rep.ImagePath != "high.jpg" {
		t.Fatalf("expected high.jpg as representative, got %s", rep.ImagePath)
	}
}

func TestMergeExtractedData(t *testing.T) {
	rep := makeResult("rep.jpg", models.EntityTypeAsset, 0.9, &models.ExtractedAssetData{
		Name: "Pump",
		// Description is nil — should be filled from member.
	}, nil)

	members := []models.ImageAnalysisResult{
		makeResult("m.jpg", models.EntityTypeAsset, 0.7, &models.ExtractedAssetData{
			Name:         "Pump",
			Description:  strPtr("A hydraulic pump"),
			SerialNumber: strPtr("SN-MEMBER"),
		}, nil),
	}

	mergeExtractedData(&rep, members)

	if rep.ExtractedData.GetDescription() != "A hydraulic pump" {
		t.Fatalf("expected description to be filled, got %q", rep.ExtractedData.GetDescription())
	}
	if rep.ExtractedData.GetSerialNumber() != "SN-MEMBER" {
		t.Fatalf("expected serial_number to be filled, got %q", rep.ExtractedData.GetSerialNumber())
	}
}

func TestLevenshteinRatio(t *testing.T) {
	tests := []struct {
		a, b string
		min  float64
		max  float64
	}{
		{"kitten", "kitten", 1.0, 1.0},
		{"kitten", "sitting", 0.5, 0.65},
		{"", "", 1.0, 1.0},
		{"abc", "", 0.0, 0.0},
		{"", "xyz", 0.0, 0.0},
		{"flaw", "lawn", 0.4, 0.6},
	}

	for _, tc := range tests {
		ratio := levenshteinRatio(tc.a, tc.b)
		if ratio < tc.min || ratio > tc.max {
			t.Errorf("levenshteinRatio(%q, %q) = %f, want [%f, %f]", tc.a, tc.b, ratio, tc.min, tc.max)
		}
	}
}

func TestPhashSimilarity(t *testing.T) {
	// Identical hashes.
	if sim := phashSimilarity("abcdef1234567890", "abcdef1234567890"); sim != 1.0 {
		t.Fatalf("identical hashes should be 1.0, got %f", sim)
	}

	// Completely different (all bits differ for small values is unlikely, just check > 0).
	sim := phashSimilarity("0000000000000000", "ffffffffffffffff")
	if sim >= 1.0 || sim < 0.0 {
		t.Fatalf("different hashes should be in [0,1), got %f", sim)
	}

	// Invalid hex.
	if sim := phashSimilarity("not-hex", "also-not"); sim != 0.0 {
		t.Fatalf("invalid hex should return 0.0, got %f", sim)
	}
}

func TestUnclassifiedIsolation(t *testing.T) {
	g := NewGrouper(0.75)

	results := []models.ImageAnalysisResult{
		makeResult("u1.jpg", models.EntityTypeUnclassified, 0.3, nil, nil),
		makeResult("u2.jpg", models.EntityTypeUnclassified, 0.2, nil, nil),
		makeResult("a1.jpg", models.EntityTypeAsset, 0.9, &models.ExtractedAssetData{Name: "Pump"}, nil),
	}

	groups := g.GroupImages(results)

	// Each unclassified should be its own group with no members.
	unclassifiedCount := 0
	for _, grp := range groups {
		if grp.Primary.Classification.PrimaryType == models.EntityTypeUnclassified {
			unclassifiedCount++
			if len(grp.Members) != 0 {
				t.Fatalf("unclassified groups should have no members, got %d", len(grp.Members))
			}
		}
	}
	if unclassifiedCount != 2 {
		t.Fatalf("expected 2 unclassified groups, got %d", unclassifiedCount)
	}
}
