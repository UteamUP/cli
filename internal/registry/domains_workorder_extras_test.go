package registry

import (
	"testing"
)

// --- Workorder Template "create-workorder" action ---

// wotCreateAction resolves the create-workorder action via the package-shared
// findDomain/findAction helpers (declared in domains_journal_test.go).
func wotCreateAction(t *testing.T) *Action {
	t.Helper()
	d := findDomain("workorder-template")
	if d == nil {
		t.Fatal("expected workorder-template domain to be registered")
	}
	a := findAction(d, "create-workorder")
	if a == nil {
		t.Fatal("expected create-workorder action on the workorder-template domain")
	}
	return a
}

// The create-workorder action spawns an open work order from a template via
// its public GUID. The tool name is the backend contract — a typo here ships
// a command that always 404s server-side.
func TestWorkorderTemplateCreateWorkorderTool(t *testing.T) {
	a := wotCreateAction(t)
	if a.ToolName != "UteamupWorkorderTemplateCreateFromTemplateByGuid" {
		t.Errorf("create-workorder: expected tool UteamupWorkorderTemplateCreateFromTemplateByGuid, got %q", a.ToolName)
	}
}

// --template carries the template's public GUID and is the only required
// input. Losing Required would let callers fire a body-less request the
// backend rejects.
func TestWorkorderTemplateCreateWorkorderRequiresTemplate(t *testing.T) {
	a := wotCreateAction(t)
	var seen bool
	for _, f := range a.Flags {
		if f.Name == "template" {
			seen = true
			if !f.Required {
				t.Error("flag --template must be marked Required")
			}
		}
	}
	if !seen {
		t.Error("missing required flag --template")
	}
}

// name/description/priority/notes are overrides — forcing any of them would
// defeat the "no asset or resolution note required" usability goal.
func TestWorkorderTemplateCreateWorkorderOverridesAreOptional(t *testing.T) {
	a := wotCreateAction(t)
	mustBeOptional := map[string]bool{
		"name":        true,
		"description": true,
		"priority":    true,
		"notes":       true,
	}
	for _, f := range a.Flags {
		if mustBeOptional[f.Name] && f.Required {
			t.Errorf("flag --%s must be optional", f.Name)
		}
	}
}
