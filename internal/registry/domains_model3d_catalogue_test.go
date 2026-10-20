package registry

import (
	"slices"
	"strings"
	"testing"
)

func model3DCatalogueAction(t *testing.T, name string) (*Domain, Action) {
	t.Helper()
	domain := findDomain("model3d-catalogue")
	if domain == nil {
		t.Fatal("model3d-catalogue domain is not registered")
	}
	for _, action := range domain.Actions {
		if action.Name == name {
			return domain, action
		}
	}
	t.Fatalf("model3d-catalogue action %q is not registered", name)
	return nil, Action{}
}

func model3DCatalogueBodyNames(action Action) map[string]FlagDef {
	names := map[string]FlagDef{}
	for _, flag := range action.Flags {
		name := flag.BodyName
		if name == "" {
			name = toCamelCase(flag.Name)
		}
		names[name] = flag
	}
	return names
}

func TestModel3DCatalogueActionsResolveToRealControllerRoutes(t *testing.T) {
	t.Parallel()
	domain := findDomain("model3d-catalogue")
	if domain == nil {
		t.Fatal("model3d-catalogue domain is not registered")
	}
	if domain.APIPath != "/api/model3d/catalogue" {
		t.Fatalf("APIPath = %q, want /api/model3d/catalogue", domain.APIPath)
	}

	const guid = "0f8fad5b-d9cb-469f-a165-70867728950e"
	tests := []struct {
		action, method, path string
		args                 map[string]any
	}{
		{"search", "GET", "/api/model3d/catalogue/search", map[string]any{}},
		{"get", "GET", "/api/model3d/catalogue/" + guid, map[string]any{"entryGuid": guid}},
		{"copy", "POST", "/api/model3d/catalogue/copy", map[string]any{}},
		{"share", "POST", "/api/model3d/catalogue/share", map[string]any{}},
		{"withdraw", "POST", "/api/model3d/catalogue/" + guid + "/withdraw", map[string]any{"entryGuid": guid}},
		{"mine", "GET", "/api/model3d/catalogue/mine", map[string]any{}},
	}
	if len(tests) != len(domain.Actions) {
		t.Fatalf("route table covers %d actions but the domain declares %d", len(tests), len(domain.Actions))
	}

	for _, test := range tests {
		t.Run(test.action, func(t *testing.T) {
			domain, action := model3DCatalogueAction(t, test.action)
			if action.MCPOnly {
				t.Fatalf("%s is MCPOnly — the catalogue API is served over REST", test.action)
			}
			method := action.HTTPMethod
			if method == "" {
				method = HTTPMethod[action.Name]
			}
			if method != test.method {
				t.Fatalf("method = %q, want %q", method, test.method)
			}
			path, _ := buildRESTPath(domain, action, test.args)
			if path != test.path {
				t.Fatalf("path = %q, want %q", path, test.path)
			}
			if strings.Contains(path, "{") {
				t.Fatalf("%s left an unexpanded placeholder: %s", test.action, path)
			}
			if path != strings.ToLower(path) {
				t.Fatalf("%s path %q is not lowercase", test.action, path)
			}
		})
	}
}

func TestModel3DCatalogueToolNamesMatchTheBackendTools(t *testing.T) {
	t.Parallel()
	for action, tool := range map[string]string{
		"search":   "uteamup_model3d_catalogue_search",
		"get":      "uteamup_model3d_catalogue_get",
		"copy":     "uteamup_model3d_catalogue_copy",
		"share":    "uteamup_model3d_catalogue_share",
		"withdraw": "uteamup_model3d_catalogue_withdraw",
		"mine":     "uteamup_model3d_catalogue_mine",
	} {
		_, a := model3DCatalogueAction(t, action)
		if a.ToolName != tool {
			t.Errorf("%s ToolName = %q, want %q", action, a.ToolName, tool)
		}
	}
}

func TestModel3DCatalogueSearchSendsTheBackendQueryFields(t *testing.T) {
	t.Parallel()
	_, search := model3DCatalogueAction(t, "search")
	names := model3DCatalogueBodyNames(search)
	for _, field := range []string{"search", "manufacturer", "modelNumber", "partNumber", "gtin", "category", "page", "pageSize"} {
		flag, ok := names[field]
		if !ok {
			t.Fatalf("search sends no %s; Model3DCatalogueSearchQuery binds it", field)
		}
		if flag.Required {
			t.Errorf("search --%s must be optional", flag.Name)
		}
	}
	if size := names["pageSize"]; size.Type != "int" || size.Default != 20 {
		t.Errorf("search --page-size = %+v, want an int defaulting to 20 (backend caps at 50)", size)
	}
	if !slices.Equal(names["category"].AllowedValues, []string{"asset", "part", "tool", "chemical", "other"}) {
		t.Errorf("search --category allows %v", names["category"].AllowedValues)
	}
}

