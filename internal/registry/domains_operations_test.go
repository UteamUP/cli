package registry

import "testing"

func TestOperationalRouteOptimizeUsesPublicGuid(t *testing.T) {
	domain := findDomain("route")
	if domain == nil {
		t.Fatal("route domain is not registered")
	}

	var optimize *Action
	for index := range domain.Actions {
		if domain.Actions[index].Name == "optimize" {
			optimize = &domain.Actions[index]
			break
		}
	}
	if optimize == nil {
		t.Fatal("route optimize action is not registered")
	}
	if optimize.ToolName != "UteamupOperationalRouteOptimize" {
		t.Fatalf("tool = %q", optimize.ToolName)
	}
	if len(optimize.Args) != 1 ||
		optimize.Args[0].Name != "routeGuid" ||
		optimize.Args[0].Type != "string" {
		t.Fatalf("optimize must use one public route GUID: %+v", optimize.Args)
	}
}
