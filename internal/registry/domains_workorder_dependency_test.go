package registry

import (
	"reflect"
	"testing"
)

func domainAction(t *testing.T, domainName, actionName string) (*Domain, Action) {
	t.Helper()
	domain := findDomain(domainName)
	if domain == nil {
		t.Fatalf("%s domain is not registered", domainName)
	}
	for _, action := range domain.Actions {
		if action.Name == actionName {
			return domain, action
		}
	}
	t.Fatalf("%s action %q is not registered", domainName, actionName)
	return nil, Action{}
}

func TestWorkorderDependencyRoutesUseGuidIdentity(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		args     map[string]any
		path     string
		consumed []string
	}{
		{name: "list", args: map[string]any{"workorderGuid": "wo-guid"},
			path: "/api/workorder/wo-guid/dependencies", consumed: []string{"workorderGuid"}},
		{name: "create", args: map[string]any{"workorderGuid": "wo-guid"},
			path: "/api/workorder/wo-guid/dependencies", consumed: []string{"workorderGuid"}},
		{name: "delete", args: map[string]any{"workorderGuid": "wo-guid", "dependencyGuid": "dep-guid"},
			path: "/api/workorder/wo-guid/dependencies/dep-guid", consumed: []string{"workorderGuid", "dependencyGuid"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			domain, action := domainAction(t, "workorder-dependency", testCase.name)
			path, consumed := buildRESTPath(domain, action, testCase.args)
			if path != testCase.path {
				t.Fatalf("path = %q, want %q", path, testCase.path)
			}
			if !reflect.DeepEqual(consumed, testCase.consumed) {
				t.Fatalf("consumed = %v, want %v", consumed, testCase.consumed)
			}
			for _, arg := range action.Args {
				if arg.Type != "uuid" {
					t.Fatalf("public identity arg is not UUID typed: %+v", arg)
				}
			}
		})
	}
}

func TestWorkorderDependencyCreateSendsOnlyThePrerequisiteGuid(t *testing.T) {
	t.Parallel()

	_, create := domainAction(t, "workorder-dependency", "create")
	if len(create.Flags) != 1 {
		t.Fatalf("create flags = %d, want 1", len(create.Flags))
	}
	flag := create.Flags[0]
	if flag.Name != "prerequisite-workorder-guid" || flag.BodyName != "prerequisiteWorkorderGuid" ||
		flag.Type != "non-empty-uuid" || !flag.Required {
		t.Fatalf("unexpected create flag: %+v", flag)
	}
}

func TestWorkorderDependencyToolNamesMirrorBackend(t *testing.T) {
	t.Parallel()

	want := map[string]string{
		"list":   "UteamupWorkorderDependencyList",
		"create": "UteamupWorkorderDependencyCreate",
		"delete": "UteamupWorkorderDependencyDelete",
	}
	for actionName, toolName := range want {
		_, action := domainAction(t, "workorder-dependency", actionName)
		if action.ToolName != toolName {
			t.Fatalf("%s tool = %q, want %q", actionName, action.ToolName, toolName)
		}
	}
}
