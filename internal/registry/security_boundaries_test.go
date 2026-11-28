package registry

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/uteamup/cli/internal/client"
	"github.com/uteamup/cli/internal/logging"
)

func TestRouteValuesRejectEncodedAndLiteralSyntax(t *testing.T) {
	action := Action{Name: "get", RESTPath: "by-guid/{guid}"}
	for _, v := range []string{"../admin", "%2fadmin", "id?x=y", "..", "a#b"} {
		if validateRouteValues(action, map[string]any{"guid": v}) == nil {
			t.Fatal(v)
		}
	}
	path, _ := expandPathTemplate("by-guid/{guid}", map[string]any{"guid": "with space"})
	if path != "by-guid/with%20space" {
		t.Fatal(path)
	}
}
func TestSensitiveArgvIsRejectedBeforeClientCreation(t *testing.T) {
	called := false
	factory := func() (*client.APIClient, error) { called = true; return nil, nil }
	format := "json"
	action := Action{Name: "redeem", Flags: []FlagDef{{Name: "token", Type: "string", Sensitive: true, Required: true}}}
	command := buildActionCommand(&Domain{Name: "test"}, action, factory, logging.New(logging.LevelError), &format, nil)
	command.SetArgs([]string{"--token", "secret"})
	err := command.Execute()
	if err == nil || !strings.Contains(err.Error(), "argv") || called {
		t.Fatalf("err=%v client=%v", err, called)
	}
}

func TestActualBookingVerificationCredentialRejectsArgvBeforeClientCreation(t *testing.T) {
	called := false
	factory := func() (*client.APIClient, error) { called = true; return nil, nil }
	format := "json"
	for _, domain := range DefaultRegistry.domains {
		for _, action := range domain.Actions {
			if action.ToolName != "UteamupSalesBookingVerify" {
				continue
			}
			command := buildActionCommand(domain, action, factory, logging.New(logging.LevelError), &format, nil)
			command.SetArgs([]string{"1", "--token", "private-one-time-code"})
			if err := command.Execute(); err == nil || !strings.Contains(err.Error(), "argv") || called {
				t.Fatalf("credential protection failed: %v", err)
			}
			return
		}
	}
	t.Fatal("booking verification action missing")
}

func TestDeviceTokenMutationRejectsArgvBeforeClientCreation(t *testing.T) {
	for _, actionName := range []string{"register", "delete"} {
		called := false
		factory := func() (*client.APIClient, error) { called = true; return nil, nil }
		format := "json"
		var domain *Domain
		var action *Action
		for i := range DefaultRegistry.domains {
			if DefaultRegistry.domains[i].Name != "devicetoken" {
				continue
			}
			domain = DefaultRegistry.domains[i]
			for j := range domain.Actions {
				if domain.Actions[j].Name == actionName {
					action = &domain.Actions[j]
				}
			}
		}
		if action == nil {
			t.Fatalf("device token %s action is missing", actionName)
		}
		command := buildActionCommand(domain, *action, factory, logging.New(logging.LevelError), &format, nil)
		args := []string{"--token", "device-token-canary"}
		if actionName == "register" {
			args = append(args, "--platform", "ios")
		}
		command.SetArgs(args)
		err := command.Execute()
		if err == nil || !strings.Contains(err.Error(), "argv") || called {
			t.Fatalf("%s sensitive token argv error=%v client=%v", actionName, err, called)
		}
	}
}
func TestCredentialActionsDisableExport(t *testing.T) {
	for _, pair := range [][2]string{{"apikey", "create"}, {"admin-user", "reset-password"}} {
		_ = pair
	}
	for _, domain := range DefaultRegistry.domains {
		for _, action := range domain.Actions {
			if action.ToolName == "UteamupTenantApiKeyCreate" || action.ToolName == "UteamupAdminUserResetPassword" {
				if !action.DisableResponseExport {
					t.Fatal(action.ToolName)
				}
			}
		}
	}
}

func TestDocumentCapabilityActionsDisableExport(t *testing.T) {
	for _, name := range []string{"ticket", "download-link"} {
		if !findDomainAction(t, "document-upload", name).DisableResponseExport {
			t.Errorf("document-upload %s must not export capability URLs", name)
		}
	}
}

func TestDocumentUploadResponseExportBoundary(t *testing.T) {
	const guid = "11111111-1111-4111-8111-111111111111"
	for _, tc := range []struct {
		name    string
		args    []string
		payload string
		export  bool
	}{
		{"ticket", []string{"--file-name", "report.pdf", "--content-type", "application/pdf", "--size-bytes", "42"},
			`{"uploadUrl":"https://api.uteamup.com/api/mcp/upload-tickets/upload-canary"}`, false},
		{"download-link", []string{guid},
			`{"downloadUrl":"https://api.uteamup.com/api/mcp/download-links/download-canary"}`, false},
		{"status", []string{guid}, `{"status":"pending"}`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			action := findDomainAction(t, "document-upload", tc.name)
			var calls atomic.Int32
			apiClient := projectTransportClient(t, "apikey", client.RetryOptions{MaxRetries: 0}, func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				var request struct {
					ID     int                   `json:"id"`
					Method string                `json:"method"`
					Params client.ToolCallParams `json:"params"`
				}
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Errorf("decode request: %v", err)
				}
				if r.Method != http.MethodPost || r.URL.Path != "/mcp" || request.Method != "tools/call" || request.Params.Name != mcpToolName(action.ToolName) {
					t.Error("request did not reach the registered MCP tool")
				}
				_ = json.NewEncoder(w).Encode(map[string]any{
					"jsonrpc": "2.0", "id": request.ID,
					"result": map[string]any{"content": []map[string]string{{"type": "text", "text": tc.payload}}},
				})
			})
			exportDir := t.TempDir()
			format := "json"
			command := buildActionCommand(findDomain("document-upload"), *action,
				func() (*client.APIClient, error) { return apiClient, nil }, logging.New(logging.LevelError),
				&format, &ExportConfig{Enabled: true, Dir: exportDir})
			command.SetArgs(tc.args)
			var runErr error
			stdout := captureRegistryStdout(t, func() { runErr = command.Execute() })
			if runErr != nil || calls.Load() != 1 {
				t.Fatalf("execution error = %v, requests = %d; want one successful request", runErr, calls.Load())
			}
			var got, want map[string]any
			if err := json.Unmarshal([]byte(stdout), &got); err != nil {
				t.Fatalf("decode stdout: %v", err)
			}
			if err := json.Unmarshal([]byte(tc.payload), &want); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatal("response must remain available on stdout")
			}
			if tc.export {
				assertOnlyFiles(t, exportDir, "document-upload_status.json")
			} else {
				assertOnlyFiles(t, exportDir)
			}
		})
	}
}

func protectedRegistrySecret(t *testing.T, value string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "protected-input")
	if err := os.WriteFile(path, []byte(value), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}
