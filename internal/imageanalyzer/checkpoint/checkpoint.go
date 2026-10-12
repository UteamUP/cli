// Package checkpoint persists only CLI-owned, tenant- and scan-scoped image state.
package checkpoint

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/uteamup/cli/internal/security"
)

const checkpointKind = "uteamup.image-checkpoint"
const maxCheckpointBytes int64 = 16 * 1024 * 1024

var hashPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)
var tenantPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

type Scope struct {
	APIOrigin  string `json:"api_origin"`
	TenantGUID string `json:"tenant_guid"`
	InputRoot  string `json:"input_root"`
	Profile    string `json:"profile"`
}

func NewScope(origin, tenant, inputRoot, profile string) (Scope, error) {
	normalized, err := security.Origin(origin)
	if err != nil {
		return Scope{}, err
	}
	tenant = strings.ToLower(strings.TrimSpace(tenant))
	if !tenantPattern.MatchString(tenant) || tenant == "00000000-0000-0000-0000-000000000000" {
		return Scope{}, fmt.Errorf("authenticated tenant GUID is required for checkpoint scope")
	}
	if profile == "" || len(profile) > 256 || security.SafeText(profile) != profile {
		return Scope{}, fmt.Errorf("authenticated profile is required for checkpoint scope")
	}
	root, err := filepath.Abs(inputRoot)
	if err != nil {
		return Scope{}, err
	}
	return Scope{APIOrigin: normalized, TenantGUID: tenant, InputRoot: filepath.Clean(root), Profile: profile}, nil
}

// ManagedDirectory is created through anchored parent descriptors; ambient YAML
// cannot give the image analyzer deletion authority over an arbitrary pathname.
func ManagedDirectory() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	root, err := security.OpenDirectory(home)
	if err != nil {
		return "", err
	}
	defer root.Close()
	if err := root.Mkdir(".uteamup", 0700); err != nil && !os.IsExist(err) {
		return "", err
	}
	state, err := security.OpenDirectory(filepath.Join(home, ".uteamup"))
	if err != nil {
		return "", err
	}
	defer state.Close()
	if err := state.Mkdir("image-checkpoints", 0700); err != nil && !os.IsExist(err) {
		return "", err
	}
	path := filepath.Join(home, ".uteamup", "image-checkpoints")
	check, err := security.OpenDirectory(path)
	if err != nil {
		return "", err
	}
	check.Close()
	return path, nil
}

func Path(scope Scope) (string, error) {
	dir, err := ManagedDirectory()
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(scope)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return filepath.Join(dir, hex.EncodeToString(sum[:])+".json"), nil
}

func ValidatePath(path string) error {
	dir, err := ManagedDirectory()
	if err != nil {
		return err
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	if filepath.Dir(absolute) != dir || filepath.Ext(absolute) != ".json" {
		return fmt.Errorf("checkpoint_file must be a JSON file inside %s", dir)
	}
	return nil
}

type Status struct {
	ProcessedCount int            `json:"processed_count"`
	StartedAt      string         `json:"started_at"`
	LastUpdated    string         `json:"last_updated"`
	TypeBreakdown  map[string]int `json:"type_breakdown"`
	FlaggedCount   int            `json:"flagged_for_review"`
}

type Checkpoint struct {
	path, name, lockPath, lockName string
	root                           *os.Root
	scope                          Scope
	processed                      map[string]json.RawMessage
	startedAt, lastUpdated, lockID string
	identity, lockIdentity         os.FileInfo
	mu                             sync.Mutex
}

type checkpointFile struct {
	Kind           string                     `json:"kind"`
	Version        int                        `json:"version"`
	Scope          Scope                      `json:"scope"`
	StartedAt      string                     `json:"started_at"`
	LastUpdated    string                     `json:"last_updated"`
	ProcessedCount int                        `json:"processed_count"`
	Processed      map[string]json.RawMessage `json:"processed"`
}

type lockFile struct {
	Kind    string `json:"kind"`
	Scope   Scope  `json:"scope"`
	ID      string `json:"id"`
	PID     int    `json:"pid"`
	Started string `json:"started"`
}

func decodeOwned(data []byte, value any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return fmt.Errorf("unexpected checkpoint content")
	}
	return nil
}

