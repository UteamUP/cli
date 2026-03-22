package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"github.com/uteamup/cli/internal/config"
)

var (
	imageOutputDir string
	imageModel     string
	imageAPIKey    string
	imageDryRun    bool
	imageNoRename  bool
	imageConfig    string
	imageVerbose   bool
)

var imageCmd = &cobra.Command{
	Use:     "image",
	Aliases: []string{"img", "images"},
	Short:   "Analyze images for CMMS inventory data",
	Long: `Analyze images using the UteamUP Image Analyzer tool.

The image analyzer uses AI (Google Gemini) to process batches of images,
extracting CMMS-relevant inventory data and exporting results to CSV.

This command requires the UteamUP Image Analyzer Python tool to be installed.
See: https://github.com/uteamup/image-analyzer`,
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
  uteamup img analyze ./photos --model gemini-2.5-pro --api-key AIza...
  uteamup img analyze /path/to/images -o /path/to/output --model gemini-2.5-flash --verbose
  ut img analyze ./images --no-rename`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		imagePath := args[0]

		// Load Gemini settings from CLI config (profile defaults)
		if cfg, err := config.Load(); err == nil {
			if profile, err := cfg.ActiveProfileConfig(); err == nil {
				// Use config values as defaults when flags aren't explicitly set
				if imageAPIKey == "" && profile.GeminiAPIKey != "" {
					imageAPIKey = profile.GeminiAPIKey
				}
				if imageModel == "" && profile.GeminiModel != "" {
					imageModel = profile.GeminiModel
				}
			}
		}

		// Resolve the image path to absolute
		absImagePath, err := filepath.Abs(imagePath)
		if err != nil {
			return fmt.Errorf("resolving image path: %w", err)
		}

		// Check that the image path exists
		info, err := os.Stat(absImagePath)
		if err != nil {
			return fmt.Errorf("image path %q does not exist: %w", absImagePath, err)
		}
		if !info.IsDir() {
			return fmt.Errorf("image path %q is not a directory", absImagePath)
		}

		// Locate the image analyzer
		analyzerDir, err := findAnalyzerDir()
		if err != nil {
			return err
		}

		// Check for Python venv
		pythonBin := filepath.Join(analyzerDir, ".venv", "bin", "python")
		if runtime.GOOS == "windows" {
			pythonBin = filepath.Join(analyzerDir, ".venv", "Scripts", "python.exe")
		}
		if _, err := os.Stat(pythonBin); err != nil {
			return fmt.Errorf("Python virtual environment not found at %s\n\nSetup instructions:\n  cd %s\n  python3 -m venv .venv\n  .venv/bin/pip install -r requirements.txt",
				pythonBin, analyzerDir)
		}

		// Resolve output directory to absolute
		absOutputDir, err := filepath.Abs(imageOutputDir)
		if err != nil {
			return fmt.Errorf("resolving output path: %w", err)
		}

		// Build the command arguments
		cmdArgs := []string{"-m", "image_analyzer", "analyze", "--folder", absImagePath, "--output", absOutputDir}

		if imageDryRun {
			cmdArgs = append(cmdArgs, "--dry-run")
		}
		if imageNoRename {
			cmdArgs = append(cmdArgs, "--no-rename")
		}
		if imageVerbose {
			cmdArgs = append(cmdArgs, "--verbose")
		}
		if imageConfig != "" {
			absConfig, err := filepath.Abs(imageConfig)
			if err != nil {
				return fmt.Errorf("resolving config path: %w", err)
			}
			cmdArgs = append(cmdArgs, "--config", absConfig)
		}

		// Build and run the command
		execCmd := exec.Command(pythonBin, cmdArgs...)
		execCmd.Dir = analyzerDir
		execCmd.Env = append(os.Environ(), "PYTHONPATH="+filepath.Join(analyzerDir, "src"))

		// Pass API key and model via env vars (Python config reads these)
		if imageAPIKey != "" {
			execCmd.Env = append(execCmd.Env, "GEMINI_API_KEY="+imageAPIKey)
		}
		if imageModel != "" {
			execCmd.Env = append(execCmd.Env, "GEMINI_MODEL="+imageModel)
		}

		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr
		execCmd.Stdin = os.Stdin

		// Count images in source folder for status
		imageCount := countImages(absImagePath)
		model := imageModel
		if model == "" {
			model = "gemini-3.1-flash-lite-preview (default)"
		}
		fmt.Printf("\n=== UteamUP Image Analyzer ===\n")
		fmt.Printf("  Source:  %s\n", absImagePath)
		fmt.Printf("  Output:  %s\n", absOutputDir)
		fmt.Printf("  Images:  %d found\n", imageCount)
		fmt.Printf("  Model:   %s\n", model)
		if imageDryRun {
			fmt.Printf("  Mode:    DRY RUN (cost estimate only)\n")
		}
		fmt.Printf("==============================\n\n")

		if imageVerbose {
			fmt.Fprintf(os.Stderr, "Analyzer: %s\n", analyzerDir)
			fmt.Fprintf(os.Stderr, "Command:  %s %s\n", pythonBin, strings.Join(cmdArgs, " "))
		}

		if err := execCmd.Run(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				os.Exit(exitErr.ExitCode())
			}
			return fmt.Errorf("running image analyzer: %w", err)
		}

		return nil
	},
}

// findAnalyzerDir locates the UteamUP Image Analyzer installation.
// It checks (in order):
//  1. UTEAMUP_IMAGE_ANALYZER_PATH environment variable
//  2. Sibling directory ../UteamUP_ImageAnalyzer relative to the CLI binary
//  3. Sibling directory ../UteamUP_ImageAnalyzer relative to the current working directory
//  4. ~/UteamUP_ImageAnalyzer
//  5. ~/UteamUP_Development/ActiveProjects/UteamUP_ImageAnalyzer
func findAnalyzerDir() (string, error) {
	// 1. Environment variable
	if envPath := os.Getenv("UTEAMUP_IMAGE_ANALYZER_PATH"); envPath != "" {
		absPath, err := filepath.Abs(envPath)
		if err != nil {
			return "", fmt.Errorf("resolving UTEAMUP_IMAGE_ANALYZER_PATH: %w", err)
		}
		if isAnalyzerDir(absPath) {
			return absPath, nil
		}
		return "", fmt.Errorf("UTEAMUP_IMAGE_ANALYZER_PATH=%q does not contain a valid image analyzer installation", envPath)
	}

	// 2. Sibling directory relative to CLI binary
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		siblingPath := filepath.Join(exeDir, "..", "UteamUP_ImageAnalyzer")
		if absPath, err := filepath.Abs(siblingPath); err == nil && isAnalyzerDir(absPath) {
			return absPath, nil
		}
	}

	// 3. Sibling directory relative to current working directory
	if cwd, err := os.Getwd(); err == nil {
		cwdSibling := filepath.Join(cwd, "..", "UteamUP_ImageAnalyzer")
		if absPath, err := filepath.Abs(cwdSibling); err == nil && isAnalyzerDir(absPath) {
			return absPath, nil
		}
	}

	// 4. Home directory
	if homeDir, err := os.UserHomeDir(); err == nil {
		homePath := filepath.Join(homeDir, "UteamUP_ImageAnalyzer")
		if isAnalyzerDir(homePath) {
			return homePath, nil
		}
		// 5. Common dev path
		devPath := filepath.Join(homeDir, "UteamUP_Development", "ActiveProjects", "UteamUP_ImageAnalyzer")
		if isAnalyzerDir(devPath) {
			return devPath, nil
		}
	}

	return "", fmt.Errorf(`UteamUP Image Analyzer not found.

Searched locations:
  1. UTEAMUP_IMAGE_ANALYZER_PATH environment variable (not set)
  2. Sibling directory ../UteamUP_ImageAnalyzer (relative to CLI binary)
  3. Sibling directory ../UteamUP_ImageAnalyzer (relative to current directory)
  4. ~/UteamUP_ImageAnalyzer
  5. ~/UteamUP_Development/ActiveProjects/UteamUP_ImageAnalyzer

Install the image analyzer:
  git clone https://github.com/UteamUP/ImageAnalyzer ~/UteamUP_ImageAnalyzer
  cd ~/UteamUP_ImageAnalyzer
  python3 -m venv .venv
  .venv/bin/pip install -r requirements.txt

Or set the UTEAMUP_IMAGE_ANALYZER_PATH environment variable to point to your installation.`)
}

// countImages counts image files in a directory (recursively).
func countImages(dir string) int {
	imageExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".webp": true,
		".heic": true, ".heif": true, ".tiff": true, ".bmp": true,
	}
	count := 0
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if imageExts[ext] {
			count++
		}
		return nil
	})
	return count
}

// isAnalyzerDir checks whether the given directory looks like a valid image analyzer installation.
func isAnalyzerDir(dir string) bool {
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return false
	}
	// Check for the src/image_analyzer directory as a marker
	srcDir := filepath.Join(dir, "src", "image_analyzer")
	if info, err := os.Stat(srcDir); err == nil && info.IsDir() {
		return true
	}
	return false
}

func init() {
	imageAnalyzeCmd.Flags().StringVarP(&imageOutputDir, "output", "o", "./Output", "Output folder for analysis results")
	imageAnalyzeCmd.Flags().StringVar(&imageModel, "model", "", "Gemini model: gemini-pro-latest (always newest), gemini-3.1-pro-preview, gemini-3.1-flash-lite-preview, gemini-3-pro-preview, gemini-2.5-pro, gemini-2.5-flash")
	imageAnalyzeCmd.Flags().StringVar(&imageAPIKey, "api-key", "", "Google Gemini API key (overrides GEMINI_API_KEY env var)")
	imageAnalyzeCmd.Flags().BoolVar(&imageDryRun, "dry-run", false, "Estimate cost only, do not process images")
	imageAnalyzeCmd.Flags().BoolVar(&imageNoRename, "no-rename", false, "Skip image renaming after analysis")
	imageAnalyzeCmd.Flags().StringVar(&imageConfig, "config", "", "Path to config.yaml override")
	imageAnalyzeCmd.Flags().BoolVarP(&imageVerbose, "verbose", "V", false, "Enable verbose output")

	imageCmd.AddCommand(imageAnalyzeCmd)
}
