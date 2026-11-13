package registry

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/uteamup/cli/internal/auth"
	"github.com/uteamup/cli/internal/client"
	"github.com/uteamup/cli/internal/logging"
)

// Customers and vendors are archive-only (2026-11-13): the CLI offers archive and restore, never delete.
func TestCustomerAndVendorHaveNoDeleteCommand(t *testing.T) {
	for _, name := range []string{"customer", "vendor"} {
		for _, action := range findDomainByName(t, name).Actions {
			if action.Name == "delete" || action.HTTPMethod == "DELETE" || strings.HasSuffix(action.ToolName, "Delete") {
				t.Fatalf("%s must be archive-only, found %+v", name, action)
			}
		}
	}
}

func TestCustomerMutationMetadataUsesNormalRoutesWithoutChangingOriginal(t *testing.T) {
	const id = "11111111-1111-4111-8111-111111111111"
	const operation = "22222222-2222-4222-8222-222222222222"
	const revision = "2026-10-16T10:00:00.1234567Z"
	for _, action := range []string{"create", "update"} {
		t.Run(action, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv("USERPROFILE", home)
			body := map[string]any{"name": "Reviewed", "contactGuids": []any{id}, "idempotencyKey": operation, "mutationOutcomeVersion": float64(1)}
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
				path := "/api/customer"
				method := "POST"
				if action != "create" {
					path += "/" + id
					method = "PUT"
				}
				if r.Method != method || r.URL.Path != path {
					t.Errorf("wrong normal route: %s %s", r.Method, r.URL)
				}
				if r.URL.RawQuery != "" {
					t.Errorf("unexpected query: %s", r.URL.RawQuery)
				}
				{
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
			cmd := buildDomainCommand(findDomainByName(t, "customer"), func() (*client.APIClient, error) { return api, nil }, logging.New(logging.LevelError), &format, &ExportConfig{})
			args := []string{action}
			if action != "create" {
				args = append(args, id)
			}
			args = append(args, "--confirm", "--from-json", file)
			cmd.SetArgs(args)
			if err := cmd.Execute(); err != nil || count != 1 {
				t.Fatalf("request failed: %v count=%d", err, count)
			}
		})
	}
}
