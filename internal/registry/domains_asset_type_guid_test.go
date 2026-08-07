package registry

import "testing"

func TestAssetTypeActionsMirrorGuidMCPContracts(t *testing.T) {
	domain := findDomainByName(t, "asset-type")
	for _, action := range domain.Actions {
		if !action.MCPOnly {
			t.Errorf("%s must use its governed MCP tool", action.Name)
		}
		for _, arg := range action.Args {
			if arg.Name == "id" || arg.Type == "int" {
				t.Errorf("%s exposes forbidden integer identity arg %+v", action.Name, arg)
			}
		}
	}

	for _, name := range []string{"get", "update", "delete", "compatible-parts"} {
		action := findDomainAction(t, "asset-type", name)
		if len(action.Args) != 1 {
			t.Fatalf("%s args = %d, want 1", name, len(action.Args))
		}
		arg := action.Args[0]
		if arg.Name != "assetTypeGuid" || !arg.Required || arg.Type != "uuid" {
			t.Errorf("%s assetTypeGuid is miswired: %+v", name, arg)
		}
	}

	for _, name := range []string{"create", "update"} {
		action := findDomainAction(t, "asset-type", name)
		request := actionFlagByName(t, action, "request-file")
		if !request.Required || !request.JSONFile || request.BodyName != "model" {
			t.Errorf("%s request-file is miswired: %+v", name, request)
		}
	}
}

func TestAssetTypeMeterActionsAreGuidOnly(t *testing.T) {
	domain := findDomainByName(t, "asset-type-meter")
	for _, action := range domain.Actions {
		if !action.MCPOnly {
			t.Errorf("%s must use its governed MCP tool", action.Name)
		}
		for _, flag := range action.Flags {
			if flag.Name == "asset-type-id" || flag.Name == "definition-id" {
				t.Errorf("%s exposes forbidden integer identity flag %q", action.Name, flag.Name)
			}
		}

		assetType := actionFlagByName(t, &action, "asset-type-guid")
		if !assetType.Required || assetType.Type != "uuid" || assetType.BodyName != "assetTypeGuid" {
			t.Errorf("%s asset-type-guid is miswired: %+v", action.Name, assetType)
		}
	}

	remove := findDomainAction(t, "asset-type-meter", "remove")
	definition := actionFlagByName(t, remove, "definition-guid")
	if !definition.Required || definition.Type != "uuid" || definition.BodyName != "definitionGuid" {
		t.Errorf("remove definition-guid is miswired: %+v", definition)
	}

	toggle := findDomainAction(t, "asset-type-meter", "toggle")
	metered := actionFlagByName(t, toggle, "metered")
	if metered.Type != "bool" || metered.BodyName != "isMetered" {
		t.Errorf("toggle metered flag is miswired: %+v", metered)
	}
}

func TestAssetTypeReferencesOnAssetActionsAreGuidOnly(t *testing.T) {
	create := findDomainAction(t, "asset", "create")
	if !create.MCPOnly {
		t.Error("asset create must use the GUID-only MCP model contract")
	}
	request := actionFlagByName(t, create, "request-file")
	if !request.Required || !request.JSONFile || request.BodyName != "model" {
		t.Errorf("asset create request-file is miswired: %+v", request)
	}

	update := findDomainAction(t, "asset", "update")
	for _, flag := range update.Flags {
		switch flag.Name {
		case "asset-type-id", "asset-type-ids", "primary-asset-type-id", "from-json":
			t.Errorf("asset update exposes stale identity/payload flag %q", flag.Name)
		}
	}
	assetTypes := actionFlagByName(t, update, "asset-type-guids")
	if assetTypes.Type != "stringSlice" {
		t.Errorf("update asset-type-guids is miswired: %+v", assetTypes)
	}
	primary := actionFlagByName(t, update, "primary-asset-type-guid")
	if primary.Type != "uuid" {
		t.Errorf("update primary-asset-type-guid is miswired: %+v", primary)
	}

	if update.RESTPath != "by-guid/{assetGuid}" || update.HTTPMethod != "PUT" {
		t.Errorf("asset update GUID route is miswired: %+v", update)
	}
	if len(update.Args) != 1 || update.Args[0].Name != "assetGuid" || update.Args[0].Type != "uuid" {
		t.Errorf("asset update GUID argument is miswired: %+v", update.Args)
	}

	specs := findDomainAction(t, "asset", "get-specs")
	if specs.RESTPath != "by-guid/{assetGuid}/effective-attribute-definitions" {
		t.Errorf("asset get-specs GUID route is miswired: %q", specs.RESTPath)
	}
	if len(specs.Args) != 1 || specs.Args[0].Name != "assetGuid" || specs.Args[0].Type != "uuid" {
		t.Errorf("asset get-specs GUID argument is miswired: %+v", specs.Args)
	}
}
