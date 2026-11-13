package registry

func init() {
	Register(&Domain{
		Name: "search", Description: "Search readable records across the current company",
		Actions: []Action{{
			Name: "global", Description: "Search names, identifiers, indexed text and #tags across permitted record types",
			ToolName: "uteamup_global_search", HTTPMethod: "GET", RESTPath: "global",
			Args: []ArgDef{{Name: "query", QueryName: "query", Description: "Search text (3-1000 characters)", Required: true, Type: "string"}},
			Flags: []FlagDef{
				{Name: "max-results", QueryName: "maxResults", Description: "Matches per type (capped at 25, or 100 with one type)", Type: "int", Default: 5},
				{Name: "entity-types", QueryName: "entityTypes", Description: "Optional comma-separated types such as asset,part,workorder,help,iotalert", Type: "string"},
				{Name: "changed-since", QueryName: "changedSince", Description: "Only records changed on or after this UTC time (ISO-8601)", Type: "string"},
				{Name: "page", QueryName: "page", Description: "Page number when exactly one type is searched", Type: "int", Default: 1},
			},
		}, {
			Name: "scope", Description: "List the record types you may search",
			ToolName: "uteamup_global_search", HTTPMethod: "GET", RESTPath: "scope",
		}},
	})
}
