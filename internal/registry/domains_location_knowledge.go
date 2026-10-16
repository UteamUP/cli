package registry

import (
	"encoding/json"
	"fmt"
)

func locationKnowledgeActions() []Action {
	args := []ArgDef{{Name: "locationGuid", Description: "Location public GUID", Type: "non-empty-uuid", Required: true},
		{Name: "articleGuid", Description: "Article public GUID", Type: "non-empty-uuid", Required: true}}
	fields := jsonFlag()
	fields.Required = true
	fields.Description = "Original reviewed JSON with idempotencyKey, mutationOutcomeVersion=1 and explicit expectedLinkGuid"
	flags := []FlagDef{fields, {Name: "confirm", Description: "Confirm this reviewed relation change", Type: "bool", Required: true, MustBeTrue: true, LocalOnly: true}}
	return []Action{
		{Name: "location-link", ToolName: "UteamupLocationKnowledgeLink", HTTPMethod: "POST", RESTPath: "linked/location/{locationGuid}/{articleGuid}/reviewed", Args: args, Flags: flags, Description: "Link an absent reviewed Location relation; no legacy fallback"},
		{Name: "location-unlink", ToolName: "UteamupLocationKnowledgeUnlink", HTTPMethod: "DELETE", RESTPath: "linked/location/{locationGuid}/{articleGuid}/reviewed", Args: args, Flags: flags, Description: "Unlink the exact reviewed Location junction; no legacy fallback"},
	}
}

func validateLocationKnowledgeOriginal(action Action, body map[string]any) error {
	if action.ToolName != "UteamupLocationKnowledgeLink" && action.ToolName != "UteamupLocationKnowledgeUnlink" {
		return nil
	}
	key, ok := body["idempotencyKey"].(string)
	if !ok || !uuidValuePattern.MatchString(key) || key == emptyUUIDValue || body["mutationOutcomeVersion"] != json.Number("1") {
		return fmt.Errorf("reviewed relation requires original operation GUID and mutationOutcomeVersion 1")
	}
	expected, exists := body["expectedLinkGuid"]
	if !exists {
		return fmt.Errorf("expectedLinkGuid must be explicitly provided")
	}
	if action.ToolName == "UteamupLocationKnowledgeLink" {
		if expected != nil {
			return fmt.Errorf("link requires expectedLinkGuid null")
		}
		return nil
	}
	guid, ok := expected.(string)
	if !ok || !uuidValuePattern.MatchString(guid) || guid == emptyUUIDValue {
		return fmt.Errorf("unlink requires exact reviewed junction GUID")
	}
	return nil
}