func Load(path string, scope Scope, resume bool) (*Checkpoint, error) {
	normalized, err := NewScope(scope.APIOrigin, scope.TenantGUID, scope.InputRoot, scope.Profile)
	if err != nil || normalized != scope {
		return nil, fmt.Errorf("canonical authenticated checkpoint scope is required")
	}
	if err := ValidatePath(path); err != nil {
		return nil, err
	}
	root, err := security.OpenDirectory(filepath.Dir(path))
	if err != nil {
		return nil, err
	}
	cp := &Checkpoint{path: path, name: filepath.Base(path), lockPath: path + ".lock", lockName: filepath.Base(path) + ".lock", root: root, scope: scope, processed: make(map[string]json.RawMessage), startedAt: time.Now().Format(time.RFC3339)}
	data, err := security.ReadRootFile(root, cp.name, maxCheckpointBytes)
	if os.IsNotExist(err) {
		return cp, nil
	}
	if err != nil {
		root.Close()
		return nil, fmt.Errorf("reading checkpoint: %w", err)
	}
	var stored checkpointFile
	if err := decodeOwned(data, &stored); err != nil || stored.Kind != checkpointKind || stored.Version != 1 || stored.Scope != scope || stored.Processed == nil || stored.ProcessedCount != len(stored.Processed) || stored.StartedAt == "" {
		root.Close()
		return nil, fmt.Errorf("checkpoint is not owned by this authenticated tenant, backend and scan; legacy or unrelated state was preserved")
	}
	if !resume {
		root.Close()
		return nil, fmt.Errorf("unfinished checkpoint exists; use --resume to restore it")
	}
	for hash, result := range stored.Processed {
		if !hashPattern.MatchString(hash) || !json.Valid(result) {
			root.Close()
			return nil, fmt.Errorf("checkpoint result is invalid")
		}
	}
	cp.identity, err = root.Lstat(cp.name)
	if err != nil {
		root.Close()
		return nil, err
	}
	cp.processed, cp.startedAt, cp.lastUpdated = stored.Processed, stored.StartedAt, stored.LastUpdated
	return cp, nil
}

func (cp *Checkpoint) Close() { cp.root.Close() }

func (cp *Checkpoint) AcquireLock() error {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	// Serialize stale-lock reclamation as well as exclusive lock creation.
	// A crashed acquisition leaves this guard in place and fails closed.
	guard := cp.lockName + ".acquiring"
	if err := cp.root.Mkdir(guard, 0700); err != nil {
		return fmt.Errorf("checkpoint lock acquisition is in progress or interrupted: %w", err)
	}
	defer cp.root.Remove(guard)
	if info, err := cp.root.Lstat(cp.lockName); err == nil {
		data, err := security.ReadRootFile(cp.root, cp.lockName, 4096)
		var existing lockFile
		if err != nil || decodeOwned(data, &existing) != nil || existing.Kind != checkpointKind || existing.Scope != cp.scope || existing.PID <= 0 || existing.ID == "" {
			return fmt.Errorf("unowned or invalid checkpoint lock was preserved")
		}
		if processAlive(existing.PID) {
			return fmt.Errorf("another process (PID %d) is using this checkpoint", existing.PID)
		}
		current, err := cp.root.Lstat(cp.lockName)
		if err != nil || !os.SameFile(info, current) {
			return fmt.Errorf("checkpoint lock identity changed")
		}
		if err := cp.root.Remove(cp.lockName); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		return err
	}
	cp.lockID = hex.EncodeToString(id)
	data, err := json.Marshal(lockFile{Kind: checkpointKind, Scope: cp.scope, ID: cp.lockID, PID: os.Getpid(), Started: time.Now().Format(time.RFC3339)})
	if err != nil {
		return err
	}
	file, err := cp.root.OpenFile(cp.lockName, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("acquiring checkpoint lock: %w", err)
	}
	if _, err = file.Write(data); err != nil {
		file.Close()
		return err
	}
	if err = file.Sync(); err != nil {
		file.Close()
		return err
	}
	cp.lockIdentity, err = file.Stat()
	closeErr := file.Close()
	if err != nil {
		return err
	}
	return closeErr
}

