package registry

import (
	"testing"
)

// findMeterScheduleDomain locates the registered meter-schedule domain.
// Centralised here so every test gets the same lookup behaviour.
func findMeterScheduleDomain(t *testing.T) *Domain {
	t.Helper()
	for _, dom := range DefaultRegistry.Domains() {
		if dom.Name == "meter-schedule" {
			return dom
		}
	}
	t.Fatal("expected meter-schedule domain to be registered")
	return nil
}

func TestMeterScheduleDomainRegistered(t *testing.T) {
	d := findMeterScheduleDomain(t)
	if d.APIPath != "/api/meter-reading-schedules" {
		t.Errorf("APIPath = %q, want %q", d.APIPath, "/api/meter-reading-schedules")
	}

	// The Guid-first aliases must exist so users can type either the long
	// or the short name. `meter-reading-schedule` matches the backend
	// controller route prefix.
	wantAliases := map[string]bool{
		"meter-reading-schedule":  false,
		"meter-schedules":         false,
		"ms":                      false,
	}
	for _, a := range d.Aliases {
		if _, ok := wantAliases[a]; ok {
			wantAliases[a] = true
		}
	}
	for alias, found := range wantAliases {
		if !found {
			t.Errorf("expected alias %q on meter-schedule domain", alias)
		}
	}
}

// All schedule-targeting actions must take a Guid (string) positional arg,
// never an int. Catches accidental int-id regressions per the GUIDs-In rule.
func TestMeterScheduleActionsAreGuidKeyed(t *testing.T) {
	d := findMeterScheduleDomain(t)

	type actionExpect struct {
		name    string
		argName string
	}
	cases := []actionExpect{
		{"get", "guid"},
		{"update", "guid"},
		{"delete", "guid"},
		{"compliance-asset", "asset-guid"},
		{"initialize", "asset-guid"},
	}

	for _, c := range cases {
		var found *Action
		for i := range d.Actions {
			if d.Actions[i].Name == c.name {
				found = &d.Actions[i]
				break
			}
		}
		if found == nil {
			t.Errorf("expected action %q on meter-schedule domain", c.name)
			continue
		}
		if len(found.Args) == 0 {
			t.Errorf("action %q expected at least one positional arg", c.name)
			continue
		}
		arg := found.Args[0]
		if arg.Name != c.argName {
			t.Errorf("action %q: first arg = %q, want %q (Guid-first rule)", c.name, arg.Name, c.argName)
		}
		if arg.Type != "string" {
			t.Errorf("action %q: first arg type = %q, want %q (Guids are strings, never int)", c.name, arg.Type, "string")
		}
		if !arg.Required {
			t.Errorf("action %q: first arg %q must be Required", c.name, arg.Name)
		}
	}
}

// `create` declares the asset + attribute via Guid flags. Catches a
// regression where someone reverts to the int variants.
func TestMeterScheduleCreateUsesGuidFlags(t *testing.T) {
	d := findMeterScheduleDomain(t)
	var create *Action
	for i := range d.Actions {
		if d.Actions[i].Name == "create" {
			create = &d.Actions[i]
			break
		}
	}
	if create == nil {
		t.Fatal("expected `create` action on meter-schedule domain")
	}
	if create.ToolName != "UteamupMeterscheduleCreateByGuid" {
		t.Errorf("create ToolName = %q, want %q", create.ToolName, "UteamupMeterscheduleCreateByGuid")
	}

	expected := map[string]string{
		"asset-guid":                "string",
		"attribute-definition-guid": "string",
		"interval-seconds":          "int",
	}
	got := make(map[string]string)
	for _, f := range create.Flags {
		got[f.Name] = f.Type
	}
	for name, ty := range expected {
		gotType, ok := got[name]
		if !ok {
			t.Errorf("create flag %q missing", name)
			continue
		}
		if gotType != ty {
			t.Errorf("create flag %q type = %q, want %q", name, gotType, ty)
		}
	}

	// Explicit regression guard: must NOT carry the legacy int flags.
	if _, present := got["asset-id"]; present {
		t.Error("create must not expose legacy int flag `asset-id` — Guid-first only")
	}
	if _, present := got["attribute-definition-id"]; present {
		t.Error("create must not expose legacy int flag `attribute-definition-id` — Guid-first only")
	}
}

// Verifies the compliance-asset action drives the /compliance/asset/{assetGuid}
// path, not the legacy /compliance/{assetId} route.
func TestMeterScheduleComplianceAssetUsesGuidPath(t *testing.T) {
	d := findMeterScheduleDomain(t)
	var action *Action
	for i := range d.Actions {
		if d.Actions[i].Name == "compliance-asset" {
			action = &d.Actions[i]
			break
		}
	}
	if action == nil {
		t.Fatal("expected `compliance-asset` action on meter-schedule domain")
	}
	if action.RESTPath != "compliance/asset/{assetGuid}" {
		t.Errorf("compliance-asset RESTPath = %q, want %q", action.RESTPath, "compliance/asset/{assetGuid}")
	}
	if action.HTTPMethod != "GET" {
		t.Errorf("compliance-asset HTTPMethod = %q, want %q", action.HTTPMethod, "GET")
	}
}
