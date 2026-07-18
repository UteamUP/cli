package registry

import "testing"

// The timesheet read tools live on the time-entry domain (aliased `timesheet`)
// but are served by TimesheetController via a RESTBasePath override.

func TestTimesheetAliasStillOnTimeEntryDomain(t *testing.T) {
	d := findDomain("time-entry")
	if d == nil {
		t.Fatal("expected time-entry domain to be registered")
	}
	hasAlias := false
	for _, a := range d.Aliases {
		if a == "timesheet" {
			hasAlias = true
		}
	}
	if !hasAlias {
		t.Errorf("time-entry domain missing 'timesheet' alias, got %v", d.Aliases)
	}
}

func TestTimesheetWeeklyMineActionWired(t *testing.T) {
	action := findDomainAction(t, "time-entry", "weekly-mine")
	if action.ToolName != "UteamupTimesheetWeeklyMine" {
		t.Errorf("weekly-mine ToolName = %q, want %q", action.ToolName, "UteamupTimesheetWeeklyMine")
	}
	if action.HTTPMethod != "GET" || action.RESTBasePath != "/api/timesheet" || action.RESTPath != "weekly/me" {
		t.Errorf("weekly-mine = %s %q + %q, want GET /api/timesheet + weekly/me", action.HTTPMethod, action.RESTBasePath, action.RESTPath)
	}
	f := findFlag(action, "week-start")
	if f == nil || !f.Required || f.Type != "string" || f.BodyName != "weekStart" {
		t.Errorf("weekly-mine must have a required string 'week-start' flag mapping to weekStart, got %+v", f)
	}
}

func TestTimesheetPendingApprovalsActionWired(t *testing.T) {
	action := findDomainAction(t, "time-entry", "pending-approvals")
	if action.ToolName != "UteamupTimesheetPendingApprovals" {
		t.Errorf("pending-approvals ToolName = %q, want %q", action.ToolName, "UteamupTimesheetPendingApprovals")
	}
	if action.HTTPMethod != "GET" || action.RESTBasePath != "/api/timesheet" || action.RESTPath != "pending-approval" {
		t.Errorf("pending-approvals = %s %q + %q, want GET /api/timesheet + pending-approval", action.HTTPMethod, action.RESTBasePath, action.RESTPath)
	}
	if len(action.Args) != 0 || len(action.Flags) != 0 {
		t.Errorf("pending-approvals should not need args or flags, got args=%+v flags=%+v", action.Args, action.Flags)
	}
}

func TestTimesheetRoutesResolve(t *testing.T) {
	d := findDomain("time-entry")
	if d == nil {
		t.Fatal("expected time-entry domain to be registered")
	}
	actions := map[string]Action{}
	for _, action := range d.Actions {
		actions[action.Name] = action
	}

	cases := []struct {
		name string
		want string
	}{
		{"weekly-mine", "/api/timesheet/weekly/me"},
		{"pending-approvals", "/api/timesheet/pending-approval"},
	}

	for _, tc := range cases {
		action, ok := actions[tc.name]
		if !ok {
			t.Fatalf("missing time-entry action %q", tc.name)
		}
		got, _ := buildRESTPath(d, action, map[string]any{})
		if got != tc.want {
			t.Errorf("%s path = %q, want %q", tc.name, got, tc.want)
		}
	}
}
