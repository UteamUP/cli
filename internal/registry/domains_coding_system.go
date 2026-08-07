package registry

func init() {
	Register(&Domain{
		Name:        "codingsystem",
		Aliases:     []string{"cs", "coding"},
		Description: "Manage industrial coding systems (KKS, RDS, RDS-PP, RDS-PS, ISO 14224)",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List available coding systems for the tenant",
				ToolName:    "UteamupCodingsystemList",
				MCPOnly:     true,
			},
			{
				Name:        "tree",
				Description: "Browse code catalog tree hierarchy",
				ToolName:    "UteamupCodingsystemTree",
				MCPOnly:     true,
				Flags: []FlagDef{
					{Name: "coding-system-guid", Short: "c", BodyName: "codingSystemGuid", Description: "Coding system GUID", Required: true, Type: "uuid"},
					{Name: "parent-guid", Short: "p", BodyName: "parentGuid", Description: "Parent entry GUID (omit for root level)", Type: "uuid"},
				},
			},
			{
				Name:        "search",
				Description: "Search assets by coding system code prefix",
				ToolName:    "UteamupCodingsystemSearchAssets",
				MCPOnly:     true,
				Args:        []ArgDef{{Name: "query", BodyName: "codePrefix", Description: "Code prefix to search (e.g., '1-HLA')", Required: true, Type: "string"}},
			},
			{
				Name:        "next-code",
				Description: "Get next available code for a parent entry",
				ToolName:    "UteamupCodingsystemNextCode",
				MCPOnly:     true,
				Flags: []FlagDef{
					{Name: "parent-guid", Short: "p", BodyName: "parentEntryGuid", Description: "Parent entry GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "assign",
				Description: "Assign a code catalog entry to an asset",
				ToolName:    "UteamupCodingsystemAssignCode",
				MCPOnly:     true,
				Flags: []FlagDef{
					{Name: "asset-guid", Short: "a", BodyName: "assetGuid", Description: "Asset GUID", Required: true, Type: "uuid"},
					{Name: "entry-guid", Short: "e", BodyName: "codeCatalogEntryGuid", Description: "Code catalog entry GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "workorders",
				Description: "Get work orders by code branch prefix",
				ToolName:    "UteamupCodingsystemWorkorders",
				MCPOnly:     true,
				Args:        []ArgDef{{Name: "prefix", BodyName: "codeBranchPrefix", Description: "Code branch prefix (e.g., '1-HLA')", Required: true, Type: "string"}},
			},
			{
				Name:        "create-workorder",
				Description: "Create a work order from a code catalog entry",
				ToolName:    "UteamupCodingsystemCreateWorkorder",
				MCPOnly:     true,
				Flags: []FlagDef{
					{Name: "entry-guid", Short: "e", BodyName: "codeCatalogEntryGuid", Description: "Code catalog entry GUID", Required: true, Type: "uuid"},
					{Name: "title", BodyName: "name", Description: "Work order title", Required: true, Type: "string"},
					{Name: "description", Description: "Work order description", Type: "string"},
					{Name: "priority", Description: "Priority (1=Low, 2=Medium, 3=High, 4=Critical)", Default: 2, Type: "int"},
				},
			},
		},
	})
}
