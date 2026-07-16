package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/uteamup/cli/internal/auth"
	"github.com/uteamup/cli/internal/config"
	"github.com/uteamup/cli/internal/mediaanalyzer"
	vaconfig "github.com/uteamup/cli/internal/videoanalyzer/config"
	"github.com/uteamup/cli/internal/videoanalyzer/pipeline"
)

var (
	videoOutputDir           string
	videoDryRun              bool
	videoConfig              string
	videoSimilarityThreshold float64
	videoConfidenceThreshold float64
	videoTimeout             time.Duration
)

var videoCmd = &cobra.Command{
	Use:     "video",
	Aliases: []string{"vid", "videos"},
	Short:   "Analyze videos for CMMS inventory data",
	Long: `Analyze videos using the UteamUP Video Analyzer.

Media is sent to the authenticated UteamUP backend. The server selects the
tenant's governed AI route, including Tenant BYOK when it is active.`,
}

var videoAnalyzeCmd = &cobra.Command{
	Use:   "analyze <path>",
	Short: "Analyze video files for CMMS inventory data",
	Long: `Analyze video files (MP4, MOV) in the specified path using AI-powered video analysis.

The analyzer uploads each video to UteamUP, extracts CMMS-relevant data
(equipment type, manufacturer, model, condition, timestamps), deduplicates
entities across frames and videos, and exports results to CSV files.

GIF files found in the input path are reported for separate processing with
the image analyzer.

Examples:
  uteamup video analyze ./videos
  uteamup video analyze ./recording.mp4 --dry-run
  ut vid analyze ./videos -o ./results --verbose
  ut video analyze ./walkthrough.mov --timeout 10m`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		videoPath := args[0]

		token, err := auth.LoadToken()
		if err != nil {
			return fmt.Errorf("loading authentication: %w", err)
		}

		// Load CLI config for profile settings.
		cliCfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading CLI config: %w", err)
		}
		profile, err := cliCfg.ActiveProfileConfig()
		if err != nil {
			return fmt.Errorf("loading active profile: %w", err)
		}

		if err := validateMediaTenant(profile, token); err != nil {
			return err
		}
		apiClient, err := newMediaAPIClient(profile, videoTimeout)
		if err != nil {
			return err
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
		fmt.Printf("\n=== UteamUP Video Analyzer ===\n")
		fmt.Printf("  Source:  %s\n", absVideoPath)
		fmt.Printf("  Output:  %s\n", absOutputDir)
		tenantName := token.TenantName
		if tenantName == "" {
			tenantName = "authenticated tenant"
		}
		fmt.Printf("  Tenant:  %s\n", tenantName)
		fmt.Printf("  AI:      server-governed route\n")
		if videoDryRun {
			fmt.Printf("  Mode:    DRY RUN (validation and upload scope only)\n")
		}
		fmt.Printf("==============================\n")

		// Create and run pipeline.
		return pipeline.NewPipeline(cfg, mediaanalyzer.New(apiClient)).Run(cmd.Context())
	},
}

func init() {
	videoAnalyzeCmd.Flags().StringVarP(&videoOutputDir, "output", "o", "./Output", "Output folder for analysis results")
	videoAnalyzeCmd.Flags().BoolVar(&videoDryRun, "dry-run", false, "Validate and show upload scope without processing videos")
	videoAnalyzeCmd.Flags().StringVar(&videoConfig, "config", "", "Path to config.yaml override")
	videoAnalyzeCmd.Flags().Float64Var(&videoSimilarityThreshold, "similarity-threshold", 0.75, "Grouping similarity threshold (0.0-1.0)")
	videoAnalyzeCmd.Flags().Float64Var(&videoConfidenceThreshold, "confidence-threshold", 0.5, "Minimum confidence to classify (0.0-1.0)")
	videoAnalyzeCmd.Flags().DurationVar(&videoTimeout, "timeout", 10*time.Minute, "Maximum time for each backend media request (max 15m)")

	videoCmd.AddCommand(videoAnalyzeCmd)
}
