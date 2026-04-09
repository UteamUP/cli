package registry

func init() {
	Register(&Domain{
		Name:        "codingsystem",
		Aliases:     []string{"cs", "coding"},
		Description: "Manage industrial coding systems (KKS, RDS, ISO 14224)",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List available coding systems for the tenant",
				ToolName:    "UteamupCodingsystemList",
			},
			{
				Name:        "tree",
				Description: "Browse code catalog tree hierarchy",
				ToolName:    "UteamupCodingsystemTree",
				Flags: []FlagDef{
					{Name: "coding-system-id", Short: "c", Description: "Coding system ID", Required: true, Type: "int"},
					{Name: "parent-id", Short: "p", Description: "Parent entry ID (omit for root level)", Type: "int"},
				},
			},
			{
				Name:        "search",
				Description: "Search assets by coding system code prefix",
				ToolName:    "UteamupCodingsystemSearchAssets",
				Args:        []ArgDef{{Name: "query", Description: "Code prefix to search (e.g., '1-HLA')", Required: true, Type: "string"}},
			},
			{
				Name:        "next-code",
				Description: "Get next available code for a parent entry",
				ToolName:    "UteamupCodingsystemNextCode",
				Flags: []FlagDef{
					{Name: "coding-system-id", Short: "c", Description: "Coding system ID", Required: true, Type: "int"},
					{Name: "parent-id", Short: "p", Description: "Parent entry ID", Required: true, Type: "int"},
				},
			},
			{
				Name:        "assign",
				Description: "Assign a code catalog entry to an asset",
				ToolName:    "UteamupCodingsystemAssignCode",
				Flags: []FlagDef{
					{Name: "asset-id", Short: "a", Description: "Asset ID", Required: true, Type: "int"},
					{Name: "entry-id", Short: "e", Description: "Code catalog entry ID", Required: true, Type: "int"},
				},
			},
			{
				Name:        "workorders",
				Description: "Get work orders by code branch prefix",
				ToolName:    "UteamupCodingsystemWorkorders",
				Args:        []ArgDef{{Name: "prefix", Description: "Code branch prefix (e.g., '1-HLA')", Required: true, Type: "string"}},
			},
			{
				Name:        "create-workorder",
				Description: "Create a work order from a code catalog entry",
				ToolName:    "UteamupCodingsystemCreateWorkorder",
				Flags: []FlagDef{
					{Name: "entry-id", Short: "e", Description: "Code catalog entry ID", Required: true, Type: "int"},
					{Name: "title", Description: "Work order title", Required: true, Type: "string"},
					{Name: "description", Description: "Work order description", Type: "string"},
					{Name: "priority", Description: "Priority (Low, Medium, High, Critical)", Default: "Medium", Type: "string"},
					{Name: "from-json", Description: "JSON file with work order data", Type: "string"},
				},
			},
		},
	})
}
