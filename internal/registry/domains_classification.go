package registry

// Classification writes keep the existing normal REST domain routes.
func classificationActions(domain string) []Action {
	identity := []ArgDef{{Name: "guid", Description: domain + " public GUID", Required: true, Type: "non-empty-uuid"}}
	fields := jsonFlag()
	fields.Required = true
	fields.Description = "Reviewed JSON; optional idempotencyKey and mutationOutcomeVersion=1 (plus exact expectedUpdatedAt for update) negotiate a minimal receipt. Retain original JSON unchanged after uncertain delivery"
	confirmation := FlagDef{Name: "confirm", Description: "Confirm the reviewed classification change", Type: "bool", Required: true, MustBeTrue: true, LocalOnly: true}
	return []Action{
		{Name: "workorders", Description: "List visible associated workorders by classification GUID", ToolName: "Uteamup" + domain + "Workorders", HTTPMethod: "GET", RESTPath: "by-guid/{guid}/workorders", Args: identity, Flags: paginationFlags()},
		{Name: "list", Description: "List visible records", ToolName: "Uteamup" + domain + "List", Flags: paginationFlags()},
		{Name: "get", Description: "Get by public GUID", ToolName: "Uteamup" + domain + "Get", RESTPath: "by-guid/{guid}", Args: identity},
		{Name: "create", Description: "Create from reviewed JSON", ToolName: "Uteamup" + domain + "Create", Flags: []FlagDef{fields, confirmation}},
		{Name: "update", Description: "Update the reviewed public GUID", ToolName: "Uteamup" + domain + "Update", RESTPath: "by-guid/{guid}", Args: identity, Flags: []FlagDef{fields, confirmation}},
		{Name: "delete", Description: "Delete the reviewed public GUID; domain association restrictions still apply", ToolName: "Uteamup" + domain + "Delete", RESTPath: "by-guid/{guid}", Args: identity, Flags: []FlagDef{confirmation,
			{Name: "idempotency-key", Description: "Original operation GUID; supply all three metadata flags together", Type: "non-empty-uuid", QueryName: "idempotencyKey"},
			{Name: "expected-updated-at", Description: "Exact canonical reviewed UTC revision; retain unchanged on retry", Type: "string", QueryName: "expectedUpdatedAt"},
			{Name: "mutation-outcome-version", Description: "Minimal outcome version (1)", Type: "int", QueryName: "mutationOutcomeVersion"},
		}},
	}
}

// categoryActions adds archive and restore: a category in use is archived instead of deleted. Archived
// categories leave lists, searches, pickers and synced phones; list shows them only with --archived.
func categoryActions() []Action {
	actions := classificationActions("Category")
	for index := range actions {
		if actions[index].Name == "list" {
			actions[index].Description += "; archived categories are hidden unless you pass --archived"
			actions[index].Flags = append(actions[index].Flags, FlagDef{
				Name: "archived", QueryName: "archived", Type: "bool",
				Description: "true lists only archived categories; omit it for categories in use",
			})
		}
	}
	identity := []ArgDef{{Name: "guid", Description: "Category public GUID", Required: true, Type: "non-empty-uuid"}}
	return append(actions,
		Action{
			Name: "archive",
			Description: "Archive a category: it leaves lists, searches, pickers and synced phones and stays on every " +
				"record that uses it. Use it for a category in use, which cannot be deleted. Needs Category.Delete",
			ToolName:        "UteamupCategoryArchive",
			HTTPMethod:      "POST",
			RESTPath:        "by-guid/{guid}/archive",
			RejectExtraArgs: true,
			Args:            identity,
		},
		Action{
			Name:            "restore",
			Description:     "Restore an archived category so it shows in lists, searches and pickers again. Needs Category.Delete",
			ToolName:        "UteamupCategoryRestore",
			HTTPMethod:      "POST",
			RESTPath:        "by-guid/{guid}/restore",
			RejectExtraArgs: true,
			Args:            identity,
		},
	)
}
