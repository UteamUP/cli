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
	if a.ToolName != "UteamupProjectBudgetGet" || a.HTTPMethod != "" || a.RESTPath != "{projectGuid}/budget" {
		t.Errorf("budget get must be GET {projectGuid}/budget, got %+v", a)
	}
	assertProjectGuidArgs(t, a, "")
	if len(a.Flags) != 0 {
		t.Errorf("budget get should take no flags, got %d", len(a.Flags))
	}
}

// --- project-risk ---

func TestProjectRiskDomainRegistered(t *testing.T) {
	d := findDomain("project-risk")
	if d == nil {
		t.Fatal("expected project-risk domain to be registered")
	}
	// ProjectRiskController routes under /api/projects/{projectGuid}/risks —
	// plural base, so the domain needs the explicit APIPath.
	if d.APIPath != "/api/projects" {
		t.Errorf("project-risk APIPath = %q, want /api/projects", d.APIPath)
	}
	expected := map[string]bool{"project-risks": true, "risks": true}
	for _, alias := range d.Aliases {
		delete(expected, alias)
	}
	if len(expected) > 0 {
		t.Errorf("missing aliases: %v", expected)
	}
}

func TestProjectRiskActionRouteTemplates(t *testing.T) {
	cases := []struct {
		action   string
		tool     string
		method   string
		restPath string
		subGuid  string
	}{
		{"list", "UteamupProjectRiskList", "", "{projectGuid}/risks", ""},
		{"get", "UteamupProjectRiskGet", "", "{projectGuid}/risks/{riskGuid}", "riskGuid"},
		{"create", "UteamupProjectRiskCreate", "", "{projectGuid}/risks", ""},
		{"update", "UteamupProjectRiskUpdate", "", "{projectGuid}/risks/{riskGuid}", "riskGuid"},
		{"delete", "UteamupProjectRiskDelete", "", "{projectGuid}/risks/{riskGuid}", "riskGuid"},
		{"set-status", "UteamupProjectRiskSetStatus", "PUT", "{projectGuid}/risks/{riskGuid}/status", "riskGuid"},
	}
	for _, c := range cases {
		a := findDomainAction(t, "project-risk", c.action)
		if a.ToolName != c.tool || a.HTTPMethod != c.method || a.RESTPath != c.restPath {
			t.Errorf("project-risk %s: want tool=%s method=%q path=%s, got tool=%s method=%q path=%s",
				c.action, c.tool, c.method, c.restPath, a.ToolName, a.HTTPMethod, a.RESTPath)
		}
		assertProjectGuidArgs(t, a, c.subGuid)
	}
}

func TestProjectRiskListStatusFilterOptional(t *testing.T) {
	a := findDomainAction(t, "project-risk", "list")
	f := findFlag(a, "status")
	// GET flags ride the query string; the backend binds a nullable
	// ProjectRiskStatus, so the filter must stay optional with no default.
	if f == nil || f.Required || f.Type != "string" || f.Default != nil {
		t.Errorf("list `status` must be an optional string flag with no default, got %+v", f)
	}
}

func TestProjectRiskCreateFlags(t *testing.T) {
	a := findDomainAction(t, "project-risk", "create")
	title := findFlag(a, "title")
	if title == nil || !title.Required || title.Type != "string" {
		t.Errorf("create must have a required string `title` flag, got %+v", title)
	}
	for _, required := range []string{"probability", "impact"} {
		f := findFlag(a, required)
		if f == nil || !f.Required || f.Type != "int" {
			t.Errorf("create `%s` must be a required int flag, got %+v", required, f)
		}
	}
	for _, optional := range []string{"description", "category", "mitigation-plan", "owner-guid", "review-date"} {
		f := findFlag(a, optional)
		if f == nil || f.Required || f.Type != "string" {
			t.Errorf("create `%s` must be an optional string flag, got %+v", optional, f)
		}
	}
	if status := findFlag(a, "status"); status != nil {
		t.Errorf("create must not expose a `status` flag (ProjectRiskCreateModel has no Status), got %+v", status)
	}
}

func TestProjectRiskUpdateFlags(t *testing.T) {
	a := findDomainAction(t, "project-risk", "update")
	// PUT binds ProjectRiskUpdateModel : ProjectRiskCreateModel — title,
	// probability, impact, status and category are required so a full update
	// never silently resets a field (omitted category re-defaults to "Other").
	for _, required := range []string{"title", "category", "probability", "impact", "status"} {
		f := findFlag(a, required)
		if f == nil || !f.Required {
			t.Errorf("update `%s` flag must be required, got %+v", required, f)
		}
	}
}

