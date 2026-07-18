package registry

import "testing"

func TestWorkforceGroupDomainIsGuidFirst(t *testing.T) {
	d := findDomain("workforce-group")
	if d == nil {
		t.Fatal("expected workforce-group domain to be registered")
	}
	if d.APIPath != "/api/workforcegroups" {
		t.Fatalf("workforce-group APIPath = %q, want /api/workforcegroups", d.APIPath)
	}

	actions := map[string]Action{}
	for _, action := range d.Actions {
		actions[action.Name] = action
	}

	for _, name := range []string{"get", "update", "delete"} {
		action, ok := actions[name]
		if !ok {
			t.Fatalf("missing workforce-group action %q", name)
		}
		if len(action.Args) != 1 || action.Args[0].Name != "groupGuid" {
			t.Fatalf("%s args = %+v, want single groupGuid arg", name, action.Args)
		}
		if action.Args[0].Type != "string" {
			t.Fatalf("%s groupGuid type = %q, want string", name, action.Args[0].Type)
		}
		if action.RESTPath != "by-guid/{groupGuid}" {
			t.Fatalf("%s RESTPath = %q, want by-guid/{groupGuid}", name, action.RESTPath)
		}
	}
}

func TestWorkforceGroupGuidRoutesResolve(t *testing.T) {
	d := findDomain("workforce-group")
	if d == nil {
		t.Fatal("expected workforce-group domain to be registered")
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
		{"get", map[string]any{"groupGuid": "group-1"}, "/api/workforcegroups/by-guid/group-1"},
		{"update", map[string]any{"groupGuid": "group-1"}, "/api/workforcegroups/by-guid/group-1"},
		{"delete", map[string]any{"groupGuid": "group-1"}, "/api/workforcegroups/by-guid/group-1"},
		{"members", map[string]any{"groupGuid": "group-1"}, "/api/workforcegroups/by-guid/group-1/members"},
		{"member-add", map[string]any{"groupGuid": "group-1"}, "/api/workforcegroups/by-guid/group-1/members"},
		{
			"member-remove",
			map[string]any{"groupGuid": "group-1", "memberGuid": "member-1"},
			"/api/workforcegroups/by-guid/group-1/members/by-guid/member-1",
		},
	}

	for _, tc := range cases {
		action, ok := actions[tc.name]
		if !ok {
			t.Fatalf("missing workforce-group action %q", tc.name)
		}
		got, consumed := buildRESTPath(d, action, tc.args)
		if got != tc.want {
			t.Fatalf("%s path = %q, want %q", tc.name, got, tc.want)
		}
		if len(consumed) != len(action.Args) {
			t.Fatalf("%s consumed = %v, want %d path args", tc.name, consumed, len(action.Args))
		}
	}
}

func TestWorkforceTrainingDomainIsGuidFirst(t *testing.T) {
	d := findDomain("workforce-training")
	if d == nil {
		t.Fatal("expected workforce-training domain to be registered")
	}
	if d.APIPath != "/api/workforcegrouprequiredtraining" {
		t.Fatalf("workforce-training APIPath = %q, want /api/workforcegrouprequiredtraining", d.APIPath)
	}

	actions := map[string]Action{}
	for _, action := range d.Actions {
		actions[action.Name] = action
	}

	for _, name := range []string{"update", "delete"} {
		action, ok := actions[name]
		if !ok {
			t.Fatalf("missing workforce-training action %q", name)
		}
		if len(action.Args) != 1 || action.Args[0].Name != "trainingGuid" {
			t.Fatalf("%s args = %+v, want single trainingGuid arg", name, action.Args)
		}
		if action.RESTPath != "by-guid/{trainingGuid}" {
			t.Fatalf("%s RESTPath = %q, want by-guid/{trainingGuid}", name, action.RESTPath)
		}
	}

	list, ok := actions["list"]
	if !ok {
		t.Fatal("missing workforce-training list action")
	}
	if len(list.Flags) != 1 || list.Flags[0].Name != "group-guid" {
		t.Fatalf("list flags = %+v, want group-guid filter", list.Flags)
	}
}

