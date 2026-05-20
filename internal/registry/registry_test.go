package registry

import (
	"reflect"
	"sort"
	"testing"
)

// TestBuildRESTPathUpdateSubRoutes locks in the update-<sub> sub-route convention:
// `update-status` → /{id}/status (explicit case), `update-notes` → /{id}/notes
// (generic fallback). Without these, PATCH endpoints route to the basePath and
// produce 405/404 from the backend.
func TestBuildRESTPathUpdateSubRoutes(t *testing.T) {
	domain := &Domain{Name: "bugsandfeatures", APIPath: "/api/bugsandfeatures"}
	cases := []struct {
		actionName string
		argKey     string
		argValue   any
		want       string
	}{
		{"update-status", "externalGuid", "abc-123", "/api/bugsandfeatures/abc-123/status"},
		{"update-notes", "externalGuid", "abc-123", "/api/bugsandfeatures/abc-123/notes"},
		{"update-status", "id", 42, "/api/bugsandfeatures/42/status"},
		{"update-priority", "externalGuid", "g1", "/api/bugsandfeatures/g1/priority"},
		{"get", "externalGuid", "g1", "/api/bugsandfeatures/g1"},
	}
	for _, tc := range cases {
		t.Run(tc.actionName, func(t *testing.T) {
			got, _ := buildRESTPath(domain, Action{Name: tc.actionName}, map[string]any{tc.argKey: tc.argValue})
			if got != tc.want {
				t.Errorf("buildRESTPath(%s) = %q, want %q", tc.actionName, got, tc.want)
			}
		})
	}
}

func TestHTTPMethodForUpdateNotes(t *testing.T) {
	if HTTPMethod["update-notes"] != "PATCH" {
		t.Errorf("update-notes HTTPMethod = %q, want PATCH", HTTPMethod["update-notes"])
	}
}

// TestBuildRESTPathTemplate locks in the sub-resource path-template routing
// for verbs that don't fit the standard CRUD pattern. Without it, the CLI's
// `comments-list` / `comments-add` / `attachments-list` / `attachments-delete`
// verbs fall through to the basePath and produce wrong-shape requests (the
// `comments-add` regression that surfaced when posting to bug
// c6ec7720-… on prod sent the body in the query string of a GET).
func TestBuildRESTPathTemplate(t *testing.T) {
	domain := &Domain{Name: "bugsandfeatures", APIPath: "/api/bugsandfeatures"}
	cases := []struct {
		name         string
		action       Action
		args         map[string]any
		wantPath     string
		wantConsumed []string
	}{
		{
			name:         "comments-list expands {bugExternalGuid}",
			action:       Action{Name: "comments-list", RESTPath: "{bugExternalGuid}/comments"},
			args:         map[string]any{"bugExternalGuid": "g1"},
			wantPath:     "/api/bugsandfeatures/g1/comments",
			wantConsumed: []string{"bugExternalGuid"},
		},
		{
			name:         "comments-add expands {bugExternalGuid}",
			action:       Action{Name: "comments-add", RESTPath: "{bugExternalGuid}/comments"},
			args:         map[string]any{"bugExternalGuid": "g1", "bodyHtml": "hi"},
			wantPath:     "/api/bugsandfeatures/g1/comments",
			wantConsumed: []string{"bugExternalGuid"},
		},
		{
			name:         "attachments-delete expands two placeholders",
			action:       Action{Name: "attachments-delete", RESTPath: "{bugExternalGuid}/attachments/{attachmentExternalGuid}"},
			args:         map[string]any{"bugExternalGuid": "g1", "attachmentExternalGuid": "a1"},
			wantPath:     "/api/bugsandfeatures/g1/attachments/a1",
			wantConsumed: []string{"bugExternalGuid", "attachmentExternalGuid"},
		},
		{
			name:     "literal RESTPath without placeholders is preserved (legacy)",
			action:   Action{Name: "list", RESTPath: "all"},
			args:     map[string]any{},
			wantPath: "/api/bugsandfeatures/all",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotPath, gotConsumed := buildRESTPath(domain, tc.action, tc.args)
			if gotPath != tc.wantPath {
				t.Errorf("path = %q, want %q", gotPath, tc.wantPath)
			}
			if tc.wantConsumed != nil {
				sort.Strings(gotConsumed)
				sort.Strings(tc.wantConsumed)
				if !reflect.DeepEqual(gotConsumed, tc.wantConsumed) {
					t.Errorf("consumed = %v, want %v", gotConsumed, tc.wantConsumed)
				}
			}
		})
	}
}

