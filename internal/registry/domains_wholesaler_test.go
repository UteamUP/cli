package registry

import "testing"

func findWholesalerDomain() *Domain {
	for _, d := range DefaultRegistry.Domains() {
		if d.Name == "wholesaler" {
			return d
		}
	}
	return nil
}

func TestWholesalerDomainRegistered(t *testing.T) {
	d := findWholesalerDomain()
	if d == nil {
		t.Fatal("expected wholesaler domain to be registered")
	}
	if d.Description == "" {
		t.Error("wholesaler domain must have a Description")
	}
}

func TestWholesalerActionsWired(t *testing.T) {
	d := findWholesalerDomain()
	if d == nil {
		t.Fatal("expected wholesaler domain to be registered")
	}
	expected := map[string]string{
		"list":         "UteamupWholesalerList",
		"get":          "UteamupWholesalerGet",
		"applications": "UteamupWholesalerApplicationsList",
		"catalog":      "UteamupWholesalerCatalogGet",
		"me":           "UteamupWholesalerMyGet",
	}
	actions := map[string]Action{}
	for _, a := range d.Actions {
		actions[a.Name] = a
	}
	for name, tool := range expected {
		a, ok := actions[name]
		if !ok {
			t.Errorf("missing wholesaler action %q", name)
			continue
		}
		if a.ToolName != tool {
			t.Errorf("action %q maps to %q, want %q", name, a.ToolName, tool)
		}
	}

	guidActions := []string{"get", "catalog"}
	for _, name := range guidActions {
		a := actions[name]
		hasGuid := false
		for _, f := range a.Flags {
			if f.Name == "guid" && f.Required && f.Type == "string" {
				hasGuid = true
			}
		}
		if !hasGuid {
			t.Errorf("action %q must take a required string --guid flag", name)
		}
	}
}
