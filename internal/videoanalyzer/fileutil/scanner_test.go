package fileutil

import (
	"os"
	"path/filepath"
	"testing"
)

// createTempDir creates a temporary directory for testing.
func createTempDir(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}

// writeTempFile writes the given bytes to a named file inside the directory.
func writeTempFile(t *testing.T, dir, name string, data []byte) {
	t.Helper()
	path := filepath.Join(dir, name)
	// Ensure parent directories exist for nested paths
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("failed to create parent dirs: %v", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("failed to write temp file %s: %v", name, err)
	}
}

func makeMP4Bytes() []byte {
	data := make([]byte, 12)
	copy(data[4:8], "ftyp")
	copy(data[8:12], "isom")
	return data
}

func makeMOVBytes() []byte {
	data := make([]byte, 12)
	copy(data[4:8], "ftyp")
	copy(data[8:12], "qt  ")
	return data
}

func makeGIFBytes() []byte {
	data := make([]byte, 12)
	copy(data[0:6], "GIF89a")
	return data
}

func makePNGBytes() []byte {
	return []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x00}
}

func makeTXTBytes() []byte {
	return []byte("hello world!")
}

func TestScanPath_SingleMP4File(t *testing.T) {
	dir := createTempDir(t)
	writeTempFile(t, dir, "test.mp4", makeMP4Bytes())

	scanner := NewScanner(100)
	result, err := scanner.ScanPath(filepath.Join(dir, "test.mp4"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Videos) != 1 {
		t.Errorf("expected 1 video, got %d", len(result.Videos))
	}
	if len(result.GIFFiles) != 0 {
		t.Errorf("expected 0 GIFs, got %d", len(result.GIFFiles))
	}
	if len(result.Skipped) != 0 {
		t.Errorf("expected 0 skipped, got %d", len(result.Skipped))
	}
}

func TestScanPath_SingleGIFFile(t *testing.T) {
	dir := createTempDir(t)
	writeTempFile(t, dir, "animation.gif", makeGIFBytes())

	scanner := NewScanner(100)
	result, err := scanner.ScanPath(filepath.Join(dir, "animation.gif"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Videos) != 0 {
		t.Errorf("expected 0 videos, got %d", len(result.Videos))
	}
	if len(result.GIFFiles) != 1 {
		t.Errorf("expected 1 GIF, got %d", len(result.GIFFiles))
	}
	if len(result.Skipped) != 0 {
		t.Errorf("expected 0 skipped, got %d", len(result.Skipped))
	}
}

func TestScanPath_UnsupportedFile(t *testing.T) {
	dir := createTempDir(t)
	writeTempFile(t, dir, "image.png", makePNGBytes())

	scanner := NewScanner(100)
	result, err := scanner.ScanPath(filepath.Join(dir, "image.png"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Videos) != 0 {
		t.Errorf("expected 0 videos, got %d", len(result.Videos))
	}
	if len(result.GIFFiles) != 0 {
		t.Errorf("expected 0 GIFs, got %d", len(result.GIFFiles))
	}
	if len(result.Skipped) != 1 {
		t.Errorf("expected 1 skipped, got %d", len(result.Skipped))
	}
}

func TestScanPath_DirectoryWithMixedFiles(t *testing.T) {
	dir := createTempDir(t)
	writeTempFile(t, dir, "video.mp4", makeMP4Bytes())
	writeTempFile(t, dir, "movie.mov", makeMOVBytes())
	writeTempFile(t, dir, "animation.gif", makeGIFBytes())
	writeTempFile(t, dir, "image.png", makePNGBytes())
	writeTempFile(t, dir, "readme.txt", makeTXTBytes())

	scanner := NewScanner(100)
	result, err := scanner.ScanPath(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Videos) != 2 {
		t.Errorf("expected 2 videos, got %d", len(result.Videos))
	}
	if len(result.GIFFiles) != 1 {
		t.Errorf("expected 1 GIF, got %d", len(result.GIFFiles))
	}
	if len(result.Skipped) != 2 {
		t.Errorf("expected 2 skipped, got %d", len(result.Skipped))
	}
}

func TestScanPath_EmptyDirectory(t *testing.T) {
	dir := createTempDir(t)

	scanner := NewScanner(100)
	result, err := scanner.ScanPath(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Videos) != 0 {
		t.Errorf("expected 0 videos, got %d", len(result.Videos))
	}
	if len(result.GIFFiles) != 0 {
		t.Errorf("expected 0 GIFs, got %d", len(result.GIFFiles))
	}
	if len(result.Skipped) != 0 {
		t.Errorf("expected 0 skipped, got %d", len(result.Skipped))
	}
}

func TestScanPath_FileSizeLimit(t *testing.T) {
	dir := createTempDir(t)

	// Create a file that exceeds the 1 MB limit (2 MB)
	bigData := make([]byte, 2*1024*1024)
	copy(bigData[4:8], "ftyp")
	copy(bigData[8:12], "isom")
	writeTempFile(t, dir, "big.mp4", bigData)

	// Create a file within the limit (500 KB)
	smallData := make([]byte, 500*1024)
	copy(smallData[4:8], "ftyp")
	copy(smallData[8:12], "isom")
	writeTempFile(t, dir, "small.mp4", smallData)

	scanner := NewScanner(1) // 1 MB limit
	result, err := scanner.ScanPath(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Videos) != 1 {
		t.Errorf("expected 1 video (small.mp4), got %d", len(result.Videos))
	}
	if len(result.Skipped) != 1 {
		t.Errorf("expected 1 skipped (big.mp4), got %d", len(result.Skipped))
	}
	if len(result.Videos) == 1 && result.Videos[0].Filename != "small.mp4" {
		t.Errorf("expected small.mp4 to be included, got %s", result.Videos[0].Filename)
	}
}

func TestScanPath_NonexistentPath(t *testing.T) {
	scanner := NewScanner(100)
	_, err := scanner.ScanPath("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Fatal("expected error for nonexistent path, got nil")
	}
}

func TestScanPath_RecursiveSubdirectories(t *testing.T) {
	dir := createTempDir(t)
	writeTempFile(t, dir, "root.mp4", makeMP4Bytes())
	writeTempFile(t, dir, "sub1/video1.mov", makeMOVBytes())
	writeTempFile(t, dir, "sub1/sub2/video2.mp4", makeMP4Bytes())

	scanner := NewScanner(100)
	result, err := scanner.ScanPath(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Videos) != 3 {
		t.Errorf("expected 3 videos across nested dirs, got %d", len(result.Videos))
	}

	// Verify sorted by path
	for i := 1; i < len(result.Videos); i++ {
		if result.Videos[i].Path < result.Videos[i-1].Path {
			t.Errorf("videos not sorted by path: %s came after %s",
				result.Videos[i].Path, result.Videos[i-1].Path)
		}
	}
}
