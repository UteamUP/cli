package registry

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/uteamup/cli/internal/auth"
	"github.com/uteamup/cli/internal/client"
	clierrors "github.com/uteamup/cli/internal/errors"
	"github.com/uteamup/cli/internal/logging"
)

func TestRBACCommandsUseTenantRESTForBothAuthenticationMethods(t *testing.T) {
	const guid = "11111111-1111-4111-8111-111111111111"
	for _, authMethod := range []string{"login", "api-key"} {
		for _, tc := range []struct {
			domain, method, path string
			args                 []string
		}{
			{"role", "GET", "/api/role/tenant", []string{"list"}},
			{"asset-type", "GET", "/api/asset-type", []string{"list", "--include-inactive"}},
			{"asset-type", "GET", "/api/asset-type/" + guid, []string{"get", guid}},
			{"asset-type", "POST", "/api/asset-type", []string{"create"}},
			{"asset-type", "PUT", "/api/asset-type/" + guid, []string{"update", guid}},
			{"asset-type", "DELETE", "/api/asset-type/" + guid, []string{"delete", guid}},
			{"asset-type", "POST", "/api/asset-type/attributes/" + guid + "/fill-existing", []string{"fill-existing-assets", guid, "--value", "12.5"}},
			{"asset-type", "GET", "/api/asset-type/by-guid/" + guid + "/compatible-parts", []string{"compatible-parts", guid}},
		} {
			t.Run(authMethod+"/"+tc.domain+"/"+tc.args[0], func(t *testing.T) {
				directory := t.TempDir()
				t.Setenv("HOME", directory)
				t.Setenv("USERPROFILE", directory)
				seen := false
				server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					seen = true
					if r.Method != tc.method || r.URL.Path != tc.path {
						t.Errorf("request = %s %s, want %s %s", r.Method, r.URL.Path, tc.method, tc.path)
					}
					if r.Header.Get("Authorization") != "Bearer rbac-human-token" || r.Header.Get("X-Tenant-Guid") != guid {
						t.Error("human identity and tenant scope must reach REST unchanged")
					}
					if tc.args[0] == "list" && tc.domain == "asset-type" && r.URL.Query().Get("includeInactive") != "true" {
						t.Error("includeInactive must be a query parameter")
					}
					if tc.args[0] == "create" || tc.args[0] == "update" {
						var body map[string]any
						if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body["name"] != "Reviewed type" || body["model"] != nil {
							t.Errorf("REST must receive the reviewed model at the body root: %v", body)
						}
					}
					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write([]byte(`[]`))
				}))
				defer server.Close()
				if err := auth.SaveToken(&auth.TokenData{APIOrigin: server.URL, AccessToken: "rbac-human-token", AuthMethod: authMethod,
					ExpiresAt: time.Now().Add(time.Hour), TenantGUID: guid}); err != nil {
					t.Fatal(err)
				}
				args := append([]string{}, tc.args...)
				if tc.args[0] == "create" || tc.args[0] == "update" {
					file := filepath.Join(directory, "type.json")
					if err := os.WriteFile(file, []byte(`{"name":"Reviewed type"}`), 0600); err != nil {
						t.Fatal(err)
					}
					args = append(args, "--request-file", file)
				}
				api := client.NewAPIClient(server.URL, time.Second, true, client.RetryOptions{MaxRetries: 0}, logging.New(logging.LevelError))
				format := "json"
				command := buildDomainCommand(findDomainByName(t, tc.domain), func() (*client.APIClient, error) { return api, nil },
					logging.New(logging.LevelError), &format, &ExportConfig{})
				command.SetArgs(args)
				if err := command.Execute(); err != nil {
					t.Fatal(err)
				}
				if !seen {
					t.Fatal("command never reached the REST authorization boundary")
				}
			})
		}
	}
}

func TestRBACCommandsPreserveRESTPermissionDenials(t *testing.T) {
	for _, authMethod := range []string{"login", "api-key"} {
		for _, domain := range []string{"role", "asset-type"} {
			t.Run(authMethod+"/"+domain, func(t *testing.T) {
				directory := t.TempDir()
				t.Setenv("HOME", directory)
				t.Setenv("USERPROFILE", directory)
				requests := 0
				const body = `{"error":"Permission denied"}`
				server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					requests++
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusForbidden)
					_, _ = w.Write([]byte(body))
				}))
				t.Cleanup(server.Close)
				if err := auth.SaveToken(&auth.TokenData{APIOrigin: server.URL, AccessToken: "rbac-denied-token",
					AuthMethod: authMethod, ExpiresAt: time.Now().Add(time.Hour),
					TenantGUID: "11111111-1111-4111-8111-111111111111"}); err != nil {
					t.Fatal(err)
				}
				api := client.NewAPIClient(server.URL, time.Second, true, client.RetryOptions{MaxRetries: 0}, logging.New(logging.LevelError))
				format := "json"
				command := buildDomainCommand(findDomainByName(t, domain), func() (*client.APIClient, error) { return api, nil },
					logging.New(logging.LevelError), &format, &ExportConfig{})
				command.SilenceErrors, command.SilenceUsage = true, true
				command.SetArgs([]string{"list"})
				var runErr error
				stdout := captureRegistryStdout(t, func() { runErr = command.Execute() })
				var apiErr *clierrors.APIError
				if !errors.As(runErr, &apiErr) || apiErr.StatusCode != http.StatusForbidden || apiErr.Body != body {
					t.Fatalf("server denial was not preserved: %v", runErr)
				}
				if requests != 1 || stdout != "" {
					t.Fatalf("denied command: requests=%d stdout=%q", requests, stdout)
				}
			})
		}
	}
}
