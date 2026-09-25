package client

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func assertDirEntries(t *testing.T, dir string, want ...string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	if len(entries) != len(want) {
		t.Fatalf("dir %s holds %d entries, want %v", dir, len(entries), want)
	}
	for index, entry := range entries {
		if entry.Name() != want[index] {
			t.Fatalf("dir entry %d = %s, want %s", index, entry.Name(), want[index])
		}
	}
}

func TestWriteDownloadFileOverLimitRemovesTemp(t *testing.T) {
	dir := t.TempDir()
	outputPath := filepath.Join(dir, "too-large.pdf")

	written, err := writeDownloadFileWithLimit(outputPath, strings.NewReader("0123456789"), 4)
	if err == nil || !strings.Contains(err.Error(), "4 byte limit") {
		t.Fatalf("error = %v, want the byte-limit refusal", err)
	}
	if written != 0 {
		t.Fatalf("written = %d, want 0 on refusal", written)
	}
	if _, statErr := os.Stat(outputPath); !os.IsNotExist(statErr) {
		t.Fatalf("over-limit download left %s behind", outputPath)
	}
	assertDirEntries(t, dir)

	written, err = writeDownloadFileWithLimit(outputPath, strings.NewReader("0123"), 4)
	if err != nil || written != 4 {
		t.Fatalf("exactly-at-limit write = %d, %v; want 4, nil", written, err)
	}
	content, err := os.ReadFile(outputPath)
	if err != nil || string(content) != "0123" {
		t.Fatalf("at-limit content = %q, %v", content, err)
	}
	assertDirEntries(t, dir, "too-large.pdf")
}

func TestWriteDownloadFileRefusesExisting(t *testing.T) {
	dir := t.TempDir()
	outputPath := filepath.Join(dir, "reports.csv")
	if err := os.WriteFile(outputPath, []byte("original"), 0o600); err != nil {
		t.Fatalf("seed existing file: %v", err)
	}

	written, err := WriteDownloadFile(outputPath, strings.NewReader("replacement"))
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("error = %v, want an already-exists refusal", err)
	}
	if written != 0 {
		t.Fatalf("written = %d, want 0", written)
	}
	content, err := os.ReadFile(outputPath)
	if err != nil || string(content) != "original" {
		t.Fatalf("existing file changed to %q, %v", content, err)
	}
	assertDirEntries(t, dir, "reports.csv")

	if _, err := WriteDownloadFile("  ", strings.NewReader("x")); err == nil ||
		!strings.Contains(err.Error(), "path is required") {
		t.Fatalf("blank path error = %v", err)
	}
}
