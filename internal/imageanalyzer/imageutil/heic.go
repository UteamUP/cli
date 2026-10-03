package imageutil

import (
	"context"
	"fmt"
	"github.com/uteamup/cli/internal/security"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// heicMagicSignatures contains the ftyp box brand values that identify
// HEIC/HEIF files. The ftyp box starts at byte offset 4 in the file.
var heicMagicSignatures = []string{
	"ftypheic",
	"ftypheix",
	"ftypmif1",
	"ftypmsf1",
	"ftypheis",
	"ftyphevc",
}

// IsHEIC returns true if filePath has a .heic or .heif extension
// (case-insensitive).
func IsHEIC(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	return ext == ".heic" || ext == ".heif"
}

// hasHEICMagicBytes checks whether the file starts with HEIC/HEIF
// magic bytes (ftyp box at offset 4).
func hasHEICMagicBytes(filePath string) bool {
	f, err := security.OpenRegular(filePath, 100*1024*1024)
	if err != nil {
		return false
	}
	defer f.Close()

	header := make([]byte, 12)
	n, err := f.Read(header)
	if err != nil || n < 12 {
		return false
	}

	// The ftyp box brand starts at byte 4.
	brand := string(header[4:12])
	for _, sig := range heicMagicSignatures {
		if strings.HasPrefix(brand, sig) {
			return true
		}
	}
	return false
}

// ConvertHEICToJPEG converts a HEIC/HEIF file to JPEG bytes.
//
// On macOS it uses the built-in `sips` command which reliably handles
// HEIC files. On other platforms it returns an error with guidance.
func ConvertHEICToJPEG(filePath string) ([]byte, error) {
	if !hasHEICMagicBytes(filePath) && !IsHEIC(filePath) {
		return nil, fmt.Errorf("file %q does not appear to be a HEIC/HEIF image", filePath)
	}

	if runtime.GOOS == "darwin" {
		return convertHEICViaSips(filePath)
	}

	return nil, fmt.Errorf(
		"HEIC/HEIF conversion is not supported on %s; "+
			"convert the file to JPEG manually or use macOS where `sips` is available",
		runtime.GOOS,
	)
}

// convertHEICViaSips uses the macOS built-in `sips` command to convert
// HEIC to JPEG.
func convertHEICViaSips(filePath string) ([]byte, error) {
	source, err := security.ReadFile(filePath, 15*1024*1024)
	if err != nil {
		return nil, err
	}
	input, err := os.CreateTemp("", "heic-source-*.heic")
	if err != nil {
		return nil, err
	}
	defer os.Remove(input.Name())
	if _, err := input.Write(source); err != nil {
		input.Close()
		return nil, err
	}
	input.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	dimensions, err := exec.CommandContext(ctx, "sips", "-g", "pixelWidth", "-g", "pixelHeight", input.Name()).Output()
	if err != nil {
		return nil, err
	}
	width, height := 0, 0
	for _, line := range strings.Split(string(dimensions), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 {
			value, _ := strconv.Atoi(fields[1])
			switch fields[0] {
			case "pixelWidth:":
				width = value
			case "pixelHeight:":
				height = value
			}
		}
	}
	if err := validateImageDimensions(width, height, maxOutputImageDimension); err != nil {
		return nil, err
	}
	tmpFile, err := os.CreateTemp("", "heic-convert-*.jpg")
	if err != nil {
		return nil, fmt.Errorf("create temp file for HEIC conversion: %w", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	cmd := exec.CommandContext(ctx, "sips", "-s", "format", "jpeg", input.Name(), "--out", tmpPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("sips conversion failed: %w\noutput: %s", err, string(output))
	}

	data, err := security.ReadFile(tmpPath, 15*1024*1024)
	if err != nil {
		return nil, fmt.Errorf("read converted JPEG: %w", err)
	}

	return data, nil
}
