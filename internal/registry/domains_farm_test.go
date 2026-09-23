package registry

import "testing"

func TestFarmUnitDomainUsesGuidAndExactRESTPath(t *testing.T) {
	domain := findDomain("farm-unit")
	if domain == nil {
		t.Fatal("farm-unit domain is not registered")
	}
	if domain.APIPath != "/api/farm/units" {
		t.Fatalf("APIPath = %q, want /api/farm/units", domain.APIPath)
	}
	if got := findAction(domain, "list"); got == nil || got.ToolName != "UteamupFarmUnitList" {
		t.Errorf("list action must mirror the farm unit MCP tool")
	}
	get := findAction(domain, "get")
	if get == nil || get.ToolName != "UteamupFarmUnitGet" {
		t.Fatal("get action must mirror the farm unit MCP tool")
	}
	if len(get.Args) != 1 || get.Args[0].Name != "externalGuid" || get.Args[0].Type != "string" {
		t.Errorf("get action must take one public GUID: %+v", get.Args)
	}
}
