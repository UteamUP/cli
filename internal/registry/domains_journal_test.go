package registry

import (
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/uteamup/cli/internal/auth"
	"github.com/uteamup/cli/internal/client"
	"github.com/uteamup/cli/internal/logging"
)

// --- Journal domain ---

func TestJournalDomainRegistered(t *testing.T) {
	d := findDomain("journal")
	if d == nil {
		t.Fatal("expected journal domain to be registered")
	}
}

func TestJournalDomainAliases(t *testing.T) {
	d := findDomain("journal")
	if d == nil {
		t.Fatal("expected journal domain to be registered")
	}
	expected := map[string]bool{"journals": true}
	for _, alias := range d.Aliases {
		delete(expected, alias)
	}
	if len(expected) > 0 {
		t.Errorf("missing aliases: %v", expected)
	}
}

func TestJournalDomainActions(t *testing.T) {
	d := findDomain("journal")
	if d == nil {
		t.Fatal("expected journal domain to be registered")
	}

	expected := map[string]string{
		"by-code":           "UteamupJournalGetByCode",
		"by-asset":          "UteamupJournalGetByAsset",
		"import":            "UteamupJournalImport",
		"create-from-image": "UteamupJournalCreateFromImage",
		"search-assets":     "UteamupAssetMentionSearch",
		"search-workorders": "UteamupWorkorderMentionSearch",
		"search-tools":      "UteamupToolMentionSearch",
		"search-parts":      "UteamupPartMentionSearch",
		"search-chemicals":  "UteamupChemicalMentionSearch",
		"search-locations":  "UteamupLocationMentionSearch",
		"search-customers":  "UteamupCustomerMentionSearch",
		"search-contacts":   "UteamupContactMentionSearch",
		"search-vendors":    "UteamupVendorMentionSearch",
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

func TestJournalListUsesSearchAdapter(t *testing.T) {
	d := findDomain("journal")
	if d == nil {
		t.Fatal("expected journal domain to be registered")
	}
	action := findAction(d, "list")
	if action == nil {
		t.Fatal("expected list action on journal domain")
	}
	if action.HTTPMethod != "POST" || action.RESTPath != "search" {
		t.Fatalf("journal list route = %s %q, want POST search", action.HTTPMethod, action.RESTPath)
	}

	path, consumed := buildRESTPath(d, *action, nil)
	if path != "/api/journal/search" {
		t.Fatalf("journal list path = %q, want /api/journal/search", path)
	}
	if len(consumed) != 0 {
		t.Fatalf("journal list unexpectedly consumed path args: %v", consumed)
	}

	flags := flagsToMap(action.Flags)
	if flags["page"].BodyName != "pageNumber" {
		t.Errorf("list --page body name = %q, want pageNumber", flags["page"].BodyName)
	}
	if flags["page-size"].BodyName != "pageSize" {
		t.Errorf("list --page-size body name = %q, want pageSize", flags["page-size"].BodyName)
	}
	if flags["page"].Default != 1 || flags["page-size"].Default != 20 {
		t.Errorf("list pagination defaults = page %v, size %v; want 1 and 20", flags["page"].Default, flags["page-size"].Default)
	}
}

func TestJournalCrudActionsUseCanonicalGuidRoutes(t *testing.T) {
	d := findDomain("journal")
	if d == nil {
		t.Fatal("expected journal domain to be registered")
	}

	journalGUID := "11111111-2222-4333-8444-555555555555"
	for _, tc := range []struct {
		name   string
		method string
	}{
		{name: "get", method: "GET"},
		{name: "delete", method: "DELETE"},
	} {
		action := findAction(d, tc.name)
		if action == nil {
			t.Fatalf("expected %s action on journal domain", tc.name)
		}
		if action.HTTPMethod != tc.method {
			t.Errorf("%s method = %q, want %q", tc.name, action.HTTPMethod, tc.method)
		}
		if action.RESTPath != "by-guid/{journalGuid}" {
			t.Errorf("%s RESTPath = %q, want by-guid/{journalGuid}", tc.name, action.RESTPath)
		}
		assertJournalGuidArg(t, tc.name, action.Args, "journalGuid")

		path, consumed := buildRESTPath(d, *action, map[string]any{"journalGuid": journalGUID})
		if path != "/api/journal/by-guid/"+journalGUID {
			t.Errorf("%s path = %q, want /api/journal/by-guid/%s", tc.name, path, journalGUID)
		}
		if len(consumed) != 1 || consumed[0] != "journalGuid" {
			t.Errorf("%s consumed args = %v, want [journalGuid]", tc.name, consumed)
		}
	}
}

func TestJournalCreateAndUpdateUseNativeMultipart(t *testing.T) {
	const journal = "11111111-2222-4333-8444-555555555555"
	for _, name := range []string{"create", "update"} {
		action := findAction(findDomain("journal"), name)
		if action == nil || action.MCPOnly || !action.FormBody || !reflect.DeepEqual(action.FormObjectFields, []string{"mutation"}) {
			t.Fatalf("%s must use the reviewed multipart REST adapter: %+v", name, action)
		}
		fromJSON := flagsToMap(action.Flags)["from-json"]
		if !fromJSON.Required || !fromJSON.RootJSONObjectFile || fromJSON.JSONFile || fromJSON.BodyName != "" {
			t.Fatalf("%s must preserve the unwrapped reviewed model: %+v", name, fromJSON)
		}
		want := "/api/journal"
		if name == "update" {
			want += "/by-guid/" + journal
			assertJournalGuidArg(t, name, action.Args, "journalGuid")
		}
		if path, _ := buildRESTPath(findDomain("journal"), *action, map[string]any{"journalGuid": journal}); path != want {
			t.Fatalf("%s path = %q, want %q", name, path, want)
		}
	}
}

func TestJournalMultipartCommandsPreserveReviewedIdentityForBothSignIns(t *testing.T) {
	const journal = "11111111-2222-4333-8444-555555555555"
	const operation = "aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee"
	const tenant = "55555555-5555-4555-8555-555555555555"
	for _, method := range []string{"browser", "apikey"} {
		for _, action := range []string{"create", "update"} {
			t.Run(method+"/"+action, func(t *testing.T) {
				t.Setenv("HOME", t.TempDir())
				t.Setenv("USERPROFILE", os.Getenv("HOME"))
				calls := 0
				server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					calls++
					wantMethod, wantPath := http.MethodPost, "/api/journal"
					if action == "update" {
						wantMethod, wantPath = http.MethodPut, wantPath+"/by-guid/"+journal
					}
					if r.Method != wantMethod || r.URL.Path != wantPath {
						t.Errorf("route = %s %s", r.Method, r.URL.Path)
					}
					if r.Header.Get("X-Tenant-Guid") != tenant || r.Header.Get("Authorization") != "Bearer journal-fixture" || r.Header.Get("X-Requested-With") != "XMLHttpRequest" {
						t.Error("normal authenticated tenant and CSRF headers must be preserved")
					}
					if err := r.ParseMultipartForm(1 << 20); err != nil {
						t.Error(err)
						return
					}
					want := map[string][]string{"Title": {"Þróunarteymið"}, "Content": {"**Unnið** og yfirfarið."}, "IsMarkdown": {"true"}, "Mutation.operationGuid": {operation}, "Mutation.expectedUpdatedAt": {"2026-11-15T20:00:00.1234567Z"}}
					if !reflect.DeepEqual(r.MultipartForm.Value, want) {
						t.Errorf("reviewed form changed: %#v", r.MultipartForm.Value)
					}
					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write([]byte(`{"operationGuid":"` + operation + `","journalGuid":"` + journal + `","tenantGuid":"` + tenant + `"}`))
				}))
				defer server.Close()
				if err := auth.SaveToken(&auth.TokenData{APIOrigin: server.URL, AccessToken: "journal-fixture", AuthMethod: method, ExpiresAt: time.Now().Add(time.Hour), TenantGUID: tenant}); err != nil {
					t.Fatal(err)
				}
				api := client.NewAPIClient(server.URL, time.Second, true, client.RetryOptions{MaxRetries: 0}, logging.New(logging.LevelError))
				path := writeRegistryJSONFixture(t, `{"title":"Þróunarteymið","content":"**Unnið** og yfirfarið.","isMarkdown":true,"mutation":{"operationGuid":"`+operation+`","expectedUpdatedAt":"2026-11-15T20:00:00.1234567Z"}}`)
				format := "json"
				command := buildDomainCommand(findDomain("journal"), func() (*client.APIClient, error) { return api, nil }, logging.New(logging.LevelError), &format, &ExportConfig{})
				args := []string{action}
				if action == "update" {
					args = append(args, journal)
				}
				command.SetArgs(append(args, "--from-json", path))
				if err := command.Execute(); err != nil {
					t.Fatal(err)
				}
				if calls != 1 {
					t.Fatalf("requests=%d, want one native REST mutation", calls)
				}
			})
		}
	}
}

