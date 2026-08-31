package registry

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/uteamup/cli/internal/auth"
	"github.com/uteamup/cli/internal/client"
	"github.com/uteamup/cli/internal/logging"
)

const qmsAuditValidGUID = "11111111-2222-4333-8444-555555555555"

type qmsAuditRouteExpectation struct {
	method string
	path   string
	tool   string
}

type qmsAuditMutationExpectation struct {
	concurrency  bool
	confirmation bool
}

func qmsAuditDomain(t *testing.T, name, apiPath string, actionCount int) *Domain {
	t.Helper()
	for _, domain := range DefaultRegistry.Domains() {
		if domain.Name != name {
			continue
		}
		if domain.APIPath != apiPath {
			t.Fatalf("%s APIPath = %q, want %q", name, domain.APIPath, apiPath)
		}
		if len(domain.Actions) != actionCount {
			t.Fatalf("%s action count = %d, want %d", name, len(domain.Actions), actionCount)
		}
		return domain
	}
	t.Fatalf("%s domain is not registered", name)
	return nil
}

func qmsAuditAction(t *testing.T, domain *Domain, name string) Action {
	t.Helper()
	for _, action := range domain.Actions {
		if action.Name == name {
			return action
		}
	}
	t.Fatalf("%s action %q is not registered", domain.Name, name)
	return Action{}
}

func qmsAuditFlag(t *testing.T, action Action, name string) FlagDef {
	t.Helper()
	for _, flag := range action.Flags {
		if flag.Name == name {
			return flag
		}
	}
	t.Fatalf("%s flag --%s is not registered", action.Name, name)
	return FlagDef{}
}

func qmsAuditAssertRoutes(
	t *testing.T,
	domain *Domain,
	expected map[string]qmsAuditRouteExpectation,
) {
	t.Helper()
	if len(expected) != len(domain.Actions) {
		t.Fatalf("route expectations = %d, domain actions = %d", len(expected), len(domain.Actions))
	}
	for actionName, want := range expected {
		action := qmsAuditAction(t, domain, actionName)
		if action.HTTPMethod != want.method || action.ToolName != want.tool {
			t.Errorf(
				"%s contract = method %q tool %q, want method %q tool %q",
				actionName,
				action.HTTPMethod,
				action.ToolName,
				want.method,
				want.tool,
			)
		}
		arguments := make(map[string]any, len(action.Args))
		for _, argument := range action.Args {
			if len(argument.AllowedValues) > 0 {
				arguments[argument.Name] = argument.AllowedValues[0]
			} else {
				arguments[argument.Name] = qmsAuditValidGUID
			}
			if argument.Type != "string" && argument.Type != "non-empty-uuid" {
				t.Errorf("%s argument %s type = %q", actionName, argument.Name, argument.Type)
			}
		}
		path, consumed := buildRESTPath(domain, action, arguments)
		if path != want.path {
			t.Errorf("%s path = %q, want %q", actionName, path, want.path)
		}
		if len(consumed) != len(action.Args) {
			t.Errorf("%s consumed args = %v, want all %d path args", actionName, consumed, len(action.Args))
		}
	}
}

func qmsAuditAssertSearchFlags(t *testing.T, action Action, expected map[string]FlagDef) {
	t.Helper()
	if len(action.Flags) != len(expected) {
		t.Fatalf("%s search flags = %d, want %d", action.Name, len(action.Flags), len(expected))
	}
	seen := make(map[string]bool, len(action.Flags))
	for _, flag := range action.Flags {
		want, ok := expected[flag.Name]
		if !ok {
			t.Errorf("unexpected %s search flag --%s", action.Name, flag.Name)
			continue
		}
		if seen[flag.Name] {
			t.Errorf("duplicate %s search flag --%s", action.Name, flag.Name)
		}
		seen[flag.Name] = true
		if flag.QueryName != want.QueryName || flag.Type != want.Type ||
			!reflect.DeepEqual(flag.Default, want.Default) {
			t.Errorf("%s --%s = %+v, want query=%q type=%q default=%v", action.Name, flag.Name, flag, want.QueryName, want.Type, want.Default)
		}
	}
	for name := range expected {
		if !seen[name] {
			t.Errorf("missing %s search flag --%s", action.Name, name)
		}
	}
}

