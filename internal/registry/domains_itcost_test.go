package registry

import "testing"

func itCostAction(t *testing.T, name string) Action {
	t.Helper()
	domain := findDomain("itcost")
	if domain == nil {
		t.Fatal("expected itcost domain to be registered")
	}
	for _, action := range domain.Actions {
		if action.Name == name {
			return action
		}
	}
	t.Fatalf("missing itcost action %q", name)
	return Action{}
}

func TestITCostDomainMirrorsMCPTools(t *testing.T) {
	expected := map[string]string{
		"summary":          "UteamupITManagementCostSummary",
		"breakdown":        "UteamupITManagementCostBreakdown",
		"budgets":          "UteamupITManagementCostBudgetsList",
		"savings":          "UteamupITManagementCostRecommendationsList",
		"saving-workorder": "UteamupITManagementCostRecommendationCreateWorkorder",
		"licences":         "UteamupITManagementLicencesList",
	}
	domain := findDomain("itcost")
	if domain == nil {
		t.Fatal("expected itcost domain to be registered")
	}
	if len(domain.Actions) != len(expected) {
		t.Errorf("itcost has %d actions, want %d: no account or secret commands", len(domain.Actions), len(expected))
	}
	for name, tool := range expected {
		if action := itCostAction(t, name); action.ToolName != tool {
			t.Errorf("%s maps to %q, want %q", name, action.ToolName, tool)
		}
	}
}

func TestITCostActionsRouteToTheBackendControllers(t *testing.T) {
	const guid = "11111111-2222-4333-8444-555555555555"
	domain := findDomain("itcost")
	cases := []struct {
		action string
		method string
		path   string
		args   map[string]any
	}{
		{"summary", "GET", "/api/itmanagement/cost/summary", map[string]any{}},
		{"breakdown", "GET", "/api/itmanagement/cost/breakdown", map[string]any{"groupBy": "service"}},
		{"budgets", "GET", "/api/itmanagement/cost/budgets", map[string]any{}},
		{"savings", "GET", "/api/itmanagement/cost/recommendations", map[string]any{"take": 50}},
		{"saving-workorder", "POST", "/api/itmanagement/cost/recommendations/" + guid + "/create-workorder", map[string]any{"recommendationGuid": guid, "workorderTemplateGuid": guid}},
		{"licences", "GET", "/api/itmanagement/cost/licences", map[string]any{}},
	}
	for _, tc := range cases {
		t.Run(tc.action, func(t *testing.T) {
			action := itCostAction(t, tc.action)
			if action.HTTPMethod != tc.method {
				t.Errorf("method = %q, want %q", action.HTTPMethod, tc.method)
			}
			path, consumed := buildRESTPath(domain, action, tc.args)
			if path != tc.path {
				t.Errorf("path = %q, want %q", path, tc.path)
			}
			if tc.action == "saving-workorder" && (len(consumed) != 1 || consumed[0] != "recommendationGuid") {
				t.Errorf("consumed = %v, want the saving GUID taken out of the body", consumed)
			}
		})
	}
}

func TestITCostFlagsUseTheBackendParameterNames(t *testing.T) {
	want := map[string]map[string]string{
		"breakdown":        {"group-by": "groupBy", "tag-key": "tagKey", "account-guid": "accountGuid"},
		"savings":          {"status": "statuses", "limit": "take"},
		"saving-workorder": {"template-guid": "workorderTemplateGuid"},
	}
	for actionName, flags := range want {
		byName := map[string]FlagDef{}
		for _, flag := range itCostAction(t, actionName).Flags {
			byName[flag.Name] = flag
			if flag.Name == "id" || flag.BodyName == "id" {
				t.Errorf("%s exposes an integer id flag", actionName)
			}
		}
		for flag, body := range flags {
			if byName[flag].BodyName != body {
				t.Errorf("%s --%s sends %q, backend binds %q", actionName, flag, byName[flag].BodyName, body)
			}
		}
	}
}

func TestAssetTcoRoutesByGuid(t *testing.T) {
	const guid = "11111111-2222-4333-8444-555555555555"
	domain := findDomain("asset")
	var tco *Action
	for i := range domain.Actions {
		if domain.Actions[i].Name == "tco" {
			tco = &domain.Actions[i]
		}
	}
	if tco == nil || tco.ToolName != "UteamupAssetTco" {
		t.Fatal("asset tco must mirror the uteamup_asset_tco MCP tool")
	}
	path, _ := buildRESTPath(domain, *tco, map[string]any{"assetGuid": guid, "months": 12})
	if path != "/api/asset/"+guid+"/tco" {
		t.Errorf("path = %q", path)
	}
}
