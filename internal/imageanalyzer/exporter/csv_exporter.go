// Package exporter provides CSV export, image renaming, and summary report
// generation for grouped image analysis results.
package exporter

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/uteamup/cli/internal/imageanalyzer/models"
)

// CSVExporter writes per-entity-type CSV files from grouped analysis results.
type CSVExporter struct {
	outputFolder        string
	renamedImagesFolder string
	renameImages        bool
	renamePattern       string
}

// NewExporter creates a CSVExporter, ensuring output directories exist.
func NewExporter(outputFolder, renamedImagesFolder string, renameImages bool, renamePattern string) *CSVExporter {
	if renamedImagesFolder == "" {
		renamedImagesFolder = outputFolder
	}
	_ = os.MkdirAll(outputFolder, 0o755)
	_ = os.MkdirAll(renamedImagesFolder, 0o755)
	return &CSVExporter{
		outputFolder:        outputFolder,
		renamedImagesFolder: renamedImagesFolder,
		renameImages:        renameImages,
		renamePattern:       renamePattern,
	}
}

// ExportCSVs buckets groups by entity type and writes one CSV per type.
// Unclassified results are wrapped as single-member pseudo-groups.
// Returns a map of entity_type -> csv_path.
func (e *CSVExporter) ExportCSVs(groups []models.ImageGroup, unclassified []models.ImageAnalysisResult) (map[string]string, error) {
	byType := make(map[models.EntityType][]models.ImageGroup)
	for _, g := range groups {
		etype := g.Primary.Classification.PrimaryType
		byType[etype] = append(byType[etype], g)
	}

	// Wrap each unclassified result as a pseudo-group.
	for _, r := range unclassified {
		byType[models.EntityTypeUnclassified] = append(byType[models.EntityTypeUnclassified], models.ImageGroup{
			Primary:        r,
			Members:        nil,
			GroupConfidence: r.Classification.Confidence,
		})
	}

	result := make(map[string]string)
	for etype, etypeGroups := range byType {
		if len(etypeGroups) == 0 {
			continue
		}
		columns, ok := models.CSVColumnsByType[etype]
		if !ok {
			continue
		}
		csvPath := filepath.Join(e.outputFolder, string(etype)+"s.csv")
		if err := e.writeCSV(csvPath, columns, etypeGroups, etype); err != nil {
			return nil, fmt.Errorf("writing %s CSV: %w", etype, err)
		}
		result[string(etype)] = csvPath
	}
	return result, nil
}

// ExportVendorCSV writes a CSV file containing detected vendor data.
// Returns the file path on success.
func (e *CSVExporter) ExportVendorCSV(vendors []models.DetectedVendor) (string, error) {
	csvPath := filepath.Join(e.outputFolder, "vendors.csv")
	f, err := os.Create(csvPath)
	if err != nil {
		return "", fmt.Errorf("creating vendor CSV: %w", err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	if err := w.Write(models.VendorCSVColumns); err != nil {
		return "", err
	}

	for _, v := range vendors {
		record := []string{
			v.Name,
			v.Description,
			v.Email,
			v.Website,
			v.PhoneNumber,
			strings.Join(v.EntityNames, "; "),
			strings.Join(v.EntityTypes, "; "),
			fmt.Sprintf("%d", v.Count),
			strings.Join(v.ImagePaths, "; "),
		}
		if err := w.Write(record); err != nil {
			return "", err
		}
	}

	return csvPath, nil
}

// ExportLocationCSV writes a CSV file containing detected location data.
// Returns the file path on success.
func (e *CSVExporter) ExportLocationCSV(locations []models.DetectedLocation) (string, error) {
	csvPath := filepath.Join(e.outputFolder, "locations.csv")
	f, err := os.Create(csvPath)
	if err != nil {
		return "", fmt.Errorf("creating location CSV: %w", err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	if err := w.Write(models.LocationCSVColumns); err != nil {
		return "", err
	}

	for _, loc := range locations {
		record := []string{
			loc.Name,
			loc.Description,
			loc.Street,
			loc.City,
			loc.State,
			loc.ZipCode,
			loc.PostalCode,
			loc.Country,
			loc.GooglePlaceId,
			loc.FormattedAddress,
			fmt.Sprintf("%f", loc.Latitude),
			fmt.Sprintf("%f", loc.Longitude),
			loc.GoogleMapsUrl,
			loc.Source,
			strings.Join(loc.EntityNames, "; "),
			strings.Join(loc.EntityTypes, "; "),
			fmt.Sprintf("%d", loc.Count),
			strings.Join(loc.ImagePaths, "; "),
		}
		if err := w.Write(record); err != nil {
			return "", err
		}
	}

	return csvPath, nil
}

// writeCSV writes a single CSV file for the given entity type.
func (e *CSVExporter) writeCSV(path string, columns []string, groups []models.ImageGroup, etype models.EntityType) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	if err := w.Write(columns); err != nil {
		return err
	}

	for _, g := range groups {
		row := e.buildRow(g, etype)
		record := make([]string, len(columns))
		for i, col := range columns {
			record[i] = row[col]
		}
		if err := w.Write(record); err != nil {
			return err
		}
	}
	return nil
}

// buildRow constructs a column-name-to-value map for a single group.
func (e *CSVExporter) buildRow(group models.ImageGroup, etype models.EntityType) map[string]string {
	primary := group.Primary
	row := make(map[string]string)

	if etype == models.EntityTypeUnclassified {
		row["original_filename"] = primary.OriginalFilename
		row["image_path"] = primary.ImagePath
		row["confidence_score"] = fmt.Sprintf("%g", primary.Classification.Confidence)
		row["flagged_for_review"] = fmt.Sprintf("%t", primary.FlaggedForReview)
		row["review_reason"] = primary.ReviewReason
		row["classification_reasoning"] = primary.Classification.Reasoning
		row["related_to"] = primary.RelatedTo
		return row
	}

	// Extract data fields via the typed ToMap().
	data := primary.ExtractedData
	switch {
	case data.Asset != nil:
		for k, v := range data.Asset.ToMap() {
			row[k] = v
		}
	case data.Tool != nil:
		for k, v := range data.Tool.ToMap() {
			row[k] = v
		}
	case data.Part != nil:
		for k, v := range data.Part.ToMap() {
			row[k] = v
		}
	case data.Chemical != nil:
		for k, v := range data.Chemical.ToMap() {
			row[k] = v
		}
	}

	// Populate vendor_name from suggested_vendor or manufacturer_brand.
	vendorName := data.GetVendorName()
	row["vendor_name"] = vendorName

	// Populate location_name from suggested_location.
	locationName := data.GetLocationName()
	row["location_name"] = locationName

	row["related_to"] = primary.RelatedTo
	row["image_paths"] = strings.Join(group.AllImagePaths(), "; ")
	row["original_filenames"] = strings.Join(group.AllOriginalFilenames(), "; ")
	row["confidence_score"] = fmt.Sprintf("%g", primary.Classification.Confidence)
	row["flagged_for_review"] = fmt.Sprintf("%t", primary.FlaggedForReview)
	row["review_reason"] = primary.ReviewReason

	return row
}