func qmsAuditAssertMutationFlags(
	t *testing.T,
	domain *Domain,
	expected map[string]qmsAuditMutationExpectation,
) {
	t.Helper()
	for actionName, want := range expected {
		action := qmsAuditAction(t, domain, actionName)
		requestFile := qmsAuditFlag(t, action, "request-file")
		if !requestFile.Required || requestFile.Type != "string" ||
			!requestFile.RootJSONObjectFile || requestFile.Short != "f" {
			t.Errorf("%s request-file contract = %+v", actionName, requestFile)
		}

		idempotency := qmsAuditFlag(t, action, "idempotency-key")
		if !idempotency.Required || idempotency.Type != "string" ||
			idempotency.HeaderName != "Idempotency-Key" || idempotency.BodyName != "" ||
			!strings.Contains(idempotency.Description, "8-128 byte") {
			t.Errorf("%s idempotency contract = %+v", actionName, idempotency)
		}

		concurrency, hasConcurrency := qmsAuditFindFlag(action, "concurrency-token")
		if hasConcurrency != want.concurrency {
			t.Errorf("%s concurrency flag present = %v, want %v", actionName, hasConcurrency, want.concurrency)
		} else if hasConcurrency && (!concurrency.Required || concurrency.Type != "string" ||
			!concurrency.Sensitive || concurrency.HeaderName != "If-Match" || !concurrency.StrongETag) {
			t.Errorf("%s concurrency contract = %+v", actionName, concurrency)
		}

		confirmation, hasConfirmation := qmsAuditFindFlag(action, "confirm")
		if hasConfirmation != want.confirmation {
			t.Errorf("%s confirmation flag present = %v, want %v", actionName, hasConfirmation, want.confirmation)
		} else if hasConfirmation && (!confirmation.Required || confirmation.Type != "bool" ||
			!confirmation.LocalOnly || !confirmation.MustBeTrue) {
			t.Errorf("%s confirmation contract = %+v", actionName, confirmation)
		}
	}

	for _, action := range domain.Actions {
		if _, found := qmsAuditFindFlag(action, "dry-run"); found {
			t.Errorf("%s must not advertise an unsupported dry-run", action.Name)
		}
	}
}

func qmsAuditFindFlag(action Action, name string) (FlagDef, bool) {
	for _, flag := range action.Flags {
		if flag.Name == name {
			return flag, true
		}
	}
	return FlagDef{}, false
}

func TestQualityAuditNonEmptyUUIDValidationIsOptIn(t *testing.T) {
	t.Parallel()

	nonEmptyAction := Action{
		Name: "get",
		Args: []ArgDef{{Name: "recordGuid", Required: true, Type: "non-empty-uuid"}},
	}
	if err := validateActionInput(&cobra.Command{}, []string{emptyUUIDValue}, nonEmptyAction); err == nil ||
		!strings.Contains(err.Error(), "non-empty GUID") {
		t.Fatalf("non-empty UUID accepted Guid.Empty: %v", err)
	}
	if err := validateActionInput(&cobra.Command{}, []string{qmsAuditValidGUID}, nonEmptyAction); err != nil {
		t.Fatalf("non-empty UUID rejected valid GUID: %v", err)
	}

	legacyAction := Action{
		Name: "get",
		Args: []ArgDef{{Name: "recordGuid", Required: true, Type: "uuid"}},
	}
	if err := validateActionInput(&cobra.Command{}, []string{emptyUUIDValue}, legacyAction); err != nil {
		t.Fatalf("legacy uuid behavior changed: %v", err)
	}
}

