package registry

func init() {
	Register(&Domain{
		Name:        "journal",
		Aliases:     []string{"journals"},
		Description: "Manage journal entries and query by code or asset",
		Actions: append(crudActions("Journal"),
			Action{
				Name:        "by-code",
				Description: "List journal entries linked to a code catalog entry",
				ToolName:    "UteamupJournalByCode",
				RESTPath:    "by-code",
				Args:        []ArgDef{{Name: "code-catalog-entry-id", Description: "Code catalog entry ID", Required: true, Type: "int"}},
				Flags: []FlagDef{
					{Name: "page", Short: "p", Description: "Page number", Default: 1, Type: "int"},
					{Name: "page-size", Short: "s", Description: "Items per page", Default: 25, Type: "int"},
				},
			},
			Action{
				Name:        "by-asset",
				Description: "List journal entries for an asset",
				ToolName:    "UteamupJournalByAsset",
				RESTPath:    "by-asset",
				Args:        []ArgDef{{Name: "asset-id", Description: "Asset ID", Required: true, Type: "int"}},
				Flags: []FlagDef{
					{Name: "page", Short: "p", Description: "Page number", Default: 1, Type: "int"},
					{Name: "page-size", Short: "s", Description: "Items per page", Default: 25, Type: "int"},
				},
			},
		),
	})
}
