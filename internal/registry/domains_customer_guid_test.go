package registry

import "testing"

func TestCustomerActionsUseGuidRoutesAndReviewedWrites(t *testing.T) {
	domain := findDomainByName(t, "customer")
	if domain.APIPath != "/api/customer" {
		t.Fatalf("customer API path = %q", domain.APIPath)
	}
	for _, name := range []string{"get", "update", "delete"} {
		action := findDomainAction(t, "customer", name)
		if action.MCPOnly || action.RESTPath != "{customerGuid}" || len(action.Args) != 1 {
			t.Fatalf("%s must use the normal GUID route", name)
		}
		arg := action.Args[0]
		if arg.Name != "customerGuid" || arg.Type != "non-empty-uuid" || !arg.Required {
			t.Errorf("%s identity is not a required nonempty public GUID", name)
		}
	}
	for _, name := range []string{"create", "update", "delete"} {
		action := findDomainAction(t, "customer", name)
		confirm := actionFlagByName(t, action, "confirm")
		if !confirm.Required || !confirm.MustBeTrue || !confirm.LocalOnly {
			t.Errorf("%s must require confirmation without changing the REST DTO", name)
		}
		if name != "delete" {
			fields := actionFlagByName(t, action, "from-json")
			if !fields.Required || !fields.RootJSONObjectFile {
				t.Errorf("%s requires the complete reviewed JSON object", name)
			}
		}
	}
	search := findDomainAction(t, "customer", "search")
	if search.HTTPMethod != "GET" || !search.UseDomainBasePath || search.RESTPath != "" || search.Args[0].QueryName != "nameFilter" {
		t.Fatal("search must use the existing paginated nameFilter endpoint")
	}
}