func TestQualityAuditOptionalGUIDQueryRejectsGuidEmpty(t *testing.T) {
	t.Parallel()
	action := Action{
		Name: "search",
		Flags: []FlagDef{{
			Name:      "record-guid",
			QueryName: "recordGuid",
			Type:      "non-empty-uuid",
		}},
	}
	command := buildActionCommand(&Domain{Name: "audit"}, action, nil, nil, nil, nil)
	if err := command.Flags().Set("record-guid", emptyUUIDValue); err != nil {
		t.Fatalf("set record-guid: %v", err)
	}
	if err := validateActionInput(command, nil, action); err == nil ||
		!strings.Contains(err.Error(), "non-empty GUID") {
		t.Fatalf("optional GUID query accepted Guid.Empty: %v", err)
	}
}

func TestQualityAuditRequiredGUIDFlagRejectsGuidEmpty(t *testing.T) {
	t.Parallel()
	action := Action{
		Name: "mutate",
		Flags: []FlagDef{{
			Name:     "record-guid",
			Required: true,
			Type:     "non-empty-uuid",
		}},
	}
	command := buildActionCommand(&Domain{Name: "audit"}, action, nil, nil, nil, nil)
	if err := command.Flags().Set("record-guid", emptyUUIDValue); err != nil {
		t.Fatalf("set record-guid: %v", err)
	}
	if err := validateActionInput(command, nil, action); err == nil ||
		!strings.Contains(err.Error(), "non-empty GUID") {
		t.Fatalf("required GUID flag accepted Guid.Empty: %v", err)
	}
}

func TestQualityAuditConfirmationFailsClosed(t *testing.T) {
	t.Parallel()
	domain := qmsAuditDomain(t, "audit", "/api/quality/audits", 12)
	action := qmsAuditAction(t, domain, "transition")
	command := buildActionCommand(domain, action, nil, nil, nil, nil)
	if err := command.Flags().Set("confirm", "false"); err != nil {
		t.Fatalf("set confirm: %v", err)
	}
	if err := validateActionInput(
		command,
		[]string{qmsAuditValidGUID, qualityAuditTransitionActionKeys[0]},
		action,
	); err == nil || !strings.Contains(err.Error(), "--confirm must be explicitly enabled") {
		t.Fatalf("confirmed mutation accepted --confirm=false: %v", err)
	}
}

