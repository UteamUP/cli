package scanner

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/uteamup/cli/internal/imageanalyzer/imageutil"
	"github.com/uteamup/cli/internal/imageanalyzer/models"
)

// ImageScanner walks an image folder, extracts metadata, and discovers images.
type ImageScanner struct {
	supportedFormats map[string]bool
	imageFolder      string
	recursive        bool
	maxDimension     int
	maxFileSizeMB    int
}

// NewScanner creates a new ImageScanner.
//
// supportedFormats should contain extensions with leading dots (e.g. ".jpg").
// If supportedFormats is empty, a sensible default set is used.
func NewScanner(imageFolder string, recursive bool, supportedFormats []string, maxDimension, maxFileSizeMB int) *ImageScanner {
	fmts := make(map[string]bool, len(supportedFormats))
	for _, f := range supportedFormats {
		ext := strings.ToLower(f)
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		fmts[ext] = true
	}

	return &ImageScanner{
		supportedFormats: fmts,
		imageFolder:      imageFolder,
		recursive:        recursive,
		maxDimension:     maxDimension,
		maxFileSizeMB:    maxFileSizeMB,
	}
}

// ScanFolder walks the configured image folder and returns metadata for
// every valid image found. It respects the recursive flag, filters by
// supported extensions (case-insensitive), validates images via
// imageutil.IsValidImage, computes hashes, and extracts EXIF data.
// Results are sorted by file path.
func (s *ImageScanner) ScanFolder() ([]models.ImageInfo, error) {
	info, err := os.Stat(s.imageFolder)
	if err != nil || !info.IsDir() {
		return nil, fmt.Errorf("image folder not found or not a directory: %s", s.imageFolder)
	}

	var images []models.ImageInfo

	err = filepath.WalkDir(s.imageFolder, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			log.Printf("warning: walk error at %s: %v", path, walkErr)
			return nil
		}

		// If not recursive, skip subdirectories (but not the root itself).
		if !s.recursive && d.IsDir() && path != s.imageFolder {
			return filepath.SkipDir
		}

		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !s.supportedFormats[ext] {
			return nil
		}

		if !imageutil.IsValidImage(path) {
			log.Printf("skipped invalid image: %s", path)
			return nil
		}

		fi, err := d.Info()
		if err != nil {
			log.Printf("warning: cannot stat %s: %v", path, err)
			return nil
		}

		sha256, phash := ComputeSHA256Must(path), ""
		if ph, err := ComputePerceptualHash(path); err == nil {
			phash = ph
		}

		img := models.ImageInfo{
			Path:           path,
			Filename:       filepath.Base(path),
			Extension:      ext,
			FileSizeBytes:  fi.Size(),
			SHA256Hash:     sha256,
			PerceptualHash: phash,
			EXIFMetadata:   ExtractEXIF(path),
		}
		images = append(images, img)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scan folder %s: %w", s.imageFolder, err)
	}

	sort.Slice(images, func(i, j int) bool {
		return images[i].Path < images[j].Path
	})

	log.Printf("scan complete: %d images found in %s", len(images), s.imageFolder)
	return images, nil
}

// ComputeSHA256Must computes the SHA-256 hash and returns the hex string.
// On error it returns an empty string.
func ComputeSHA256Must(filePath string) string {
	h, err := ComputeSHA256(filePath)
	if err != nil {
		log.Printf("sha256 failed for %s: %v", filePath, err)
		return ""
	}
	return h
}
