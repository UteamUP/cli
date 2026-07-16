// Package pipeline orchestrates the 4-phase image analysis pipeline:
// scan, analyze, group, and export.
package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"

	"github.com/uteamup/cli/internal/imageanalyzer/checkpoint"
	iaconfig "github.com/uteamup/cli/internal/imageanalyzer/config"
	"github.com/uteamup/cli/internal/imageanalyzer/exporter"
	"github.com/uteamup/cli/internal/imageanalyzer/grouper"
	"github.com/uteamup/cli/internal/imageanalyzer/imageutil"
	"github.com/uteamup/cli/internal/imageanalyzer/models"
	"github.com/uteamup/cli/internal/imageanalyzer/scanner"
	"github.com/uteamup/cli/internal/mediaanalyzer"
)

// Pipeline orchestrates the 4-phase image analysis pipeline.
type Pipeline struct {
	config   *iaconfig.AppConfig
	analyzer *mediaanalyzer.Analyzer
}

// NewPipeline creates a new Pipeline with the given configuration.
func NewPipeline(cfg *iaconfig.AppConfig, analyzer *mediaanalyzer.Analyzer) *Pipeline {
	return &Pipeline{config: cfg, analyzer: analyzer}
}

// Run executes the full pipeline: scan -> analyze -> group -> export.
func (p *Pipeline) Run(ctx context.Context) error {
	startTime := time.Now()

	// ── Phase 1: Scan ──────────────────────────────────────────────────
	fmt.Println("\nPhase 1: Scanning images...")
	log.Printf("Phase 1: Scanning images folder=%s", p.config.Scan.ImageFolder)

	sc := scanner.NewScanner(
		p.config.Scan.ImageFolder,
		p.config.Scan.Recursive,
		p.config.Scan.SupportedFormats,
		p.config.Scan.MaxImageDimension,
		p.config.Scan.MaxFileSizeMB,
	)

	allImages, err := sc.ScanFolder()
	if err != nil {
		return fmt.Errorf("scanning folder: %w", err)
	}

	if len(allImages) == 0 {
		fmt.Println("No images found in folder:", p.config.Scan.ImageFolder)
		return nil
	}

	// Detect duplicates.
	uniqueImages, duplicatePairs := scanner.DetectDuplicates(allImages)
	duplicatesFound := len(duplicatePairs)

	// Detect iPhone edit pairs.
	editPairs := scanner.DetectIPhonePairs(uniqueImages)

	// Build set of images to analyze (skip edit variants).
	editVariantPaths := make(map[string]bool)
	for _, variants := range editPairs {
		for _, v := range variants {
			editVariantPaths[v] = true
		}
	}
	var imagesToAnalyze []models.ImageInfo
	for _, img := range uniqueImages {
		if !editVariantPaths[img.Path] {
			imagesToAnalyze = append(imagesToAnalyze, img)
		}
	}

	fmt.Printf("\n  Scan Summary:\n")
	fmt.Printf("    Total found:    %d\n", len(allImages))
	fmt.Printf("    Unique:         %d\n", len(uniqueImages))
	fmt.Printf("    To analyze:     %d\n", len(imagesToAnalyze))
	fmt.Printf("    Duplicates:     %d\n", duplicatesFound)
	fmt.Printf("    Edit pairs:     %d\n\n", len(editPairs))

	// Report GPS stats.
	gpsCount := 0
	for _, img := range imagesToAnalyze {
		if img.HasGPS {
			gpsCount++
		}
	}
	if gpsCount > 0 {
		fmt.Printf("    With GPS data:  %d\n\n", gpsCount)
	}

	// Dry-run: estimate cost and stop.
	if p.config.Processing.DryRun {
		p.printDryRun(imagesToAnalyze, duplicatesFound, editPairs)
		return nil
	}

	// ── Phase 2: Analyze ───────────────────────────────────────────────
	fmt.Println("Phase 2: Analyzing images through UteamUP...")
	log.Printf("Phase 2: Analyzing images through tenant-scoped backend routing")

	cp, err := checkpoint.Load(p.config.Processing.CheckpointFile)
	if err != nil {
		return fmt.Errorf("loading checkpoint: %w", err)
	}

	if err := cp.AcquireLock(); err != nil {
		return fmt.Errorf("checkpoint lock: %w", err)
	}
	defer cp.ReleaseLock()

	// Restore previously processed results.
	var results []models.ImageAnalysisResult
	for _, raw := range cp.GetResults() {
		// Handle both single result and array-of-results checkpoint formats.
		var arr []models.ImageAnalysisResult
		if json.Unmarshal(raw, &arr) == nil {
			results = append(results, arr...)
			continue
		}
		var single models.ImageAnalysisResult
		if json.Unmarshal(raw, &single) == nil {
			results = append(results, single)
		}
	}

	totalImages := len(imagesToAnalyze)

	// Overall progress bar (0% to 100% across all images).
	overallBar := progressbar.NewOptions(totalImages,
		progressbar.OptionSetDescription("  Overall progress"),
		progressbar.OptionSetWidth(40),
		progressbar.OptionShowCount(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "█",
			SaucerHead:    "█",
			SaucerPadding: "░",
			BarStart:      "[",
			BarEnd:        "]",
		}),
		progressbar.OptionOnCompletion(func() { fmt.Println() }),
	)

	totalCredits := 0
	credentialSource := ""
	modelAlias := ""
	for i, imageInfo := range imagesToAnalyze {
		if err := ctx.Err(); err != nil {
			return err
		}
		// Skip already processed.
		if cp.IsProcessed(imageInfo.SHA256Hash) {
			_ = overallBar.Add(1)
			continue
		}

		// Per-image header.
		fmt.Printf("\n  ── Image %d/%d: %s ──\n", i+1, totalImages, imageInfo.Filename)

		// Per-image progress: 3 steps (load → analyze → save).
		imgBar := progressbar.NewOptions(3,
			progressbar.OptionSetDescription("    Loading"),
			progressbar.OptionSetWidth(30),
			progressbar.OptionSetTheme(progressbar.Theme{
				Saucer:        "▓",
				SaucerHead:    "▓",
				SaucerPadding: "░",
				BarStart:      "[",
				BarEnd:        "]",
			}),
			progressbar.OptionOnCompletion(func() { fmt.Println() }),
		)

		// Step 1: Load image.
		imgBar.Describe("    Loading image")
		imgBytes, err := imageutil.LoadImageBytes(
			imageInfo.Path,
			p.config.Scan.MaxImageDimension,
		)
		if err != nil {
			log.Printf("Failed to load image %s: %v", imageInfo.Path, err)
			failResult := createFailResult(imageInfo, fmt.Sprintf("Load error: %v", err))
			results = append(results, failResult)
			raw, _ := json.Marshal(failResult)
			_ = cp.AddResult(imageInfo.SHA256Hash, raw)
			_ = imgBar.Add(3)
			_ = overallBar.Add(1)
			continue
		}
		_ = imgBar.Add(1)

		// Step 2: Analyze through the authenticated backend task route.
		imgBar.Describe("    Analyzing")
		analysis, err := p.analyzer.AnalyzePhoto(ctx, imageInfo.Path, imgBytes)
		if err != nil {
			return fmt.Errorf("analyzing image %q: %w", imageInfo.Filename, err)
		}
		imageResults := analysis.Items
		totalCredits += analysis.Receipt.CreditsCharged
		credentialSource = analysis.Receipt.CredentialSource
		modelAlias = analysis.Receipt.ModelAlias
		_ = imgBar.Add(1)

		// Step 3: Process results.
		imgBar.Describe("    Saving results")
		for j := range imageResults {
			// Attach iPhone edit pair paths.
			if pairs, ok := editPairs[imageInfo.Filename]; ok {
				imageResults[j].PairedImages = pairs
			}

			// Apply confidence threshold.
			if imageResults[j].Classification.Confidence > 0 &&
				imageResults[j].Classification.Confidence < p.config.Processing.ConfidenceThreshold {
				imageResults[j].Classification.PrimaryType = models.EntityTypeUnclassified
				imageResults[j].FlaggedForReview = true
				imageResults[j].ReviewReason = fmt.Sprintf(
					"Low confidence: %.2f", imageResults[j].Classification.Confidence,
				)
			}
		}

		results = append(results, imageResults...)

		// Checkpoint all results for this image.
		raw, _ := json.Marshal(imageResults)
		_ = cp.AddResult(imageInfo.SHA256Hash, raw)
		_ = imgBar.Add(1)

		// Per-image summary.
		fmt.Printf("    Entities: %d\n", len(imageResults))

		// Update overall progress.
		_ = overallBar.Add(1)
	}

	fmt.Printf("\n  Analysis complete: %d results\n\n", len(results))

	// ── Phase 3: Group ─────────────────────────────────────────────────
	fmt.Println("Phase 3: Grouping images...")
	log.Printf("Phase 3: Grouping images")

	var classified, unclassified []models.ImageAnalysisResult
	for _, r := range results {
		if r.Classification.PrimaryType == models.EntityTypeUnclassified {
			unclassified = append(unclassified, r)
		} else {
			classified = append(classified, r)
		}
	}

	imgGrouper := grouper.NewGrouper(p.config.Processing.GroupingSimilarityThreshold)
	groups := imgGrouper.GroupImages(classified)

	fmt.Printf("  Groups formed:  %d\n", len(groups))
	fmt.Printf("  Unclassified:   %d\n\n", len(unclassified))

	// ── Phase 3b: Extract vendors and locations ────────────────────────
	fmt.Println("Phase 3b: Extracting vendors and locations...")
	log.Printf("Phase 3b: Extracting vendors and locations")

	vendors := extractVendors(groups)
	locations := extractLocations(groups, allImages)

	fmt.Printf("  Vendors found:   %d\n", len(vendors))
	fmt.Printf("  Locations found: %d\n\n", len(locations))

	// ── Phase 4: Export ────────────────────────────────────────────────
	fmt.Println("Phase 4: Exporting CSVs...")
	log.Printf("Phase 4: Exporting CSVs")

	exp := exporter.NewExporter(
		p.config.Scan.OutputFolder,
		p.config.Scan.RenamedImagesFolder,
		p.config.Processing.RenameImages,
		p.config.Processing.RenamePattern,
	)

	csvFiles, err := exp.ExportCSVs(groups, unclassified)
	if err != nil {
		return fmt.Errorf("exporting CSVs: %w", err)
	}
	for entityType, csvPath := range csvFiles {
		fmt.Printf("  CSV written: %s -> %s\n", entityType, csvPath)
	}

	if len(vendors) > 0 {
		vendorPath, err := exp.ExportVendorCSV(vendors)
		if err != nil {
			log.Printf("Warning: failed to export vendor CSV: %v", err)
		} else {
			fmt.Printf("  CSV written: vendors -> %s\n", vendorPath)
		}
	}

	if len(locations) > 0 {
		locationPath, err := exp.ExportLocationCSV(locations)
		if err != nil {
			log.Printf("Warning: failed to export location CSV: %v", err)
		} else {
			fmt.Printf("  CSV written: locations -> %s\n", locationPath)
		}
	}

	if p.config.Processing.RenameImages {
		renameMap, err := exp.RenameImages(groups)
		if err != nil {
			return fmt.Errorf("renaming images: %w", err)
		}
		fmt.Printf("  Images renamed: %d\n", len(renameMap))
	}

	duration := time.Since(startTime).Seconds()
	_, err = exp.GenerateSummaryReport(groups, unclassified, duration, duplicatesFound)
	if err != nil {
		return fmt.Errorf("generating summary report: %w", err)
	}

	// Clean up checkpoint on success.
	_ = cp.Delete()

	fmt.Printf("\n=== Pipeline complete in %.1fs ===\n", duration)
	if credentialSource != "" {
		fmt.Printf("  AI source: %s", credentialSource)
		if modelAlias != "" {
			fmt.Printf(" · %s", modelAlias)
		}
		fmt.Println()
	}
	fmt.Printf("  Credits charged: %d\n", totalCredits)
	return nil
}

