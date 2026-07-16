package imageutil

import "testing"

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