func TestJournalImportActionsUseExactMCPContracts(t *testing.T) {
	d := findDomain("journal")
	if d == nil {
		t.Fatal("expected journal domain to be registered")
	}

	for _, tc := range []struct {
		name       string
		toolName   string
		argNames   []string
		flagBodies map[string]string
	}{
		{
			name:       "import",
			toolName:   "UteamupJournalImport",
			argNames:   []string{"fileName", "fileContentBase64"},
			flagBodies: map[string]string{"title": "title", "summary": "summary", "target-journal-guid": "targetJournalGuid"},
		},
		{
			name:       "create-from-image",
			toolName:   "UteamupJournalCreateFromImage",
			argNames:   []string{"imageFileName", "imageContentBase64"},
			flagBodies: map[string]string{"title": "title"},
		},
	} {
		action := findAction(d, tc.name)
		if action == nil {
			t.Fatalf("expected %s action on journal domain", tc.name)
		}
		if action.ToolName != tc.toolName || !action.MCPOnly {
			t.Fatalf("%s transport = tool %q MCPOnly=%v, want MCP tool %q", tc.name, action.ToolName, action.MCPOnly, tc.toolName)
		}
		if action.RESTBasePath != "" || action.RESTPath != "" || action.HTTPMethod != "" {
			t.Fatalf("%s must not fall through to the multipart REST endpoint: %+v", tc.name, action)
		}
		if len(action.Args) != len(tc.argNames) {
			t.Fatalf("%s args = %+v, want %v", tc.name, action.Args, tc.argNames)
		}
		for index, name := range tc.argNames {
			if action.Args[index].Name != name || !action.Args[index].Required {
				t.Errorf("%s arg %d = %+v, want required %s", tc.name, index, action.Args[index], name)
			}
		}
		flags := flagsToMap(action.Flags)
		for flagName, bodyName := range tc.flagBodies {
			if flags[flagName].BodyName != bodyName {
				t.Errorf("%s --%s maps to %q, want %q", tc.name, flagName, flags[flagName].BodyName, bodyName)
			}
		}
	}
}

