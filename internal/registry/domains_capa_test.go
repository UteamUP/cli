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

const (
	testCorrectivePreventiveActionGUID = "11111111-1111-4111-8111-111111111111"
	testCAPAEvidenceGUID               = "22222222-2222-4222-8222-222222222222"
	testCAPANonConformanceGUID         = "33333333-3333-4333-8333-333333333333"
	testCAPASourceLinkGUID             = "44444444-4444-4444-8444-444444444444"
)

func findCAPADomain(t *testing.T) *Domain {
	t.Helper()
	for _, domain := range DefaultRegistry.Domains() {
		if domain.Name == "capa" {
			return domain
		}
	}
	t.Fatal("capa domain is not registered")
	return nil
}

func findCAPAAction(t *testing.T, domain *Domain, name string) Action {
	t.Helper()
	for _, action := range domain.Actions {
		if action.Name == name {
			return action
		}
	}
	t.Fatalf("capa action %q is not registered", name)
	return Action{}
}

func TestCAPADomainMirrorsTheNineGovernedOperations(t *testing.T) {
	t.Parallel()
	domain := findCAPADomain(t)

	if domain.APIPath != "/api/quality/corrective-preventive-actions" {
		t.Fatalf("APIPath = %q", domain.APIPath)
	}
	wantAliases := []string{"corrective-preventive-action", "corrective-preventive-actions"}
	if !reflect.DeepEqual(domain.Aliases, wantAliases) {
		t.Fatalf("aliases = %v, want %v", domain.Aliases, wantAliases)
	}

	type actionContract struct {
		tool       string
		method     string
		restPath   string
		domainBase bool
	}
	want := map[string]actionContract{
		"search": {
			tool:       "UteamupCorrectivePreventiveActionSearch",
			method:     "GET",
			domainBase: true,
		},
		"get": {
			tool:     "UteamupCorrectivePreventiveActionGet",
			method:   "GET",
			restPath: "{correctivePreventiveActionGuid}",
		},
		"create": {
			tool:   "UteamupCorrectivePreventiveActionCreate",
			method: "POST",
		},
		"update": {
			tool:     "UteamupCorrectivePreventiveActionUpdate",
			method:   "PUT",
			restPath: "{correctivePreventiveActionGuid}",
		},
		"transition": {
			tool:     "UteamupCorrectivePreventiveActionTransition",
			method:   "POST",
			restPath: "{correctivePreventiveActionGuid}/transitions/{actionKey}",
		},
		"evidence-add": {
			tool:     "UteamupCorrectivePreventiveActionEvidenceAdd",
			method:   "POST",
			restPath: "{correctivePreventiveActionGuid}/evidence",
		},
		"evidence-revoke": {
			tool:     "UteamupCorrectivePreventiveActionEvidenceRevoke",
			method:   "POST",
			restPath: "{correctivePreventiveActionGuid}/evidence/{evidenceGuid}/revoke",
		},
		"source-ncr-add": {
			tool:     "UteamupCorrectivePreventiveActionSourceNonConformanceAdd",
			method:   "POST",
			restPath: "{correctivePreventiveActionGuid}/source-non-conformances/{sourceNonConformanceGuid}",
		},
		"source-ncr-revoke": {
			tool:     "UteamupCorrectivePreventiveActionSourceNonConformanceRevoke",
			method:   "POST",
			restPath: "{correctivePreventiveActionGuid}/source-non-conformances/{sourceLinkGuid}/revoke",
		},
	}

	if len(domain.Actions) != len(want) {
		t.Fatalf("actions = %d, want %d", len(domain.Actions), len(want))
	}
	for _, action := range domain.Actions {
		contract, ok := want[action.Name]
		if !ok {
			t.Errorf("unexpected action %q", action.Name)
			continue
		}
		if action.ToolName != contract.tool || action.HTTPMethod != contract.method ||
			action.RESTPath != contract.restPath || action.UseDomainBasePath != contract.domainBase {
			t.Errorf("%s declaration = %+v, want %+v", action.Name, action, contract)
		}
		if err := validateActionDefinition(action); err != nil {
			t.Errorf("%s declaration is invalid: %v", action.Name, err)
		}
		delete(want, action.Name)
	}
	if len(want) != 0 {
		t.Errorf("missing actions: %v", want)
	}
}

