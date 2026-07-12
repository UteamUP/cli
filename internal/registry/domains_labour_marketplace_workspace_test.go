package registry

import "testing"

func TestLabourMarketplaceWorkspaceDomain(t *testing.T) {
	var domain *Domain
	for _, candidate := range DefaultRegistry.Domains() {
		if candidate.Name == "labour-marketplace-workspace" {
			domain = candidate
			break
		}
	}
	if domain == nil {
		t.Fatal("labour-marketplace-workspace domain not registered")
	}
	if domain.APIPath != "/api/labour-marketplace" {
		t.Fatalf("unexpected API path %q", domain.APIPath)
	}
	if len(domain.Actions) != 1 {
		t.Fatalf("expected one action, got %d", len(domain.Actions))
	}
	action := domain.Actions[0]
	if action.Name != "me" || action.ToolName != "UteamupLabourMarketplaceWorkspaceMe" {
		t.Fatalf("unexpected workspace action %#v", action)
	}
	if action.RESTPath != "workspace/me" || action.HTTPMethod != "GET" {
		t.Fatalf("unexpected REST contract %s %s", action.HTTPMethod, action.RESTPath)
	}
	if len(action.Args) != 0 || len(action.Flags) != 0 {
		t.Fatal("current-user workspace must not accept party or user selectors")
	}
}
