package registry

import "testing"

func TestPropertyDamageDomainActions(t *testing.T) {
	d := findDomain("property-damage")
	if d == nil || d.APIPath != "/api/propertydamage" {
		t.Fatal("property damage domain or API path missing")
	}
	tools := map[string]string{
		"list":   "UteamupPropertydamageList",
		"get":    "UteamupPropertydamageGet",
		"create": "UteamupPropertydamageCreate",
		"update": "UteamupPropertydamageUpdate",
	}
	for _, action := range d.Actions {
		if want, ok := tools[action.Name]; ok && action.ToolName != want {
			t.Errorf("%s uses %s, want %s", action.Name, action.ToolName, want)
		}
		delete(tools, action.Name)
	}
	if len(tools) != 0 {
		t.Errorf("missing actions: %v", tools)
	}
}
