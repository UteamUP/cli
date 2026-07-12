package registry

import "testing"

func TestIoTDomainMirrorsMCPTools(t *testing.T) {
	domain := findDomain("iot")
	if domain == nil {
		t.Fatal("expected iot domain to be registered")
	}
	expected := map[string]string{
		"status":     "UteamupIoTEnvironmentStatus",
		"monitoring": "UteamupIoTMonitoringDashboard",
		"telemetry":  "UteamupIoTTelemetryPoints",
		"rules":      "UteamupIoTRulesList",
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
		t.Errorf("missing iot action %q", missing)
	}
}

func TestIoTTelemetryUsesGuidFiltersAndBoundedLimit(t *testing.T) {
	domain := findDomain("iot")
	var telemetry Action
	for _, action := range domain.Actions {
		if action.Name == "telemetry" {
			telemetry = action
		}
	}
	flags := map[string]FlagDef{}
	for _, flag := range telemetry.Flags {
		flags[flag.Name] = flag
	}
	for _, name := range []string{"device-guid", "asset-guid", "attribute-definition-guid", "before-point-guid"} {
		if flags[name].Type != "string" {
			t.Errorf("%s must be a GUID string flag", name)
		}
	}
	if flags["limit"].Default != 100 {
		t.Errorf("telemetry limit default = %v, want 100", flags["limit"].Default)
	}
}
