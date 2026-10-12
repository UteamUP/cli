package client

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/uteamup/cli/internal/auth"
	"github.com/uteamup/cli/internal/logging"
)

func TestRetriesRequireReadSemanticsOrDeduplicatedMedia(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	file := filepath.Join(t.TempDir(), "photo.png")
	if err := os.WriteFile(file, []byte("stable payload"), 0600); err != nil {
		t.Fatal(err)
	}
	key := "e1cddc64-a986-45de-b9b4-f68f4dc82631"
	tests := []struct {
		name, method, path, kind, key string
		attempts                      int
	}{
		{"read", "GET", "/api/assets", "rest", "", 2},
		{"create", "POST", "/api/assets", "rest", "", 1},
		{"lost create response", "POST", "/api/assets", "rest", "", 1},
		{"lost photo response", "POST", "/api/inventoryai/analyze-photo", "limited", key, 2},
		{"delete", "DELETE", "/api/assets/one", "rest", "", 1},
		{"tool", "POST", "/mcp", "tool", "", 1},
		{"multipart", "POST", "/api/stock/import", "upload", key, 1},
		{"unknown media", "POST", "/api/assets", "limited", key, 1},
		{"media missing key", "POST", "/api/inventoryai/analyze-photo", "limited", "", 1},
		{"media invalid key", "POST", "/api/inventoryai/analyze-photo", "limited", "bad", 1},
		{"photo replay", "POST", "/api/inventoryai/analyze-photo", "limited", key, 2},
		{"video replay", "POST", "/api/inventoryai/analyze-video", "limited", key, 2},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			attempts := 0
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				attempts++
				if tc.kind == "limited" {
					if r.Header.Get("Idempotency-Key") != tc.key {
						t.Error("idempotency key changed")
					}
					part, _, err := r.FormFile("file")
					if err != nil {
						t.Error(err)
					} else {
						contents, _ := io.ReadAll(part)
						part.Close()
						if string(contents) != "stable payload" {
							t.Error("retry upload changed")
						}
					}
				} else {
					_, _ = io.Copy(io.Discard, r.Body)
				}
				if strings.HasPrefix(tc.name, "lost") {
					connection, _, err := w.(http.Hijacker).Hijack()
					if err != nil {
						t.Error(err)
						return
					}
					connection.Close()
					return
				}
				w.WriteHeader(http.StatusServiceUnavailable)
				_, _ = w.Write([]byte(`{"error":"provider unavailable"}`))
			}))
			defer server.Close()
			if err := auth.SaveToken(&auth.TokenData{APIOrigin: server.URL, AccessToken: "test-token", AuthMethod: "apikey", ExpiresAt: time.Now().Add(time.Hour), TenantGUID: "b966b8c7-04a4-45d4-aa51-519ecf2ef13a"}); err != nil {
				t.Fatal(err)
			}
			client := NewAPIClient(server.URL, time.Second, true, RetryOptions{MaxRetries: 1, BaseDelay: time.Millisecond, MaxDelay: time.Millisecond}, logging.New(logging.LevelError))
			headers := map[string]string{"Idempotency-Key": tc.key}
			var err error
			switch tc.kind {
			case "tool":
				_, err = client.CallTool(context.Background(), "create_asset", nil)
			case "upload":
				_, err = client.CallRESTUpload(context.Background(), tc.method, tc.path, "file", file, nil, headers, "")
			case "limited":
				_, err = client.CallRESTUploadLimited(context.Background(), tc.method, tc.path, "file", file, "photo.png", "image/png", 1024, headers)
			default:
				_, err = client.CallREST(context.Background(), tc.method, tc.path, nil, headers, "")
			}
			if err == nil || attempts != tc.attempts {
				t.Fatalf("error=%v attempts=%d, want %d", err, attempts, tc.attempts)
			}
		})
	}
}
