package registry

import (
	"reflect"
	"strings"
	"testing"
)

func TestShiftHandoverCriticalItemActionsMirrorGuidContracts(t *testing.T) {
	domain := findRegisteredDomain(t, "shift-handover")
	tests := []struct {
		name      string
		tool      string
		path      string
		itemArg   string
		itemValue string
	}{
		{"section-critical-acknowledge", "UteamupShiftHandoverAcknowledgeCriticalSection", "by-guid/{handoverGuid}/sections/{sectionGuid}/acknowledge-critical", "sectionGuid", "22222222-2222-4222-8222-222222222222"},
		{"carryover-critical-acknowledge", "UteamupShiftHandoverAcknowledgeCriticalCarryOver", "by-guid/{handoverGuid}/carryovers/{carryOverGuid}/acknowledge-critical", "carryOverGuid", "33333333-3333-4333-8333-333333333333"},
	}
	handoverGuid := "11111111-1111-4111-8111-111111111111"

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			action := findShiftHandoverAction(t, domain, test.name)
			if action.ToolName != test.tool || action.HTTPMethod != "PUT" || action.RESTPath != test.path {
				t.Fatalf("action = %+v", action)
			}
			if len(action.Args) != 2 || action.Args[0].Type != "uuid" || action.Args[1].Type != "uuid" {
				t.Fatalf("args must be two GUIDs: %+v", action.Args)
			}
			path, consumed := buildRESTPath(domain, *action, map[string]any{
				"handoverGuid": handoverGuid,
				test.itemArg:   test.itemValue,
			})
			if path != "/api/shifthandover/"+replaceGuidPlaceholders(test.path, handoverGuid, test.itemValue) {
				t.Fatalf("path = %q", path)
			}
			if !reflect.DeepEqual(consumed, []string{"handoverGuid", test.itemArg}) {
				t.Fatalf("consumed = %v", consumed)
			}
			assertCriticalItemFlag(t, action, "notes", "notes", "string", false)
			assertCriticalItemFlag(t, action, "concurrency-token", "concurrencyToken", "string", true)
		})
	}
}

func replaceGuidPlaceholders(path, handoverGuid, itemGuid string) string {
	result := path
	result = strings.ReplaceAll(result, "{handoverGuid}", handoverGuid)
	result = strings.ReplaceAll(result, "{sectionGuid}", itemGuid)
	return strings.ReplaceAll(result, "{carryOverGuid}", itemGuid)
}

func assertCriticalItemFlag(
	t *testing.T,
	action *Action,
	name, bodyName, flagType string,
	required bool,
) {
	t.Helper()
	for _, flag := range action.Flags {
		if flag.Name == name {
			actualBodyName := flag.BodyName
			if actualBodyName == "" {
				actualBodyName = toCamelCase(flag.Name)
			}
			if actualBodyName != bodyName || flag.Type != flagType || flag.Required != required {
				t.Fatalf("flag = %+v", flag)
			}
			return
		}
	}
	t.Fatalf("flag %q missing", name)
}
