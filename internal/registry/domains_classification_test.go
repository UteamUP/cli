package registry

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/uteamup/cli/internal/auth"
	"github.com/uteamup/cli/internal/client"
	"github.com/uteamup/cli/internal/logging"
)

func TestCategoryMutationMetadataUsesNormalRoutesWithoutChangingOriginal(t *testing.T) {
	const id = "11111111-1111-4111-8111-111111111111"
	const operation = "22222222-2222-4222-8222-222222222222"
	const revision = "2026-10-16T10:00:00.1234567Z"
	for _, action := range []string{"create", "update", "delete", "legacy-delete"} {
		t.Run(action, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv("USERPROFILE", home)
			body := map[string]any{"name": "Reviewed", "idempotencyKey": operation, "mutationOutcomeVersion": float64(1)}
			if action == "update" {
				body["expectedUpdatedAt"] = revision
			}
			payload, _ := json.Marshal(body)
			file := filepath.Join(home, "original.json")
			if err := os.WriteFile(file, payload, 0600); err != nil {
				t.Fatal(err)
			}
			count := 0
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				count++
				path := "/api/category"
				method := "POST"
				if action != "create" {
					path += "/by-guid/" + id
					method = "PUT"
				}
				if action == "delete" || action == "legacy-delete" {
					method = "DELETE"
				}
				if r.Method != method || r.URL.Path != path {
					t.Errorf("wrong normal route: %s %s", r.Method, r.URL)
				}
				if action == "delete" {
					q := r.URL.Query()
					if len(q) != 3 || q.Get("idempotencyKey") != operation || q.Get("expectedUpdatedAt") != revision || q.Get("mutationOutcomeVersion") != "1" {
						t.Errorf("changed deletion metadata: %v", q)
					}
				} else if r.URL.RawQuery != "" {
					t.Errorf("unexpected query: %s", r.URL.RawQuery)
				}
				if action == "create" || action == "update" {
					var actual map[string]any
					if err := json.NewDecoder(r.Body).Decode(&actual); err != nil {
						t.Error(err)
					}
					if !reflect.DeepEqual(actual, body) {
						t.Errorf("original changed: %#v", actual)
					}
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"schemaVersion":1,"outcome":"applied"}`))
			}))
			defer server.Close()
			if err := auth.SaveToken(&auth.TokenData{APIOrigin: server.URL, AccessToken: "test", ExpiresAt: time.Now().Add(time.Hour), TenantGUID: id}); err != nil {
				t.Fatal(err)
			}
			api := client.NewAPIClient(server.URL, time.Second, true, client.RetryOptions{MaxRetries: 0}, logging.New(logging.LevelError))
			format := "json"
			cmd := buildDomainCommand(findDomainByName(t, "category"), func() (*client.APIClient, error) { return api, nil }, logging.New(logging.LevelError), &format, &ExportConfig{})
			args := []string{action, "--confirm"}
			if action == "legacy-delete" {
				args[0] = "delete"
			}
			if action != "create" {
				args = append(args, id)
			}
			if action == "create" || action == "update" {
				args = append(args, "--from-json", file)
			}
			if action == "delete" {
				args = append(args, "--idempotency-key", operation, "--expected-updated-at", revision, "--mutation-outcome-version", "1")
			}
			cmd.SetArgs(args)
			if err := cmd.Execute(); err != nil || count != 1 {
				t.Fatalf("request failed: %v count=%d", err, count)
			}
		})
	}
}

func TestTagMutationMetadataUsesNormalRoutesWithoutChangingOriginal(t *testing.T) {
	const id = "11111111-1111-4111-8111-111111111111"
	const operation = "22222222-2222-4222-8222-222222222222"
	const revision = "2026-10-16T10:00:00.1234567Z"
	for _, action := range []string{"create", "update", "delete", "legacy-delete"} {
		t.Run(action, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv("USERPROFILE", home)
			body := map[string]any{"name": "Reviewed", "idempotencyKey": operation, "mutationOutcomeVersion": float64(1)}
			if action == "update" {
				body["expectedUpdatedAt"] = revision
			}
			payload, _ := json.Marshal(body)
			file := filepath.Join(home, "original.json")
			if err := os.WriteFile(file, payload, 0600); err != nil {
				t.Fatal(err)
			}
			count := 0
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				count++
				path := "/api/tag"
				method := "POST"
				if action != "create" {
					path += "/by-guid/" + id
					method = "PUT"
				}
				if action == "delete" || action == "legacy-delete" {
					method = "DELETE"
				}
				if r.Method != method || r.URL.Path != path {
					t.Errorf("wrong normal route: %s %s", r.Method, r.URL)
				}
				if action == "delete" {
					q := r.URL.Query()
					if len(q) != 3 || q.Get("idempotencyKey") != operation || q.Get("expectedUpdatedAt") != revision || q.Get("mutationOutcomeVersion") != "1" {
						t.Errorf("changed deletion metadata: %v", q)
					}
				} else if r.URL.RawQuery != "" {
					t.Errorf("unexpected query: %s", r.URL.RawQuery)
				}
				if action == "create" || action == "update" {
					var actual map[string]any
					if err := json.NewDecoder(r.Body).Decode(&actual); err != nil {
						t.Error(err)
					}
					if !reflect.DeepEqual(actual, body) {
						t.Errorf("original changed: %#v", actual)
					}
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"schemaVersion":1,"outcome":"applied"}`))
			}))
			defer server.Close()
			if err := auth.SaveToken(&auth.TokenData{APIOrigin: server.URL, AccessToken: "test", ExpiresAt: time.Now().Add(time.Hour), TenantGUID: id}); err != nil {
				t.Fatal(err)
			}
			api := client.NewAPIClient(server.URL, time.Second, true, client.RetryOptions{MaxRetries: 0}, logging.New(logging.LevelError))
			format := "json"
			cmd := buildDomainCommand(findDomainByName(t, "tag"), func() (*client.APIClient, error) { return api, nil }, logging.New(logging.LevelError), &format, &ExportConfig{})
			args := []string{action, "--confirm"}
			if action == "legacy-delete" {
				args[0] = "delete"
			}
			if action != "create" {
				args = append(args, id)
			}
			if action == "create" || action == "update" {
				args = append(args, "--from-json", file)
			}
			if action == "delete" {
				args = append(args, "--idempotency-key", operation, "--expected-updated-at", revision, "--mutation-outcome-version", "1")
			}
			cmd.SetArgs(args)
			if err := cmd.Execute(); err != nil || count != 1 {
				t.Fatalf("request failed: %v count=%d", err, count)
			}
		})
	}
}

