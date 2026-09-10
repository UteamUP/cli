package registry

import "testing"

func TestAssetStateDomainMirrorsTheControllerRoutes(t *testing.T) {
	domain := findDomain("asset-state")
	if domain == nil {
		t.Fatal("asset-state domain is not registered")
	}
	if domain.APIPath != "/api/assetstate" {
		t.Fatalf("unexpected API path: %q", domain.APIPath)
	}

	actions := map[string]*Action{}
	for index := range domain.Actions {
		actions[domain.Actions[index].Name] = &domain.Actions[index]
	}

	// Every route is keyed by the asset's public GUID, never its integer id.
	for _, name := range []string{"get", "history", "availability", "set-operational"} {
		action := actions[name]
		if action == nil {
			t.Fatalf("%s action is missing", name)
		}
		if len(action.Args) != 1 || action.Args[0].Name != "assetGuid" ||
			action.Args[0].Type != "non-empty-uuid" {
			t.Fatalf("%s must take a single assetGuid argument, got %+v", name, action.Args)
		}
	}

	// Reads must stay reads: a GET that silently became a PUT would mutate on a lookup.
	for _, name := range []string{"get", "history", "availability"} {
		if actions[name].HTTPMethod != "GET" {
			t.Fatalf("%s must be a GET, got %q", name, actions[name].HTTPMethod)
		}
	}

	availability := actions["availability"]
	if len(availability.Flags) != 2 {
		t.Fatalf("availability needs both window bounds, got %+v", availability.Flags)
	}
	for _, flag := range availability.Flags {
		if !flag.Required {
			t.Fatalf("availability flag %q must be required — an unbounded window scans the whole log", flag.Name)
		}
		if flag.QueryName == "" {
			t.Fatalf("availability flag %q must travel as a query parameter on a GET", flag.Name)
		}
	}

	set := actions["set-operational"]
	if set.HTTPMethod != "PUT" {
		t.Fatalf("set-operational must be a PUT, got %q", set.HTTPMethod)
	}
	if set.RESTPath != "asset/by-guid/{assetGuid}/operational" {
		t.Fatalf("unexpected set-operational path: %q", set.RESTPath)
	}

	flags := map[string]*FlagDef{}
	for index := range set.Flags {
		flags[set.Flags[index].Name] = &set.Flags[index]
	}
	toState := flags["to-state"]
	if toState == nil || !toState.Required || toState.BodyName != "toState" {
		t.Fatalf("set-operational needs a required toState body flag, got %+v", toState)
	}
	// The reason and the originating work order are what make the interval log auditable.
	if flags["reason"] == nil || flags["workorder-guid"] == nil ||
		flags["workorder-guid"].BodyName != "workorderGuid" {
		t.Fatalf("set-operational must carry reason and workorderGuid, got %+v", set.Flags)
	}
}
