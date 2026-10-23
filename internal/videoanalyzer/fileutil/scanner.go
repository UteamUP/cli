package fileutil

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/uteamup/cli/internal/security"
)

// VideoFile holds metadata about a discovered video file.
type VideoFile struct {
	Path      string
	Filename  string
	MIMEType  MIMEType
	SizeBytes int64
}

// ScanResult holds the results of scanning a path for video files.
type ScanResult struct {
	Videos   []VideoFile
	GIFFiles []string // GIF file paths to route to image analyzer
	Skipped  []string // Unsupported files that were skipped
}

// Scanner walks a directory or validates a single file, filtering by MIME type.
type Scanner struct {
	maxFileSizeMB int
	recursive     bool
}

// NewScanner creates a new Scanner with the given file size limit in megabytes.
func NewScanner(maxFileSizeMB int, recursive bool) *Scanner {
	return &Scanner{maxFileSizeMB: maxFileSizeMB, recursive: recursive}
}

// ScanPath scans the given path (file or directory) for video files.
// It uses DetectMIME to classify each file, routes GIFs to GIFFiles,
// enforces the file size limit, and skips unsupported formats.
// If path is a single file, it validates just that file.
// If path is a directory, it follows the configured recursion boundary.
func (s *Scanner) ScanPath(path string) (*ScanResult, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, fmt.Errorf("stat path: %w", err)
	}

	result := &ScanResult{}
	budget := security.NewMediaScanBudget()

	if !info.IsDir() {
		if err := budget.AddEntry(); err != nil {
			return nil, err
		}
		if err := s.processFile(path, info, result, budget); err != nil {
			return nil, err
		}
		return result, nil
	}

	err = security.WalkMediaTree(path, s.recursive, budget, func(p string, d fs.DirEntry) error {
		fi, err := d.Info()
		if err != nil {
			log.Printf("skip %s: %s", security.SafeText(p), security.SafeText(err.Error()))
			return nil
		}
		return s.processFile(p, fi, result, budget)
	})
	if err != nil {
		return nil, fmt.Errorf("walk directory: %w", err)
	}

	sort.Slice(result.Videos, func(i, j int) bool {
		return result.Videos[i].Path < result.Videos[j].Path
	})

	return result, nil
}

// processFile classifies a single file and appends it to the appropriate
// category in the scan result.
func (s *Scanner) processFile(
	path string,
	info os.FileInfo,
	result *ScanResult,
	budget *security.MediaScanBudget,
) error {
	if !info.Mode().IsRegular() {
		log.Printf("skip %s: upload source is not a regular file", security.SafeText(filepath.Base(path)))
		result.Skipped = append(result.Skipped, path)
		return nil
	}
	maxBytes := int64(s.maxFileSizeMB) * 1024 * 1024
	if info.Size() <= 0 || info.Size() > maxBytes {
		log.Printf("skip %s: file size is outside the 1-%d MB limit", security.SafeText(filepath.Base(path)), s.maxFileSizeMB)
		result.Skipped = append(result.Skipped, path)
		return nil
	}

	security.BindSource(path, info)
	mt, err := DetectMIME(path)
	if err != nil {
		log.Printf("skip %s: MIME detection failed: %s", security.SafeText(filepath.Base(path)), security.SafeText(err.Error()))
		result.Skipped = append(result.Skipped, path)
		return nil
	}

	if IsUnsupported(mt) {
		log.Printf("skip %s: unsupported format", security.SafeText(filepath.Base(path)))
		result.Skipped = append(result.Skipped, path)
		return nil
	}
	if err := budget.AddMediaFile(info.Size()); err != nil {
		return fmt.Errorf("media scan budget exceeded: %w", err)
	}

	if IsGIF(mt) {
		result.GIFFiles = append(result.GIFFiles, path)
		return nil
	}

	result.Videos = append(result.Videos, VideoFile{
		Path:      path,
		Filename:  filepath.Base(path),
		MIMEType:  mt,
		SizeBytes: info.Size(),
	})
	return nil
}
