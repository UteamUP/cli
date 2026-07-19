package registry

import (
	"strings"
	"testing"
)

func TestBookableResourceDomainMirrorsBackendToolsAndGuidRoutes(t *testing.T) {
	domain := findDomain("bookable-resource")
	if domain == nil {
		t.Fatal("bookable-resource domain is not registered")
	}
	if domain.APIPath != "/api/bookableresources" {
		t.Fatalf("APIPath = %q, want /api/bookableresources", domain.APIPath)
	}

	expected := map[string]struct {
		tool   string
		method string
		path   string
	}{
		"list":               {"UteamupBookableResourceList", "GET", ""},
		"get":                {"UteamupBookableResourceGet", "GET", "{resourceGuid}"},
		"create":             {"UteamupBookableResourceCreate", "POST", ""},
		"update":             {"UteamupBookableResourceUpdate", "PUT", "{resourceGuid}"},
		"pool-members-set":   {"UteamupBookableResourcePoolMembersSet", "PUT", "{poolGuid}/members"},
		"territory-list":     {"UteamupServiceTerritoryList", "GET", "territories"},
		"territory-create":   {"UteamupServiceTerritoryCreate", "POST", "territories"},
		"territory-update":   {"UteamupServiceTerritoryUpdate", "PUT", "territories/{territoryGuid}"},
		"requirement-list":   {"UteamupBookableResourceRequirementList", "GET", "workorders/{workorderGuid}/requirements"},
		"requirement-create": {"UteamupBookableResourceRequirementCreate", "POST", "workorders/{workorderGuid}/requirements"},
		"requirement-update": {"UteamupBookableResourceRequirementUpdate", "PUT", "requirements/{requirementGuid}"},
		"route-estimate":     {"UteamupBookableResourceRouteEstimate", "POST", "route-estimate"},
	}

	for actionName, want := range expected {
		action := findAction(domain, actionName)
		if action == nil {
			t.Fatalf("action %q is not registered", actionName)
		}
		if action.ToolName != want.tool ||
			action.HTTPMethod != want.method ||
			action.RESTPath != want.path {
			t.Errorf(
				"%s = tool %q method %q path %q, want %q %q %q",
				actionName,
				action.ToolName,
				action.HTTPMethod,
				action.RESTPath,
				want.tool,
				want.method,
				want.path,
			)
		}
		for _, arg := range action.Args {
			if strings.HasSuffix(arg.Name, "Id") || arg.Type == "int" {
				t.Errorf("%s exposes database identifier argument %+v", actionName, arg)
			}
		}
	}
}

func TestBookableResourceUpdatesSendReviewedVersionInQuery(t *testing.T) {
	domain := findDomain("bookable-resource")
	for _, actionName := range []string{
		"update",
		"territory-update",
		"requirement-update",
	} {
		action := findAction(domain, actionName)
		if action == nil {
			t.Fatalf("action %q is not registered", actionName)
		}
		found := false
		for _, flag := range action.Flags {
			if flag.Name != "expected-updated-at" {
				continue
			}
			found = true
			if !flag.Required || flag.QueryName != "expectedUpdatedAt" {
				t.Errorf(
					"%s ExpectedUpdatedAt = required %v query %q",
					actionName,
					flag.Required,
					flag.QueryName,
				)
			}
		}
		if !found {
			t.Errorf("%s is missing --expected-updated-at", actionName)
		}
	}
}

func TestAppendQueryParametersPreservesPathAndEscapesReviewedTimestamp(t *testing.T) {
	path := appendQueryParameters(
		"/api/bookableresources/resource-guid",
		map[string]any{
			"expectedUpdatedAt": "2026-07-19T12:30:00+00:00",
		},
	)

	want := "/api/bookableresources/resource-guid?" +
		"expectedUpdatedAt=2026-07-19T12%3A30%3A00%2B00%3A00"
	if path != want {
		t.Fatalf("path = %q, want %q", path, want)
	}
}

func TestBookableResourceStructuredInputsUseJSONFiles(t *testing.T) {
	domain := findDomain("bookable-resource")
	action := findAction(domain, "pool-members-set")
	if action == nil {
		t.Fatal("pool-members-set action is not registered")
	}
	for _, flag := range action.Flags {
		if flag.Name == "members-file" {
			if !flag.Required || !flag.JSONFile || flag.BodyName != "members" {
				t.Fatalf("members-file flag is not a required members JSON payload: %+v", flag)
			}
			return
		}
	}
	t.Fatal("pool-members-set is missing --members-file")
}
