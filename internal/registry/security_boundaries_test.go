package registry

import (
	"github.com/uteamup/cli/internal/client"
	"github.com/uteamup/cli/internal/logging"
	"os"
	"path/filepath"
	"strings"
	"testing"
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

func protectedRegistrySecret(t *testing.T, value string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "protected-input")
	if err := os.WriteFile(path, []byte(value), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}
