package registry

import "testing"

func TestProjectGovernanceDomainsAreGuidFirst(t *testing.T) {
	domains := []string{
		"project-member",
		"project-dependency",
		"project-activity",
		"project-comment",
		"project-baseline",
		"project-change-request",
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

func TestProjectGovernanceMutationRoutes(t *testing.T) {
	cases := []struct {
		domain string
		action string
		method string
		path   string
		tool   string
	}{
		{"project-member", "add", "POST", "{projectGuid}/members", "uteamup_project_members_add"},
		{"project-member", "update", "PUT", "{projectGuid}/members/{memberGuid}", "uteamup_project_members_update"},
		{"project-dependency", "remove", "DELETE", "{projectGuid}/dependencies/{dependencyGuid}", "uteamup_project_dependencies_remove"},
		{"project-comment", "update", "PUT", "{projectGuid}/comments/{commentGuid}", "uteamup_project_comments_update"},
		{"project-baseline", "capture", "POST", "{projectGuid}/baselines", "uteamup_project_baselines_capture"},
		{"project-change-request", "apply", "POST", "{projectGuid}/change-requests/{requestGuid}/apply", "uteamup_project_change_requests_apply"},
	}

	for _, testCase := range cases {
		action := findDomainAction(t, testCase.domain, testCase.action)
		if action.HTTPMethod != testCase.method || action.RESTPath != testCase.path || action.ToolName != testCase.tool {
			t.Errorf("%s %s: want %s %s %s, got %s %s %s", testCase.domain, testCase.action,
				testCase.method, testCase.path, testCase.tool, action.HTTPMethod, action.RESTPath, action.ToolName)
		}
	}
}
