package checkpoint

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func tmpCheckpointPath(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	dir, err := ManagedDirectory()
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Join(dir, "test_checkpoint.json")
}

func TestConcurrentAcquisitionHasOnlyOneOwner(t *testing.T) {
	path := tmpCheckpointPath(t)
	var checkpoints []*Checkpoint
	for i := 0; i < 10; i++ {
		cp, err := Load(path, testScope(t), false)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(cp.Close)
		checkpoints = append(checkpoints, cp)
	}
	start := make(chan struct{})
	owners := make(chan *Checkpoint, len(checkpoints))
	var workers sync.WaitGroup
	for _, cp := range checkpoints {
		workers.Add(1)
		go func(cp *Checkpoint) {
			defer workers.Done()
			<-start
			if cp.AcquireLock() == nil {
				owners <- cp
			}
		}(cp)
	}
	close(start)
	workers.Wait()
	close(owners)
	if len(owners) != 1 {
		t.Fatalf("concurrent owners = %d, want 1", len(owners))
	}
	for cp := range owners {
		cp.ReleaseLock()
	}
}

func TestLoadNew(t *testing.T) {
	path := tmpCheckpointPath(t)

	cp, err := Load(path, testScope(t), true)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	t.Cleanup(cp.Close)
	if len(cp.processed) != 0 {
		t.Fatalf("expected empty processed map, got %d entries", len(cp.processed))
	}
	if cp.startedAt == "" {
		t.Fatal("expected startedAt to be set")
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	path := tmpCheckpointPath(t)

	cp, err := Load(path, testScope(t), true)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	t.Cleanup(cp.Close)

	if err := cp.AcquireLock(); err != nil {
		t.Fatal(err)
	}
	defer cp.ReleaseLock()

	result := json.RawMessage(`{"classification":{"primary_type":"vehicle"},"flagged_for_review":false}`)
	if err := cp.AddResult(strings.Repeat("a", 64), result); err != nil {
		t.Fatalf("AddResult failed: %v", err)
	}

	result2 := json.RawMessage(`{"classification":{"primary_type":"building"},"flagged_for_review":true}`)
	if err := cp.AddResult(strings.Repeat("b", 64), result2); err != nil {
		t.Fatalf("AddResult failed: %v", err)
	}

	// Reload from disk.
	cp2, err := Load(path, testScope(t), true)
	if err != nil {
		t.Fatalf("Reload failed: %v", err)
	}
	t.Cleanup(cp2.Close)
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

	cp, err := Load(path, testScope(t), true)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	t.Cleanup(cp.Close)

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

	cp, err := Load(path, testScope(t), true)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	t.Cleanup(cp.Close)

	if cp.IsProcessed(strings.Repeat("1", 64)) {
		t.Fatal("expected hash1 to not be processed")
	}

	if err := cp.AcquireLock(); err != nil {
		t.Fatal(err)
	}
	defer cp.ReleaseLock()

	result := json.RawMessage(`{"data":"test"}`)
	if err := cp.AddResult(strings.Repeat("1", 64), result); err != nil {
		t.Fatalf("AddResult failed: %v", err)
	}

	if !cp.IsProcessed(strings.Repeat("1", 64)) {
		t.Fatal("expected hash1 to be processed after AddResult")
	}
}

func TestDelete(t *testing.T) {
	path := tmpCheckpointPath(t)

	cp, err := Load(path, testScope(t), true)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	t.Cleanup(cp.Close)

	if err := cp.AcquireLock(); err != nil {
		t.Fatal(err)
	}
	defer cp.ReleaseLock()

	// Write some data so the file exists on disk.
	if err := cp.AddResult(strings.Repeat("c", 64), json.RawMessage(`{}`)); err != nil {
		t.Fatalf("AddResult failed: %v", err)
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

func testScope(t *testing.T) Scope {
	t.Helper()
	scope, err := NewScope("https://api.uteamup.com:443/", "b966b8c7-04a4-45d4-aa51-519ecf2ef13a", filepath.Join(os.TempDir(), "uteamup-test-photos"), "default")
	if err != nil {
		t.Fatal(err)
	}
	return scope
}

func TestLoadRefusesUnownedFilesAndMismatchedScope(t *testing.T) {
	path := tmpCheckpointPath(t)
	unrelated := []byte(`{"important":"keep this JSON"}`)
	if err := os.WriteFile(path, unrelated, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path, testScope(t), true); err == nil {
		t.Fatal("unowned JSON accepted")
	}
	after, _ := os.ReadFile(path)
	if string(after) != string(unrelated) {
		t.Fatal("unowned state changed")
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	cp, err := Load(path, testScope(t), false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(cp.Close)
	t.Cleanup(cp.Close)
	if err := cp.AcquireLock(); err != nil {
		t.Fatal(err)
	}
	defer cp.ReleaseLock()
	if err := cp.AddResult(strings.Repeat("a", 64), json.RawMessage(`[]`)); err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(*Scope){
		func(s *Scope) { s.TenantGUID = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa" },
		func(s *Scope) { s.APIOrigin = "https://devback.uteamup.com" },
		func(s *Scope) { s.InputRoot = filepath.Join(s.InputRoot, "another") },
		func(s *Scope) { s.Profile = "other-profile" },
	} {
		scope := testScope(t)
		change(&scope)
		if _, err := Load(path, scope, true); err == nil {
			t.Fatal("foreign scope accepted")
		}
	}
	if _, err := Load(path, testScope(t), false); err == nil {
		t.Fatal("checkpoint restored without explicit resume")
	}
}

func TestDeletePreservesSwappedStateAndSymlinkLocks(t *testing.T) {
	path := tmpCheckpointPath(t)
	cp, err := Load(path, testScope(t), false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(cp.Close)
	t.Cleanup(cp.Close)
	outside := filepath.Join(t.TempDir(), "private.json")
	if err := os.WriteFile(outside, []byte(`{"keep":true}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, cp.lockPath); err != nil {
		t.Skip(err)
	}
	if err := cp.AcquireLock(); err == nil {
		t.Fatal("linked lock accepted")
	}
	if err := os.Remove(cp.lockPath); err != nil {
		t.Fatal(err)
	}
	if err := cp.AcquireLock(); err != nil {
		t.Fatal(err)
	}
	defer cp.ReleaseLock()
	if err := cp.AddResult(strings.Repeat("a", 64), json.RawMessage(`[]`)); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(`{"keep":true}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := cp.Delete(); err == nil {
		t.Fatal("swapped state deleted")
	}
	data, _ := os.ReadFile(path)
	if string(data) != `{"keep":true}` {
		t.Fatal("swapped file was changed")
	}
	data, _ = os.ReadFile(outside)
	if string(data) != `{"keep":true}` {
		t.Fatal("linked target was changed")
	}
}
