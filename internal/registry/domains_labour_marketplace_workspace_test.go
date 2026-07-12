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
	if len(domain.Actions) != 3 {
		t.Fatalf("expected three actions, got %d", len(domain.Actions))
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
	timesheets := domain.Actions[1]
	if timesheets.Name != "timesheets" || timesheets.ToolName != "UteamupLabourAgreementTimesheets" {
		t.Fatalf("unexpected timesheets action %#v", timesheets)
	}
	if timesheets.RESTPath != "agreements/{agreementGuid}/timesheets" || timesheets.HTTPMethod != "GET" {
		t.Fatalf("unexpected timesheets REST contract %s %s", timesheets.HTTPMethod, timesheets.RESTPath)
	}
	if len(timesheets.Args) != 1 || timesheets.Args[0].Name != "agreementGuid" || !timesheets.Args[0].Required {
		t.Fatal("timesheets must require exactly one agreement GUID argument")
	}
	replacement := domain.Actions[2]
	if replacement.Name != "replace-worker" || replacement.ToolName != "UteamupLabourWorkerDispatchReplace" {
		t.Fatalf("unexpected replacement action %#v", replacement)
	}
	if replacement.RESTPath != "dispatches/{dispatchGuid}/replacement" || replacement.HTTPMethod != "POST" {
		t.Fatalf("unexpected replacement REST contract %s %s", replacement.HTTPMethod, replacement.RESTPath)
	}
	if len(replacement.Args) != 1 || replacement.Args[0].Name != "dispatchGuid" || !replacement.Args[0].Required {
		t.Fatal("replacement must require the dispatch GUID")
	}
	if len(replacement.Flags) != 2 || !replacement.Flags[0].Required || !replacement.Flags[1].Required {
		t.Fatal("replacement membership GUID and reason must both be required")
	}
}
