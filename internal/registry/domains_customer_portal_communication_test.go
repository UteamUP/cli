package registry

import "testing"

func TestCustomerPortalCommunicationDomainsAreGuidOnly(t *testing.T) {
	for _, name := range []string{"customer-message", "customer-rating"} {
		domain := findDomain(name)
		if domain == nil {
			t.Fatalf("%s domain is not registered", name)
		}
		if domain.APIPath != "/api/customerportal" {
			t.Fatalf("%s has unexpected API path %q", name, domain.APIPath)
		}
		for _, action := range domain.Actions {
			if len(action.Args) != 1 || action.Args[0].Name != "portalUserGuid" || action.Args[0].Type != "uuid" {
				t.Fatalf("%s %s must use a portalUserGuid UUID argument: %#v", name, action.Name, action.Args)
			}
			for _, flag := range action.Flags {
				if flag.BodyName == "toUserId" || flag.BodyName == "workorderId" || flag.BodyName == "projectId" {
					t.Fatalf("%s %s exposes legacy integer field %q", name, action.Name, flag.BodyName)
				}
			}
		}
	}
}
