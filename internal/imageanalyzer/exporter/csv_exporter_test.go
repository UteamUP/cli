package exporter

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/uteamup/cli/internal/imageanalyzer/models"
)

// helper to create a string pointer.
func strPtr(s string) *string { return &s }

func makeAssetGroup(name string, confidence float64, flagged bool) models.ImageGroup {
	return models.ImageGroup{
		Primary: models.ImageAnalysisResult{
			ImagePath:        "/tmp/img1.jpg",
			OriginalFilename: "img1.jpg",
			Classification: models.ClassificationResult{
				PrimaryType: models.EntityTypeAsset,
				Confidence:  confidence,
				Reasoning:   "looks like an asset",
			},
			ExtractedData: models.ExtractedData{
				Asset: &models.ExtractedAssetData{
					Name:              name,
					Description:       strPtr("A test asset"),
					SerialNumber:      strPtr("SN123"),
					ManufacturerBrand: strPtr("BrandX"),
				},
			},
			FlaggedForReview: flagged,
			ReviewReason:     "",
			RelatedTo:        "parent-asset",
		},
		Members:         nil,
		GroupConfidence: confidence,
	}
}

func TestExportCSVsAssets(t *testing.T) {
	dir := t.TempDir()
	exp := NewExporter(dir, "", false, "")

	groups := []models.ImageGroup{makeAssetGroup("Pump A", 0.95, false)}

	result, err := exp.ExportCSVs(groups, nil)
	if err != nil {
		t.Fatalf("ExportCSVs: %v", err)
	}
	csvPath, ok := result["asset"]
	if !ok {
		t.Fatal("expected 'asset' key in result")
	}

	f, err := os.Open(csvPath)
	if err != nil {
		t.Fatalf("open CSV: %v", err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("read CSV: %v", err)
	}

	// Header + 1 data row.
	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}

	header := records[0]
	// Verify columns match AssetCSVColumns.
	if len(header) != len(models.AssetCSVColumns) {
		t.Fatalf("header length %d != AssetCSVColumns %d", len(header), len(models.AssetCSVColumns))
	}
	for i, col := range models.AssetCSVColumns {
		if header[i] != col {
			t.Errorf("header[%d] = %q, want %q", i, header[i], col)
		}
	}

	// Check name value in the data row.
	row := records[1]
	if row[0] != "Pump A" {
		t.Errorf("name = %q, want %q", row[0], "Pump A")
	}
}

