package scanner

import "github.com/uteamup/cli/internal/imageanalyzer/imageutil"

// sanitizeViaImageutil is a test helper that calls imageutil.SanitizeFilename.
func sanitizeViaImageutil(name string) string {
	return imageutil.SanitizeFilename(name)
}
