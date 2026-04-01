// Package checkpoint provides resume capability with atomic writes and lock files.
package checkpoint

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

// CheckpointStatus holds summary information about the checkpoint state.
type CheckpointStatus struct {
	ProcessedCount int            `json:"processed_count"`
	StartedAt      string         `json:"started_at"`
	LastUpdated    string         `json:"last_updated"`
	TypeBreakdown  map[string]int `json:"type_breakdown"`
	FlaggedCount   int            `json:"flagged_for_review"`
}

// Checkpoint persists processing state for resume-after-interruption.
type Checkpoint struct {
	path        string
	processed   map[string]json.RawMessage
	startedAt   string
	lastUpdated string
	lockPath    string
	mu          sync.Mutex
}

// checkpointFile is the JSON structure persisted to disk.
type checkpointFile struct {
	StartedAt      string                     `json:"started_at"`
	LastUpdated    string                     `json:"last_updated"`
	ProcessedCount int                        `json:"processed_count"`
	Processed      map[string]json.RawMessage `json:"processed"`
}

// lockFile is the JSON structure for the lock file.
type lockFile struct {
	PID     int    `json:"pid"`
	Started string `json:"started"`
}

// Load reads an existing checkpoint from disk, or creates a new one if the file does not exist.
func Load(checkpointPath string) (*Checkpoint, error) {
	cp := &Checkpoint{
		path:      checkpointPath,
		processed: make(map[string]json.RawMessage),
		lockPath:  checkpointPath + ".lock",
	}

	data, err := os.ReadFile(checkpointPath)
	if err != nil {
		if os.IsNotExist(err) {
			cp.startedAt = time.Now().Format(time.RFC3339)
			return cp, nil
		}
		return nil, fmt.Errorf("reading checkpoint: %w", err)
	}

	var cf checkpointFile
	if err := json.Unmarshal(data, &cf); err != nil {
		return nil, fmt.Errorf("parsing checkpoint: %w", err)
	}

	cp.processed = cf.Processed
	if cp.processed == nil {
		cp.processed = make(map[string]json.RawMessage)
	}
	cp.startedAt = cf.StartedAt
	cp.lastUpdated = cf.LastUpdated

	log.Printf("checkpoint: loaded, processed_count=%d, started_at=%s", len(cp.processed), cp.startedAt)
	return cp, nil
}

// AcquireLock writes a lock file with the current PID. If a lock file already exists,
// it checks whether the owning process is still alive. Stale locks are removed.
func (cp *Checkpoint) AcquireLock() error {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	if _, err := os.Stat(cp.lockPath); err == nil {
		// Lock file exists — check if process is alive.
		data, readErr := os.ReadFile(cp.lockPath)
		if readErr == nil {
			var lf lockFile
			if json.Unmarshal(data, &lf) == nil && lf.PID > 0 {
				proc, findErr := os.FindProcess(lf.PID)
				if findErr == nil {
					// Signal 0 checks if process exists without sending a signal.
					if err := proc.Signal(syscall.Signal(0)); err == nil {
						return fmt.Errorf("another process (PID %d) is using this checkpoint; delete %s if the process is not running", lf.PID, cp.lockPath)
					}
				}
				log.Printf("checkpoint: stale lock file found (PID %d), removing", lf.PID)
			} else {
				log.Printf("checkpoint: corrupt lock file found, removing")
			}
		}
		os.Remove(cp.lockPath)
	}

	lf := lockFile{
		PID:     os.Getpid(),
		Started: time.Now().Format(time.RFC3339),
	}
	data, err := json.Marshal(lf)
	if err != nil {
		return fmt.Errorf("marshalling lock: %w", err)
	}
	if err := os.WriteFile(cp.lockPath, data, 0644); err != nil {
		return fmt.Errorf("writing lock file: %w", err)
	}
	return nil
}

// ReleaseLock removes the lock file.
func (cp *Checkpoint) ReleaseLock() {
	os.Remove(cp.lockPath)
}

// IsProcessed returns true if the given file hash has already been processed.
func (cp *Checkpoint) IsProcessed(fileHash string) bool {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	_, ok := cp.processed[fileHash]
	return ok
}

// AddResult stores a result for the given file hash and atomically saves the checkpoint.
func (cp *Checkpoint) AddResult(fileHash string, result json.RawMessage) error {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	cp.processed[fileHash] = result
	cp.lastUpdated = time.Now().Format(time.RFC3339)
	return cp.save()
}

// GetResults returns all stored results.
func (cp *Checkpoint) GetResults() []json.RawMessage {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	results := make([]json.RawMessage, 0, len(cp.processed))
	for _, v := range cp.processed {
		results = append(results, v)
	}
	return results
}

// Delete removes both the checkpoint file and the lock file.
func (cp *Checkpoint) Delete() error {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	for _, p := range []string{cp.path, cp.lockPath} {
		if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("removing %s: %w", p, err)
		}
	}
	return nil
}

// GetStatus returns summary information about the checkpoint state.
func (cp *Checkpoint) GetStatus() CheckpointStatus {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	typeCounts := make(map[string]int)
	flagged := 0

	for _, raw := range cp.processed {
		var result map[string]interface{}
		if json.Unmarshal(raw, &result) != nil {
			continue
		}
		if classification, ok := result["classification"].(map[string]interface{}); ok {
			entityType, _ := classification["primary_type"].(string)
			if entityType == "" {
				entityType = "unknown"
			}
			typeCounts[entityType]++
		}
		if flaggedVal, ok := result["flagged_for_review"].(bool); ok && flaggedVal {
			flagged++
		}
	}

	return CheckpointStatus{
		ProcessedCount: len(cp.processed),
		StartedAt:      cp.startedAt,
		LastUpdated:    cp.lastUpdated,
		TypeBreakdown:  typeCounts,
		FlaggedCount:   flagged,
	}
}

// save writes the checkpoint atomically via temp file + rename.
func (cp *Checkpoint) save() error {
	cf := checkpointFile{
		StartedAt:      cp.startedAt,
		LastUpdated:    cp.lastUpdated,
		ProcessedCount: len(cp.processed),
		Processed:      cp.processed,
	}

	data, err := json.MarshalIndent(cf, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling checkpoint: %w", err)
	}

	dir := filepath.Dir(cp.path)
	tmpFile, err := os.CreateTemp(dir, "checkpoint-*.json.tmp")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("writing temp file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("closing temp file: %w", err)
	}

	if err := os.Rename(tmpPath, cp.path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("renaming temp to checkpoint: %w", err)
	}
	return nil
}
