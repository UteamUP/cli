package registry

func init() {
	// Routes mirror AssetDependencyController (/api/assetdependency/*). Edges are
	// tenant-level facts about the equipment, not group content, so the backend gates
	// them with Asset.View / Asset.Update. Every identifier is a GUID.
	assetGuid := ArgDef{Name: "assetGuid", Description: "Asset GUID", Required: true, Type: "non-empty-uuid"}
	dependencyGuid := ArgDef{Name: "dependencyGuid", Description: "Dependency GUID", Required: true, Type: "non-empty-uuid"}

	Register(&Domain{
		Name:        "asset-dependency",
		Aliases:     []string{"adep", "asset-dependencies"},
		Description: "Record how assets are connected and simulate what one failure takes down",
		APIPath:     "/api/assetdependency",
		Actions: []Action{
			{
				Name:        "list",
				Description: "Every edge touching one asset, incoming and outgoing",
				ToolName:    "UteamupAssetDependencyListByAsset",
				RESTPath:    "by-asset/{assetGuid}",
				Args:        []ArgDef{assetGuid},
			},
			{
				Name:        "create",
				Description: "Create an edge and, with --create-reverse, its mirror",
				ToolName:    "UteamupAssetDependencyCreate",
				Flags: []FlagDef{
					{Name: "asset-group-guid", BodyName: "assetGroupGuid", Description: "Group whose diagram owns this rule", Type: "non-empty-uuid"},
					{Name: "source-asset-guid", BodyName: "sourceAssetGuid", Description: "Asset the edge starts at", Required: true, Type: "non-empty-uuid"},
					{Name: "target-asset-guid", BodyName: "targetAssetGuid", Description: "Asset the edge points at", Required: true, Type: "non-empty-uuid"},
					{Name: "dependency-type", BodyName: "dependencyType", Description: "DependsOn (default), Feeds, Powers, Controls, Cools, Protects, Communicates or Other", Default: "DependsOn", Type: "string"},
					{Name: "failure-impact", BodyName: "failureImpact", Description: "What the target experiences when the source fails: None (default), Degraded, Stopped or SafetyRisk", Default: "None", Type: "string"},
					{Name: "conditions", BodyName: "conditions", Description: "AND/OR observations: {match: all|any, items: [{assetGuid, state}]}", Type: "string", JSONFile: true},
					{Name: "impact-delay-minutes", BodyName: "impactDelayMinutes", Description: "How long the target survives before the impact lands", Type: "int"},
					{Name: "notes", Description: "Why the two assets are connected", Type: "string"},
					{Name: "create-reverse", Description: "JSON object for the mirrored target-to-source edge: {dependencyType, failureImpact, impactDelayMinutes}", Type: "string", JSONFile: true, BodyName: "createReverse"},
				},
			},
			{
				Name:        "update",
				Description: "Patch an edge; endpoints are immutable, so delete and recreate to move one",
				ToolName:    "UteamupAssetDependencyUpdate",
				RESTPath:    "{dependencyGuid}",
				Args:        []ArgDef{dependencyGuid},
				Flags: []FlagDef{
					{Name: "dependency-type", BodyName: "dependencyType", Description: "DependsOn, Feeds, Powers, Controls, Cools, Protects, Communicates or Other", Type: "string"},
					{Name: "failure-impact", BodyName: "failureImpact", Description: "None, Degraded, Stopped or SafetyRisk", Type: "string"},
					{Name: "conditions", BodyName: "conditions", Description: "AND/OR observations: {match: all|any, items: [{assetGuid, state}]}", Type: "string", JSONFile: true},
					{Name: "impact-delay-minutes", BodyName: "impactDelayMinutes", Description: "How long the target survives before the impact lands", Type: "int"},
					{Name: "notes", Description: "Why the two assets are connected", Type: "string"},
				},
			},
			{
				Name:        "delete",
				Description: "Delete an edge; a mirrored edge is a separate row and survives",
				ToolName:    "UteamupAssetDependencyDelete",
				RESTPath:    "{dependencyGuid}",
				Args:        []ArgDef{dependencyGuid},
			},
			{
				Name:        "impact",
				Description: "Simulate one asset failing and report what it takes down across the tenant",
				ToolName:    "UteamupAssetDependencyImpactAnalyze",
				RESTPath:    "by-asset/{assetGuid}/impact",
				Args:        []ArgDef{assetGuid},
				Flags: []FlagDef{
					{Name: "propagate-degraded", QueryName: "propagateDegraded", Description: "Let merely degraded assets pass the failure on, capped at degraded", Type: "bool"},
					{Name: "max-depth", QueryName: "maxDepth", Description: "Hop limit; clamped to 1-100 and echoed back on the result", Type: "int"},
				},
			},
			{
				Name:        "neighbourhood",
				Description: "Everything reachable from one asset within --depth hops, with its groups",
				ToolName:    "UteamupAssetDependencyNeighbourhood",
				RESTPath:    "by-asset/{assetGuid}/neighbourhood",
				Args:        []ArgDef{assetGuid},
				Flags: []FlagDef{
					{Name: "depth", QueryName: "depth", Description: "How many hops to walk, clamped to 1-3", Default: 2, Type: "int"},
				},
			},
		},
	})
}
