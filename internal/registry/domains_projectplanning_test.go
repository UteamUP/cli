package registry

import (
	"testing"
)

// --- Project planning domains: project-stage, project-output, project-budget ---

func findDomainAction(t *testing.T, domainName, actionName string) *Action {
	t.Helper()
	d := findDomain(domainName)
	if d == nil {
		t.Fatalf("expected %s domain to be registered", domainName)
	}
	for i := range d.Actions {
		if d.Actions[i].Name == actionName {
			return &d.Actions[i]
		}
	}
	t.Fatalf("expected `%s` action on %s domain", actionName, domainName)
	return nil
}

func findFlag(action *Action, name string) *FlagDef {
	for i := range action.Flags {
		if action.Flags[i].Name == name {
			return &action.Flags[i]
		}
	}
	return nil
}

// assertProjectGuidArgs verifies the action's positional args are exactly
// projectGuid (+ optional subGuid), all required strings — the names must
// literally match the RESTPath placeholders or expandPathTemplate leaves the
// raw token in the URL.
func assertProjectGuidArgs(t *testing.T, action *Action, subGuid string) {
	t.Helper()
	want := 1
	if subGuid != "" {
		want = 2
	}
	if len(action.Args) != want {
		t.Fatalf("%s expected %d positional args, got %+v", action.Name, want, action.Args)
	}
	if action.Args[0].Name != "projectGuid" || !action.Args[0].Required || action.Args[0].Type != "string" {
		t.Errorf("%s first arg must be required string 'projectGuid', got %+v", action.Name, action.Args[0])
	}
	if subGuid != "" && (action.Args[1].Name != subGuid || !action.Args[1].Required || action.Args[1].Type != "string") {
		t.Errorf("%s second arg must be required string %q, got %+v", action.Name, subGuid, action.Args[1])
	}
}

// --- project-stage ---

func TestProjectStageDomainRegistered(t *testing.T) {
	d := findDomain("project-stage")
	if d == nil {
		t.Fatal("expected project-stage domain to be registered")
	}
	// ProjectStageController routes under /api/projects (plural) — NOT the
	// /api/project base the `project` domain auto-derives.
	if d.APIPath != "/api/projects" {
		t.Errorf("project-stage APIPath = %q, want /api/projects", d.APIPath)
	}
	expected := map[string]bool{"project-stages": true, "stages": true}
	for _, alias := range d.Aliases {
		delete(expected, alias)
	}
	if len(expected) > 0 {
		t.Errorf("missing aliases: %v", expected)
	}
}

func TestProjectStageActionRouteTemplates(t *testing.T) {
	// Method "" means derived from the action name via the HTTPMethod map
	// (list/get→GET, create→POST, update→PUT, delete→DELETE).
	cases := []struct {
		action   string
		tool     string
		method   string
		restPath string
		subGuid  string
	}{
		{"list", "UteamupProjectStageList", "", "{projectGuid}/stages", ""},
		{"get", "UteamupProjectStageGet", "", "{projectGuid}/stages/{stageGuid}", "stageGuid"},
		{"create", "UteamupProjectStageCreate", "", "{projectGuid}/stages", ""},
		{"update", "UteamupProjectStageUpdate", "", "{projectGuid}/stages/{stageGuid}", "stageGuid"},
		{"delete", "UteamupProjectStageDelete", "", "{projectGuid}/stages/{stageGuid}", "stageGuid"},
		{"advance", "UteamupProjectStageAdvance", "POST", "{projectGuid}/stages/{stageGuid}/advance", "stageGuid"},
		{"reorder", "UteamupProjectStageReorder", "PUT", "{projectGuid}/stages/reorder", ""},
		{"set-status", "UteamupProjectStageSetStatus", "PUT", "{projectGuid}/stages/{stageGuid}/status", "stageGuid"},
	}
	for _, c := range cases {
		a := findDomainAction(t, "project-stage", c.action)
		if a.ToolName != c.tool || a.HTTPMethod != c.method || a.RESTPath != c.restPath {
			t.Errorf("project-stage %s: want tool=%s method=%q path=%s, got tool=%s method=%q path=%s",
				c.action, c.tool, c.method, c.restPath, a.ToolName, a.HTTPMethod, a.RESTPath)
		}
		assertProjectGuidArgs(t, a, c.subGuid)
	}
}

