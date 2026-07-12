package registry

import (
	"reflect"
	"strings"
	"testing"
)

func TestShiftHandoverCarryOverActionsUseGuidFirstContracts(t *testing.T) {
	domain := findRegisteredDomain(t, "shift-handover")

	tests := []struct {
		name       string
		toolName   string
		httpMethod string
		restPath   string
		argNames   []string
	}{
		{"carryovers", "UteamupShiftHandoverGetCarryOvers", "GET", "by-guid/{handoverGuid}/carryovers", []string{"handoverGuid"}},
		{"carryover-create", "UteamupShiftHandoverCreateCarryOver", "POST", "by-guid/{handoverGuid}/carryovers", []string{"handoverGuid"}},
		{"carryover-update", "UteamupShiftHandoverUpdateCarryOver", "PUT", "by-guid/{handoverGuid}/carryovers/{carryOverGuid}", []string{"handoverGuid", "carryOverGuid"}},
		{"carryover-delete", "UteamupShiftHandoverDeleteCarryOver", "DELETE", "by-guid/{handoverGuid}/carryovers/{carryOverGuid}", []string{"handoverGuid", "carryOverGuid"}},
		{"carryover-convert", "UteamupShiftHandoverConvertCarryOverToWorkOrder", "POST", "by-guid/{handoverGuid}/carryovers/{carryOverGuid}/convert-to-workorder", []string{"handoverGuid", "carryOverGuid"}},
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
			if strings.Contains(action.RESTPath, "{id}") ||
				strings.Contains(action.RESTPath, "{handoverId}") ||
				strings.Contains(action.RESTPath, "{carryOverId}") {
				t.Errorf("RESTPath exposes an integer identity placeholder: %q", action.RESTPath)
			}
		})
	}
}

func TestShiftHandoverCarryOverRoutesConsumeGuidArguments(t *testing.T) {
	domain := findRegisteredDomain(t, "shift-handover")
	handoverGuid := "11111111-2222-3333-4444-555555555555"
	carryOverGuid := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"

	tests := []struct {
		name         string
		args         map[string]any
		wantPath     string
		wantConsumed []string
	}{
		{"carryovers", map[string]any{"handoverGuid": handoverGuid}, "/api/shifthandover/by-guid/" + handoverGuid + "/carryovers", []string{"handoverGuid"}},
		{"carryover-create", map[string]any{"handoverGuid": handoverGuid}, "/api/shifthandover/by-guid/" + handoverGuid + "/carryovers", []string{"handoverGuid"}},
		{"carryover-update", map[string]any{"handoverGuid": handoverGuid, "carryOverGuid": carryOverGuid}, "/api/shifthandover/by-guid/" + handoverGuid + "/carryovers/" + carryOverGuid, []string{"handoverGuid", "carryOverGuid"}},
		{"carryover-delete", map[string]any{"handoverGuid": handoverGuid, "carryOverGuid": carryOverGuid}, "/api/shifthandover/by-guid/" + handoverGuid + "/carryovers/" + carryOverGuid, []string{"handoverGuid", "carryOverGuid"}},
		{"carryover-convert", map[string]any{"handoverGuid": handoverGuid, "carryOverGuid": carryOverGuid}, "/api/shifthandover/by-guid/" + handoverGuid + "/carryovers/" + carryOverGuid + "/convert-to-workorder", []string{"handoverGuid", "carryOverGuid"}},
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

func TestShiftHandoverCarryOverMutationFlagsMatchBackend(t *testing.T) {
	domain := findRegisteredDomain(t, "shift-handover")

	assertBodyFlag := func(t *testing.T, action *Action, name, bodyName, flagType string, required bool) {
		t.Helper()
		for _, flag := range action.Flags {
			if flag.Name != name {
				continue
			}
			actualBodyName := flag.BodyName
			if actualBodyName == "" {
				actualBodyName = toCamelCase(flag.Name)
			}
			if actualBodyName != bodyName || flag.Type != flagType || flag.Required != required || flag.HeaderName != "" {
				t.Fatalf("flag %q = %+v (body %q), want body=%q type=%q required=%v", name, flag, actualBodyName, bodyName, flagType, required)
			}
			return
		}
		t.Fatalf("flag %q missing from %s", name, action.Name)
	}

	create := findShiftHandoverAction(t, domain, "carryover-create")
	assertBodyFlag(t, create, "description", "description", "string", true)
	assertBodyFlag(t, create, "priority", "priority", "int", false)
	assertBodyFlag(t, create, "original-handover-guid", "originalShiftHandoverGuid", "uuid", false)
	assertBodyFlag(t, create, "concurrency-token", "concurrencyToken", "string", true)

	update := findShiftHandoverAction(t, domain, "carryover-update")
	assertBodyFlag(t, update, "description", "description", "string", false)
	assertBodyFlag(t, update, "status", "status", "string", false)
	assertBodyFlag(t, update, "priority", "priority", "int", false)
	assertBodyFlag(t, update, "concurrency-token", "concurrencyToken", "string", true)

	deleteAction := findShiftHandoverAction(t, domain, "carryover-delete")
	assertBodyFlag(t, deleteAction, "concurrency-token", "concurrencyToken", "string", true)

	convert := findShiftHandoverAction(t, domain, "carryover-convert")
	assertBodyFlag(t, convert, "concurrency-token", "concurrencyToken", "string", true)
	for _, flag := range convert.Flags {
		if flag.Name == "idempotency-key" && flag.Required && flag.HeaderName == "Idempotency-Key" && flag.BodyName == "" {
			return
		}
	}
	t.Fatal("carryover-convert must require an Idempotency-Key header")
}
