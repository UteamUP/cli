package registry

import (
	"strings"
	"testing"
)

func serviceAgreementAction(t *testing.T, name string) (*Domain, Action) {
	t.Helper()
	domain := findDomain("service-agreement")
	if domain == nil {
		t.Fatal("service-agreement domain is not registered")
	}
	for _, action := range domain.Actions {
		if action.Name == name {
			return domain, action
		}
	}
	t.Fatalf("service-agreement action %q is not registered", name)
	return nil, Action{}
}

func TestServiceAgreementRoutesAreGuidOnly(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name string
		path string
	}{
		{name: "get", path: "/api/service-agreements/agreement-guid"},
		{name: "update", path: "/api/service-agreements/agreement-guid"},
		{name: "approve", path: "/api/service-agreements/agreement-guid/approve"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			domain, action := serviceAgreementAction(t, testCase.name)
			path, consumed := buildRESTPath(domain, action, map[string]any{"agreementGuid": "agreement-guid"})
			if path != testCase.path {
				t.Fatalf("path = %q, want %q", path, testCase.path)
			}
			if len(consumed) != 1 || strings.Contains(strings.ToLower(action.RESTPath), "{id}") {
				t.Fatalf("route is not GUID-only: %q consumed=%v", action.RESTPath, consumed)
			}
			if len(action.Args) != 1 || action.Args[0].Type != "uuid" {
				t.Fatalf("public identity argument is not a UUID: %+v", action.Args)
			}
		})
	}
}

func TestServiceAgreementWritesCarryIdempotencyAndReviewedVersion(t *testing.T) {
	t.Parallel()
	for _, actionName := range []string{"create", "update", "approve"} {
		_, action := serviceAgreementAction(t, actionName)
		assertServiceAgreementFlag(t, action, "idempotency-key", "idempotencyKey", true)
	}
	for _, actionName := range []string{"update", "approve"} {
		_, action := serviceAgreementAction(t, actionName)
		assertServiceAgreementFlag(t, action, "expected-updated-at", "expectedUpdatedAt", true)
	}
	_, create := serviceAgreementAction(t, "create")
	assertServiceAgreementFlag(t, create, "contract-guid", "contractGuid", true)
	assertServiceAgreementFlag(t, create, "customer-guid", "customerGuid", true)
	coverage := findServiceAgreementFlag(create, "coverage-json")
	if coverage == nil || !coverage.JSONFile || coverage.BodyName != "coverage" {
		t.Fatalf("coverage must be supplied as reviewed JSON: %+v", coverage)
	}
}

func assertServiceAgreementFlag(t *testing.T, action Action, name, bodyName string, required bool) {
	t.Helper()
	flag := findServiceAgreementFlag(action, name)
	if flag == nil || flag.BodyName != bodyName || flag.Required != required {
		t.Fatalf("%s flag = %+v, want body=%q required=%t", name, flag, bodyName, required)
	}
}

func findServiceAgreementFlag(action Action, name string) *FlagDef {
	for index := range action.Flags {
		if action.Flags[index].Name == name {
			return &action.Flags[index]
		}
	}
	return nil
}
