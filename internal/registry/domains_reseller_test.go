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
