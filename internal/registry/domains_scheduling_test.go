package registry

import "testing"

func TestShiftCrudIsGuidFirst(t *testing.T) {
	d := findDomain("shift")
	if d == nil {
		t.Fatal("expected shift domain to be registered")
	}

	actions := map[string]Action{}
	for _, action := range d.Actions {
		actions[action.Name] = action
	}

	for _, name := range []string{"get", "update", "delete"} {
		action, ok := actions[name]
		if !ok {
			t.Fatalf("missing shift action %q", name)
		}
		if len(action.Args) != 1 || action.Args[0].Name != "externalGuid" {
			t.Fatalf("shift %s args = %+v, want single externalGuid arg", name, action.Args)
		}
		if action.Args[0].Type != "string" {
			t.Fatalf("shift %s externalGuid type = %q, want string", name, action.Args[0].Type)
		}
		if action.RESTPath != "by-guid/{externalGuid}" {
			t.Fatalf("shift %s RESTPath = %q, want by-guid/{externalGuid}", name, action.RESTPath)
		}
	}
}

func TestShiftGuidRoutesResolve(t *testing.T) {
	d := findDomain("shift")
	if d == nil {
		t.Fatal("expected shift domain to be registered")
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
		{"get", map[string]any{"externalGuid": "shift-1"}, "/api/shift/by-guid/shift-1"},
		{"update", map[string]any{"externalGuid": "shift-1"}, "/api/shift/by-guid/shift-1"},
		{"delete", map[string]any{"externalGuid": "shift-1"}, "/api/shift/by-guid/shift-1"},
		{"pattern-list", map[string]any{"shiftGuid": "shift-1"}, "/api/shift/by-guid/shift-1/patterns"},
		{"pattern-get", map[string]any{"patternGuid": "pattern-1"}, "/api/shift/patterns/by-guid/pattern-1"},
		{"pattern-update", map[string]any{"patternGuid": "pattern-1"}, "/api/shift/patterns/by-guid/pattern-1"},
		{"pattern-delete", map[string]any{"patternGuid": "pattern-1"}, "/api/shift/patterns/by-guid/pattern-1"},
	}

	for _, tc := range cases {
		action, ok := actions[tc.name]
		if !ok {
			t.Fatalf("missing shift action %q", tc.name)
		}
		got, consumed := buildRESTPath(d, action, tc.args)
		if got != tc.want {
			t.Fatalf("%s path = %q, want %q", tc.name, got, tc.want)
		}
		if len(consumed) != 1 {
			t.Fatalf("%s consumed = %v, want exactly one path arg", tc.name, consumed)
		}
	}
}

func TestShiftInstanceGuidRoutesResolve(t *testing.T) {
	d := findDomain("shift-instance")
	if d == nil {
		t.Fatal("expected shift-instance domain to be registered")
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
		{"get", map[string]any{"instanceGuid": "instance-1"}, "/api/shiftinstance/by-guid/instance-1"},
		{"update", map[string]any{"instanceGuid": "instance-1"}, "/api/shiftinstance/by-guid/instance-1"},
		{"delete", map[string]any{"instanceGuid": "instance-1"}, "/api/shiftinstance/by-guid/instance-1"},
		{"approve", map[string]any{"instanceGuid": "instance-1"}, "/api/shiftinstance/by-guid/instance-1/approve"},
		{"status", map[string]any{"instanceGuid": "instance-1"}, "/api/shiftinstance/by-guid/instance-1/status"},
	}

	for _, tc := range cases {
		action, ok := actions[tc.name]
		if !ok {
			t.Fatalf("missing shift-instance action %q", tc.name)
		}
		got, consumed := buildRESTPath(d, action, tc.args)
		if got != tc.want {
			t.Fatalf("%s path = %q, want %q", tc.name, got, tc.want)
		}
		if len(consumed) != 1 {
			t.Fatalf("%s consumed = %v, want one path arg", tc.name, consumed)
		}
	}
}

func TestShiftRequestGuidRoutesResolve(t *testing.T) {
	d := findDomain("shift-request")
	if d == nil {
		t.Fatal("expected shift-request domain to be registered")
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
		{"get", map[string]any{"requestGuid": "request-1"}, "/api/shiftrequest/by-guid/request-1"},
		{"approve", map[string]any{"requestGuid": "request-1"}, "/api/shiftrequest/by-guid/request-1/approve"},
		{"deny", map[string]any{"requestGuid": "request-1"}, "/api/shiftrequest/by-guid/request-1/deny"},
		{"withdraw", map[string]any{"requestGuid": "request-1"}, "/api/shiftrequest/by-guid/request-1"},
	}

	for _, tc := range cases {
		action, ok := actions[tc.name]
		if !ok {
			t.Fatalf("missing shift-request action %q", tc.name)
		}
		got, consumed := buildRESTPath(d, action, tc.args)
		if got != tc.want {
			t.Fatalf("%s path = %q, want %q", tc.name, got, tc.want)
		}
		if len(consumed) != 1 {
			t.Fatalf("%s consumed = %v, want one path arg", tc.name, consumed)
		}
	}
}

func TestShiftAssignmentGuidRoutesResolve(t *testing.T) {
	d := findDomain("shift-assignment")
	if d == nil {
		t.Fatal("expected shift-assignment domain to be registered")
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
		{"instance", map[string]any{"instanceGuid": "instance-1"}, "/api/shiftuserassignment/instance/by-guid/instance-1"},
		{"update", map[string]any{"assignmentGuid": "assignment-1"}, "/api/shiftuserassignment/by-guid/assignment-1"},
		{"delete", map[string]any{"assignmentGuid": "assignment-1"}, "/api/shiftuserassignment/by-guid/assignment-1"},
		{"unavailable", map[string]any{"assignmentGuid": "assignment-1"}, "/api/shiftuserassignment/by-guid/assignment-1/unavailable"},
	}

	for _, tc := range cases {
		action, ok := actions[tc.name]
		if !ok {
			t.Fatalf("missing shift-assignment action %q", tc.name)
		}
		got, consumed := buildRESTPath(d, action, tc.args)
		if got != tc.want {
			t.Fatalf("%s path = %q, want %q", tc.name, got, tc.want)
		}
		if len(consumed) != 1 {
			t.Fatalf("%s consumed = %v, want one path arg", tc.name, consumed)
		}
	}
}

func TestShiftHandoverPreviousUsesShiftGuid(t *testing.T) {
	action := findDomainAction(t, "shift-handover", "previous")
	if action.ToolName != "UteamupShiftHandoverGetPrevious" {
		t.Fatalf("previous ToolName = %q, want UteamupShiftHandoverGetPrevious", action.ToolName)
	}
	if len(action.Args) != 1 || action.Args[0].Name != "shiftGuid" {
		t.Fatalf("previous args = %+v, want single shiftGuid arg", action.Args)
	}

	d := findDomain("shift-handover")
	if d == nil {
		t.Fatal("expected shift-handover domain to be registered")
	}
	got, consumed := buildRESTPath(d, *action, map[string]any{"shiftGuid": "shift-1"})
	if got != "/api/shifthandover/previous/by-guid/shift-1" {
		t.Fatalf("previous path = %q, want /api/shifthandover/previous/by-guid/shift-1", got)
	}
	if len(consumed) != 1 {
		t.Fatalf("previous consumed = %v, want one path arg", consumed)
	}
}