func TestCAPARoutesUseOnlyExactGUIDScopedPaths(t *testing.T) {
	t.Parallel()
	domain := findCAPADomain(t)
	tests := []struct {
		name string
		args map[string]any
		want string
	}{
		{name: "search", want: "/api/quality/corrective-preventive-actions"},
		{name: "get", args: map[string]any{"correctivePreventiveActionGuid": testCorrectivePreventiveActionGUID}, want: "/api/quality/corrective-preventive-actions/" + testCorrectivePreventiveActionGUID},
		{name: "create", want: "/api/quality/corrective-preventive-actions"},
		{name: "update", args: map[string]any{"correctivePreventiveActionGuid": testCorrectivePreventiveActionGUID}, want: "/api/quality/corrective-preventive-actions/" + testCorrectivePreventiveActionGUID},
		{name: "transition", args: map[string]any{"correctivePreventiveActionGuid": testCorrectivePreventiveActionGUID, "actionKey": "capa.submit"}, want: "/api/quality/corrective-preventive-actions/" + testCorrectivePreventiveActionGUID + "/transitions/capa.submit"},
		{name: "evidence-add", args: map[string]any{"correctivePreventiveActionGuid": testCorrectivePreventiveActionGUID}, want: "/api/quality/corrective-preventive-actions/" + testCorrectivePreventiveActionGUID + "/evidence"},
		{name: "evidence-revoke", args: map[string]any{"correctivePreventiveActionGuid": testCorrectivePreventiveActionGUID, "evidenceGuid": testCAPAEvidenceGUID}, want: "/api/quality/corrective-preventive-actions/" + testCorrectivePreventiveActionGUID + "/evidence/" + testCAPAEvidenceGUID + "/revoke"},
		{name: "source-ncr-add", args: map[string]any{"correctivePreventiveActionGuid": testCorrectivePreventiveActionGUID, "sourceNonConformanceGuid": testCAPANonConformanceGUID}, want: "/api/quality/corrective-preventive-actions/" + testCorrectivePreventiveActionGUID + "/source-non-conformances/" + testCAPANonConformanceGUID},
		{name: "source-ncr-revoke", args: map[string]any{"correctivePreventiveActionGuid": testCorrectivePreventiveActionGUID, "sourceLinkGuid": testCAPASourceLinkGUID}, want: "/api/quality/corrective-preventive-actions/" + testCorrectivePreventiveActionGUID + "/source-non-conformances/" + testCAPASourceLinkGUID + "/revoke"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			action := findCAPAAction(t, domain, test.name)
			path, _ := buildRESTPath(domain, action, test.args)
			if path != test.want {
				t.Errorf("path = %q, want %q", path, test.want)
			}
			if strings.ContainsAny(path, "{}") {
				t.Errorf("path contains an unexpanded placeholder: %s", path)
			}
		})
	}
}

func TestCAPASearchFiltersAreExplicitQueryParameters(t *testing.T) {
	t.Parallel()
	domain := findCAPADomain(t)
	search := findCAPAAction(t, domain, "search")

	want := map[string]FlagDef{
		"owning-site-location-guid": {QueryName: "owningSiteLocationGuid", Type: "uuid"},
		"status":                    {QueryName: "status", Type: "string"},
		"type":                      {QueryName: "type", Type: "string"},
		"initial-risk-level":        {QueryName: "initialRiskLevel", Type: "string"},
		"owner-user-guid":           {QueryName: "ownerUserGuid", Type: "uuid"},
		"query":                     {QueryName: "query", Type: "string"},
		"page":                      {QueryName: "page", Type: "int", Default: 1},
		"page-size":                 {QueryName: "pageSize", Type: "int", Default: 25},
	}
	if len(search.Flags) != len(want) {
		t.Fatalf("search flags = %d, want %d", len(search.Flags), len(want))
	}
	for _, flag := range search.Flags {
		expected, ok := want[flag.Name]
		if !ok {
			t.Errorf("unexpected search flag %q", flag.Name)
			continue
		}
		if flag.QueryName != expected.QueryName || flag.Type != expected.Type ||
			!reflect.DeepEqual(flag.Default, expected.Default) {
			t.Errorf("%s = %+v, want QueryName=%q Type=%q Default=%v", flag.Name, flag, expected.QueryName, expected.Type, expected.Default)
		}
		if flag.BodyName != "" || flag.HeaderName != "" {
			t.Errorf("search filter %q can escape the query string: %+v", flag.Name, flag)
		}
		delete(want, flag.Name)
	}
	if len(want) != 0 {
		t.Errorf("missing search flags: %v", want)
	}
}

