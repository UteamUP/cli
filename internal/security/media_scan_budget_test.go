package security

import (
	"strings"
	"testing"
)

func TestMediaScanBudgetBoundsEntriesFilesAndBytes(t *testing.T) {
	t.Run("entry count", func(t *testing.T) {
		budget := &MediaScanBudget{maximumEntries: 2}
		if err := budget.AddEntry(); err != nil {
			t.Fatal(err)
		}
		if err := budget.AddEntry(); err != nil {
			t.Fatal(err)
		}
		if err := budget.AddEntry(); err == nil || !strings.Contains(err.Error(), "2-entry") {
			t.Fatalf("third entry error = %v, want traversal limit", err)
		}
	})

	t.Run("file count", func(t *testing.T) {
		budget := &MediaScanBudget{maximumFiles: 1, maximumBytes: 10}
		if err := budget.AddMediaFile(5); err != nil {
			t.Fatal(err)
		}
		if err := budget.AddMediaFile(1); err == nil || !strings.Contains(err.Error(), "1-file") {
			t.Fatalf("second file error = %v, want file limit", err)
		}
	})

	t.Run("aggregate bytes", func(t *testing.T) {
		budget := &MediaScanBudget{maximumFiles: 3, maximumBytes: 10}
		if err := budget.AddMediaFile(6); err != nil {
			t.Fatal(err)
		}
		if err := budget.AddMediaFile(4); err != nil {
			t.Fatal(err)
		}
		if err := budget.AddMediaFile(1); err == nil || !strings.Contains(err.Error(), "10-byte") {
			t.Fatalf("third file error = %v, want aggregate byte limit", err)
		}
	})
}
