package client

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/uteamup/cli/internal/auth"
	"github.com/uteamup/cli/internal/logging"
)

func TestUploadContentTypeUsesSupportedAttachmentMediaTypes(t *testing.T) {
	tests := map[string]string{
		"report.txt":  "text/plain",
		"codes.csv":   "text/csv",
		"manual.pdf":  "application/pdf",
		"photo.webp":  "image/webp",
		"video.webm":  "video/webm",
		"unknown.bin": "application/octet-stream",
	}

	for name, want := range tests {
		if got := uploadContentType(name); got != want {
			t.Errorf("uploadContentType(%q) = %q, want %q", name, got, want)
		}
	}
}

func TestCallRESTUploadLimitedUsesAuthenticatedGuidScopedMultipart(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	if err := auth.SaveToken(&auth.TokenData{
		AccessToken: "secret-access-token",
		ExpiresAt:   time.Now().Add(time.Hour),
		TenantID:    42,
		TenantGUID:  "b966b8c7-04a4-45d4-aa51-519ecf2ef13a",
	}); err != nil {
		t.Fatal(err)
	}

	uploadPath := filepath.Join(t.TempDir(), "source.mp4")
	if err := os.WriteFile(uploadPath, []byte("safe-media"), 0o600); err != nil {
		t.Fatal(err)
	}

	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/api/inventoryai/analyze-video" {
			t.Errorf("path = %q", request.URL.Path)
		}
		if request.Header.Get("Authorization") != "Bearer secret-access-token" {
			t.Errorf("bearer authentication was overridden: %q", request.Header.Get("Authorization"))
		}
		if request.Header.Get("X-Tenant-Guid") != "b966b8c7-04a4-45d4-aa51-519ecf2ef13a" {
			t.Error("missing tenant GUID header")
		}
		if request.Header.Get("X-Tenant-ID") != "" {
			t.Error("integer tenant ID leaked across the public boundary")
		}
		if request.Header.Get("Idempotency-Key") != "e1cddc64-a986-45de-b9b4-f68f4dc82631" {
			t.Error("missing idempotency key")
		}

		file, header, err := request.FormFile("file")
		if err != nil {
			t.Errorf("reading multipart file: %v", err)
			response.WriteHeader(http.StatusBadRequest)
			return
		}
		defer file.Close()
		if header.Filename != "unsafe-name.mp4" {
			t.Errorf("sanitized filename = %q", header.Filename)
		}
		contents, _ := io.ReadAll(file)
		if string(contents) != "safe-media" {
			t.Errorf("contents = %q", contents)
		}
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{"items":[]}`))
	}))
	defer server.Close()

	apiClient := NewAPIClient(server.URL, time.Second, true, RetryOptions{MaxRetries: 0}, logging.New(logging.LevelError))
	_, err := apiClient.CallRESTUploadLimited(
		context.Background(),
		"POST",
		"/api/inventoryai/analyze-video",
		"file",
		uploadPath,
		"../unsafe\n-name.mp4",
		"video/mp4",
		1024,
		map[string]string{
			"Idempotency-Key": "e1cddc64-a986-45de-b9b4-f68f4dc82631",
			"Authorization":   "Bearer attacker-controlled",
			"X-Tenant-Guid":   "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
		},
	)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCallRESTUploadLimitedRejectsOversizedFileBeforeRequest(t *testing.T) {
	path := filepath.Join(t.TempDir(), "large.bin")
	if err := os.WriteFile(path, []byte("too-large"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := validateUploadFile(path, 3); err == nil {
		t.Fatal("expected oversized upload validation error")
	}
}
