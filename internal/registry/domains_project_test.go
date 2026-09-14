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
		"search":                           "UteamupProjectSearch",
		"my-projects":                      "UteamupProjectMyProjects",
		"set-status":                       "UteamupProjectSetStatus",
		"set-priority":                     "UteamupProjectSetPriority",
		"set-owner":                        "UteamupProjectSetOwner",
		"templates":                        "UteamupProjectTemplateList",
		"from-template":                    "UteamupProjectCreateFromTemplate",
		"save-as-template":                 "UteamupProjectSaveAsTemplate",
		"template-add-workorder-templates": "UteamupProjectTemplateAddWorkorderTemplates",
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

func TestProjectCrudIsGuidFirst(t *testing.T) {
	d := findDomain("project")
	if d == nil {
		t.Fatal("expected project domain to be registered")
	}

	actionMap := make(map[string]Action)
	for _, a := range d.Actions {
		actionMap[a.Name] = a
	}

	// get/update/delete must take the project ExternalGuid positional arg
	// (string), never a legacy integer id — GUIDs In, Integer IDs Out.
	for _, name := range []string{"get", "update", "delete"} {
		a, ok := actionMap[name]
		if !ok {
			t.Errorf("missing CRUD action %q", name)
			continue
		}
		if len(a.Args) != 1 {
			t.Errorf("action %q: expected 1 positional arg, got %d", name, len(a.Args))
			continue
		}
		if a.Args[0].Name != "externalGuid" {
			t.Errorf("action %q: expected positional arg %q, got %q", name, "externalGuid", a.Args[0].Name)
		}
		if a.Args[0].Type != "string" {
			t.Errorf("action %q: identity arg must be string (guid), got %q", name, a.Args[0].Type)
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
		{"set-owner", "UteamupProjectSetOwner", "by-guid/{projectGuid}/owner-guid/{ownerGuid}", "ownerGuid", "string"},
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

// --- Project templates ---

// projectToolFlag is the MCP tool argument a flag must reach, with the flag's
// CLI type and whether the command requires it.
type projectToolFlag struct {
	argument string
	flagType string
	required bool
}

// projectToolArgument resolves the argument name executeAction sends a flag
// under: its BodyName, or the camelCased flag name when BodyName is empty.
func projectToolArgument(flag FlagDef) string {
	if flag.BodyName != "" {
		return flag.BodyName
	}
	return toCamelCase(flag.Name)
}

// assertProjectToolGUIDArg verifies the action takes exactly one positional arg:
// a required public GUID sent under the tool's exact argument name.
func assertProjectToolGUIDArg(t *testing.T, action *Action, argument string) {
	t.Helper()
	if len(action.Args) != 1 {
		t.Fatalf("%s positional args = %+v, want exactly one %s", action.Name, action.Args, argument)
	}
	arg := action.Args[0]
	sent := arg.Name
	if arg.BodyName != "" {
		sent = arg.BodyName
	}
	if arg.Name != argument || sent != argument || !arg.Required || arg.Type != "uuid" {
		t.Errorf("%s positional arg = %+v, want required uuid %q sent as %q", action.Name, arg, argument, argument)
	}
}

// assertProjectToolFlags verifies the action's flags are exactly want. Every
// flag becomes a tool argument, so an extra or misnamed flag breaks the contract.
func assertProjectToolFlags(t *testing.T, action *Action, want map[string]projectToolFlag) {
	t.Helper()
	if len(action.Flags) != len(want) {
		t.Errorf("%s flags = %d, want exactly %d", action.Name, len(action.Flags), len(want))
	}
	for _, flag := range action.Flags {
		expected, ok := want[flag.Name]
		if !ok {
			t.Errorf("%s exposes unexpected flag --%s", action.Name, flag.Name)
			continue
		}
		if argument := projectToolArgument(flag); argument != expected.argument {
			t.Errorf("%s --%s is sent as %q, want tool argument %q", action.Name, flag.Name, argument, expected.argument)
		}
		if flag.Type != expected.flagType || flag.Required != expected.required {
			t.Errorf("%s --%s = type %q required %v, want type %q required %v",
				action.Name, flag.Name, flag.Type, flag.Required, expected.flagType, expected.required)
		}
	}
	for name := range want {
		if findFlag(action, name) == nil {
			t.Errorf("%s is missing flag --%s", action.Name, name)
		}
	}
}

func TestProjectTemplateActionsCallTheirMCPTools(t *testing.T) {
	// The template tools have no REST adapter. Routed over REST, `templates`
	// would silently GET /api/project and list projects instead of templates.
	expected := map[string]string{
		"templates":                        "UteamupProjectTemplateList",
		"from-template":                    "UteamupProjectCreateFromTemplate",
		"save-as-template":                 "UteamupProjectSaveAsTemplate",
		"template-add-workorder-templates": "UteamupProjectTemplateAddWorkorderTemplates",
	}
	for name, tool := range expected {
		a := findDomainAction(t, "project", name)
		if a.ToolName != tool {
			t.Errorf("%s: expected tool %q, got %q", name, tool, a.ToolName)
		}
		if !a.MCPOnly {
			t.Errorf("%s must use the MCP transport because its tool has no REST adapter", name)
		}
	}
}

func TestProjectTemplatesIsArgless(t *testing.T) {
	// UteamupProjectTemplateList takes no arguments.
	a := findDomainAction(t, "project", "templates")
	if len(a.Args) != 0 || len(a.Flags) != 0 {
		t.Errorf("templates must take no args or flags, got args %+v flags %+v", a.Args, a.Flags)
	}
}

func TestProjectFromTemplateMapsToolArguments(t *testing.T) {
	a := findDomainAction(t, "project", "from-template")
	assertProjectToolGUIDArg(t, a, "templateGuid")
	assertProjectToolFlags(t, a, map[string]projectToolFlag{
		"name":             {"name", "string", true},
		"start-date-utc":   {"startDateUtc", "string", true},
		"owner-guid":       {"ownerGuid", "uuid", true},
		"description":      {"description", "string", false},
		"project-code":     {"projectCode", "string", false},
		"end-date-utc":     {"endDateUtc", "string", false},
		"priority":         {"priority", "int", false},
		"budget":           {"budget", "float", false},
		"customer-guid":    {"customerGuid", "uuid", false},
		"location-guid":    {"locationGuid", "uuid", false},
		"asset-group-guid": {"assetGroupGuid", "uuid", false},
	})

	// Defaults mirror the tool's own (priority 0, budget 0) and must be typed
	// literals: float flag defaults are float literals (CLI rule 7).
	if priority := findFlag(a, "priority"); priority != nil {
		if v, ok := priority.Default.(int); !ok || v != 0 {
			t.Errorf("priority Default must be the int literal 0, got %T (%v)", priority.Default, priority.Default)
		}
	}
	if budget := findFlag(a, "budget"); budget != nil {
		if v, ok := budget.Default.(float64); !ok || v != 0 {
			t.Errorf("budget Default must be the float64 literal 0.0, got %T (%v)", budget.Default, budget.Default)
		}
	}
}

func TestProjectSaveAsTemplateMapsToolArguments(t *testing.T) {
	a := findDomainAction(t, "project", "save-as-template")
	assertProjectToolGUIDArg(t, a, "projectGuid")
	assertProjectToolFlags(t, a, map[string]projectToolFlag{
		"name":            {"name", "string", true},
		"project-code":    {"projectCode", "string", false},
		"description":     {"description", "string", false},
		"workorder-guids": {"workorderGuids", "stringSlice", false},
	})
}

func TestProjectTemplateAddWorkorderTemplatesUsesModelFile(t *testing.T) {
	// Like create/update, the complete model rides a parsed JSON file sent under
	// the tool's named `model` argument, never an unwrapped REST body.
	a := findDomainAction(t, "project", "template-add-workorder-templates")
	assertProjectToolGUIDArg(t, a, "projectGuid")
	if len(a.Flags) != 1 {
		t.Fatalf("template-add-workorder-templates flags = %+v, want one model file flag", a.Flags)
	}
	model := a.Flags[0]
	if model.Name != "from-json" || model.BodyName != "model" || model.Type != "string" ||
		!model.Required || !model.JSONFile || model.RootJSONObjectFile {
		t.Errorf("model flag = %+v, want required --from-json JSON file sent as model", model)
	}
}