// extractVendors iterates all groups and extracts unique vendor names from
// suggested_vendor and manufacturer_brand fields.
func extractVendors(groups []models.ImageGroup) []models.DetectedVendor {
	vendorMap := make(map[string]*models.DetectedVendor) // key: lowercase vendor name

	for _, g := range groups {
		data := g.Primary.ExtractedData
		entityName := data.GetName()
		entityType := string(g.Primary.Classification.PrimaryType)
		imagePaths := g.AllImagePaths()

		// Collect vendor names from suggested_vendor and manufacturer_brand.
		var vendorNames []string
		vendorName := data.GetVendorName()
		if vendorName != "" {
			vendorNames = append(vendorNames, vendorName)
		}
		brand := data.GetBrand()
		if brand != "" && !strings.EqualFold(brand, vendorName) {
			vendorNames = append(vendorNames, brand)
		}

		for _, vn := range vendorNames {
			key := strings.ToLower(strings.TrimSpace(vn))
			if key == "" {
				continue
			}
			if existing, ok := vendorMap[key]; ok {
				existing.Count++
				existing.EntityNames = appendUnique(existing.EntityNames, entityName)
				existing.EntityTypes = appendUnique(existing.EntityTypes, entityType)
				existing.ImagePaths = appendUnique(existing.ImagePaths, imagePaths...)
			} else {
				vendorMap[key] = &models.DetectedVendor{
					Name:        vn,
					EntityNames: []string{entityName},
					EntityTypes: []string{entityType},
					ImagePaths:  imagePaths,
					Count:       1,
				}
			}
		}
	}

	vendors := make([]models.DetectedVendor, 0, len(vendorMap))
	for _, v := range vendorMap {
		vendors = append(vendors, *v)
	}
	return vendors
}