func TestExportCSVsChemicals(t *testing.T) {
	dir := t.TempDir()
	exp := NewExporter(dir, "", false, "")

	group := models.ImageGroup{
		Primary: models.ImageAnalysisResult{
			ImagePath:        "/tmp/chem.jpg",
			OriginalFilename: "chem.jpg",
			Classification: models.ClassificationResult{
				PrimaryType: models.EntityTypeChemical,
				Confidence:  0.88,
			},
			ExtractedData: models.ExtractedData{
				Chemical: &models.ExtractedChemicalData{
					Name:                    "Acetone",
					HazardStatements:        []string{"H225", "H319", "H336"},
					PrecautionaryStatements: []string{"P210", "P261"},
					UnitOfMeasure:           "L",
				},
			},
			RelatedTo: "",
		},
		GroupConfidence: 0.88,
	}

	result, err := exp.ExportCSVs([]models.ImageGroup{group}, nil)
	if err != nil {
		t.Fatalf("ExportCSVs: %v", err)
	}
	csvPath := result["chemical"]

	f, err := os.Open(csvPath)
	if err != nil {
		t.Fatalf("open CSV: %v", err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("read CSV: %v", err)
	}

	header := records[0]
	row := records[1]

	// Find hazard_statements column index.
	hsIdx := -1
	psIdx := -1
	for i, col := range header {
		if col == "hazard_statements" {
			hsIdx = i
		}
		if col == "precautionary_statements" {
			psIdx = i
		}
	}
	if hsIdx < 0 || psIdx < 0 {
		t.Fatal("hazard_statements or precautionary_statements column not found")
	}

	if row[hsIdx] != "H225; H319; H336" {
		t.Errorf("hazard_statements = %q, want %q", row[hsIdx], "H225; H319; H336")
	}
	if row[psIdx] != "P210; P261" {
		t.Errorf("precautionary_statements = %q, want %q", row[psIdx], "P210; P261")
	}
}

func TestExportCSVsUnclassified(t *testing.T) {
	dir := t.TempDir()
	exp := NewExporter(dir, "", false, "")

	unclassified := []models.ImageAnalysisResult{
		{
			ImagePath:        "/tmp/mystery.png",
			OriginalFilename: "mystery.png",
			Classification: models.ClassificationResult{
				PrimaryType: models.EntityTypeUnclassified,
				Confidence:  0.3,
				Reasoning:   "unclear image",
			},
			FlaggedForReview: true,
			ReviewReason:     "low confidence",
			RelatedTo:        "",
		},
	}

	result, err := exp.ExportCSVs(nil, unclassified)
	if err != nil {
		t.Fatalf("ExportCSVs: %v", err)
	}
	csvPath, ok := result["unclassified"]
	if !ok {
		t.Fatal("expected 'unclassified' key")
	}

	f, err := os.Open(csvPath)
	if err != nil {
		t.Fatalf("open CSV: %v", err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("read CSV: %v", err)
	}

	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}

	header := records[0]
	row := records[1]

	// Verify unclassified columns.
	for i, col := range models.UnclassifiedCSVColumns {
		if header[i] != col {
			t.Errorf("header[%d] = %q, want %q", i, header[i], col)
		}
	}

	// Check original_filename.
	if row[0] != "mystery.png" {
		t.Errorf("original_filename = %q, want %q", row[0], "mystery.png")
	}
	// Check flagged_for_review.
	flaggedIdx := -1
	for i, col := range header {
		if col == "flagged_for_review" {
			flaggedIdx = i
		}
	}
	if flaggedIdx >= 0 && row[flaggedIdx] != "true" {
		t.Errorf("flagged_for_review = %q, want %q", row[flaggedIdx], "true")
	}
}

func TestRenameImages(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	// Create a temp source image.
	srcFile := filepath.Join(srcDir, "photo.jpg")
	if err := os.WriteFile(srcFile, []byte("fake-image-data"), 0o644); err != nil {
		t.Fatal(err)
	}

	exp := NewExporter(t.TempDir(), dstDir, true, "")

	group := models.ImageGroup{
		Primary: models.ImageAnalysisResult{
			ImagePath:        srcFile,
			OriginalFilename: "photo.jpg",
			Classification: models.ClassificationResult{
				PrimaryType: models.EntityTypeAsset,
				Confidence:  0.9,
			},
			ExtractedData: models.ExtractedData{
				Asset: &models.ExtractedAssetData{
					Name: "Test Pump",
				},
			},
		},
		GroupConfidence: 0.9,
	}

	mapping, err := exp.RenameImages([]models.ImageGroup{group})
	if err != nil {
		t.Fatalf("RenameImages: %v", err)
	}

	if len(mapping) != 1 {
		t.Fatalf("expected 1 mapping, got %d", len(mapping))
	}

	for _, newPath := range mapping {
		// Verify file exists.
		if _, err := os.Stat(newPath); err != nil {
			t.Errorf("renamed file not found: %v", err)
		}
		base := filepath.Base(newPath)
		if !strings.HasPrefix(base, "asset_test_pump_001_") {
			t.Errorf("unexpected filename: %s", base)
		}
		if !strings.HasSuffix(base, ".jpg") {
			t.Errorf("unexpected extension: %s", base)
		}
	}
}

func TestEmptyGroups(t *testing.T) {
	dir := t.TempDir()
	exp := NewExporter(dir, "", false, "")

	result, err := exp.ExportCSVs(nil, nil)
	if err != nil {
		t.Fatalf("ExportCSVs: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty result, got %d entries", len(result))
	}
}

func TestSummaryReport(t *testing.T) {
	dir := t.TempDir()
	exp := NewExporter(dir, "", false, "")

	groups := []models.ImageGroup{makeAssetGroup("Pump", 0.9, true)}
	unclassified := []models.ImageAnalysisResult{
		{
			ImagePath:        "/tmp/x.jpg",
			OriginalFilename: "x.jpg",
			Classification: models.ClassificationResult{
				PrimaryType: models.EntityTypeUnclassified,
				Confidence:  0.2,
			},
			FlaggedForReview: false,
		},
	}

	report, err := exp.GenerateSummaryReport(groups, unclassified, 120.0, 3)
	if err != nil {
		t.Fatalf("GenerateSummaryReport: %v", err)
	}

	if !strings.Contains(report, "# Image Analysis Summary Report") {
		t.Error("missing report title")
	}
	if !strings.Contains(report, "Flagged for review | 1") {
		t.Error("missing flagged count")
	}
	if !strings.Contains(report, "Duplicates found | 3") {
		t.Error("missing duplicates count")
	}
	if !strings.Contains(report, "2.0 min") {
		t.Error("missing duration")
	}
	if !strings.Contains(report, "UteamUP Image Analyzer") {
		t.Error("missing generator attribution")
	}

	// Verify file was written.
	reportPath := filepath.Join(dir, "summary_report.md")
	if _, err := os.Stat(reportPath); err != nil {
		t.Errorf("report file not created: %v", err)
	}
}
