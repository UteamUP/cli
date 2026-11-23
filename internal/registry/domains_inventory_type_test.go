package registry

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/uteamup/cli/internal/auth"
	"github.com/uteamup/cli/internal/client"
	"github.com/uteamup/cli/internal/logging"
)

// The inventory type domain mirrors the routes the UTE-1075 API actually serves:
//
//	GET   /api/inventory-type/{kind}
//	GET   /api/inventory-type/{kind}/{typeGuid}
//	POST  /api/inventory-type/{kind}
//	PUT   /api/inventory-type/{kind}/{typeGuid}
//	POST  /api/inventory-type/{kind}/{typeGuid}/fields
//	PUT   /api/inventory-type/{kind}/{typeGuid}/fields/{fieldGuid}
//	GET   /api/inventory-type/{kind}/items/{itemGuid}
//	PATCH /api/inventory-type/{kind}/items/{itemGuid}
func TestInventoryTypeActionsMirrorTheExactApiRoutes(t *testing.T) {
	domain := findDomainByName(t, "inventory-type")
	if domain.APIPath != "/api/inventory-type" {
		t.Fatalf("inventory-type base path = %q", domain.APIPath)
	}

	expected := map[string]struct {
		method, path, tool string
	}{
		"list":         {"GET", "{kind}", "UteamupInventoryTypeList"},
		"get":          {"GET", "{kind}/{typeGuid}", "UteamupInventoryTypeGet"},
		"create":       {"POST", "{kind}", "UteamupInventoryTypeCreate"},
		"update":       {"PUT", "{kind}/{typeGuid}", "UteamupInventoryTypeUpdate"},
		"field-create": {"POST", "{kind}/{typeGuid}/fields", "UteamupInventoryTypeFieldCreate"},
		"field-update": {"PUT", "{kind}/{typeGuid}/fields/{fieldGuid}", "UteamupInventoryTypeFieldUpdate"},
		"item-get":     {"GET", "{kind}/items/{itemGuid}", "UteamupInventoryTypeItemGet"},
		"item-save":    {"PATCH", "{kind}/items/{itemGuid}", "UteamupInventoryTypeItemSave"},
	}

	if len(domain.Actions) != len(expected) {
		t.Fatalf("inventory-type has %d actions, want %d", len(domain.Actions), len(expected))
	}

	for name, want := range expected {
		action := findDomainAction(t, "inventory-type", name)
		if action.HTTPMethod != want.method || action.RESTPath != want.path || action.ToolName != want.tool {
			t.Fatalf("%s does not mirror the exact API route/tool: %+v", name, action)
		}
		if action.MCPOnly {
			t.Fatalf("%s must use the REST route the API serves", name)
		}
	}
}

// The registry must not invent a route the backend does not serve. An extra action
// would post to a route that answers 404, and a fabricated tool name would point at
// a tool that does not exist. The generic fall-through test only checks that a route
// is declared, not that it exists upstream, so the domain checks this itself.
func TestInventoryTypeDeclaresNoRouteTheBackendDoesNotServe(t *testing.T) {
	domain := findDomainByName(t, "inventory-type")

	served := map[string]bool{
		"GET {kind}":                               true,
		"GET {kind}/{typeGuid}":                    true,
		"POST {kind}":                              true,
		"PUT {kind}/{typeGuid}":                    true,
		"POST {kind}/{typeGuid}/fields":            true,
		"PUT {kind}/{typeGuid}/fields/{fieldGuid}": true,
		"GET {kind}/items/{itemGuid}":              true,
		"PATCH {kind}/items/{itemGuid}":            true,
	}

	for _, action := range domain.Actions {
		route := action.HTTPMethod + " " + action.RESTPath
		if !served[route] {
			t.Errorf("%s declares the unserved route %q", action.Name, route)
		}
	}
}

