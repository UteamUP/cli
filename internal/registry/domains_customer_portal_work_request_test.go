package registry

import "testing"

func customerPortalWorkRequestDomain(t *testing.T) *Domain {
	t.Helper()
	domain := findDomain("customer-portal-work-request")
	if domain == nil {
		t.Fatal("customer-portal-work-request domain is not registered")
	}
	return domain
}

func customerPortalWorkRequestAction(t *testing.T, name string) Action {
	t.Helper()
	for _, action := range customerPortalWorkRequestDomain(t).Actions {
		if action.Name == name {
			return action
		}
	}
	t.Fatalf("customer-portal-work-request action %q is not registered", name)
	return Action{}
}

func TestCustomerPortalWorkRequestRoutesAreGuidOnly(t *testing.T) {
	domain := customerPortalWorkRequestDomain(t)
	if domain.APIPath != "/api/customerportalworkrequests" {
		t.Fatalf("unexpected API path: %q", domain.APIPath)
	}

	get := customerPortalWorkRequestAction(t, "get")
	status := customerPortalWorkRequestAction(t, "status-update")
	if get.RESTPath != "by-guid/{workRequestGuid}" {
		t.Fatalf("unexpected get path: %q", get.RESTPath)
	}
	if status.RESTPath != "by-guid/{workRequestGuid}/status" {
		t.Fatalf("unexpected status path: %q", status.RESTPath)
	}
	for _, action := range []Action{get, status} {
		if len(action.Args) != 1 || action.Args[0].Type != "uuid" {
			t.Fatalf("action %q must expose one UUID identity argument: %#v", action.Name, action.Args)
		}
	}
}

func TestCustomerPortalWorkRequestCreateUsesRelationshipGuids(t *testing.T) {
	create := customerPortalWorkRequestAction(t, "create")
	fields := map[string]FlagDef{}
	for _, flag := range create.Flags {
		fields[flag.BodyName] = flag
	}

	for _, field := range []string{"customerPortalUserExternalGuid", "customerExternalGuid"} {
		flag, ok := fields[field]
		if !ok || !flag.Required || flag.Type != "uuid" {
			t.Fatalf("GUID field %q is missing or invalid: %#v", field, flag)
		}
	}
	for _, legacy := range []string{"customerPortalUserId", "customerId"} {
		if _, ok := fields[legacy]; ok {
			t.Fatalf("legacy integer relationship field %q must not be exposed", legacy)
		}
	}
}
