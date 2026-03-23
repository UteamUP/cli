package pipeline

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"

	"github.com/uteamup/cli/internal/imageanalyzer/exporter"
	"github.com/uteamup/cli/internal/imageanalyzer/geocoder"
	"github.com/uteamup/cli/internal/imageanalyzer/grouper"
	"github.com/uteamup/cli/internal/imageanalyzer/models"
	"github.com/uteamup/cli/internal/videoanalyzer/analyzer"
	vaconfig "github.com/uteamup/cli/internal/videoanalyzer/config"
	"github.com/uteamup/cli/internal/videoanalyzer/fileutil"
	"github.com/uteamup/cli/internal/videoanalyzer/gps"
	"github.com/uteamup/cli/internal/videoanalyzer/vendor"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// Pipeline orchestrates the 4-phase video analysis pipeline.
type Pipeline struct {
	config *vaconfig.AppConfig
}

// NewPipeline creates a new Pipeline with the given configuration.
func NewPipeline(cfg *vaconfig.AppConfig) *Pipeline {
	return &Pipeline{config: cfg}
}

// Run executes the full pipeline: validate → upload+analyze → deduplicate → export.
func (p *Pipeline) Run() error {
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

	// Dry run: show cost estimate and exit.
	if p.config.Processing.DryRun {
		estimate := analyzer.EstimateCost(len(scanResult.Videos))
		fmt.Printf("\n=== Dry Run Cost Estimate ===\n")
		fmt.Printf("  Videos:        %d\n", len(scanResult.Videos))
		fmt.Printf("  Est. tokens:   %d input + %d output\n", estimate.InputTokens, estimate.OutputTokens)
		fmt.Printf("  Est. cost:     $%.4f\n", estimate.EstimatedCostUSD)
		fmt.Printf("=============================\n")
		return nil
	}

	// ── Phase 2: Upload + Analyze ──────────────────────────────────────
	fmt.Println("\nPhase 2: Analyzing videos...")

	ctx := context.Background()

	va, err := analyzer.NewVideoAnalyzer(ctx, analyzer.Config{
		APIKey:            p.config.Gemini.APIKey,
		Model:             p.config.Gemini.Model,
		MaxOutputTokens:   p.config.Gemini.MaxOutputTokens,
		Temperature:       p.config.Gemini.Temperature,
		RequestsPerMinute: p.config.Gemini.RequestsPerMinute,
		MaxRetries:        p.config.Gemini.MaxRetries,
		PollIntervalSec:   p.config.Processing.PollIntervalSec,
		PollTimeoutSec:    p.config.Processing.ProcessingTimeoutSec,
	})
	if err != nil {
		return fmt.Errorf("creating video analyzer: %w", err)
	}
	defer va.Close()

	var allResults []models.ImageAnalysisResult
	var gpsLocations []videoGPSData
	totalVideos := len(scanResult.Videos)

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
		// Per-video header.
		fmt.Printf("\n  ── Video %d/%d: %s (%s) ──\n", i+1, totalVideos, video.Filename, formatFileSize(video.SizeBytes))

		// Check max cost.
		if p.config.Processing.MaxCost != nil {
			currentCost := va.CostTracker().TotalCost()
			if currentCost.EstimatedCostUSD >= *p.config.Processing.MaxCost {
				fmt.Printf("  Max cost limit reached ($%.4f >= $%.4f). Stopping.\n",
					currentCost.EstimatedCostUSD, *p.config.Processing.MaxCost)
				break
			}
		}

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

		// Step 1: Upload + Process (handled inside AnalyzeVideo with spinner).
		videoBar.Describe("    Uploading & processing")
		results, err := va.AnalyzeVideo(ctx, video.Path, string(video.MIMEType))
		if err != nil {
			log.Printf("Phase 2: error analyzing %s: %v", video.Filename, err)
			fmt.Printf("    Error: %v (skipping)\n", err)
			_ = videoBar.Add(4) // Complete the per-video bar.
			_ = overallBar.Add(1)
			continue
		}
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

	cost := va.CostTracker().TotalCost()
	fmt.Printf("\n  Videos:    %d processed\n", cost.VideosProcessed)
	fmt.Printf("  Entities:  %d found\n", len(allResults))
	fmt.Printf("  Cost:      $%.4f\n", cost.EstimatedCostUSD)

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

	// Vendor enrichment.
	if len(vendors) > 0 {
		fmt.Printf("  Enriching %d vendors...\n", len(vendors))
		vendors = p.enrichVendors(ctx, vendors)
	}

	// Extract and geocode locations.
	locations := extractLocations(groups, gpsLocations)
	geocoderAPIKey := p.config.Gemini.GoogleMapsAPIKey
	if geocoderAPIKey != "" {
		geo := geocoder.NewGeocoder(geocoderAPIKey)
		for i, loc := range locations {
			if loc.HasGPS && loc.FormattedAddress == "" {
				result, err := geo.ReverseGeocode(loc.Latitude, loc.Longitude)
				if err == nil {
					locations[i].FormattedAddress = result.FormattedAddress
					locations[i].Source = "gps_reverse_geocoded"
					if locations[i].Name == "" {
						locations[i].Name = result.LocationName
					}
				}
			}
		}
	} else if hasGPSLocations(locations) {
		geo := geocoder.NewGeocoder("")
		for i, loc := range locations {
			if loc.HasGPS && loc.FormattedAddress == "" {
				result, err := geo.ReverseGeocode(loc.Latitude, loc.Longitude)
				if err == nil {
					locations[i].FormattedAddress = result.FormattedAddress
					locations[i].Source = "gps_reverse_geocoded"
					if locations[i].Name == "" {
						locations[i].Name = result.LocationName
					}
				}
			}
		}
	}

	// Assign locations to entities.
	assignLocations(groups, locations)

	fmt.Printf("  Vendors found:   %d\n", len(vendors))
	fmt.Printf("  Locations found: %d\n\n", len(locations))

	// ── Phase 4: Export ────────────────────────────────────────────────
	fmt.Println("Phase 4: Exporting CSVs...")

	exp := exporter.NewExporter(
		p.config.Scan.OutputFolder,
		p.config.Scan.OutputFolder, // No separate renamed folder for videos.
		false,                       // No renaming for videos.
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
	fmt.Printf("  Videos processed:  %d\n", cost.VideosProcessed)
	fmt.Printf("  Entities found:    %d\n", len(deduped))
	fmt.Printf("  Groups:            %d\n", len(groups))
	fmt.Printf("  Cost:              $%.4f\n", cost.EstimatedCostUSD)
	fmt.Printf("  Elapsed:           %s\n", elapsed.Round(time.Second))
	fmt.Printf("===============================\n")

	return nil
}

// videoGPSData holds GPS coordinates extracted from a video file.
type videoGPSData struct {
	videoPath string
	lat, lng  float64
}

// enrichVendors enriches vendor data via Gemini lookups.
func (p *Pipeline) enrichVendors(ctx context.Context, vendors []models.DetectedVendor) []models.DetectedVendor {
	client, err := genai.NewClient(ctx, option.WithAPIKey(p.config.Gemini.APIKey))
	if err != nil {
		log.Printf("vendor enrichment: failed to create client: %v", err)
		return vendors
	}
	defer client.Close()

	model := client.GenerativeModel(p.config.Gemini.Model)
	model.ResponseMIMEType = "application/json"

	enricher := vendor.NewEnricher(model)

	for i, v := range vendors {
		enriched := enricher.Enrich(ctx, v.Name)
		if enriched != nil && enriched.Source == "ai_enriched" {
			// Store enrichment data in the vendor's entity names as metadata.
			// The CSV exporter will pick up the vendor name as-is.
			if enriched.FullName != "" {
				vendors[i].Name = enriched.FullName
			}
		}
	}

	return vendors
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
				Source:      "gemini_suggested",
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

// hasGPSLocations checks if any locations have GPS data.
func hasGPSLocations(locations []models.DetectedLocation) bool {
	for _, l := range locations {
		if l.HasGPS {
			return true
		}
	}
	return false
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
