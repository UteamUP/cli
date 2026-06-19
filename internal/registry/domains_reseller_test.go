package registry

import "testing"

func findResellerDomain() *Domain {
	for _, d := range DefaultRegistry.Domains() {
		if d.Name == "reseller" {
			return d
		}
	}
	return nil
}

func TestResellerDomainRegistered(t *testing.T) {
	d := findResellerDomain()
	if d == nil {
		t.Fatal("expected reseller domain to be registered")
	}
	if d.Description == "" {
		t.Error("reseller domain must have a Description")
	}
	if len(d.Aliases) == 0 {
		t.Error("reseller domain should have aliases")
	}
}

func TestResellerActionsWired(t *testing.T) {
	d := findResellerDomain()
	if d == nil {
		t.Fatal("expected reseller domain to be registered")
	}
	expected := map[string]string{
		"list":             "UteamupResellerList",
		"get":              "UteamupResellerGet",
		"applications":     "UteamupResellerApplicationsList",
		"tenants":          "UteamupResellerTenantsList",
		"earnings":         "UteamupResellerEarningsList",
		"program-defaults": "UteamupResellerProgramDefaultsGet",
		// New actions — 2026-06 reseller program overhaul
		"application-get": "UteamupResellerMyApplicationGet",
		"checklist":        "UteamupResellerApplicationChecksGet",
		"referral-codes":   "UteamupResellerMyReferralCodesGet",
		"tenant-manager":   "UteamupResellerMyTenantManagerGet",
	}
	got := map[string]string{}
	for _, a := range d.Actions {
		got[a.Name] = a.ToolName
	}
	for action, tool := range expected {
		if got[action] != tool {
			t.Errorf("expected reseller action %q to map to %q, got %q", action, tool, got[action])
		}
	}
}

func TestResellerNewActionsHaveNoSpoofingFlags(t *testing.T) {
	d := findResellerDomain()
	if d == nil {
		t.Fatal("expected reseller domain to be registered")
	}
	// Self-serve tools must take zero flags (identity comes from Bearer token, not a CLI arg).
	noFlagActions := []string{"application-get", "referral-codes", "tenant-manager"}
	actionMap := map[string]*Action{}
	for i := range d.Actions {
		actionMap[d.Actions[i].Name] = &d.Actions[i]
	}
	for _, name := range noFlagActions {
		a, ok := actionMap[name]
		if !ok {
			t.Errorf("expected action %q to be registered", name)
			continue
		}
		if len(a.Flags) != 0 {
			t.Errorf("action %q must have no flags (spoofing guard), got %d", name, len(a.Flags))
		}
	}
}

func TestResellerChecklistFlagPresent(t *testing.T) {
	d := findResellerDomain()
	if d == nil {
		t.Fatal("expected reseller domain to be registered")
	}
	for _, a := range d.Actions {
		if a.Name != "checklist" {
			continue
		}
		if len(a.Flags) == 0 {
			t.Error("checklist action must have at least one flag (application-guid)")
			return
		}
		found := false
		for _, f := range a.Flags {
			if f.Name == "application-guid" && f.Required {
				found = true
				break
			}
		}
		if !found {
			t.Error("checklist action must have a required application-guid flag")
		}
		return
	}
	t.Error("checklist action not found in reseller domain")
}