func TestJournalMentionSearchActionsUseControllerRoutes(t *testing.T) {
	d := findDomain("journal")
	if d == nil {
		t.Fatal("expected journal domain to be registered")
	}

	for _, tc := range []struct {
		name     string
		basePath string
	}{
		{name: "search-assets", basePath: "/api/asset"},
		{name: "search-workorders", basePath: "/api/workorder"},
		{name: "search-tools", basePath: "/api/tool"},
		{name: "search-parts", basePath: "/api/part"},
		{name: "search-chemicals", basePath: "/api/chemical"},
		{name: "search-locations", basePath: "/api/location"},
		{name: "search-customers", basePath: "/api/customer"},
		{name: "search-contacts", basePath: "/api/contact"},
		{name: "search-vendors", basePath: "/api/vendor"},
	} {
		action := findAction(d, tc.name)
		if action == nil {
			t.Fatalf("expected %s action on journal domain", tc.name)
		}
		if action.MCPOnly || action.HTTPMethod != "GET" || action.RESTBasePath != tc.basePath || action.RESTPath != "mention-search" {
			t.Fatalf("%s REST adapter is invalid: %+v", tc.name, action)
		}
		if len(action.Args) != 1 || action.Args[0].Name != "query" || action.Args[0].QueryName != "query" || !action.Args[0].Required {
			t.Fatalf("%s query binding = %+v", tc.name, action.Args)
		}
		limit := flagsToMap(action.Flags)["limit"]
		if limit.QueryName != "limit" || limit.Default != 8 {
			t.Fatalf("%s limit binding = %+v", tc.name, limit)
		}

		path, consumed := buildRESTPath(d, *action, nil)
		path = appendQueryParameters(path, map[string]any{"query": "pump 1", "limit": 8})
		want := tc.basePath + "/mention-search?limit=8&query=pump+1"
		if path != want || len(consumed) != 0 {
			t.Fatalf("%s resolved request = %q consumed=%v, want %q", tc.name, path, consumed, want)
		}
	}
}

