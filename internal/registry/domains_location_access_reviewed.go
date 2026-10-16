package registry

import (
	"encoding/json"
	"fmt"
	"time"
)

func locationAccessReviewedActions() []Action {
	fields := jsonFlag()
	fields.Required = true
	fields.Description = "Original reviewed schedule JSON, operation GUID, receipt version 1 and exact revision for update/delete"
	flags := []FlagDef{fields, {Name: "confirm", Description: "Confirm the reviewed schedule change", Type: "bool", Required: true, MustBeTrue: true, LocalOnly: true}}
	location := ArgDef{Name: "locationGuid", Description: "Location public GUID", Type: "non-empty-uuid", Required: true}
	code := ArgDef{Name: "scheduleGuid", Description: "Schedule public GUID", Type: "non-empty-uuid", Required: true}
	return []Action{
		{Name: "create-reviewed", ToolName: "UteamupLocationAccessScheduleCreateReviewed", RESTBasePath: "/api/location", HTTPMethod: "POST", RESTPath: "{locationGuid}/access-schedules/reviewed", Args: []ArgDef{location}, Flags: flags, Description: "Create an independent Location access schedule"},
		{Name: "update-reviewed", ToolName: "UteamupLocationAccessScheduleUpdateReviewed", RESTBasePath: "/api/location", HTTPMethod: "PUT", RESTPath: "access-schedules/{scheduleGuid}/reviewed", Args: []ArgDef{code}, Flags: flags, Description: "Update scalar schedule fields at the exact revision"},
		{Name: "delete-reviewed", ToolName: "UteamupLocationAccessScheduleDeleteReviewed", RESTBasePath: "/api/location", HTTPMethod: "DELETE", RESTPath: "access-schedules/{scheduleGuid}/reviewed", Args: []ArgDef{code}, Flags: flags, Description: "Delete the exact Location access schedule"},
	}
}

func validateLocationAccessReviewedOriginal(action Action, body map[string]any) error {
	switch action.ToolName {
	case "UteamupLocationAccessScheduleCreateReviewed", "UteamupLocationAccessScheduleUpdateReviewed", "UteamupLocationAccessScheduleDeleteReviewed":
	default:
		return nil
	}
	key, ok := body["idempotencyKey"].(string)
	if !ok || !uuidValuePattern.MatchString(key) || key == emptyUUIDValue || body["mutationOutcomeVersion"] != json.Number("1") {
		return fmt.Errorf("complete original schedule operation metadata is required")
	}
	revision := body["expectedUpdatedAtUtc"]
	if action.ToolName == "UteamupLocationAccessScheduleCreateReviewed" {
		if revision != nil {
			return fmt.Errorf("create cannot carry a prior revision")
		}
	} else {
		text, ok := revision.(string)
		const layout = "2006-01-02T15:04:05.9999999Z"
		parsed, err := time.Parse(layout, text)
		if !ok || !teamRevisionPattern.MatchString(text) || err != nil || parsed.IsZero() || parsed.Format(layout) != text {
			return fmt.Errorf("exact canonical UTC schedule revision is required")
		}
	}
	return nil
}
