package registry

import "testing"

func TestFloorplanReadsUsePublicGuidArguments(t *testing.T) {
	domain := findDomain("floorplan")
	if domain == nil {
		t.Fatal("floorplan domain is missing")
	}
	if domain.APIPath != "/api/floorplan" {
		t.Fatalf("APIPath = %q", domain.APIPath)
	}
	for _, action := range domain.Actions {
		for _, arg := range action.Args {
			if arg.Type != "uuid" || !arg.Required {
				t.Fatalf("%s argument %s is not a required GUID", action.Name, arg.Name)
			}
		}
	}
	want := map[string]string{
		"get": "{guid}", "by-location": "location/{locationGuid}",
	}
	for _, action := range domain.Actions {
		if path, ok := want[action.Name]; ok {
			if action.HTTPMethod != "GET" || action.RESTPath != path || action.MCPOnly {
				t.Fatalf("%s does not use its controller read route", action.Name)
			}
			delete(want, action.Name)
		}
		if action.Name == "download" || action.Name == "by-location-floor" {
			if !action.MCPOnly {
				t.Fatalf("%s must not fall back to a nonexistent JSON REST route", action.Name)
			}
		}
	}
	if len(want) != 0 {
		t.Fatalf("missing controller reads: %v", want)
	}
}