func TestJournalByCodeArgsAndFlags(t *testing.T) {
	d := findDomain("journal")
	if d == nil {
		t.Fatal("expected journal domain to be registered")
	}

	action := findAction(d, "by-code")
	if action == nil {
		t.Fatal("expected by-code action on journal domain")
	}

	if action.HTTPMethod != "GET" || action.RESTPath != "by-code/{codeCatalogEntryGuid}" {
		t.Fatalf("by-code route = %s %q, want GET by-code/{codeCatalogEntryGuid}", action.HTTPMethod, action.RESTPath)
	}
	assertJournalGuidArg(t, "by-code", action.Args, "codeCatalogEntryGuid")

	codeGUID := "22222222-3333-4444-8555-666666666666"
	path, consumed := buildRESTPath(d, *action, map[string]any{"codeCatalogEntryGuid": codeGUID})
	if path != "/api/journal/by-code/"+codeGUID {
		t.Errorf("by-code path = %q, want /api/journal/by-code/%s", path, codeGUID)
	}
	if len(consumed) != 1 || consumed[0] != "codeCatalogEntryGuid" {
		t.Errorf("by-code consumed args = %v, want [codeCatalogEntryGuid]", consumed)
	}

	// Must have page and page-size flags
	flagMap := flagsToMap(action.Flags)
	if _, ok := flagMap["page"]; !ok {
		t.Error("by-code action missing page flag")
	}
	if _, ok := flagMap["page-size"]; !ok {
		t.Error("by-code action missing page-size flag")
	}
	if flagMap["page"].QueryName != "pageNumber" || flagMap["page-size"].QueryName != "pageSize" {
		t.Errorf("by-code pagination query binding is wrong: %+v", action.Flags)
	}
}

func TestJournalByAssetArgsAndFlags(t *testing.T) {
	d := findDomain("journal")
	if d == nil {
		t.Fatal("expected journal domain to be registered")
	}

	action := findAction(d, "by-asset")
	if action == nil {
		t.Fatal("expected by-asset action on journal domain")
	}

	if action.HTTPMethod != "GET" || action.RESTPath != "by-asset/{assetGuid}" {
		t.Fatalf("by-asset route = %s %q, want GET by-asset/{assetGuid}", action.HTTPMethod, action.RESTPath)
	}
	assertJournalGuidArg(t, "by-asset", action.Args, "assetGuid")

	assetGUID := "33333333-4444-4555-8666-777777777777"
	path, consumed := buildRESTPath(d, *action, map[string]any{"assetGuid": assetGUID})
	if path != "/api/journal/by-asset/"+assetGUID {
		t.Errorf("by-asset path = %q, want /api/journal/by-asset/%s", path, assetGUID)
	}
	if len(consumed) != 1 || consumed[0] != "assetGuid" {
		t.Errorf("by-asset consumed args = %v, want [assetGuid]", consumed)
	}

	// Must have page and page-size flags
	flagMap := flagsToMap(action.Flags)
	if _, ok := flagMap["page"]; !ok {
		t.Error("by-asset action missing page flag")
	}
	if _, ok := flagMap["page-size"]; !ok {
		t.Error("by-asset action missing page-size flag")
	}
	if flagMap["page"].QueryName != "pageNumber" || flagMap["page-size"].QueryName != "pageSize" {
		t.Errorf("by-asset pagination query binding is wrong: %+v", action.Flags)
	}
}

