package pipeline

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"

	"github.com/uteamup/cli/internal/imageanalyzer/exporter"
	"github.com/uteamup/cli/internal/imageanalyzer/grouper"
	"github.com/uteamup/cli/internal/imageanalyzer/models"
	"github.com/uteamup/cli/internal/mediaanalyzer"
	vaconfig "github.com/uteamup/cli/internal/videoanalyzer/config"
	"github.com/uteamup/cli/internal/videoanalyzer/fileutil"
	"github.com/uteamup/cli/internal/videoanalyzer/gps"
)

// Pipeline orchestrates the 4-phase video analysis pipeline.
type Pipeline struct {
	config   *vaconfig.AppConfig
	analyzer *mediaanalyzer.Analyzer
}

// NewPipeline creates a new Pipeline with the given configuration.
func NewPipeline(cfg *vaconfig.AppConfig, analyzer *mediaanalyzer.Analyzer) *Pipeline {
	return &Pipeline{config: cfg, analyzer: analyzer}
}

// Run executes the full pipeline: validate → upload+analyze → deduplicate → export.
func (p *Pipeline) Run(ctx context.Context) error {
	startTime := time.Now()

	// ── Phase 1: Validate ──────────────────────────────────────────────
	fmt.Println("\nPhase 1: Scanning for video files...")
	log.Printf("Phase 1: Scanning folder=%s", p.config.Scan.VideoFolder)

	scanner := fileutil.NewScanner(p.config.Scan.MaxFileSizeMB)
	scanResult, err := scanner.ScanPath(p.config.Scan.VideoFolder)
	if err != nil {
		return fmt.Errorf("scanning for videos: %w", err)
	}

	fmt.Printf("  Videos found:  %d\n", len(scanResult.Videos))
	fmt.Printf("  GIFs found:    %d (route to image analyzer)\n", len(scanResult.GIFFiles))
	fmt.Printf("  Skipped:       %d\n", len(scanResult.Skipped))

	if len(scanResult.Videos) == 0 {
		fmt.Printf("\nNo supported video files found in %s\n", p.config.Scan.VideoFolder)
		return nil
	}

	if len(scanResult.GIFFiles) > 0 {
		fmt.Println("\n  GIF files detected — use 'uteamup image analyze' to process them:")
		for _, g := range scanResult.GIFFiles {
			fmt.Printf("    %s\n", g)
		}
	}

	// Dry run: report scope without inventing provider-specific prices.
	if p.config.Processing.DryRun {
		var totalBytes int64
		for _, video := range scanResult.Videos {
			totalBytes += video.SizeBytes
		}
		fmt.Printf("\n=== Dry Run Upload Scope ===\n")
		fmt.Printf("  Videos:        %d\n", len(scanResult.Videos))
		fmt.Printf("  Input size:    %.2f MB\n", float64(totalBytes)/(1024*1024))
		fmt.Printf("  AI route:      Resolved by UteamUP for the authenticated tenant\n")
		fmt.Printf("  Est. cost:     Unavailable without executing the governed route\n")
		fmt.Printf("=============================\n")
		return nil
	}

	// ── Phase 2: Upload + Analyze ──────────────────────────────────────
	fmt.Println("\nPhase 2: Analyzing videos...")

	var allResults []models.ImageAnalysisResult
	var gpsLocations []videoGPSData
	totalCredits := 0
	credentialSource := ""
	modelAlias := ""
	totalVideos := len(scanResult.Videos)
	processedVideos := 0

	// Overall progress bar (0% to 100% across all videos).
	overallBar := progressbar.NewOptions(totalVideos,
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

	for i, video := range scanResult.Videos {
		if err := ctx.Err(); err != nil {
			return err
		}
		// Per-video header.
		fmt.Printf("\n  ── Video %d/%d: %s (%s) ──\n", i+1, totalVideos, video.Filename, formatFileSize(video.SizeBytes))

		// Per-video progress: 4 steps (upload → process → analyze → extract GPS).
		videoBar := progressbar.NewOptions(4,
			progressbar.OptionSetDescription("    Uploading"),
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

		// Step 1: Upload + Process through the authenticated backend route.
		videoBar.Describe("    Uploading & processing")
		analysis, err := p.analyzer.AnalyzeVideo(ctx, video.Path, string(video.MIMEType))
		if err != nil {
			return fmt.Errorf("analyzing video %q: %w", video.Filename, err)
		}
		results := analysis.Items
		processedVideos++
		totalCredits += analysis.Receipt.CreditsCharged
		credentialSource = analysis.Receipt.CredentialSource
		modelAlias = analysis.Receipt.ModelAlias
		_ = videoBar.Add(2) // Upload + process done.

		// Step 2: Parse results.
		videoBar.Describe("    Parsing results")
		allResults = append(allResults, results...)
		_ = videoBar.Add(1)

		// Step 3: Extract GPS.
		videoBar.Describe("    Extracting GPS")
		gpsData, found, gpsErr := gps.ExtractGPS(video.Path)
		if gpsErr != nil {
			log.Printf("Phase 2: GPS extraction error for %s: %v", video.Filename, gpsErr)
		}
		if found {
			gpsLocations = append(gpsLocations, videoGPSData{
				videoPath: video.Path,
				lat:       gpsData.Latitude,
				lng:       gpsData.Longitude,
			})
		}
		_ = videoBar.Add(1)

		// Per-video summary.
		fmt.Printf("    Entities: %d", len(results))
		if found {
			fmt.Printf(" | GPS: %.4f, %.4f", gpsData.Latitude, gpsData.Longitude)
		}
		fmt.Println()

		// Update overall progress.
		_ = overallBar.Add(1)
	}

	fmt.Printf("\n  Videos:    %d processed\n", processedVideos)
	fmt.Printf("  Entities:  %d found\n", len(allResults))
	fmt.Printf("  Credits:   %d charged\n", totalCredits)

	if len(allResults) == 0 {
		fmt.Println("\nNo entities detected in any video.")
		return nil
	}

	// ── Phase 3: Deduplicate ───────────────────────────────────────────
	fmt.Println("\nPhase 3: Deduplicating entities...")

	// Temporal dedup (within same video).
	dedupWindow := p.config.Processing.TemporalDedupWindowSec
	if dedupWindow <= 0 {
		dedupWindow = 30
	}
	deduped := TemporalDedup(allResults, dedupWindow)
	fmt.Printf("  After temporal dedup: %d entities (was %d)\n", len(deduped), len(allResults))

	// Cross-video dedup via grouper.
	g := grouper.NewGrouper(p.config.Processing.GroupingSimilarityThreshold)
	groups := g.GroupImages(deduped)
	fmt.Printf("  After cross-video grouping: %d groups\n", len(groups))

	// Extract vendors.
	vendors := extractVendors(groups)

	// Extract local GPS and AI-suggested locations. The CLI does not call
	// third-party URL or geocoding services directly.
	locations := extractLocations(groups, gpsLocations)

	// Assign locations to entities.
	assignLocations(groups, locations)

	fmt.Printf("  Vendors found:   %d\n", len(vendors))
	fmt.Printf("  Locations found: %d\n\n", len(locations))

	// ── Phase 4: Export ────────────────────────────────────────────────
	fmt.Println("Phase 4: Exporting CSVs...")

	exp := exporter.NewExporter(
		p.config.Scan.OutputFolder,
		p.config.Scan.OutputFolder, // No separate renamed folder for videos.
		false,                      // No renaming for videos.
		"",
	)

	// Separate classified groups from unclassified.
	var classified []models.ImageGroup
	var unclassified []models.ImageAnalysisResult
	for _, group := range groups {
		if group.Primary.Classification.PrimaryType == models.EntityTypeUnclassified {
			unclassified = append(unclassified, group.Primary)
			unclassified = append(unclassified, group.Members...)
		} else {
			classified = append(classified, group)
		}
	}

	csvPaths, err := exp.ExportCSVs(classified, unclassified)
	if err != nil {
		return fmt.Errorf("exporting CSVs: %w", err)
	}

	for entityType, path := range csvPaths {
		fmt.Printf("  %s → %s\n", entityType, path)
	}

	if len(vendors) > 0 {
		vendorPath, err := exp.ExportVendorCSV(vendors)
		if err != nil {
			log.Printf("Phase 4: vendor CSV export error: %v", err)
		} else {
			fmt.Printf("  vendors → %s\n", vendorPath)
		}
	}

	if len(locations) > 0 {
		locationPath, err := exp.ExportLocationCSV(locations)
		if err != nil {
			log.Printf("Phase 4: location CSV export error: %v", err)
		} else {
			fmt.Printf("  locations → %s\n", locationPath)
		}
	}

	// Print summary.
	elapsed := time.Since(startTime)
	fmt.Printf("\n=== Video Analysis Complete ===\n")
	fmt.Printf("  Videos processed:  %d\n", processedVideos)
	fmt.Printf("  Entities found:    %d\n", len(deduped))
	fmt.Printf("  Groups:            %d\n", len(groups))
	fmt.Printf("  Credits charged:   %d\n", totalCredits)
	if credentialSource != "" {
		fmt.Printf("  AI source:         %s", credentialSource)
		if modelAlias != "" {
			fmt.Printf(" · %s", modelAlias)
		}
		fmt.Println()
	}
	fmt.Printf("  Elapsed:           %s\n", elapsed.Round(time.Second))
	fmt.Printf("===============================\n")

	return nil
}

// videoGPSData holds GPS coordinates extracted from a video file.
type videoGPSData struct {
	videoPath string
	lat, lng  float64
}

// extractVendors aggregates unique vendors from grouped results.
func extractVendors(groups []models.ImageGroup) []models.DetectedVendor {
	vendorMap := make(map[string]*models.DetectedVendor)

	for _, g := range groups {
		allResults := append([]models.ImageAnalysisResult{g.Primary}, g.Members...)
		for _, r := range allResults {
			vn := r.ExtractedData.GetVendorName()
			if vn == "" {
				continue
			}
			entityName := r.ExtractedData.GetName()
			entityType := string(r.Classification.PrimaryType)
			imagePaths := g.AllImagePaths()

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

// extractLocations creates DetectedLocation entries from GPS data and AI-suggested locations.
func extractLocations(groups []models.ImageGroup, gpsData []videoGPSData) []models.DetectedLocation {
	var locations []models.DetectedLocation

	// Build GPS lookup by video path.
	gpsMap := make(map[string]videoGPSData)
	for _, g := range gpsData {
		gpsMap[g.videoPath] = g
	}

	// Create GPS-based locations.
	for _, g := range groups {
		entityName := g.Primary.ExtractedData.GetName()
		entityType := string(g.Primary.Classification.PrimaryType)

		for _, path := range g.AllImagePaths() {
			if gd, ok := gpsMap[path]; ok {
				locations = append(locations, models.DetectedLocation{
					Latitude:    gd.lat,
					Longitude:   gd.lng,
					HasGPS:      true,
					Source:      "gps_video_metadata",
					EntityNames: []string{entityName},
					EntityTypes: []string{entityType},
					ImagePaths:  []string{path},
					Count:       1,
				})
			}
		}
	}

	// Add AI-suggested locations.
	for _, g := range groups {
		loc := g.Primary.ExtractedData.GetLocationName()
		if loc != "" {
			entityName := g.Primary.ExtractedData.GetName()
			entityType := string(g.Primary.Classification.PrimaryType)
			locations = append(locations, models.DetectedLocation{
				Name:        loc,
				Source:      "ai_suggested",
				EntityNames: []string{entityName},
				EntityTypes: []string{entityType},
				ImagePaths:  g.AllImagePaths(),
				Count:       1,
			})
		}
	}

	return locations
}

// assignLocations assigns location names to entities in groups based on available location data.
func assignLocations(groups []models.ImageGroup, locations []models.DetectedLocation) {
	// Build a map of video path to location name for GPS locations.
	pathToLocation := make(map[string]string)
	for _, loc := range locations {
		name := loc.Name
		if name == "" {
			name = loc.FormattedAddress
		}
		if name == "" {
			name = fmt.Sprintf("%.4f, %.4f", loc.Latitude, loc.Longitude)
		}
		for _, p := range loc.ImagePaths {
			if _, exists := pathToLocation[p]; !exists {
				pathToLocation[p] = name
			}
		}
	}

	// Assign location to primary results in groups where applicable.
	for i := range groups {
		if locName, ok := pathToLocation[groups[i].Primary.ImagePath]; ok {
			if groups[i].Primary.EXIFMetadata == nil {
				groups[i].Primary.EXIFMetadata = make(map[string]interface{})
			}
			groups[i].Primary.EXIFMetadata["assigned_location"] = locName
		}
	}
}

// appendUnique appends items to a slice only if they're not already present.
func appendUnique(slice []string, items ...string) []string {
	existing := make(map[string]bool, len(slice))
	for _, s := range slice {
		existing[s] = true
	}
	for _, item := range items {
		if !existing[item] {
			slice = append(slice, item)
			existing[item] = true
		}
	}
	return slice
}

// formatFileSize returns a human-readable file size string.
func formatFileSize(bytes int64) string {
	const (
		kb = 1024
		mb = kb * 1024
		gb = mb * 1024
	)
	switch {
	case bytes >= gb:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(gb))
	case bytes >= mb:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(mb))
	case bytes >= kb:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(kb))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