func TestModel3DCatalogueCopyBodyMatchesTheBackendModel(t *testing.T) {
	t.Parallel()
	_, copyAction := model3DCatalogueAction(t, "copy")
	names := model3DCatalogueBodyNames(copyAction)
	entry, ok := names["entryGuid"]
	if !ok || !entry.Required {
		t.Fatal("copy must send a required entryGuid")
	}
	kind, ok := names["itemKind"]
	if !ok || kind.Required {
		t.Fatal("copy must send an optional itemKind")
	}
	if !slices.Equal(kind.AllowedValues, []string{"asset", "part", "tool", "chemical"}) {
		t.Errorf("copy --item-kind allows %v, want the four item kinds", kind.AllowedValues)
	}
	if item, ok := names["itemGuid"]; !ok || item.Required {
		t.Fatal("copy must send an optional itemGuid")
	}
}

func TestModel3DCatalogueShareBodyMatchesTheBackendModel(t *testing.T) {
	t.Parallel()
	_, share := model3DCatalogueAction(t, "share")
	names := model3DCatalogueBodyNames(share)
	for _, field := range []string{"model3DGuid", "name", "category", "manufacturer", "modelNumber", "partNumber", "gtin", "rightsConfirmed", "containsOnlinePhotos"} {
		if _, ok := names[field]; !ok {
			t.Fatalf("share sends no %s; Model3DShareRequestModel needs it", field)
		}
	}
	for _, field := range []string{"model3DGuid", "name", "rightsConfirmed"} {
		if !names[field].Required {
			t.Errorf("share --%s must be required", names[field].Name)
		}
	}
	rights := names["rightsConfirmed"]
	if rights.Type != "bool" || !rights.MustBeTrue || rights.LocalOnly {
		t.Fatalf("share --rights-confirmed = %+v, want a sent bool that must be explicitly true", rights)
	}
	if photos := names["containsOnlinePhotos"]; photos.Type != "bool" || photos.Default != false {
		t.Errorf("share --contains-online-photos = %+v, want a bool defaulting to false", photos)
	}
}

func TestModel3DCatalogueBoundaryIsGuidOnly(t *testing.T) {
	t.Parallel()
	domain := findDomain("model3d-catalogue")
	for _, action := range domain.Actions {
		for _, arg := range action.Args {
			if arg.Type == "int" || arg.Name == "id" {
				t.Errorf("%s arg %s is an integer id — the boundary is GUID-only", action.Name, arg.Name)
			}
			if strings.HasSuffix(arg.Name, "Guid") && arg.Type != "non-empty-uuid" {
				t.Errorf("%s arg %s is %s, want non-empty-uuid", action.Name, arg.Name, arg.Type)
			}
		}
		for _, flag := range action.Flags {
			if strings.HasSuffix(flag.Name, "-guid") && flag.Type != "non-empty-uuid" {
				t.Errorf("%s --%s is %s, want non-empty-uuid", action.Name, flag.Name, flag.Type)
			}
			if flag.Type == "int" && strings.HasSuffix(flag.Name, "-id") {
				t.Errorf("%s --%s is an integer id — the boundary is GUID-only", action.Name, flag.Name)
			}
		}
	}
}

// The tenant comes from the session and the sharer is never revealed, so no
// action may let the caller name a tenant or a user.
func TestModel3DCatalogueTakesNoTenantOrUserArguments(t *testing.T) {
	t.Parallel()
	domain := findDomain("model3d-catalogue")
	forbidden := func(name string) bool {
		lower := strings.ToLower(name)
		return strings.Contains(lower, "tenant") || strings.Contains(lower, "user")
	}
	for _, action := range domain.Actions {
		for _, arg := range action.Args {
			if forbidden(arg.Name) || forbidden(arg.BodyName) {
				t.Errorf("%s arg %s names a tenant or user", action.Name, arg.Name)
			}
		}
		for _, flag := range action.Flags {
			if forbidden(flag.Name) || forbidden(flag.BodyName) || forbidden(flag.HeaderName) || forbidden(flag.QueryName) {
				t.Errorf("%s --%s names a tenant or user", action.Name, flag.Name)
			}
		}
	}
}

// Review is Global Admin work in the web UI only; the CLI must not reach it.
func TestModel3DCatalogueExposesNoModerationActions(t *testing.T) {
	t.Parallel()
	domain := findDomain("model3d-catalogue")
	for _, action := range domain.Actions {
		for _, word := range []string{"approve", "reject", "publish", "moderate", "admin", "resolve-report"} {
			if strings.Contains(action.Name, word) {
				t.Errorf("action %s exposes catalogue moderation", action.Name)
			}
		}
		if strings.Contains(action.RESTBasePath, "admin") || strings.Contains(action.RESTPath, "admin") {
			t.Errorf("action %s routes to the admin catalogue API", action.Name)
		}
		if strings.Contains(action.ToolName, "admin") {
			t.Errorf("action %s calls an admin tool %s", action.Name, action.ToolName)
		}
	}
}
