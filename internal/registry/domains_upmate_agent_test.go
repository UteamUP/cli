package registry

import (
	"strings"
	"testing"
)

func TestUpmateAgentDomainTargetsTheAgentController(t *testing.T) {
	t.Parallel()
	domain := findDomain("upmate-agent")
	if domain == nil {
		t.Fatal("upmate-agent domain is not registered")
	}
	if domain.APIPath != "/api/upmateagent" {
		t.Fatalf("API path = %q, want /api/upmateagent", domain.APIPath)
	}

	cases := []struct {
		action string
		method string
		args   map[string]any
		path   string
	}{
		{"list", "GET", nil, "/api/upmateagent"},
		{"get", "GET", map[string]any{"agentGuid": "a1"}, "/api/upmateagent/by-guid/a1"},
		{"capabilities", "GET", nil, "/api/upmateagent/capabilities"},
		{"run", "POST", map[string]any{"agentGuid": "a1", "entityType": "Workorder"}, "/api/upmateagent/by-guid/a1/run"},
		{"runs", "GET", map[string]any{"agentGuid": "a1"}, "/api/upmateagent/by-guid/a1/runs"},
		{"run-get", "GET", map[string]any{"runGuid": "r1"}, "/api/upmateagent/runs/r1"},
		{"run-cancel", "POST", map[string]any{"runGuid": "r1"}, "/api/upmateagent/runs/r1/cancel"},
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

func TestUpmateAgentDomainExposesNoIntegerIdentifiers(t *testing.T) {
	t.Parallel()
	domain := findDomain("upmate-agent")
	if domain == nil {
		t.Fatal("upmate-agent domain is not registered")
	}
	for _, action := range domain.Actions {
		for _, arg := range action.Args {
			if arg.Type == "int" && (arg.Name == "id" || strings.HasSuffix(arg.Name, "Id")) {
				t.Fatalf("%s exposes integer identifier %q", action.Name, arg.Name)
			}
		}
	}
}
