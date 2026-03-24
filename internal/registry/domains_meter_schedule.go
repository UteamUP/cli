package registry

func init() {
	Register(&Domain{
		Name:        "meter-schedule",
		Description: "Manage meter reading schedules",
		Actions: []Action{
			{
				Name:        "create",
				Description: "Create a meter reading schedule",
				ToolName:    "UteamupMeterScheduleCreate",
				Flags: []FlagDef{
					{Name: "asset-id", Description: "Asset ID", Required: true, Type: "int"},
					{Name: "attribute-definition-id", Description: "Attribute definition ID", Required: true, Type: "int"},
					{Name: "interval-seconds", Description: "Reading interval in seconds", Required: true, Type: "int"},
					{Name: "label", Description: "Schedule label", Type: "string"},
					{Name: "preferred-time", Description: "Preferred time of day (HH:mm)", Type: "string"},
					{Name: "timezone", Description: "Timezone (e.g. UTC, Europe/London)", Type: "string"},
				},
			},
			{
				Name:        "list",
				Description: "List meter reading schedules",
				ToolName:    "UteamupMeterScheduleList",
				Flags: append(paginationFlags(),
					FlagDef{Name: "asset-id", Description: "Filter by asset ID", Type: "int"},
				),
			},
			{
				Name:        "overdue",
				Description: "List overdue meter reading schedules",
				ToolName:    "UteamupMeterScheduleOverdue",
				Flags:       paginationFlags(),
			},
			{
				Name:        "compliance",
				Description: "Get meter reading compliance summary",
				ToolName:    "UteamupMeterScheduleCompliance",
				Flags: []FlagDef{
					{Name: "asset-id", Description: "Filter by asset ID", Type: "int"},
				},
			},
			{
				Name:        "initialize",
				Description: "Initialize meter schedules from asset type defaults",
				ToolName:    "UteamupMeterScheduleInitialize",
				Flags: []FlagDef{
					{Name: "asset-id", Description: "Asset ID", Required: true, Type: "int"},
				},
			},
		},
	})
}