// Every action rejects undeclared positional input, so a stray argument can never
// reach the backend as part of the path.
func TestInventoryTypeActionsRejectExtraPositionalInput(t *testing.T) {
	domain := findDomainByName(t, "inventory-type")

	for _, action := range domain.Actions {
		if !action.RejectExtraArgs {
			t.Errorf("%s must reject extra positional input", action.Name)
		}
	}
}

// The kind is a route segment, so it is constrained to the three inventory kinds the
// API serves. A free-form value could otherwise reshape the path.
func TestInventoryTypeKindIsConstrainedToTheThreeInventoryKinds(t *testing.T) {
	domain := findDomainByName(t, "inventory-type")
	for _, action := range domain.Actions {
		if len(action.Args) == 0 {
			t.Fatalf("%s has no kind argument", action.Name)
		}
		kind := action.Args[0]
		if kind.Name != "kind" || !kind.Required || kind.Type != "string" {
			t.Fatalf("%s kind is miswired: %+v", action.Name, kind)
		}
		if len(kind.AllowedValues) != 3 {
			t.Fatalf("%s kind must allow exactly the three inventory kinds: %+v", action.Name, kind.AllowedValues)
		}
		for _, value := range kind.AllowedValues {
			switch value {
			case "part", "tool", "chemical":
			default:
				t.Fatalf("%s kind allows an unknown kind %q", action.Name, value)
			}
		}
	}
}

// Inventory types are GUID-first and never expose the internal integer identity.
func TestInventoryTypeActionsAreGuidOnly(t *testing.T) {
	domain := findDomainByName(t, "inventory-type")
	for _, action := range domain.Actions {
		for _, arg := range action.Args {
			if arg.Name == "id" || arg.Type == "int" {
				t.Errorf("%s exposes a forbidden integer identity arg %+v", action.Name, arg)
			}
		}
		for _, flag := range action.Flags {
			if flag.Name == "id" || flag.Type == "int" && flag.Name == "type-id" {
				t.Errorf("%s exposes a forbidden integer identity flag %+v", action.Name, flag)
			}
		}
	}

	for _, name := range []string{"get", "update"} {
		action := findDomainAction(t, "inventory-type", name)
		if len(action.Args) != 2 || action.Args[1].Name != "typeGuid" ||
			!action.Args[1].Required || action.Args[1].Type != "non-empty-uuid" {
			t.Errorf("%s typeGuid is miswired: %+v", name, action.Args)
		}
	}

	for _, name := range []string{"field-create"} {
		action := findDomainAction(t, "inventory-type", name)
		if len(action.Args) != 2 || action.Args[1].Name != "typeGuid" ||
			!action.Args[1].Required || action.Args[1].Type != "non-empty-uuid" {
			t.Errorf("%s typeGuid is miswired: %+v", name, action.Args)
		}
	}

	fieldUpdate := findDomainAction(t, "inventory-type", "field-update")
	if len(fieldUpdate.Args) != 3 || fieldUpdate.Args[2].Name != "fieldGuid" ||
		!fieldUpdate.Args[2].Required || fieldUpdate.Args[2].Type != "non-empty-uuid" {
		t.Errorf("field-update fieldGuid is miswired: %+v", fieldUpdate.Args)
	}

	for _, name := range []string{"item-get", "item-save"} {
		action := findDomainAction(t, "inventory-type", name)
		if len(action.Args) != 2 || action.Args[1].Name != "itemGuid" ||
			!action.Args[1].Required || action.Args[1].Type != "non-empty-uuid" {
			t.Errorf("%s itemGuid is miswired: %+v", name, action.Args)
		}
	}
}

// Types are archived, never deleted: the API has no hard delete for them.
func TestInventoryTypeOffersNoDelete(t *testing.T) {
	domain := findDomainByName(t, "inventory-type")
	for _, action := range domain.Actions {
		if action.Name == "delete" || action.HTTPMethod == "DELETE" {
			t.Errorf("inventory-type must not offer a delete action: %+v", action.Name)
		}
	}
}

