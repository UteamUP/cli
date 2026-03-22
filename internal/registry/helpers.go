package registry

// paginationFlags returns standard pagination flags used across most list actions.
func paginationFlags() []FlagDef {
	return []FlagDef{
		{Name: "page", Short: "p", Description: "Page number", Default: 1, Type: "int"},
		{Name: "page-size", Short: "s", Description: "Items per page", Default: 25, Type: "int"},
	}
}

// searchFlags returns pagination + query flags for search actions.
func searchFlags() []FlagDef {
	return append(paginationFlags(), FlagDef{Name: "filter", Short: "f", Description: "Filter/search term", Type: "string"})
}

// idArg returns a standard required integer ID positional argument.
func idArg() []ArgDef {
	return []ArgDef{{Name: "id", Description: "Record ID", Required: true, Type: "int"}}
}

// stringIDArg returns a standard required string ID positional argument (for GUIDs).
func stringIDArg() []ArgDef {
	return []ArgDef{{Name: "id", Description: "Record ID", Required: true, Type: "string"}}
}

// queryArg returns a required search query positional argument.
func queryArg() []ArgDef {
	return []ArgDef{{Name: "query", Description: "Search term", Required: true, Type: "string"}}
}

// jsonFlag returns the --from-json flag for JSON file input.
func jsonFlag() FlagDef {
	return FlagDef{Name: "from-json", Description: "JSON file with request data", Type: "string"}
}

// nameFlag returns a required --name flag.
func nameFlag() FlagDef {
	return FlagDef{Name: "name", Description: "Name", Required: true, Type: "string"}
}

// crudActions returns standard CRUD actions for a domain.
func crudActions(entityPrefix string) []Action {
	return []Action{
		{Name: "list", Description: "List records", ToolName: "Uteamup" + entityPrefix + "List", Flags: paginationFlags()},
		{Name: "get", Description: "Get by ID", ToolName: "Uteamup" + entityPrefix + "Get", Args: idArg()},
		{Name: "create", Description: "Create a record", ToolName: "Uteamup" + entityPrefix + "Create", Flags: []FlagDef{jsonFlag()}},
		{Name: "update", Description: "Update a record", ToolName: "Uteamup" + entityPrefix + "Update", Args: idArg(), Flags: []FlagDef{jsonFlag()}},
		{Name: "delete", Description: "Delete a record", ToolName: "Uteamup" + entityPrefix + "Delete", Args: idArg()},
	}
}

// listGetActions returns list + get actions for read-only or simple domains.
func listGetActions(entityPrefix string) []Action {
	return []Action{
		{Name: "list", Description: "List records", ToolName: "Uteamup" + entityPrefix + "List", Flags: paginationFlags()},
		{Name: "get", Description: "Get by ID", ToolName: "Uteamup" + entityPrefix + "Get", Args: idArg()},
	}
}
