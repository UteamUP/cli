package registry

import "testing"

func TestAssetEnergyUsesBoundedGuidReadRoutes(t *testing.T) {
	domain := findDomainByName(t, "asset-energy")
	if domain.APIPath != "/api/assets" {
		t.Fatalf("asset-energy API path = %q", domain.APIPath)
	}
	for _, name := range []string{"list", "summary"} {
		action := findDomainAction(t, "asset-energy", name)
		wantPath := "energy"
		if name == "summary" {
			wantPath += "/summary"
		}
		path, _ := buildRESTPath(domain, *action, nil)
		if action.MCPOnly || action.RESTPath != wantPath || path != "/api/assets/"+wantPath {
			t.Fatalf("%s is not bound to the existing bounded read: %+v, %s", name, action, path)
		}
		asset := actionFlagByName(t, action, "asset-guid")
		if asset.Type != "non-empty-uuid" || asset.BodyName != "assetGuid" || asset.Required {
			t.Fatalf("%s filter must be an optional public asset GUID: %+v", name, asset)
		}
		for _, legacy := range []string{"asset-id", "id", "tenant-id"} {
			if actionFlagExists(action, legacy) {
				t.Errorf("%s exposes legacy %s", name, legacy)
			}
		}
	}
	list := findDomainAction(t, "asset-energy", "list")
	if actionFlagByName(t, list, "page-size").Default != 100 {
		t.Error("energy list must request at most the agreed default 100 rows")
	}
	if len(actionFlagByName(t, list, "sort-by").AllowedValues) != 5 || len(actionFlagByName(t, list, "sort-order").AllowedValues) != 2 {
		t.Error("server sorting fields must remain constrained to the read contract")
	}
}

func TestAssetEnergyCommandsRequirePublicGuidsAndReviewedModels(t *testing.T) {
	for _, name := range []string{"get", "create", "update", "delete"} {
		action := findDomainAction(t, "asset-energy", name)
		if !action.MCPOnly || len(action.Args) != 1 || !action.RejectExtraArgs {
			t.Fatalf("%s must use its governed GUID tool: %+v", name, action)
		}
		arg := action.Args[0]
		wantName, wantBody := "record-guid", "recordGuid"
		if name == "create" {
			wantName, wantBody = "asset-guid", "assetGuid"
		}
		if arg.Name != wantName || arg.BodyName != wantBody || !arg.Required || arg.Type != "non-empty-uuid" {
			t.Errorf("%s identifier is not a public GUID: %+v", name, arg)
		}
		if name == "create" || name == "update" {
			model := actionFlagByName(t, action, "request-file")
			if !model.JSONFile || !model.Required || model.BodyName != "model" {
				t.Errorf("%s model must preserve structured reviewed JSON: %+v", name, model)
			}
		}
	}
}
