package checkpoint

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func tmpCheckpointPath(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return filepath.Join(dir, "test_checkpoint.json")
}

func TestLoadNew(t *testing.T) {
	path := tmpCheckpointPath(t)

	cp, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(cp.processed) != 0 {
		t.Fatalf("expected empty processed map, got %d entries", len(cp.processed))
	}
	if cp.startedAt == "" {
		t.Fatal("expected startedAt to be set")
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	path := tmpCheckpointPath(t)

	cp, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	result := json.RawMessage(`{"classification":{"primary_type":"vehicle"},"flagged_for_review":false}`)
	if err := cp.AddResult("abc123", result); err != nil {
		t.Fatalf("AddResult failed: %v", err)
	}

	result2 := json.RawMessage(`{"classification":{"primary_type":"building"},"flagged_for_review":true}`)
	if err := cp.AddResult("def456", result2); err != nil {
		t.Fatalf("AddResult failed: %v", err)
	}

	// Reload from disk.
	cp2, err := Load(path)
	if err != nil {
		t.Fatalf("Reload failed: %v", err)
	}
	if len(cp2.processed) != 2 {
		t.Fatalf("expected 2 processed entries, got %d", len(cp2.processed))
	}

	results := cp2.GetResults()
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestLockAcquireRelease(t *testing.T) {
	path := tmpCheckpointPath(t)

	cp, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if err := cp.AcquireLock(); err != nil {
		t.Fatalf("AcquireLock failed: %v", err)
	}

	// Lock file should exist.
	lockPath := path + ".lock"
	if _, err := os.Stat(lockPath); os.IsNotExist(err) {
		t.Fatal("expected lock file to exist")
	}

	cp.ReleaseLock()

	// Lock file should be gone.
	if _, err := os.Stat(lockPath); !os.IsNotExist(err) {
		t.Fatal("expected lock file to be removed after release")
	}
}

func TestIsProcessed(t *testing.T) {
	path := tmpCheckpointPath(t)

	cp, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cp.IsProcessed("hash1") {
		t.Fatal("expected hash1 to not be processed")
	}

	result := json.RawMessage(`{"data":"test"}`)
	if err := cp.AddResult("hash1", result); err != nil {
		t.Fatalf("AddResult failed: %v", err)
	}

	if !cp.IsProcessed("hash1") {
		t.Fatal("expected hash1 to be processed after AddResult")
	}
}

func TestDelete(t *testing.T) {
	path := tmpCheckpointPath(t)

	cp, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Write some data so the file exists on disk.
	if err := cp.AddResult("x", json.RawMessage(`{}`)); err != nil {
		t.Fatalf("AddResult failed: %v", err)
	}
	if err := cp.AcquireLock(); err != nil {
		t.Fatalf("AcquireLock failed: %v", err)
	}

	if err := cp.Delete(); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Both checkpoint and lock files should be gone.
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("expected checkpoint file to be deleted")
	}
	lockPath := path + ".lock"
	if _, err := os.Stat(lockPath); !os.IsNotExist(err) {
		t.Fatal("expected lock file to be deleted")
	}
}
