package imageutil

import (
	"regexp"
	"strings"
)

var (
	reNonAlphanumUnderscore = regexp.MustCompile(`[^a-z0-9_]`)
	reMultipleUnderscores   = regexp.MustCompile(`_{2,}`)
)

// SanitizeFilename cleans a filename for safe, consistent use:
//   - lowercased
//   - spaces and hyphens replaced with underscores
//   - all characters except [a-z0-9_] removed from the stem
//   - multiple underscores collapsed
//   - leading/trailing underscores trimmed from stem
//   - extension is preserved untouched
func SanitizeFilename(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))

	// Split stem from extension.
	var stem, ext string
	if dotIdx := strings.LastIndex(name, "."); dotIdx > 0 {
		stem = name[:dotIdx]
		ext = name[dotIdx:] // includes the dot
	} else {
		stem = name
		ext = ""
	}

	// Replace spaces and hyphens with underscores.
	stem = strings.ReplaceAll(stem, " ", "_")
	stem = strings.ReplaceAll(stem, "-", "_")

	// Remove everything except [a-z0-9_].
	stem = reNonAlphanumUnderscore.ReplaceAllString(stem, "")

	// Collapse multiple underscores.
	stem = reMultipleUnderscores.ReplaceAllString(stem, "_")

	// Trim leading/trailing underscores.
	stem = strings.Trim(stem, "_")

	return stem + ext
}
