package registry

// Location public arguments use GUIDs; integer tool arguments are no longer supported.
func locationActions() []Action {
	domain := "Location"
	identity := []ArgDef{{Name: "locationGuid", Description: domain + " public GUID", Required: true, Type: "non-empty-uuid"}}
	fields := jsonFlag()
	fields.Required = true
	fields.Description = "Reviewed JSON; optional idempotencyKey and mutationOutcomeVersion=1 (plus exact expectedUpdatedAt for update) negotiate a minimal receipt. Retain original JSON unchanged after uncertain delivery"
	confirmation := FlagDef{Name: "confirm", Description: "Confirm the reviewed location change", Type: "bool", Required: true, MustBeTrue: true, LocalOnly: true}
	return append([]Action{
		{Name: "list", Description: "List visible records; --search matches part of the name, ignoring case", ToolName: "Uteamup" + domain + "List", Flags: append(paginationFlags(),
			FlagDef{Name: "search", Description: "Case-insensitive part of the location name; a hyphen also matches an en or em dash", Type: "string", QueryName: "nameFilter"})},
		{Name: "get", Description: "Get by public GUID", ToolName: "Uteamup" + domain + "GetByGuid", RESTPath: "{locationGuid}", Args: identity},
		{Name: "node-tree", Description: "Read the floors, rooms and areas inside a location; an optional type filter returns flat matching nodes with parent GUIDs", ToolName: "UteamupLocationNodeTree", HTTPMethod: "GET", RESTPath: "{locationGuid}/nodes", Args: identity,
			Flags: []FlagDef{{Name: "node-type-guid", Description: "Public node-type GUID from node-types", Type: "non-empty-uuid", QueryName: "locationNodeTypeGuid"}}},
		{Name: "create", Description: "Create from reviewed JSON", ToolName: "Uteamup" + domain + "Create", Flags: []FlagDef{fields, confirmation}},
		{Name: "update", Description: "Update the reviewed public GUID", ToolName: "Uteamup" + domain + "Update", RESTPath: "{locationGuid}", Args: identity, Flags: []FlagDef{fields, confirmation}},
		{Name: "delete", Description: "Delete the reviewed public GUID; domain association restrictions still apply", ToolName: "Uteamup" + domain + "Delete", RESTPath: "{locationGuid}", Args: identity, Flags: []FlagDef{confirmation,
			{Name: "idempotency-key", Description: "Original operation GUID; supply all three metadata flags together", Type: "non-empty-uuid", QueryName: "idempotencyKey"},
			{Name: "expected-updated-at", Description: "Exact canonical reviewed UTC revision; retain unchanged on retry", Type: "string", QueryName: "expectedUpdatedAt"},
			{Name: "mutation-outcome-version", Description: "Minimal outcome version (1)", Type: "int", QueryName: "mutationOutcomeVersion"},
		}},
	}, append(locationNodeActions(), locationNodeTypeActions()...)...)
}
