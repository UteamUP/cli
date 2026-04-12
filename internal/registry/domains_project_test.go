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

	// "search" and "my-projects" are the two project-specific actions layered
	// on top of the standard CRUD set. The CRUD set (list/get/create/update/delete)
	// is covered by registry_test.go.
	expected := map[string]string{
		"search":      "UteamupProjectSearch",
		"my-projects": "UteamupProjectMyProjects",
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