func TestProjectRiskSetStatusFlag(t *testing.T) {
	a := findDomainAction(t, "project-risk", "set-status")
	f := findFlag(a, "status")
	if f == nil || !f.Required || f.Type != "string" {
		t.Errorf("set-status must have a required string `status` flag, got %+v", f)
	}
}

// --- project-insights ---

func TestProjectInsightsDomainRegistered(t *testing.T) {
	d := findDomain("project-insights")
	if d == nil {
		t.Fatal("expected project-insights domain to be registered")
	}
	// ProjectInsightsController routes under /api/projects (plural).
	if d.APIPath != "/api/projects" {
		t.Errorf("project-insights APIPath = %q, want /api/projects", d.APIPath)
	}
	expected := map[string]bool{"insights": true}
	for _, alias := range d.Aliases {
		delete(expected, alias)
	}
	if len(expected) > 0 {
		t.Errorf("missing aliases: %v", expected)
	}
}

func TestProjectInsightsConflictsActionWired(t *testing.T) {
	a := findDomainAction(t, "project-insights", "conflicts")
	// Method "" — `conflicts` is not in the HTTPMethod map and has no
	// update- prefix, so the runtime falls back to GET.
	if a.ToolName != "UteamupProjectGetConflicts" || a.HTTPMethod != "" || a.RESTPath != "{projectGuid}/conflicts" {
		t.Errorf("conflicts must be GET {projectGuid}/conflicts, got %+v", a)
	}
	assertProjectGuidArgs(t, a, "")
	if len(a.Flags) != 0 {
		t.Errorf("conflicts should take no flags, got %d", len(a.Flags))
	}
}

func TestProjectInsightsPortfolioActionWired(t *testing.T) {
	a := findDomainAction(t, "project-insights", "portfolio")
	if a.ToolName != "UteamupProjectGetPortfolio" || a.HTTPMethod != "" || a.RESTPath != "portfolio" {
		t.Errorf("portfolio must be GET portfolio, got %+v", a)
	}
	if len(a.Args) != 0 {
		t.Errorf("portfolio is tenant-scoped and takes no positional args, got %+v", a.Args)
	}
	// camelCase(page-size) = pageSize matches the backend [FromQuery] binding.
	for _, name := range []string{"page", "page-size"} {
		f := findFlag(a, name)
		if f == nil || f.Type != "int" {
			t.Errorf("portfolio must have an int `%s` pagination flag, got %+v", name, f)
		}
	}
	status := findFlag(a, "status")
	// Optional int with no default — omitted must stay off the query string so
	// the backend's nullable status filter stays null.
	if status == nil || status.Required || status.Type != "int" || status.Default != nil {
		t.Errorf("portfolio `status` must be an optional int flag with no default, got %+v", status)
	}
}

// --- cost-budget-threshold ---

func TestCostBudgetThresholdDomainRegistered(t *testing.T) {
	d := findDomain("cost-budget-threshold")
	if d == nil {
		t.Fatal("expected cost-budget-threshold domain to be registered")
	}
	// CostBudgetThresholdController is tenant-scoped under its own base path.
	if d.APIPath != "/api/costbudgetthresholds" {
		t.Errorf("cost-budget-threshold APIPath = %q, want /api/costbudgetthresholds", d.APIPath)
	}
	expected := map[string]bool{"cost-budget-thresholds": true, "budget-thresholds": true}
	for _, alias := range d.Aliases {
		delete(expected, alias)
	}
	if len(expected) > 0 {
		t.Errorf("missing aliases: %v", expected)
	}
}

func TestCostBudgetThresholdActionRouteTemplates(t *testing.T) {
	cases := []struct {
		action   string
		tool     string
		restPath string
		hasGuid  bool
	}{
		{"list", "UteamupCostBudgetThresholdList", "", false},
		{"create", "UteamupCostBudgetThresholdCreate", "", false},
		{"update", "UteamupCostBudgetThresholdUpdate", "{thresholdGuid}", true},
		{"delete", "UteamupCostBudgetThresholdDelete", "{thresholdGuid}", true},
	}
	for _, c := range cases {
		a := findDomainAction(t, "cost-budget-threshold", c.action)
		// Method derived from the action name (list/create/update/delete).
		if a.ToolName != c.tool || a.HTTPMethod != "" || a.RESTPath != c.restPath {
			t.Errorf("cost-budget-threshold %s: want tool=%s method=%q path=%q, got tool=%s method=%q path=%q",
				c.action, c.tool, "", c.restPath, a.ToolName, a.HTTPMethod, a.RESTPath)
		}
		if !c.hasGuid {
			if len(a.Args) != 0 {
				t.Errorf("%s should take no positional args, got %+v", c.action, a.Args)
			}
			continue
		}
		if len(a.Args) != 1 || a.Args[0].Name != "thresholdGuid" || !a.Args[0].Required || a.Args[0].Type != "string" {
			t.Errorf("%s expected single required string positional arg 'thresholdGuid', got %+v", c.action, a.Args)
		}
	}
}