func TestCAPAMutationsRequireRootJSONIdempotencyAndConcurrency(t *testing.T) {
	t.Parallel()
	domain := findCAPADomain(t)
	mutationNames := []string{
		"create",
		"update",
		"transition",
		"evidence-add",
		"evidence-revoke",
		"source-ncr-add",
		"source-ncr-revoke",
	}
	confirmationActions := map[string]bool{
		"transition":        true,
		"evidence-revoke":   true,
		"source-ncr-revoke": true,
	}

	for _, name := range mutationNames {
		action := findCAPAAction(t, domain, name)
		flags := capaFlagsByName(action)

		requestFile, ok := flags["request-file"]
		if !ok || !requestFile.Required || requestFile.Type != "string" ||
			!requestFile.RootJSONObjectFile || requestFile.JSONFile ||
			requestFile.BodyName != "" || requestFile.HeaderName != "" || requestFile.QueryName != "" {
			t.Errorf("%s request-file = %+v", name, requestFile)
		}
		idempotency, ok := flags["idempotency-key"]
		if !ok || !idempotency.Required || idempotency.Type != "uuid" ||
			idempotency.HeaderName != "Idempotency-Key" || idempotency.BodyName != "" {
			t.Errorf("%s idempotency-key = %+v", name, idempotency)
		}

		concurrency, hasConcurrency := flags["concurrency-token"]
		if name == "create" {
			if hasConcurrency {
				t.Errorf("create unexpectedly exposes concurrency-token: %+v", concurrency)
			}
		} else if !hasConcurrency || !concurrency.Required || concurrency.Type != "string" ||
			concurrency.HeaderName != "If-Match" || !concurrency.StrongETag || !concurrency.Sensitive {
			t.Errorf("%s concurrency-token = %+v", name, concurrency)
		}

		confirm, hasConfirm := flags["confirm"]
		if confirmationActions[name] {
			if !hasConfirm || !confirm.Required || confirm.Type != "bool" ||
				!confirm.MustBeTrue || !confirm.LocalOnly || confirm.Default != nil ||
				confirm.BodyName != "" || confirm.HeaderName != "" || confirm.QueryName != "" {
				t.Errorf("%s confirmation = %+v", name, confirm)
			}
		} else if hasConfirm {
			t.Errorf("%s unexpectedly requires high-impact confirmation", name)
		}
	}
}

func TestCAPAPublicSurfaceHasNoIntegerOrIdentitySpoofingInputs(t *testing.T) {
	t.Parallel()
	domain := findCAPADomain(t)
	forbidden := map[string]bool{
		"id":                true,
		"tenant":            true,
		"tenantid":          true,
		"tenantguid":        true,
		"user":              true,
		"userid":            true,
		"userguid":          true,
		"actor":             true,
		"actorid":           true,
		"actorguid":         true,
		"createdbyuserid":   true,
		"createdbyuserguid": true,
	}

	for _, action := range domain.Actions {
		for _, argument := range action.Args {
			if argument.Type == "int" {
				t.Errorf("%s exposes integer positional argument %+v", action.Name, argument)
			}
			assertCAPAInputNameIsNotServerOwned(t, action.Name, argument.Name, forbidden)
			assertCAPAInputNameIsNotServerOwned(t, action.Name, argument.BodyName, forbidden)
			assertCAPAInputNameIsNotServerOwned(t, action.Name, argument.QueryName, forbidden)
		}
		for _, flag := range action.Flags {
			if flag.Type == "int" && flag.QueryName == "" {
				t.Errorf("%s exposes non-pagination integer flag %+v", action.Name, flag)
			}
			assertCAPAInputNameIsNotServerOwned(t, action.Name, flag.Name, forbidden)
			assertCAPAInputNameIsNotServerOwned(t, action.Name, flag.BodyName, forbidden)
			assertCAPAInputNameIsNotServerOwned(t, action.Name, flag.QueryName, forbidden)
		}
	}
}

func TestCAPATransitionKeysAreTheExactClosedAllowlist(t *testing.T) {
	t.Parallel()
	domain := findCAPADomain(t)
	transition := findCAPAAction(t, domain, "transition")
	want := []string{
		"capa.submit",
		"capa.require-rca",
		"capa.complete-rca",
		"capa.complete-plan",
		"capa.approve-plan",
		"capa.return-plan",
		"capa.complete-actions",
		"capa.reach-effectiveness-window",
		"capa.approve-effective",
		"capa.mark-ineffective",
		"capa.resubmit-plan",
		"capa.reopen",
		"capa.resume-assessment",
	}

	if !reflect.DeepEqual(correctivePreventiveActionTransitionActionKeys, want) {
		t.Fatalf("declared transition keys = %v, want %v", correctivePreventiveActionTransitionActionKeys, want)
	}
	if len(transition.Args) != 2 || !reflect.DeepEqual(transition.Args[1].AllowedValues, want) {
		t.Fatalf("transition action-key argument = %+v, want exact allowlist %v", transition.Args, want)
	}

	assertCAPAValidationError(t, transition, []string{testCorrectivePreventiveActionGUID, "capa.submit/../close"}, "invalid actionKey")
	get := findCAPAAction(t, domain, "get")
	assertCAPAValidationError(t, get, []string{"42"}, "invalid correctivePreventiveActionGuid")
	sourceAdd := findCAPAAction(t, domain, "source-ncr-add")
	assertCAPAValidationError(t, sourceAdd, []string{testCorrectivePreventiveActionGUID, "17"}, "invalid sourceNonConformanceGuid")
}

