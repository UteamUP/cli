package registry

import "testing"

func TestAgentSkillDomainRoutesThroughMcpTools(t *testing.T) {
	t.Parallel()
	domain := findDomain("agent-skill")
	if domain == nil {
		t.Fatal("agent-skill domain is not registered")
	}
	want := map[string]string{"list": "UteamupAgentSkillList", "get": "UteamupAgentSkillGet"}
	if len(domain.Actions) != len(want) {
		t.Fatalf("actions = %d, want %d", len(domain.Actions), len(want))
	}
	for _, action := range domain.Actions {
		if want[action.Name] != action.ToolName || !action.MCPOnly {
			t.Fatalf("action contract = %+v", action)
		}
		if action.Name == "get" && (len(action.Args) != 1 || action.Args[0].Name != "name" || !action.Args[0].Required) {
			t.Fatalf("get must require the skill name: %+v", action.Args)
		}
	}
}