// TestExpandPathTemplateUnknownPlaceholder verifies that an unknown placeholder
// stops expansion at that point so callers see the raw token rather than a
// silently-wrong URL.
func TestExpandPathTemplateUnknownPlaceholder(t *testing.T) {
	got, consumed := expandPathTemplate("{a}/x/{b}", map[string]any{"a": "1"})
	want := "1/x/{b}"
	if got != want {
		t.Errorf("expandPathTemplate = %q, want %q", got, want)
	}
	if !reflect.DeepEqual(consumed, []string{"a"}) {
		t.Errorf("consumed = %v, want [a]", consumed)
	}
}

// TestBugsAndFeaturesCommentsAddDeclaration locks in the registry declaration
// for `comments-add` so a future refactor can't silently regress the
// HTTPMethod / RESTPath / BodyName fields. This is the verb the user hit on
// the c6ec7720 prod bug — we need it to stay POST + path-templated + body-
// renamed forever.
func TestBugsAndFeaturesCommentsAddDeclaration(t *testing.T) {
	var domain *Domain
	for _, d := range DefaultRegistry.Domains() {
		if d.Name == "bugsandfeatures" {
			domain = d
			break
		}
	}
	if domain == nil {
		t.Fatal("bugsandfeatures domain not registered")
	}
	var addAction *Action
	for i := range domain.Actions {
		if domain.Actions[i].Name == "comments-add" {
			addAction = &domain.Actions[i]
			break
		}
	}
	if addAction == nil {
		t.Fatal("comments-add action not registered")
	}
	if addAction.HTTPMethod != "POST" {
		t.Errorf("HTTPMethod = %q, want POST", addAction.HTTPMethod)
	}
	if addAction.RESTPath != "{bugExternalGuid}/comments" {
		t.Errorf("RESTPath = %q, want {bugExternalGuid}/comments", addAction.RESTPath)
	}
	wantBodyNames := map[string]string{
		"text":    "bodyHtml",
		"parent":  "parentCommentExternalGuid",
		"mention": "mentionedGlobalAdminGuids",
	}
	for _, f := range addAction.Flags {
		if want, ok := wantBodyNames[f.Name]; ok {
			if f.BodyName != want {
				t.Errorf("flag %q BodyName = %q, want %q", f.Name, f.BodyName, want)
			}
			delete(wantBodyNames, f.Name)
		}
	}
	for missing := range wantBodyNames {
		t.Errorf("flag %q not declared on comments-add", missing)
	}
}

