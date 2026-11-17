package registry

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/uteamup/cli/internal/auth"
	"github.com/uteamup/cli/internal/client"
	"github.com/uteamup/cli/internal/logging"
)

const partRequestTestGUID = "11111111-1111-4111-8111-111111111111"

type capturedPartRequest struct {
	method string
	path   string
	query  url.Values
	body   map[string]any
}

func findPartRequestAction(t *testing.T, name string) (*Domain, Action) {
	t.Helper()
	domain := findDomain("part-request")
	if domain == nil {
		t.Fatal("part-request domain is not registered")
	}
	for _, action := range domain.Actions {
		if action.Name == name {
			return domain, action
		}
	}
	t.Fatalf("part-request %s is not registered", name)
	return nil, Action{}
}

// runPartRequestAction executes one action against an isolated TLS server and
// returns the captured request, or nil when the CLI refused before transport.
func runPartRequestAction(t *testing.T, name string, args ...string) (*capturedPartRequest, error) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	var captured *capturedPartRequest
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen := &capturedPartRequest{method: r.Method, path: r.URL.Path, query: r.URL.Query()}
		raw, _ := io.ReadAll(r.Body)
		if len(bytes.TrimSpace(raw)) > 0 {
			decoder := json.NewDecoder(bytes.NewReader(raw))
			decoder.UseNumber()
			if err := decoder.Decode(&seen.body); err != nil {
				t.Errorf("request body is not a JSON object: %v", err)
			}
		}
		captured = seen
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"guid":"` + partRequestTestGUID + `"}`))
	}))
	defer server.Close()
	if err := auth.SaveToken(&auth.TokenData{APIOrigin: server.URL, AccessToken: "isolated-test-token", ExpiresAt: time.Now().Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}
	logger := logging.New(logging.LevelError)
	format := "json"
	factory := func() (*client.APIClient, error) {
		return client.NewAPIClient(server.URL, time.Second, true, client.RetryOptions{MaxRetries: 0}, logger), nil
	}
	domain, action := findPartRequestAction(t, name)
	command := buildActionCommand(domain, action, factory, logger, &format, nil)
	command.SetOut(io.Discard)
	command.SetErr(io.Discard)
	command.SetArgs(args)
	err := command.Execute()
	return captured, err
}

func TestPartRequestDomainMirrorsMCPToolsOverRest(t *testing.T) {
	domain := findDomain("part-request")
	if domain == nil || domain.APIPath != "/api/part-requests" {
		t.Fatalf("part-request must route to /api/part-requests: %+v", domain)
	}
	want := map[string]struct{ tool, method string }{
		"list":   {"uteamup_part_request_list", "GET"},
		"get":    {"uteamup_part_request_get", "GET"},
		"create": {"uteamup_part_request_create", "POST"},
		"update": {"uteamup_part_request_update", "PUT"},
	}
	if len(domain.Actions) != len(want) {
		t.Fatalf("part-request must expose exactly list/get/create/update, got %d actions", len(domain.Actions))
	}
	for _, action := range domain.Actions {
		expected, ok := want[action.Name]
		if !ok {
			t.Fatalf("unexpected part-request action %q (approve/reject/bundle stay in the app)", action.Name)
		}
		if got := mcpToolName(action.ToolName); got != expected.tool {
			t.Errorf("%s maps to MCP tool %q, want %q", action.Name, got, expected.tool)
		}
		if action.MCPOnly {
			t.Errorf("%s has a REST endpoint and must not be MCPOnly (human sign-in sessions cannot call MCP)", action.Name)
		}
		if HTTPMethod[action.Name] != expected.method || action.HTTPMethod != "" {
			t.Errorf("%s must use the standard %s verb", action.Name, expected.method)
		}
		if !action.RejectExtraArgs {
			t.Errorf("%s must reject undeclared positional input", action.Name)
		}
		if err := validateActionDefinition(action); err != nil {
			t.Errorf("%s: %v", action.Name, err)
		}
	}
}

// The read-only uteamup-agent wrapper allows only read verbs; list/get must be
// the only actions under those names and the write verbs must not be GETs.
func TestPartRequestReadAndWriteVerbsAreClassifiedForReadOnlyWrapper(t *testing.T) {
	readVerbs := map[string]bool{"list": true, "get": true, "search": true, "show": true, "view": true}
	domain := findDomain("part-request")
	for _, action := range domain.Actions {
		isRead := HTTPMethod[action.Name] == "GET"
		if isRead != readVerbs[action.Name] {
			t.Errorf("%s: GET=%v but read-only wrapper classification=%v", action.Name, isRead, readVerbs[action.Name])
		}
	}
}

func TestPartRequestListSendsFiltersAndOffsetPaging(t *testing.T) {
	captured, err := runPartRequestAction(t, "list",
		"--status", "pending", "--urgency", "high",
		"--asset-guid", partRequestTestGUID, "--workorder-guid", partRequestTestGUID,
		"--page", "3", "--page-size", "20")
	if err != nil {
		t.Fatal(err)
	}
	if captured == nil || captured.method != http.MethodGet || captured.path != "/api/part-requests" {
		t.Fatalf("wrong list request: %+v", captured)
	}
	want := map[string]string{
		"status": "pending", "urgency": "high",
		"assetGuid": partRequestTestGUID, "workorderGuid": partRequestTestGUID,
		"skip": "40", "take": "20",
	}
	for key, value := range want {
		if got := captured.query.Get(key); got != value {
			t.Errorf("query %s = %q, want %q", key, got, value)
		}
	}
	for _, key := range []string{"page", "pageSize"} {
		if captured.query.Has(key) {
			t.Errorf("query %s must be translated to skip/take, not sent", key)
		}
	}
}

func TestPartRequestListDefaultsToFirstPageOfFifty(t *testing.T) {
	captured, err := runPartRequestAction(t, "list")
	if err != nil {
		t.Fatal(err)
	}
	if captured == nil || captured.query.Get("skip") != "0" || captured.query.Get("take") != "50" {
		t.Fatalf("default paging must be skip=0 take=50: %+v", captured)
	}
	for _, key := range []string{"status", "urgency", "assetGuid", "workorderGuid"} {
		if captured.query.Has(key) {
			t.Errorf("unset filter %s must not be sent", key)
		}
	}
}

func TestPartRequestRejectsInvalidInputBeforeTransport(t *testing.T) {
	cases := []struct {
		action string
		args   []string
		want   string
	}{
		{"list", []string{"--page", "0"}, "--page"},
		{"list", []string{"--page-size", "0"}, "--page-size"},
		{"list", []string{"--page-size", "101"}, "--page-size"},
		{"list", []string{"--page", "2147483647", "--page-size", "100"}, "out of range"},
		{"list", []string{"--status", "Pending"}, "must be one of"},
		{"list", []string{"--urgency", "urgent"}, "must be one of"},
		{"list", []string{"--asset-guid", "not-a-guid"}, "GUID"},
		{"get", []string{"00000000-0000-0000-0000-000000000000"}, "non-empty GUID"},
		{"get", []string{"../" + partRequestTestGUID}, "requestGuid"},
		{"create", []string{"--part-name", " "}, "cannot be empty"},
		{"create", []string{"--part-name", "SSD", "--quantity", "1e3"}, "quantity"},
		{"create", []string{"--part-name", "SSD", "--part-guid", "x"}, "GUID"},
		{"update", []string{partRequestTestGUID, "--part-name", "SSD"}, "quantity"},
		{"update", []string{partRequestTestGUID, "extra", "--part-name", "SSD", "--quantity", "1"}, "positional"},
	}
	for _, c := range cases {
		t.Run(c.action+" "+strings.Join(c.args, " "), func(t *testing.T) {
			captured, err := runPartRequestAction(t, c.action, c.args...)
			if captured != nil {
				t.Fatalf("invalid input reached transport: %+v", captured)
			}
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("error %v must mention %q", err, c.want)
			}
		})
	}
}

func TestPartRequestGetRoutesByPublicGuid(t *testing.T) {
	captured, err := runPartRequestAction(t, "get", partRequestTestGUID)
	if err != nil {
		t.Fatal(err)
	}
	if captured == nil || captured.method != http.MethodGet || captured.path != "/api/part-requests/"+partRequestTestGUID {
		t.Fatalf("wrong get request: %+v", captured)
	}
	if len(captured.query) != 0 {
		t.Fatalf("the route GUID must not leak into the query: %v", captured.query)
	}
}

func TestPartRequestCreateSendsCreateModelBody(t *testing.T) {
	captured, err := runPartRequestAction(t, "create",
		"--part-name", "Netac NV7000-t 1TB NVMe SSD (slot 2)",
		"--quantity", "2.50",
		"--part-guid", partRequestTestGUID,
		"--part-number", "NV7000-T-1TB",
		"--asset-guid", partRequestTestGUID,
		"--workorder-guid", partRequestTestGUID,
		"--preferred-vendor-guid", partRequestTestGUID,
		"--delivery-location-guid", partRequestTestGUID,
		"--uom", "pcs",
		"--estimated-unit-cost", "149.90",
		"--currency", "EUR",
		"--urgency", "critical",
		"--required-by-date", "2026-12-01T00:00:00Z",
		"--notes", "Planned for slot 2")
	if err != nil {
		t.Fatal(err)
	}
	if captured == nil || captured.method != http.MethodPost || captured.path != "/api/part-requests" {
		t.Fatalf("wrong create request: %+v", captured)
	}
	want := map[string]any{
		"partName":             "Netac NV7000-t 1TB NVMe SSD (slot 2)",
		"quantity":             json.Number("2.50"),
		"partGuid":             partRequestTestGUID,
		"partNumber":           "NV7000-T-1TB",
		"assetGuid":            partRequestTestGUID,
		"workorderGuid":        partRequestTestGUID,
		"preferredVendorGuid":  partRequestTestGUID,
		"deliveryLocationGuid": partRequestTestGUID,
		"uom":                  "pcs",
		"estimatedUnitCost":    json.Number("149.90"),
		"currency":             "EUR",
		"urgency":              "critical",
		"requiredByDate":       "2026-12-01T00:00:00Z",
		"reviewNotes":          "Planned for slot 2",
	}
	if len(captured.body) != len(want) {
		t.Errorf("create body has %d fields, want %d: %v", len(captured.body), len(want), captured.body)
	}
	for key, value := range want {
		if captured.body[key] != value {
			t.Errorf("body %s = %#v, want %#v", key, captured.body[key], value)
		}
	}
	for _, forbidden := range []string{"notes", "status", "tenantGuid", "requestedByUserId"} {
		if _, ok := captured.body[forbidden]; ok {
			t.Errorf("body must not carry %s", forbidden)
		}
	}
}

func TestPartRequestCreateMinimalDefaultsToMediumUrgency(t *testing.T) {
	captured, err := runPartRequestAction(t, "create", "--part-name", "Fan belt")
	if err != nil {
		t.Fatal(err)
	}
	if captured == nil || captured.body["partName"] != "Fan belt" || captured.body["urgency"] != "medium" {
		t.Fatalf("minimal create must send part name and medium urgency: %+v", captured)
	}
	if _, ok := captured.body["quantity"]; ok {
		t.Error("omitted quantity must use the backend default of 1, not a CLI value")
	}
}

func TestPartRequestUpdatePutsReplacementByPublicGuid(t *testing.T) {
	captured, err := runPartRequestAction(t, "update", partRequestTestGUID,
		"--part-name", "Fan belt", "--quantity", "3", "--urgency", "low", "--notes", "Swap to OEM")
	if err != nil {
		t.Fatal(err)
	}
	if captured == nil || captured.method != http.MethodPut || captured.path != "/api/part-requests/"+partRequestTestGUID {
		t.Fatalf("wrong update request: %+v", captured)
	}
	if _, ok := captured.body["requestGuid"]; ok {
		t.Error("route GUID must not be duplicated in the body")
	}
	if captured.body["partName"] != "Fan belt" || captured.body["quantity"] != json.Number("3") ||
		captured.body["urgency"] != "low" || captured.body["reviewNotes"] != "Swap to OEM" {
		t.Fatalf("wrong update body: %v", captured.body)
	}
}

func TestOffsetPagingRequiresPlainIntPageFlags(t *testing.T) {
	base := Action{Name: "list", OffsetPagingMaxPageSize: 10}
	if err := validateActionDefinition(base); err == nil {
		t.Fatal("offset paging without page flags must be rejected")
	}
	base.Flags = []FlagDef{
		{Name: "page", Type: "int", Default: 1},
		{Name: "page-size", Type: "int", Default: 10, QueryName: "pageSize"},
	}
	if err := validateActionDefinition(base); err == nil {
		t.Fatal("offset paging must reject a page-size flag that already has its own query name")
	}
	base.Flags[1].QueryName = ""
	if err := validateActionDefinition(base); err != nil {
		t.Fatalf("plain int page flags must be accepted: %v", err)
	}
}

func TestPartRequestInvalidPagingFailsBeforeClientCreation(t *testing.T) {
	domain, action := findPartRequestAction(t, "list")
	reached := errors.New("client factory reached")
	format := "json"
	command := buildActionCommand(domain, action, func() (*client.APIClient, error) {
		return nil, reached
	}, logging.New(logging.LevelError), &format, nil)
	command.SetOut(io.Discard)
	command.SetErr(io.Discard)
	command.SetArgs([]string{"--page-size", "500"})
	if err := command.Execute(); err == nil || errors.Is(err, reached) {
		t.Fatalf("invalid page size must fail before a client is created: %v", err)
	}
}
