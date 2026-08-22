package registry

import (
	"strings"
	"testing"
)

func digitalTwinAction(t *testing.T, name string) (*Domain, Action) {
	t.Helper()
	domain := findDomain("digitaltwin")
	if domain == nil {
		t.Fatal("digitaltwin domain is not registered")
	}
	for _, action := range domain.Actions {
		if action.Name == name {
			return domain, action
		}
	}
	t.Fatalf("digitaltwin action %q is not registered", name)
	return nil, Action{}
}

func TestDigitalTwinActionsResolveToRealControllerRoutes(t *testing.T) {
	t.Parallel()
	domain := findDomain("digitaltwin")
	if domain == nil {
		t.Fatal("digitaltwin domain is not registered")
	}
	if domain.APIPath != "/api/digitaltwin" {
		t.Fatalf("APIPath = %q, want /api/digitaltwin", domain.APIPath)
	}

	tests := []struct {
		action, method string
		args           map[string]any
		path           string
	}{
		// DigitalTwinController — /api/digitaltwin
		{"get-by-asset", "GET", map[string]any{"assetGuid": "a"}, "/api/digitaltwin/asset/a"},
		{"get", "GET", map[string]any{"twinGuid": "t"}, "/api/digitaltwin/t"},
		{"create", "POST", nil, "/api/digitaltwin"},
		{"update", "PUT", map[string]any{"twinGuid": "t"}, "/api/digitaltwin/t"},
		{"delete", "DELETE", map[string]any{"twinGuid": "t"}, "/api/digitaltwin/t"},
		{"live", "GET", map[string]any{"twinGuid": "t"}, "/api/digitaltwin/t/live"},
		{"suggest-anchors", "POST", map[string]any{"twinGuid": "t"}, "/api/digitaltwin/t/suggest-anchors"},

		// DigitalTwinAnchorController — /api/digitaltwinanchor
		{"anchors", "GET", map[string]any{"twinGuid": "t"}, "/api/digitaltwinanchor/twin/t"},
		{"anchor-create", "POST", nil, "/api/digitaltwinanchor"},
		{"anchor-bulk-create", "POST", nil, "/api/digitaltwinanchor/bulk"},
		{"anchor-update", "PUT", map[string]any{"anchorGuid": "n"}, "/api/digitaltwinanchor/n"},
		{"anchor-delete", "DELETE", map[string]any{"anchorGuid": "n"}, "/api/digitaltwinanchor/n"},
	}
	if len(tests) != len(domain.Actions) {
		t.Fatalf("route table covers %d actions but the domain declares %d", len(tests), len(domain.Actions))
	}

	for _, test := range tests {
		t.Run(test.action, func(t *testing.T) {
			domain, action := digitalTwinAction(t, test.action)
			if action.MCPOnly {
				t.Fatalf("%s is MCPOnly — the Digital Twin API is REST-only", test.action)
			}
			if action.HTTPMethod != test.method {
				t.Fatalf("method = %q, want %q (names outside the HTTPMethod map silently default to GET)",
					action.HTTPMethod, test.method)
			}
			args := test.args
			if args == nil {
				args = map[string]any{}
			}
			path, _ := buildRESTPath(domain, action, args)
			if path != test.path {
				t.Fatalf("path = %q, want %q", path, test.path)
			}
			if strings.Contains(path, "{") {
				t.Fatalf("%s left an unexpanded placeholder: %s", test.action, path)
			}
		})
	}
}

func TestDigitalTwinAnchorActionsOverrideTheDomainBasePath(t *testing.T) {
	t.Parallel()
	domain := findDomain("digitaltwin")
	if domain == nil {
		t.Fatal("digitaltwin domain is not registered")
	}
	for _, action := range domain.Actions {
		wantOverride := strings.HasPrefix(action.Name, "anchor") // anchors, anchor-*
		if wantOverride && action.RESTBasePath != digitalTwinAnchorAPIPath {
			t.Errorf("%s RESTBasePath = %q, want %q — anchors live on a separate controller",
				action.Name, action.RESTBasePath, digitalTwinAnchorAPIPath)
		}
		if !wantOverride && action.RESTBasePath != "" {
			t.Errorf("%s RESTBasePath = %q, want the domain APIPath", action.Name, action.RESTBasePath)
		}
	}
}