// TestBugsAndFeaturesCreateIdempotencyKeyHeader locks in that the
// idempotency-key flag on `bugs create` is routed via HeaderName, not BodyName.
// The backend reads `[FromHeader(Name = "Idempotency-Key")]` and rejects bodies
// that try to carry it as `idempotencyKey`. Regression guard for CLI-1
// (bug 14074b1d-abbc-4ecf-9d10-444de997255b).
func TestBugsAndFeaturesCreateIdempotencyKeyHeader(t *testing.T) {
	var domain *Domain
	for _, d := range DefaultRegistry.Domains() {
		if d.Name == "bugsandfeatures" {
			domain = d
			break
		}
	}
	if domain == nil {
		t.Fatal("bugsandfeatures domain not registered")
	}
	var createAction *Action
	for i := range domain.Actions {
		if domain.Actions[i].Name == "create" {
			createAction = &domain.Actions[i]
			break
		}
	}
	if createAction == nil {
		t.Fatal("create action not registered")
	}
	var idem *FlagDef
	for i := range createAction.Flags {
		if createAction.Flags[i].Name == "idempotency-key" {
			idem = &createAction.Flags[i]
			break
		}
	}
	if idem == nil {
		t.Fatal("idempotency-key flag not declared on bugs create")
	}
	if idem.HeaderName != "Idempotency-Key" {
		t.Errorf("idempotency-key HeaderName = %q, want %q", idem.HeaderName, "Idempotency-Key")
	}
	if idem.BodyName != "" {
		t.Errorf("idempotency-key BodyName = %q, want empty (header-only flag)", idem.BodyName)
	}
}

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"page", "page"},
		{"page-size", "pageSize"},
		{"sort-by", "sortBy"},
		{"sort-order", "sortOrder"},
		{"from-json", "fromJson"},
		{"asset-type-id", "assetTypeId"},
		{"a", "a"},
		{"", ""},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := toCamelCase(tc.input)
			if result != tc.expected {
				t.Errorf("toCamelCase(%q) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestDefaultRegistryHasDomains(t *testing.T) {
	domains := DefaultRegistry.Domains()
	if len(domains) == 0 {
		t.Fatal("expected at least one registered domain")
	}

	// Check that our 3 starter domains are registered
	domainNames := make(map[string]bool)
	for _, d := range domains {
		domainNames[d.Name] = true
	}

	expected := []string{"asset", "workorder", "user"}
	for _, name := range expected {
		if !domainNames[name] {
			t.Errorf("expected domain %q to be registered", name)
		}
	}
}

func TestAssetDomainActions(t *testing.T) {
	var assetDomain *Domain
	for _, d := range DefaultRegistry.Domains() {
		if d.Name == "asset" {
			assetDomain = d
			break
		}
	}
	if assetDomain == nil {
		t.Fatal("asset domain not found")
	}

	expectedActions := []string{"list", "get", "get-by-guid", "get-assigned-stock", "create", "update", "delete", "search"}
	actionNames := make(map[string]bool)
	actionByName := make(map[string]Action)
	for _, a := range assetDomain.Actions {
		actionNames[a.Name] = true
		actionByName[a.Name] = a
	}

	for _, name := range expectedActions {
		if !actionNames[name] {
			t.Errorf("expected action %q in asset domain", name)
		}
	}

	// New endpoint must mirror the backend MCP tool name + accept assetGuid.
	if got := actionByName["get-assigned-stock"]; got.ToolName != "UteamupAssetGetAssignedStock" {
		t.Errorf("get-assigned-stock ToolName = %q, want UteamupAssetGetAssignedStock", got.ToolName)
	}
	if got := actionByName["get-assigned-stock"]; len(got.Args) != 1 || got.Args[0].Name != "assetGuid" {
		t.Errorf("get-assigned-stock should take a single required assetGuid arg, got %+v", got.Args)
	}
}

func TestWorkorderDomainAliases(t *testing.T) {
	var woDomain *Domain
	for _, d := range DefaultRegistry.Domains() {
		if d.Name == "workorder" {
			woDomain = d
			break
		}
	}
	if woDomain == nil {
		t.Fatal("workorder domain not found")
	}

	hasWO := false
	for _, alias := range woDomain.Aliases {
		if alias == "wo" {
			hasWO = true
		}
	}
	if !hasWO {
		t.Error("workorder domain should have 'wo' alias")
	}
}

func TestDomainToolNames(t *testing.T) {
	for _, d := range DefaultRegistry.Domains() {
		for _, a := range d.Actions {
			if a.ToolName == "" {
				t.Errorf("domain %s action %s has empty ToolName", d.Name, a.Name)
			}
		}
	}
}
