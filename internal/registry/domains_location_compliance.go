package registry

import (
	"encoding/json"
	"fmt"
	"time"
)

func locationComplianceActions() []Action {
	fields := jsonFlag()
	fields.Required = true
	fields.Description = "Original reviewed code JSON, operation GUID, receipt version 1 and exact revision for update/delete"
	flags := []FlagDef{fields, {Name: "confirm", Description: "Confirm the reviewed code change", Type: "bool", Required: true, MustBeTrue: true, LocalOnly: true}}
	location := ArgDef{Name: "locationGuid", Description: "Location public GUID", Type: "non-empty-uuid", Required: true}
	code := ArgDef{Name: "codeGuid", Description: "Code public GUID", Type: "non-empty-uuid", Required: true}
	return []Action{
		{Name: "location-create", ToolName: "UteamupLocationComplianceCodeCreate", RESTBasePath: "/api/compliance", HTTPMethod: "POST", RESTPath: "entity/location/by-guid/{locationGuid}/codes/reviewed", Args: []ArgDef{location}, Flags: flags, Description: "Create an independent Location-owned code"},
		{Name: "location-update", ToolName: "UteamupLocationComplianceCodeUpdate", RESTBasePath: "/api/compliance", HTTPMethod: "PUT", RESTPath: "entity/location/by-guid/{locationGuid}/codes/by-guid/{codeGuid}/reviewed", Args: []ArgDef{location, code}, Flags: flags, Description: "Update scalar code fields at the exact revision"},
		{Name: "location-delete", ToolName: "UteamupLocationComplianceCodeDelete", RESTBasePath: "/api/compliance", HTTPMethod: "DELETE", RESTPath: "entity/location/by-guid/{locationGuid}/codes/by-guid/{codeGuid}/reviewed", Args: []ArgDef{location, code}, Flags: flags, Description: "Delete the exact Location-owned code"},
	}
}

func validateLocationComplianceOriginal(action Action, body map[string]any) error {
	switch action.ToolName {
	case "UteamupLocationComplianceCodeCreate", "UteamupLocationComplianceCodeUpdate", "UteamupLocationComplianceCodeDelete":
	default:
		return nil
	}
	key, ok := body["idempotencyKey"].(string)
	if !ok || !uuidValuePattern.MatchString(key) || key == emptyUUIDValue || body["mutationOutcomeVersion"] != json.Number("1") {
		return fmt.Errorf("complete original compliance operation metadata is required")
	}
	revision := body["expectedUpdatedAt"]
	if action.ToolName == "UteamupLocationComplianceCodeCreate" {
		if revision != nil {
			return fmt.Errorf("create cannot carry a prior revision")
		}
	} else {
		text, ok := revision.(string)
		const layout = "2006-01-02T15:04:05.9999999Z"
		parsed, err := time.Parse(layout, text)
		if !ok || !teamRevisionPattern.MatchString(text) || err != nil || parsed.IsZero() || parsed.Format(layout) != text {
			return fmt.Errorf("exact canonical UTC code revision is required")
		}
	}
	for _, name := range []string{"assetId", "partId", "toolId", "chemicalId", "locationId", "customerId", "vendorId", "documentGuids"} {
		if body[name] != nil {
			return fmt.Errorf("%s is not part of reviewed Location code mutations", name)
		}
	}
	return nil
}
