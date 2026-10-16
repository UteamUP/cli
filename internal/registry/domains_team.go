package registry

import (
	"fmt"
	"github.com/spf13/cobra"
	"regexp"
	"strings"
	"time"
)

func teamActions() []Action {
	identity := []ArgDef{{Name: "externalGuid", Description: "Team public GUID", Required: true, Type: "non-empty-uuid"}}
	fields := jsonFlag()
	fields.Description = "Team JSON; optional idempotencyKey and mutationOutcomeVersion=1 (plus exact expectedUpdatedAt on update) return a minimal receipt; retain original JSON unchanged on retry"
	return []Action{
		{Name: "list", Description: "List visible teams", ToolName: "UteamupTeamList", Flags: paginationFlags()},
		{Name: "get", Description: "Get a team by public GUID", ToolName: "UteamupTeamGet", RESTPath: "{externalGuid}", Args: identity},
		{Name: "create", Description: "Create a team; optional original metadata negotiates a minimal acknowledgment", ToolName: "UteamupTeamCreate", Flags: []FlagDef{fields}},
		{Name: "update", Description: "Update a team by public GUID; retain exact original metadata for retries", ToolName: "UteamupTeamUpdate", RESTPath: "{externalGuid}", Args: identity, Flags: []FlagDef{fields}},
		{Name: "delete", Description: "Delete a team by public GUID; optional original metadata negotiates a minimal acknowledgment", ToolName: "UteamupTeamDelete", RESTPath: "{externalGuid}", Args: identity, Flags: []FlagDef{
			{Name: "idempotency-key", Description: "Original nonempty operation GUID; supply all three mutation metadata flags together", Type: "non-empty-uuid", QueryName: "idempotencyKey"},
			{Name: "expected-updated-at", Description: "Exact canonical reviewed UTC revision (seconds or up to seven fractional digits, without trailing zeros); retain unchanged on retry", Type: "string", QueryName: "expectedUpdatedAt"},
			{Name: "mutation-outcome-version", Description: "Minimal outcome version (1); requires original operation and reviewed revision", Type: "int", QueryName: "mutationOutcomeVersion"},
		}},
	}
}

var teamRevisionPattern = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}(\.[0-9]{1,7})?Z$`)

func validateTeamDeleteMetadata(cmd *cobra.Command, action Action) error {
	if action.ToolName != "UteamupTeamDelete" {
		return nil
	}
	flags := []string{"idempotency-key", "expected-updated-at", "mutation-outcome-version"}
	supplied := 0
	for _, name := range flags {
		if cmd.Flags().Changed(name) {
			supplied++
		}
	}
	if supplied == 0 {
		return nil
	}
	if supplied != len(flags) {
		return fmt.Errorf("Team delete requires all three original mutation metadata flags together")
	}
	key, _ := cmd.Flags().GetString("idempotency-key")
	version, _ := cmd.Flags().GetInt("mutation-outcome-version")
	if !uuidValuePattern.MatchString(key) || strings.EqualFold(key, emptyUUIDValue) || version != 1 {
		return fmt.Errorf("Team delete requires a nonempty operation GUID and mutation outcome version 1")
	}
	revision, _ := cmd.Flags().GetString("expected-updated-at")
	const layout = "2006-01-02T15:04:05.9999999Z"
	parsed, err := time.Parse(layout, revision)
	if !teamRevisionPattern.MatchString(revision) || err != nil || parsed.Year() < 1 || parsed.Format(layout) != revision || parsed.IsZero() {
		return fmt.Errorf("Team delete requires the exact canonical original UTC revision, with at most seven fractional digits and no trailing zeros")
	}
	return nil
}
