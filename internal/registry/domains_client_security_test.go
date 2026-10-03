package registry

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/uteamup/cli/internal/auth"
	"github.com/uteamup/cli/internal/client"
	"github.com/uteamup/cli/internal/logging"
)

func clientSecurityDomain(t *testing.T) *Domain {
	t.Helper()
	for _, domain := range DefaultRegistry.Domains() {
		if domain.Name == "client-security" {
			return domain
		}
	}
	t.Fatal("client-security domain missing")
	return nil
}

func TestClientSecurityPolicyUsesActiveTenantAndBackendOwnedTools(t *testing.T) {
	domain := clientSecurityDomain(t)
	expected := map[string]string{
		"get":       "UteamupTenantClientSecurityPolicyGet",
		"effective": "UteamupTenantClientSecurityPolicyEffective",
		"set":       "UteamupTenantClientSecurityPolicySet",
	}
	if domain.APIPath != "/api/tenantpolicies/client-security" || len(domain.Actions) != len(expected) {
		t.Fatal("unexpected client policy API surface")
	}
	for _, action := range domain.Actions {
		if action.ToolName != expected[action.Name] || len(action.Args) != 0 {
			t.Errorf("unexpected tool or public identity override: %+v", action)
		}
		if err := validateActionDefinition(action); err != nil {
			t.Errorf("%s definition: %v", action.Name, err)
		}
		if action.Name != "set" && len(action.Flags) != 0 {
			t.Errorf("%s must use authenticated active-tenant scope", action.Name)
		}
	}
}

func TestClientSecurityConfirmationFailsBeforeCreatingClient(t *testing.T) {
	for _, confirmation := range []string{"", "--confirm=false"} {
		t.Run(confirmation, func(t *testing.T) {
			created := false
			format := "json"
			command := buildDomainCommand(clientSecurityDomain(t), func() (*client.APIClient, error) {
				created = true
				return nil, nil
			}, logging.New(logging.LevelError), &format, &ExportConfig{})
			command.SilenceErrors = true
			command.SilenceUsage = true
			arguments := []string{"set", "--file", filepath.Join(t.TempDir(), "unused.json")}
			if confirmation != "" {
				arguments = append(arguments, confirmation)
			}
			command.SetArgs(arguments)
			if err := command.Execute(); err == nil || !strings.Contains(err.Error(), "confirm") {
				t.Fatalf("policy mutation accepted without explicit confirmation: %v", err)
			}
			if created {
				t.Fatal("client created before confirmation validation")
			}
		})
	}
}

func TestClientSecurityTransportPreservesVersionAndExcludesLocalConfirmation(t *testing.T) {
	const tenantGuid = "55555555-5555-4555-8555-555555555555"
	for _, action := range []string{"get", "effective", "set"} {
		t.Run(action, func(t *testing.T) {
			policy := map[string]any{
				"expectedPolicyVersion": float64(7), "offlineEnabled": true, "maxOfflineDays": float64(60),
				"localPassphraseAllowed": true, "osAuthenticationAllowed": false,
				"lockOnOsLock": true, "lockOnSleep": true, "lockOnSessionChange": true,
				"lockOnShutdown": true, "idleLockMinutes": float64(15), "allowPwaLifecycleFallback": true,
			}
			encoded, err := json.Marshal(policy)
			if err != nil {
				t.Fatal(err)
			}
			requestFile := writeRegistryJSONFixture(t, string(encoded))
			home := t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv("USERPROFILE", home)
			requests := 0
			server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
				requests++
				wantPath := "/api/tenantpolicies/client-security"
				wantMethod := http.MethodGet
				if action == "effective" {
					wantPath += "/effective"
				} else if action == "set" {
					wantMethod = http.MethodPut
				}
				if request.Method != wantMethod || request.URL.Path != wantPath || request.URL.RawQuery != "" {
					t.Errorf("unexpected request: %s %s", request.Method, request.URL.String())
				}
				if request.Header.Get("X-Tenant-Guid") != tenantGuid {
					t.Errorf("request does not use active tenant GUID")
				}
				if action == "set" {
					var actual map[string]any
					if err := json.NewDecoder(request.Body).Decode(&actual); err != nil {
						t.Errorf("decode request: %v", err)
					} else if !reflect.DeepEqual(actual, policy) {
						t.Errorf("reviewed policy changed or local confirmation leaked: %#v", actual)
					}
				}
				response.Header().Set("Content-Type", "application/json")
				_, _ = response.Write([]byte(`{"tenantGuid":"` + tenantGuid + `","policyVersion":8}`))
			}))
			t.Cleanup(server.Close)
			if err := auth.SaveToken(&auth.TokenData{
				APIOrigin: server.URL, AccessToken: "client-policy-contract-token",
				ExpiresAt: time.Now().Add(time.Hour), TenantGUID: tenantGuid,
			}); err != nil {
				t.Fatal(err)
			}
			apiClient := client.NewAPIClient(server.URL, time.Second, true,
				client.RetryOptions{MaxRetries: 0}, logging.New(logging.LevelError))
			format := "json"
			command := buildDomainCommand(clientSecurityDomain(t), func() (*client.APIClient, error) {
				return apiClient, nil
			}, logging.New(logging.LevelError), &format, &ExportConfig{})
			command.SilenceErrors = true
			command.SilenceUsage = true
			arguments := []string{action}
			if action == "set" {
				arguments = append(arguments, "--file", requestFile, "--confirm")
			}
			command.SetArgs(arguments)
			if err := command.Execute(); err != nil {
				t.Fatal(err)
			}
			if requests != 1 {
				t.Errorf("request count = %d, want 1", requests)
			}
		})
	}
}
