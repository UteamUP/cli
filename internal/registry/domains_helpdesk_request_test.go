package registry

import "testing"

func helpdeskDomainAction(t *testing.T, domainName, actionName string) (*Domain, Action) {
	t.Helper()
	domain := findDomain(domainName)
	if domain == nil {
		t.Fatalf("%s domain is not registered", domainName)
	}
	for _, action := range domain.Actions {
		if action.Name == actionName {
			return domain, action
		}
	}
	t.Fatalf("%s action %q is not registered", domainName, actionName)
	return nil, Action{}
}

func TestHelpdeskRequestRoutesMirrorTheController(t *testing.T) {
	cases := []struct {
		action, method, path string
	}{
		{"queue", "GET", "queue"},
		{"review", "PUT", "{requestGuid}/status"},
		{"convert", "POST", "{requestGuid}/convert"},
		{"create", "POST", ""},
	}
	for _, tc := range cases {
		domain, action := helpdeskDomainAction(t, "helpdesk-request", tc.action)
		if domain.APIPath != "/api/helpdesk/requests" {
			t.Fatalf("unexpected API path: %q", domain.APIPath)
		}
		if action.HTTPMethod != tc.method || action.RESTPath != tc.path {
			t.Fatalf("%s: got %s %q, want %s %q", tc.action, action.HTTPMethod, action.RESTPath, tc.method, tc.path)
		}
	}
}

func TestHelpdeskRequestIdentifiersAreGuids(t *testing.T) {
	for _, name := range []string{"review", "convert"} {
		_, action := helpdeskDomainAction(t, "helpdesk-request", name)
		if len(action.Args) != 1 || action.Args[0].Name != "requestGuid" || action.Args[0].Type != "uuid" {
			t.Fatalf("%s must take exactly one UUID requestGuid argument: %#v", name, action.Args)
		}
	}
	_, convert := helpdeskDomainAction(t, "helpdesk-request", "convert")
	if len(convert.Flags) != 1 || convert.Flags[0].BodyName != "templateGuid" || convert.Flags[0].Type != "uuid" || convert.Flags[0].Required {
		t.Fatalf("convert must take an optional UUID templateGuid: %#v", convert.Flags)
	}
}

func TestHelpdeskQueueFiltersTravelOnTheQueryString(t *testing.T) {
	for _, name := range []string{"queue", "list"} {
		_, action := helpdeskDomainAction(t, "helpdesk-request", name)
		for _, flag := range action.Flags {
			if flag.QueryName == "" {
				t.Fatalf("%s flag --%s must be a query parameter on a GET", name, flag.Name)
			}
		}
	}
}

func TestHelpdeskIntakeSetSendsOnlyTheIntakeFields(t *testing.T) {
	domain, set := helpdeskDomainAction(t, "helpdesk-intake", "set")
	if domain.APIPath != "/api/tenant" || set.HTTPMethod != "PATCH" || set.RESTPath != "{tenantGuid}/helpdesk-settings" {
		t.Fatalf("unexpected intake set route: %s %s/%s", set.HTTPMethod, domain.APIPath, set.RESTPath)
	}
	// The PATCH must not carry license counts or approval flags: the backend treats an omitted
	// field as unchanged, so the CLI can never zero a tenant's extra helpdesk licenses.
	for _, flag := range set.Flags {
		if flag.BodyName != "helpdeskIntakeMode" && flag.BodyName != "helpdeskDefaultTemplateGuid" {
			t.Fatalf("unexpected body field %q on helpdesk-intake set", flag.BodyName)
		}
	}
	_, get := helpdeskDomainAction(t, "helpdesk-intake", "get")
	if get.HTTPMethod != "GET" || get.RESTPath != "{tenantGuid}/helpdesk-licenses" {
		t.Fatalf("unexpected intake get route: %s %q", get.HTTPMethod, get.RESTPath)
	}
}
