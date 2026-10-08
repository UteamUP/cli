package registry

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/uteamup/cli/internal/client"
)

func TestTenantSamlDomainActionsMirrorOwnerRestAndMcpContracts(t *testing.T) {
	const tenant = "11111111-1111-4111-8111-111111111111"
	const domain = "22222222-2222-4222-8222-222222222222"
	for _, scenario := range []struct {
		action, method, path string
		domainArg            bool
	}{
		{"list", "GET", "/api/tenant/saml/domains", false},
		{"register", "POST", "/api/tenant/saml/domains", false},
		{"verify", "POST", "/api/tenant/saml/domains/" + domain + "/verify", true},
		{"disable", "DELETE", "/api/tenant/saml/domains/" + domain, true},
	} {
		t.Run(scenario.action, func(t *testing.T) {
			name := "saml-domain-" + scenario.action
			action := findDomainAction(t, "tenant", name)
			if action.ToolName != "uteamup_tenant_saml_domain_"+scenario.action || action.HTTPMethod != scenario.method || action.MCPOnly {
				t.Fatalf("domain contract drift: %+v", action)
			}
			for _, arg := range action.Args {
				if arg.Type != "non-empty-uuid" {
					t.Fatal("domain boundary exposes non-GUID identifiers")
				}
			}
			apiClient := projectTransportClient(t, "saml", client.RetryOptions{}, func(w http.ResponseWriter, r *http.Request) {
				if r.Method != scenario.method || r.URL.Path != scenario.path || r.URL.Query().Get("tenantGuid") != tenant {
					t.Errorf("domain operation changed scope or route: %s %s", r.Method, r.URL)
				}
				if scenario.action == "register" {
					var body map[string]any
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil || !reflect.DeepEqual(body, map[string]any{"domain": "iteggs.com"}) {
						t.Errorf("registration must send only the owning domain DTO: %v", body)
					}
				}
				if scenario.action == "disable" {
					w.WriteHeader(204)
					return
				}
				_, _ = w.Write([]byte(`{"enabled":false}`))
			})
			args := []string{name, tenant}
			if scenario.domainArg {
				args = append(args, domain)
			}
			if scenario.action == "register" {
				args = append(args, "--domain", "iteggs.com")
			}
			if err := executeProjectTransport(t, apiClient, "tenant", args); err != nil {
				t.Fatal(err)
			}
		})
	}
	for _, name := range []string{"saml-domain-list", "saml-domain-register", "saml-domain-verify", "saml-domain-disable"} {
		action := findDomainAction(t, "tenant", name)
		if action.Args[0].QueryName != "tenantGuid" || !action.RejectExtraArgs || !strings.HasPrefix(action.RESTPath, "saml/domains") {
			t.Fatalf("tenant authority must stay outside the body: %+v", action)
		}
	}
}
