package registry

import (
	"reflect"
	"strings"
	"testing"
)

func TestShiftHandoverItemLinkActionsUseGuidFirstContracts(t *testing.T) {
	domain := findRegisteredDomain(t, "shift-handover")
	tests := []struct {
		name       string
		toolName   string
		httpMethod string
		restPath   string
		argNames   []string
	}{
		{"section-links", "UteamupShiftHandoverGetSectionLinks", "GET", "by-guid/{handoverGuid}/sections/{sectionGuid}/links", []string{"handoverGuid", "sectionGuid"}},
		{"section-link-create", "UteamupShiftHandoverCreateSectionLink", "POST", "by-guid/{handoverGuid}/sections/{sectionGuid}/links", []string{"handoverGuid", "sectionGuid"}},
		{"section-link-delete", "UteamupShiftHandoverDeleteSectionLink", "DELETE", "by-guid/{handoverGuid}/sections/{sectionGuid}/links/{linkGuid}", []string{"handoverGuid", "sectionGuid", "linkGuid"}},
		{"carryover-links", "UteamupShiftHandoverGetCarryOverLinks", "GET", "by-guid/{handoverGuid}/carryovers/{carryOverGuid}/links", []string{"handoverGuid", "carryOverGuid"}},
		{"carryover-link-create", "UteamupShiftHandoverCreateCarryOverLink", "POST", "by-guid/{handoverGuid}/carryovers/{carryOverGuid}/links", []string{"handoverGuid", "carryOverGuid"}},
		{"carryover-link-delete", "UteamupShiftHandoverDeleteCarryOverLink", "DELETE", "by-guid/{handoverGuid}/carryovers/{carryOverGuid}/links/{linkGuid}", []string{"handoverGuid", "carryOverGuid", "linkGuid"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			action := findShiftHandoverAction(t, domain, test.name)
			if action.ToolName != test.toolName {
				t.Errorf("ToolName = %q, want %q", action.ToolName, test.toolName)
			}
			if action.HTTPMethod != test.httpMethod || action.RESTPath != test.restPath {
				t.Errorf("route = %s %s, want %s %s", action.HTTPMethod, action.RESTPath, test.httpMethod, test.restPath)
			}
			if len(action.Args) != len(test.argNames) {
				t.Fatalf("args = %+v, want %v", action.Args, test.argNames)
			}
			for index, name := range test.argNames {
				arg := action.Args[index]
				if arg.Name != name || arg.Type != "uuid" || !arg.Required {
					t.Errorf("arg[%d] = %+v, want required UUID %q", index, arg, name)
				}
			}
			if strings.Contains(action.RESTPath, "{id}") || strings.Contains(action.RESTPath, "Id}") {
				t.Errorf("RESTPath exposes an integer identity placeholder: %q", action.RESTPath)
			}
		})
	}
}

func TestShiftHandoverItemLinkRoutesConsumeEveryGuidArgument(t *testing.T) {
	domain := findRegisteredDomain(t, "shift-handover")
	handoverGuid := "11111111-2222-4333-8444-555555555555"
	sectionGuid := "aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee"
	carryOverGuid := "99999999-8888-4777-8666-555555555555"
	linkGuid := "12345678-1234-4123-8123-123456789abc"
	tests := []struct {
		name         string
		args         map[string]any
		wantPath     string
		wantConsumed []string
	}{
		{"section-links", map[string]any{"handoverGuid": handoverGuid, "sectionGuid": sectionGuid}, "/api/shifthandover/by-guid/" + handoverGuid + "/sections/" + sectionGuid + "/links", []string{"handoverGuid", "sectionGuid"}},
		{"section-link-create", map[string]any{"handoverGuid": handoverGuid, "sectionGuid": sectionGuid}, "/api/shifthandover/by-guid/" + handoverGuid + "/sections/" + sectionGuid + "/links", []string{"handoverGuid", "sectionGuid"}},
		{"section-link-delete", map[string]any{"handoverGuid": handoverGuid, "sectionGuid": sectionGuid, "linkGuid": linkGuid}, "/api/shifthandover/by-guid/" + handoverGuid + "/sections/" + sectionGuid + "/links/" + linkGuid, []string{"handoverGuid", "sectionGuid", "linkGuid"}},
		{"carryover-links", map[string]any{"handoverGuid": handoverGuid, "carryOverGuid": carryOverGuid}, "/api/shifthandover/by-guid/" + handoverGuid + "/carryovers/" + carryOverGuid + "/links", []string{"handoverGuid", "carryOverGuid"}},
		{"carryover-link-create", map[string]any{"handoverGuid": handoverGuid, "carryOverGuid": carryOverGuid}, "/api/shifthandover/by-guid/" + handoverGuid + "/carryovers/" + carryOverGuid + "/links", []string{"handoverGuid", "carryOverGuid"}},
		{"carryover-link-delete", map[string]any{"handoverGuid": handoverGuid, "carryOverGuid": carryOverGuid, "linkGuid": linkGuid}, "/api/shifthandover/by-guid/" + handoverGuid + "/carryovers/" + carryOverGuid + "/links/" + linkGuid, []string{"handoverGuid", "carryOverGuid", "linkGuid"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			action := findShiftHandoverAction(t, domain, test.name)
			path, consumed := buildRESTPath(domain, *action, test.args)
			if path != test.wantPath {
				t.Errorf("path = %q, want %q", path, test.wantPath)
			}
			if !reflect.DeepEqual(consumed, test.wantConsumed) {
				t.Errorf("consumed = %v, want %v", consumed, test.wantConsumed)
			}
		})
	}
}

func TestShiftHandoverItemLinkMutationFlagsMatchBackendModels(t *testing.T) {
	domain := findRegisteredDomain(t, "shift-handover")
	for _, name := range []string{"section-link-create", "carryover-link-create"} {
		action := findShiftHandoverAction(t, domain, name)
		assertItemLinkFlag(t, action, "linked-entity-type", "linkedEntityType", "string")
		assertItemLinkFlag(t, action, "linked-entity-guid", "linkedEntityGuid", "uuid")
		assertItemLinkFlag(t, action, "concurrency-token", "concurrencyToken", "string")
	}
	for _, name := range []string{"section-link-delete", "carryover-link-delete"} {
		action := findShiftHandoverAction(t, domain, name)
		if len(action.Flags) != 1 {
			t.Fatalf("%s flags = %+v, want only concurrency token", name, action.Flags)
		}
		assertItemLinkFlag(t, action, "concurrency-token", "concurrencyToken", "string")
	}
}

func assertItemLinkFlag(t *testing.T, action *Action, name, bodyName, flagType string) {
	t.Helper()
	for _, flag := range action.Flags {
		if flag.Name != name {
			continue
		}
		if flag.BodyName != bodyName || flag.Type != flagType || !flag.Required || flag.HeaderName != "" {
			t.Fatalf("%s flag %q = %+v, want required body %q type %q", action.Name, name, flag, bodyName, flagType)
		}
		return
	}
	t.Fatalf("%s flag %q missing", action.Name, name)
}
