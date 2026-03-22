package exporter

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/uteamup/cli/internal/imageanalyzer/imageutil"
	"github.com/uteamup/cli/internal/imageanalyzer/models"
)

// RenameImages copies images to the renamed-images folder with descriptive
// filenames following the pattern: {entity_type}_{sanitized_name}_{seq:03d}_{YYYYMMDD}.{ext}.
// Originals are never moved. Returns a mapping of original_path -> new_path.
func (e *CSVExporter) RenameImages(groups []models.ImageGroup) (map[string]string, error) {
	if !e.renameImages {
		return map[string]string{}, nil
	}

	mapping := make(map[string]string)
	seqCounters := make(map[string]int)
	today := time.Now().Format("20060102")

	for _, group := range groups {
		etype := string(group.Primary.Classification.PrimaryType)
		name := imageutil.SanitizeFilename(group.Primary.ExtractedData.GetName())
		if name == "" {
			name = "unnamed"
		}

		for _, imgPath := range group.AllImagePaths() {
			if _, err := os.Stat(imgPath); os.IsNotExist(err) {
				continue
			}

			ext := strings.TrimPrefix(filepath.Ext(imgPath), ".")
			key := fmt.Sprintf("%s_%s", etype, name)
			seq := seqCounters[key]
			if seq == 0 {
				seq = 1
			}

			newName := fmt.Sprintf("%s_%s_%03d_%s.%s", etype, name, seq, today, ext)
			dest := filepath.Join(e.renamedImagesFolder, newName)

			// Handle collisions by incrementing seq.
			for {
				if _, err := os.Stat(dest); os.IsNotExist(err) {
					break
				}
				seq++
				newName = fmt.Sprintf("%s_%s_%03d_%s.%s", etype, name, seq, today, ext)
				dest = filepath.Join(e.renamedImagesFolder, newName)
			}

			seqCounters[key] = seq + 1

			if err := copyFile(imgPath, dest); err != nil {
				return nil, fmt.Errorf("copying %s to %s: %w", imgPath, dest, err)
			}
			mapping[imgPath] = dest
		}
	}
	return mapping, nil
}

// copyFile copies src to dst using io.Copy, preserving content only.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
