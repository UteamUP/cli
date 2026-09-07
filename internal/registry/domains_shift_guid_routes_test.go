package registry

import "testing"

func TestShiftTemplateDomainIsRetired(t *testing.T) {
	if d := findDomain("shift-template"); d != nil {
		t.Fatalf("shift-template domain should stay retired; found %+v", d)
	}
}

func TestShiftCrudIsGuidFirst(t *testing.T) {
	d := findDomain("shift")
	if d == nil {
		t.Fatal("expected shift domain to be registered")
	}

	actions := map[string]Action{}
	for _, action := range d.Actions {
		actions[action.Name] = action
	}

	for _, name := range []string{"update", "delete"} {
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

	get, ok := actions["get"]
	if !ok {
		t.Fatal("missing shift action \"get\"")
	}
	if len(get.Args) != 1 || get.Args[0].Name != "shiftGuid" || get.Args[0].Type != "uuid" {
		t.Fatalf("shift get args = %+v, want single shiftGuid UUID arg", get.Args)
	}
	if get.RESTPath != "by-guid/{shiftGuid}" {
		t.Fatalf("shift get RESTPath = %q, want by-guid/{shiftGuid}", get.RESTPath)
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
		{"get", map[string]any{"shiftGuid": "shift-1"}, "/api/shift/by-guid/shift-1"},
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

func TestShiftAssignmentUserDateUsesPublicGuid(t *testing.T) {
	action := findDomainAction(t, "shift-assignment", "user-date")
	if len(action.Args) != 2 || action.Args[0].Name != "userGuid" || action.Args[0].Type != "uuid" {
		t.Fatalf("user-date must accept a public worker GUID: %+v", action.Args)
	}
	d := findDomain("shift-assignment")
	guid := "8c7ae65f-73a8-4af1-a5e2-d0145348257b"
	got, consumed := buildRESTPath(d, *action, map[string]any{"userGuid": guid, "date": "2026-09-07"})
	want := "/api/shiftuserassignment/user/by-guid/" + guid + "/date/2026-09-07"
	if got != want || len(consumed) != 2 {
		t.Fatalf("user-date route = %q (%v), want %q and two path arguments", got, consumed, want)
	}
}

func TestWorkerOwnShiftCommandsUseOwnRoutes(t *testing.T) {
	cases := []struct {
		domain, action, tool, method, path string
		args                               map[string]any
	}{
		{"shift-assignment", "mine", "UteamupShiftUserAssignmentMine", "GET", "/api/shiftuserassignment/me/range", nil},
		{"shift-assignment", "confirm-mine", "UteamupShiftUserAssignmentConfirmMine", "PUT", "/api/shiftuserassignment/me/by-guid/assignment/confirm", map[string]any{"assignmentGuid": "assignment"}},
		{"shift-assignment", "decline-mine", "UteamupShiftUserAssignmentDeclineMine", "PUT", "/api/shiftuserassignment/me/by-guid/assignment/decline", map[string]any{"assignmentGuid": "assignment"}},
		{"shift-assignment", "self-assign", "UteamupShiftUserAssignmentSelfAssign", "POST", "/api/shiftuserassignment/me/self-assign", nil},
		{"shift-request", "mine", "UteamupShiftRequestMine", "GET", "/api/shiftrequest/mine", nil},
		{"shift-request", "incoming", "UteamupShiftRequestIncoming", "GET", "/api/shiftrequest/mine/counterpart", nil},
		{"shift-request", "respond", "UteamupShiftRequestRespondCounterpart", "PUT", "/api/shiftrequest/mine/counterpart/by-guid/request/respond", map[string]any{"requestGuid": "request"}},
		{"shift-request", "withdraw-mine", "UteamupShiftRequestWithdrawMine", "DELETE", "/api/shiftrequest/mine/by-guid/request", map[string]any{"requestGuid": "request"}},
	}
	for _, tc := range cases {
		t.Run(tc.domain+"/"+tc.action, func(t *testing.T) {
			action := findDomainAction(t, tc.domain, tc.action)
			got, _ := buildRESTPath(findDomain(tc.domain), *action, tc.args)
			if got != tc.path || action.HTTPMethod != tc.method || action.ToolName != tc.tool {
				t.Fatalf("own route = %s %s (%s), want %s %s (%s)", action.HTTPMethod, got, action.ToolName, tc.method, tc.path, tc.tool)
			}
			for _, arg := range action.Args {
				if arg.Name == "userGuid" || arg.Name == "userId" || arg.Name == "tenantGuid" {
					t.Fatalf("own command must derive actor from authentication: %s", arg.Name)
				}
			}
		})
	}
	response := findDomainAction(t, "shift-request", "respond")
	if len(response.Flags) != 1 || response.Flags[0].Name != "accept" || !response.Flags[0].Required {
		t.Fatal("counterpart response requires an explicit accept decision")
	}
	withdraw := findDomainAction(t, "shift-request", "withdraw-mine")
	if len(withdraw.Flags) != 1 || withdraw.Flags[0].QueryName != "concurrencyToken" || !withdraw.Flags[0].Required {
		t.Fatal("own withdrawal requires the latest concurrency token")
	}
}
