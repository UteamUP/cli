package registry

import "testing"

func TestProjectGovernanceDomainsAreGuidFirst(t *testing.T) {
	domains := []string{
		"project-field-context",
		"project-member",
		"project-dependency",
		"project-activity",
		"project-comment",
		"project-baseline",
		"project-change-request",
		"project-source-line",
		"project-requirement",
	}

	for _, domainName := range domains {
		domain := findDomain(domainName)
		if domain == nil {
			t.Errorf("expected %s domain to be registered", domainName)
			continue
		}
		for _, action := range domain.Actions {
			if len(action.Args) == 0 || action.Args[0].Name != "projectGuid" || action.Args[0].Type != "string" {
				t.Errorf("%s %s must start with a string projectGuid argument, got %+v", domainName, action.Name, action.Args)
			}
			for _, argument := range action.Args {
				if argument.Type == "int" && argument.Name != "limit" {
					t.Errorf("%s %s leaks an integer identity argument: %+v", domainName, action.Name, argument)
				}
			}
		}
	}
}

func TestProjectFieldContextReadRoute(t *testing.T) {
	action := findDomainAction(t, "project-field-context", "get")
	if action.HTTPMethod != "" ||
		action.RESTPath != "{projectGuid}/field-context" ||
		action.ToolName != "UteamupProjectFieldContextGet" {
		t.Errorf(
			"project-field-context get: want GET {projectGuid}/field-context UteamupProjectFieldContextGet, got %s %s %s",
			action.HTTPMethod,
			action.RESTPath,
			action.ToolName,
		)
	}
}

func TestProjectGovernanceMutationRoutes(t *testing.T) {
	cases := []struct {
		domain string
		action string
		method string
		path   string
		tool   string
	}{
		{"project-member", "add", "POST", "{projectGuid}/members", "UteamupProjectMembersAdd"},
		{"project-member", "update", "PUT", "{projectGuid}/members/{memberGuid}", "UteamupProjectMembersUpdate"},
		{"project-dependency", "remove", "DELETE", "{projectGuid}/dependencies/{dependencyGuid}", "UteamupProjectDependenciesRemove"},
		{"project-comment", "update", "PUT", "{projectGuid}/comments/{commentGuid}", "UteamupProjectCommentsUpdate"},
		{"project-baseline", "capture", "POST", "{projectGuid}/baselines", "UteamupProjectBaselinesCapture"},
		{"project-baseline", "scope", "", "{projectGuid}/baselines/{baselineGuid}/scope", "UteamupProjectBaselineScopeGet"},
		{"project-change-request", "apply", "POST", "{projectGuid}/change-requests/{changeRequestGuid}/apply", "UteamupProjectChangeRequestsApply"},
		{"project-change-request", "get", "", "{projectGuid}/change-requests/{requestGuid}", "UteamupProjectChangeRequestsGet"},
		{"project-change-request", "history", "", "{projectGuid}/change-requests/{requestGuid}/history", "UteamupProjectChangeRequestsHistory"},
		{"project-source-line", "list", "", "{projectGuid}/intake/lines", "UteamupProjectSourceLinesList"},
		{"project-requirement", "get", "", "{projectGuid}/requirements/{requirementGuid}", "UteamupProjectRequirementGet"},
		{"project-requirement", "verification", "", "{projectGuid}/requirements/verification-reference/{stepGuid}", "UteamupProjectVerificationReferenceGet"},
	}

	for _, testCase := range cases {
		action := findDomainAction(t, testCase.domain, testCase.action)
		if action.HTTPMethod != testCase.method || action.RESTPath != testCase.path || action.ToolName != testCase.tool {
			t.Errorf("%s %s: want %s %s %s, got %s %s %s", testCase.domain, testCase.action,
				testCase.method, testCase.path, testCase.tool, action.HTTPMethod, action.RESTPath, action.ToolName)
		}
	}
}

func TestProjectChangeDecisionsAcceptVersionedPayload(t *testing.T) {
	for _, name := range []string{"submit", "approve", "reject", "apply"} {
		action := findDomainAction(t, "project-change-request", name)
		found := false
		for _, flag := range action.Flags {
			if flag.Name == "from-json" {
				found = true
			}
		}
		if !found {
			t.Errorf("%s must accept requestGuid, expectedUpdatedAt and the inspected review fingerprint", name)
		}
	}
}
