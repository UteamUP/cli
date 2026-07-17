package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/uteamup/cli/internal/auth"
	"github.com/uteamup/cli/internal/config"
	"github.com/uteamup/cli/internal/imageanalyzer/checkpoint"
	iaconfig "github.com/uteamup/cli/internal/imageanalyzer/config"
	"github.com/uteamup/cli/internal/imageanalyzer/pipeline"
	"github.com/uteamup/cli/internal/mediaanalyzer"
)

var (
	imageOutputDir           string
	imageDryRun              bool
	imageNoRename            bool
	imageConfig              string
	imageResume              bool
	imageSimilarityThreshold float64
	imageConfidenceThreshold float64
	imageTimeout             time.Duration
)

var imageCmd = &cobra.Command{
	Use:     "image",
	Aliases: []string{"img", "images"},
	Short:   "Analyze images for CMMS inventory data",
	Long: `Analyze images using the UteamUP Image Analyzer.

Media is sent to the authenticated UteamUP backend. The server selects the
tenant's governed AI route, including Tenant BYOK when it is active.`,
}

var imageAnalyzeCmd = &cobra.Command{
	Use:   "analyze <path>",
	Short: "Analyze images in a folder for CMMS inventory data",
	Long: `Analyze all images in the specified folder using AI-powered image analysis.

The analyzer processes images in batches, extracts CMMS-relevant data
(equipment type, manufacturer, model, condition, etc.), and exports
results to CSV files.

Examples:
  uteamup image analyze ./photos
  uteamup image analyze ./photos --output ./results --dry-run
  uteamup img analyze /path/to/images -o /path/to/output --verbose
  ut img analyze ./images --no-rename
  ut img analyze ./photos --confidence-threshold 0.7
  ut img analyze ./photos --resume`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		imagePath := args[0]

		cliConfig, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading CLI config: %w", err)
		}
		profile, err := cliConfig.ActiveProfileConfig()
		if err != nil {
			return fmt.Errorf("loading active profile: %w", err)
		}
		token, err := auth.LoadToken()
		if err != nil {
			return fmt.Errorf("loading authentication: %w", err)
		}
		if err := validateMediaTenant(profile, token); err != nil {
			return err
		}
		apiClient, err := newMediaAPIClient(profile, imageTimeout)
		if err != nil {
			return err
		}

		// Resolve the image path to absolute.
		absImagePath, err := filepath.Abs(imagePath)
		if err != nil {
			return fmt.Errorf("resolving image path: %w", err)
		}

		// Check that the image path exists.
		info, err := os.Stat(absImagePath)
		if err != nil {
			return fmt.Errorf("image path %q does not exist: %w", absImagePath, err)
		}
		if !info.IsDir() {
			return fmt.Errorf("image path %q is not a directory", absImagePath)
		}

		// Resolve output directory to absolute.
		absOutputDir, err := filepath.Abs(imageOutputDir)
		if err != nil {
			return fmt.Errorf("resolving output path: %w", err)
		}

		// Build config options from CLI flags.
		var opts []iaconfig.Option
		opts = append(opts, iaconfig.WithFolderOverride(absImagePath))
		opts = append(opts, iaconfig.WithOutputOverride(absOutputDir))
		opts = append(opts, iaconfig.WithDryRun(imageDryRun))
		opts = append(opts, iaconfig.WithNoRename(imageNoRename))

		// Load config (YAML file + env vars + CLI flag overrides).
		configPath := imageConfig
		if configPath == "" {
			configPath = "config.yaml"
		}
		cfg, err := iaconfig.LoadConfig(configPath, opts...)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		// Apply thresholds if explicitly set.
		if cmd.Flags().Changed("similarity-threshold") {
			cfg.Processing.GroupingSimilarityThreshold = imageSimilarityThreshold
		}
		if cmd.Flags().Changed("confidence-threshold") {
			cfg.Processing.ConfidenceThreshold = imageConfidenceThreshold
		}

		// Validate config.
		if errs := cfg.Validate(); len(errs) > 0 {
			fmt.Fprintln(os.Stderr, "Configuration errors:")
			for _, e := range errs {
				fmt.Fprintf(os.Stderr, "  - %s\n", e)
			}
			return fmt.Errorf("invalid configuration")
		}

		// Count images for banner.
		imageExts := map[string]bool{
			".jpg": true, ".jpeg": true, ".png": true, ".webp": true,
			".heic": true, ".heif": true, ".tiff": true, ".bmp": true,
		}
		imageCount := 0
		_ = filepath.WalkDir(absImagePath, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			ext := strings.ToLower(filepath.Ext(path))
			if imageExts[ext] {
				imageCount++
			}
			return nil
		})

		// Print banner.
		fmt.Printf("\n=== UteamUP Image Analyzer ===\n")
		fmt.Printf("  Source:  %s\n", absImagePath)
		fmt.Printf("  Output:  %s\n", absOutputDir)
		fmt.Printf("  Images:  %d found\n", imageCount)
		tenantName := strings.TrimSpace(token.TenantName)
		if tenantName == "" {
			tenantName = "authenticated tenant"
		}
		fmt.Printf("  Tenant:  %s\n", tenantName)
		fmt.Printf("  AI:      server-governed route\n")
		if imageDryRun {
			fmt.Printf("  Mode:    DRY RUN (validation and upload scope only)\n")
		}
		if imageResume {
			fmt.Printf("  Resume:  enabled\n")
		}
		fmt.Printf("==============================\n")

		// Create and run pipeline.
		return pipeline.NewPipeline(cfg, mediaanalyzer.New(apiClient)).Run(cmd.Context())
	},
}

var imageStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of an in-progress image analysis",
	Long: `Display the current checkpoint status for an image analysis run.

Shows the number of processed images, type breakdown, and timing information
from the checkpoint file.

Examples:
  uteamup image status
  ut img status`,
	RunE: func(cmd *cobra.Command, args []string) error {
		checkpointPath := ".checkpoint.json"
		if imageConfig != "" {
			// Try to load checkpoint path from config.
			if cfg, err := iaconfig.LoadConfig(imageConfig); err == nil {
				checkpointPath = cfg.Processing.CheckpointFile
			}
		}

		cp, err := checkpoint.Load(checkpointPath)
		if err != nil {
			return fmt.Errorf("loading checkpoint: %w", err)
		}

		status := cp.GetStatus()

		if status.ProcessedCount == 0 {
			fmt.Println("No checkpoint found. No analysis in progress.")
			return nil
		}

		fmt.Println("\n=== Image Analysis Status ===")
		fmt.Printf("  Processed:    %d images\n", status.ProcessedCount)
		fmt.Printf("  Started:      %s\n", status.StartedAt)
		fmt.Printf("  Last updated: %s\n", status.LastUpdated)
		fmt.Printf("  Flagged:      %d\n", status.FlaggedCount)

		if len(status.TypeBreakdown) > 0 {
			fmt.Println("\n  Type breakdown:")
			for entityType, count := range status.TypeBreakdown {
				fmt.Printf("    %-15s %d\n", entityType, count)
			}
		}
		fmt.Println("=============================")
		return nil
	},
}

func init() {
	imageAnalyzeCmd.Flags().StringVarP(&imageOutputDir, "output", "o", "./Output", "Output folder for analysis results")
	imageAnalyzeCmd.Flags().BoolVar(&imageDryRun, "dry-run", false, "Validate and show upload scope without processing images")
	imageAnalyzeCmd.Flags().BoolVar(&imageNoRename, "no-rename", false, "Skip image renaming after analysis")
	imageAnalyzeCmd.Flags().StringVar(&imageConfig, "config", "", "Path to config.yaml override")
	imageAnalyzeCmd.Flags().BoolVar(&imageResume, "resume", false, "Resume from checkpoint if available")
	imageAnalyzeCmd.Flags().Float64Var(&imageSimilarityThreshold, "similarity-threshold", 0.75, "Grouping similarity threshold (0.0-1.0)")
	imageAnalyzeCmd.Flags().Float64Var(&imageConfidenceThreshold, "confidence-threshold", 0.5, "Minimum confidence to classify (0.0-1.0)")
	imageAnalyzeCmd.Flags().DurationVar(&imageTimeout, "timeout", 5*time.Minute, "Maximum time for each backend media request (max 15m)")

	imageCmd.AddCommand(imageAnalyzeCmd)
	imageCmd.AddCommand(imageStatusCmd)
}