// extractLocations extracts unique locations from GPS data and suggested_location fields.
// Nearby GPS coordinates (within ~100m) are clustered as the same location.
func extractLocations(groups []models.ImageGroup, allImages []models.ImageInfo) []models.DetectedLocation {
	var locations []models.DetectedLocation

	// Build a map of image path -> ImageInfo for GPS lookup.
	imageMap := make(map[string]models.ImageInfo, len(allImages))
	for _, img := range allImages {
		imageMap[img.Path] = img
	}

	// Extract GPS-based locations from images.
	type gpsCluster struct {
		lat, lng    float64
		entityNames []string
		entityTypes []string
		imagePaths  []string
		count       int
	}
	var gpsClusters []gpsCluster

	for _, g := range groups {
		entityName := g.Primary.ExtractedData.GetName()
		entityType := string(g.Primary.Classification.PrimaryType)

		for _, imgPath := range g.AllImagePaths() {
			img, ok := imageMap[imgPath]
			if !ok || !img.HasGPS {
				continue
			}

			// Find existing cluster within ~100m.
			found := false
			for i := range gpsClusters {
				if haversineDistance(gpsClusters[i].lat, gpsClusters[i].lng, img.GPSLatitude, img.GPSLongitude) < 100 {
					gpsClusters[i].count++
					gpsClusters[i].entityNames = appendUnique(gpsClusters[i].entityNames, entityName)
					gpsClusters[i].entityTypes = appendUnique(gpsClusters[i].entityTypes, entityType)
					gpsClusters[i].imagePaths = appendUnique(gpsClusters[i].imagePaths, imgPath)
					found = true
					break
				}
			}
			if !found {
				gpsClusters = append(gpsClusters, gpsCluster{
					lat:         img.GPSLatitude,
					lng:         img.GPSLongitude,
					entityNames: []string{entityName},
					entityTypes: []string{entityType},
					imagePaths:  []string{imgPath},
					count:       1,
				})
			}
		}
	}

	// Convert GPS clusters to DetectedLocations.
	for _, c := range gpsClusters {
		locations = append(locations, models.DetectedLocation{
			Latitude:    c.lat,
			Longitude:   c.lng,
			HasGPS:      true,
			Source:      "gps_exif",
			EntityNames: c.entityNames,
			EntityTypes: c.entityTypes,
			ImagePaths:  c.imagePaths,
			Count:       c.count,
		})
	}

	// Extract suggested locations from governed backend analysis (for entities without GPS).
	suggestedMap := make(map[string]*models.DetectedLocation) // key: lowercase location name
	for _, g := range groups {
		data := g.Primary.ExtractedData
		locName := data.GetLocationName()
		if locName == "" {
			continue
		}

		entityName := data.GetName()
		entityType := string(g.Primary.Classification.PrimaryType)
		imagePaths := g.AllImagePaths()

		key := strings.ToLower(strings.TrimSpace(locName))
		if existing, ok := suggestedMap[key]; ok {
			existing.Count++
			existing.EntityNames = appendUnique(existing.EntityNames, entityName)
			existing.EntityTypes = appendUnique(existing.EntityTypes, entityType)
			existing.ImagePaths = appendUnique(existing.ImagePaths, imagePaths...)
		} else {
			suggestedMap[key] = &models.DetectedLocation{
				Name:        locName,
				Source:      "ai_suggested",
				EntityNames: []string{entityName},
				EntityTypes: []string{entityType},
				ImagePaths:  imagePaths,
				Count:       1,
			}
		}
	}

	for _, loc := range suggestedMap {
		// Check if this suggested location overlaps with a GPS location by name.
		duplicate := false
		for i := range locations {
			if strings.EqualFold(locations[i].Name, loc.Name) {
				// Merge into existing GPS location.
				locations[i].EntityNames = appendUnique(locations[i].EntityNames, loc.EntityNames...)
				locations[i].EntityTypes = appendUnique(locations[i].EntityTypes, loc.EntityTypes...)
				locations[i].ImagePaths = appendUnique(locations[i].ImagePaths, loc.ImagePaths...)
				locations[i].Count += loc.Count
				duplicate = true
				break
			}
		}
		if !duplicate {
			locations = append(locations, *loc)
		}
	}

	return locations
}

