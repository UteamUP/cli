package imageutil

import (
	"fmt"
	"os"
)

// LoadImageBytes loads an image from filePath, converts HEIC if needed,
// resizes to fit within maxDimension, and returns JPEG bytes ready for
// API submission (e.g. Gemini).
func LoadImageBytes(filePath string, maxDimension int) ([]byte, error) {
	var rawBytes []byte

	if IsHEIC(filePath) {
		jpegBytes, err := ConvertHEICToJPEG(filePath)
		if err != nil {
			return nil, fmt.Errorf("convert HEIC %q: %w", filePath, err)
		}
		rawBytes = jpegBytes
	} else {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("read image %q: %w", filePath, err)
		}
		rawBytes = data
	}

	resized, err := ResizeImage(rawBytes, maxDimension)
	if err != nil {
		return nil, fmt.Errorf("resize image %q: %w", filePath, err)
	}

	return resized, nil
}
