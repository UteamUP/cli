package registry

import (
	"testing"
)

func TestCodingSystemDomainRegistered(t *testing.T) {
	var csDomain *Domain
	for _, d := range DefaultRegistry.Domains() {
		if d.Name == "codingsystem" {
			csDomain = d
			break
		}
	}
	if csDomain == nil {
		t.Fatal("expected codingsystem domain to be registered")
	}

	// Verify aliases
	expectedAliases := map[string]bool{"cs": true, "coding": true}
	for _, alias := range csDomain.Aliases {
		if !expectedAliases[alias] {
			t.Errorf("unexpected alias %q", alias)
		}
		delete(expectedAliases, alias)
	}
	if len(expectedAliases) > 0 {
		t.Errorf("missing aliases: %v", expectedAliases)
	}
}

func TestCodingSystemDomainActions(t *testing.T) {
	var csDomain *Domain
	for _, d := range DefaultRegistry.Domains() {
		if d.Name == "codingsystem" {
			csDomain = d
			break
		}
	}
	if csDomain == nil {
		t.Fatal("expected codingsystem domain to be registered")
	}

	expectedActions := map[string]string{
		"list":             "UteamupCodingsystemList",
		"tree":             "UteamupCodingsystemTree",
		"search":           "UteamupCodingsystemSearchAssets",
		"next-code":        "UteamupCodingsystemNextCode",
		"assign":           "UteamupCodingsystemAssignCode",
		"workorders":       "UteamupCodingsystemWorkorders",
		"create-workorder": "UteamupCodingsystemCreateWorkorder",
	}

	actionMap := make(map[string]string)
	for _, a := range csDomain.Actions {
		actionMap[a.Name] = a.ToolName
	}

	for name, tool := range expectedActions {
		if actual, ok := actionMap[name]; !ok {
			t.Errorf("missing action %q", name)
		} else if actual != tool {
			t.Errorf("action %q: expected tool %q, got %q", name, tool, actual)
		}
	}
}

func TestCodingSystemTreeFlags(t *testing.T) {
	var csDomain *Domain
	for _, d := range DefaultRegistry.Domains() {
		if d.Name == "codingsystem" {
			csDomain = d
			break
		}
	}
	if csDomain == nil {
		t.Fatal("expected codingsystem domain to be registered")
	}

	var treeAction *Action
	for i := range csDomain.Actions {
		if csDomain.Actions[i].Name == "tree" {
			treeAction = &csDomain.Actions[i]
			break
		}
	}
	if treeAction == nil {
		t.Fatal("expected tree action")
	}

	// Should have coding-system-id (required) and parent-id (optional)
	flagMap := make(map[string]FlagDef)
	for _, f := range treeAction.Flags {
		flagMap[f.Name] = f
	}

	csFlag, ok := flagMap["coding-system-id"]
	if !ok {
		t.Fatal("missing coding-system-id flag")
	}
	if !csFlag.Required {
		t.Error("coding-system-id should be required")
	}

	parentFlag, ok := flagMap["parent-id"]
	if !ok {
		t.Fatal("missing parent-id flag")
	}
	if parentFlag.Required {
		t.Error("parent-id should not be required")
	}
}

func TestWorkorderDomainHasByCodeAction(t *testing.T) {
	var woDomain *Domain
	for _, d := range DefaultRegistry.Domains() {
		if d.Name == "workorder" {
			woDomain = d
			break
		}
	}
	if woDomain == nil {
		t.Fatal("expected workorder domain to be registered")
	}

	found := false
	for _, a := range woDomain.Actions {
		if a.Name == "by-code" {
			found = true
			if a.ToolName != "UteamupCodingsystemWorkorders" {
				t.Errorf("by-code action: expected tool UteamupCodingsystemWorkorders, got %q", a.ToolName)
			}
			break
		}
	}
	if !found {
		t.Error("expected by-code action on workorder domain")
	}
}
