package fileutil

import (
	"os"
	"path/filepath"
	"testing"
)

// createTempFile writes the given bytes to a temporary file and returns its path.
func createTempFile(t *testing.T, data []byte) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "testfile")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	return path
}

func TestDetectMIME_MP4(t *testing.T) {
	data := make([]byte, 12)
	copy(data[4:8], "ftyp")
	copy(data[8:12], "isom")
	path := createTempFile(t, data)

	mt, err := DetectMIME(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mt != MIMETypeMP4 {
		t.Errorf("expected %s, got %s", MIMETypeMP4, mt)
	}
}

func TestDetectMIME_MP4_AVC1(t *testing.T) {
	data := make([]byte, 12)
	copy(data[4:8], "ftyp")
	copy(data[8:12], "avc1")
	path := createTempFile(t, data)

	mt, err := DetectMIME(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mt != MIMETypeMP4 {
		t.Errorf("expected %s, got %s", MIMETypeMP4, mt)
	}
}

func TestDetectMIME_MOV(t *testing.T) {
	data := make([]byte, 12)
	copy(data[4:8], "ftyp")
	copy(data[8:12], "qt  ")
	path := createTempFile(t, data)

	mt, err := DetectMIME(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mt != MIMETypeMOV {
		t.Errorf("expected %s, got %s", MIMETypeMOV, mt)
	}
}

func TestDetectMIME_GIF87a(t *testing.T) {
	data := []byte("GIF87a\x00\x00\x00\x00\x00\x00")
	path := createTempFile(t, data)

	mt, err := DetectMIME(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mt != MIMETypeGIF {
		t.Errorf("expected %s, got %s", MIMETypeGIF, mt)
	}
}

func TestDetectMIME_GIF89a(t *testing.T) {
	data := []byte("GIF89a\x00\x00\x00\x00\x00\x00")
	path := createTempFile(t, data)

	mt, err := DetectMIME(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mt != MIMETypeGIF {
		t.Errorf("expected %s, got %s", MIMETypeGIF, mt)
	}
}

func TestDetectMIME_PNG(t *testing.T) {
	data := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x00}
	path := createTempFile(t, data)

	mt, err := DetectMIME(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mt != MIMETypeUnsupported {
		t.Errorf("expected %s, got %s", MIMETypeUnsupported, mt)
	}
}

func TestDetectMIME_TextFile(t *testing.T) {
	data := []byte("hello world!")
	path := createTempFile(t, data)

	mt, err := DetectMIME(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mt != MIMETypeUnsupported {
		t.Errorf("expected %s, got %s", MIMETypeUnsupported, mt)
	}
}

func TestDetectMIME_EmptyFile(t *testing.T) {
	path := createTempFile(t, []byte{})

	_, err := DetectMIME(path)
	if err == nil {
		t.Fatal("expected error for empty file, got nil")
	}
}

func TestDetectMIME_TruncatedFile(t *testing.T) {
	data := []byte{0x00, 0x01, 0x02}
	path := createTempFile(t, data)

	mt, err := DetectMIME(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mt != MIMETypeUnsupported {
		t.Errorf("expected %s, got %s", MIMETypeUnsupported, mt)
	}
}

func TestDetectMIME_NonexistentFile(t *testing.T) {
	_, err := DetectMIME("/nonexistent/path/to/file.mp4")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

func TestIsVideo(t *testing.T) {
	tests := []struct {
		mt   MIMEType
		want bool
	}{
		{MIMETypeMP4, true},
		{MIMETypeMOV, true},
		{MIMETypeGIF, false},
		{MIMETypeUnsupported, false},
	}
	for _, tt := range tests {
		if got := IsVideo(tt.mt); got != tt.want {
			t.Errorf("IsVideo(%s) = %v, want %v", tt.mt, got, tt.want)
		}
	}
}

func TestIsGIF(t *testing.T) {
	tests := []struct {
		mt   MIMEType
		want bool
	}{
		{MIMETypeGIF, true},
		{MIMETypeMP4, false},
		{MIMETypeMOV, false},
		{MIMETypeUnsupported, false},
	}
	for _, tt := range tests {
		if got := IsGIF(tt.mt); got != tt.want {
			t.Errorf("IsGIF(%s) = %v, want %v", tt.mt, got, tt.want)
		}
	}
}

func TestIsUnsupported(t *testing.T) {
	tests := []struct {
		mt   MIMEType
		want bool
	}{
		{MIMETypeUnsupported, true},
		{MIMETypeMP4, false},
		{MIMETypeMOV, false},
		{MIMETypeGIF, false},
	}
	for _, tt := range tests {
		if got := IsUnsupported(tt.mt); got != tt.want {
			t.Errorf("IsUnsupported(%s) = %v, want %v", tt.mt, got, tt.want)
		}
	}
}
