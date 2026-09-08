package registry

func init() {
	// Routes mirror AssetGroupController and its three partials
	// (Impact, Labels, Generation) at /api/assetgroup/*. Every identifier is a GUID;
	// the int-keyed twins are not exposed here.
	groupGuid := ArgDef{Name: "groupGuid", Description: "Asset group GUID", Required: true, Type: "non-empty-uuid"}
	dependencyGuid := ArgDef{Name: "dependencyGuid", Description: "Group dependency GUID", Required: true, Type: "non-empty-uuid"}
	generationGuid := ArgDef{Name: "generationGuid", Description: "Link generation GUID", Required: true, Type: "non-empty-uuid"}

	Register(&Domain{
		Name:        "asset-group",
		Aliases:     []string{"ag", "asset-groups"},
		Description: "Group assets, draw how they are connected, and simulate what a failure does",
		APIPath:     "/api/assetgroup",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List the tenant's asset groups with member and dependency counts",
				ToolName:    "UteamupAssetGroupList",
				Flags: []FlagDef{
					{Name: "search", Description: "Case-insensitive name/description filter", Type: "string"},
					{Name: "kind", Description: "Restrict to one kind: System, ProductionLine, Area, Fleet, Utility, SafetySystem or Custom", Type: "string"},
					{Name: "include-inactive", BodyName: "includeInactive", Description: "Include soft-deleted groups", Type: "bool"},
					{Name: "page", Short: "p", Description: "Page number", Default: 1, Type: "int"},
					{Name: "page-size", Short: "s", BodyName: "pageSize", Description: "Rows per page (1-200)", Default: 50, Type: "int"},
				},
			},
			{
				Name:        "get",
				Description: "Read one group with its members and both directions of its group edges",
				ToolName:    "UteamupAssetGroupGet",
				RESTPath:    "by-guid/{groupGuid}",
				Args:        []ArgDef{groupGuid},
			},
			{
				Name:        "create",
				Description: "Create a group, optionally seeded with member assets",
				ToolName:    "UteamupAssetGroupCreate",
				Flags: []FlagDef{
					{Name: "name", Description: "Group name", Required: true, Type: "string"},
					{Name: "description", Description: "What the group covers", Type: "string"},
					{Name: "kind", Description: "System, ProductionLine, Area, Fleet, Utility, SafetySystem or Custom", Default: "Custom", Type: "string"},
					{Name: "icon-name", BodyName: "iconName", Description: "Icon name shown on the group card", Type: "string"},
					{Name: "color", Description: "Accent colour, e.g. #c2410c", Type: "string"},
					{Name: "member-asset-guid", BodyName: "memberAssetGuids", Description: "Asset GUID to seed the group with, repeatable", Type: "stringSlice"},
				},
			},
			{
				Name:        "update",
				Description: "Update the group header; membership is changed through members-add / members-remove",
				ToolName:    "UteamupAssetGroupUpdate",
				RESTPath:    "by-guid/{groupGuid}",
				Args:        []ArgDef{groupGuid},
				Flags: []FlagDef{
					{Name: "name", Description: "Group name", Required: true, Type: "string"},
					{Name: "description", Description: "What the group covers", Type: "string"},
					{Name: "kind", Description: "System, ProductionLine, Area, Fleet, Utility, SafetySystem or Custom", Default: "Custom", Type: "string"},
					{Name: "icon-name", BodyName: "iconName", Description: "Icon name shown on the group card", Type: "string"},
					{Name: "color", Description: "Accent colour, e.g. #c2410c", Type: "string"},
					{Name: "active", BodyName: "isActive", Description: "false soft-deletes the group, true restores it", Type: "bool"},
				},
			},
			{
				Name:        "delete",
				Description: "Soft-delete the group; it stays visible with --include-inactive",
				ToolName:    "UteamupAssetGroupDelete",
				RESTPath:    "by-guid/{groupGuid}",
				Args:        []ArgDef{groupGuid},
			},
			{
				Name:        "members-add",
				Description: "Add assets to the group and return the resulting membership",
				ToolName:    "UteamupAssetGroupMembersAdd",
				HTTPMethod:  "POST",
				RESTPath:    "by-guid/{groupGuid}/members",
				Args:        []ArgDef{groupGuid},
				Flags: []FlagDef{
					{Name: "asset-guid", BodyName: "assetGuids", Description: "Asset GUID to add, repeatable", Required: true, Type: "stringSlice"},
					{Name: "role", Description: "Role given to every added asset: Primary, Supporting (default) or Spare", Type: "string"},
				},
			},
			{
				Name:        "members-remove",
				Description: "Remove one asset from the group; the asset itself is untouched",
				ToolName:    "UteamupAssetGroupMembersRemove",
				HTTPMethod:  "DELETE",
				RESTPath:    "by-guid/{groupGuid}/members/{assetGuid}",
				Args: []ArgDef{
					groupGuid,
					{Name: "assetGuid", Description: "Asset GUID to remove", Required: true, Type: "non-empty-uuid"},
				},
			},
			{
				Name:        "diagram",
				Description: "Member nodes with their saved positions plus every member-to-member edge",
				ToolName:    "UteamupAssetGroupDiagramGet",
				RESTPath:    "by-guid/{groupGuid}/diagram",
				Args:        []ArgDef{groupGuid},
			},
			{
				Name:        "positions-save",
				Description: "Persist dragged node positions and, when supplied, the canvas zoom/pan",
				ToolName:    "UteamupAssetGroupPositionsSave",
				HTTPMethod:  "PUT",
				RESTPath:    "by-guid/{groupGuid}/diagram/positions",
				Args:        []ArgDef{groupGuid},
				Flags: []FlagDef{
					{Name: "from-json", Description: "JSON object with positions[] ({assetGuid,x,y}) and the optional canvasState", Required: true, Type: "string", RootJSONObjectFile: true},
				},
			},
			{
				Name:        "auto-layout",
				Description: "Lay every member out on the layered dependency layout and save the positions",
				ToolName:    "UteamupAssetGroupAutoLayout",
				HTTPMethod:  "POST",
				RESTPath:    "by-guid/{groupGuid}/diagram/auto-layout",
				Args:        []ArgDef{groupGuid},
			},
			{
				Name:        "map",
				Description: "The tenant-level group map: every active group and every group-to-group edge",
				ToolName:    "UteamupAssetGroupMapGet",
				RESTPath:    "map",
			},
			{
				Name:        "dependencies",
				Description: "Every group-to-group edge touching the group, incoming and outgoing",
				ToolName:    "UteamupAssetGroupDependenciesList",
				RESTPath:    "by-guid/{groupGuid}/dependencies",
				Args:        []ArgDef{groupGuid},
			},
			{
				Name:        "dependency-create",
				Description: "Create a group-to-group edge",
				ToolName:    "UteamupAssetGroupDependenciesCreate",
				HTTPMethod:  "POST",
				RESTPath:    "dependencies",
				Flags: []FlagDef{
					{Name: "source-group-guid", BodyName: "sourceGroupGuid", Description: "Group the edge starts at", Required: true, Type: "non-empty-uuid"},
					{Name: "target-group-guid", BodyName: "targetGroupGuid", Description: "Group the edge points at", Required: true, Type: "non-empty-uuid"},
					{Name: "dependency-type", BodyName: "dependencyType", Description: "DependsOn (default), Feeds, Powers, Controls, Cools, Protects, Communicates or Other", Default: "DependsOn", Type: "string"},
					{Name: "failure-impact", BodyName: "failureImpact", Description: "What the target experiences when the source fails: None (default), Degraded, Stopped or SafetyRisk", Default: "None", Type: "string"},
					{Name: "impact-delay-minutes", BodyName: "impactDelayMinutes", Description: "How long the target survives before the impact lands", Type: "int"},
					{Name: "notes", Description: "Why the two systems are connected", Type: "string"},
				},
			},
			{
				Name:        "dependency-delete",
				Description: "Delete a group-to-group edge",
				ToolName:    "UteamupAssetGroupDependenciesDelete",
				HTTPMethod:  "DELETE",
				RESTPath:    "dependencies/{dependencyGuid}",
				Args:        []ArgDef{dependencyGuid},
			},
			{
				Name:        "impact",
				Description: "Simulate a whole group failing and report what it takes down across the tenant",
				ToolName:    "UteamupAssetGroupImpactAnalyze",
				RESTPath:    "impact",
				Args: []ArgDef{
					{Name: "failedGroupGuid", Description: "Group to simulate as failed", Required: true, Type: "non-empty-uuid", QueryName: "failedGroupGuid"},
				},
				Flags: []FlagDef{
					{Name: "propagate-degraded", QueryName: "propagateDegraded", Description: "Let merely degraded assets pass the failure on, capped at degraded", Type: "bool"},
					{Name: "max-depth", QueryName: "maxDepth", Description: "Hop limit; clamped to 1-100 and echoed back on the result", Type: "int"},
				},
			},
			{
				Name:        "blast-radius",
				Description: "Per-member reachability: who takes every other member down on their own",
				ToolName:    "UteamupAssetGroupBlastRadius",
				RESTPath:    "by-guid/{groupGuid}/blast-radius",
				Args:        []ArgDef{groupGuid},
			},
			{
				Name:        "live-status",
				Description: "Current status of every member with the signal that produced it",
				ToolName:    "UteamupAssetGroupLiveStatus",
				RESTPath:    "by-guid/{groupGuid}/live-status",
				Args:        []ArgDef{groupGuid},
			},
			{
				Name:        "label",
				Description: "What a printable label encodes: QR target, barcode value and assigned codes",
				ToolName:    "UteamupAssetGroupLabelPayload",
				RESTPath:    "by-guid/{groupGuid}/label-payload",
				Args:        []ArgDef{groupGuid},
			},
			{
				Name:        "links-generate",
				Description: "Ask UPMate to propose the connections; spends AI credits and writes nothing",
				ToolName:    "UteamupAssetGroupLinksGenerate",
				HTTPMethod:  "POST",
				RESTPath:    "links/generate",
				Flags: []FlagDef{
					{Name: "prompt", Description: "How the equipment is wired, in plain language (10-4000 chars)", Required: true, Type: "string"},
					{Name: "mode", Description: "AssetLinks (default, edges between a group's members) or GroupLinks (edges between groups)", Default: "AssetLinks", Type: "string"},
					{Name: "asset-group-guid", BodyName: "assetGroupGuid", Description: "Group whose members to connect; required in AssetLinks mode", Type: "uuid"},
					{Name: "document-guid", BodyName: "documentGuids", Description: "Document GUID to attach as evidence (e.g. the P&ID), up to five, repeatable", Type: "stringSlice"},
					{Name: "request-guid", BodyName: "requestGuid", Description: "Caller-generated GUID that deduplicates a retried generation", Required: true, Type: "non-empty-uuid"},
				},
			},
			{
				Name:        "links-get",
				Description: "Re-read a stored proposal, so reviewing it again costs no credits",
				ToolName:    "UteamupAssetGroupLinksGenerationGet",
				RESTPath:    "links/generations/{generationGuid}",
				Args:        []ArgDef{generationGuid},
			},
			{
				Name:        "links-apply",
				Description: "Write the proposals a human accepted from one stored generation",
				ToolName:    "UteamupAssetGroupLinksApply",
				HTTPMethod:  "POST",
				RESTPath:    "links/from-generation/{generationGuid}",
				Args:        []ArgDef{generationGuid},
				Flags: []FlagDef{
					{Name: "accepted-link-id", BodyName: "acceptedLinkIds", Description: "Proposal id to write, repeatable; anything not listed is discarded", Type: "stringSlice"},
					{Name: "from-json", Description: "JSON object with acceptedLinkIds[] and the optional per-link overrides[]", Type: "string", RootJSONObjectFile: true},
				},
			},
		},
	})
}
