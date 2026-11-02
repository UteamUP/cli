package registry

import "testing"

func TestPartStockLevelsUsesGuidAndRESTForHumanAndKeySessions(t *testing.T) {
	for _, domain := range DefaultRegistry.Domains() {
		if domain.Name != "part" {
			continue
		}
		for _, action := range domain.Actions {
			if action.Name != "stock-levels" {
				continue
			}
			if action.ToolName != "UteamupPartGetStockLevels" || action.MCPOnly || action.HTTPMethod != "GET" || len(action.Args) != 1 {
				t.Fatalf("stock levels action has the wrong transport: %+v", action)
			}
			arg := action.Args[0]
			if arg.Name != "partGuid" || arg.Type != "non-empty-uuid" || !arg.Required {
				t.Fatalf("stock levels must require a public part GUID: %+v", arg)
			}
			guid := "11111111-1111-4111-8111-111111111111"
			path, consumed := buildRESTPath(domain, action, map[string]any{"partGuid": guid})
			if path != "/api/part/by-guid/"+guid+"/stock-levels" || len(consumed) != 1 {
				t.Fatalf("wrong stock-levels route: %q %+v", path, consumed)
			}
			return
		}
	}
	t.Fatal("part stock-levels was not registered")
}
