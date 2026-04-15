package registry

// Logbook-import CLI surface — mirrors LogbookImport MCP tools.

func init() {
	Register(&Domain{
		Name:        "logbook-import",
		Aliases:     []string{"logimp"},
		Description: "Inspect parsed Word-logbook (.docx) imports before committing entries as Journals",
		Actions: []Action{
			{
				Name:        "get",
				Description: "Get a parsed logbook import with its entries",
				ToolName:    "UteamupLogbookImportGet",
				Flags: []FlagDef{
					{Name: "batch-id", Short: "b", Description: "Batch ID of the logbook import", Required: true, Type: "int"},
				},
			},
		},
	})
}
