package registry

import (
	"reflect"
	"strings"
	"testing"
)

func TestShiftHandoverAuditActionsMirrorGuidFirstMcpTools(t *testing.T) {
	domain := findRegisteredDomain(t, "shift-handover")
	tests := []struct {
		name       string
		toolName   string
		httpMethod string
		restPath   string
	}{
		{"history", "UteamupShiftHandoverGetHistory", "GET", "by-guid/{handoverGuid}/history"},
		{"signature-create", "UteamupShiftHandoverCreateSignature", "POST", "by-guid/{handoverGuid}/signatures"},
		{"audit-export", "UteamupShiftHandoverExportAudit", "GET", "by-guid/{handoverGuid}/audit-export"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			action := findShiftHandoverAction(t, domain, test.name)
			if action.ToolName != test.toolName {
				t.Errorf("ToolName = %q, want %q", action.ToolName, test.toolName)
			}
			if action.HTTPMethod != test.httpMethod || action.RESTPath != test.restPath {
				t.Errorf(
					"route = %s %s, want %s %s",
					action.HTTPMethod,
					action.RESTPath,
					test.httpMethod,
					test.restPath,
				)
			}
			if len(action.Args) != 1 ||
				action.Args[0].Name != "handoverGuid" ||
				action.Args[0].Type != "uuid" ||
				!action.Args[0].Required {
				t.Errorf("args = %+v, want one required handoverGuid UUID", action.Args)
			}
			if strings.Contains(action.RESTPath, "{id}") ||
				strings.Contains(action.RESTPath, "{handoverId}") {
				t.Errorf("RESTPath exposes an integer identity: %q", action.RESTPath)
			}
		})
	}
}

func TestShiftHandoverAuditRoutesResolveAndConsumeGuid(t *testing.T) {
	domain := findRegisteredDomain(t, "shift-handover")
	handoverGUID := "11111111-2222-4333-8444-555555555555"

	for _, name := range []string{"history", "signature-create", "audit-export"} {
		t.Run(name, func(t *testing.T) {
			action := findShiftHandoverAction(t, domain, name)
			path, consumed := buildRESTPath(
				domain,
				*action,
				map[string]any{"handoverGuid": handoverGUID},
			)
			wantPath := "/api/shifthandover/" + strings.ReplaceAll(
				action.RESTPath,
				"{handoverGuid}",
				handoverGUID,
			)
			if path != wantPath {
				t.Errorf("path = %q, want %q", path, wantPath)
			}
			if !reflect.DeepEqual(consumed, []string{"handoverGuid"}) {
				t.Errorf("consumed = %v, want [handoverGuid]", consumed)
			}
		})
	}
}

func TestShiftHandoverSignatureFlagsMatchBackendCreateModel(t *testing.T) {
	domain := findRegisteredDomain(t, "shift-handover")
	action := findShiftHandoverAction(t, domain, "signature-create")
	want := []struct {
		name     string
		bodyName string
		typeName string
		required bool
	}{
		{"purpose", "purpose", "string", true},
		{"method", "method", "string", true},
		{"meaning", "meaning", "string", true},
		{"device-key-identifier", "deviceKeyIdentifier", "string", false},
		{"idempotency-key", "idempotencyKey", "uuid", true},
	}

	if len(action.Flags) != len(want) {
		t.Fatalf("flags = %+v, want %+v", action.Flags, want)
	}
	for index, expected := range want {
		flag := action.Flags[index]
		bodyName := flag.BodyName
		if bodyName == "" {
			bodyName = toCamelCase(flag.Name)
		}
		if flag.Name != expected.name ||
			bodyName != expected.bodyName ||
			flag.Type != expected.typeName ||
			flag.Required != expected.required {
			t.Errorf("flag[%d] = %+v (body %q), want %+v", index, flag, bodyName, expected)
		}
		if flag.HeaderName != "" {
			t.Errorf("flag %q must be sent in the JSON body", flag.Name)
		}
	}
}
