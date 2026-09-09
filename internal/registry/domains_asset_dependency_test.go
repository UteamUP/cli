package registry

import "testing"

// The asset-dependency domain mirrors AssetDependencyController (/api/assetdependency/*).
// Edges are tenant-level facts about the equipment, so the backend gates them with the
// asset permissions rather than the group ones — the CLI surface must not suggest
// otherwise by routing under /api/assetgroup.

func assetDependencyDomain(t *testing.T) *Domain {
	t.Helper()
	d := findDomain("asset-dependency")
	if d == nil {
		t.Fatal("expected asset-dependency domain to be registered")
	}
	return d
}

func TestAssetDependencyDomainRoutesUnderAssetdependency(t *testing.T) {
	d := assetDependencyDomain(t)
	if d.APIPath != "/api/assetdependency" {
		t.Errorf("APIPath = %q, want /api/assetdependency", d.APIPath)
	}
	hasAlias := false
	for _, alias := range d.Aliases {
		if alias == "adep" {
			hasAlias = true
		}
	}
	if !hasAlias {
		t.Errorf("expected the adep alias, got %v", d.Aliases)
	}
}

func TestAssetDependencyActionsMatchBackendContract(t *testing.T) {
	d := assetDependencyDomain(t)
	cases := []struct{ name, tool, method, path string }{
		{"list", "UteamupAssetDependencyListByAsset", "", "by-asset/{assetGuid}"},
		{"create", "UteamupAssetDependencyCreate", "", ""},
		{"update", "UteamupAssetDependencyUpdate", "", "{dependencyGuid}"},
		{"delete", "UteamupAssetDependencyDelete", "", "{dependencyGuid}"},
		{"impact", "UteamupAssetDependencyImpactAnalyze", "", "by-asset/{assetGuid}/impact"},
		{"neighbourhood", "UteamupAssetDependencyNeighbourhood", "", "by-asset/{assetGuid}/neighbourhood"},
	}

	for _, c := range cases {
		a := findAction(d, c.name)
		if a == nil {
			t.Errorf("expected %s action on asset-dependency", c.name)
			continue
		}
		if a.ToolName != c.tool {
			t.Errorf("%s: tool = %q, want %q", c.name, a.ToolName, c.tool)
		}
		if a.HTTPMethod != c.method || a.RESTPath != c.path {
			t.Errorf("%s: route = %s %q, want %s %q", c.name, a.HTTPMethod, a.RESTPath, c.method, c.path)
		}
	}
	if len(d.Actions) != len(cases) {
		t.Errorf("asset-dependency has %d actions, want %d", len(d.Actions), len(cases))
	}
}

func TestAssetDependencyRoutesExpandFromTheirArgs(t *testing.T) {
	d := assetDependencyDomain(t)
	cases := []struct {
		action string
		args   map[string]any
		want   string
	}{
		{"list", map[string]any{"assetGuid": "a-1"}, "/api/assetdependency/by-asset/a-1"},
		{"update", map[string]any{"dependencyGuid": "d-1"}, "/api/assetdependency/d-1"},
		{"delete", map[string]any{"dependencyGuid": "d-1"}, "/api/assetdependency/d-1"},
		{"impact", map[string]any{"assetGuid": "a-1"}, "/api/assetdependency/by-asset/a-1/impact"},
		{"neighbourhood", map[string]any{"assetGuid": "a-1"}, "/api/assetdependency/by-asset/a-1/neighbourhood"},
	}

	for _, c := range cases {
		action := findAction(d, c.action)
		if action == nil {
			t.Fatalf("missing asset-dependency action %q", c.action)
		}
		got, consumed := buildRESTPath(d, *action, c.args)
		if got != c.want {
			t.Errorf("%s path = %q, want %q", c.action, got, c.want)
		}
		if len(consumed) != 1 {
			t.Errorf("%s consumed = %v, want one path arg", c.action, consumed)
		}
	}
}

// A mutual relation is two rows, each with its own impact — there is no bidirectional
// flag. Losing --create-reverse would silently make every mutual link one-way.
func TestAssetDependencyCreateCarriesTheEdgeAndItsMirror(t *testing.T) {
	d := assetDependencyDomain(t)
	create := findAction(d, "create")
	if create == nil {
		t.Fatal("expected create action on asset-dependency")
	}

	flags := flagsToMap(create.Flags)
	for name, want := range map[string]struct {
		bodyName string
		typ      string
		required bool
	}{
		"asset-group-guid":     {"assetGroupGuid", "non-empty-uuid", false},
		"source-asset-guid":    {"sourceAssetGuid", "non-empty-uuid", true},
		"target-asset-guid":    {"targetAssetGuid", "non-empty-uuid", true},
		"dependency-type":      {"dependencyType", "string", false},
		"failure-impact":       {"failureImpact", "string", false},
		"impact-delay-minutes": {"impactDelayMinutes", "int", false},
		"create-reverse":       {"createReverse", "string", false},
	} {
		flag, ok := flags[name]
		if !ok {
			t.Errorf("missing --%s on asset-dependency create", name)
			continue
		}
		if flag.BodyName != want.bodyName {
			t.Errorf("--%s BodyName = %q, want %q", name, flag.BodyName, want.bodyName)
		}
		if flag.Type != want.typ {
			t.Errorf("--%s Type = %q, want %q", name, flag.Type, want.typ)
		}
		if flag.Required != want.required {
			t.Errorf("--%s Required = %v, want %v", name, flag.Required, want.required)
		}
	}
	if !flags["create-reverse"].JSONFile {
		t.Error("--create-reverse must be a JSON file flag: the mirror carries its own impact")
	}
}

// None is the safe default: an edge whose impact was never stated must not silently
// propagate a stopped failure through the simulation.
func TestAssetDependencyCreateDefaultsToTheHarmlessImpact(t *testing.T) {
	d := assetDependencyDomain(t)
	create := findAction(d, "create")
	if create == nil {
		t.Fatal("expected create action on asset-dependency")
	}

	flags := flagsToMap(create.Flags)
	if flags["failure-impact"].Default != "None" {
		t.Errorf("failure-impact default = %v, want None", flags["failure-impact"].Default)
	}
	if flags["dependency-type"].Default != "DependsOn" {
		t.Errorf("dependency-type default = %v, want DependsOn", flags["dependency-type"].Default)
	}
}

func TestAssetDependencyReadsSendTheirBoundsOnTheQueryString(t *testing.T) {
	d := assetDependencyDomain(t)
	for action, want := range map[string]map[string]string{
		"impact":        {"propagate-degraded": "propagateDegraded", "max-depth": "maxDepth"},
		"neighbourhood": {"depth": "depth"},
	} {
		found := findAction(d, action)
		if found == nil {
			t.Fatalf("missing asset-dependency action %q", action)
		}
		flags := flagsToMap(found.Flags)
		for name, queryName := range want {
			flag, ok := flags[name]
			if !ok {
				t.Errorf("%s: missing --%s", action, name)
				continue
			}
			if flag.QueryName != queryName {
				t.Errorf("%s: --%s QueryName = %q, want %q", action, name, flag.QueryName, queryName)
			}
		}
	}
}