func TestQualityAuditRepresentativeRequestsPreserveTransportContracts(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	if err := auth.SaveToken(&auth.TokenData{
		AccessToken: "quality-audit-contract-token",
		ExpiresAt:   time.Now().Add(time.Hour),
		TenantGUID:  "55555555-5555-4555-8555-555555555555",
	}); err != nil {
		t.Fatalf("save test token: %v", err)
	}

	tests := []struct {
		name             string
		domainName       string
		domainPath       string
		domainActions    int
		actionName       string
		arguments        []string
		method           string
		path             string
		requestJSON      string
		wantBody         map[string]any
		idempotencyKey   string
		concurrencyToken string
		confirmation     bool
	}{
		{
			name:           "audit program create sends an unwrapped evolving DTO",
			domainName:     "audit-program",
			domainPath:     "/api/quality/audit-programs",
			domainActions:  7,
			actionName:     "create",
			method:         http.MethodPost,
			path:           "/api/quality/audit-programs",
			requestJSON:    `{"title":"FY 2027 audits","futureField":{"enabled":true}}`,
			wantBody:       map[string]any{"title": "FY 2027 audits", "futureField": map[string]any{"enabled": true}},
			idempotencyKey: "audit-program-create-key-01",
		},
		{
			name:             "audit update formats a strong concurrency ETag",
			domainName:       "audit",
			domainPath:       "/api/quality/audits",
			domainActions:    12,
			actionName:       "update",
			arguments:        []string{qmsAuditValidGUID},
			method:           http.MethodPut,
			path:             "/api/quality/audits/" + qmsAuditValidGUID,
			requestJSON:      `{"title":"Pump audit revised","futureField":{"enabled":true}}`,
			wantBody:         map[string]any{"title": "Pump audit revised", "futureField": map[string]any{"enabled": true}},
			idempotencyKey:   "audit-update-key-02",
			concurrencyToken: "audit-version-4",
		},
		{
			name:             "audit finding transition keeps confirmation local",
			domainName:       "audit-finding",
			domainPath:       "/api/quality/audit-findings",
			domainActions:    11,
			actionName:       "transition",
			arguments:        []string{qmsAuditValidGUID, "quality-audit-finding.issue"},
			method:           http.MethodPost,
			path:             "/api/quality/audit-findings/" + qmsAuditValidGUID + "/transitions/quality-audit-finding.issue",
			requestJSON:      `{"reason":"Issue the reviewed finding"}`,
			wantBody:         map[string]any{"reason": "Issue the reviewed finding"},
			idempotencyKey:   "finding-transition-key-03",
			concurrencyToken: "finding-version-9",
			confirmation:     true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			requestPath := writeRegistryJSONFixture(t, test.requestJSON)
			server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
				if request.Method != test.method {
					t.Errorf("method = %s, want %s", request.Method, test.method)
				}
				if request.URL.Path != test.path {
					t.Errorf("path = %q, want %q", request.URL.Path, test.path)
				}
				if request.URL.RawQuery != "" {
					t.Errorf("query = %q; local confirmation or route identity leaked", request.URL.RawQuery)
				}
				if request.Header.Get("Idempotency-Key") != test.idempotencyKey {
					t.Errorf("Idempotency-Key = %q, want %q", request.Header.Get("Idempotency-Key"), test.idempotencyKey)
				}
				wantIfMatch := ""
				if test.concurrencyToken != "" {
					wantIfMatch = `"` + test.concurrencyToken + `"`
				}
				if request.Header.Get("If-Match") != wantIfMatch {
					t.Errorf("If-Match = %q, want %q", request.Header.Get("If-Match"), wantIfMatch)
				}

				bodyBytes, err := io.ReadAll(request.Body)
				if err != nil {
					t.Errorf("read body: %v", err)
				}
				var body map[string]any
				if err := json.Unmarshal(bodyBytes, &body); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if !reflect.DeepEqual(body, test.wantBody) {
					t.Errorf("body = %#v, want unwrapped %#v", body, test.wantBody)
				}
				for _, forbidden := range []string{"request", "confirm", "auditGuid", "findingGuid", "programmeGuid", "actionKey"} {
					if _, leaked := body[forbidden]; leaked {
						t.Errorf("local or route-only field %q leaked into body: %#v", forbidden, body)
					}
				}

				response.Header().Set("Content-Type", "application/json")
				_, _ = response.Write([]byte(`{}`))
			}))
			t.Cleanup(server.Close)

			apiClient := client.NewAPIClient(
				server.URL,
				time.Second,
				true,
				client.RetryOptions{MaxRetries: 0},
				logging.New(logging.LevelError),
			)
			domain := qmsAuditDomain(t, test.domainName, test.domainPath, test.domainActions)
			format := "json"
			command := buildDomainCommand(
				domain,
				func() (*client.APIClient, error) { return apiClient, nil },
				logging.New(logging.LevelError),
				&format,
				&ExportConfig{},
			)
			command.SilenceErrors = true
			command.SilenceUsage = true
			commandArguments := append([]string{test.actionName}, test.arguments...)
			commandArguments = append(
				commandArguments,
				"--request-file", requestPath,
				"--idempotency-key", test.idempotencyKey,
			)
			if test.concurrencyToken != "" {
				commandArguments = append(commandArguments, "--concurrency-token", test.concurrencyToken)
			}
			if test.confirmation {
				commandArguments = append(commandArguments, "--confirm")
			}
			command.SetArgs(commandArguments)
			if err := command.Execute(); err != nil {
				t.Fatalf("%s command error = %v", test.actionName, err)
			}
		})
	}
}
