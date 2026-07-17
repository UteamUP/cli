package registry

import (
	"strings"
	"testing"
)

func TestImprovementProjectDomainUsesGuidFirstRoutes(t *testing.T) {
	domain := findImprovementDomain(t)
	if domain.APIPath != "/api/improvementproject" {
		t.Fatalf("APIPath = %q, want /api/improvementproject", domain.APIPath)
	}

	expectedPaths := map[string]string{
		"get":               "by-guid/{projectGuid}",
		"update":            "by-guid/{projectGuid}",
		"delete":            "by-guid/{projectGuid}",
		"approve":           "by-guid/{projectGuid}/approve",
		"start":             "by-guid/{projectGuid}/start",
		"complete":          "by-guid/{projectGuid}/complete",
		"cancel":            "by-guid/{projectGuid}/cancel",
		"hold":              "by-guid/{projectGuid}/hold",
		"create-workorder":  "by-guid/{projectGuid}/create-workorder",
		"linked-workorders": "by-guid/{projectGuid}/linked-workorders",
		"actions":           "by-guid/{projectGuid}/actions",
		"action-create":     "by-guid/{projectGuid}/actions",
		"action-update":     "by-guid/{projectGuid}/actions/by-guid/{actionGuid}",
		"action-delete":     "by-guid/{projectGuid}/actions/by-guid/{actionGuid}",
		"action-complete":   "by-guid/{projectGuid}/actions/by-guid/{actionGuid}/complete",
		"pdca":              "by-guid/{projectGuid}/pdca",
		"pdca-add":          "by-guid/{projectGuid}/pdca",
		"pdca-complete":     "by-guid/{projectGuid}/pdca/by-guid/{entryGuid}/complete",
	}

	for name, expectedPath := range expectedPaths {
		action := findImprovementAction(t, domain, name)
		if action.RESTPath != expectedPath {
			t.Errorf("%s RESTPath = %q, want %q", name, action.RESTPath, expectedPath)
		}
		if strings.Contains(action.RESTPath, "{id}") {
			t.Errorf("%s exposes an integer identity in %q", name, action.RESTPath)
		}
		for _, arg := range action.Args {
			if arg.Name == "id" || arg.Type == "int" {
				t.Errorf("%s exposes non-GUID identity argument %+v", name, arg)
			}
			if strings.HasSuffix(arg.Name, "Guid") && arg.Type != "uuid" {
				t.Errorf("%s argument %s type = %q, want uuid", name, arg.Name, arg.Type)
			}
		}
	}
}

func TestImprovementProjectNestedRoutesExpandAllGuids(t *testing.T) {
	domain := findImprovementDomain(t)
	testCases := []struct {
		actionName string
		args       map[string]any
		expected   string
	}{
		{
			actionName: "action-complete",
			args: map[string]any{
				"projectGuid": "11111111-1111-1111-1111-111111111111",
				"actionGuid":  "22222222-2222-2222-2222-222222222222",
			},
			expected: "/api/improvementproject/by-guid/11111111-1111-1111-1111-111111111111/actions/by-guid/22222222-2222-2222-2222-222222222222/complete",
		},
		{
			actionName: "pdca-complete",
			args: map[string]any{
				"projectGuid": "11111111-1111-1111-1111-111111111111",
				"entryGuid":   "33333333-3333-3333-3333-333333333333",
			},
			expected: "/api/improvementproject/by-guid/11111111-1111-1111-1111-111111111111/pdca/by-guid/33333333-3333-3333-3333-333333333333/complete",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.actionName, func(t *testing.T) {
			action := findImprovementAction(t, domain, testCase.actionName)
			path, consumed := buildRESTPath(domain, *action, testCase.args)
			if path != testCase.expected {
				t.Fatalf("path = %q, want %q", path, testCase.expected)
			}
			if len(consumed) != len(testCase.args) {
				t.Errorf("consumed %d arguments, want %d", len(consumed), len(testCase.args))
			}
		})
	}
}

func findImprovementDomain(t *testing.T) *Domain {
	t.Helper()
	for _, domain := range DefaultRegistry.Domains() {
		if domain.Name == "improvement-project" {
			return domain
		}
	}
	t.Fatal("improvement-project domain is not registered")
	return nil
}

func findImprovementAction(t *testing.T, domain *Domain, name string) *Action {
	t.Helper()
	for index := range domain.Actions {
		if domain.Actions[index].Name == name {
			return &domain.Actions[index]
		}
	}
	t.Fatalf("%s action is not registered for %s", name, domain.Name)
	return nil
}
