package registry

import "testing"

func findOnCallDomain(t *testing.T) *Domain {
	t.Helper()
	for _, d := range DefaultRegistry.Domains() {
		if d.Name == "oncall" {
			return d
		}
	}
	t.Fatal("expected oncall domain to be registered")
	return nil
}

func TestOnCallDomainRegistered(t *testing.T) {
	d := findOnCallDomain(t)
	if d.Description == "" {
		t.Error("oncall domain must have a Description")
	}
	if d.APIPath != "/api/oncall" {
		t.Errorf("oncall APIPath = %q, want %q", d.APIPath, "/api/oncall")
	}
	hasAlias := false
	for _, a := range d.Aliases {
		if a == "on-call" {
			hasAlias = true
		}
	}
	if !hasAlias {
		t.Errorf("oncall domain missing 'on-call' alias, got %v", d.Aliases)
	}
}

func TestOnCallWhoActionWired(t *testing.T) {
	d := findOnCallDomain(t)
	var who *Action
	for i := range d.Actions {
		if d.Actions[i].Name == "who" {
			who = &d.Actions[i]
		}
	}
	if who == nil {
		t.Fatal("expected 'who' action on oncall domain")
	}
	if who.ToolName != "UteamupOnCallWho" {
		t.Errorf("who ToolName = %q, want %q", who.ToolName, "UteamupOnCallWho")
	}
	if who.HTTPMethod != "GET" {
		t.Errorf("who HTTPMethod = %q, want GET", who.HTTPMethod)
	}
	if who.RESTPath != "{schedule-guid}/who" {
		t.Errorf("who RESTPath = %q, want %q", who.RESTPath, "{schedule-guid}/who")
	}
	// schedule-guid is a required uuid positional arg
	if len(who.Args) != 1 || who.Args[0].Name != "schedule-guid" || !who.Args[0].Required || who.Args[0].Type != "uuid" {
		t.Errorf("who must take a required uuid 'schedule-guid' arg, got %+v", who.Args)
	}
	// optional 'at' query flag
	hasAt := false
	for _, f := range who.Flags {
		if f.Name == "at" {
			hasAt = true
		}
	}
	if !hasAt {
		t.Errorf("who must expose an optional 'at' flag, got %+v", who.Flags)
	}
}

func TestOnCallScheduleActionsWired(t *testing.T) {
	d := findOnCallDomain(t)
	byName := map[string]*Action{}
	for i := range d.Actions {
		byName[d.Actions[i].Name] = &d.Actions[i]
	}

	list, ok := byName["schedule-list"]
	if !ok {
		t.Fatal("expected 'schedule-list' action")
	}
	if list.HTTPMethod != "GET" || list.RESTPath != "schedules" {
		t.Errorf("schedule-list = %s %q, want GET \"schedules\"", list.HTTPMethod, list.RESTPath)
	}

	create, ok := byName["schedule-create"]
	if !ok {
		t.Fatal("expected 'schedule-create' action")
	}
	if create.HTTPMethod != "POST" || create.RESTPath != "schedules" {
		t.Errorf("schedule-create = %s %q, want POST \"schedules\"", create.HTTPMethod, create.RESTPath)
	}
	var nameFlag *FlagDef
	for i := range create.Flags {
		if create.Flags[i].Name == "name" {
			nameFlag = &create.Flags[i]
		}
	}
	if nameFlag == nil || !nameFlag.Required {
		t.Errorf("schedule-create must have a required 'name' flag, got %+v", create.Flags)
	}
}

func TestOnCallLayerAddActionWired(t *testing.T) {
	d := findOnCallDomain(t)
	var la *Action
	for i := range d.Actions {
		if d.Actions[i].Name == "layer-add" {
			la = &d.Actions[i]
		}
	}
	if la == nil {
		t.Fatal("expected 'layer-add' action")
	}
	if la.HTTPMethod != "POST" || la.RESTPath != "{schedule-guid}/layers" {
		t.Errorf("layer-add = %s %q, want POST \"{schedule-guid}/layers\"", la.HTTPMethod, la.RESTPath)
	}
	if len(la.Args) != 1 || la.Args[0].Name != "schedule-guid" || la.Args[0].Type != "uuid" {
		t.Errorf("layer-add must take a uuid 'schedule-guid' arg, got %+v", la.Args)
	}
	byFlag := map[string]*FlagDef{}
	for i := range la.Flags {
		byFlag[la.Flags[i].Name] = &la.Flags[i]
	}
	// user flag: repeatable, required, maps to orderedUserGuids body field
	user, ok := byFlag["user"]
	if !ok || user.Type != "stringSlice" || !user.Required || user.BodyName != "orderedUserGuids" {
		t.Errorf("layer-add 'user' flag must be a required stringSlice → orderedUserGuids, got %+v", user)
	}
	if sm, ok := byFlag["shift-minutes"]; !ok || sm.BodyName != "shiftLengthMinutes" || !sm.Required {
		t.Errorf("layer-add 'shift-minutes' must map to shiftLengthMinutes and be required, got %+v", sm)
	}
	// precedence default must be an int literal (not a runtime-panicking float mismatch)
	if p, ok := byFlag["precedence"]; !ok || p.Default != 1 {
		t.Errorf("layer-add 'precedence' default should be 1, got %+v", p)
	}
}
