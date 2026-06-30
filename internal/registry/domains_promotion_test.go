package registry

import "testing"

func findPromotionDomain(t *testing.T) *Domain {
	t.Helper()
	for _, d := range DefaultRegistry.Domains() {
		if d.Name == "promotion" {
			return d
		}
	}
	t.Fatal("expected promotion domain to be registered")
	return nil
}

func findPromotionAction(t *testing.T, name string) *Action {
	t.Helper()
	d := findPromotionDomain(t)
	for i := range d.Actions {
		if d.Actions[i].Name == name {
			return &d.Actions[i]
		}
	}
	t.Fatalf("expected %q action on promotion domain", name)
	return nil
}

func TestPromotionDomainRegistered(t *testing.T) {
	d := findPromotionDomain(t)
	if d.Description == "" {
		t.Error("promotion domain must have a Description")
	}
	if d.APIPath != "/api/promotion" {
		t.Errorf("promotion APIPath = %q, want %q", d.APIPath, "/api/promotion")
	}
	wantAliases := map[string]bool{"promotions": true, "discount": true, "discounts": true}
	for _, a := range d.Aliases {
		delete(wantAliases, a)
	}
	if len(wantAliases) != 0 {
		t.Errorf("promotion domain missing aliases: %v (got %v)", wantAliases, d.Aliases)
	}
}

func TestPromotionActionsWired(t *testing.T) {
	expected := map[string]string{
		"list":        "UteamupPromotionList",
		"get":         "UteamupPromotionGet",
		"create":      "UteamupPromotionCreate",
		"update":      "UteamupPromotionUpdate",
		"archive":     "UteamupPromotionArchive",
		"grant":       "UteamupPromotionGrant",
		"grant-adhoc": "UteamupPromotionGrantAdhoc",
		"redemptions": "UteamupPromotionRedemptions",
		"revoke":      "UteamupPromotionRevokeRedemption",
	}
	for name, tool := range expected {
		a := findPromotionAction(t, name)
		if a.ToolName != tool {
			t.Errorf("action %q ToolName = %q, want %q", name, a.ToolName, tool)
		}
	}
}

func TestPromotionGuidRoutesAreGuidFirst(t *testing.T) {
	// get/update/archive/redemptions are keyed by the promotion GUID via by-guid/{guid}.
	for _, name := range []string{"get", "update", "archive", "redemptions"} {
		a := findPromotionAction(t, name)
		if len(a.Args) != 1 || a.Args[0].Name != "guid" || a.Args[0].Type != "string" {
			t.Errorf("action %q must take a single string `guid` arg, got %+v", name, a.Args)
		}
	}
	// revoke targets a redemption GUID under a nested path.
	revoke := findPromotionAction(t, "revoke")
	if revoke.RESTPath != "redemption/by-guid/{guid}" {
		t.Errorf("revoke RESTPath = %q, want %q", revoke.RESTPath, "redemption/by-guid/{guid}")
	}
	if revoke.HTTPMethod != "DELETE" {
		t.Errorf("revoke HTTPMethod = %q, want DELETE", revoke.HTTPMethod)
	}
}

func TestPromotionGrantWiring(t *testing.T) {
	grant := findPromotionAction(t, "grant")
	if grant.HTTPMethod != "POST" || grant.RESTPath != "by-guid/{guid}/grant" {
		t.Errorf("grant routing = %s %s, want POST by-guid/{guid}/grant", grant.HTTPMethod, grant.RESTPath)
	}
	// tenant-guid must be a required flag (camelCases to tenantGuid in the body).
	var foundTenant bool
	for _, f := range grant.Flags {
		if f.Name == "tenant-guid" {
			foundTenant = true
			if !f.Required {
				t.Error("grant --tenant-guid must be required")
			}
		}
	}
	if !foundTenant {
		t.Error("grant action must declare a --tenant-guid flag")
	}
}

func TestPlanGetMigratedToGuid(t *testing.T) {
	var planDomain *Domain
	for _, d := range DefaultRegistry.Domains() {
		if d.Name == "plan" {
			planDomain = d
		}
	}
	if planDomain == nil {
		t.Fatal("expected plan domain to be registered")
	}
	for i := range planDomain.Actions {
		if planDomain.Actions[i].Name == "get" {
			a := planDomain.Actions[i]
			if a.RESTPath != "by-guid/{guid}" {
				t.Errorf("plan get RESTPath = %q, want by-guid/{guid} (int id deprecated)", a.RESTPath)
			}
			if len(a.Args) != 1 || a.Args[0].Name != "guid" || a.Args[0].Type != "string" {
				t.Errorf("plan get must take a single string `guid` arg, got %+v", a.Args)
			}
			return
		}
	}
	t.Fatal("expected plan get action")
}
