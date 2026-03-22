package imageutil

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
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
	f, err := os.Open(filePath)
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
	tmpFile, err := os.CreateTemp("", "heic-convert-*.jpg")
	if err != nil {
		return nil, fmt.Errorf("create temp file for HEIC conversion: %w", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	cmd := exec.Command("sips", "-s", "format", "jpeg", filePath, "--out", tmpPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("sips conversion failed: %w\noutput: %s", err, string(output))
	}

	data, err := os.ReadFile(tmpPath)
	if err != nil {
		return nil, fmt.Errorf("read converted JPEG: %w", err)
	}

	return data, nil
}
