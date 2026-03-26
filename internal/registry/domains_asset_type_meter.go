package registry

func init() {
	Register(&Domain{
		Name:        "asset-type-meter",
		Description: "Manage asset type meter definitions",
		Actions: []Action{
			{
				Name:        "toggle",
				Description: "Toggle IsMetered on an asset type",
				ToolName:    "UteamupAssetTypeMeterToggle",
				Flags: []FlagDef{
					{Name: "asset-type-id", Description: "Asset type ID", Required: true, Type: "int"},
					{Name: "metered", Description: "Whether the asset type is metered (true/false)", Required: true, Type: "string"},
				},
			},
			{
				Name:        "list",
				Description: "List meter definitions for an asset type",
				ToolName:    "UteamupAssetTypeMeterListDefinitions",
				Flags: []FlagDef{
					{Name: "asset-type-id", Description: "Asset type ID", Required: true, Type: "int"},
				},
			},
			{
				Name:        "add",
				Description: "Add a meter definition to an asset type",
				ToolName:    "UteamupAssetTypeMeterAddDefinition",
				Flags: []FlagDef{
					{Name: "asset-type-id", Description: "Asset type ID", Required: true, Type: "int"},
					{Name: "name", Description: "Meter definition name", Required: true, Type: "string"},
					{Name: "unit", Description: "Unit of measurement (e.g. km, °C, PSI)", Required: true, Type: "string"},
					{Name: "min", Description: "Alert threshold minimum value", Type: "int"},
					{Name: "max", Description: "Alert threshold maximum value", Type: "int"},
					{Name: "interval", Description: "Expected reading interval in seconds", Type: "int"},
				},
			},
			{
				Name:        "remove",
				Description: "Remove a meter definition from an asset type",
				ToolName:    "UteamupAssetTypeMeterRemoveDefinition",
				Flags: []FlagDef{
					{Name: "asset-type-id", Description: "Asset type ID", Required: true, Type: "int"},
					{Name: "definition-id", Description: "Attribute definition ID to remove", Required: true, Type: "int"},
				},
			},
		},
	})
}