// The field key and data type are the field's immutable identity, so both are
// constrained to the contract's exact shape.
func TestInventoryTypeFieldActionsConstrainTheFieldIdentity(t *testing.T) {
	for _, name := range []string{"field-create", "field-update"} {
		action := findDomainAction(t, "inventory-type", name)
		key := actionFlagByName(t, action, "field-key")
		if !key.Required || key.Type != "string" {
			t.Errorf("%s field-key is miswired: %+v", name, key)
		}
		dataType := actionFlagByName(t, action, "data-type")
		if !dataType.Required || len(dataType.AllowedValues) != 6 {
			t.Errorf("%s data-type is miswired: %+v", name, dataType)
		}
		for _, value := range dataType.AllowedValues {
			switch value {
			case "string", "integer", "float", "boolean", "date", "array":
			default:
				t.Errorf("%s data-type allows an unknown type %q", name, value)
			}
		}
	}
}

// The selection patch is merged into the request-body root, because the API binds
// InventoryTypeSelectionModel as the body itself and not under a wrapper field.
func TestInventoryTypeItemSaveSendsTheSelectionAsTheRequestBodyRoot(t *testing.T) {
	action := findDomainAction(t, "inventory-type", "item-save")
	file := actionFlagByName(t, action, "file")
	if !file.Required || !file.RootJSONObjectFile {
		t.Fatalf("item-save file is miswired: %+v", file)
	}
	if file.BodyName != "" || file.JSONFile {
		t.Fatalf("item-save file must not wrap the selection: %+v", file)
	}
}

// The list route reads its paging and inactive filter from the query string.
func TestInventoryTypeListReadsItsFiltersFromTheQueryString(t *testing.T) {
	action := findDomainAction(t, "inventory-type", "list")
	inactive := actionFlagByName(t, action, "include-inactive")
	if inactive.Type != "bool" || inactive.Default != false {
		t.Fatalf("list include-inactive is miswired: %+v", inactive)
	}
	page := actionFlagByName(t, action, "page")
	size := actionFlagByName(t, action, "page-size")
	if page.Type != "int" || size.Type != "int" {
		t.Fatalf("list paging is miswired: %+v %+v", page, size)
	}
}

