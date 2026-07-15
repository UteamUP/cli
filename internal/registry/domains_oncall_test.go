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

func TestOnCallCalloutSummaryActionWired(t *testing.T) {
	d := findOnCallDomain(t)
	var summary *Action
	for i := range d.Actions {
		if d.Actions[i].Name == "callout-summary" {
			summary = &d.Actions[i]
		}
	}
	if summary == nil {
		t.Fatal("expected 'callout-summary' action")
	}
	if summary.ToolName != "UteamupOnCallCalloutSummary" {
		t.Errorf("callout-summary ToolName = %q, want %q", summary.ToolName, "UteamupOnCallCalloutSummary")
	}
	if summary.HTTPMethod != "GET" || summary.RESTPath != "callouts/summary" {
		t.Errorf("callout-summary = %s %q, want GET \"callouts/summary\"", summary.HTTPMethod, summary.RESTPath)
	}
	if len(summary.Args) != 0 || len(summary.Flags) != 0 {
		t.Errorf("callout-summary should not need args or flags, got args=%+v flags=%+v", summary.Args, summary.Flags)
	}
}

func TestOnCallCalendarActionWired(t *testing.T) {
	d := findOnCallDomain(t)
	var calendar *Action
	for i := range d.Actions {
		if d.Actions[i].Name == "calendar" {
			calendar = &d.Actions[i]
		}
	}
	if calendar == nil {
		t.Fatal("expected 'calendar' action")
	}
	if calendar.ToolName != "UteamupOnCallCalendar" {
		t.Errorf("calendar ToolName = %q, want %q", calendar.ToolName, "UteamupOnCallCalendar")
	}
	if calendar.HTTPMethod != "GET" || calendar.RESTPath != "{schedule-guid}/calendar.ics" {
		t.Errorf("calendar = %s %q, want GET \"{schedule-guid}/calendar.ics\"", calendar.HTTPMethod, calendar.RESTPath)
	}
	if len(calendar.Args) != 1 || calendar.Args[0].Name != "schedule-guid" || !calendar.Args[0].Required || calendar.Args[0].Type != "uuid" {
		t.Errorf("calendar must take a required uuid 'schedule-guid' arg, got %+v", calendar.Args)
	}
	byFlag := map[string]*FlagDef{}
	for i := range calendar.Flags {
		byFlag[calendar.Flags[i].Name] = &calendar.Flags[i]
	}
	if _, ok := byFlag["from"]; !ok {
		t.Errorf("calendar missing 'from' flag, got %+v", calendar.Flags)
	}
	if _, ok := byFlag["to"]; !ok {
		t.Errorf("calendar missing 'to' flag, got %+v", calendar.Flags)
	}
}

func TestOnCallCalendarSubscriptionActionsWired(t *testing.T) {
	d := findOnCallDomain(t)
	byName := map[string]*Action{}
	for i := range d.Actions {
		byName[d.Actions[i].Name] = &d.Actions[i]
	}

	cases := []struct {
		name   string
		tool   string
		method string
	}{
		{"calendar-subscription-get", "UteamupOnCallCalendarSubscriptionGet", "GET"},
		{"calendar-subscription-rotate", "UteamupOnCallCalendarSubscriptionRotate", "POST"},
		{"calendar-subscription-revoke", "UteamupOnCallCalendarSubscriptionRevoke", "DELETE"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			action, ok := byName[tc.name]
			if !ok {
				t.Fatalf("expected %q action", tc.name)
			}
			if action.ToolName != tc.tool {
				t.Errorf("%s ToolName = %q, want %q", tc.name, action.ToolName, tc.tool)
			}
			if action.HTTPMethod != tc.method || action.RESTPath != "{schedule-guid}/calendar-subscription" {
				t.Errorf("%s = %s %q, want %s \"{schedule-guid}/calendar-subscription\"", tc.name, action.HTTPMethod, action.RESTPath, tc.method)
			}
			if len(action.Args) != 1 || action.Args[0].Name != "schedule-guid" || !action.Args[0].Required || action.Args[0].Type != "uuid" {
				t.Errorf("%s must take a required uuid 'schedule-guid' arg, got %+v", tc.name, action.Args)
			}
			if len(action.Flags) != 0 {
				t.Errorf("%s should not take flags, got %+v", tc.name, action.Flags)
			}
		})
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

func TestOnCallOverrideAddActionWired(t *testing.T) {
	d := findOnCallDomain(t)
	var oa *Action
	for i := range d.Actions {
		if d.Actions[i].Name == "override-add" {
			oa = &d.Actions[i]
		}
	}
	if oa == nil {
		t.Fatal("expected 'override-add' action")
	}
	if oa.HTTPMethod != "POST" || oa.RESTPath != "{schedule-guid}/overrides" {
		t.Errorf("override-add = %s %q, want POST \"{schedule-guid}/overrides\"", oa.HTTPMethod, oa.RESTPath)
	}
	byFlag := map[string]*FlagDef{}
	for i := range oa.Flags {
		byFlag[oa.Flags[i].Name] = &oa.Flags[i]
	}
	if u, ok := byFlag["user"]; !ok || u.BodyName != "targetUserGuid" || !u.Required {
		t.Errorf("override-add 'user' must be required → targetUserGuid, got %+v", u)
	}
	if s, ok := byFlag["start"]; !ok || s.BodyName != "startAt" {
		t.Errorf("override-add 'start' must map to startAt, got %+v", s)
	}
	if e, ok := byFlag["end"]; !ok || e.BodyName != "endAt" {
		t.Errorf("override-add 'end' must map to endAt, got %+v", e)
	}
}

func TestOnCallClassifyStandbyActionWired(t *testing.T) {
	d := findOnCallDomain(t)
	var cs *Action
	for i := range d.Actions {
		if d.Actions[i].Name == "classify-standby" {
			cs = &d.Actions[i]
		}
	}
	if cs == nil {
		t.Fatal("expected 'classify-standby' action")
	}
	if cs.HTTPMethod != "POST" || cs.RESTPath != "classify-standby" {
		t.Errorf("classify-standby = %s %q, want POST \"classify-standby\"", cs.HTTPMethod, cs.RESTPath)
	}
	byFlag := map[string]*FlagDef{}
	for i := range cs.Flags {
		byFlag[cs.Flags[i].Name] = &cs.Flags[i]
	}
	// float flag default must be a float literal (registry type-assert panic guard)
	if c, ok := byFlag["callouts-per-week"]; !ok {
		t.Error("classify-standby missing 'callouts-per-week'")
	} else if _, isFloat := c.Default.(float64); !isFloat {
		t.Errorf("'callouts-per-week' default must be a float literal, got %T", c.Default)
	}
	if r, ok := byFlag["response-minutes"]; !ok || r.BodyName != "responseTimeMinutes" {
		t.Errorf("'response-minutes' must map to responseTimeMinutes, got %+v", r)
	}
}
