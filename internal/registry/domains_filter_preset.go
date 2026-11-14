package registry

// Personal saved filters on one hub list (workorders, assets, asset-types, ...). They act only for the signed-in
// user in the current company, so no action takes a user or tenant argument. The hub key is lowercase with
// hyphens; the backend refuses any other shape.
func init() {
	hubKey := ArgDef{Name: "hubKey", Description: "Hub identifier, e.g. workorders, assets, asset-types", Required: true, Type: "string"}
	presetGuid := ArgDef{Name: "presetGuid", Description: "GUID of the saved filter", Required: true, Type: "non-empty-uuid"}
	saveFlags := []FlagDef{
		{Name: "name", Description: "Name shown on the saved-filter button, 1 to 100 characters", Type: "string", Required: true},
		{Name: "filter-file", BodyName: "filter", Description: "JSON file holding the filter object (at most 8 KB)", Type: "string", Required: true, JSONFile: true},
		{Name: "default", BodyName: "isDefault", Description: "Open the hub with this filter", Type: "bool", Default: false},
	}

	Register(&Domain{
		Name:        "filter-preset",
		Description: "Manage your personal saved filters on hub lists",
		APIPath:     "/api/user/filter-presets",
		Actions: []Action{
			{
				Name:            "list",
				Description:     "List your saved filters on one hub, by name",
				ToolName:        "UteamupFilterPresetList",
				HTTPMethod:      "GET",
				RESTPath:        "{hubKey}",
				RejectExtraArgs: true,
				Args:            []ArgDef{hubKey},
			},
			{
				Name:            "create",
				Description:     "Save a new filter on a hub. Names are unique per hub; at most 20 filters per hub",
				ToolName:        "UteamupFilterPresetCreate",
				HTTPMethod:      "POST",
				RESTPath:        "{hubKey}",
				RejectExtraArgs: true,
				Args:            []ArgDef{hubKey},
				Flags:           saveFlags,
			},
			{
				Name:            "update",
				Description:     "Rename a saved filter and replace its filter object",
				ToolName:        "UteamupFilterPresetUpdate",
				HTTPMethod:      "PUT",
				RESTPath:        "{hubKey}/{presetGuid}",
				RejectExtraArgs: true,
				Args:            []ArgDef{hubKey, presetGuid},
				Flags:           saveFlags,
			},
			{
				Name:            "set-default",
				Description:     "Open the hub with this saved filter; the previous default is cleared",
				ToolName:        "UteamupFilterPresetSetDefault",
				HTTPMethod:      "PUT",
				RESTPath:        "{hubKey}/{presetGuid}/set-default",
				RejectExtraArgs: true,
				Args:            []ArgDef{hubKey, presetGuid},
			},
			{
				Name:            "remove",
				Description:     "Remove one of your saved filters (a personal view setting, so it is removed, not archived)",
				ToolName:        "UteamupFilterPresetRemove",
				HTTPMethod:      "DELETE",
				RESTPath:        "{hubKey}/{presetGuid}",
				RejectExtraArgs: true,
				Args:            []ArgDef{hubKey, presetGuid},
			},
		},
	})
}
