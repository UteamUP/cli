package registry

func init() {
	Register(&Domain{
		Name: "search", Description: "Search readable records across the current company",
		Actions: []Action{{
			Name: "global", Description: "Search names, identifiers, indexed text and #tags across permitted record types",
			ToolName: "uteamup_global_search", HTTPMethod: "GET", RESTPath: "global",
			Args: []ArgDef{{Name: "query", QueryName: "query", Description: "Search text (3-1000 characters)", Required: true, Type: "string"}},
			Flags: []FlagDef{
				{Name: "max-results", QueryName: "maxResults", Description: "Matches per type (capped at 25)", Type: "int", Default: 5},
				{Name: "entity-types", QueryName: "entityTypes", Description: "Optional comma-separated types such as asset,part,workorder", Type: "string"},
			},
		}},
	})
}
