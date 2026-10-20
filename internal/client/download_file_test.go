package client

import (
	"context"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsAuthorizedBlobOrigin(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		trusted []string
		want    bool
	}{
		{name: "Azure Blob account", url: "https://stuteamup.blob.core.windows.net/container/file?sig=secret", want: true},
		{name: "Azure suffix lookalike", url: "https://blob.core.windows.net.attacker.example/file", want: false},
		{name: "Azure unexpected port", url: "https://stuteamup.blob.core.windows.net:8443/file", want: false},
		{name: "custom exact origin", url: "https://Blob.Internal.Example:8443/container/file?sig=secret", trusted: []string{"https://blob.internal.example:8443"}, want: true},
		{name: "custom origin wrong port", url: "https://blob.internal.example:9443/file", trusted: []string{"https://blob.internal.example:8443"}, want: false},
		{name: "custom origin wrong subdomain", url: "https://other.blob.internal.example:8443/file", trusted: []string{"https://blob.internal.example:8443"}, want: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parsed, err := url.Parse(test.url)
			if err != nil {
				t.Fatal(err)
			}
			if got := isAuthorizedBlobOrigin(parsed, test.trusted); got != test.want {
				t.Fatalf("isAuthorizedBlobOrigin(%q) = %v, want %v", test.url, got, test.want)
			}
		})
	}
}

func TestDownloadFileRejectsUntrustedOriginBeforeNetworkOrFileWrite(t *testing.T) {
	dir := t.TempDir()
	outputPath := filepath.Join(dir, "attachment.pdf")
	apiClient := NewAPIClient("https://api.uteamup.com", 0, false, RetryOptions{}, nil)

	if _, err := apiClient.DownloadFile(context.Background(), "https://127.0.0.1/private", outputPath); err == nil ||
		!strings.Contains(err.Error(), "origin is not trusted") {
		t.Fatalf("download error = %v, want untrusted-origin rejection", err)
	}
	if _, err := os.Stat(outputPath); !os.IsNotExist(err) {
		t.Fatalf("untrusted download created output: stat error = %v", err)
	}
}

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
