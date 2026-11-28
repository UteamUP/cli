package registry

import "testing"

// UTE-1606: the CLI mirrors the backend MCP tool UteamupBillingCatalogApprovalsPending and the
// REST read GET /api/billing/admin/governance/pending-approvals (global-admin maker-checker queue).
func TestBillingCatalogDomainRoutes(t *testing.T) {
	t.Parallel()
	domain := findDomain("billing-catalog")
	if domain == nil {
		t.Fatal("billing-catalog domain is not registered")
	}
	if domain.APIPath != "/api/billing/admin/governance" {
		t.Fatalf("APIPath = %q, want /api/billing/admin/governance", domain.APIPath)
	}
	var pending *Action
	for i := range domain.Actions {
		if domain.Actions[i].Name == "pending-approvals" {
			pending = &domain.Actions[i]
		}
	}
	if pending == nil {
		t.Fatal("pending-approvals action is not registered")
	}
	if pending.ToolName != "UteamupBillingCatalogApprovalsPending" {
		t.Fatalf("ToolName = %q, want UteamupBillingCatalogApprovalsPending", pending.ToolName)
	}
	if pending.HTTPMethod != "" && pending.HTTPMethod != "GET" {
		t.Fatalf("HTTPMethod = %q, want a derived GET", pending.HTTPMethod)
	}
	path, consumed := buildRESTPath(domain, *pending, map[string]any{})
	if path != "/api/billing/admin/governance/pending-approvals" {
		t.Fatalf("path = %q, want /api/billing/admin/governance/pending-approvals", path)
	}
	if len(consumed) != 0 {
		t.Fatalf("consumed %d path args, want 0", len(consumed))
	}
	for _, arg := range pending.Args {
		if arg.Type == "int" {
			t.Fatalf("pending-approvals declares int identity arg %s — public identifiers must be GUIDs", arg.Name)
		}
	}
}
