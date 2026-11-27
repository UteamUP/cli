package registry

import (
	"errors"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/uteamup/cli/internal/client"
	"github.com/uteamup/cli/internal/logging"
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
		{"update-severity", "externalGuid", "abc-123", "/api/bugsandfeatures/abc-123/severity"},
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

func TestRedactSensitiveActionValues(t *testing.T) {
	action := Action{Flags: []FlagDef{
		{Name: "code", BodyName: "numericCode", Sensitive: true},
		{Name: "request-key", HeaderName: "X-Request-Key", Sensitive: true},
		{Name: "note"},
	}}
	args := map[string]any{
		"numericCode": "12345678",
		"note":        "handover review",
	}
	headers := map[string]string{"X-Request-Key": "secret-header"}

	redactedArgs, redactedHeaders := redactSensitiveActionValues(action, args, headers)

	if redactedArgs["numericCode"] != "[REDACTED]" {
		t.Errorf("sensitive body value = %v, want [REDACTED]", redactedArgs["numericCode"])
	}
	if redactedArgs["note"] != "handover review" {
		t.Errorf("ordinary body value changed: %v", redactedArgs["note"])
	}
	if redactedHeaders["X-Request-Key"] != "[REDACTED]" {
		t.Errorf("sensitive header value = %q, want [REDACTED]", redactedHeaders["X-Request-Key"])
	}
	if args["numericCode"] != "12345678" || headers["X-Request-Key"] != "secret-header" {
		t.Error("redaction mutated the outgoing request values")
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
			name:         "ping-reporter expands {bugExternalGuid}",
			action:       Action{Name: "ping-reporter", RESTPath: "{bugExternalGuid}/ping-reporter"},
			args:         map[string]any{"bugExternalGuid": "g1"},
			wantPath:     "/api/bugsandfeatures/g1/ping-reporter",
			wantConsumed: []string{"bugExternalGuid"},
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

// TestBugsAndFeaturesAttachmentsUploadDeclaration locks in the registry
// declaration for `attachments-upload`: POST to the {bugExternalGuid}/attachments
// sub-route with the `file` flag marked UploadFile so the CLI sends a real
// multipart upload. Without HTTPMethod/RESTPath the verb falls through to the
// basePath, and without UploadFile the file path is sent as a plain string —
// both silently break the upload. Regression guard against a refactor dropping
// any of the three fields.
func TestBugsAndFeaturesAttachmentsUploadDeclaration(t *testing.T) {
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
	var uploadAction *Action
	for i := range domain.Actions {
		if domain.Actions[i].Name == "attachments-upload" {
			uploadAction = &domain.Actions[i]
			break
		}
	}
	if uploadAction == nil {
		t.Fatal("attachments-upload action not registered")
	}
	if uploadAction.HTTPMethod != "POST" {
		t.Errorf("HTTPMethod = %q, want POST", uploadAction.HTTPMethod)
	}
	if uploadAction.RESTPath != "{bugExternalGuid}/attachments" {
		t.Errorf("RESTPath = %q, want {bugExternalGuid}/attachments", uploadAction.RESTPath)
	}
	var file *FlagDef
	for i := range uploadAction.Flags {
		if uploadAction.Flags[i].Name == "file" {
			file = &uploadAction.Flags[i]
			break
		}
	}
	if file == nil {
		t.Fatal("file flag not declared on attachments-upload")
	}
	if !file.UploadFile {
		t.Error("file flag UploadFile = false, want true (multipart upload)")
	}
}

// TestBugsAndFeaturesReporterConversation locks in the reporter-conversation registry surface:
// the `--share-with-reporter` flag on comments-add (BodyName isVisibleToReporter, bool) and the
// GET `conversation` read action. Keeps the CLI in lockstep with the backend reporter endpoints.
func TestBugsAndFeaturesReporterConversation(t *testing.T) {
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

	var addAction, convAction *Action
	for i := range domain.Actions {
		switch domain.Actions[i].Name {
		case "comments-add":
			addAction = &domain.Actions[i]
		case "conversation":
			convAction = &domain.Actions[i]
		}
	}

	if addAction == nil {
		t.Fatal("comments-add action not registered")
	}
	var share *FlagDef
	for i := range addAction.Flags {
		if addAction.Flags[i].Name == "share-with-reporter" {
			share = &addAction.Flags[i]
			break
		}
	}
	if share == nil {
		t.Fatal("share-with-reporter flag not declared on comments-add")
	}
	if share.BodyName != "isVisibleToReporter" {
		t.Errorf("share-with-reporter BodyName = %q, want isVisibleToReporter", share.BodyName)
	}
	if share.Type != "bool" {
		t.Errorf("share-with-reporter Type = %q, want bool", share.Type)
	}

	if convAction == nil {
		t.Fatal("conversation action not registered")
	}
	if convAction.HTTPMethod != "GET" {
		t.Errorf("conversation HTTPMethod = %q, want GET", convAction.HTTPMethod)
	}
	if convAction.RESTPath != "{bugExternalGuid}/conversation" {
		t.Errorf("conversation RESTPath = %q, want {bugExternalGuid}/conversation", convAction.RESTPath)
	}
}

// TestBugsAndFeaturesMine locks in the `mine` read action that powers the
// reporter "My reports" page: GET against the static `mine` suffix, tool
// uteamup_bug_report_list_mine, no positional args. Keeps the CLI in lockstep
// with the backend BugsAndFeaturesController `[HttpGet("mine")]` endpoint.
func TestBugsAndFeaturesMine(t *testing.T) {
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

	var mineAction *Action
	for i := range domain.Actions {
		if domain.Actions[i].Name == "mine" {
			mineAction = &domain.Actions[i]
			break
		}
	}
	if mineAction == nil {
		t.Fatal("mine action not registered")
	}
	if mineAction.ToolName != "uteamup_bug_report_list_mine" {
		t.Errorf("mine ToolName = %q, want uteamup_bug_report_list_mine", mineAction.ToolName)
	}
	if mineAction.HTTPMethod != "GET" {
		t.Errorf("mine HTTPMethod = %q, want GET", mineAction.HTTPMethod)
	}
	if mineAction.RESTPath != "mine" {
		t.Errorf("mine RESTPath = %q, want mine", mineAction.RESTPath)
	}
	if len(mineAction.Args) != 0 {
		t.Errorf("mine should take no positional args, got %d", len(mineAction.Args))
	}
}

// TestBugsAndFeaturesUpdateSeverityDeclaration locks in the registry declaration
// for `update-severity`: a REST PATCH to the {externalGuid}/severity sub-route
// (generic update-<sub> rule), tool uteamup_bugsandfeatures_update_severity, a
// required --severity flag constrained to the four backend enum values, and an
// optional --note. Without the AllowedValues constraint a free-form severity
// would reach the backend and 400; without the update- prefix the verb would
// fall through to the basePath and 404/405.
func TestBugsAndFeaturesUpdateSeverityDeclaration(t *testing.T) {
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
	var action *Action
	for i := range domain.Actions {
		if domain.Actions[i].Name == "update-severity" {
			action = &domain.Actions[i]
			break
		}
	}
	if action == nil {
		t.Fatal("update-severity action not registered")
	}
	if action.MCPOnly {
		t.Error("update-severity must use the REST severity endpoint, not MCPOnly")
	}
	if action.ToolName != "uteamup_bugsandfeatures_update_severity" {
		t.Errorf("ToolName = %q, want uteamup_bugsandfeatures_update_severity", action.ToolName)
	}
	if len(action.Args) != 1 || action.Args[0].Name != "externalGuid" || !action.Args[0].Required {
		t.Fatalf("update-severity must take a single required externalGuid arg, got %+v", action.Args)
	}

	var severity, note *FlagDef
	for i := range action.Flags {
		switch action.Flags[i].Name {
		case "severity":
			severity = &action.Flags[i]
		case "note":
			note = &action.Flags[i]
		}
	}
	if severity == nil {
		t.Fatal("severity flag not declared on update-severity")
	}
	if !severity.Required || severity.Type != "string" {
		t.Errorf("severity flag must be a required string, got required=%v type=%q", severity.Required, severity.Type)
	}
	want := []string{"Low", "Medium", "High", "Critical"}
	if !reflect.DeepEqual(severity.AllowedValues, want) {
		t.Errorf("severity AllowedValues = %v, want %v", severity.AllowedValues, want)
	}
	if severity.BodyName != "" {
		t.Errorf("severity BodyName = %q, want empty (default camelCase severity)", severity.BodyName)
	}
	if note == nil {
		t.Fatal("note flag not declared on update-severity")
	}
	if note.Required {
		t.Error("note flag must be optional")
	}

	if err := validateActionDefinition(*action); err != nil {
		t.Fatalf("validateActionDefinition: %v", err)
	}
	path, consumed := buildRESTPath(domain, *action, map[string]any{"externalGuid": "g1", "severity": "Critical"})
	if path != "/api/bugsandfeatures/g1/severity" {
		t.Errorf("path = %q, want /api/bugsandfeatures/g1/severity", path)
	}
	if !reflect.DeepEqual(consumed, []string{"externalGuid"}) {
		t.Errorf("consumed = %v, want [externalGuid] (guid must not double-leak into the body)", consumed)
	}
}

// TestBugsAndFeaturesUpdateSeverityValidation proves the --severity enum gate
// runs before any transport: an undeclared value is rejected locally, while a
// declared value reaches the API client factory. Mirrors the MIME-enum guard.
func TestBugsAndFeaturesUpdateSeverityValidation(t *testing.T) {
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
	var action Action
	found := false
	for _, a := range domain.Actions {
		if a.Name == "update-severity" {
			action = a
			found = true
			break
		}
	}
	if !found {
		t.Fatal("update-severity action not registered")
	}

	for _, value := range []string{"Low", "Medium", "High", "Critical", "Urgent", "critical"} {
		t.Run(value, func(t *testing.T) {
			reached := errors.New("transport reached")
			called := false
			format := "json"
			command := buildActionCommand(domain, action, func() (*client.APIClient, error) {
				called = true
				return nil, reached
			}, logging.New(logging.LevelError), &format, nil)
			command.SetArgs([]string{"11111111-1111-4111-8111-111111111111", "--severity", value})
			err := command.Execute()
			valid := value == "Low" || value == "Medium" || value == "High" || value == "Critical"
			if valid {
				if !called || !errors.Is(err, reached) {
					t.Fatalf("declared severity %q was blocked: called=%v err=%v", value, called, err)
				}
			} else {
				if called || err == nil || !strings.Contains(err.Error(), "must be one of") {
					t.Fatalf("undeclared severity %q reached transport: called=%v err=%v", value, called, err)
				}
			}
		})
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

	expectedActions := []string{"list", "get", "get-by-guid", "duplicate", "get-assigned-stock", "get-documents-aggregated", "create", "update", "delete", "search"}
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

	// Aggregated-documents endpoint must mirror the backend MCP tool name + accept assetGuid.
	if got := actionByName["get-documents-aggregated"]; got.ToolName != "UteamupAssetGetAggregatedDocuments" {
		t.Errorf("get-documents-aggregated ToolName = %q, want UteamupAssetGetAggregatedDocuments", got.ToolName)
	}
	if got := actionByName["get-documents-aggregated"]; len(got.Args) != 1 || got.Args[0].Name != "assetGuid" {
		t.Errorf("get-documents-aggregated should take a single required assetGuid arg, got %+v", got.Args)
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

func TestActionFlagsDoNotReusePersistentShorthands(t *testing.T) {
	reserved := map[string]string{
		"o": "output",
		"P": "profile",
		"v": "verbose",
	}

	for _, domain := range DefaultRegistry.Domains() {
		for _, action := range domain.Actions {
			for _, flag := range action.Flags {
				if persistentName, conflict := reserved[flag.Short]; conflict {
					t.Errorf(
						"domain %s action %s flag %s reuses -%s reserved by --%s",
						domain.Name,
						action.Name,
						flag.Name,
						flag.Short,
						persistentName,
					)
				}
			}
		}
	}
}

func TestBugAttachmentDownloadStreamsTheSignedURL(t *testing.T) {
	domain := findDomain("bugsandfeatures")
	if domain == nil {
		t.Fatal("bugsandfeatures domain not found")
	}

	for _, action := range domain.Actions {
		if action.Name != "attachments-download" {
			continue
		}
		if action.RESTPath != "{bugExternalGuid}/attachments/{attachmentExternalGuid}/url" ||
			action.DownloadURLField != "sasUrl" || action.DownloadOutputFlag != "out" ||
			action.DownloadDefaultArg != "attachmentExternalGuid" || !action.DisableResponseExport {
			t.Fatalf("attachment download action is not safely wired: %+v", action)
		}
		return
	}

	t.Fatal("attachments-download action not found")
}

// TestReportDomainAliasesAreUniqueAcrossRegistry is scoped to the report template
// and schedule domains: the registry already carries unrelated duplicates
// (codes, automations) that belong to other changes.
func TestReportDomainAliasesAreUniqueAcrossRegistry(t *testing.T) {
	reportDomains := map[string]bool{"report-template": true, "report-schedule": true}
	owners := map[string]string{}
	for _, name := range []string{"report-template", "report-schedule"} {
		domain := findDomain(name)
		if domain == nil {
			t.Fatalf("expected %s domain to be registered", name)
		}
		for _, token := range append([]string{domain.Name}, domain.Aliases...) {
			if owner, taken := owners[token]; taken {
				t.Errorf("%q is used by both %s and %s", token, owner, name)
			}
			owners[token] = name
		}
	}

	for _, domain := range DefaultRegistry.Domains() {
		if reportDomains[domain.Name] {
			continue
		}
		for _, token := range append([]string{domain.Name}, domain.Aliases...) {
			if owner, taken := owners[token]; taken {
				t.Errorf("%q of domain %s collides with the %s domain", token, domain.Name, owner)
			}
		}
	}
}

// TestBugsAndFeaturesVerificationSetDeclaration locks in `verification-set` (UTE-2101): a REST
// PATCH to {externalGuid}/verification carrying verificationRequested + verificationSteps.
func TestBugsAndFeaturesVerificationSetDeclaration(t *testing.T) {
	var action *Action
	for _, d := range DefaultRegistry.Domains() {
		if d.Name != "bugsandfeatures" {
			continue
		}
		for i := range d.Actions {
			if d.Actions[i].Name == "verification-set" {
				action = &d.Actions[i]
			}
		}
		got, _ := buildRESTPath(d, Action{Name: "verification-set", RESTPath: "{externalGuid}/verification"}, map[string]any{"externalGuid": "g1"})
		if got != "/api/bugsandfeatures/g1/verification" {
			t.Errorf("REST path = %q, want /api/bugsandfeatures/g1/verification", got)
		}
	}
	if action == nil {
		t.Fatal("verification-set action not registered")
	}
	if action.ToolName != "uteamup_bugsandfeatures_update_verification" || action.HTTPMethod != "PATCH" {
		t.Errorf("ToolName/HTTPMethod = %q/%q", action.ToolName, action.HTTPMethod)
	}
	bodies := map[string]string{}
	for _, f := range action.Flags {
		bodies[f.Name] = f.BodyName
	}
	if bodies["requested"] != "verificationRequested" || bodies["steps"] != "verificationSteps" {
		t.Errorf("flag body names = %v", bodies)
	}
}
