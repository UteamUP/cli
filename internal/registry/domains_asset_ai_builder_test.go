package registry

import "testing"

// The asset-ai-builder domain mirrors AssetAiBuilderController
// (/api/asset/{assetGuid}/ai-builder/*). Generate spends AI credits and apply writes
// records, so both must stay POSTs, and apply must send the reviewed JSON as the body root.

func assetAiBuilderDomain(t *testing.T) *Domain {
	t.Helper()
	d := findDomain("asset-ai-builder")
	if d == nil {
		t.Fatal("expected asset-ai-builder domain to be registered")
	}
	return d
}

func TestAssetAiBuilderActionsMatchBackendContract(t *testing.T) {
	d := assetAiBuilderDomain(t)
	if d.APIPath != "/api/asset" {
		t.Errorf("APIPath = %q, want /api/asset", d.APIPath)
	}

	cases := []struct{ name, tool, method, path string }{
		{"readiness", "UteamupAssetAiBuilderReadiness", "", "{assetGuid}/ai-builder/readiness"},
		{"generate", "UteamupAssetAiBuilderGenerate", "POST", "{assetGuid}/ai-builder/generate"},
		{"apply", "UteamupAssetAiBuilderApply", "POST", "{assetGuid}/ai-builder/apply"},
		{"history", "UteamupAssetAiBuilderHistory", "", "{assetGuid}/ai-builder/history"},
	}
	for _, c := range cases {
		a := findAction(d, c.name)
		if a == nil {
			t.Errorf("expected %s action on asset-ai-builder", c.name)
			continue
		}
		if a.ToolName != c.tool {
			t.Errorf("%s: tool = %q, want %q", c.name, a.ToolName, c.tool)
		}
		if a.HTTPMethod != c.method || a.RESTPath != c.path {
			t.Errorf("%s: route = %s %q, want %s %q", c.name, a.HTTPMethod, a.RESTPath, c.method, c.path)
		}
		if len(a.Args) != 1 || a.Args[0].Name != "assetGuid" || !a.Args[0].Required || a.Args[0].Type != "non-empty-uuid" {
			t.Errorf("%s: expected one required non-empty-uuid assetGuid arg, got %+v", c.name, a.Args)
		}
	}
	if len(d.Actions) != len(cases) {
		t.Errorf("asset-ai-builder has %d actions, want %d", len(d.Actions), len(cases))
	}
}

func TestAssetAiBuilderGenerateRequiresKinds(t *testing.T) {
	a := findAction(assetAiBuilderDomain(t), "generate")
	var kinds *FlagDef
	for i := range a.Flags {
		if a.Flags[i].Name == "kinds" {
			kinds = &a.Flags[i]
		}
	}
	if kinds == nil || !kinds.Required || kinds.Type != "stringSlice" {
		t.Fatalf("generate must take a required stringSlice --kinds flag, got %+v", kinds)
	}
}

func TestAssetAiBuilderApplySendsTheReviewedFileAsTheBodyRoot(t *testing.T) {
	a := findAction(assetAiBuilderDomain(t), "apply")
	if len(a.Flags) != 1 {
		t.Fatalf("apply should take exactly one flag, got %d", len(a.Flags))
	}
	f := a.Flags[0]
	if f.Name != "file" || !f.Required || !f.RootJSONObjectFile || f.JSONFile {
		t.Errorf("apply --file must be a required root JSON object file, got %+v", f)
	}
}

func TestAssetAiBuilderTakesNoTenantOrUserIdentityFlags(t *testing.T) {
	for _, a := range assetAiBuilderDomain(t).Actions {
		for _, f := range a.Flags {
			if f.Name == "tenant-guid" || f.Name == "user-id" {
				t.Errorf("%s: identity comes from the login, not from --%s", a.Name, f.Name)
			}
		}
	}
}
