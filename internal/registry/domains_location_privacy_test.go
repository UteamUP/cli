package registry

import (
	"strings"
	"testing"
)

func TestLocationPrivacyDomainWired(t *testing.T) {
	d := findDomain("location-privacy")
	if d == nil {
		t.Fatal("expected location-privacy domain to be registered")
	}
	if d.APIPath != "/api/locationprivacy" {
		t.Errorf("APIPath = %q, want /api/locationprivacy", d.APIPath)
	}
}

func TestLocationPrivacyActionsMapToTheMcpTools(t *testing.T) {
	cases := []struct {
		action, tool, method, path string
	}{
		{"grants", "UteamupLocationPrivacyMyGrants", "GET", "my/grants"},
		{"access-log", "UteamupLocationPrivacyMyAccessLog", "GET", "my/access-log"},
		{"revoke", "UteamupLocationPrivacyRevokeGrant", "POST", "grants/{grantGuid}/revoke"},
		{"request", "UteamupLocationPrivacyRequestShare", "POST", "requests"},
		{"policy-get", "UteamupLocationPrivacyGetPolicy", "GET", "policy"},
		{"policy-update", "UteamupLocationPrivacyUpdatePolicy", "PUT", "policy"},
	}
	for _, c := range cases {
		action := findDomainAction(t, "location-privacy", c.action)
		if action.ToolName != c.tool || action.HTTPMethod != c.method || action.RESTPath != c.path {
			t.Errorf("%s = %s %s %q, want %s %s %q", c.action, action.ToolName, action.HTTPMethod, action.RESTPath, c.tool, c.method, c.path)
		}
	}
}

func TestLocationPrivacyRevokeBuildsAGuidPath(t *testing.T) {
	d := findDomain("location-privacy")
	action := findDomainAction(t, "location-privacy", "revoke")
	if len(action.Args) != 1 || action.Args[0].Name != "grantGuid" || action.Args[0].Type != "non-empty-uuid" || !action.Args[0].Required {
		t.Fatalf("revoke must take one required non-empty-uuid grantGuid, got %+v", action.Args)
	}
	path, consumed := buildRESTPath(d, *action, map[string]any{"grantGuid": "11111111-1111-4111-8111-111111111111"})
	if path != "/api/locationprivacy/grants/11111111-1111-4111-8111-111111111111/revoke" {
		t.Errorf("revoke path = %q", path)
	}
	if len(consumed) != 1 || consumed[0] != "grantGuid" {
		t.Errorf("revoke must consume grantGuid from the body, got %v", consumed)
	}
}

func TestLocationPrivacyIdentifiersAreGuidOnly(t *testing.T) {
	d := findDomain("location-privacy")
	for _, action := range d.Actions {
		for _, arg := range action.Args {
			if strings.HasSuffix(strings.ToLower(arg.Name), "guid") && arg.Type != "uuid" && arg.Type != "non-empty-uuid" {
				t.Errorf("%s arg %s must be a uuid, got %q", action.Name, arg.Name, arg.Type)
			}
			if arg.Type == "int" {
				t.Errorf("%s arg %s must not be an integer identifier", action.Name, arg.Name)
			}
		}
		for _, flag := range action.Flags {
			if strings.HasSuffix(flag.Name, "-guid") && flag.Type != "uuid" && flag.Type != "non-empty-uuid" {
				t.Errorf("%s flag %s must be a uuid, got %q", action.Name, flag.Name, flag.Type)
			}
			if flag.Name == "id" || strings.HasSuffix(flag.Name, "-id") {
				t.Errorf("%s flag %s is an integer identifier", action.Name, flag.Name)
			}
		}
	}
}

func TestLocationPrivacySelfActionsTakeNoUserOrTenant(t *testing.T) {
	for _, name := range []string{"grants", "access-log", "revoke"} {
		action := findDomainAction(t, "location-privacy", name)
		for _, arg := range action.Args {
			lower := strings.ToLower(arg.Name)
			if strings.Contains(lower, "user") || strings.Contains(lower, "tenant") {
				t.Errorf("self-service %s must not take %s; the person comes from authentication", name, arg.Name)
			}
		}
		for _, flag := range action.Flags {
			if strings.Contains(flag.Name, "user") || strings.Contains(flag.Name, "tenant") {
				t.Errorf("self-service %s must not take --%s; the person comes from authentication", name, flag.Name)
			}
		}
	}
}

func TestLocationPrivacyHasNoApproveOrGrantAction(t *testing.T) {
	d := findDomain("location-privacy")
	for _, action := range d.Actions {
		for _, forbidden := range []string{"approve", "decline", "grant-create", "create", "emergency"} {
			if strings.Contains(action.Name, forbidden) {
				t.Errorf("location-privacy must not expose %q: consent comes only from the person in the app", action.Name)
			}
		}
	}
}

func TestLocationPrivacyRequestFlagsMatchTheBackendModel(t *testing.T) {
	action := findDomainAction(t, "location-privacy", "request")
	for _, name := range []string{"subject-user-guid", "scopes", "reason"} {
		f := findFlag(action, name)
		if f == nil || !f.Required {
			t.Errorf("request must have a required --%s flag, got %+v", name, f)
		}
	}
	if f := findFlag(action, "duration-days"); f == nil || f.Type != "int" {
		t.Errorf("request must have an int --duration-days flag, got %+v", f)
	}
	if f := findFlag(findDomainAction(t, "location-privacy", "policy-update"), "policy-file"); f == nil || !f.RootJSONObjectFile || !f.Required {
		t.Errorf("policy-update must read a required root JSON object file, got %+v", f)
	}
}
