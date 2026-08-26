package imageutil

import (
	"bytes"
	"compress/bzip2"
	"encoding/base64"
	"io"
	"strings"
	"testing"
)

// This is the compressed upstream x/image WebP regression fixture that
// references Huffman group 65535. Vulnerable decoders allocated about 170 MiB
// for unused Huffman groups before decoding the image.
const largeHuffmanIndexWebPBzip2Base64 = "QlpoOTFBWSZTWdD9D0AABRR+zdwAyACAAIBAVyRRgAACRAAAAMAAQAAADiAAUKAAAAACSkBoNAGmR6SbqnOnIsnOD++mypKryggARGSBAJJsPfvWaUIY/M1I0suEACPH/i7kinChIaH6HoA="

func TestResizeImageRejectsExcessiveWebPHuffmanGroups(t *testing.T) {
	compressed, err := base64.StdEncoding.DecodeString(largeHuffmanIndexWebPBzip2Base64)
	if err != nil {
		t.Fatalf("decode fixture: %v", err)
	}

	payload, err := io.ReadAll(bzip2.NewReader(bytes.NewReader(compressed)))
	if err != nil {
		t.Fatalf("decompress fixture: %v", err)
	}

	_, err = ResizeImage(payload, 2048)
	if err == nil || !strings.Contains(err.Error(), "vp8l: too many Huffman trees") {
		t.Fatalf("expected bounded VP8L rejection, got %v", err)
	}
}

func TestValidateImageDimensionsRejectsDecompressionBombShapes(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
		output int
	}{
		{name: "excessive width", width: maxSourceImageDimension + 1, height: 1, output: 2048},
		{name: "excessive pixels", width: 10_000, height: 10_000, output: 2048},
		{name: "invalid output", width: 100, height: 100, output: maxOutputImageDimension + 1},
		{name: "zero width", width: 0, height: 100, output: 2048},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := validateImageDimensions(test.width, test.height, test.output); err == nil {
				t.Fatal("expected unsafe dimensions to be rejected")
			}
		})
	}
}

func TestValidateImageDimensionsAcceptsBoundedImage(t *testing.T) {
	if err := validateImageDimensions(4096, 4096, 2048); err != nil {
		t.Fatal(err)
	}
}
