// Package pipeline orchestrates the 4-phase image analysis pipeline:
// scan, analyze, group, and export.
package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/schollz/progressbar/v3"

	"github.com/uteamup/cli/internal/imageanalyzer/analyzer"
	"github.com/uteamup/cli/internal/imageanalyzer/checkpoint"
	iaconfig "github.com/uteamup/cli/internal/imageanalyzer/config"
	"github.com/uteamup/cli/internal/imageanalyzer/exporter"
	"github.com/uteamup/cli/internal/imageanalyzer/grouper"
	"github.com/uteamup/cli/internal/imageanalyzer/imageutil"
	"github.com/uteamup/cli/internal/imageanalyzer/models"
	"github.com/uteamup/cli/internal/imageanalyzer/scanner"
)

// Pipeline orchestrates the 4-phase image analysis pipeline.
type Pipeline struct {
	config *iaconfig.AppConfig
}

// NewPipeline creates a new Pipeline with the given configuration.
func NewPipeline(cfg *iaconfig.AppConfig) *Pipeline {
	return &Pipeline{config: cfg}
}

// Run executes the full pipeline: scan -> analyze -> group -> export.
func (p *Pipeline) Run() error {
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

	// Dry-run: estimate cost and stop.
	if p.config.Processing.DryRun {
		p.printDryRun(imagesToAnalyze, duplicatesFound, editPairs)
		return nil
	}

	// ── Phase 2: Analyze ───────────────────────────────────────────────
	fmt.Println("Phase 2: Analyzing images with Gemini...")
	log.Printf("Phase 2: Analyzing images with Gemini")

	geminiAnalyzer, err := analyzer.NewGeminiAnalyzer(p.config.Gemini)
	if err != nil {
		return fmt.Errorf("creating analyzer: %w", err)
	}

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

	bar := progressbar.NewOptions(len(imagesToAnalyze),
		progressbar.OptionSetDescription("Analyzing"),
		progressbar.OptionShowCount(),
		progressbar.OptionSetWidth(40),
		progressbar.OptionClearOnFinish(),
	)

	ctx := context.Background()
	for _, imageInfo := range imagesToAnalyze {
		_ = bar.Add(1)

		// Skip already processed.
		if cp.IsProcessed(imageInfo.SHA256Hash) {
			continue
		}

		// Check budget.
		if p.config.Processing.MaxCost != nil {
			nextCost := analyzer.EstimateCost(1).EstimatedTotalCostUSD
			spent := geminiAnalyzer.TotalCost().EstimatedTotalCostUSD
			if spent+nextCost > *p.config.Processing.MaxCost {
				fmt.Printf("\n  Budget limit reached: spent $%.4f of $%.2f cap\n",
					spent, *p.config.Processing.MaxCost)
				break
			}
		}

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
			continue
		}

		imageResults, err := geminiAnalyzer.AnalyzeImage(ctx, imageInfo.Path, imgBytes)
		if err != nil {
			log.Printf("Failed to analyze image %s: %v", imageInfo.Path, err)
			failResult := createFailResult(imageInfo, fmt.Sprintf("Analysis error: %v", err))
			results = append(results, failResult)
			raw, _ := json.Marshal(failResult)
			_ = cp.AddResult(imageInfo.SHA256Hash, raw)
			continue
		}

		for i := range imageResults {
			// Attach iPhone edit pair paths.
			if pairs, ok := editPairs[imageInfo.Filename]; ok {
				imageResults[i].PairedImages = pairs
			}

			// Apply confidence threshold.
			if imageResults[i].Classification.Confidence < p.config.Processing.ConfidenceThreshold {
				imageResults[i].Classification.PrimaryType = models.EntityTypeUnclassified
				imageResults[i].FlaggedForReview = true
				imageResults[i].ReviewReason = fmt.Sprintf(
					"Low confidence: %.2f", imageResults[i].Classification.Confidence,
				)
			}
		}

		results = append(results, imageResults...)

		// Checkpoint all results for this image.
		raw, _ := json.Marshal(imageResults)
		_ = cp.AddResult(imageInfo.SHA256Hash, raw)
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
	return nil
}

// printDryRun displays a cost estimate without making API calls.
func (p *Pipeline) printDryRun(images []models.ImageInfo, duplicatesFound int, editPairs map[string][]string) {
	estimate := analyzer.EstimateCost(len(images))
	rpm := p.config.Gemini.RequestsPerMinute
	if rpm < 1 {
		rpm = 1
	}
	estMinutes := float64(len(images)) / float64(rpm)

	fmt.Println("\n=== DRY RUN — Cost Estimate ===")
	fmt.Printf("Images to analyze:  %d\n", len(images))
	fmt.Printf("Duplicates skipped: %d\n", duplicatesFound)
	fmt.Printf("iPhone edit pairs:  %d (variants skipped)\n", len(editPairs))
	fmt.Printf("Model:              %s\n", p.config.Gemini.Model)
	fmt.Printf("Est. input tokens:  %d\n", estimate.EstimatedInputTokens)
	fmt.Printf("Est. output tokens: %d\n", estimate.EstimatedOutputTokens)
	fmt.Printf("Est. total cost:    $%.4f\n", estimate.EstimatedTotalCostUSD)
	fmt.Printf("Est. time:          %.1f minutes\n", estMinutes)
	fmt.Printf("                    (at %d req/min)\n", p.config.Gemini.RequestsPerMinute)
	if p.config.Processing.MaxCost != nil {
		fmt.Printf("Budget cap:         $%.2f\n", *p.config.Processing.MaxCost)
	}
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