func TestCAPARevocationsDeclareRetainOnlyAssociationSemantics(t *testing.T) {
	t.Parallel()
	domain := findCAPADomain(t)
	evidenceDescription := strings.ToLower(findCAPAAction(t, domain, "evidence-revoke").Description)
	for _, required := range []string{
		"association",
		"does not delete",
		"document",
		"documentversion",
		"sharepoint",
		"microsoft graph",
		"provider content",
		"immutable history",
	} {
		if !strings.Contains(evidenceDescription, required) {
			t.Errorf("evidence-revoke description %q does not mention %q", evidenceDescription, required)
		}
	}
	sourceDescription := strings.ToLower(findCAPAAction(t, domain, "source-ncr-revoke").Description)
	for _, required := range []string{"association", "does not delete", "source ncr", "immutable link history"} {
		if !strings.Contains(sourceDescription, required) {
			t.Errorf("source-ncr-revoke description %q does not mention %q", sourceDescription, required)
		}
	}
}

func TestCAPASourceNCRAddKeepsRouteIdentitySeparateFromRootBody(t *testing.T) {
	domain := findCAPADomain(t)
	requestPath := writeRegistryJSONFixture(t, `{
		"nonConformanceGuid":"`+testCAPANonConformanceGUID+`",
		"sourceReason":"Recurring supplier escape"
	}`)
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	if err := auth.SaveToken(&auth.TokenData{
		AccessToken: "capa-contract-token",
		ExpiresAt:   time.Now().Add(time.Hour),
		TenantGUID:  "55555555-5555-4555-8555-555555555555",
	}); err != nil {
		t.Fatalf("save test token: %v", err)
	}

	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", request.Method)
		}
		wantPath := "/api/quality/corrective-preventive-actions/" +
			testCorrectivePreventiveActionGUID +
			"/source-non-conformances/" + testCAPANonConformanceGUID
		if request.URL.Path != wantPath {
			t.Errorf("path = %q, want %q", request.URL.Path, wantPath)
		}
		if request.Header.Get("Idempotency-Key") != testCAPAEvidenceGUID {
			t.Errorf("Idempotency-Key = %q", request.Header.Get("Idempotency-Key"))
		}
		if request.Header.Get("If-Match") != `"capa-version-3"` {
			t.Errorf("If-Match = %q", request.Header.Get("If-Match"))
		}
		bodyBytes, err := io.ReadAll(request.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
		}
		var body map[string]any
		if err := json.Unmarshal(bodyBytes, &body); err != nil {
			t.Errorf("decode body: %v", err)
		}
		wantBody := map[string]any{
			"nonConformanceGuid": testCAPANonConformanceGUID,
			"sourceReason":       "Recurring supplier escape",
		}
		if !reflect.DeepEqual(body, wantBody) {
			t.Errorf("body = %#v, want %#v", body, wantBody)
		}
		if _, leaked := body["sourceNonConformanceGuid"]; leaked {
			t.Errorf("route-only sourceNonConformanceGuid leaked into body: %#v", body)
		}
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{"correctivePreventiveAction":{}}`))
	}))
	t.Cleanup(server.Close)

	apiClient := client.NewAPIClient(
		server.URL,
		time.Second,
		true,
		client.RetryOptions{MaxRetries: 0},
		logging.New(logging.LevelError),
	)
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
	command.SetArgs([]string{
		"source-ncr-add",
		testCorrectivePreventiveActionGUID,
		testCAPANonConformanceGUID,
		"--request-file", requestPath,
		"--idempotency-key", testCAPAEvidenceGUID,
		"--concurrency-token", "capa-version-3",
	})
	if err := command.Execute(); err != nil {
		t.Fatalf("source-ncr-add command error = %v", err)
	}
}

func capaFlagsByName(action Action) map[string]FlagDef {
	flags := make(map[string]FlagDef, len(action.Flags))
	for _, flag := range action.Flags {
		flags[flag.Name] = flag
	}
	return flags
}

func assertCAPAInputNameIsNotServerOwned(t *testing.T, actionName, name string, forbidden map[string]bool) {
	t.Helper()
	canonical := strings.NewReplacer("-", "", "_", "").Replace(strings.ToLower(name))
	if forbidden[canonical] {
		t.Errorf("%s exposes server-owned input %q", actionName, name)
	}
}

func assertCAPAValidationError(t *testing.T, action Action, args []string, want string) {
	t.Helper()
	command := &cobra.Command{}
	if err := validateActionInput(command, args, action); err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("validateActionInput() error = %v, want containing %q", err, want)
	}
}
