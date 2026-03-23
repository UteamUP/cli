package fileutil

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
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
}

// NewScanner creates a new Scanner with the given file size limit in megabytes.
func NewScanner(maxFileSizeMB int) *Scanner {
	return &Scanner{maxFileSizeMB: maxFileSizeMB}
}

// ScanPath scans the given path (file or directory) for video files.
// It uses DetectMIME to classify each file, routes GIFs to GIFFiles,
// enforces the file size limit, and skips unsupported formats.
// If path is a single file, it validates just that file.
// If path is a directory, it walks recursively.
func (s *Scanner) ScanPath(path string) (*ScanResult, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat path: %w", err)
	}

	result := &ScanResult{}

	if !info.IsDir() {
		s.processFile(path, info, result)
		return result, nil
	}

	err = filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		fi, err := d.Info()
		if err != nil {
			log.Printf("skip %s: %v", p, err)
			return nil
		}
		s.processFile(p, fi, result)
		return nil
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
func (s *Scanner) processFile(path string, info os.FileInfo, result *ScanResult) {
	mt, err := DetectMIME(path)
	if err != nil {
		log.Printf("skip %s: MIME detection failed: %v", filepath.Base(path), err)
		result.Skipped = append(result.Skipped, path)
		return
	}

	if IsUnsupported(mt) {
		log.Printf("skip %s: unsupported format", filepath.Base(path))
		result.Skipped = append(result.Skipped, path)
		return
	}

	if IsGIF(mt) {
		result.GIFFiles = append(result.GIFFiles, path)
		return
	}

	// Check file size limit for video files
	maxBytes := int64(s.maxFileSizeMB) * 1024 * 1024
	if info.Size() > maxBytes {
		log.Printf("skip %s: file size %d bytes exceeds limit %d MB",
			filepath.Base(path), info.Size(), s.maxFileSizeMB)
		result.Skipped = append(result.Skipped, path)
		return
	}

	result.Videos = append(result.Videos, VideoFile{
		Path:      path,
		Filename:  filepath.Base(path),
		MIMEType:  mt,
		SizeBytes: info.Size(),
	})
}
