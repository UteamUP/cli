package registry

import (
	"strings"
	"testing"
)

func findCodeDomain(t *testing.T) *Domain {
	t.Helper()
	for _, dom := range DefaultRegistry.Domains() {
		if dom.Name == "code" {
			return dom
		}
	}
	t.Fatal("expected code domain to be registered")
	return nil
}

func TestCodeDomainTargetsPluralRoute(t *testing.T) {
	d := findCodeDomain(t)
	// CodesController routes at api/codes (plural) — the auto-derived
	// "/api/code" base never matched a backend route.
	if d.APIPath != "/api/codes" {
		t.Errorf("code domain APIPath = %q, want %q", d.APIPath, "/api/codes")
	}
	if len(d.Aliases) != 1 || d.Aliases[0] != "codes" {
		t.Errorf("code domain aliases = %+v, want [codes]", d.Aliases)
	}
}

func TestCodeResolveActionWired(t *testing.T) {
	d := findCodeDomain(t)
	var action *Action
	for i := range d.Actions {
		if d.Actions[i].Name == "resolve" {
			action = &d.Actions[i]
			break
		}
	}
	if action == nil {
		t.Fatal("expected `resolve` action on code domain")
	}

	if action.ToolName != "UteamupCodeResolve" {
		t.Errorf("resolve ToolName = %q, want %q", action.ToolName, "UteamupCodeResolve")
	}
	// Default GET — the resolver is a read (soft-miss 200, never a 404).
	if action.HTTPMethod != "" {
		t.Errorf("resolve HTTPMethod = %q, want \"\" (defaults to GET)", action.HTTPMethod)
	}
	if action.RESTPath != "resolve/{value}" {
		t.Errorf("resolve RESTPath = %q, want %q (GET api/codes/resolve/{value})", action.RESTPath, "resolve/{value}")
	}
	if len(action.Args) != 1 || action.Args[0].Name != "value" || !action.Args[0].Required || action.Args[0].Type != "string" {
		t.Fatalf("resolve expected single required string positional arg 'value', got %+v", action.Args)
	}
}

// The resolver's target vocabulary is the operator's only clue about what a scan can
// return. CodeResolveResponseModel.TargetType gained assetGroup when asset groups
// shipped; a description that still stops at `asset` teaches the wrong answer.
func TestCodeResolveDescriptionListsEveryTargetType(t *testing.T) {
	d := findCodeDomain(t)
	action := findAction(d, "resolve")
	if action == nil {
		t.Fatal("expected `resolve` action on code domain")
	}

	for _, target := range []string{"stockItem", "stockItemUnit", "stockBin", "asset", "assetGroup", "part", "tool", "chemical", "workPermit", "unknown"} {
		if !strings.Contains(action.Description, target) {
			t.Errorf("resolve description is missing the %q target type: %q", target, action.Description)
		}
	}
}

func TestCodeTargetActionsBuildTheTargetRoutes(t *testing.T) {
	d := findCodeDomain(t)
	const target = "44444444-4444-4444-4444-444444444444"
	const code = "55555555-5555-5555-5555-555555555555"

	cases := []struct {
		name   string
		method string
		tool   string
		args   map[string]any
		want   string
	}{
		{"target-list", "GET", "UteamupCodeListForTarget", map[string]any{"targetType": "part", "targetGuid": target}, "/api/codes/targets/part/" + target},
		{"target-register", "POST", "UteamupCodeRegisterForTarget", map[string]any{"targetType": "stockbin", "targetGuid": target}, "/api/codes/targets/stockbin/" + target},
		{"target-generate", "POST", "UteamupCodeGenerateForTarget", map[string]any{"targetType": "tool", "targetGuid": target}, "/api/codes/targets/tool/" + target + "/generate"},
		{"target-generate", "POST", "UteamupCodeGenerateForTarget", map[string]any{"targetType": "asset", "targetGuid": target}, "/api/codes/targets/asset/" + target + "/generate"},
		{"target-list", "GET", "UteamupCodeListForTarget", map[string]any{"targetType": "assetgroup", "targetGuid": target}, "/api/codes/targets/assetgroup/" + target},
		{"target-remove", "DELETE", "UteamupCodeRemoveFromTarget", map[string]any{"targetType": "workpermit", "targetGuid": target, "codeGuid": code}, "/api/codes/targets/workpermit/" + target + "/" + code},
	}

	for _, tc := range cases {
		action := findAction(d, tc.name)
		if action == nil {
			t.Fatalf("missing code action %q", tc.name)
		}
		if action.HTTPMethod != tc.method || action.ToolName != tc.tool {
			t.Errorf("%s = %s %s, want %s %s", tc.name, action.HTTPMethod, action.ToolName, tc.method, tc.tool)
		}
		got, consumed := buildRESTPath(d, *action, tc.args)
		if got != tc.want {
			t.Errorf("%s path = %q, want %q", tc.name, got, tc.want)
		}
		if len(consumed) != len(tc.args) {
			t.Errorf("%s consumed %v, want every positional arg", tc.name, consumed)
		}
	}
}