func TestWorkforceTrainingGuidRoutesResolve(t *testing.T) {
	d := findDomain("workforce-training")
	if d == nil {
		t.Fatal("expected workforce-training domain to be registered")
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
		{
			"update",
			map[string]any{"trainingGuid": "training-1"},
			"/api/workforcegrouprequiredtraining/by-guid/training-1",
		},
		{
			"delete",
			map[string]any{"trainingGuid": "training-1"},
			"/api/workforcegrouprequiredtraining/by-guid/training-1",
		},
	}

	for _, tc := range cases {
		action, ok := actions[tc.name]
		if !ok {
			t.Fatalf("missing workforce-training action %q", tc.name)
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

func TestWorkforcePlanningActionsMirrorMcpTools(t *testing.T) {
	expected := map[string]string{
		"grid":                  "UteamupWorkforcePlanningGrid",
		"assign":                "UteamupWorkforceAssignWorkOrder",
		"available-technicians": "UteamupWorkforceGetAvailableTechnicians",
		"group-members":         "UteamupWorkforceGetGroupMembers",
	}
	for name, toolName := range expected {
		action := findDomainAction(t, "workforce-planning", name)
		if action.ToolName != toolName {
			t.Errorf("%s ToolName = %q, want %q", name, action.ToolName, toolName)
		}
	}
}

func TestWorkforcePlanningDomainReplacedCrudStub(t *testing.T) {
	d := findDomain("workforce-planning")
	if d == nil {
		t.Fatal("expected workforce-planning domain to be registered")
	}
	if d.APIPath != "/api/workforceplanning" {
		t.Errorf("workforce-planning APIPath = %q, want /api/workforceplanning", d.APIPath)
	}
	hasAlias := false
	for _, a := range d.Aliases {
		if a == "wp" {
			hasAlias = true
		}
	}
	if !hasAlias {
		t.Errorf("workforce-planning missing 'wp' alias, got %v", d.Aliases)
	}
	// The generic CRUD stub is gone — no list/get/create/update/delete.
	for _, action := range d.Actions {
		switch action.Name {
		case "list", "get", "create", "update", "delete":
			t.Errorf("workforce-planning must not keep the crudActions stub action %q", action.Name)
		}
	}
}

func TestWorkforcePlanningGridActionWired(t *testing.T) {
	action := findDomainAction(t, "workforce-planning", "grid")
	if action.HTTPMethod != "GET" || action.RESTPath != "grid" {
		t.Errorf("grid = %s %q, want GET \"grid\"", action.HTTPMethod, action.RESTPath)
	}
	for _, name := range []string{"date-from", "date-to"} {
		f := findFlag(action, name)
		if f == nil || !f.Required || f.Type != "string" {
			t.Errorf("grid must have a required string %q flag, got %+v", name, f)
		}
	}
	if f := findFlag(action, "category-guid"); f == nil || f.BodyName != "categoryGuids" {
		t.Errorf("grid 'category-guid' must map to categoryGuids, got %+v", f)
	}
	for _, name := range []string{"location-guid", "customer-guid"} {
		if f := findFlag(action, name); f == nil || f.Required {
			t.Errorf("grid %q must be an optional flag, got %+v", name, f)
		}
	}
}

func TestWorkforcePlanningAssignActionMapsToWorkorderAssigneeRoute(t *testing.T) {
	action := findDomainAction(t, "workforce-planning", "assign")
	if action.HTTPMethod != "PUT" {
		t.Errorf("assign HTTPMethod = %q, want PUT", action.HTTPMethod)
	}
	if action.RESTBasePath != "/api/workorder" || action.RESTPath != "by-guid/{workOrderGuid}/assignee" {
		t.Errorf("assign route = %q + %q, want /api/workorder + by-guid/{workOrderGuid}/assignee", action.RESTBasePath, action.RESTPath)
	}
	if len(action.Args) != 1 || action.Args[0].Name != "workOrderGuid" || !action.Args[0].Required || action.Args[0].Type != "uuid" {
		t.Errorf("assign must take a required uuid 'workOrderGuid' arg, got %+v", action.Args)
	}
	if f := findFlag(action, "assignee-user-guid"); f == nil || !f.Required || f.BodyName != "assigneeGuid" {
		t.Errorf("assign 'assignee-user-guid' must be required and map to assigneeGuid, got %+v", f)
	}
	if f := findFlag(action, "override-reason"); f == nil || f.Required || f.BodyName != "qualificationOverrideReason" {
		t.Errorf("assign 'override-reason' must be optional and map to qualificationOverrideReason, got %+v", f)
	}
}

func TestWorkforcePlanningAvailableTechniciansActionWired(t *testing.T) {
	action := findDomainAction(t, "workforce-planning", "available-technicians")
	if action.HTTPMethod != "GET" || action.RESTBasePath != "/api/schedule" || action.RESTPath != "available-technicians" {
		t.Errorf("available-technicians = %s %q + %q, want GET /api/schedule + available-technicians", action.HTTPMethod, action.RESTBasePath, action.RESTPath)
	}
	if f := findFlag(action, "date"); f == nil || !f.Required {
		t.Errorf("available-technicians must have a required 'date' flag, got %+v", f)
	}
	if f := findFlag(action, "required-skill-guid"); f == nil || f.BodyName != "requiredSkillGuids" {
		t.Errorf("'required-skill-guid' must map to requiredSkillGuids, got %+v", f)
	}
	if f := findFlag(action, "required-certificate-guid"); f == nil || f.BodyName != "requiredCertificateGuids" {
		t.Errorf("'required-certificate-guid' must map to requiredCertificateGuids, got %+v", f)
	}
	if f := findFlag(action, "team-guid"); f == nil || f.Required {
		t.Errorf("'team-guid' must be an optional flag, got %+v", f)
	}
}

func TestWorkforcePlanningRoutesResolve(t *testing.T) {
	d := findDomain("workforce-planning")
	if d == nil {
		t.Fatal("expected workforce-planning domain to be registered")
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
		{"grid", map[string]any{}, "/api/workforceplanning/grid"},
		{"assign", map[string]any{"workOrderGuid": "wo-1"}, "/api/workorder/by-guid/wo-1/assignee"},
		{"available-technicians", map[string]any{}, "/api/schedule/available-technicians"},
		{"group-members", map[string]any{"groupGuid": "group-1"}, "/api/workforcegroups/by-guid/group-1/members"},
	}

	for _, tc := range cases {
		action, ok := actions[tc.name]
		if !ok {
			t.Fatalf("missing workforce-planning action %q", tc.name)
		}
		got, consumed := buildRESTPath(d, action, tc.args)
		if got != tc.want {
			t.Errorf("%s path = %q, want %q", tc.name, got, tc.want)
		}
		if len(consumed) != len(action.Args) {
			t.Errorf("%s consumed = %v, want %d path args", tc.name, consumed, len(action.Args))
		}
	}
}
