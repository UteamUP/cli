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
				MCPOnly:     true,
				Flags: []FlagDef{
					{Name: "asset-type-guid", BodyName: "assetTypeGuid", Description: "Asset type GUID", Required: true, Type: "uuid"},
					{Name: "metered", BodyName: "isMetered", Description: "Whether the asset type is metered", Required: true, Type: "bool"},
				},
			},
			{
				Name:        "list",
				Description: "List meter definitions for an asset type",
				ToolName:    "UteamupAssetTypeMeterListDefinitions",
				MCPOnly:     true,
				Flags: []FlagDef{
					{Name: "asset-type-guid", BodyName: "assetTypeGuid", Description: "Asset type GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "add",
				Description: "Add a meter definition to an asset type",
				ToolName:    "UteamupAssetTypeMeterAddDefinition",
				MCPOnly:     true,
				Flags: []FlagDef{
					{Name: "asset-type-guid", BodyName: "assetTypeGuid", Description: "Asset type GUID", Required: true, Type: "uuid"},
					{Name: "name", Description: "Meter definition name", Required: true, Type: "string"},
					{Name: "unit", Description: "Unit of measurement (e.g. km, °C, PSI)", Required: true, Type: "string"},
					{Name: "min", BodyName: "alertThresholdMin", Description: "Alert threshold minimum value", Type: "float"},
					{Name: "max", BodyName: "alertThresholdMax", Description: "Alert threshold maximum value", Type: "float"},
					{Name: "interval", BodyName: "expectedReadingIntervalSeconds", Description: "Expected reading interval in seconds", Type: "int"},
				},
			},
			{
				Name:        "remove",
				Description: "Remove a meter definition from an asset type",
				ToolName:    "UteamupAssetTypeMeterRemoveDefinition",
				MCPOnly:     true,
				Flags: []FlagDef{
					{Name: "asset-type-guid", BodyName: "assetTypeGuid", Description: "Asset type GUID", Required: true, Type: "uuid"},
					{Name: "definition-guid", BodyName: "definitionGuid", Description: "Attribute definition GUID", Required: true, Type: "uuid"},
				},
			},
		},
	})
}
