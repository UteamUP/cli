package registry

import (
	"slices"
	"testing"
)

func productWebFilesAction(t *testing.T, name string) (*Domain, Action) {
	t.Helper()
	domain := findDomain("product-web-files")
	if domain == nil {
		t.Fatal("product-web-files domain is not registered")
	}
	for _, action := range domain.Actions {
		if action.Name == name {
			return domain, action
		}
	}
	t.Fatalf("product-web-files action %q is not registered", name)
	return nil, Action{}
}

func TestProductWebFilesActionsResolveToRealControllerRoutes(t *testing.T) {
	t.Parallel()
	tests := map[string]struct{ method, path string }{
		"search": {"POST", "/api/productwebfiles/search"},
		"import": {"POST", "/api/productwebfiles/import"},
	}
	domain := findDomain("product-web-files")
	if domain == nil {
		t.Fatal("product-web-files domain is not registered")
	}
	if len(tests) != len(domain.Actions) {
		t.Fatalf("route table covers %d actions but the domain declares %d", len(tests), len(domain.Actions))
	}
	for name, want := range tests {
		domain, action := productWebFilesAction(t, name)
		if action.MCPOnly {
			t.Fatalf("%s is MCPOnly — the API is served over REST", name)
		}
		if action.HTTPMethod != want.method {
			t.Errorf("%s method = %q, want %s", name, action.HTTPMethod, want.method)
		}
		if path, _ := buildRESTPath(domain, action, map[string]any{}); path != want.path {
			t.Errorf("%s path = %q, want %q", name, path, want.path)
		}
	}
}

func TestProductWebFilesToolNamesMatchTheBackendTools(t *testing.T) {
	t.Parallel()
	for action, tool := range map[string]string{
		"search": "uteamup_product_web_files_search",
		"import": "uteamup_product_web_files_import",
	} {
		_, a := productWebFilesAction(t, action)
		if a.ToolName != tool {
			t.Errorf("%s ToolName = %q, want %q", action, a.ToolName, tool)
		}
	}
}

func TestProductWebFilesBodiesMatchTheBackendModels(t *testing.T) {
	t.Parallel()
	required := map[string][]string{
		"search": {"itemKind", "itemGuid"},
		"import": {"itemKind", "itemGuid", "url"},
	}
	for name, fields := range required {
		_, action := productWebFilesAction(t, name)
		flags := map[string]FlagDef{}
		for _, flag := range action.Flags {
			body := flag.BodyName
			if body == "" {
				body = toCamelCase(flag.Name)
			}
			flags[body] = flag
			if body == "tenantGuid" || body == "userId" {
				t.Errorf("%s must not send %s; identity comes from the token", name, body)
			}
		}
		for _, field := range fields {
			flag, ok := flags[field]
			if !ok {
				t.Fatalf("%s sends no %s", name, field)
			}
			if !flag.Required {
				t.Errorf("%s --%s must be required", name, flag.Name)
			}
		}
		if kind := flags["itemKind"]; !slices.Equal(kind.AllowedValues, []string{"asset", "part", "tool", "chemical"}) {
			t.Errorf("%s --item-kind allows %v, want the four item kinds", name, kind.AllowedValues)
		}
	}
}