// haversineDistance returns the distance in meters between two GPS coordinates.
func haversineDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadiusM = 6371000.0
	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusM * c
}

// appendUnique appends values to a slice, skipping duplicates.
func appendUnique(slice []string, values ...string) []string {
	seen := make(map[string]bool, len(slice))
	for _, s := range slice {
		seen[s] = true
	}
	for _, v := range values {
		if !seen[v] {
			slice = append(slice, v)
			seen[v] = true
		}
	}
	return slice
}

// printDryRun reports the deterministic upload scope without making AI calls.
// Pricing is route-specific and therefore cannot be truthfully estimated by
// the CLI without asking the backend.
func (p *Pipeline) printDryRun(images []models.ImageInfo, duplicatesFound int, editPairs map[string][]string) {
	var totalBytes int64
	for _, image := range images {
		totalBytes += image.FileSizeBytes
	}

	fmt.Println("\n=== DRY RUN — Upload Scope ===")
	fmt.Printf("Images to analyze:  %d\n", len(images))
	fmt.Printf("Duplicates skipped: %d\n", duplicatesFound)
	fmt.Printf("iPhone edit pairs:  %d (variants skipped)\n", len(editPairs))
	fmt.Printf("Input size:         %.2f MB\n", float64(totalBytes)/(1024*1024))
	fmt.Println("AI route:           Resolved by UteamUP for the authenticated tenant")
	fmt.Println("Estimated cost:     Unavailable without executing the governed route")
	fmt.Println("================================")
}

// createFailResult builds an unclassified result for images that failed to load or analyze.
func createFailResult(img models.ImageInfo, reason string) models.ImageAnalysisResult {
	return models.ImageAnalysisResult{
		ImagePath:        img.Path,
		OriginalFilename: img.Filename,
		FileHashSHA256:   img.SHA256Hash,
		PerceptualHash:   img.PerceptualHash,
		Classification: models.ClassificationResult{
			PrimaryType: models.EntityTypeUnclassified,
			Confidence:  0.0,
			Reasoning:   reason,
		},
		FlaggedForReview: true,
		ReviewReason:     reason,
		ProcessedAt:      time.Now(),
	}
}
