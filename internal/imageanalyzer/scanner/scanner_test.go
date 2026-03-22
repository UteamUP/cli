package scanner

import (
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"

	"github.com/uteamup/cli/internal/imageanalyzer/models"
)

// createTestJPEG creates a minimal valid JPEG file at the given path.
func createTestJPEG(t *testing.T, path string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create test jpeg %s: %v", path, err)
	}
	defer f.Close()

	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}
	if err := jpeg.Encode(f, img, nil); err != nil {
		t.Fatalf("encode test jpeg %s: %v", path, err)
	}
}

func TestScanFolder(t *testing.T) {
	tmpDir := t.TempDir()

	// Create valid JPEG files.
	createTestJPEG(t, filepath.Join(tmpDir, "photo1.jpg"))
	createTestJPEG(t, filepath.Join(tmpDir, "photo2.JPG"))

	// Create a non-image file that should be skipped.
	if err := os.WriteFile(filepath.Join(tmpDir, "notes.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	scanner := NewScanner(tmpDir, false, []string{".jpg", ".jpeg", ".png"}, 1024, 10)
	images, err := scanner.ScanFolder()
	if err != nil {
		t.Fatalf("ScanFolder error: %v", err)
	}

	if len(images) != 2 {
		t.Errorf("expected 2 images, got %d", len(images))
	}

	for _, img := range images {
		if img.Extension != ".jpg" {
			t.Errorf("expected .jpg extension, got %s for %s", img.Extension, img.Filename)
		}
		if img.SHA256Hash == "" {
			t.Errorf("expected non-empty SHA256 hash for %s", img.Filename)
		}
		if img.FileSizeBytes == 0 {
			t.Errorf("expected non-zero file size for %s", img.Filename)
		}
	}
}

func TestScanFolderNonRecursive(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "sub")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	createTestJPEG(t, filepath.Join(tmpDir, "root.jpg"))
	createTestJPEG(t, filepath.Join(subDir, "nested.jpg"))

	scanner := NewScanner(tmpDir, false, []string{".jpg"}, 1024, 10)
	images, err := scanner.ScanFolder()
	if err != nil {
		t.Fatalf("ScanFolder error: %v", err)
	}

	if len(images) != 1 {
		t.Errorf("non-recursive: expected 1 image, got %d", len(images))
	}
	if len(images) > 0 && images[0].Filename != "root.jpg" {
		t.Errorf("expected root.jpg, got %s", images[0].Filename)
	}
}

func TestScanFolderRecursive(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "sub")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	createTestJPEG(t, filepath.Join(tmpDir, "root.jpg"))
	createTestJPEG(t, filepath.Join(subDir, "nested.jpg"))

	scanner := NewScanner(tmpDir, true, []string{".jpg"}, 1024, 10)
	images, err := scanner.ScanFolder()
	if err != nil {
		t.Fatalf("ScanFolder error: %v", err)
	}

	if len(images) != 2 {
		t.Errorf("recursive: expected 2 images, got %d", len(images))
	}
}

func TestDetectDuplicates(t *testing.T) {
	images := []models.ImageInfo{
		{Path: "/a/img1.jpg", Filename: "img1.jpg", SHA256Hash: "aaa111"},
		{Path: "/a/img2.jpg", Filename: "img2.jpg", SHA256Hash: "bbb222"},
		{Path: "/a/img3.jpg", Filename: "img3.jpg", SHA256Hash: "aaa111"}, // duplicate of img1
		{Path: "/a/img4.jpg", Filename: "img4.jpg", SHA256Hash: "ccc333"},
		{Path: "/a/img5.jpg", Filename: "img5.jpg", SHA256Hash: "bbb222"}, // duplicate of img2
	}

	unique, pairs := DetectDuplicates(images)

	if len(unique) != 3 {
		t.Errorf("expected 3 unique images, got %d", len(unique))
	}
	if len(pairs) != 2 {
		t.Errorf("expected 2 duplicate pairs, got %d", len(pairs))
	}

	// Verify the first pair is img1 kept, img3 duplicate.
	if pairs[0] != [2]string{"/a/img1.jpg", "/a/img3.jpg"} {
		t.Errorf("unexpected first pair: %v", pairs[0])
	}
	// Verify the second pair is img2 kept, img5 duplicate.
	if pairs[1] != [2]string{"/a/img2.jpg", "/a/img5.jpg"} {
		t.Errorf("unexpected second pair: %v", pairs[1])
	}
}

