package registry

import "testing"

func TestWorkforceCapacityDomainsRegistered(t *testing.T) {
	for _, name := range []string{"capacity", "capacity-scenario"} {
		d := findDomainByName(t, name)
		if d == nil {
			t.Fatalf("expected %q domain to be registered", name)
		}
		if d.Description == "" {
			t.Errorf("%q domain must have a Description", name)
		}
	}
}

func TestCapacityReadinessActionWired(t *testing.T) {
	d := findDomainByName(t, "capacity")
	if d == nil {
		t.Fatal("expected capacity domain")
	}
	if d.APIPath != "/api/workforcecapacity" {
		t.Errorf("unexpected APIPath %q", d.APIPath)
	}
	var readiness *Action
	for i := range d.Actions {
		if d.Actions[i].Name == "readiness" {
			readiness = &d.Actions[i]
		}
	}
	if readiness == nil {
		t.Fatal("expected a readiness action")
	}
	if readiness.HTTPMethod != "GET" || readiness.RESTPath != "readiness" {
		t.Errorf("readiness must be GET readiness, got %s %s", readiness.HTTPMethod, readiness.RESTPath)
	}
}

func TestCapacityScenarioActionsWired(t *testing.T) {
	d := findDomainByName(t, "capacity-scenario")
	if d == nil {
		t.Fatal("expected capacity-scenario domain")
	}
	if d.APIPath != "/api/workforcecapacityscenarios" {
		t.Errorf("unexpected APIPath %q", d.APIPath)
	}
	expected := map[string]string{
		"list":   "UteamupWorkforceCapacityScenarioList",
		"get":    "UteamupWorkforceCapacityScenarioGet",
		"create": "UteamupWorkforceCapacityScenarioCreate",
		"clone":  "UteamupWorkforceCapacityScenarioClone",
		"delete": "UteamupWorkforceCapacityScenarioDelete",
	}
	got := map[string]string{}
	for _, a := range d.Actions {
		got[a.Name] = a.ToolName
	}
	for action, toolName := range expected {
		if got[action] != toolName {
			t.Errorf("expected capacity-scenario action %q to map to %q, got %q", action, toolName, got[action])
		}
	}
}
