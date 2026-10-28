package registry

import "testing"

func TestStockPlannedIncomingGuidRoutesAndMethods(t *testing.T) {
	guid := "11111111-1111-4111-8111-111111111111"
	for _, entry := range []struct{ name, method, path string }{
		{"list", "GET", "/api/stock/items/" + guid + "/planned-incoming"},
		{"get", "GET", "/api/stock/planned-incoming/" + guid},
		{"create", "POST", "/api/stock/items/" + guid + "/planned-incoming"},
		{"update", "PUT", "/api/stock/planned-incoming/" + guid},
		{"convert", "POST", "/api/stock/planned-incoming/" + guid + "/purchase-order-draft"},
	} {
		action := findStockAction(t, "planned-incoming-"+entry.name)
		if action.HTTPMethod != entry.method || len(action.Args) != 1 || action.Args[0].Type != "uuid" || !action.Args[0].Required {
			t.Fatalf("incorrect public contract: %+v", action)
		}
		path, consumed := buildRESTPath(findStockDomain(t), *action, map[string]any{action.Args[0].Name: guid})
		if path != entry.path || len(consumed) != 1 {
			t.Fatalf("%s: route %q, consumed %v", entry.name, path, consumed)
		}
	}
}

func TestStockPlannedIncomingPreservesExactQuantityAndReviewedVersion(t *testing.T) {
	for _, name := range []string{"create", "update"} {
		action := findStockAction(t, "planned-incoming-"+name)
		found := false
		for _, flag := range action.Flags {
			if flag.Name == "quantity" {
				found = flag.Type == "decimal" && flag.BodyName == "quantity" && flag.Required
			}
		}
		if !found {
			t.Fatalf("%s must preserve decimal quantity", name)
		}
	}
	for _, name := range []string{"update", "convert"} {
		found := false
		for _, flag := range findStockAction(t, "planned-incoming-"+name).Flags {
			if flag.Name == "expected-version-guid" {
				found = flag.Type == "uuid" && flag.BodyName == "expectedVersionGuid" && flag.Required
			}
		}
		if !found {
			t.Fatalf("%s must require reviewed version", name)
		}
	}
}
