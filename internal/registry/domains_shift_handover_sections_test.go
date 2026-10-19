package registry

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/uteamup/cli/internal/auth"
	"github.com/uteamup/cli/internal/client"
	"github.com/uteamup/cli/internal/logging"
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
			if len(action.Flags) != len(test.flags)+1 {
				t.Fatalf("flags = %+v, want body %+v and required retry header", action.Flags, test.flags)
			}
			assertHandoverMutationFlags(t, test.name, action.Flags)
			key := action.Flags[len(action.Flags)-1]
			if key.Name != "idempotency-key" || key.HeaderName != "Idempotency-Key" || key.BodyName != "" {
				t.Fatalf("last flag must be an HTTP header only: %+v", key)
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

// Exercises the real command builder and HTTP client; these cases remain unrun until the goal gate.
func TestShiftHandoverSectionKeysReachNormalTransportWithoutLeakingIntoJson(t *testing.T) {
	const parent = "11111111-2222-4333-8444-555555555555"
	const child = "aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee"
	const key = "12345678-1234-4234-8234-123456789abc"
	for _, name := range []string{"section-create", "section-update", "section-delete", "sections-reorder"} {
		t.Run(name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv("USERPROFILE", home)
			seen := 0
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
				seen++
				expectedPath := "/api/shifthandover/by-guid/" + parent + "/sections"
				if name == "section-update" || name == "section-delete" {
					expectedPath += "/" + child
				}
				if name == "sections-reorder" {
					expectedPath += "/reorder"
				}
				expectedMethod := map[string]string{"section-create": "POST", "section-update": "PUT", "section-delete": "DELETE", "sections-reorder": "PUT"}[name]
				if request.URL.Path != expectedPath || request.Method != expectedMethod {
					t.Errorf("wrong request %s %s", request.Method, request.URL.Path)
				}
				if request.Header.Get("Idempotency-Key") != key {
					t.Error("original GUID key missing or changed")
				}
				var body map[string]any
				if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
					t.Error(err)
				}
				if body["concurrencyToken"] != "opaque-original-token" {
					t.Error("original concurrency token changed")
				}
				if _, exists := body["idempotencyKey"]; exists {
					t.Error("header leaked into JSON")
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"message":"confirmed"}`))
			}))
			defer server.Close()
			if err := auth.SaveToken(&auth.TokenData{APIOrigin: server.URL, AccessToken: "isolated-token", ExpiresAt: time.Now().Add(time.Hour)}); err != nil {
				t.Fatal(err)
			}
			logger := logging.New(logging.LevelError)
			format := "json"
			factory := func() (*client.APIClient, error) {
				return client.NewAPIClient(server.URL, time.Second, true, client.RetryOptions{MaxRetries: 0}, logger), nil
			}
			domain := findRegisteredDomain(t, "shift-handover")
			action := findShiftHandoverAction(t, domain, name)
			args := []string{parent}
			if name == "section-update" || name == "section-delete" {
				args = append(args, child)
			}
			args = append(args, "--concurrency-token", "opaque-original-token")
			if name == "section-create" {
				args = append(args, "--section-type", "5")
			}
			if name == "sections-reorder" {
				args = append(args, "--section-guids", child+","+parent)
			}
			missing := buildActionCommand(domain, *action, factory, logger, &format, nil)
			missing.SetArgs(args)
			if err := missing.Execute(); err == nil || seen != 0 {
				t.Fatal("missing required retry key must refuse before HTTP")
			}
			command := buildActionCommand(domain, *action, factory, logger, &format, nil)
			command.SetArgs(append(append([]string{}, args...), "--idempotency-key", key))
			if err := command.Execute(); err != nil {
				t.Fatal(err)
			}
			if seen != 1 {
				t.Fatalf("expected one explicit request, got %d", seen)
			}
		})
	}
}
