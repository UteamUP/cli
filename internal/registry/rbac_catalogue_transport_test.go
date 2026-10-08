package registry

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/uteamup/cli/internal/auth"
	"github.com/uteamup/cli/internal/client"
	clierrors "github.com/uteamup/cli/internal/errors"
	"github.com/uteamup/cli/internal/logging"
)

// Exercise every registered action, including write, upload, download and MCP-only
// commands. A 403 must produce an error with no success output or transport fallback.
func TestEveryRegisteredActionPreservesPermissionDenial(t *testing.T) {
	const guid = "11111111-1111-4111-8111-111111111111"
	const body = `{"error":"Permission denied"}`
	directory := t.TempDir()
	t.Setenv("HOME", directory)
	t.Setenv("USERPROFILE", directory)
	file := filepath.Join(directory, "request.json")
	if err := os.WriteFile(file, []byte(`{}`), 0600); err != nil {
		t.Fatal(err)
	}
	var requests atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if r.Header.Get("X-Tenant-Guid") != guid || r.Header.Get("Authorization") != "Bearer rbac-denied" {
			t.Error("command lost the authenticated tenant boundary")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(body))
	}))
	defer server.Close()
	for _, method := range []string{"login", "saml", "apikey"} {
		if err := auth.SaveToken(&auth.TokenData{APIOrigin: server.URL, AccessToken: "rbac-denied",
			AuthMethod: method, ExpiresAt: time.Now().Add(time.Hour), TenantGUID: guid}); err != nil {
			t.Fatal(err)
		}
		api := client.NewAPIClient(server.URL, time.Second, true, client.RetryOptions{MaxRetries: 0}, logging.New(logging.LevelError))
		for _, domain := range DefaultRegistry.Domains() {
			for _, action := range domain.Actions {
				t.Run(method+"/"+domain.Name+"/"+action.Name, func(t *testing.T) {
					args := []string{}
					for _, argument := range action.Args {
						value := rbacInputValue(argument.Type, guid)
						if len(argument.AllowedValues) > 0 {
							value = argument.AllowedValues[0]
						}
						args = append(args, value)
					}
					for _, flag := range action.Flags {
						if !flag.Required && !flag.MustBeTrue {
							continue
						}
						value := rbacInputValue(flag.Type, guid)
						if len(flag.AllowedValues) > 0 {
							value = flag.AllowedValues[0]
						}
						if flag.JSONFile || flag.RootJSONObjectFile || flag.UploadFile {
							value = file
						}
						if flag.StrongETag {
							value = `"rbac-version"`
						}
						if flag.Sensitive {
							protected := filepath.Join(t.TempDir(), "protected-input")
							if err := os.WriteFile(protected, []byte(value), 0600); err != nil {
								t.Fatal(err)
							}
							args = append(args, "--"+flag.Name+"-file="+protected)
						} else {
							args = append(args, "--"+flag.Name+"="+value)
						}
					}
					format := "json"
					command := buildActionCommand(domain, action, func() (*client.APIClient, error) { return api, nil },
						logging.New(logging.LevelError), &format, &ExportConfig{})
					command.SilenceErrors, command.SilenceUsage = true, true
					command.SetArgs(args)
					before := requests.Load()
					var runErr error
					stdout := captureRegistryStdout(t, func() { runErr = command.Execute() })
					if method != "apikey" && action.MCPOnly {
						var authErr *clierrors.AuthError
						if !errors.As(runErr, &authErr) || requests.Load() != before || stdout != "" {
							t.Fatalf("API-key-only action did not fail closed for a human login: %v", runErr)
						}
						return
					}
					var apiErr *clierrors.APIError
					if !errors.As(runErr, &apiErr) || apiErr.StatusCode != http.StatusForbidden || apiErr.Body != body {
						t.Fatalf("permission denial was not preserved: %v", runErr)
					}
					if requests.Load()-before != 1 || stdout != "" {
						t.Fatalf("denied action: requests=%d output=%q", requests.Load()-before, stdout)
					}
				})
			}
		}
	}
}

func rbacInputValue(kind, guid string) string {
	switch kind {
	case "uuid", "non-empty-uuid", "stringSlice":
		return guid
	case "int", "float", "decimal":
		return "1"
	case "bool":
		return "true"
	default:
		return "rbac-test"
	}
}
