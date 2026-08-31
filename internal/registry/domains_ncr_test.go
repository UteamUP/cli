package registry

import (
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

const (
	testNonConformanceGUID = "11111111-1111-4111-8111-111111111111"
	testEvidenceGUID       = "22222222-2222-4222-8222-222222222222"
	testTargetGUID         = "33333333-3333-4333-8333-333333333333"
	testLinkGUID           = "44444444-4444-4444-8444-444444444444"
)

func findNCRDomain(t *testing.T) *Domain {
	t.Helper()
	for _, domain := range DefaultRegistry.Domains() {
		if domain.Name == "ncr" {
			return domain
		}
	}
	t.Fatal("ncr domain is not registered")
	return nil
}

func findNCRAction(t *testing.T, domain *Domain, name string) Action {
	t.Helper()
	for _, action := range domain.Actions {
		if action.Name == name {
			return action
		}
	}
	t.Fatalf("ncr action %q is not registered", name)
	return Action{}
}

func TestNCRDomainMirrorsTheNineGovernedOperations(t *testing.T) {
	t.Parallel()
	domain := findNCRDomain(t)

	if domain.APIPath != "/api/quality/non-conformances" {
		t.Fatalf("APIPath = %q", domain.APIPath)
	}
	wantAliases := []string{"nonconformance", "non-conformance"}
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
			tool:       "UteamupNonConformanceSearch",
			method:     "GET",
			domainBase: true,
		},
		"get": {
			tool:     "UteamupNonConformanceGet",
			method:   "GET",
			restPath: "{nonConformanceGuid}",
		},
		"create": {
			tool:   "UteamupNonConformanceCreate",
			method: "POST",
		},
		"update": {
			tool:     "UteamupNonConformanceUpdate",
			method:   "PUT",
			restPath: "{nonConformanceGuid}",
		},
		"transition": {
			tool:     "UteamupNonConformanceTransition",
			method:   "POST",
			restPath: "{nonConformanceGuid}/transitions/{actionKey}",
		},
		"evidence-add": {
			tool:     "UteamupNonConformanceEvidenceAdd",
			method:   "POST",
			restPath: "{nonConformanceGuid}/evidence",
		},
		"evidence-revoke": {
			tool:     "UteamupNonConformanceEvidenceRevoke",
			method:   "POST",
			restPath: "{nonConformanceGuid}/evidence/{evidenceGuid}/revoke",
		},
		"link-add": {
			tool:     "UteamupNonConformanceLinkAdd",
			method:   "POST",
			restPath: "{nonConformanceGuid}/links/{linkKind}/{targetGuid}",
		},
		"link-revoke": {
			tool:     "UteamupNonConformanceLinkRevoke",
			method:   "POST",
			restPath: "{nonConformanceGuid}/links/{linkKind}/{linkGuid}/revoke",
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

func TestNCRRoutesUseOnlyExactGUIDScopedPaths(t *testing.T) {
	t.Parallel()
	domain := findNCRDomain(t)
	tests := []struct {
		name string
		args map[string]any
		want string
	}{
		{name: "search", want: "/api/quality/non-conformances"},
		{name: "get", args: map[string]any{"nonConformanceGuid": testNonConformanceGUID}, want: "/api/quality/non-conformances/" + testNonConformanceGUID},
		{name: "create", want: "/api/quality/non-conformances"},
		{name: "update", args: map[string]any{"nonConformanceGuid": testNonConformanceGUID}, want: "/api/quality/non-conformances/" + testNonConformanceGUID},
		{name: "transition", args: map[string]any{"nonConformanceGuid": testNonConformanceGUID, "actionKey": "ncr.submit"}, want: "/api/quality/non-conformances/" + testNonConformanceGUID + "/transitions/ncr.submit"},
		{name: "evidence-add", args: map[string]any{"nonConformanceGuid": testNonConformanceGUID}, want: "/api/quality/non-conformances/" + testNonConformanceGUID + "/evidence"},
		{name: "evidence-revoke", args: map[string]any{"nonConformanceGuid": testNonConformanceGUID, "evidenceGuid": testEvidenceGUID}, want: "/api/quality/non-conformances/" + testNonConformanceGUID + "/evidence/" + testEvidenceGUID + "/revoke"},
		{name: "link-add", args: map[string]any{"nonConformanceGuid": testNonConformanceGUID, "linkKind": "workorder", "targetGuid": testTargetGUID}, want: "/api/quality/non-conformances/" + testNonConformanceGUID + "/links/workorder/" + testTargetGUID},
		{name: "link-revoke", args: map[string]any{"nonConformanceGuid": testNonConformanceGUID, "linkKind": "project", "linkGuid": testLinkGUID}, want: "/api/quality/non-conformances/" + testNonConformanceGUID + "/links/project/" + testLinkGUID + "/revoke"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			action := findNCRAction(t, domain, test.name)
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

func TestNCRSearchFiltersAreExplicitQueryParameters(t *testing.T) {
	t.Parallel()
	domain := findNCRDomain(t)
	search := findNCRAction(t, domain, "search")

	want := map[string]FlagDef{
		"owning-site-location-guid": {QueryName: "owningSiteLocationGuid", Type: "uuid"},
		"status":                    {QueryName: "status", Type: "string"},
		"severity":                  {QueryName: "severity", Type: "string"},
		"risk-level":                {QueryName: "riskLevel", Type: "string"},
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

func TestNCRMutationsRequireRootJSONIdempotencyAndConcurrency(t *testing.T) {
	t.Parallel()
	domain := findNCRDomain(t)
	mutationNames := []string{
		"create",
		"update",
		"transition",
		"evidence-add",
		"evidence-revoke",
		"link-add",
		"link-revoke",
	}
	confirmationActions := map[string]bool{
		"transition":      true,
		"evidence-revoke": true,
		"link-revoke":     true,
	}

	for _, name := range mutationNames {
		action := findNCRAction(t, domain, name)
		flags := ncrFlagsByName(action)

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

func TestNCRPublicSurfaceHasNoIntegerOrActorSpoofingInputs(t *testing.T) {
	t.Parallel()
	domain := findNCRDomain(t)
	forbidden := map[string]bool{
		"id":                true,
		"tenantid":          true,
		"tenantguid":        true,
		"userid":            true,
		"userguid":          true,
		"actorid":           true,
		"actorguid":         true,
		"reporteruserid":    true,
		"reporteruserguid":  true,
		"createdbyuserid":   true,
		"createdbyuserguid": true,
		"source":            true,
	}

	for _, action := range domain.Actions {
		for _, argument := range action.Args {
			if argument.Type == "int" {
				t.Errorf("%s exposes integer positional argument %+v", action.Name, argument)
			}
			assertNCRInputNameIsNotServerOwned(t, action.Name, argument.Name, forbidden)
			assertNCRInputNameIsNotServerOwned(t, action.Name, argument.BodyName, forbidden)
			assertNCRInputNameIsNotServerOwned(t, action.Name, argument.QueryName, forbidden)
		}
		for _, flag := range action.Flags {
			if flag.Type == "int" && flag.QueryName == "" {
				t.Errorf("%s exposes non-pagination integer flag %+v", action.Name, flag)
			}
			assertNCRInputNameIsNotServerOwned(t, action.Name, flag.Name, forbidden)
			assertNCRInputNameIsNotServerOwned(t, action.Name, flag.BodyName, forbidden)
			assertNCRInputNameIsNotServerOwned(t, action.Name, flag.QueryName, forbidden)
		}
	}
}

func TestNCRRouteSegmentsAreClosedAllowlists(t *testing.T) {
	t.Parallel()
	domain := findNCRDomain(t)
	transition := findNCRAction(t, domain, "transition")
	if !reflect.DeepEqual(transition.Args[1].AllowedValues, nonConformanceTransitionActionKeys) {
		t.Fatalf("transition action keys = %v", transition.Args[1].AllowedValues)
	}

	wantLinkKinds := []string{
		"workorder",
		"project",
		"asset",
		"location",
		"part",
		"stock-item",
		"vendor",
		"vehicle-inspection",
		"root-cause-analysis",
		"operational-route-execution",
	}
	for _, name := range []string{"link-add", "link-revoke"} {
		action := findNCRAction(t, domain, name)
		if !reflect.DeepEqual(action.Args[1].AllowedValues, wantLinkKinds) {
			t.Errorf("%s link kinds = %v, want %v", name, action.Args[1].AllowedValues, wantLinkKinds)
		}
		if containsExact(action.Args[1].AllowedValues, "customer-quality-case") {
			t.Errorf("%s exposes unavailable customer-quality-case relationships", name)
		}
	}

	assertNCRValidationError(t, transition, []string{testNonConformanceGUID, "ncr.submit/../close"}, "invalid actionKey")
	linkAdd := findNCRAction(t, domain, "link-add")
	assertNCRValidationError(t, linkAdd, []string{testNonConformanceGUID, "customer-quality-case", testTargetGUID}, "invalid linkKind")
	get := findNCRAction(t, domain, "get")
	assertNCRValidationError(t, get, []string{"42"}, "invalid nonConformanceGuid")
}

func ncrFlagsByName(action Action) map[string]FlagDef {
	flags := make(map[string]FlagDef, len(action.Flags))
	for _, flag := range action.Flags {
		flags[flag.Name] = flag
	}
	return flags
}

func assertNCRInputNameIsNotServerOwned(t *testing.T, actionName, name string, forbidden map[string]bool) {
	t.Helper()
	canonical := strings.NewReplacer("-", "", "_", "").Replace(strings.ToLower(name))
	if forbidden[canonical] {
		t.Errorf("%s exposes server-owned input %q", actionName, name)
	}
}

func assertNCRValidationError(t *testing.T, action Action, args []string, want string) {
	t.Helper()
	command := &cobra.Command{}
	if err := validateActionInput(command, args, action); err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("validateActionInput() error = %v, want containing %q", err, want)
	}
}
