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

// ResizeImage decodes an image from imgBytes (any registered format),
// checks if the largest dimension exceeds maxDimension, and resizes
// preserving aspect ratio using CatmullRom interpolation. The result
// is re-encoded as JPEG at quality 90. If no resize is needed the
// original bytes are returned unchanged.
func ResizeImage(imgBytes []byte, maxDimension int) ([]byte, error) {
	src, _, err := image.Decode(bytes.NewReader(imgBytes))
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}

	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if width <= maxDimension && height <= maxDimension {
		return imgBytes, nil
	}

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

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, dst, &jpeg.Options{Quality: 90}); err != nil {
		return nil, fmt.Errorf("encode resized jpeg: %w", err)
	}

	return buf.Bytes(), nil
}
