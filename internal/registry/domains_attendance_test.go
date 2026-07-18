package registry

import "testing"

func TestAttendanceStationDomainWired(t *testing.T) {
	d := findDomain("attendance-station")
	if d == nil {
		t.Fatal("expected attendance-station domain to be registered")
	}
	if d.APIPath != "/api/attendancestation" {
		t.Errorf("attendance-station APIPath = %q, want /api/attendancestation", d.APIPath)
	}
	hasAlias := false
	for _, a := range d.Aliases {
		if a == "attendance" {
			hasAlias = true
		}
	}
	if !hasAlias {
		t.Errorf("attendance-station missing 'attendance' alias, got %v", d.Aliases)
	}
}

func TestAttendanceStationListActionWired(t *testing.T) {
	action := findDomainAction(t, "attendance-station", "list")
	if action.ToolName != "UteamupAttendanceStationList" {
		t.Errorf("list ToolName = %q, want %q", action.ToolName, "UteamupAttendanceStationList")
	}
	f := findFlag(action, "active-only")
	if f == nil {
		t.Fatalf("list must expose an 'active-only' flag, got %+v", action.Flags)
	}
	if f.Type != "bool" || f.Required {
		t.Errorf("'active-only' must be an optional bool flag, got %+v", f)
	}
	if def, ok := f.Default.(bool); !ok || !def {
		t.Errorf("'active-only' default must be the bool literal true, got %#v", f.Default)
	}
}

func TestAttendanceStationGetActionWired(t *testing.T) {
	action := findDomainAction(t, "attendance-station", "get")
	if action.ToolName != "UteamupAttendanceStationGetByGuid" {
		t.Errorf("get ToolName = %q, want %q", action.ToolName, "UteamupAttendanceStationGetByGuid")
	}
	if action.RESTPath != "{stationGuid}" {
		t.Errorf("get RESTPath = %q, want %q", action.RESTPath, "{stationGuid}")
	}
	if len(action.Args) != 1 || action.Args[0].Name != "stationGuid" || !action.Args[0].Required || action.Args[0].Type != "uuid" {
		t.Errorf("get must take a required uuid 'stationGuid' arg, got %+v", action.Args)
	}
}

func TestAttendanceCorrectionsPendingActionWired(t *testing.T) {
	action := findDomainAction(t, "attendance-station", "corrections-pending")
	if action.ToolName != "UteamupAttendanceCorrectionsPending" {
		t.Errorf("corrections-pending ToolName = %q, want %q", action.ToolName, "UteamupAttendanceCorrectionsPending")
	}
	if action.HTTPMethod != "GET" || action.RESTPath != "attendance/corrections/pending" {
		t.Errorf("corrections-pending = %s %q, want GET \"attendance/corrections/pending\"", action.HTTPMethod, action.RESTPath)
	}
	if len(action.Args) != 0 || len(action.Flags) != 0 {
		t.Errorf("corrections-pending should not need args or flags, got args=%+v flags=%+v", action.Args, action.Flags)
	}
}

func TestAttendanceStationRoutesResolve(t *testing.T) {
	d := findDomain("attendance-station")
	if d == nil {
		t.Fatal("expected attendance-station domain to be registered")
	}
	actions := map[string]Action{}
	for _, action := range d.Actions {
		actions[action.Name] = action
	}

	cases := []struct {
		name string
		args map[string]any
		want string
	}{
		{"list", map[string]any{}, "/api/attendancestation"},
		{"get", map[string]any{"stationGuid": "station-1"}, "/api/attendancestation/station-1"},
		{"corrections-pending", map[string]any{}, "/api/attendancestation/attendance/corrections/pending"},
	}

	for _, tc := range cases {
		action, ok := actions[tc.name]
		if !ok {
			t.Fatalf("missing attendance-station action %q", tc.name)
		}
		got, _ := buildRESTPath(d, action, tc.args)
		if got != tc.want {
			t.Errorf("%s path = %q, want %q", tc.name, got, tc.want)
		}
	}
}
