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
	if a.HTTPMethod != "POST" || a.RESTPath != "{templateGuid}/create-workorder" {
		t.Errorf("create-workorder route = %s %q, want POST {templateGuid}/create-workorder", a.HTTPMethod, a.RESTPath)
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
			if f.BodyName != "templateGuid" {
				t.Errorf("flag --template BodyName = %q, want templateGuid", f.BodyName)
			}
		}
	}
	if !seen {
		t.Error("missing required flag --template")
	}
}

func TestWorkorderTemplateCreateWorkorderMirrorsCanonicalGuidOverrides(t *testing.T) {
	a := wotCreateAction(t)
	want := map[string]string{
		"asset-guids":           "assetGuids",
		"part-guids":            "partGuids",
		"tool-guids":            "toolGuids",
		"chemical-guids":        "chemicalGuids",
		"task-list-guids":       "taskListGuids",
		"check-list-guids":      "checkListGuids",
		"location-guid":         "locationGuid",
		"location-floor-guid":   "locationFloorGuid",
		"primary-assignee-guid": "primaryAssigneeGuid",
		"estimated-duration":    "estimatedDuration",
		"estimated-cost":        "estimatedCost",
	}
	seen := map[string]bool{}
	for _, flag := range a.Flags {
		bodyName, ok := want[flag.Name]
		if !ok {
			continue
		}
		seen[flag.Name] = true
		if flag.BodyName != bodyName || flag.Required {
			t.Errorf("flag --%s = BodyName %q Required %v, want %q and optional", flag.Name, flag.BodyName, flag.Required, bodyName)
		}
	}
	for name := range want {
		if !seen[name] {
			t.Errorf("missing canonical override flag --%s", name)
		}
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