func TestDigitalTwinMCPToolNamesMatchTheBackendTools(t *testing.T) {
	t.Parallel()
	// Only these two exist as MCP tools; every other ToolName is metadata and the REST
	// route is what resolves at runtime.
	_, byAsset := digitalTwinAction(t, "get-by-asset")
	if byAsset.ToolName != "UteamupDigitalTwinGet" {
		t.Errorf("get-by-asset ToolName = %q, want UteamupDigitalTwinGet", byAsset.ToolName)
	}
	if len(byAsset.Args) != 1 || byAsset.Args[0].Name != "assetGuid" {
		t.Errorf("get-by-asset args = %+v, want a single assetGuid", byAsset.Args)
	}
	_, live := digitalTwinAction(t, "live")
	if live.ToolName != "UteamupDigitalTwinLiveValues" {
		t.Errorf("live ToolName = %q, want UteamupDigitalTwinLiveValues", live.ToolName)
	}
	if len(live.Args) != 1 || live.Args[0].Name != "twinGuid" {
		t.Errorf("live args = %+v, want a single twinGuid", live.Args)
	}
}

func TestDigitalTwinBoundaryIsGuidOnlyWithFloatLiteralDefaults(t *testing.T) {
	t.Parallel()
	domain := findDomain("digitaltwin")
	if domain == nil {
		t.Fatal("digitaltwin domain is not registered")
	}
	for _, action := range domain.Actions {
		for _, arg := range action.Args {
			if arg.Type == "int" {
				t.Errorf("%s arg %s is int — every Digital Twin identifier is a GUID", action.Name, arg.Name)
			}
			if !strings.HasSuffix(arg.Name, "Guid") {
				t.Errorf("%s arg %s is not a GUID positional", action.Name, arg.Name)
			}
		}
		for _, flag := range action.Flags {
			if flag.Type != "float" || flag.Default == nil {
				continue
			}
			// An untyped int default panics in buildActionCommand's type assertion.
			if _, ok := flag.Default.(float64); !ok {
				t.Errorf("%s flag %s has a non-float default %#v", action.Name, flag.Name, flag.Default)
			}
		}
	}
}

func TestDigitalTwinAnchorUpdateSendsNoUnchangedDefaults(t *testing.T) {
	t.Parallel()
	// A partial anchor update must not carry defaults: executeAction sends a flag's
	// Default when the user did not set it, which would overwrite the stored value.
	_, update := digitalTwinAction(t, "anchor-update")
	for _, flag := range update.Flags {
		if flag.Default != nil {
			t.Errorf("anchor-update flag %s declares default %#v — it would overwrite the stored value",
				flag.Name, flag.Default)
		}
		if flag.Required {
			t.Errorf("anchor-update flag %s is required — the update model is fully optional", flag.Name)
		}
	}
}

func TestDigitalTwinNestedPayloadsUseJSONFileFlags(t *testing.T) {
	t.Parallel()
	// nodes[] and anchors[] cannot be expressed as flat flags.
	for _, test := range []struct{ action, flag, bodyName string }{
		{"suggest-anchors", "nodes-file", "nodes"},
		{"anchor-bulk-create", "anchors-file", "anchors"},
	} {
		_, action := digitalTwinAction(t, test.action)
		var found bool
		for _, flag := range action.Flags {
			if flag.Name != test.flag {
				continue
			}
			found = true
			if !flag.JSONFile {
				t.Errorf("%s flag %s is not marked JSONFile — the path would be sent as a string",
					test.action, test.flag)
			}
			if flag.BodyName != test.bodyName {
				t.Errorf("%s flag %s BodyName = %q, want %q", test.action, test.flag, flag.BodyName, test.bodyName)
			}
			if !flag.Required {
				t.Errorf("%s flag %s must be required", test.action, test.flag)
			}
		}
		if !found {
			t.Errorf("%s does not declare a %s flag", test.action, test.flag)
		}
	}
}
