package registry

import "testing"

func findPartnerDomain() *Domain {
	for _, d := range DefaultRegistry.Domains() {
		if d.Name == "partner" {
			return d
		}
	}
	return nil
}

func TestPartnerDomainRegistered(t *testing.T) {
	d := findPartnerDomain()
	if d == nil {
		t.Fatal("expected partner domain to be registered")
	}
	if d.Description == "" {
		t.Error("partner domain must have a Description")
	}
	if len(d.Aliases) == 0 {
		t.Error("partner domain should have aliases")
	}
}

func TestPartnerActionsWired(t *testing.T) {
	d := findPartnerDomain()
	if d == nil {
		t.Fatal("expected partner domain to be registered")
	}
	expected := map[string]string{
		"list":             "UteamupPartnerList",
		"get":              "UteamupPartnerGet",
		"applications":     "UteamupPartnerApplicationsList",
		"tenants":          "UteamupPartnerTenantsList",
		"earnings":         "UteamupPartnerEarningsList",
		"program-defaults": "UteamupPartnerProgramDefaultsGet",
		// New actions — 2026-06 partner program overhaul
		"application-get": "UteamupPartnerMyApplicationGet",
		"checklist":       "UteamupPartnerApplicationChecksGet",
		"meetings":        "UteamupPartnerApplicationMeetingsGet",
		"referral-codes":  "UteamupPartnerMyReferralCodesGet",
		"tenant-manager":  "UteamupPartnerMyTenantManagerGet",
	}
	got := map[string]string{}
	for _, a := range d.Actions {
		got[a.Name] = a.ToolName
	}
	for action, tool := range expected {
		if got[action] != tool {
			t.Errorf("expected partner action %q to map to %q, got %q", action, tool, got[action])
		}
	}
}

// The CLI calls REST, and an action without a RESTPath falls back to GET /api/partner: on
// 2026-09-22 `partner get --guid` and `partner tenants --partner-guid` printed the partner list.
// Each action's args are built from its own flags, as runCommand builds them, so a placeholder
// that does not match its flag's camelCase name fails here too.
func TestPartnerActionsRouteToTheirOwnEndpoints(t *testing.T) {
	d := findPartnerDomain()
	if d == nil {
		t.Fatal("expected partner domain to be registered")
	}
	want := map[string]string{
		"list":             "/api/partner",
		"get":              "/api/partner/<guid>",
		"applications":     "/api/partner/application",
		"tenants":          "/api/partner/<partner-guid>/tenant",
		"earnings":         "/api/partner/<partner-guid>/earning",
		"program-defaults": "/api/partner/program-defaults",
		"application-get":  "/api/partner/me/application",
		"checklist":        "/api/partner/application/<application-guid>/checks",
		"meetings":         "/api/partner/application/<application-guid>/meetings",
		"referral-codes":   "/api/partner/me/referral-codes",
		"tenant-manager":   "/api/partner/my-tenant-manager",
	}
	for _, action := range d.Actions {
		expected, ok := want[action.Name]
		if !ok {
			t.Errorf("partner action %q has no expected route here", action.Name)
			continue
		}
		args := map[string]any{}
		for _, flag := range action.Flags {
			args[toCamelCase(flag.Name)] = "<" + flag.Name + ">"
		}
		got, consumed := buildRESTPath(d, action, args)
		if got != expected {
			t.Errorf("%s path = %q, want %q", action.Name, got, expected)
		}
		if len(consumed) != len(action.Flags) {
			t.Errorf("%s consumed %v, want every GUID flag in the path", action.Name, consumed)
		}
	}
	if len(d.Actions) != len(want) {
		t.Errorf("partner has %d actions, want %d", len(d.Actions), len(want))
	}
}

func TestPartnerNewActionsHaveNoSpoofingFlags(t *testing.T) {
	d := findPartnerDomain()
	if d == nil {
		t.Fatal("expected partner domain to be registered")
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

func TestPartnerChecklistFlagPresent(t *testing.T) {
	d := findPartnerDomain()
	if d == nil {
		t.Fatal("expected partner domain to be registered")
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
	t.Error("checklist action not found in partner domain")
}

func TestPartnerMeetingsFlagPresent(t *testing.T) {
	d := findPartnerDomain()
	if d == nil {
		t.Fatal("expected partner domain to be registered")
	}
	for _, a := range d.Actions {
		if a.Name != "meetings" {
			continue
		}
		found := false
		for _, f := range a.Flags {
			if f.Name == "application-guid" && f.Required {
				found = true
				break
			}
		}
		if !found {
			t.Error("meetings action must have a required application-guid flag")
		}
		return
	}
	t.Error("meetings action not found in partner domain")
}
