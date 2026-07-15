package registry

import (
	"strings"
	"testing"
)

func customerPortalUserAction(t *testing.T, name string) (*Domain, Action) {
	t.Helper()
	domain := findDomain("customer-portal-user")
	if domain == nil {
		t.Fatal("customer-portal-user domain is not registered")
	}

	for _, action := range domain.Actions {
		if action.Name == name {
			return domain, action
		}
	}

	t.Fatalf("customer-portal-user action %q is not registered", name)
	return nil, Action{}
}

func TestCustomerPortalUserRoutesAreGuidOnly(t *testing.T) {
	t.Parallel()

	for _, actionName := range []string{"get", "update", "delete"} {
		domain, action := customerPortalUserAction(t, actionName)
		path, consumed := buildRESTPath(domain, action, map[string]any{
			"userExternalGuid": "user-guid",
		})
		if path != "/api/customerportalusers/by-guid/user-guid" {
			t.Fatalf("%s path = %q", actionName, path)
		}
		if len(consumed) != 1 {
			t.Fatalf("%s consumed = %v, want one GUID arg", actionName, consumed)
		}
		if len(action.Args) != 1 || action.Args[0].Type != "uuid" {
			t.Fatalf("%s args = %+v, want one UUID arg", actionName, action.Args)
		}
		if strings.Contains(strings.ToLower(action.RESTPath), "{id") {
			t.Fatalf("%s exposes an integer identity route: %q", actionName, action.RESTPath)
		}
	}
}

func TestCustomerPortalUserCreateUsesCustomerGuidAndSensitivePassword(t *testing.T) {
	t.Parallel()

	_, action := customerPortalUserAction(t, "create")
	flags := make(map[string]FlagDef, len(action.Flags))
	for _, flag := range action.Flags {
		flags[flag.Name] = flag
	}

	customer := flags["customer-external-guid"]
	if !customer.Required || customer.BodyName != "customerExternalGuid" || customer.Type != "string" {
		t.Fatalf("customer GUID flag = %+v", customer)
	}
	if _, exists := flags["customer-id"]; exists {
		t.Fatal("integer customer-id flag must not be registered")
	}
	if !flags["password"].Sensitive {
		t.Fatal("password flag must be redacted from diagnostic logs")
	}
}

func TestCustomerPortalUserDomainDoesNotExposeLogin(t *testing.T) {
	t.Parallel()

	domain := findDomain("customer-portal-user")
	if domain == nil {
		t.Fatal("customer-portal-user domain is not registered")
	}
	for _, action := range domain.Actions {
		if action.Name == "login" {
			t.Fatal("disabled portal login must not be exposed by the CLI")
		}
	}
}
