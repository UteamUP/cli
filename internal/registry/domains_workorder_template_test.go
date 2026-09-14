package registry

import "testing"

// Bug c5f9a884 — /workorders/templates scheduled templates need browser-visible
// generation proof. The CLI exposes the same backend "Generate now" endpoint
// via a domain verb. Losing this registration would silently leave the CLI in
// drift with the backend; assert all three contract points: the domain
// exists, the action exists with the right tool name, and the GUID arg is
// required.

func TestWorkorderTemplateDomainHasRunScheduleNowAction(t *testing.T) {
	d := findDomain("workorder-template")
	if d == nil {
		t.Fatal("expected workorder-template domain to be registered")
	}

	var action *Action
	for i := range d.Actions {
		if d.Actions[i].Name == "run-schedule-now" {
			action = &d.Actions[i]
			break
		}
	}
	if action == nil {
		t.Fatal("expected run-schedule-now action on the workorder-template domain")
	}

	if action.ToolName != "UteamupWorkorderTemplateRunScheduleNow" {
		t.Errorf("run-schedule-now: expected tool UteamupWorkorderTemplateRunScheduleNow, got %q", action.ToolName)
	}

	if len(action.Args) != 1 {
		t.Fatalf("run-schedule-now: expected exactly one positional arg, got %d", len(action.Args))
	}
	arg := action.Args[0]
	if arg.Name != "scheduleGuid" {
		t.Errorf("run-schedule-now: expected arg named 'scheduleGuid', got %q", arg.Name)
	}
	if !arg.Required {
		t.Error("run-schedule-now: scheduleGuid arg must be marked Required")
	}
	if arg.Type != "string" {
		t.Errorf("run-schedule-now: scheduleGuid arg type must be 'string' (a guid is passed as a string), got %q", arg.Type)
	}
}

func TestWorkorderTemplateDomainHasWotAlias(t *testing.T) {
	d := findDomain("workorder-template")
	if d == nil {
		t.Fatal("expected workorder-template domain to be registered")
	}

	hasWot := false
	for _, alias := range d.Aliases {
		if alias == "wot" {
			hasWot = true
			break
		}
	}
	if !hasWot {
		t.Errorf("workorder-template domain must keep the 'wot' alias; got %v", d.Aliases)
	}
}

func TestWorkorderTemplateDomainHasApprovedActiveRead(t *testing.T) {
	d := findDomain("workorder-template")
	if d == nil {
		t.Fatal("expected workorder-template domain to be registered")
	}

	var action *Action
	for index := range d.Actions {
		if d.Actions[index].Name == "active" {
			action = &d.Actions[index]
			break
		}
	}
	if action == nil {
		t.Fatal("expected active action on workorder-template domain")
	}
	if action.ToolName != "UteamupWorkorderTemplateGetActive" {
		t.Errorf("ToolName = %q, want UteamupWorkorderTemplateGetActive", action.ToolName)
	}
	if action.HTTPMethod != "GET" || action.RESTPath != "" || len(action.Args) != 0 {
		t.Errorf("active route = method %q path %q args %+v, want bounded GET base route",
			action.HTTPMethod, action.RESTPath, action.Args)
	}

	flags := make(map[string]FlagDef, len(action.Flags))
	for _, flag := range action.Flags {
		flags[flag.Name] = flag
	}
	if flag, ok := flags["is-active"]; !ok || flag.Default != true || flag.Type != "bool" {
		t.Errorf("is-active = %+v, want bool default true", flag)
	}
	if flag, ok := flags["page-size"]; !ok || flag.Default != 20 || flag.Type != "int" {
		t.Errorf("page-size = %+v, want int default 20", flag)
	}
}

// The list action filters by project ownership. Without --project-guid the
// general library is listed, which excludes workorder templates owned by a
// project template; with it, only that project's or project template's own
// templates come back. The REST route and the UteamupWorkorderTemplateList tool
// both bind the filter as projectGuid.
func TestWorkorderTemplateListFiltersByProjectGuid(t *testing.T) {
	action := findDomainAction(t, "workorder-template", "list")
	if action.ToolName != "UteamupWorkorderTemplateList" {
		t.Errorf("list: expected tool UteamupWorkorderTemplateList, got %q", action.ToolName)
	}
	if action.MCPOnly || action.HTTPMethod != "" || action.RESTPath != "" || len(action.Args) != 0 {
		t.Errorf("list route = mcpOnly %v method %q path %q args %+v, want the plain GET collection route",
			action.MCPOnly, action.HTTPMethod, action.RESTPath, action.Args)
	}

	flag := findFlag(action, "project-guid")
	if flag == nil {
		t.Fatal("list must expose an optional --project-guid filter")
	}
	// uuid validation matters here because GET flags are copied into the query
	// string unescaped. A default would filter every call and hide the library.
	if flag.BodyName != "projectGuid" || flag.Type != "uuid" || flag.Required || flag.Default != nil {
		t.Errorf("--project-guid = %+v, want optional uuid sent as projectGuid with no default", *flag)
	}

	// The filter joins the standard pagination flags instead of replacing them.
	for _, name := range []string{"page", "page-size"} {
		if findFlag(action, name) == nil {
			t.Errorf("list lost its --%s pagination flag", name)
		}
	}
}