func TestProjectStageCreateFlags(t *testing.T) {
	a := findDomainAction(t, "project-stage", "create")
	name := findFlag(a, "name")
	if name == nil || !name.Required || name.Type != "string" {
		t.Errorf("create must have a required string `name` flag, got %+v", name)
	}
	order := findFlag(a, "order")
	if order == nil || !order.Required || order.Type != "int" {
		t.Errorf("create must have a required int `order` flag, got %+v", order)
	}
	for _, optional := range []string{"gate-criteria-json", "start-date", "due-date"} {
		f := findFlag(a, optional)
		if f == nil || f.Required || f.Type != "string" {
			t.Errorf("create `%s` must be an optional string flag, got %+v", optional, f)
		}
	}
}

func TestProjectStageUpdateFlags(t *testing.T) {
	a := findDomainAction(t, "project-stage", "update")
	// PUT binds ProjectStageUpdateModel : ProjectStageCreateModel — name,
	// order, and status are non-nullable on the backend, so the CLI requires
	// them to avoid silently resetting fields on a full update.
	for _, required := range []string{"name", "order", "status"} {
		f := findFlag(a, required)
		if f == nil || !f.Required {
			t.Errorf("update `%s` flag must be required, got %+v", required, f)
		}
	}
}

func TestProjectStageReorderFlag(t *testing.T) {
	a := findDomainAction(t, "project-stage", "reorder")
	f := findFlag(a, "stage-guids")
	if f == nil || !f.Required || f.Type != "stringSlice" {
		t.Fatalf("reorder must have a required stringSlice `stage-guids` flag, got %+v", f)
	}
	// camelCase(stage-guids) = stageGuids matches the backend
	// ProjectStageReorderModel.StageGuids binding — no BodyName override needed.
	if f.BodyName != "" {
		t.Errorf("stage-guids should rely on default camelCase body name, got BodyName=%q", f.BodyName)
	}
}

func TestProjectStageSetStatusFlag(t *testing.T) {
	a := findDomainAction(t, "project-stage", "set-status")
	f := findFlag(a, "status")
	if f == nil || !f.Required || f.Type != "string" {
		t.Errorf("set-status must have a required string `status` flag, got %+v", f)
	}
}

// --- project-output ---

func TestProjectOutputDomainRegistered(t *testing.T) {
	d := findDomain("project-output")
	if d == nil {
		t.Fatal("expected project-output domain to be registered")
	}
	if d.APIPath != "/api/projects" {
		t.Errorf("project-output APIPath = %q, want /api/projects", d.APIPath)
	}
	expected := map[string]bool{"project-outputs": true, "output-items": true}
	for _, alias := range d.Aliases {
		delete(expected, alias)
	}
	if len(expected) > 0 {
		t.Errorf("missing aliases: %v", expected)
	}
}

func TestProjectOutputActionRouteTemplates(t *testing.T) {
	cases := []struct {
		action   string
		tool     string
		method   string
		restPath string
		subGuid  string
	}{
		{"list", "UteamupProjectOutputItemList", "", "{projectGuid}/outputitems", ""},
		{"get", "UteamupProjectOutputItemGet", "", "{projectGuid}/outputitems/{itemGuid}", "itemGuid"},
		{"create", "UteamupProjectOutputItemCreate", "", "{projectGuid}/outputitems", ""},
		{"update", "UteamupProjectOutputItemUpdate", "", "{projectGuid}/outputitems/{itemGuid}", "itemGuid"},
		{"delete", "UteamupProjectOutputItemDelete", "", "{projectGuid}/outputitems/{itemGuid}", "itemGuid"},
		{"deliver", "UteamupProjectOutputItemDeliver", "POST", "{projectGuid}/outputitems/{itemGuid}/deliver", "itemGuid"},
	}
	for _, c := range cases {
		a := findDomainAction(t, "project-output", c.action)
		if a.ToolName != c.tool || a.HTTPMethod != c.method || a.RESTPath != c.restPath {
			t.Errorf("project-output %s: want tool=%s method=%q path=%s, got tool=%s method=%q path=%s",
				c.action, c.tool, c.method, c.restPath, a.ToolName, a.HTTPMethod, a.RESTPath)
		}
		assertProjectGuidArgs(t, a, c.subGuid)
	}
}

