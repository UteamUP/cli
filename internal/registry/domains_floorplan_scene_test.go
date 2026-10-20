package registry

import (
	"strings"
	"testing"
)

func floorplanSceneAction(t *testing.T, name string) (*Domain, Action) {
	t.Helper()
	domain := findDomain("floorplan-scene")
	if domain == nil {
		t.Fatal("floorplan-scene domain is not registered")
	}
	for _, action := range domain.Actions {
		if action.Name == name {
			return domain, action
		}
	}
	t.Fatalf("floorplan-scene action %q is not registered", name)
	return nil, Action{}
}

func TestFloorplanSceneActionsResolveToRealControllerRoutes(t *testing.T) {
	t.Parallel()
	domain := findDomain("floorplan-scene")
	if domain == nil {
		t.Fatal("floorplan-scene domain is not registered")
	}
	if domain.APIPath != "/api/floorplanscene" {
		t.Fatalf("APIPath = %q, want /api/floorplanscene", domain.APIPath)
	}
	if len(domain.Aliases) != 1 || domain.Aliases[0] != "scene3d" {
		t.Fatalf("Aliases = %v, want [scene3d]", domain.Aliases)
	}

	location := map[string]any{"locationGuid": "l"}
	tests := []struct {
		action, method string
		path           string
	}{
		{"get", "GET", "/api/floorplanscene/location/l"},
		{"items", "GET", "/api/floorplanscene/location/l/items"},
		{"save", "PUT", "/api/floorplanscene/location/l"},
		{"delete", "DELETE", "/api/floorplanscene/location/l"},
		{"estimate", "GET", "/api/floorplanscene/location/l/generate-estimate"},
		{"generate", "POST", "/api/floorplanscene/location/l/generate"},
	}
	if len(tests) != len(domain.Actions) {
		t.Fatalf("route table covers %d actions but the domain declares %d", len(tests), len(domain.Actions))
	}

	for _, test := range tests {
		t.Run(test.action, func(t *testing.T) {
			domain, action := floorplanSceneAction(t, test.action)
			if action.MCPOnly {
				t.Fatalf("%s is MCPOnly — the floorplan scene API is served over REST", test.action)
			}
			if action.HTTPMethod != test.method {
				t.Fatalf("method = %q, want %q (names outside the HTTPMethod map silently default to GET)",
					action.HTTPMethod, test.method)
			}
			path, _ := buildRESTPath(domain, action, location)
			if path != test.path {
				t.Fatalf("path = %q, want %q", path, test.path)
			}
			if strings.Contains(path, "{") {
				t.Fatalf("%s left an unexpanded placeholder: %s", test.action, path)
			}
		})
	}
}

func TestFloorplanSceneMCPToolNamesMatchTheBackendTools(t *testing.T) {
	t.Parallel()
	// The backend registers exactly these five tools; estimate has none and its
	// ToolName is metadata only.
	for action, tool := range map[string]string{
		"get":      "UteamupFloorplanSceneGet",
		"items":    "uteamup_floorplan_scene_items",
		"save":     "UteamupFloorplanSceneSave",
		"delete":   "UteamupFloorplanSceneDelete",
		"generate": "UteamupFloorplanSceneGenerate",
	} {
		_, a := floorplanSceneAction(t, action)
		if a.ToolName != tool {
			t.Errorf("%s ToolName = %q, want %q", action, a.ToolName, tool)
		}
	}
}

func TestFloorplanSceneBoundaryIsLocationGuidOnly(t *testing.T) {
	t.Parallel()
	domain := findDomain("floorplan-scene")
	if domain == nil {
		t.Fatal("floorplan-scene domain is not registered")
	}
	for _, action := range domain.Actions {
		if len(action.Args) != 1 {
			t.Errorf("%s args = %+v, want a single locationGuid", action.Name, action.Args)
			continue
		}
		arg := action.Args[0]
		if arg.Name != "locationGuid" || arg.Type != "uuid" || !arg.Required {
			t.Errorf("%s arg = %+v, want a required uuid locationGuid", action.Name, arg)
		}
		for _, flag := range action.Flags {
			if flag.Type == "int" && strings.HasSuffix(flag.Name, "-id") {
				t.Errorf("%s flag %s is an integer id — the boundary is GUID-only", action.Name, flag.Name)
			}
		}
	}
}

func TestFloorplanSceneSaveSendsTheSaveModel(t *testing.T) {
	t.Parallel()
	_, save := floorplanSceneAction(t, "save")
	want := map[string]struct {
		bodyName, flagType string
		required, jsonFile bool
	}{
		"name":             {"", "string", true, false},
		"expected-version": {"expectedVersion", "int", false, false},
		"scene-file":       {"scene", "string", true, true},
		"generated-by-ai":  {"generatedByAi", "bool", false, false},
	}
	if len(save.Flags) != len(want) {
		t.Fatalf("save declares %d flags, want %d", len(save.Flags), len(want))
	}
	for _, flag := range save.Flags {
		w, ok := want[flag.Name]
		if !ok {
			t.Errorf("save declares unexpected flag %s", flag.Name)
			continue
		}
		if flag.BodyName != w.bodyName || flag.Type != w.flagType ||
			flag.Required != w.required || flag.JSONFile != w.jsonFile {
			t.Errorf("save flag %s = %+v, want %+v", flag.Name, flag, w)
		}
		if flag.Name == "expected-version" {
			// An untyped non-int default panics in buildActionCommand's int assertion.
			if v, ok := flag.Default.(int); !ok || v != 0 {
				t.Errorf("expected-version default = %#v, want int 0", flag.Default)
			}
		}
	}
}
