package scanner

import (
	"log"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/uteamup/cli/internal/imageanalyzer/models"
)

var (
	// iPhoneEditRE matches iPhone edited variants: IMG_EXXXX
	iPhoneEditRE = regexp.MustCompile(`(?i)^IMG_E(\d{4})`)
	// iPhoneOriginalRE matches iPhone originals: IMG_XXXX
	iPhoneOriginalRE = regexp.MustCompile(`(?i)^IMG_(\d{4})`)
)

// DetectIPhonePairs finds iPhone original/edited image pairs.
//
// iPhones save the original as IMG_XXXX.jpg and the edited version as
// IMG_EXXXX.jpg. This function matches them by the 4-digit number.
//
// The returned map keys are the original filenames, and the values are
// slices of edit-variant file paths.
//
// Side-effect: sets IsIPhoneEdit and PairedWith on matching ImageInfo
// objects in the input slice.
func DetectIPhonePairs(images []models.ImageInfo) map[string][]string {
	// Build lookups: number -> ImageInfo index for originals and edits.
	originals := make(map[string]int)    // number -> index in images
	edits := make(map[string][]int)      // number -> indices in images

	for i := range images {
		stem := strings.TrimSuffix(images[i].Filename, filepath.Ext(images[i].Filename))

		if m := iPhoneEditRE.FindStringSubmatch(stem); m != nil {
			number := m[1]
			edits[number] = append(edits[number], i)
			continue
		}

		if m := iPhoneOriginalRE.FindStringSubmatch(stem); m != nil {
			number := m[1]
			originals[number] = i
		}
	}

	pairs := make(map[string][]string)

	for number, editIndices := range edits {
		origIdx, exists := originals[number]
		if !exists {
			continue
		}

		primary := &images[origIdx]
		var variantPaths []string

		for _, ei := range editIndices {
			images[ei].IsIPhoneEdit = true
			images[ei].PairedWith = primary.Filename
			variantPaths = append(variantPaths, images[ei].Path)
		}

		pairs[primary.Filename] = variantPaths
		log.Printf("iPhone edit pair: %s -> %v", primary.Filename, variantPaths)
	}

	log.Printf("iPhone edit detection complete: %d pairs found", len(pairs))
	return pairs
}
