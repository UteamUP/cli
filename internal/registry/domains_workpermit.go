package registry

func init() {
	Register(&Domain{
		Name:        "workpermit",
		Aliases:     []string{"work-permit"},
		Description: "Work permits: biosecurity zone entry declarations and the current user's zone status",
		APIPath:     "/api/workpermit",
		Actions: []Action{
			{
				Name:        "biosecurity-status",
				Description: "Show whether a work order is in a biosecurity zone and whether you hold a valid signed entry declaration for it",
				ToolName:    "UteamupBiosecurityStatus",
				RESTPath:    "biosecurity-status",
				HTTPMethod:  "GET",
				Flags: []FlagDef{
					{Name: "workorder-guid", Description: "Work order GUID", Required: true, Type: "non-empty-uuid", QueryName: "workorderGuid"},
				},
			},
			{
				Name:        "biosecurity-entry",
				Description: "Record a biosecurity entry declaration (a Draft permit with the four disinfection lines) for a zone, or return your open draft",
				ToolName:    "UteamupBiosecurityEntry",
				RESTPath:    "biosecurity-entry",
				HTTPMethod:  "POST",
				Flags: []FlagDef{
					{Name: "location-guid", Description: "Biosecurity zone location GUID", Required: true, Type: "non-empty-uuid"},
					{Name: "workorder-guid", Description: "Optional work order GUID the entry is made for", Type: "uuid"},
				},
			},
		},
	})
}
