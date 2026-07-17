package registry

import (
	"reflect"
	"strings"
	"testing"
)

func TestShiftHandoverSectionActionsUseGuidFirstContracts(t *testing.T) {
	domain := findRegisteredDomain(t, "shift-handover")

	tests := []struct {
		name       string
		toolName   string
		httpMethod string
		restPath   string
		argNames   []string
	}{
		{"sections", "UteamupShiftHandoverGetSections", "GET", "by-guid/{handoverGuid}/sections", []string{"handoverGuid"}},
		{"section-create", "UteamupShiftHandoverCreateSection", "POST", "by-guid/{handoverGuid}/sections", []string{"handoverGuid"}},
		{"section-update", "UteamupShiftHandoverUpdateSection", "PUT", "by-guid/{handoverGuid}/sections/{sectionGuid}", []string{"handoverGuid", "sectionGuid"}},
		{"section-delete", "UteamupShiftHandoverDeleteSection", "DELETE", "by-guid/{handoverGuid}/sections/{sectionGuid}", []string{"handoverGuid", "sectionGuid"}},
		{"sections-reorder", "UteamupShiftHandoverReorderSections", "PUT", "by-guid/{handoverGuid}/sections/reorder", []string{"handoverGuid"}},
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
				strings.Contains(action.RESTPath, "{sectionId}") {
				t.Errorf("RESTPath exposes an integer identity placeholder: %q", action.RESTPath)
			}
		})
	}
}

func TestShiftHandoverSectionRoutesResolveAndConsumeGuidArguments(t *testing.T) {
	domain := findRegisteredDomain(t, "shift-handover")
	handoverGUID := "11111111-2222-3333-4444-555555555555"
	sectionGUID := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"

	tests := []struct {
		name         string
		args         map[string]any
		wantPath     string
		wantConsumed []string
	}{
		{"sections", map[string]any{"handoverGuid": handoverGUID}, "/api/shifthandover/by-guid/" + handoverGUID + "/sections", []string{"handoverGuid"}},
		{"section-create", map[string]any{"handoverGuid": handoverGUID}, "/api/shifthandover/by-guid/" + handoverGUID + "/sections", []string{"handoverGuid"}},
		{"section-update", map[string]any{"handoverGuid": handoverGUID, "sectionGuid": sectionGUID}, "/api/shifthandover/by-guid/" + handoverGUID + "/sections/" + sectionGUID, []string{"handoverGuid", "sectionGuid"}},
		{"section-delete", map[string]any{"handoverGuid": handoverGUID, "sectionGuid": sectionGUID}, "/api/shifthandover/by-guid/" + handoverGUID + "/sections/" + sectionGUID, []string{"handoverGuid", "sectionGuid"}},
		{"sections-reorder", map[string]any{"handoverGuid": handoverGUID}, "/api/shifthandover/by-guid/" + handoverGUID + "/sections/reorder", []string{"handoverGuid"}},
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

func TestShiftHandoverSectionMutationFlagsMatchBackendModels(t *testing.T) {
	domain := findRegisteredDomain(t, "shift-handover")

	tests := []struct {
		name  string
		flags []shiftHandoverSectionFlagContract
	}{
		{
			name: "section-create",
			flags: []shiftHandoverSectionFlagContract{
				{"section-type", "sectionType", "int", true},
				{"title", "title", "string", false},
				{"content", "content", "string", false},
				{"sort-order", "sortOrder", "int", false},
				{"required", "isRequired", "bool", false},
				{"concurrency-token", "concurrencyToken", "string", true},
			},
		},
		{
			name: "section-update",
			flags: []shiftHandoverSectionFlagContract{
				{"title", "title", "string", false},
				{"content", "content", "string", false},
				{"completed", "isCompleted", "bool", false},
				{"sort-order", "sortOrder", "int", false},
				{"concurrency-token", "concurrencyToken", "string", true},
			},
		},
		{
			name: "section-delete",
			flags: []shiftHandoverSectionFlagContract{
				{"concurrency-token", "concurrencyToken", "string", true},
			},
		},
		{
			name: "sections-reorder",
			flags: []shiftHandoverSectionFlagContract{
				{"section-guids", "sectionGuids", "stringSlice", true},
				{"concurrency-token", "concurrencyToken", "string", true},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			action := findShiftHandoverAction(t, domain, test.name)
			if len(action.Flags) != len(test.flags) {
				t.Fatalf("flags = %+v, want %+v", action.Flags, test.flags)
			}
			for index, want := range test.flags {
				flag := action.Flags[index]
				bodyName := flag.BodyName
				if bodyName == "" {
					bodyName = toCamelCase(flag.Name)
				}
				if flag.Name != want.name || bodyName != want.bodyName || flag.Type != want.flagType || flag.Required != want.required {
					t.Errorf("flag[%d] = %+v (body %q), want %+v", index, flag, bodyName, want)
				}
				if flag.HeaderName != "" {
					t.Errorf("flag %q must be sent in the JSON body, not header %q", flag.Name, flag.HeaderName)
				}
			}
		})
	}
}

type shiftHandoverSectionFlagContract struct {
	name     string
	bodyName string
	flagType string
	required bool
}

func findShiftHandoverAction(t *testing.T, domain *Domain, name string) *Action {
	t.Helper()
	for index := range domain.Actions {
		if domain.Actions[index].Name == name {
			return &domain.Actions[index]
		}
	}
	t.Fatalf("shift-handover action %q missing", name)
	return nil
}
