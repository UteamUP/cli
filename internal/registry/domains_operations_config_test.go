package registry

import "testing"

func TestOperationsConfigDomainsRouteThroughMcpTools(t *testing.T) {
	t.Parallel()
	want := map[string]map[string]string{
		"inspection-template": {"list": "UteamupInspectionTemplateList", "get": "UteamupInspectionTemplateGet"},
		"workorder-prefix":    {"list": "UteamupWorkorderPrefixList", "get": "UteamupWorkorderPrefixGet"},
		"workorder-usage": {
			"tool-list": "UteamupWorkorderToolUsageList", "tool-get": "UteamupWorkorderToolUsageGet",
			"tool-update": "UteamupWorkorderToolUsageUpdate", "energy-list": "UteamupWorkorderEnergyUsageList",
			"energy-get": "UteamupWorkorderEnergyUsageGet", "energy-update": "UteamupWorkorderEnergyUsageUpdate",
		},
		"location-nfc-tag": {"list": "UteamupLocationNfcTagList", "lookup": "UteamupLocationNfcTagGetByTagId"},
	}
	for name, actions := range want {
		domain := findDomain(name)
		if domain == nil {
			t.Fatalf("%s domain is not registered", name)
		}
		if len(domain.Actions) != len(actions) {
			t.Fatalf("%s actions = %d, want %d", name, len(domain.Actions), len(actions))
		}
		for _, action := range domain.Actions {
			if actions[action.Name] != action.ToolName || !action.MCPOnly {
				t.Fatalf("%s action contract = %+v", name, action)
			}
		}
	}
}

func TestOperationsConfigGetActionsRequireAGuidOrTag(t *testing.T) {
	t.Parallel()
	cases := map[string]map[string]string{
		"inspection-template": {"get": "templateGuid"},
		"workorder-prefix":    {"get": "prefixGuid"},
		"workorder-usage":     {"tool-get": "usageGuid", "tool-update": "usageGuid", "energy-get": "usageGuid", "energy-update": "usageGuid"},
		"location-nfc-tag":    {"lookup": "tagId"},
	}
	for name, actions := range cases {
		for _, action := range findDomain(name).Actions {
			body, ok := actions[action.Name]
			if !ok {
				continue
			}
			if len(action.Args) != 1 || action.Args[0].BodyName != body || !action.Args[0].Required {
				t.Fatalf("%s %s must require %s: %+v", name, action.Name, body, action.Args)
			}
		}
	}
}

func TestWorkorderUsageUpdatesRequireTheFullRecordAndNoTenant(t *testing.T) {
	t.Parallel()
	required := map[string][]string{
		"tool-update":   {"workorder-guid", "tool-guid", "start-time"},
		"energy-update": {"workorder-guid", "utility-type", "consumption"},
	}
	for _, action := range findDomain("workorder-usage").Actions {
		flags := map[string]FlagDef{}
		for _, flag := range action.Flags {
			if flag.Name == "tenant-guid" || flag.Name == "tenant-id" {
				t.Fatalf("%s must not take a tenant flag; the tenant comes from the token", action.Name)
			}
			flags[flag.Name] = flag
		}
		for _, name := range required[action.Name] {
			if !flags[name].Required {
				t.Fatalf("%s --%s must be required", action.Name, name)
			}
		}
		if (action.Name == "tool-list" || action.Name == "energy-list") && flags["page-size"].Default != 25 {
			t.Fatalf("%s page-size default = %v, want 25", action.Name, flags["page-size"].Default)
		}
	}
}