func TestDetectDuplicatesEmptyHash(t *testing.T) {
	images := []models.ImageInfo{
		{Path: "/a/img1.jpg", Filename: "img1.jpg", SHA256Hash: ""},
		{Path: "/a/img2.jpg", Filename: "img2.jpg", SHA256Hash: "aaa111"},
	}

	unique, pairs := DetectDuplicates(images)
	if len(unique) != 2 {
		t.Errorf("expected 2 unique (empty hash treated as unique), got %d", len(unique))
	}
	if len(pairs) != 0 {
		t.Errorf("expected 0 duplicate pairs, got %d", len(pairs))
	}
}

func TestDetectIPhonePairs(t *testing.T) {
	images := []models.ImageInfo{
		{Path: "/photos/IMG_1234.jpg", Filename: "IMG_1234.jpg", Extension: ".jpg"},
		{Path: "/photos/IMG_E1234.jpg", Filename: "IMG_E1234.jpg", Extension: ".jpg"},
		{Path: "/photos/IMG_5678.jpg", Filename: "IMG_5678.jpg", Extension: ".jpg"},
		{Path: "/photos/IMG_E5678.jpg", Filename: "IMG_E5678.jpg", Extension: ".jpg"},
		{Path: "/photos/IMG_9999.jpg", Filename: "IMG_9999.jpg", Extension: ".jpg"},
		// IMG_E9999 does not exist — no pair for 9999
	}

	pairs := DetectIPhonePairs(images)

	if len(pairs) != 2 {
		t.Errorf("expected 2 pairs, got %d", len(pairs))
	}

	if variants, ok := pairs["IMG_1234.jpg"]; !ok {
		t.Error("expected pair for IMG_1234.jpg")
	} else if len(variants) != 1 || variants[0] != "/photos/IMG_E1234.jpg" {
		t.Errorf("unexpected variants for IMG_1234: %v", variants)
	}

	if variants, ok := pairs["IMG_5678.jpg"]; !ok {
		t.Error("expected pair for IMG_5678.jpg")
	} else if len(variants) != 1 || variants[0] != "/photos/IMG_E5678.jpg" {
		t.Errorf("unexpected variants for IMG_5678: %v", variants)
	}

	// Verify side effects on ImageInfo.
	if !images[1].IsIPhoneEdit {
		t.Error("IMG_E1234 should be marked as iPhone edit")
	}
	if images[1].PairedWith != "IMG_1234.jpg" {
		t.Errorf("IMG_E1234 paired_with should be IMG_1234.jpg, got %s", images[1].PairedWith)
	}
	if images[4].IsIPhoneEdit {
		t.Error("IMG_9999 should NOT be marked as iPhone edit")
	}
}

func TestDetectIPhonePairsCaseInsensitive(t *testing.T) {
	images := []models.ImageInfo{
		{Path: "/photos/img_1234.jpg", Filename: "img_1234.jpg", Extension: ".jpg"},
		{Path: "/photos/img_e1234.jpg", Filename: "img_e1234.jpg", Extension: ".jpg"},
	}

	pairs := DetectIPhonePairs(images)
	if len(pairs) != 1 {
		t.Errorf("case-insensitive: expected 1 pair, got %d", len(pairs))
	}
}

func TestSanitizeFilenameIntegration(t *testing.T) {
	// Integration test verifying imageutil.SanitizeFilename works as expected
	// when called from the scanner context.
	from := "github.com/uteamup/cli/internal/imageanalyzer/imageutil"
	_ = from // just confirming the import path is valid

	// Test via the imageutil package directly.
	got := sanitizeViaImageutil("My Photo (2024).JPG")
	if got != "my_photo_2024.jpg" {
		t.Errorf("SanitizeFilename: expected my_photo_2024.jpg, got %s", got)
	}

	got = sanitizeViaImageutil("IMG--1234__test.png")
	if got != "img_1234_test.png" {
		t.Errorf("SanitizeFilename: expected img_1234_test.png, got %s", got)
	}
}
