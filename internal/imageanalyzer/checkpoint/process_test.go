package checkpoint

import (
	"encoding/json"
	"os"
	"testing"
)

// Regression test for the Windows liveness bug: AcquireLock probed the owning
// process with syscall.Signal(0), which Windows rejects with "not supported by
// windows" for every signal except Kill. A live process therefore looked dead,
// so concurrent runs each treated the other's lock as stale and removed it.
func TestProcessAliveReportsRunningProcess(t *testing.T) {
	if !processAlive(os.Getpid()) {
		t.Fatal("processAlive reported the current (definitely running) process as dead")
	}
}

func TestProcessAliveReportsMissingProcess(t *testing.T) {
	// Implausibly high pid: above the default max on Linux (4194304) and not a
	// value Windows hands out in practice, so it should never be live.
	const missingPID = 999999999

	if processAlive(missingPID) {
		t.Fatalf("processAlive reported pid %d as running", missingPID)
	}
}

// The behaviour the bug actually broke: a lock owned by a live process must be
// respected rather than silently deleted.
func TestAcquireLockRefusesLockHeldByLiveProcess(t *testing.T) {
	path := tmpCheckpointPath(t)

	cp, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Write a lock owned by this process — the canonical "someone else is
	// running" case, since our own pid is unambiguously alive.
	data, err := json.Marshal(lockFile{PID: os.Getpid(), Started: "2026-01-01T00:00:00Z"})
	if err != nil {
		t.Fatalf("marshalling lock: %v", err)
	}
	if err := os.WriteFile(cp.lockPath, data, 0o600); err != nil {
		t.Fatalf("writing lock: %v", err)
	}

	if err := cp.AcquireLock(); err == nil {
		t.Fatal("AcquireLock succeeded despite a lock held by a live process")
	}

	if _, err := os.Stat(cp.lockPath); err != nil {
		t.Fatalf("AcquireLock deleted a live process's lock file: %v", err)
	}
}

func TestAcquireLockReplacesStaleLock(t *testing.T) {
	path := tmpCheckpointPath(t)

	cp, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	data, err := json.Marshal(lockFile{PID: 999999999, Started: "2026-01-01T00:00:00Z"})
	if err != nil {
		t.Fatalf("marshalling lock: %v", err)
	}
	if err := os.WriteFile(cp.lockPath, data, 0o600); err != nil {
		t.Fatalf("writing lock: %v", err)
	}

	if err := cp.AcquireLock(); err != nil {
		t.Fatalf("AcquireLock refused a stale lock: %v", err)
	}
	defer cp.ReleaseLock()
}