// --- project-bom ---

func TestProjectBomDomainRegistered(t *testing.T) {
	d := findDomain("project-bom")
	if d == nil {
		t.Fatal("expected project-bom domain to be registered")
	}
	// ProjectBomController routes under /api/projects/{projectGuid}/bom —
	// plural base, so the domain needs the explicit APIPath.
	if d.APIPath != "/api/projects" {
		t.Errorf("project-bom APIPath = %q, want /api/projects", d.APIPath)
	}
	expected := map[string]bool{"project-boms": true, "bom": true}
	for _, alias := range d.Aliases {
		delete(expected, alias)
	}
	if len(expected) > 0 {
		t.Errorf("missing aliases: %v", expected)
	}
}

func TestProjectBomActionRouteTemplates(t *testing.T) {
	cases := []struct {
		action   string
		tool     string
		method   string
		restPath string
		subGuid  string
	}{
		{"list", "UteamupProjectBomList", "", "{projectGuid}/bom", ""},
		{"get", "UteamupProjectBomGet", "", "{projectGuid}/bom/{itemGuid}", "itemGuid"},
		{"create", "UteamupProjectBomCreate", "", "{projectGuid}/bom", ""},
		{"update", "UteamupProjectBomUpdate", "", "{projectGuid}/bom/{itemGuid}", "itemGuid"},
		{"delete", "UteamupProjectBomDelete", "", "{projectGuid}/bom/{itemGuid}", "itemGuid"},
	}
	for _, c := range cases {
		a := findDomainAction(t, "project-bom", c.action)
		if a.ToolName != c.tool || a.HTTPMethod != c.method || a.RESTPath != c.restPath {
			t.Errorf("project-bom %s: want tool=%s method=%q path=%s, got tool=%s method=%q path=%s",
				c.action, c.tool, c.method, c.restPath, a.ToolName, a.HTTPMethod, a.RESTPath)
		}
		assertProjectGuidArgs(t, a, c.subGuid)
	}
}

func TestProjectBomCreateFlags(t *testing.T) {
	a := findDomainAction(t, "project-bom", "create")
	itemType := findFlag(a, "item-type")
	if itemType == nil || !itemType.Required || itemType.Type != "string" {
		t.Errorf("create must have a required string `item-type` flag, got %+v", itemType)
	}
	itemGuid := findFlag(a, "item-guid")
	if itemGuid == nil || !itemGuid.Required || itemGuid.Type != "string" {
		t.Errorf("create must have a required string `item-guid` flag, got %+v", itemGuid)
	}
	qty := findFlag(a, "quantity-required")
	if qty == nil || !qty.Required || qty.Type != "float" {
		t.Errorf("create must have a required float `quantity-required` flag, got %+v", qty)
	}
	// Create binds ProjectBomItemCreateModel — no quantityActual/isConsumed.
	if extra := findFlag(a, "quantity-actual"); extra != nil {
		t.Errorf("create must not expose a `quantity-actual` flag, got %+v", extra)
	}
	if extra := findFlag(a, "is-consumed"); extra != nil {
		t.Errorf("create must not expose an `is-consumed` flag, got %+v", extra)
	}
}

func TestProjectBomUpdateFlagDefaults(t *testing.T) {
	a := findDomainAction(t, "project-bom", "update")
	// PUT binds ProjectBomItemUpdateModel : ProjectBomItemCreateModel — the
	// create-base fields stay required on a full update so nothing silently resets.
	for _, required := range []string{"item-type", "item-guid", "quantity-required"} {
		f := findFlag(a, required)
		if f == nil || !f.Required {
			t.Errorf("update `%s` flag must be required, got %+v", required, f)
		}
	}
	// Float/bool flag defaults MUST be typed literals — an untyped int default
	// panics in the registry's `.(float64)`-style assertions.
	actual := findFlag(a, "quantity-actual")
	if actual == nil || actual.Type != "float" {
		t.Fatalf("update must have a float `quantity-actual` flag, got %+v", actual)
	}
	if _, ok := actual.Default.(float64); !ok {
		t.Errorf("quantity-actual Default must be a float64 literal, got %T (%v)", actual.Default, actual.Default)
	}
	consumed := findFlag(a, "is-consumed")
	if consumed == nil || consumed.Type != "bool" {
		t.Fatalf("update must have a bool `is-consumed` flag, got %+v", consumed)
	}
	if _, ok := consumed.Default.(bool); !ok {
		t.Errorf("is-consumed Default must be a bool literal, got %T (%v)", consumed.Default, consumed.Default)
	}
}

