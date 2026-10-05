package registry

func init() {
	// Routes mirror AssetAiBuilderController (/api/asset/{assetGuid}/ai-builder/*).
	// Readiness is free; generate charges per generated kind and saves nothing; apply is
	// free, saves only the reviewed items, and is idempotent per requestGuid.
	assetGuid := ArgDef{Name: "assetGuid", Description: "Asset ExternalGuid", Required: true, Type: "non-empty-uuid"}

	Register(&Domain{
		Name:        "asset-ai-builder",
		Aliases:     []string{"ai-builder"},
		Description: "Let UPMate generate checklists, resources, specs, meters and diagrams for one asset from its own information",
		APIPath:     "/api/asset",
		Actions: []Action{
			{
				Name:        "readiness",
				Description: "What UPMate can generate for the asset and what information it lacks (free)",
				ToolName:    "UteamupAssetAiBuilderReadiness",
				RESTPath:    "{assetGuid}/ai-builder/readiness",
				Args:        []ArgDef{assetGuid},
			},
			{
				Name:        "generate",
				Description: "Generate a review-first proposal; each generated kind costs AI credits, nothing is saved",
				ToolName:    "UteamupAssetAiBuilderGenerate",
				HTTPMethod:  "POST",
				RESTPath:    "{assetGuid}/ai-builder/generate",
				Args:        []ArgDef{assetGuid},
				Flags: []FlagDef{
					{Name: "kinds", Short: "k", Description: "Kinds to generate: Checklist, Tasklist, Resources, OperatingSpecs, MeterReadings, MermaidDiagram", Required: true, Type: "stringSlice"},
					{Name: "instructions", Short: "i", Description: "Optional guidance for UPMate (max 2000 characters)", Type: "string"},
				},
			},
			{
				Name:        "apply",
				Description: "Save the reviewed items of a proposal (free; each item keeps its own permission)",
				ToolName:    "UteamupAssetAiBuilderApply",
				HTTPMethod:  "POST",
				RESTPath:    "{assetGuid}/ai-builder/apply",
				Args:        []ArgDef{assetGuid},
				Flags: []FlagDef{
					{Name: "file", Short: "f", Description: "JSON with requestGuid, generationGuid and the kept checklists, tasklists, resources, operatingSpecs, meters and mermaidDiagrams", Required: true, Type: "string", RootJSONObjectFile: true},
				},
			},
			{
				Name:        "history",
				Description: "When UPMate last generated each kind for the asset, and from how many documents",
				ToolName:    "UteamupAssetAiBuilderHistory",
				RESTPath:    "{assetGuid}/ai-builder/history",
				Args:        []ArgDef{assetGuid},
			},
		},
	})
}
