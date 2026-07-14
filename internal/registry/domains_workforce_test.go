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
