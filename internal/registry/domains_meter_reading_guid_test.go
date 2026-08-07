package registry

import "testing"

func TestMeterReadingAttributeUpdateUsesGuidRequestFile(t *testing.T) {
	action := findDomainAction(t, "meter-reading", "update-attributes")
	if !action.MCPOnly {
		t.Error("update-attributes must use its governed MCP tool")
	}

	if len(action.Args) != 1 {
		t.Fatalf("update-attributes args = %d, want 1", len(action.Args))
	}
	asset := action.Args[0]
	if asset.Name != "asset-guid" || asset.BodyName != "assetGuid" || !asset.Required || asset.Type != "uuid" {
		t.Errorf("asset-guid is miswired: %+v", asset)
	}

	request := actionFlagByName(t, action, "request-file")
	if !request.Required || !request.JSONFile || request.BodyName != "request" {
		t.Errorf("request-file must bind reviewed JSON to the MCP request: %+v", request)
	}
	if actionFlagExists(action, "values-json") {
		t.Error("update-attributes must not expose the legacy integer-oriented values-json flag")
	}
}

func actionFlagExists(action *Action, name string) bool {
	for _, flag := range action.Flags {
		if flag.Name == name {
			return true
		}
	}
	return false
}