// Every route the CLI builds is exercised against a real server so a wrong path
// cannot hide behind a passing registry declaration.
func TestInventoryTypeTransportBuildsTheRealRoutes(t *testing.T) {
	const typeGuid = "11111111-1111-4111-8111-111111111111"
	const fieldGuid = "22222222-2222-4222-8222-222222222222"
	const itemGuid = "33333333-3333-4333-8333-333333333333"

	for _, scenario := range []struct {
		name       string
		args       []string
		wantMethod string
		wantPath   string
		wantQuery  map[string]string
		wantBody   map[string]any
	}{
		{
			name: "list", args: []string{"list", "part"},
			wantMethod: "GET", wantPath: "/api/inventory-type/part",
			wantQuery: map[string]string{"includeInactive": "false", "page": "1", "pageSize": "50"},
		},
		{
			name: "get", args: []string{"get", "tool", typeGuid},
			wantMethod: "GET", wantPath: "/api/inventory-type/tool/" + typeGuid,
		},
		{
			name: "create", args: []string{"create", "chemical", "--name", "Coolant"},
			wantMethod: "POST", wantPath: "/api/inventory-type/chemical",
			wantBody: map[string]any{"name": "Coolant", "isActive": true, "displayOrder": float64(0)},
		},
		{
			name: "update", args: []string{"update", "part", typeGuid, "--name", "Hydraulic", "--is-active=false"},
			wantMethod: "PUT", wantPath: "/api/inventory-type/part/" + typeGuid,
			wantBody: map[string]any{"name": "Hydraulic", "isActive": false, "displayOrder": float64(0)},
		},
		{
			name: "field-create", args: []string{"field-create", "part", typeGuid,
				"--field-key", "operating_pressure", "--name", "Operating pressure", "--data-type", "float"},
			wantMethod: "POST", wantPath: "/api/inventory-type/part/" + typeGuid + "/fields",
			wantBody: map[string]any{"fieldKey": "operating_pressure", "name": "Operating pressure",
				"dataType": "float", "isRequired": false, "isActive": true, "displayOrder": float64(0)},
		},
		{
			name: "field-update", args: []string{"field-update", "part", typeGuid, fieldGuid,
				"--field-key", "operating_pressure", "--name", "Operating pressure", "--data-type", "float"},
			wantMethod: "PUT", wantPath: "/api/inventory-type/part/" + typeGuid + "/fields/" + fieldGuid,
		},
		{
			name: "item-get", args: []string{"item-get", "tool", itemGuid},
			wantMethod: "GET", wantPath: "/api/inventory-type/tool/items/" + itemGuid,
		},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv("USERPROFILE", home)

			var received *http.Request
			var receivedBody []byte
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				received = r.Clone(r.Context())
				receivedBody, _ = io.ReadAll(r.Body)
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{}`))
			}))
			t.Cleanup(server.Close)

			if err := auth.SaveToken(&auth.TokenData{
				APIOrigin:   server.URL,
				AccessToken: "inventory-type-token",
				ExpiresAt:   time.Now().Add(time.Hour),
				TenantGUID:  "55555555-5555-4555-8555-555555555555",
			}); err != nil {
				t.Fatalf("save test token: %v", err)
			}

			apiClient := client.NewAPIClient(server.URL, time.Second, true,
				client.RetryOptions{MaxRetries: 0}, logging.New(logging.LevelError))
			format := "json"
			command := buildDomainCommand(
				findDomainByName(t, "inventory-type"),
				func() (*client.APIClient, error) { return apiClient, nil },
				logging.New(logging.LevelError),
				&format,
				&ExportConfig{},
			)
			command.SetOut(os.Stdout)
			command.SetErr(os.Stderr)
			command.SilenceErrors = true
			command.SilenceUsage = true
			command.SetArgs(scenario.args)

			if err := command.Execute(); err != nil {
				t.Fatalf("%s error = %v", scenario.name, err)
			}
			if received == nil {
				t.Fatalf("%s sent no request", scenario.name)
			}
			if received.Method != scenario.wantMethod {
				t.Errorf("method = %s, want %s", received.Method, scenario.wantMethod)
			}
			if received.URL.Path != scenario.wantPath {
				t.Errorf("path = %s, want %s", received.URL.Path, scenario.wantPath)
			}
			for key, want := range scenario.wantQuery {
				if got := received.URL.Query().Get(key); got != want {
					t.Errorf("query %s = %q, want %q", key, got, want)
				}
			}
			if scenario.wantBody != nil {
				var body map[string]any
				if err := json.Unmarshal(receivedBody, &body); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				for key, want := range scenario.wantBody {
					if body[key] != want {
						t.Errorf("body %s = %#v, want %#v", key, body[key], want)
					}
				}
			}
		})
	}
}

// The sparse value patch reaches the API as the request body root, so an omitted
// typeGuids preserves the assignment and an explicit empty array clears it.
func TestInventoryTypeItemSaveSendsTheSparseValuePatch(t *testing.T) {
	const itemGuid = "33333333-3333-4333-8333-333333333333"
	const fieldGuid = "22222222-2222-4222-8222-222222222222"

	for _, scenario := range []struct {
		name       string
		selection  string
		wantAbsent bool
		wantEmpty  bool
	}{
		{name: "preserve", selection: `{"values":[{"fieldGuid":"` + fieldGuid + `","value":0}]}`, wantAbsent: true},
		{name: "clear", selection: `{"typeGuids":[]}`},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv("USERPROFILE", home)

			path := filepath.Join(t.TempDir(), "selection.json")
			if err := os.WriteFile(path, []byte(scenario.selection), 0o600); err != nil {
				t.Fatalf("write selection: %v", err)
			}

			var received *http.Request
			var receivedBody []byte
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				received = r.Clone(r.Context())
				receivedBody, _ = io.ReadAll(r.Body)
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{}`))
			}))
			t.Cleanup(server.Close)

			if err := auth.SaveToken(&auth.TokenData{
				APIOrigin:   server.URL,
				AccessToken: "inventory-type-token",
				ExpiresAt:   time.Now().Add(time.Hour),
				TenantGUID:  "55555555-5555-4555-8555-555555555555",
			}); err != nil {
				t.Fatalf("save test token: %v", err)
			}

			apiClient := client.NewAPIClient(server.URL, time.Second, true,
				client.RetryOptions{MaxRetries: 0}, logging.New(logging.LevelError))
			format := "json"
			command := buildDomainCommand(
				findDomainByName(t, "inventory-type"),
				func() (*client.APIClient, error) { return apiClient, nil },
				logging.New(logging.LevelError),
				&format,
				&ExportConfig{},
			)
			command.SetOut(os.Stdout)
			command.SetErr(os.Stderr)
			command.SilenceErrors = true
			command.SilenceUsage = true
			command.SetArgs([]string{"item-save", "part", itemGuid, "--file", path})

			if err := command.Execute(); err != nil {
				t.Fatalf("item-save error = %v", err)
			}
			if received == nil {
				t.Fatal("item-save sent no request")
			}
			if received.Method != "PATCH" {
				t.Errorf("method = %s, want PATCH", received.Method)
			}
			if received.URL.Path != "/api/inventory-type/part/items/"+itemGuid {
				t.Errorf("path = %s", received.URL.Path)
			}

			var body map[string]any
			if err := json.Unmarshal(receivedBody, &body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if _, wrapped := body["selection"]; wrapped {
				t.Errorf("the selection must be the body root, not a wrapper: %v", body)
			}
			guids, present := body["typeGuids"]
			if scenario.wantAbsent {
				if present {
					t.Errorf("an omitted typeGuids must not be sent: %v", body)
				}
				values, ok := body["values"].([]any)
				if !ok || len(values) != 1 {
					t.Fatalf("values = %v", body["values"])
				}
				entry, ok := values[0].(map[string]any)
				if !ok || entry["fieldGuid"] != fieldGuid {
					t.Errorf("value entry = %v", values[0])
				}
				if entry["value"] != float64(0) {
					t.Errorf("a zero value must survive as a real value: %v", entry["value"])
				}
				return
			}
			empty, ok := guids.([]any)
			if !ok || len(empty) != 0 {
				t.Errorf("an empty typeGuids must clear the assignment: %v", body["typeGuids"])
			}
		})
	}
}

