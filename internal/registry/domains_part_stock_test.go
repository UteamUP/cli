package registry

import "testing"

func TestPartStockLevelsUsesGuidAndMCPTransport(t *testing.T) {
	for _, domain := range DefaultRegistry.Domains() {
		if domain.Name != "part" {
			continue
		}
		for _, action := range domain.Actions {
			if action.Name != "stock-levels" {
				continue
			}
			if action.ToolName != "UteamupPartGetStockLevels" || !action.MCPOnly || len(action.Args) != 1 {
				t.Fatalf("stock levels action has the wrong transport: %+v", action)
			}
			arg := action.Args[0]
			if arg.Name != "partGuid" || arg.Type != "string" || !arg.Required {
				t.Fatalf("stock levels must require a public part GUID: %+v", arg)
			}
			return
		}
	}
	t.Fatal("part stock-levels was not registered")
}