func assertJournalGuidArg(t *testing.T, actionName string, args []ArgDef, expectedName string) {
	t.Helper()
	if len(args) != 1 {
		t.Fatalf("%s expected exactly one GUID positional arg, got %+v", actionName, args)
	}
	if args[0].Name != expectedName || args[0].Type != "uuid" || !args[0].Required {
		t.Fatalf("%s GUID arg = %+v, want required uuid %s", actionName, args[0], expectedName)
	}
}

// --- CodeCatalog domain ---

func TestCodeCatalogDomainRegistered(t *testing.T) {
	d := findDomain("codecatalog")
	if d == nil {
		t.Fatal("expected codecatalog domain to be registered")
	}
}

func TestCodeCatalogDomainAliases(t *testing.T) {
	d := findDomain("codecatalog")
	if d == nil {
		t.Fatal("expected codecatalog domain to be registered")
	}
	expected := map[string]bool{"cc": true, "codes": true}
	for _, alias := range d.Aliases {
		delete(expected, alias)
	}
	if len(expected) > 0 {
		t.Errorf("missing aliases: %v", expected)
	}
}

func TestCodeCatalogSearchAction(t *testing.T) {
	d := findDomain("codecatalog")
	if d == nil {
		t.Fatal("expected codecatalog domain to be registered")
	}

	action := findAction(d, "search")
	if action == nil {
		t.Fatal("expected search action on codecatalog domain")
	}
	if action.ToolName != "UteamupCodeCatalogSearch" {
		t.Errorf("expected tool UteamupCodeCatalogSearch, got %q", action.ToolName)
	}
}

func TestCodeCatalogSearchArgAndFlags(t *testing.T) {
	d := findDomain("codecatalog")
	if d == nil {
		t.Fatal("expected codecatalog domain to be registered")
	}

	action := findAction(d, "search")
	if action == nil {
		t.Fatal("expected search action on codecatalog domain")
	}

	// Required positional arg: query
	if len(action.Args) == 0 {
		t.Fatal("search action should have at least one positional arg")
	}
	if action.Args[0].Name != "query" {
		t.Errorf("expected arg name query, got %q", action.Args[0].Name)
	}
	if !action.Args[0].Required {
		t.Error("query arg should be required")
	}

	// limit flag
	flagMap := flagsToMap(action.Flags)
	limitFlag, ok := flagMap["limit"]
	if !ok {
		t.Fatal("search action missing limit flag")
	}
	if limitFlag.Default != 10 {
		t.Errorf("limit default should be 10, got %v", limitFlag.Default)
	}
	if limitFlag.Required {
		t.Error("limit flag should not be required")
	}
}

// --- helpers ---

func findDomain(name string) *Domain {
	for _, d := range DefaultRegistry.Domains() {
		if d.Name == name {
			return d
		}
	}
	return nil
}

func findAction(d *Domain, name string) *Action {
	for i := range d.Actions {
		if d.Actions[i].Name == name {
			return &d.Actions[i]
		}
	}
	return nil
}

func flagsToMap(flags []FlagDef) map[string]FlagDef {
	m := make(map[string]FlagDef, len(flags))
	for _, f := range flags {
		m[f.Name] = f
	}
	return m
}
