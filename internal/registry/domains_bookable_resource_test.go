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
		"list":                {"UteamupBookableResourceList", "GET", ""},
		"crew-list":           {"UteamupBookableResourceCrewList", "GET", "crews"},
		"source-list":         {"UteamupBookableResourceSourceList", "GET", "sources"},
		"availability-search": {"UteamupBookableResourceAvailabilitySearch", "GET", "availability-search"},
		"get":                 {"UteamupBookableResourceGet", "GET", "{resourceGuid}"},
		"create":              {"UteamupBookableResourceCreate", "POST", ""},
		"update":              {"UteamupBookableResourceUpdate", "PUT", "{resourceGuid}"},
		"pool-members-set":    {"UteamupBookableResourcePoolMembersSet", "PUT", "{poolGuid}/members"},
		"territory-list":      {"UteamupServiceTerritoryList", "GET", "territories"},
		"territory-create":    {"UteamupServiceTerritoryCreate", "POST", "territories"},
		"territory-update":    {"UteamupServiceTerritoryUpdate", "PUT", "territories/{territoryGuid}"},
		"requirement-list":    {"UteamupBookableResourceRequirementList", "GET", "workorders/{workorderGuid}/requirements"},
		"requirement-create":  {"UteamupBookableResourceRequirementCreate", "POST", "workorders/{workorderGuid}/requirements"},
		"requirement-update":  {"UteamupBookableResourceRequirementUpdate", "PUT", "requirements/{requirementGuid}"},
		"route-estimate":      {"UteamupBookableResourceRouteEstimate", "POST", "route-estimate"},
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

func TestBookableResourceSourceFiltersUsePublicGuids(t *testing.T) {
	action := findAction(findDomain("bookable-resource"), "list")
	expected := map[string]string{
		"user-guid": "userGuid", "contractor-profile-guid": "contractorProfileGuid",
		"contractor-crew-guid": "contractorCrewGuid", "asset-guid": "assetGuid", "tool-guid": "toolGuid",
		"location-guid": "locationGuid",
	}
	for _, flag := range action.Flags {
		if want, ok := expected[flag.Name]; ok {
			if flag.BodyName != want || flag.Type != "string" || flag.Required {
				t.Errorf("%s must be an optional GUID filter mapped to %s: %+v", flag.Name, want, flag)
			}
			delete(expected, flag.Name)
		}
	}
	if len(expected) != 0 {
		t.Fatalf("missing resource source filters: %v", expected)
	}
}

func TestBookableResourceSaveKeysUseTheRetryHeader(t *testing.T) {
	for _, name := range []string{"create", "update"} {
		action := findAction(findDomain("bookable-resource"), name)
		found := false
		for _, flag := range action.Flags {
			if flag.Name == "idempotency-key" {
				found = true
				if flag.HeaderName != "Idempotency-Key" || flag.Type != "string" || flag.BodyName != "" {
					t.Errorf("%s key must use the GUID retry header: %+v", name, flag)
				}
			}
		}
		if !found {
			t.Errorf("%s retry key is missing", name)
		}
	}
}

func TestBookableResourcePoolSearchExposesMemberTypeFilter(t *testing.T) {
	action := findAction(findDomain("bookable-resource"), "list")
	for _, flag := range action.Flags {
		if flag.Name == "pool-member-resource-type" {
			if flag.BodyName != "poolMemberResourceType" || flag.Type != "int" {
				t.Fatalf("unexpected pool member filter: %+v", flag)
			}
			return
		}
	}
	t.Fatal("pool member resource type filter is missing")
}

func TestBookableResourceCrewLookupExposesSearchAndSavedGuid(t *testing.T) {
	action := findAction(findDomain("bookable-resource"), "crew-list")
	if action == nil {
		t.Fatal("crew-list action is missing")
	}
	expected := map[string]string{"search": "search", "crew-guid": "crewGuid", "page": "page", "page-size": "pageSize"}
	for _, flag := range action.Flags {
		if want, ok := expected[flag.Name]; ok {
			if flag.BodyName != want {
				t.Errorf("%s maps to %q, want %q", flag.Name, flag.BodyName, want)
			}
			if flag.Name == "crew-guid" && (flag.Type != "string" || flag.Required) {
				t.Errorf("crew-guid must be an optional public GUID string: %+v", flag)
			}
			delete(expected, flag.Name)
		}
	}
	if len(expected) != 0 {
		t.Fatalf("missing crew lookup filters: %v", expected)
	}
}

