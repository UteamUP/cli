package imageutil

import (
	"image"
	"os"
	"path/filepath"
	"strings"
)

// supportedExtensions lists the image file extensions this package can
// handle. HEIC/HEIF are included but require macOS sips for conversion.
var supportedExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".bmp":  true,
	".tif":  true,
	".tiff": true,
	".webp": true,
	".heic": true,
	".heif": true,
}

// IsSupportedFormat returns true if ext (with or without leading dot)
// is in the supported image format list.
func IsSupportedFormat(ext string) bool {
	ext = strings.ToLower(ext)
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	return supportedExtensions[ext]
}

// IsValidImage checks whether filePath points to a readable image file
// in a supported format. For HEIC/HEIF files it only validates the
// extension (decoding requires conversion). For other formats it
// attempts to decode the image header via image.DecodeConfig.
func IsValidImage(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	if !supportedExtensions[ext] {
		return false
	}

	// For HEIC/HEIF we cannot decode the header in pure Go, so just
	// verify the file exists and is non-empty.
	if ext == ".heic" || ext == ".heif" {
		info, err := os.Stat(filePath)
		return err == nil && info.Size() > 0
	}

	f, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer f.Close()

	_, _, err = image.DecodeConfig(f)
	return err == nil
}
