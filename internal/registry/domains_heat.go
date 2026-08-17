package registry

func init() {
	Register(&Domain{
		Name:        "heat",
		Aliases:     []string{"heatexposure", "heatnep"},
		Description: "Heat NEP exposure readings. Employer program evidence; not a medical device.",
		APIPath:     "/api/heatexposurereading",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List Heat NEP exposure readings",
				ToolName:    "UteamupHeatexposureList",
				RESTPath:    "",
				HTTPMethod:  "GET",
				Flags: []FlagDef{
					{Name: "skip", Description: "Skip offset", Default: 0, Type: "int", QueryName: "skip"},
					{Name: "take", Description: "Page size (server caps at 200)", Default: 50, Type: "int", QueryName: "take"},
					{Name: "location-guid", Description: "Optional location GUID", Type: "uuid", QueryName: "locationGuid"},
					{Name: "workorder-guid", Description: "Optional work order GUID", Type: "uuid", QueryName: "workorderGuid"},
				},
			},
			{
				Name:        "create",
				Description: "Create a Heat NEP reading. Heat index is computed; WBGT is stored only when supplied.",
				ToolName:    "UteamupHeatexposureCreate",
				RESTPath:    "",
				HTTPMethod:  "POST",
				Flags: []FlagDef{
					{Name: "from-json", BodyName: "model", Description: "JSON file containing HeatExposureReadingCreateModel", Required: true, Type: "string", JSONFile: true},
				},
			},
		},
	})
}
