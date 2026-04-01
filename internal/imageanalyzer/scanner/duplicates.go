package scanner

import (
	"log"

	"github.com/uteamup/cli/internal/imageanalyzer/models"
)

// DetectDuplicates finds duplicate images by SHA-256 hash.
// The first occurrence (by path order in the input slice) is kept.
// Returns the unique images and a list of duplicate pairs where each
// pair is [kept_path, duplicate_path].
func DetectDuplicates(images []models.ImageInfo) (unique []models.ImageInfo, duplicatePairs [][2]string) {
	seen := make(map[string]models.ImageInfo, len(images))

	for _, img := range images {
		if img.SHA256Hash == "" {
			// Cannot deduplicate without a hash; treat as unique.
			unique = append(unique, img)
			continue
		}

		if kept, exists := seen[img.SHA256Hash]; exists {
			duplicatePairs = append(duplicatePairs, [2]string{kept.Path, img.Path})
			hashPreview := img.SHA256Hash
			if len(hashPreview) > 12 {
				hashPreview = hashPreview[:12]
			}
			log.Printf("duplicate detected: %s is a duplicate of %s (hash %s)",
				img.Filename, kept.Filename, hashPreview)
		} else {
			seen[img.SHA256Hash] = img
			unique = append(unique, img)
		}
	}

	log.Printf("duplicate detection complete: %d total, %d unique, %d duplicates",
		len(images), len(unique), len(duplicatePairs))

	return unique, duplicatePairs
}
