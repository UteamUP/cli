package registry

import "testing"

func TestMenuVisibilityDomainIsRouted(t *testing.T) {
	domain := findDomain("menu-visibility")
	if domain == nil {
		t.Fatal("menu-visibility domain is not registered")
	}
	if domain.APIPath != "/api/menuvisibility" {
		t.Fatalf("menu-visibility APIPath = %q", domain.APIPath)
	}

	cases := []struct {
		action string
		method string
		path   string
		tool   string
	}{
		{"get", "GET", "", "UteamupMenuVisibilityGet"},
		{"set-tenant", "PUT", "tenant", "UteamupMenuVisibilityTenantSet"},
		{"set-me", "PUT", "me", "UteamupMenuVisibilityMeSet"},
	}

	for _, tc := range cases {
		action := findAction(domain, tc.action)
		if action == nil {
			t.Fatalf("menu-visibility %s action is missing", tc.action)
		}
		if action.HTTPMethod != tc.method || action.RESTPath != tc.path {
			t.Fatalf("menu-visibility %s route = method %q path %q, want %q %q",
				tc.action, action.HTTPMethod, action.RESTPath, tc.method, tc.path)
		}
		if action.ToolName != tc.tool {
			t.Fatalf("menu-visibility %s ToolName = %q, want %q", tc.action, action.ToolName, tc.tool)
		}
	}
}
