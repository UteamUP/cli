package registry

import (
	"slices"
	"strings"
	"testing"
)

func model3DAction(t *testing.T, name string) (*Domain, Action) {
	t.Helper()
	domain := findDomain("model3d")
	if domain == nil {
		t.Fatal("model3d domain is not registered")
	}
	for _, action := range domain.Actions {
		if action.Name == name {
			return domain, action
		}
	}
	t.Fatalf("model3d action %q is not registered", name)
	return nil, Action{}
}

func TestModel3DActionsResolveToRealControllerRoutes(t *testing.T) {
	t.Parallel()
	domain := findDomain("model3d")
	if domain == nil {
		t.Fatal("model3d domain is not registered")
	}
	if domain.APIPath != "/api/model3d" {
		t.Fatalf("APIPath = %q, want /api/model3d", domain.APIPath)
	}

	const guid = "0f8fad5b-d9cb-469f-a165-70867728950e"
	tests := []struct {
		action, method, path string
		args                 map[string]any
	}{
		{"list", "GET", "/api/model3d", map[string]any{}},
		{"get", "GET", "/api/model3d/" + guid, map[string]any{"externalGuid": guid}},
		{"delete", "DELETE", "/api/model3d/" + guid, map[string]any{"externalGuid": guid}},
		{"duplicate", "POST", "/api/model3d/" + guid + "/duplicate", map[string]any{"externalGuid": guid, "name": "Copy"}},
		{"spec-lookup", "POST", "/api/model3d/spec-lookup", map[string]any{"query": "Sony PlayStation 5 Slim"}},
		{"view-photos", "POST", "/api/model3d/view-photos", map[string]any{"query": "Sony PlayStation 5 Slim"}},
		{"link", "PUT", "/api/model3d/link", map[string]any{}},
		{"unlink", "PUT", "/api/model3d/link", map[string]any{}},
		{"resolve", "GET", "/api/model3d/item/part/" + guid, map[string]any{"itemKind": "part", "itemGuid": guid}},
	}
	if len(tests) != len(domain.Actions) {
		t.Fatalf("route table covers %d actions but the domain declares %d", len(tests), len(domain.Actions))
	}

	for _, test := range tests {
		t.Run(test.action, func(t *testing.T) {
			domain, action := model3DAction(t, test.action)
			if action.MCPOnly {
				t.Fatalf("%s is MCPOnly — the 3D model API is served over REST", test.action)
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
		})
	}
}

func TestModel3DToolNamesMatchTheBackendTools(t *testing.T) {
	t.Parallel()
	for action, tool := range map[string]string{
		"list":        "uteamup_model3d_list",
		"get":         "uteamup_model3d_get",
		"delete":      "uteamup_model3d_delete",
		"duplicate":   "uteamup_model3d_duplicate",
		"spec-lookup": "uteamup_model3d_spec_lookup",
		"view-photos": "uteamup_model3d_view_photos",
		"link":        "uteamup_model3d_link",
		"unlink":      "uteamup_model3d_link",
		"resolve":     "uteamup_model3d_resolve",
	} {
		_, a := model3DAction(t, action)
		if a.ToolName != tool {
			t.Errorf("%s ToolName = %q, want %q", action, a.ToolName, tool)
		}
	}
}

func TestModel3DLinkBodyMatchesTheBackendModel(t *testing.T) {
	t.Parallel()
	bodyNames := func(action Action) map[string]FlagDef {
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

	_, link := model3DAction(t, "link")
	names := bodyNames(link)
	for _, field := range []string{"itemKind", "itemGuid", "model3DGuid"} {
		flag, ok := names[field]
		if !ok {
			t.Fatalf("link sends no %s; Model3DLinkModel needs it", field)
		}
		if !flag.Required {
			t.Errorf("link --%s must be required", flag.Name)
		}
	}

	// Unlink is the same route with no model, which clears the item's model.
	_, unlink := model3DAction(t, "unlink")
	if _, ok := bodyNames(unlink)["model3DGuid"]; ok {
		t.Fatal("unlink must not send model3DGuid")
	}
}

func TestModel3DItemKindIsConstrainedToTheFourKinds(t *testing.T) {
	t.Parallel()
	want := []string{"asset", "part", "tool", "chemical"}
	_, resolve := model3DAction(t, "resolve")
	if !slices.Equal(resolve.Args[0].AllowedValues, want) {
		t.Errorf("resolve itemKind allows %v, want %v (a free value would change the route)", resolve.Args[0].AllowedValues, want)
	}
	for _, name := range []string{"link", "unlink"} {
		_, action := model3DAction(t, name)
		for _, flag := range action.Flags {
			if flag.Name == "item-kind" && !slices.Equal(flag.AllowedValues, want) {
				t.Errorf("%s --item-kind allows %v, want %v", name, flag.AllowedValues, want)
			}
		}
	}
}

func TestModel3DDuplicateRequiresANameSentInTheBody(t *testing.T) {
	t.Parallel()
	_, action := model3DAction(t, "duplicate")
	if len(action.Flags) != 1 {
		t.Fatalf("duplicate flags = %+v, want only --name", action.Flags)
	}
	name := action.Flags[0]
	if name.Name != "name" || !name.Required || name.Type != "string" || name.QueryName != "" {
		t.Fatalf("duplicate --name = %+v, want a required string body field (Model3DDuplicateModel.Name)", name)
	}
}

func TestModel3DSpecLookupAndViewPhotosSendOptionalQueryAndItemInTheBody(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"spec-lookup", "view-photos"} {
		_, action := model3DAction(t, name)
		want := map[string]string{"query": "string", "item-kind": "string", "item-guid": "non-empty-uuid"}
		if len(action.Flags) != len(want) {
			t.Fatalf("%s flags = %+v, want --query, --item-kind and --item-guid", name, action.Flags)
		}
		for _, flag := range action.Flags {
			if want[flag.Name] != flag.Type || flag.Required || flag.QueryName != "" {
				t.Errorf("%s --%s = %+v, want an optional %s body field", name, flag.Name, flag, want[flag.Name])
			}
		}
	}
}

func TestModel3DDeleteSendsForceAsAQueryParameter(t *testing.T) {
	t.Parallel()
	_, action := model3DAction(t, "delete")
	if len(action.Flags) != 1 || action.Flags[0].QueryName != "force" || action.Flags[0].Type != "bool" {
		t.Fatalf("delete flags = %+v, want one bool --force sent as ?force=", action.Flags)
	}
	if action.Flags[0].Default != false {
		t.Fatal("delete must not force by default; an in-use model is refused unless asked")
	}
}

func TestModel3DBoundaryIsGuidOnly(t *testing.T) {
	t.Parallel()
	domain := findDomain("model3d")
	for _, action := range domain.Actions {
		for _, arg := range action.Args {
			if arg.Type == "int" || arg.Name == "id" {
				t.Errorf("%s arg %s is an integer id — the boundary is GUID-only", action.Name, arg.Name)
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
