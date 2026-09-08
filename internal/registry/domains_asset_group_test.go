package registry

import "testing"

// The asset-group domain mirrors AssetGroupController and its Impact, Labels and
// Generation partials. Its routes are the backend contract: a typo here ships a command
// that always 404s, and a wrong HTTP method ships one that silently reads instead of
// writing.

func assetGroupDomain(t *testing.T) *Domain {
	t.Helper()
	d := findDomain("asset-group")
	if d == nil {
		t.Fatal("expected asset-group domain to be registered")
	}
	return d
}

func TestAssetGroupDomainRoutesUnderAssetgroup(t *testing.T) {
	d := assetGroupDomain(t)
	if d.APIPath != "/api/assetgroup" {
		t.Errorf("APIPath = %q, want /api/assetgroup", d.APIPath)
	}

	want := map[string]bool{"ag": false, "asset-groups": false}
	for _, alias := range d.Aliases {
		if _, ok := want[alias]; ok {
			want[alias] = true
		}
	}
	for alias, found := range want {
		if !found {
			t.Errorf("expected the %q alias, got %v", alias, d.Aliases)
		}
	}
}

func TestAssetGroupActionsMatchBackendContract(t *testing.T) {
	d := assetGroupDomain(t)
	cases := []struct{ name, tool, method, path string }{
		{"list", "UteamupAssetGroupList", "", ""},
		{"get", "UteamupAssetGroupGet", "", "by-guid/{groupGuid}"},
		{"create", "UteamupAssetGroupCreate", "", ""},
		{"update", "UteamupAssetGroupUpdate", "", "by-guid/{groupGuid}"},
		{"delete", "UteamupAssetGroupDelete", "", "by-guid/{groupGuid}"},
		{"members-add", "UteamupAssetGroupMembersAdd", "POST", "by-guid/{groupGuid}/members"},
		{"members-remove", "UteamupAssetGroupMembersRemove", "DELETE", "by-guid/{groupGuid}/members/{assetGuid}"},
		{"diagram", "UteamupAssetGroupDiagramGet", "", "by-guid/{groupGuid}/diagram"},
		{"positions-save", "UteamupAssetGroupPositionsSave", "PUT", "by-guid/{groupGuid}/diagram/positions"},
		{"auto-layout", "UteamupAssetGroupAutoLayout", "POST", "by-guid/{groupGuid}/diagram/auto-layout"},
		{"map", "UteamupAssetGroupMapGet", "", "map"},
		{"dependencies", "UteamupAssetGroupDependenciesList", "", "by-guid/{groupGuid}/dependencies"},
		{"dependency-create", "UteamupAssetGroupDependenciesCreate", "POST", "dependencies"},
		{"dependency-delete", "UteamupAssetGroupDependenciesDelete", "DELETE", "dependencies/{dependencyGuid}"},
		{"impact", "UteamupAssetGroupImpactAnalyze", "", "impact"},
		{"blast-radius", "UteamupAssetGroupBlastRadius", "", "by-guid/{groupGuid}/blast-radius"},
		{"live-status", "UteamupAssetGroupLiveStatus", "", "by-guid/{groupGuid}/live-status"},
		{"label", "UteamupAssetGroupLabelPayload", "", "by-guid/{groupGuid}/label-payload"},
		{"links-generate", "UteamupAssetGroupLinksGenerate", "POST", "links/generate"},
		{"links-get", "UteamupAssetGroupLinksGenerationGet", "", "links/generations/{generationGuid}"},
		{"links-apply", "UteamupAssetGroupLinksApply", "POST", "links/from-generation/{generationGuid}"},
	}

	for _, c := range cases {
		a := findAction(d, c.name)
		if a == nil {
			t.Errorf("expected %s action on asset-group", c.name)
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
		t.Errorf("asset-group has %d actions, want %d", len(d.Actions), len(cases))
	}
}

// Every identifier on the boundary is a GUID (Guidelines/ApiGuidelines.md). A legacy
// int `id` positional here would route to a path the backend does not serve.
func TestAssetGroupPositionalArgsAreGuidsOnly(t *testing.T) {
	d := assetGroupDomain(t)
	for _, action := range d.Actions {
		for _, arg := range action.Args {
			if arg.Type != "non-empty-uuid" {
				t.Errorf("%s: arg %s has type %q, want non-empty-uuid", action.Name, arg.Name, arg.Type)
			}
		}
	}
}

func TestAssetGroupRoutesExpandFromTheirArgs(t *testing.T) {
	d := assetGroupDomain(t)
	cases := []struct {
		action string
		args   map[string]any
		want   string
	}{
		{"get", map[string]any{"groupGuid": "g-1"}, "/api/assetgroup/by-guid/g-1"},
		{"members-remove", map[string]any{"groupGuid": "g-1", "assetGuid": "a-1"}, "/api/assetgroup/by-guid/g-1/members/a-1"},
		{"auto-layout", map[string]any{"groupGuid": "g-1"}, "/api/assetgroup/by-guid/g-1/diagram/auto-layout"},
		{"dependency-delete", map[string]any{"dependencyGuid": "d-1"}, "/api/assetgroup/dependencies/d-1"},
		{"links-apply", map[string]any{"generationGuid": "x-1"}, "/api/assetgroup/links/from-generation/x-1"},
		{"map", map[string]any{}, "/api/assetgroup/map"},
	}

	for _, c := range cases {
		action := findAction(d, c.action)
		if action == nil {
			t.Fatalf("missing asset-group action %q", c.action)
		}
		got, _ := buildRESTPath(d, *action, c.args)
		if got != c.want {
			t.Errorf("%s path = %q, want %q", c.action, got, c.want)
		}
	}
}

// The group-level simulation is a GET whose subject travels on the query string, not in
// the path — mirroring GET api/assetgroup/impact?failedGroupGuid=...
func TestAssetGroupImpactSendsItsSubjectOnTheQueryString(t *testing.T) {
	d := assetGroupDomain(t)
	impact := findAction(d, "impact")
	if impact == nil {
		t.Fatal("expected impact action on asset-group")
	}
	if len(impact.Args) != 1 || impact.Args[0].QueryName != "failedGroupGuid" {
		t.Fatalf("impact expected one failedGroupGuid query arg, got %+v", impact.Args)
	}

	flags := flagsToMap(impact.Flags)
	for name, queryName := range map[string]string{
		"propagate-degraded": "propagateDegraded",
		"max-depth":          "maxDepth",
	} {
		flag, ok := flags[name]
		if !ok {
			t.Errorf("missing --%s on asset-group impact", name)
			continue
		}
		if flag.QueryName != queryName {
			t.Errorf("--%s QueryName = %q, want %q", name, flag.QueryName, queryName)
		}
	}
}

// The generation is charged against the tenant's AI credits and is idempotent on
// requestGuid. Dropping either flag turns a retry into a second charge.
func TestAssetGroupLinksGenerateCarriesPromptAndIdempotencyKey(t *testing.T) {
	d := assetGroupDomain(t)
	generate := findAction(d, "links-generate")
	if generate == nil {
		t.Fatal("expected links-generate action on asset-group")
	}

	flags := flagsToMap(generate.Flags)
	prompt, ok := flags["prompt"]
	if !ok || !prompt.Required {
		t.Errorf("--prompt must exist and be required, got %+v", prompt)
	}
	request, ok := flags["request-guid"]
	if !ok || !request.Required || request.BodyName != "requestGuid" || request.Type != "non-empty-uuid" {
		t.Errorf("--request-guid must be a required non-empty GUID mapped to requestGuid, got %+v", request)
	}
	documents, ok := flags["document-guid"]
	if !ok || documents.BodyName != "documentGuids" || documents.Type != "stringSlice" {
		t.Errorf("--document-guid must be a repeatable slice mapped to documentGuids, got %+v", documents)
	}
}

// A diagram save is a whole-object payload; flat flags cannot express positions[].
func TestAssetGroupPositionsSaveTakesARootJSONObject(t *testing.T) {
	d := assetGroupDomain(t)
	save := findAction(d, "positions-save")
	if save == nil {
		t.Fatal("expected positions-save action on asset-group")
	}
	if len(save.Flags) != 1 {
		t.Fatalf("positions-save expected exactly one flag, got %+v", save.Flags)
	}
	flag := save.Flags[0]
	if flag.Name != "from-json" || !flag.RootJSONObjectFile || !flag.Required {
		t.Errorf("positions-save expected a required root-object --from-json, got %+v", flag)
	}
	if err := validateActionDefinition(*save); err != nil {
		t.Errorf("positions-save definition is invalid: %v", err)
	}
}
