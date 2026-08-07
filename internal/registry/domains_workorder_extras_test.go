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

func workorderSignatureDomain(t *testing.T) *Domain {
	t.Helper()
	domain := findDomain("workorder-signature")
	if domain == nil {
		t.Fatal("expected workorder-signature domain to be registered")
	}
	return domain
}

func workorderSignatureAction(t *testing.T, name string) *Action {
	t.Helper()
	action := findAction(workorderSignatureDomain(t), name)
	if action == nil {
		t.Fatalf("expected %s action on the workorder-signature domain", name)
	}
	return action
}

func TestWorkorderSignatureActionsMirrorBackendTools(t *testing.T) {
	expected := map[string]string{
		"summary":              "UteamupSignatureGetSummary",
		"compute-requirements": "UteamupSignatureComputeRequirements",
		"create-requirements":  "UteamupSignatureCreateRequirements",
		"assign-signer":        "UteamupSignatureAssignSigner",
		"remove-signer":        "UteamupSignatureRemoveSigner",
		"settings":             "UteamupSignatureGetSettings",
		"dashboard":            "UteamupSignatureGetDashboard",
		"analytics":            "UteamupSignatureGetAnalytics",
	}

	domain := workorderSignatureDomain(t)
	if len(domain.Actions) != len(expected) {
		t.Fatalf("workorder-signature actions = %d, want %d", len(domain.Actions), len(expected))
	}

	for actionName, toolName := range expected {
		action := workorderSignatureAction(t, actionName)
		if action.ToolName != toolName {
			t.Errorf("%s tool = %q, want %q", actionName, action.ToolName, toolName)
		}
		if !action.MCPOnly {
			t.Errorf("%s must use the MCP transport because it has no REST adapter", actionName)
		}
	}
}

func TestWorkorderSignatureOperationsUsePublicGuidArguments(t *testing.T) {
	expected := map[string]string{
		"summary":              "workorderGuid",
		"compute-requirements": "workorderGuid",
		"create-requirements":  "workorderGuid",
		"assign-signer":        "workorderGuid",
		"remove-signer":        "requirementGuid",
	}
	for actionName, argumentName := range expected {
		action := workorderSignatureAction(t, actionName)
		if len(action.Args) != 1 {
			t.Fatalf("%s arguments = %d, want one public GUID", actionName, len(action.Args))
		}
		argument := action.Args[0]
		if argument.Name != argumentName || argument.Type != "uuid" || !argument.Required {
			t.Errorf(
				"%s argument = %+v, want required uuid %s",
				actionName,
				argument,
				argumentName,
			)
		}
	}
}

func TestWorkorderSignatureAssignSignerUsesReviewedGuidModel(t *testing.T) {
	action := workorderSignatureAction(t, "assign-signer")
	if len(action.Flags) != 1 {
		t.Fatalf("assign-signer flags = %d, want one JSON model flag", len(action.Flags))
	}
	model := action.Flags[0]
	if model.BodyName != "model" || model.Type != "string" || !model.Required || !model.JSONFile {
		t.Errorf("assign-signer model flag = %+v, want required JSON file mapped to model", model)
	}
}

func TestWorkorderSignatureAnalyticsUsesDateArguments(t *testing.T) {
	action := workorderSignatureAction(t, "analytics")
	expected := map[string]string{
		"start-date": "startDate",
		"end-date":   "endDate",
	}
	if len(action.Flags) != len(expected) {
		t.Fatalf("analytics flags = %d, want %d", len(action.Flags), len(expected))
	}
	for _, flag := range action.Flags {
		bodyName, ok := expected[flag.Name]
		if !ok {
			t.Errorf("analytics exposes unexpected flag --%s", flag.Name)
			continue
		}
		if flag.BodyName != bodyName || flag.Type != "string" || !flag.Required {
			t.Errorf("analytics flag --%s = %+v, want required string mapped to %s", flag.Name, flag, bodyName)
		}
	}
}

func TestWorkorderSignatureRegistryRejectsInternalIdentifiers(t *testing.T) {
	internalIdentifiers := map[string]bool{
		"id":            true,
		"workorderId":   true,
		"requirementId": true,
		"signerId":      true,
		"contactId":     true,
		"signerGroupId": true,
	}

	for _, action := range workorderSignatureDomain(t).Actions {
		for _, argument := range action.Args {
			name := argument.BodyName
			if name == "" {
				name = argument.Name
			}
			if internalIdentifiers[name] || argument.Type == "int" {
				t.Errorf("%s exposes internal argument %q with type %q", action.Name, name, argument.Type)
			}
		}
		for _, flag := range action.Flags {
			name := flag.BodyName
			if name == "" {
				name = toCamelCase(flag.Name)
			}
			if internalIdentifiers[name] || flag.Type == "int" {
				t.Errorf("%s exposes internal flag %q with type %q", action.Name, name, flag.Type)
			}
		}
	}
}

func TestWorkorderSignatureSettingsAndDashboardNeedNoArguments(t *testing.T) {
	for _, actionName := range []string{"settings", "dashboard"} {
		action := workorderSignatureAction(t, actionName)
		if len(action.Args) != 0 || len(action.Flags) != 0 {
			t.Errorf("%s must not require arguments or flags", actionName)
		}
	}
}
