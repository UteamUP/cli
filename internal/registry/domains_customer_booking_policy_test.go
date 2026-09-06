package registry

import "testing"

func TestCustomerBookingServiceTypesUseThePolicyBoundaryAndGuidFilter(t *testing.T) {
	domain := findDomain("customer-booking-policy")
	if domain == nil || domain.APIPath != "/api/customerbookingpolicies" {
		t.Fatal("booking-policy lookup must use its own policy API boundary")
	}
	action := findAction(domain, "service-types")
	if action == nil || action.HTTPMethod != "GET" || action.RESTPath != "service-types" || action.ToolName != "UteamupCustomerBookingServiceTypeList" {
		t.Fatal("service-type lookup does not match the backend REST and MCP contract")
	}
	flags := map[string]FlagDef{}
	for _, flag := range action.Flags {
		flags[flag.Name] = flag
	}
	if flags["service-type-guid"].BodyName != "serviceTypeGuid" || flags["service-type-guid"].Type != "string" {
		t.Fatal("service-type selection must carry a public GUID")
	}
	if flags["page"].Default != 1 || flags["page-size"].Default != 25 {
		t.Fatal("lookup must retain bounded pagination defaults")
	}
}
