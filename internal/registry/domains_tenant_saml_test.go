package registry

import (
	"encoding/json"
	"net/http"
	"reflect"
	"sync/atomic"
	"testing"

	"github.com/spf13/cobra"
	"github.com/uteamup/cli/internal/client"
)

func TestTenantSamlCommandsMirrorBackendRoutesAndTools(t *testing.T) {
	for _, scenario := range []struct{ action, tool, method, path string }{
		{"get-saml", "uteamup_tenant_saml_get", "GET", "saml"},
		{"update-saml", "uteamup_tenant_saml_update", "PUT", "saml"},
		{"test-saml", "uteamup_tenant_saml_test", "POST", "saml/test"},
	} {
		action := findDomainAction(t, "tenant", scenario.action)
		if action.ToolName != scenario.tool || action.HTTPMethod != scenario.method || action.RESTPath != scenario.path || action.MCPOnly {
			t.Fatalf("unexpected SAML transport for %s: %+v", scenario.action, action)
		}
		if len(action.Args) != 1 || action.Args[0].Type != "non-empty-uuid" || action.Args[0].QueryName != "tenantGuid" || !action.Args[0].Required || !action.RejectExtraArgs {
			t.Fatalf("tenant must be a required GUID query, not part of the body: %+v", action.Args)
		}
		path, consumed := buildRESTPath(findDomain("tenant"), *action, map[string]any{})
		if path != "/api/tenant/"+scenario.path || len(consumed) != 0 {
			t.Fatalf("wrong SAML configuration path: %s", path)
		}
		for _, invalid := range []string{"7", "../other", "00000000-0000-0000-0000-000000000000"} {
			if err := validateActionInput(&cobra.Command{}, []string{invalid}, *action); err == nil {
				t.Errorf("%s accepted an invalid public identifier", scenario.action)
			}
		}
	}
	for _, actionName := range []string{"update-saml", "test-saml"} {
		action := findDomainAction(t, "tenant", actionName)
		if len(action.Flags) != 1 || !action.Flags[0].Required || !action.Flags[0].RootJSONObjectFile {
			t.Fatalf("%s must forward the owning DTO as root JSON", actionName)
		}
	}
	if !findDomainAction(t, "tenant", "test-saml").DisableResponseExport {
		t.Fatal("setup start response must not be automatically exported")
	}
}

func TestTenantSamlConfigurationUsesAuthenticatedRestAndQueryScope(t *testing.T) {
	for _, mode := range []string{"login", "saml", "apikey"} {
		for _, scenario := range []struct{ action, method string }{{"get-saml", "GET"}, {"update-saml", "PUT"}} {
			t.Run(mode+"/"+scenario.action, func(t *testing.T) {
				const tenant = "11111111-1111-4111-8111-111111111111"
				const payload = `{"companyCode":"iteggs","metadataUrl":"https://idp.example/metadata.xml","emailAttribute":"email","enabled":false}`
				var want map[string]any
				_ = json.Unmarshal([]byte(payload), &want)
				var calls atomic.Int32
				apiClient := projectTransportClient(t, mode, client.RetryOptions{}, func(w http.ResponseWriter, r *http.Request) {
					calls.Add(1)
					if r.URL.Path != "/api/tenant/saml" || r.Method != scenario.method || r.URL.Query().Get("tenantGuid") != tenant {
						t.Errorf("wrong SAML configuration transport: %s %s", r.Method, r.URL)
					}
					if scenario.method == "PUT" {
						var got map[string]any
						if err := json.NewDecoder(r.Body).Decode(&got); err != nil || !reflect.DeepEqual(got, want) {
							t.Errorf("configuration DTO or query-only tenant scope changed: %v", err)
						}
					}
					_, _ = w.Write([]byte(`{"enabled":false}`))
				})
				args := []string{scenario.action, tenant}
				if scenario.method == "PUT" {
					args = append(args, "--configuration-file", writeRegistryJSONFixture(t, payload))
				}
				if err := executeProjectTransport(t, apiClient, "tenant", args); err != nil {
					t.Fatal(err)
				}
				if calls.Load() != 1 {
					t.Fatal("configuration did not make exactly one owning REST request")
				}
			})
		}
	}
}
