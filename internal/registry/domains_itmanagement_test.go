package registry

import "testing"

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
		"connections":   "UteamupITManagementConnectionsList",
		"resources":     "UteamupITManagementResourcesList",
		"servicemap":    "UteamupITManagementServiceMapSummary",
		"flows":         "UteamupITManagementFlowsSummary",
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

func TestITManagementAlertsUseGuidFiltersAndBoundedLimit(t *testing.T) {
	domain := findDomain("itmanagement")
	var alerts Action
	for _, action := range domain.Actions {
		if action.Name == "alerts" {
			alerts = action
		}
	}
	flags := map[string]FlagDef{}
	for _, flag := range alerts.Flags {
		flags[flag.Name] = flag
	}
	for _, name := range []string{"asset-guid", "rule-guid"} {
		if flags[name].Type != "string" {
			t.Errorf("%s must be a GUID string flag", name)
		}
		if flags[name].BodyName == "" {
			t.Errorf("%s must map to its camelCase MCP argument", name)
		}
	}
	if flags["limit"].BodyName != "pageSize" || flags["limit"].Default != 50 {
		t.Errorf("limit must map to pageSize with default 50, got %+v", flags["limit"])
	}
	for _, name := range []string{"alerts", "alert", "alert-ack", "alert-resolve", "resources", "flows"} {
		for _, action := range domain.Actions {
			if action.Name != name {
				continue
			}
			for _, flag := range action.Flags {
				if flag.Name == "id" || flag.BodyName == "id" {
					t.Errorf("%s exposes an integer id flag", name)
				}
			}
		}
	}
}

func TestITManagementMutationsRequireTheAlertGuid(t *testing.T) {
	domain := findDomain("itmanagement")
	for _, action := range domain.Actions {
		if action.Name != "alert-ack" && action.Name != "alert-resolve" && action.Name != "alert" {
			continue
		}
		found := false
		for _, flag := range action.Flags {
			if flag.Name == "alert-guid" && flag.Required && flag.BodyName == "alertGuid" {
				found = true
			}
		}
		if !found {
			t.Errorf("%s must require --alert-guid mapped to alertGuid", action.Name)
		}
	}
}