func TestProjectOutputCreateFlags(t *testing.T) {
	a := findDomainAction(t, "project-output", "create")
	description := findFlag(a, "description")
	if description == nil || !description.Required || description.Type != "string" {
		t.Errorf("create must have a required string `description` flag, got %+v", description)
	}
	expected := findFlag(a, "expected-quantity")
	if expected == nil || !expected.Required || expected.Type != "float" {
		t.Errorf("create must have a required float `expected-quantity` flag, got %+v", expected)
	}
	customer := findFlag(a, "customer-guid")
	if customer == nil || customer.Required || customer.Type != "string" {
		t.Errorf("create `customer-guid` must be an optional string flag, got %+v", customer)
	}
}

func TestProjectOutputUpdateFlagDefaults(t *testing.T) {
	a := findDomainAction(t, "project-output", "update")
	actual := findFlag(a, "actual-quantity")
	if actual == nil || actual.Type != "float" {
		t.Fatalf("update must have a float `actual-quantity` flag, got %+v", actual)
	}
	// Float flag defaults MUST be float literals — an untyped int default
	// panics in the registry's `flag.Default.(float64)`-style assertions.
	if _, ok := actual.Default.(float64); !ok {
		t.Errorf("actual-quantity Default must be a float64 literal, got %T (%v)", actual.Default, actual.Default)
	}
	delivered := findFlag(a, "is-delivered")
	if delivered == nil || delivered.Type != "bool" {
		t.Fatalf("update must have a bool `is-delivered` flag, got %+v", delivered)
	}
	if _, ok := delivered.Default.(bool); !ok {
		t.Errorf("is-delivered Default must be a bool literal, got %T (%v)", delivered.Default, delivered.Default)
	}
}

func TestProjectOutputDeliverFlagHasNoDefault(t *testing.T) {
	a := findDomainAction(t, "project-output", "deliver")
	f := findFlag(a, "actual-quantity")
	if f == nil || f.Required || f.Type != "float" {
		t.Fatalf("deliver must have an optional float `actual-quantity` flag, got %+v", f)
	}
	// No Default on purpose: an omitted flag must stay out of the body so the
	// backend keeps the current actual quantity (ProjectOutputItemDeliverModel
	// binds a nullable decimal).
	if f.Default != nil {
		t.Errorf("deliver actual-quantity must have no Default (nullable on the backend), got %v", f.Default)
	}
}

// --- project-budget ---

func TestProjectBudgetDomainRegistered(t *testing.T) {
	d := findDomain("project-budget")
	if d == nil {
		t.Fatal("expected project-budget domain to be registered")
	}
	// ProjectBudgetController also routes under /api/projects (plural), so the
	// budget read cannot ride the `project` domain's /api/project base.
	if d.APIPath != "/api/projects" {
		t.Errorf("project-budget APIPath = %q, want /api/projects", d.APIPath)
	}
	expected := map[string]bool{"budget": true}
	for _, alias := range d.Aliases {
		delete(expected, alias)
	}
	if len(expected) > 0 {
		t.Errorf("missing aliases: %v", expected)
	}
}

func TestProjectBudgetGetActionWired(t *testing.T) {
	a := findDomainAction(t, "project-budget", "get")
	if a.ToolName != "UteamupProjectGetBudget" || a.HTTPMethod != "" || a.RESTPath != "{projectGuid}/budget" {
		t.Errorf("budget get must be GET {projectGuid}/budget, got %+v", a)
	}
	assertProjectGuidArgs(t, a, "")
	if len(a.Flags) != 0 {
		t.Errorf("budget get should take no flags, got %d", len(a.Flags))
	}
}