func TestBookableResourceSourceLookupRequiresTypeAndPublicIdentity(t *testing.T) {
	action := findAction(findDomain("bookable-resource"), "source-list")
	if action == nil {
		t.Fatal("source-list action is missing")
	}
	expected := map[string]string{"resource-type": "resourceType", "search": "search", "source-guid": "sourceGuid", "page": "page", "page-size": "pageSize"}
	for _, flag := range action.Flags {
		if want, ok := expected[flag.Name]; ok {
			if flag.BodyName != want {
				t.Errorf("%s maps to %q, want %q", flag.Name, flag.BodyName, want)
			}
			if flag.Name == "source-guid" && (flag.Type != "string" || flag.Required) {
				t.Errorf("source-guid must be an optional public GUID string: %+v", flag)
			}
			if flag.Name == "resource-type" && (flag.Type != "int" || !flag.Required) {
				t.Errorf("resource-type must be required: %+v", flag)
			}
			delete(expected, flag.Name)
		}
	}
	if len(expected) != 0 {
		t.Fatalf("missing source lookup filters: %v", expected)
	}
}

func TestBookableResourceUpdatesSendReviewedVersionInQuery(t *testing.T) {
	domain := findDomain("bookable-resource")
	for _, actionName := range []string{
		"update",
		"pool-members-set",
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
	for _, actionName := range []string{"create", "update", "pool-members-set"} {
		action := findAction(domain, actionName)
		if action == nil {
			t.Fatalf("%s action is not registered", actionName)
		}
		flagName, bodyName := "pool-members-file", "poolMembers"
		required := actionName == "pool-members-set"
		if required {
			flagName, bodyName = "members-file", "members"
		}
		found := false
		for _, flag := range action.Flags {
			if flag.Name == flagName {
				found = true
				if flag.Required != required || !flag.JSONFile || flag.BodyName != bodyName {
					t.Errorf("%s must send the member list as %s: %+v", actionName, bodyName, flag)
				}
			}
		}
		if !found {
			t.Errorf("%s is missing --%s", actionName, flagName)
		}
	}
}

func TestBookableResourceMutationsAcceptToolSource(t *testing.T) {
	domain := findDomain("bookable-resource")
	for _, actionName := range []string{"create", "update"} {
		action := findAction(domain, actionName)
		found := false
		for _, flag := range action.Flags {
			if flag.Name == "tool-guid" {
				found = true
				if flag.BodyName != "toolGuid" || flag.Type != "string" || flag.Required {
					t.Errorf("%s --tool-guid must be an optional public GUID mapped to toolGuid: %+v", actionName, flag)
				}
			}
		}
		if !found {
			t.Errorf("%s is missing --tool-guid", actionName)
		}
	}
}

func TestBookableResourceTypeHelpListsEveryType(t *testing.T) {
	domain := findDomain("bookable-resource")
	for _, actionName := range []string{"list", "create", "update", "requirement-create", "requirement-update"} {
		action := findAction(domain, actionName)
		for _, flag := range action.Flags {
			if flag.Name != "resource-type" {
				continue
			}
			for _, label := range []string{"0=technician", "1=contractor", "2=crew", "3=equipment", "4=vehicle", "5=facility", "6=pool", "7=tool"} {
				if !strings.Contains(flag.Description, label) {
					t.Errorf("%s --resource-type help is missing %q: %q", actionName, label, flag.Description)
				}
			}
		}
	}
	for _, check := range []struct{ action, flag string }{
		{"list", "pool-member-resource-type"},
		{"source-list", "resource-type"},
	} {
		for _, flag := range findAction(domain, check.action).Flags {
			if flag.Name == check.flag && !strings.Contains(flag.Description, "7=tool") {
				t.Errorf("%s --%s help is missing the tool type: %q", check.action, check.flag, flag.Description)
			}
		}
	}
}

func TestBookableResourceAvailabilitySearchSendsQueryParameters(t *testing.T) {
	action := findAction(findDomain("bookable-resource"), "availability-search")
	if action == nil {
		t.Fatal("availability-search action is missing")
	}
	expected := map[string]struct {
		query    string
		kind     string
		required bool
		def      any
	}{
		"from-utc":             {"fromUtc", "string", true, nil},
		"to-utc":               {"toUtc", "string", true, nil},
		"types":                {"types", "stringSlice", false, nil},
		"search":               {"search", "string", false, nil},
		"available-only":       {"availableOnly", "bool", false, nil},
		"include-not-rostered": {"includeNotRostered", "bool", false, true},
		"page-size-per-type":   {"pageSizePerType", "int", false, 10},
		"cursor":               {"cursor", "string", false, nil},
	}
	for _, flag := range action.Flags {
		want, ok := expected[flag.Name]
		if !ok {
			t.Errorf("unexpected availability flag %+v", flag)
			continue
		}
		if flag.QueryName != want.query || flag.Type != want.kind || flag.Required != want.required || flag.Default != want.def {
			t.Errorf("%s = %+v, want query %q type %q required %v default %v", flag.Name, flag, want.query, want.kind, want.required, want.def)
		}
		delete(expected, flag.Name)
	}
	if len(expected) != 0 {
		t.Fatalf("missing availability flags: %v", expected)
	}

	path := appendQueryParameters("/api/bookableresources/availability-search", map[string]any{
		"types": []string{"technician", "tool"},
	})
	if path != "/api/bookableresources/availability-search?types=technician&types=tool" {
		t.Fatalf("types must repeat as separate query values, got %q", path)
	}
}
