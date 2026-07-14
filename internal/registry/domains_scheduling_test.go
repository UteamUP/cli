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
