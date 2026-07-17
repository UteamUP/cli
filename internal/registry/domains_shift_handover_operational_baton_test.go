package registry

import (
	"reflect"
	"strings"
	"testing"
)

func TestShiftHandoverOperationalBatonMirrorsGuidFirstMcpTool(t *testing.T) {
	domain := findRegisteredDomain(t, "shift-handover")
	action := findShiftHandoverAction(t, domain, "operational-baton")

	if action.ToolName != "UteamupShiftHandoverGetOperationalBaton" {
		t.Errorf(
			"ToolName = %q, want UteamupShiftHandoverGetOperationalBaton",
			action.ToolName,
		)
	}
	if action.HTTPMethod != "GET" ||
		action.RESTPath != "by-guid/{handoverGuid}/operational-baton" {
		t.Errorf(
			"route = %s %s, want GET by-guid/{handoverGuid}/operational-baton",
			action.HTTPMethod,
			action.RESTPath,
		)
	}
	if len(action.Args) != 1 ||
		action.Args[0].Name != "handoverGuid" ||
		action.Args[0].Type != "uuid" ||
		!action.Args[0].Required {
		t.Errorf("args = %+v, want one required handoverGuid UUID", action.Args)
	}
	if len(action.Flags) != 0 {
		t.Errorf("flags = %+v, want a free deterministic read with no flags", action.Flags)
	}
}

func TestShiftHandoverOperationalBatonRouteResolvesAndConsumesGuid(t *testing.T) {
	domain := findRegisteredDomain(t, "shift-handover")
	action := findShiftHandoverAction(t, domain, "operational-baton")
	handoverGUID := "11111111-2222-4333-8444-555555555555"

	path, consumed := buildRESTPath(
		domain,
		*action,
		map[string]any{"handoverGuid": handoverGUID},
	)
	wantPath := "/api/shifthandover/by-guid/" +
		handoverGUID +
		"/operational-baton"

	if path != wantPath {
		t.Errorf("path = %q, want %q", path, wantPath)
	}
	if !reflect.DeepEqual(consumed, []string{"handoverGuid"}) {
		t.Errorf("consumed = %v, want [handoverGuid]", consumed)
	}
	if strings.Contains(path, "{id}") || strings.Contains(path, "{handoverId}") {
		t.Errorf("path exposes an integer identity: %q", path)
	}
}
