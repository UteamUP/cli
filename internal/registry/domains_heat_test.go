package registry

import "testing"

func TestHeatDomainRegistered(t *testing.T) {
	d := findDomain("heat")
	if d == nil {
		t.Fatal("expected heat domain to be registered")
	}
	if d.APIPath != "/api/heatexposurereading" {
		t.Fatalf("heat APIPath = %q, want /api/heatexposurereading", d.APIPath)
	}
}

func TestHeatDomainActions(t *testing.T) {
	d := findDomain("heat")
	if d == nil {
		t.Fatal("expected heat domain to be registered")
	}

	expected := map[string]string{
		"list":   "UteamupHeatexposureList",
		"create": "UteamupHeatexposureCreate",
	}

	actionMap := make(map[string]string)
	for _, a := range d.Actions {
		actionMap[a.Name] = a.ToolName
	}

	for name, tool := range expected {
		if actual, ok := actionMap[name]; !ok {
			t.Errorf("missing action %q", name)
		} else if actual != tool {
			t.Errorf("action %q: expected tool %q, got %q", name, tool, actual)
		}
	}
}
