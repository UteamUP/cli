package registry

import "testing"

func TestAutomationDomainTargetsRealControllers(t *testing.T) {
	t.Parallel()
	domain := findDomain("automation")
	if domain == nil {
		t.Fatal("automation domain is not registered")
	}
	if domain.APIPath != "/api/automation" {
		t.Fatalf("API path = %q, want /api/automation", domain.APIPath)
	}

	cases := []struct {
		action string
		method string
		args   map[string]any
		path   string
	}{
		{"list", "POST", nil, "/api/automation/search"},
		{"get", "GET", map[string]any{"externalGuid": "a1"}, "/api/automation/by-guid/a1"},
		{"workflow-publish", "POST", map[string]any{"externalGuid": "a1", "note": "v2"}, "/api/automation/by-guid/a1/publish"},
		{"workflow-validate", "POST", map[string]any{"externalGuid": "a1"}, "/api/automation/by-guid/a1/validate"},
		{"trigger", "POST", map[string]any{"externalGuid": "a1"}, "/api/automation/by-guid/a1/trigger"},
		{"runs", "POST", nil, "/api/automation/runs/search"},
		{"run-get", "GET", map[string]any{"runGuid": "r1"}, "/api/automation/runs/r1"},
		{"run-cancel", "POST", map[string]any{"runGuid": "r1"}, "/api/automation/runs/r1/cancel"},
		{"run-stats", "GET", nil, "/api/automation/runs/stats"},
		{"catalog", "GET", nil, "/api/automation/catalog"},
		{"pause", "POST", nil, "/api/automation/settings/pause"},
	}
	for _, tc := range cases {
		action := findAction(domain, tc.action)
		if action == nil {
			t.Fatalf("action %q is not registered", tc.action)
		}
		if action.HTTPMethod != tc.method {
			t.Fatalf("%s: method = %q, want %q", tc.action, action.HTTPMethod, tc.method)
		}
		path, _ := buildRESTPath(domain, *action, tc.args)
		if path != tc.path {
			t.Fatalf("%s: path = %q, want %q", tc.action, path, tc.path)
		}
	}
}

func TestAutomationDomainExposesNoIntegerIdentifiers(t *testing.T) {
	t.Parallel()
	domain := findDomain("automation")
	if domain == nil {
		t.Fatal("automation domain is not registered")
	}
	for _, action := range domain.Actions {
		for _, arg := range action.Args {
			if arg.Name == "id" || arg.Name == "automationId" || arg.Name == "runId" {
				t.Fatalf("%s exposes integer identifier %q", action.Name, arg.Name)
			}
		}
	}
}