func TestCostBudgetThresholdCreateUpdateFlags(t *testing.T) {
	for _, action := range []string{"create", "update"} {
		a := findDomainAction(t, "cost-budget-threshold", action)
		name := findFlag(a, "name")
		if name == nil || !name.Required || name.Type != "string" {
			t.Errorf("%s must have a required string `name` flag, got %+v", action, name)
		}
		pct := findFlag(a, "threshold-percentage")
		if pct == nil || !pct.Required || pct.Type != "float" {
			t.Errorf("%s must have a required float `threshold-percentage` flag, got %+v", action, pct)
		}
		// Defaults mirror the backend DTO initializers so the body stays
		// deterministic; string/bool defaults must be typed literals.
		entityType := findFlag(a, "entity-type")
		if entityType == nil || entityType.Type != "string" || entityType.Default != "Project" {
			t.Errorf("%s `entity-type` must default to \"Project\", got %+v", action, entityType)
		}
		severity := findFlag(a, "severity")
		if severity == nil || severity.Type != "string" || severity.Default != "Warning" {
			t.Errorf("%s `severity` must default to \"Warning\", got %+v", action, severity)
		}
		active := findFlag(a, "is-active")
		if active == nil || active.Type != "bool" {
			t.Fatalf("%s must have a bool `is-active` flag, got %+v", action, active)
		}
		if v, ok := active.Default.(bool); !ok || !v {
			t.Errorf("%s `is-active` Default must be the bool literal true, got %T (%v)", action, active.Default, active.Default)
		}
	}
}

// --- work-order link actions on project-stage / project-risk ---

func TestProjectStageRiskWorkorderLinkActions(t *testing.T) {
	cases := []struct {
		domain   string
		action   string
		tool     string
		method   string
		restPath string
		args     []string
	}{
		{"project-stage", "assign-workorder", "UteamupProjectStageAssignWorkorder", "PUT", "{projectGuid}/stages/{stageGuid}/workorders/{workorderGuid}", []string{"projectGuid", "stageGuid", "workorderGuid"}},
		{"project-stage", "unassign-workorder", "UteamupProjectStageUnassignWorkorder", "DELETE", "{projectGuid}/stages/{stageGuid}/workorders/{workorderGuid}", []string{"projectGuid", "stageGuid", "workorderGuid"}},
		{"project-risk", "list-workorders", "UteamupProjectRiskListWorkorders", "", "{projectGuid}/risks/{riskGuid}/workorders", []string{"projectGuid", "riskGuid"}},
		{"project-risk", "link-workorder", "UteamupProjectRiskLinkWorkorder", "POST", "{projectGuid}/risks/{riskGuid}/workorders/{workorderGuid}", []string{"projectGuid", "riskGuid", "workorderGuid"}},
		{"project-risk", "unlink-workorder", "UteamupProjectRiskUnlinkWorkorder", "DELETE", "{projectGuid}/risks/{riskGuid}/workorders/{workorderGuid}", []string{"projectGuid", "riskGuid", "workorderGuid"}},
	}
	for _, c := range cases {
		a := findDomainAction(t, c.domain, c.action)
		if a.ToolName != c.tool || a.HTTPMethod != c.method || a.RESTPath != c.restPath {
			t.Errorf("%s %s: want tool=%s method=%q path=%s, got tool=%s method=%q path=%s",
				c.domain, c.action, c.tool, c.method, c.restPath, a.ToolName, a.HTTPMethod, a.RESTPath)
		}
		if len(a.Args) != len(c.args) {
			t.Fatalf("%s %s expected %d args, got %+v", c.domain, c.action, len(c.args), a.Args)
		}
		for i, name := range c.args {
			if a.Args[i].Name != name || !a.Args[i].Required || a.Args[i].Type != "string" {
				t.Errorf("%s %s arg[%d] must be required string %q, got %+v", c.domain, c.action, i, name, a.Args[i])
			}
		}
	}
}
