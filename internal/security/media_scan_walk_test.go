package security

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWalkMediaTreeStopsAtBudgetAcrossLargeSingleDirectory(t *testing.T) {
	directory := t.TempDir()
	for index := range mediaScanReadDirBatch + 2 {
		path := filepath.Join(directory, fmt.Sprintf("entry-%03d", index))
		if err := os.WriteFile(path, nil, 0600); err != nil {
			t.Fatalf("WriteFile(%q): %v", path, err)
		}
	}

	budget := NewMediaScanBudget()
	budget.maximumEntries = mediaScanReadDirBatch + 1 // Includes the root directory.
	visited := 0
	err := WalkMediaTree(directory, true, budget, func(string, os.DirEntry) error {
		visited++
		return nil
	})
	if err == nil || !strings.Contains(err.Error(), "entry traversal limit") {
		t.Fatalf("WalkMediaTree() error = %v, want entry traversal limit", err)
	}
	if visited != mediaScanReadDirBatch {
		t.Fatalf("WalkMediaTree() visited %d entries, want exactly %d before stopping", visited, mediaScanReadDirBatch)
	}
}

func TestWalkMediaTreeDoesNotFollowSymlinkedChildDirectory(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "outside.txt"), nil, 0600); err != nil {
		t.Fatalf("WriteFile(outside file): %v", err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "linked-directory")); err != nil {
		t.Skipf("symlink setup is unavailable: %v", err)
	}

	var visited []string
	err := WalkMediaTree(root, true, NewMediaScanBudget(), func(path string, _ os.DirEntry) error {
		visited = append(visited, path)
		return nil
	})
	if err != nil {
		t.Fatalf("WalkMediaTree() error = %v", err)
	}
	for _, path := range visited {
		if filepath.Base(path) == "outside.txt" {
			t.Fatalf("WalkMediaTree() followed a symlinked directory to %q", path)
		}
	}
}

func TestWalkMediaTreeStopsAtDepthLimit(t *testing.T) {
	root := t.TempDir()
	current := root
	for depth := 0; depth < maximumMediaScanDepth; depth++ {
		current = filepath.Join(current, "d")
		if err := os.Mkdir(current, 0700); err != nil {
			t.Fatalf("Mkdir(%q): %v", current, err)
		}
	}

	err := WalkMediaTree(root, true, NewMediaScanBudget(), func(string, os.DirEntry) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "directory depth limit") {
		t.Fatalf("WalkMediaTree() error = %v, want directory depth limit", err)
	}
}
