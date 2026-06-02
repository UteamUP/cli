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
		{"open-workorders", "asset-guid"},
		{"initialize", "asset-guid"},
		{"create-workorder", "guid"},
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

// Verifies the create-workorder action POSTs to {guid}/create-workorder, mirrors
// the MCP UteamupMeterscheduleCreateWorkorder tool, and exposes the template flags.
func TestMeterScheduleCreateWorkorderAction(t *testing.T) {
	d := findMeterScheduleDomain(t)
	var action *Action
	for i := range d.Actions {
		if d.Actions[i].Name == "create-workorder" {
			action = &d.Actions[i]
			break
		}
	}
	if action == nil {
		t.Fatal("expected `create-workorder` action on meter-schedule domain")
	}
	if action.ToolName != "UteamupMeterscheduleCreateWorkorder" {
		t.Errorf("create-workorder ToolName = %q, want %q", action.ToolName, "UteamupMeterscheduleCreateWorkorder")
	}
	if action.HTTPMethod != "POST" {
		t.Errorf("create-workorder HTTPMethod = %q, want POST", action.HTTPMethod)
	}
	if action.RESTPath != "{guid}/create-workorder" {
		t.Errorf("create-workorder RESTPath = %q, want %q", action.RESTPath, "{guid}/create-workorder")
	}
	flags := make(map[string]string)
	for _, f := range action.Flags {
		flags[f.Name] = f.Type
	}
	if flags["workorder-template-guid"] != "string" {
		t.Error("create-workorder must expose a string `workorder-template-guid` flag")
	}
	if flags["use-schedule-template"] != "bool" {
		t.Error("create-workorder must expose a bool `use-schedule-template` flag")
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

// Verifies the open-workorders action drives the GUID-keyed
// /asset/{assetGuid}/open-workorders path and mirrors the MCP tool that lists
// every open meter-reading workorder for an asset.
func TestMeterScheduleOpenWorkordersAction(t *testing.T) {
	d := findMeterScheduleDomain(t)
	var action *Action
	for i := range d.Actions {
		if d.Actions[i].Name == "open-workorders" {
			action = &d.Actions[i]
			break
		}
	}
	if action == nil {
		t.Fatal("expected `open-workorders` action on meter-schedule domain")
	}
	if action.ToolName != "UteamupMeterscheduleGetOpenWorkorders" {
		t.Errorf("open-workorders ToolName = %q, want %q", action.ToolName, "UteamupMeterscheduleGetOpenWorkorders")
	}
	if action.HTTPMethod != "GET" {
		t.Errorf("open-workorders HTTPMethod = %q, want GET", action.HTTPMethod)
	}
	if action.RESTPath != "asset/{assetGuid}/open-workorders" {
		t.Errorf("open-workorders RESTPath = %q, want %q", action.RESTPath, "asset/{assetGuid}/open-workorders")
	}
}

func findMeterScheduleAction(t *testing.T, name string) *Action {
	t.Helper()
	d := findMeterScheduleDomain(t)
	for i := range d.Actions {
		if d.Actions[i].Name == name {
			return &d.Actions[i]
		}
	}
	t.Fatalf("expected %q action on meter-schedule domain", name)
	return nil
}

// Both create and update must expose the calendar-recurrence flags with the right
// types so a weekly/monthly/yearly schedule can be configured from the CLI.
func TestMeterScheduleRecurrenceFlags(t *testing.T) {
	expected := map[string]string{
		"recurrence-type":   "string",
		"days-of-week":      "stringSlice",
		"day-of-month-mode": "string",
		"day-of-month":      "int",
		"month-of-year":     "int",
	}

	for _, actionName := range []string{"create", "update"} {
		action := findMeterScheduleAction(t, actionName)
		got := make(map[string]FlagDef)
		for _, f := range action.Flags {
			got[f.Name] = f
		}
		for name, ty := range expected {
			flag, ok := got[name]
			if !ok {
				t.Errorf("%s: recurrence flag %q missing", actionName, name)
				continue
			}
			if flag.Type != ty {
				t.Errorf("%s: flag %q type = %q, want %q", actionName, name, flag.Type, ty)
			}
		}
		// daysOfWeek must serialize to the camelCase body field the backend DTO expects.
		if got["days-of-week"].BodyName != "daysOfWeek" {
			t.Errorf("%s: days-of-week BodyName = %q, want %q", actionName, got["days-of-week"].BodyName, "daysOfWeek")
		}
	}
}

// interval-seconds must NOT be Required anymore — calendar schedules don't carry one.
func TestMeterScheduleCreateIntervalOptional(t *testing.T) {
	create := findMeterScheduleAction(t, "create")
	for _, f := range create.Flags {
		if f.Name == "interval-seconds" && f.Required {
			t.Error("create flag `interval-seconds` must not be Required (calendar recurrence omits it)")
		}
	}
}

// Verifies the record-workorder action mirrors the GUID-migrated MCP tool
// UteamupMeterscheduleRecordWorkorder and is keyed by the workorder Guid, never an int.
func TestMeterScheduleRecordWorkorderAction(t *testing.T) {
	action := findMeterScheduleAction(t, "record-workorder")
	if action.ToolName != "UteamupMeterscheduleRecordWorkorder" {
		t.Errorf("record-workorder ToolName = %q, want %q", action.ToolName, "UteamupMeterscheduleRecordWorkorder")
	}
	flags := make(map[string]FlagDef)
	for _, f := range action.Flags {
		flags[f.Name] = f
	}
	if wg, ok := flags["workorder-guid"]; !ok || wg.Type != "string" || !wg.Required {
		t.Error("record-workorder must expose a required string `workorder-guid` flag (Guid-first)")
	}
	if ad, ok := flags["attribute-definition-id"]; !ok || ad.Type != "int" {
		t.Error("record-workorder must expose an int `attribute-definition-id` flag")
	}
	if rv, ok := flags["reading-value"]; !ok || rv.Type != "float" {
		t.Error("record-workorder must expose a float `reading-value` flag")
	}
	// Regression guard: must NOT carry a legacy int workorder identifier.
	if _, present := flags["workorder-id"]; present {
		t.Error("record-workorder must not expose a legacy int `workorder-id` flag — Guid-first only")
	}
}
