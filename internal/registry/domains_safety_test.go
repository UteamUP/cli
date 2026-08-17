package registry

import "testing"

func TestSafetyDomainRegistered(t *testing.T) {
	d := findDomain("safety")
	if d == nil {
		t.Fatal("expected safety domain to be registered")
	}
	if d.APIPath != "/api/safetyincident" {
		t.Fatalf("safety APIPath = %q, want /api/safetyincident", d.APIPath)
	}
}

func TestSafetyDomainActions(t *testing.T) {
	d := findDomain("safety")
	if d == nil {
		t.Fatal("expected safety domain to be registered")
	}

	expected := map[string]string{
		"list":       "UteamupSafetyincidentList",
		"get":        "UteamupSafetyincidentGet",
		"create":     "UteamupSafetyincidentCreate",
		"classify":   "UteamupSafetyincidentClassify",
		"ita-export": "UteamupOshaItaExport",
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

func TestSafetyGetUsesGuidPath(t *testing.T) {
	d := findDomain("safety")
	if d == nil {
		t.Fatal("expected safety domain to be registered")
	}
	action := findAction(d, "get")
	if action == nil {
		t.Fatal("expected get action on safety domain")
	}
	path, _ := buildRESTPath(d, *action, map[string]any{
		"guid": "11111111-1111-1111-1111-111111111111",
	})
	if path != "/api/safetyincident/by-guid/11111111-1111-1111-1111-111111111111" {
		t.Fatalf("safety get path = %q", path)
	}
}
