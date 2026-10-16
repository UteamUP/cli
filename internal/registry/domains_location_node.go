package registry

import (
	"encoding/json"
	"fmt"
	"time"
)

func locationNodeActions() []Action {
	node := []ArgDef{{Name: "nodeGuid", Description: "Reviewed public node GUID", Required: true, Type: "non-empty-uuid"}}
	location := []ArgDef{{Name: "locationGuid", Description: "Owning location public GUID", Required: true, Type: "non-empty-uuid"}}
	fields := jsonFlag()
	fields.Required = true
	fields.Description = "Reviewed node JSON; retain original operation metadata and exact revision for retries"
	confirm := FlagDef{Name: "confirm", Description: "Confirm the reviewed node change and unlink consequences", Type: "bool", Required: true, MustBeTrue: true, LocalOnly: true}
	return []Action{
		{Name: "node-create", ToolName: "UteamupLocationNodeCreate", Description: "Create a node under a location", HTTPMethod: "POST", RESTPath: "{locationGuid}/nodes", Args: location, Flags: []FlagDef{fields, confirm}},
		{Name: "node-update", ToolName: "UteamupLocationNodeUpdate", Description: "Update a node; omitted type and attributes are preserved", HTTPMethod: "PUT", RESTPath: "node/{nodeGuid}", Args: node, Flags: []FlagDef{fields, confirm}},
		{Name: "node-move", ToolName: "UteamupLocationNodeMove", Description: "Move a node; null parent selects location root", HTTPMethod: "POST", RESTPath: "node/{nodeGuid}/move", Args: node, Flags: []FlagDef{fields, confirm}},
		{Name: "node-delete", ToolName: "UteamupLocationNodeDelete", Description: "Delete a childless node without floorplan; assets and stock are unpinned", HTTPMethod: "DELETE", RESTPath: "node/{nodeGuid}", Args: node, Flags: []FlagDef{confirm,
			{Name: "idempotency-key", Description: "Original operation GUID", Type: "non-empty-uuid", QueryName: "idempotencyKey", Required: true},
			{Name: "expected-updated-at", Description: "Exact reviewed UTC revision", Type: "string", QueryName: "expectedUpdatedAt", Required: true},
			{Name: "mutation-outcome-version", Description: "Minimal outcome version 1", Type: "int", QueryName: "mutationOutcomeVersion", Required: true},
		}},
	}
}

func validateLocationNodeOriginal(action Action, body map[string]any) error {
	switch action.ToolName {
	case "UteamupLocationNodeCreate", "UteamupLocationNodeUpdate", "UteamupLocationNodeMove":
	default:
		return nil
	}
	key, ok := body["idempotencyKey"].(string)
	if !ok || !uuidValuePattern.MatchString(key) || key == emptyUUIDValue || body["mutationOutcomeVersion"] != json.Number("1") {
		return fmt.Errorf("node mutations require original operation GUID and mutationOutcomeVersion 1")
	}
	revision, exists := body["expectedUpdatedAt"]
	if action.ToolName == "UteamupLocationNodeCreate" {
		if exists && revision != nil {
			return fmt.Errorf("node create cannot carry a reviewed revision")
		}
		return nil
	}
	text, ok := revision.(string)
	const layout = "2006-01-02T15:04:05.9999999Z"
	parsed, err := time.Parse(layout, text)
	if !ok || !teamRevisionPattern.MatchString(text) || err != nil || parsed.IsZero() || parsed.Format(layout) != text {
		return fmt.Errorf("node update or move requires the exact canonical reviewed UTC revision")
	}
	return nil
}
