package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/uteamup/cli/internal/config"
	vaconfig "github.com/uteamup/cli/internal/videoanalyzer/config"
	"github.com/uteamup/cli/internal/videoanalyzer/pipeline"
)

var (
	videoOutputDir           string
	videoModel               string
	videoAPIKey              string
	videoDryRun              bool
	videoConfig              string
	videoVerbose             bool
	videoMaxCost             float64
	videoSimilarityThreshold float64
	videoConfidenceThreshold float64
	videoMapsAPIKey          string
)

var videoCmd = &cobra.Command{
	Use:     "video",
	Aliases: []string{"vid", "videos"},
	Short:   "Analyze videos for CMMS inventory data",
	Long: `Analyze videos using the UteamUP Video Analyzer.

The video analyzer uses AI (Google Gemini) to process video files,
extracting CMMS-relevant inventory data (assets, tools, parts, chemicals)
with timestamps, GPS locations, and vendor information, then exports
results to CSV.`,
}

var videoAnalyzeCmd = &cobra.Command{
	Use:   "analyze <path>",
	Short: "Analyze video files for CMMS inventory data",
	Long: `Analyze video files (MP4, MOV) in the specified path using AI-powered video analysis.

The analyzer uploads each video to Google Gemini, extracts CMMS-relevant data
(equipment type, manufacturer, model, condition, timestamps), deduplicates
entities across frames and videos, and exports results to CSV files.

GIF files found in the input path are routed to the image analyzer automatically.

Examples:
  uteamup video analyze ./videos
  uteamup video analyze ./recording.mp4 --dry-run
  uteamup vid analyze ./videos --model gemini-2.5-pro --api-key AIza...
  ut vid analyze ./videos -o ./results --verbose
  ut video analyze ./walkthrough.mov --max-cost 5.00`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		videoPath := args[0]

		// Load Gemini settings from CLI config (profile defaults).
		if cfg, err := config.Load(); err == nil {
			if profile, err := cfg.ActiveProfileConfig(); err == nil {
				if videoAPIKey == "" && profile.GeminiAPIKey != "" {
					videoAPIKey = profile.GeminiAPIKey
				}
				if videoModel == "" && profile.GeminiModel != "" {
					videoModel = profile.GeminiModel
				}
				if videoMapsAPIKey == "" && profile.GoogleMapsAPIKey != "" {
					videoMapsAPIKey = profile.GoogleMapsAPIKey
				}
			}
		}

		// Resolve paths to absolute.
		absVideoPath, err := filepath.Abs(videoPath)
		if err != nil {
			return fmt.Errorf("resolving video path: %w", err)
		}

		// Check that the path exists.
		_, err = os.Stat(absVideoPath)
		if err != nil {
			return fmt.Errorf("video path %q does not exist: %w", absVideoPath, err)
		}

		// Resolve output directory.
		absOutputDir, err := filepath.Abs(videoOutputDir)
		if err != nil {
			return fmt.Errorf("resolving output path: %w", err)
		}

		// Build config options from CLI flags.
		var opts []vaconfig.ConfigOption
		opts = append(opts, vaconfig.WithFolderOverride(absVideoPath))
		opts = append(opts, vaconfig.WithOutputOverride(absOutputDir))
		opts = append(opts, vaconfig.WithDryRun(videoDryRun))
		opts = append(opts, vaconfig.WithAPIKey(videoAPIKey))
		opts = append(opts, vaconfig.WithModel(videoModel))
		opts = append(opts, vaconfig.WithGoogleMapsAPIKey(videoMapsAPIKey))
		if cmd.Flags().Changed("max-cost") {
			opts = append(opts, vaconfig.WithMaxCost(&videoMaxCost))
		}

		// Load config.
		configPath := videoConfig
		if configPath == "" {
			configPath = "config.yaml"
		}
		cfg, err := vaconfig.LoadConfig(configPath, opts...)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		// Apply thresholds if explicitly set.
		if cmd.Flags().Changed("similarity-threshold") {
			cfg.Processing.GroupingSimilarityThreshold = videoSimilarityThreshold
		}
		if cmd.Flags().Changed("confidence-threshold") {
			cfg.Processing.ConfidenceThreshold = videoConfidenceThreshold
		}

		// Validate config.
		if errs := cfg.Validate(); len(errs) > 0 {
			fmt.Fprintln(os.Stderr, "Configuration errors:")
			for _, e := range errs {
				fmt.Fprintf(os.Stderr, "  - %s\n", e)
			}
			return fmt.Errorf("invalid configuration")
		}

		// Print banner.
		model := videoModel
		if model == "" {
			model = cfg.Gemini.Model
		}
		fmt.Printf("\n=== UteamUP Video Analyzer ===\n")
		fmt.Printf("  Source:  %s\n", absVideoPath)
		fmt.Printf("  Output:  %s\n", absOutputDir)
		fmt.Printf("  Model:   %s\n", model)
		if videoDryRun {
			fmt.Printf("  Mode:    DRY RUN (cost estimate only)\n")
		}
		fmt.Printf("==============================\n")

		// Create and run pipeline.
		return pipeline.NewPipeline(cfg).Run()
	},
}

func init() {
	videoAnalyzeCmd.Flags().StringVarP(&videoOutputDir, "output", "o", "./Output", "Output folder for analysis results")
	videoAnalyzeCmd.Flags().StringVar(&videoModel, "model", "", "Gemini model: gemini-pro-latest, gemini-3.1-pro-preview, gemini-3.1-flash-lite-preview, gemini-2.5-pro, gemini-2.5-flash")
	videoAnalyzeCmd.Flags().StringVar(&videoAPIKey, "api-key", "", "Google Gemini API key (overrides GEMINI_API_KEY env var)")
	videoAnalyzeCmd.Flags().BoolVar(&videoDryRun, "dry-run", false, "Estimate cost only, do not process videos")
	videoAnalyzeCmd.Flags().StringVar(&videoConfig, "config", "", "Path to config.yaml override")
	videoAnalyzeCmd.Flags().BoolVarP(&videoVerbose, "verbose", "V", false, "Enable verbose output")
	videoAnalyzeCmd.Flags().Float64Var(&videoMaxCost, "max-cost", 0, "Maximum budget in USD (stops when reached)")
	videoAnalyzeCmd.Flags().Float64Var(&videoSimilarityThreshold, "similarity-threshold", 0.75, "Grouping similarity threshold (0.0-1.0)")
	videoAnalyzeCmd.Flags().Float64Var(&videoConfidenceThreshold, "confidence-threshold", 0.5, "Minimum confidence to classify (0.0-1.0)")
	videoAnalyzeCmd.Flags().StringVar(&videoMapsAPIKey, "maps-api-key", "", "Google Maps API key for reverse geocoding GPS coordinates")

	videoCmd.AddCommand(videoAnalyzeCmd)
}