func TestClassificationPublicGuidAndLocalConfirmation(t *testing.T) {
	for _, domain := range []string{"category", "tag"} {
		for _, name := range []string{"get", "update", "delete"} {
			action := findDomainAction(t, domain, name)
			if action.RESTPath != "by-guid/{guid}" || len(action.Args) != 1 || action.Args[0].Type != "non-empty-uuid" {
				t.Fatalf("%s %s lost GUID routing", domain, name)
			}
		}
		for _, name := range []string{"create", "update", "delete"} {
			confirmation := actionFlagByName(t, findDomainAction(t, domain, name), "confirm")
			if !confirmation.Required || !confirmation.MustBeTrue || !confirmation.LocalOnly {
				t.Fatalf("%s %s lost local confirmation", domain, name)
			}
		}
	}
}

func TestClassificationDeleteRejectsIncompleteMetadataBeforeTransport(t *testing.T) {
	const id = "11111111-1111-4111-8111-111111111111"
	for _, domain := range []string{"category", "tag"} {
		for _, flags := range [][]string{
			{"--idempotency-key", id},
			{"--idempotency-key", id, "--expected-updated-at", "2026-10-16T10:00:00.12345678Z", "--mutation-outcome-version", "1"},
			{"--idempotency-key", id, "--expected-updated-at", "2026-10-16T10:00:00Z", "--mutation-outcome-version", "2"},
		} {
			created := false
			format := "json"
			cmd := buildDomainCommand(findDomainByName(t, domain), func() (*client.APIClient, error) { created = true; return nil, nil }, logging.New(logging.LevelError), &format, &ExportConfig{})
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true
			cmd.SetArgs(append([]string{"delete", id, "--confirm"}, flags...))
			if err := cmd.Execute(); err == nil || created {
				t.Fatalf("%s metadata reached transport: %v", domain, err)
			}
		}
	}
}
