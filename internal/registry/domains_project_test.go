package registry

import (
	"testing"
)

// --- Project domain ---

func TestProjectDomainRegistered(t *testing.T) {
	d := findDomain("project")
	if d == nil {
		t.Fatal("expected project domain to be registered")
	}
}

func TestProjectDomainAliases(t *testing.T) {
	d := findDomain("project")
	if d == nil {
		t.Fatal("expected project domain to be registered")
	}
	expected := map[string]bool{"projects": true}
	for _, alias := range d.Aliases {
		delete(expected, alias)
	}
	if len(expected) > 0 {
		t.Errorf("missing aliases: %v", expected)
	}
}

func TestProjectDomainActions(t *testing.T) {
	d := findDomain("project")
	if d == nil {
		t.Fatal("expected project domain to be registered")
	}

	// Project-specific actions layered on top of the standard CRUD set.
	// The CRUD set (list/get/create/update/delete) is covered by registry_test.go.
	expected := map[string]string{
		"search":       "UteamupProjectSearch",
		"my-projects":  "UteamupProjectMyProjects",
		"set-status":   "UteamupProjectSetStatus",
		"set-priority": "UteamupProjectSetPriority",
		"set-owner":    "UteamupProjectSetOwner",
	}

	actionMap := make(map[string]string)
	for _, a := range d.Actions {
		actionMap[a.Name] = a.ToolName
	}

	for name, tool := range expected {
		if actual, ok := actionMap[name]; !ok {
			t.Errorf("missing action %q", name)
		} else if actual != tool {
			t.Errorf("action %q: expected tool %q, got %q", name, tool, actual)
		}
	}
}

func TestProjectMyProjectsIsArgless(t *testing.T) {
	d := findDomain("project")
	if d == nil {
		t.Fatal("expected project domain to be registered")
	}

	var myProjects *Action
	for i := range d.Actions {
		if d.Actions[i].Name == "my-projects" {
			myProjects = &d.Actions[i]
			break
		}
	}
	if myProjects == nil {
		t.Fatal("expected my-projects action to exist")
	}

	// The backend resolves the user from the API key, so the CLI MUST NOT
	// require positional args or flags. Changing this contract (e.g. adding
	// a userId flag) would make the command misleading.
	if len(myProjects.Args) != 0 {
		t.Errorf("my-projects should take no args, got %d", len(myProjects.Args))
	}
	if len(myProjects.Flags) != 0 {
		t.Errorf("my-projects should take no flags, got %d", len(myProjects.Flags))
	}
}

func TestProjectByGuidSetterActionsWired(t *testing.T) {
	// GUID-keyed field setters mirror ProjectController's by-guid PUT routes.
	// Both identifiers ride the URL — no flags, no body — so both positional
	// args must literally match the RESTPath placeholders.
	cases := []struct {
		action   string
		tool     string
		restPath string
		arg2Name string
		arg2Type string
	}{
		{"set-status", "UteamupProjectSetStatus", "by-guid/{projectGuid}/status/{statusId}", "statusId", "int"},
		{"set-priority", "UteamupProjectSetPriority", "by-guid/{projectGuid}/priority/{priorityId}", "priorityId", "int"},
		{"set-owner", "UteamupProjectSetOwner", "by-guid/{projectGuid}/owner/{ownerId}", "ownerId", "string"},
	}

	for _, c := range cases {
		a := findDomainAction(t, "project", c.action)
		if a.ToolName != c.tool || a.HTTPMethod != "PUT" || a.RESTPath != c.restPath {
			t.Errorf("%s: want tool=%s method=PUT path=%s, got tool=%s method=%q path=%s",
				c.action, c.tool, c.restPath, a.ToolName, a.HTTPMethod, a.RESTPath)
		}
		if len(a.Args) != 2 {
			t.Errorf("%s should take exactly 2 positional args, got %+v", c.action, a.Args)
			continue
		}
		if a.Args[0].Name != "projectGuid" || !a.Args[0].Required || a.Args[0].Type != "string" {
			t.Errorf("%s first arg must be required string 'projectGuid', got %+v", c.action, a.Args[0])
		}
		if a.Args[1].Name != c.arg2Name || !a.Args[1].Required || a.Args[1].Type != c.arg2Type {
			t.Errorf("%s second arg must be required %s %q, got %+v", c.action, c.arg2Type, c.arg2Name, a.Args[1])
		}
		if len(a.Flags) != 0 {
			t.Errorf("%s should take no flags (identifiers ride the URL), got %d", c.action, len(a.Flags))
		}
	}
}
