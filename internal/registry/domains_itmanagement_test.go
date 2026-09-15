package registry

import "testing"

func itManagementAction(t *testing.T, name string) Action {
	t.Helper()
	domain := findDomain("itmanagement")
	if domain == nil {
		t.Fatal("expected itmanagement domain to be registered")
	}
	for _, action := range domain.Actions {
		if action.Name == name {
			return action
		}
	}
	t.Fatalf("missing itmanagement action %q", name)
	return Action{}
}

func TestITManagementDomainMirrorsMCPTools(t *testing.T) {
	domain := findDomain("itmanagement")
	if domain == nil {
		t.Fatal("expected itmanagement domain to be registered")
	}
	expected := map[string]string{
		"dashboard":     "UteamupITManagementDashboard",
		"alerts":        "UteamupITManagementAlertsList",
		"alert":         "UteamupITManagementAlertGet",
		"alert-ack":     "UteamupITManagementAlertAcknowledge",
		"alert-resolve": "UteamupITManagementAlertResolve",
		"rules":         "UteamupITManagementRulesList",
		"connections":         "UteamupITManagementConnectionsList",
		"connections-quota":   "UteamupITManagementConnectionsQuota",
		"resources":     "UteamupITManagementResourcesList",
		"servicemap":    "UteamupITManagementServiceMapSummary",
		"flows":         "UteamupITManagementFlowsSummary",
		"activity":      "UteamupITManagementActivityLog",
	}
	for _, action := range domain.Actions {
		if tool, ok := expected[action.Name]; ok {
			if action.ToolName != tool {
				t.Errorf("action %q maps to %q, want %q", action.Name, action.ToolName, tool)
			}
			delete(expected, action.Name)
		}
	}
	for missing := range expected {
		t.Errorf("missing itmanagement action %q", missing)
	}
	for _, alias := range []string{"it", "it-management"} {
		found := false
		for _, a := range domain.Aliases {
			if a == alias {
				found = true
			}
		}
		if !found {
			t.Errorf("alias %q should be declared on the itmanagement domain", alias)
		}
	}
}

// Every action must land on a real controller route. The released 2.26.0 derived
// GET /api/itmanagement for all of them, which is not a route, so each command returned 404.
func TestITManagementActionsRouteToTheBackendControllers(t *testing.T) {
	domain := findDomain("itmanagement")
	const guid = "11111111-2222-3333-4444-555555555555"
	cases := []struct {
		action string
		method string
		path   string
		args   map[string]any
	}{
		{"dashboard", "GET", "/api/itmanagement/dashboard", map[string]any{}},
		{"alerts", "GET", "/api/itmanagement/alerts", map[string]any{"state": "open"}},
		{"alert", "GET", "/api/itmanagement/alerts/" + guid, map[string]any{"alertGuid": guid}},
		{"alert-ack", "POST", "/api/itmanagement/alerts/" + guid + "/acknowledge", map[string]any{"alertGuid": guid, "reasonCode": "fixed"}},
		{"alert-resolve", "POST", "/api/itmanagement/alerts/" + guid + "/resolve", map[string]any{"alertGuid": guid, "reasonCode": "fixed"}},
		{"rules", "GET", "/api/itmanagement/rules", map[string]any{}},
		{"connections", "GET", "/api/itmanagement/connections", map[string]any{}},
		{"connections-quota", "GET", "/api/itmanagement/connections/quota", map[string]any{}},
		{"resources", "GET", "/api/itmanagement/resources", map[string]any{"mapped": false}},
		{"servicemap", "GET", "/api/itmanagement/servicemap", map[string]any{"windowHours": 24}},
		{"flows", "GET", "/api/itmanagement/network/flows", map[string]any{"windowHours": 24}},
		{"activity", "GET", "/api/itmanagement/logs/activity", map[string]any{"failedOnly": true, "rowLimit": 200}},
	}
	for _, tc := range cases {
		t.Run(tc.action, func(t *testing.T) {
			action := itManagementAction(t, tc.action)
			if action.HTTPMethod != tc.method {
				t.Errorf("method = %q, want %q", action.HTTPMethod, tc.method)
			}
			path, consumed := buildRESTPath(domain, action, tc.args)
			if path != tc.path {
				t.Errorf("path = %q, want %q", path, tc.path)
			}
			if _, usesGuid := tc.args["alertGuid"]; usesGuid && (len(consumed) != 1 || consumed[0] != "alertGuid") {
				t.Errorf("consumed = %v, want the alert GUID removed from the body", consumed)
			}
		})
	}
}

func TestITManagementQueryFlagsUseTheBackendParameterNames(t *testing.T) {
	want := map[string]map[string]string{
		"alerts":    {"severity": "minimumSeverity", "asset-guid": "assetGuid", "rule-guid": "ruleGuid", "limit": "pageSize"},
		"resources": {"include-stale": "includeStale", "type": "resourceType", "asset-guid": "assetGuid", "limit": "pageSize"},
		"flows":     {"asset-guid": "assetGuid", "window-hours": "windowHours", "denied-only": "deniedOnly", "limit": "pageSize"},
		"activity":  {"from": "fromUtc", "to": "toUtc", "failed-only": "failedOnly", "search": "textContains", "resource-group": "resourceGroup", "subscription-id": "subscriptionId", "limit": "rowLimit"},
	}
	for actionName, flags := range want {
		action := itManagementAction(t, actionName)
		byName := map[string]FlagDef{}
		for _, flag := range action.Flags {
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
	if flag := itManagementAction(t, "alerts").Flags; flag[len(flag)-2].Default != 50 {
		t.Errorf("alerts --limit must default to 50")
	}
	activity := itManagementAction(t, "activity")
	if activity.MCPOnly {
		t.Errorf("activity must use the REST logs/activity route like the other IT actions")
	}
	if flag := activity.Flags; len(flag) != 7 {
		t.Errorf("activity flags = %+v, want the seven ITActivityLogQuery filters", flag)
	} else if limit, isInt := flag[6].Default.(int); flag[6].Name != "limit" || !isInt || limit != 200 {
		t.Errorf("activity --limit must default to int 200, got %+v", flag[6])
	}
}

func TestITManagementAlertActionsTakeAValidatedAlertGuid(t *testing.T) {
	for _, name := range []string{"alert", "alert-ack", "alert-resolve"} {
		action := itManagementAction(t, name)
		if len(action.Args) != 1 {
			t.Fatalf("%s args = %+v, want one alert GUID", name, action.Args)
		}
		arg := action.Args[0]
		if arg.Name != "alert-guid" || arg.BodyName != "alertGuid" || arg.Type != "non-empty-uuid" || !arg.Required {
			t.Errorf("%s alert selector = %+v, want a required non-empty GUID mapped to alertGuid", name, arg)
		}
	}
}

// The asset page asks which Azure resource an asset is linked to: the GUID is a query filter, never a path segment.
func TestITManagementResourcesAssetGuidIsAQueryFilter(t *testing.T) {
	domain := findDomain("itmanagement")
	action := itManagementAction(t, "resources")
	const guid = "11111111-2222-3333-4444-555555555555"
	for _, flag := range action.Flags {
		if flag.Name == "asset-guid" && flag.Type != "string" {
			t.Errorf("--asset-guid type = %q, want string like the alerts flag", flag.Type)
		}
	}
	path, consumed := buildRESTPath(domain, action, map[string]any{"assetGuid": guid})
	if path != "/api/itmanagement/resources" || len(consumed) != 0 {
		t.Errorf("path = %q, consumed = %v, want the asset GUID left for the query string", path, consumed)
	}
}