func (cp *Checkpoint) ReleaseLock() {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	if cp.ownsLock() {
		_ = cp.root.Remove(cp.lockName)
	}
	cp.lockID = ""
}

func (cp *Checkpoint) ownsLock() bool {
	if cp.lockID == "" || cp.lockIdentity == nil {
		return false
	}
	info, err := cp.root.Lstat(cp.lockName)
	if err != nil || !os.SameFile(cp.lockIdentity, info) {
		return false
	}
	data, err := security.ReadRootFile(cp.root, cp.lockName, 4096)
	var lock lockFile
	return err == nil && decodeOwned(data, &lock) == nil && lock.Kind == checkpointKind && lock.Scope == cp.scope && lock.ID == cp.lockID && lock.PID == os.Getpid()
}

func (cp *Checkpoint) unchanged() bool {
	current, err := cp.root.Lstat(cp.name)
	if cp.identity == nil {
		return os.IsNotExist(err)
	}
	return err == nil && os.SameFile(cp.identity, current) && cp.identity.Size() == current.Size() && cp.identity.ModTime().Equal(current.ModTime())
}

func (cp *Checkpoint) IsProcessed(hash string) bool {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	_, ok := cp.processed[hash]
	return ok
}

func (cp *Checkpoint) AddResult(hash string, result json.RawMessage) error {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	if !cp.ownsLock() || !cp.unchanged() {
		return fmt.Errorf("checkpoint ownership changed")
	}
	if !hashPattern.MatchString(hash) || !json.Valid(result) {
		return fmt.Errorf("checkpoint result is invalid")
	}
	cp.processed[hash] = result
	cp.lastUpdated = time.Now().Format(time.RFC3339)
	stored := checkpointFile{Kind: checkpointKind, Version: 1, Scope: cp.scope, StartedAt: cp.startedAt, LastUpdated: cp.lastUpdated, ProcessedCount: len(cp.processed), Processed: cp.processed}
	data, err := json.MarshalIndent(stored, "", "  ")
	if err != nil {
		return err
	}
	if int64(len(data)) > maxCheckpointBytes {
		return fmt.Errorf("checkpoint exceeds size limit")
	}
	if err := security.WriteRootFile(cp.root, cp.name, data); err != nil {
		return err
	}
	cp.identity, err = cp.root.Lstat(cp.name)
	return err
}

// GetResults preserves the current content-hash key used to bind restored results.
func (cp *Checkpoint) GetResults() map[string]json.RawMessage {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	results := make(map[string]json.RawMessage, len(cp.processed))
	for hash, value := range cp.processed {
		results[hash] = append(json.RawMessage(nil), value...)
	}
	return results
}

func (cp *Checkpoint) Delete() error {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	if !cp.ownsLock() || !cp.unchanged() {
		return fmt.Errorf("checkpoint ownership changed; state was preserved")
	}
	if cp.identity != nil {
		if err := cp.root.Remove(cp.name); err != nil {
			return err
		}
		cp.identity = nil
	}
	if err := cp.root.Remove(cp.lockName); err != nil {
		return err
	}
	cp.lockID = ""
	return nil
}

func (cp *Checkpoint) GetStatus() Status {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	types := make(map[string]int)
	flagged := 0
	for _, raw := range cp.processed {
		var results []map[string]any
		if json.Unmarshal(raw, &results) != nil {
			var result map[string]any
			if json.Unmarshal(raw, &result) != nil {
				continue
			}
			results = []map[string]any{result}
		}
		for _, result := range results {
			if classification, ok := result["classification"].(map[string]any); ok {
				name, _ := classification["primary_type"].(string)
				if name == "" {
					name = "unknown"
				}
				types[name]++
			}
			if result["flagged_for_review"] == true {
				flagged++
			}
		}
	}
	return Status{ProcessedCount: len(cp.processed), StartedAt: cp.startedAt, LastUpdated: cp.lastUpdated, TypeBreakdown: types, FlaggedCount: flagged}
}