// The kind is the only way to address an inventory kind, so an unknown kind is
// refused before any request leaves the CLI.
func TestInventoryTypeRefusesAnUnknownKind(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("an unknown kind must not reach the server: %s %s", r.Method, r.URL.Path)
	}))
	t.Cleanup(server.Close)

	if err := auth.SaveToken(&auth.TokenData{
		APIOrigin:   server.URL,
		AccessToken: "inventory-type-token",
		ExpiresAt:   time.Now().Add(time.Hour),
		TenantGUID:  "55555555-5555-4555-8555-555555555555",
	}); err != nil {
		t.Fatalf("save test token: %v", err)
	}

	apiClient := client.NewAPIClient(server.URL, time.Second, true,
		client.RetryOptions{MaxRetries: 0}, logging.New(logging.LevelError))
	format := "json"
	command := buildDomainCommand(
		findDomainByName(t, "inventory-type"),
		func() (*client.APIClient, error) { return apiClient, nil },
		logging.New(logging.LevelError),
		&format,
		&ExportConfig{},
	)
	command.SetOut(os.Stdout)
	command.SetErr(os.Stderr)
	command.SilenceErrors = true
	command.SilenceUsage = true
	command.SetArgs([]string{"list", "asset"})

	if err := command.Execute(); err == nil {
		t.Fatal("an unknown kind must be refused")
	}
}