func TestCodeTargetTypeIsRestrictedToTheBackendTable(t *testing.T) {
	d := findCodeDomain(t)
	for _, name := range []string{"target-list", "target-register", "target-generate", "target-remove"} {
		action := findAction(d, name)
		if action == nil {
			t.Fatalf("missing code action %q", name)
		}
		got := strings.Join(action.Args[0].AllowedValues, ",")
		if action.Args[0].Name != "targetType" || got != "asset,assetgroup,part,tool,chemical,stockitem,stockbin,workpermit" {
			t.Errorf("%s targetType allowed values = %q", name, got)
		}
		if action.Args[1].Type != "non-empty-uuid" {
			t.Errorf("%s targetGuid type = %q, want non-empty-uuid", name, action.Args[1].Type)
		}
	}

	register := findAction(d, "target-register")
	required := map[string]bool{}
	for _, flag := range register.Flags {
		required[flag.Name] = flag.Required
	}
	if !required["type"] || !required["value"] || required["description"] {
		t.Errorf("target-register flags required = %v, want type and value required, description optional", required)
	}

	// The backend binds the generated code type with [FromQuery]; a body field would be ignored
	// and every generate would silently fall back to a barcode.
	generate := findAction(d, "target-generate")
	if len(generate.Flags) != 1 {
		t.Fatalf("target-generate flags = %v, want exactly the type flag", generate.Flags)
	}
	typeFlag := generate.Flags[0]
	if typeFlag.Name != "type" || typeFlag.QueryName != "type" || typeFlag.Required || typeFlag.Default != "BARCODE" {
		t.Errorf("target-generate type flag = %+v, want optional query flag defaulting to BARCODE", typeFlag)
	}
}

func TestTenantHolidayGuidRoutesResolve(t *testing.T) {
	d := findDomain("tenant-holiday")
	if d == nil {
		t.Fatal("expected tenant-holiday domain to be registered")
	}

	actions := map[string]Action{}
	for _, action := range d.Actions {
		actions[action.Name] = action
	}

	cases := []struct {
		name string
		args map[string]any
		want string
	}{
		{"year", map[string]any{"year": 2026}, "/api/tenantholiday/year/2026"},
		{"update", map[string]any{"holidayGuid": "holiday-1"}, "/api/tenantholiday/by-guid/holiday-1"},
		{"delete", map[string]any{"holidayGuid": "holiday-1"}, "/api/tenantholiday/by-guid/holiday-1"},
		{"import", map[string]any{"year": 2026}, "/api/tenantholiday/import/2026"},
	}

	for _, tc := range cases {
		action, ok := actions[tc.name]
		if !ok {
			t.Fatalf("missing tenant-holiday action %q", tc.name)
		}
		got, consumed := buildRESTPath(d, action, tc.args)
		if got != tc.want {
			t.Fatalf("%s path = %q, want %q", tc.name, got, tc.want)
		}
		if len(consumed) != 1 {
			t.Fatalf("%s consumed = %v, want one path arg", tc.name, consumed)
		}
	}
}
