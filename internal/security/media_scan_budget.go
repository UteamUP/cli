package security

import "fmt"

const (
	// MaximumMediaScanEntries bounds directory traversal even when most entries
	// are unsupported files.
	MaximumMediaScanEntries = 10000
	// MaximumMediaScanFiles bounds analysis requests and retained scan results.
	MaximumMediaScanFiles = 100
	// MaximumMediaScanBytes bounds aggregate input per scan invocation.
	MaximumMediaScanBytes int64 = 1 << 30
)

// MediaScanBudget caps directory work and aggregate media input for one scan.
type MediaScanBudget struct {
	entries int
	files   int
	bytes   int64

	maximumEntries int
	maximumFiles   int
	maximumBytes   int64
}

// NewMediaScanBudget returns the standard bounded media-scan budget.
func NewMediaScanBudget() *MediaScanBudget {
	return &MediaScanBudget{
		maximumEntries: MaximumMediaScanEntries,
		maximumFiles:   MaximumMediaScanFiles,
		maximumBytes:   MaximumMediaScanBytes,
	}
}

// AddEntry charges one filesystem entry against the traversal limit.
func (budget *MediaScanBudget) AddEntry() error {
	if budget.entries >= budget.maximumEntries {
		return fmt.Errorf("media scan exceeds the %d-entry traversal limit", budget.maximumEntries)
	}
	budget.entries++
	return nil
}

// AddMediaFile charges one candidate and its size against analysis limits.
func (budget *MediaScanBudget) AddMediaFile(sizeBytes int64) error {
	if sizeBytes <= 0 {
		return fmt.Errorf("media file size must be positive")
	}
	if budget.files >= budget.maximumFiles {
		return fmt.Errorf("media scan exceeds the %d-file analysis limit", budget.maximumFiles)
	}
	if sizeBytes > budget.maximumBytes-budget.bytes {
		return fmt.Errorf("media scan exceeds the %d-byte aggregate input limit", budget.maximumBytes)
	}
	budget.files++
	budget.bytes += sizeBytes
	return nil
}
