package registry

import (
	"strings"
	"testing"
)

func serviceEntitlementAction(t *testing.T, name string) (*Domain, Action) {
	t.Helper()
	domain := findDomain("service-entitlement")
	if domain == nil {
		t.Fatal("service-entitlement domain is not registered")
	}
	for _, action := range domain.Actions {
		if action.Name == name {
			return domain, action
		}
	}
	t.Fatalf("service-entitlement action %q is not registered", name)
	return nil, Action{}
}

func TestServiceEntitlementRoutesAreGuidOnly(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name    string
		argName string
		arg     string
		path    string
	}{
		{name: "get", argName: "entitlementGuid", arg: "entitlement-guid", path: "/api/service-entitlements/entitlement-guid"},
		{name: "update", argName: "entitlementGuid", arg: "entitlement-guid", path: "/api/service-entitlements/entitlement-guid"},
		{name: "usage-record", argName: "entitlementGuid", arg: "entitlement-guid", path: "/api/service-entitlements/entitlement-guid/usage"},
		{name: "usage-reverse", argName: "usageGuid", arg: "usage-guid", path: "/api/service-entitlements/usage/usage-guid/reverse"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			domain, action := serviceEntitlementAction(t, testCase.name)
			path, consumed := buildRESTPath(
				domain,
				action,
				map[string]any{testCase.argName: testCase.arg},
			)
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

func TestServiceEntitlementWritesCarryIdempotencyAndEvidence(t *testing.T) {
	t.Parallel()
	for _, actionName := range []string{"create", "update", "usage-record", "usage-reverse"} {
		_, action := serviceEntitlementAction(t, actionName)
		assertServiceEntitlementFlag(t, action, "idempotency-key", "idempotencyKey", true)
	}
	_, update := serviceEntitlementAction(t, "update")
	assertServiceEntitlementFlag(t, update, "expected-updated-at", "expectedUpdatedAt", true)
	_, create := serviceEntitlementAction(t, "create")
	assertServiceEntitlementFlag(t, create, "agreement-guid", "agreementGuid", true)
	_, usage := serviceEntitlementAction(t, "usage-record")
	assertServiceEntitlementFlag(t, usage, "source-guid", "sourceGuid", true)
	assertServiceEntitlementFlag(t, usage, "occurred-at", "occurredAt", true)
}

func assertServiceEntitlementFlag(
	t *testing.T,
	action Action,
	name string,
	bodyName string,
	required bool,
) {
	t.Helper()
	for _, flag := range action.Flags {
		if flag.Name == name {
			if flag.BodyName != bodyName || flag.Required != required {
				t.Fatalf("%s flag = %+v, want body=%q required=%t", name, flag, bodyName, required)
			}
			return
		}
	}
	t.Fatalf("flag %q was not registered", name)
}
