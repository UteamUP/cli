package imageutil

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/bmp"
	"golang.org/x/image/draw"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

const (
	maxSourceImageDimension = 16_384
	maxSourceImagePixels    = 50_000_000
	maxOutputImageDimension = 4_096
)

// ResizeImage decodes an image from imgBytes (any registered format),
// checks if the largest dimension exceeds maxDimension, and resizes
// preserving aspect ratio using CatmullRom interpolation. The result is
// always normalized to JPEG at quality 90 before upload so the declared MIME
// type cannot disagree with the payload.
func ResizeImage(imgBytes []byte, maxDimension int) ([]byte, error) {
	config, _, err := image.DecodeConfig(bytes.NewReader(imgBytes))
	if err != nil {
		return nil, fmt.Errorf("decode image header: %w", err)
	}
	if err := validateImageDimensions(config.Width, config.Height, maxDimension); err != nil {
		return nil, err
	}

	src, _, err := image.Decode(bytes.NewReader(imgBytes))
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}

	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	normalized := src
	if width > maxDimension || height > maxDimension {
		// Calculate new dimensions preserving aspect ratio.
		ratioW := float64(maxDimension) / float64(width)
		ratioH := float64(maxDimension) / float64(height)
		ratio := ratioW
		if ratioH < ratioW {
			ratio = ratioH
		}

		newWidth := int(float64(width) * ratio)
		newHeight := int(float64(height) * ratio)

		dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
		draw.CatmullRom.Scale(dst, dst.Bounds(), src, bounds, draw.Over, nil)
		normalized = dst
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, normalized, &jpeg.Options{Quality: 90}); err != nil {
		return nil, fmt.Errorf("encode resized jpeg: %w", err)
	}

	return buf.Bytes(), nil
}

func validateImageDimensions(width int, height int, maxOutputDimension int) error {
	if width <= 0 || height <= 0 {
		return fmt.Errorf("image dimensions must be positive")
	}
	if maxOutputDimension < 1 || maxOutputDimension > maxOutputImageDimension {
		return fmt.Errorf("maximum output image dimension must be between 1 and %d", maxOutputImageDimension)
	}
	if width > maxSourceImageDimension || height > maxSourceImageDimension ||
		int64(width)*int64(height) > maxSourceImagePixels {
		return fmt.Errorf("source image resolution exceeds the safe processing limit")
	}
	return nil
}
