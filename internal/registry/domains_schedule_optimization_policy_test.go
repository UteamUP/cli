package registry

import (
	"reflect"
	"testing"
)

func schedulePolicyAction(t *testing.T, name string) (*Domain, Action) {
	t.Helper()
	domain := findDomain("schedule-optimization-policy")
	if domain == nil {
		t.Fatal("schedule-optimization-policy domain is not registered")
	}
	for _, action := range domain.Actions {
		if action.Name == name {
			return domain, action
		}
	}
	t.Fatalf("schedule-optimization-policy action %q is not registered", name)
	return nil, Action{}
}

func TestScheduleOptimizationPolicyRoutesUseGuidIdentity(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name     string
		args     map[string]any
		path     string
		consumed []string
	}{
		{name: "list", args: map[string]any{}, path: "/api/schedule/optimization-policies"},
		{name: "get", args: map[string]any{"policyGuid": "policy-guid"}, path: "/api/schedule/optimization-policies/policy-guid", consumed: []string{"policyGuid"}},
		{name: "create", args: map[string]any{}, path: "/api/schedule/optimization-policies"},
		{name: "update", args: map[string]any{"policyGuid": "policy-guid"}, path: "/api/schedule/optimization-policies/policy-guid", consumed: []string{"policyGuid"}},
		{name: "delete", args: map[string]any{"policyGuid": "policy-guid"}, path: "/api/schedule/optimization-policies/policy-guid", consumed: []string{"policyGuid"}},
		{name: "restore", args: map[string]any{"policyGuid": "policy-guid"}, path: "/api/schedule/optimization-policies/policy-guid/restore", consumed: []string{"policyGuid"}},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			domain, action := schedulePolicyAction(t, testCase.name)
			path, consumed := buildRESTPath(domain, action, testCase.args)
			if path != testCase.path {
				t.Fatalf("path = %q, want %q", path, testCase.path)
			}
			if !reflect.DeepEqual(consumed, testCase.consumed) {
				t.Fatalf("consumed = %v, want %v", consumed, testCase.consumed)
			}
		})
	}
}

func TestScheduleOptimizationPolicyMutationsMirrorBackendContract(t *testing.T) {
	t.Parallel()
	_, create := schedulePolicyAction(t, "create")
	want := map[string]string{
		"idempotency-key": "idempotencyKey",
		"name":            "name",
		"enabled":         "isEnabled",
		"frequency":       "frequency",
		"time-zone":       "timeZoneId",
		"days-mask":       "daysOfWeekMask",
		"local-time":      "localExecutionTime",
		"horizon-days":    "horizonDays",
		"team-guids":      "teamGuids",
	}
	if len(create.Flags) != len(want) {
		t.Fatalf("create flags = %d, want %d", len(create.Flags), len(want))
	}
	for _, flag := range create.Flags {
		if want[flag.Name] != flag.BodyName {
			t.Fatalf("flag %q body = %q, want %q", flag.Name, flag.BodyName, want[flag.Name])
		}
		// The API serializes the frequency enum as a camelCase string; an int flag sent 0/1 and
		// could never express the contract the other clients read back (audit OP-003).
		if flag.Name == "frequency" {
			if flag.Type != "string" {
				t.Fatalf("frequency flag type = %q, want %q", flag.Type, "string")
			}
			if flag.Default != "daily" {
				t.Fatalf("frequency flag default = %v, want %q", flag.Default, "daily")
			}
		}
	}
}
