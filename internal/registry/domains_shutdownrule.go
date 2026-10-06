package registry

// shutdownTypeValues are the camelCase names the backend's JSON enum converter accepts.
var shutdownTypeValues = []string{"none", "partial", "full", "lockoutTagout"}

func init() {
	// Routes mirror AssetShutdownRuleController (/api/assetshutdownrule/*) and the workorder
	// shutdown override (PUT /api/workorder/by-guid/{guid}/shutdown-override). Rules are asset
	// configuration, so the backend gates them with Asset.View / Asset.Update; the override is a
	// workorder edit gated by WorkOrder.Update. Every identifier is a GUID.
	assetGuid := ArgDef{Name: "assetGuid", Description: "Asset GUID", Required: true, Type: "non-empty-uuid"}
	ruleGuid := ArgDef{Name: "ruleGuid", Description: "Shutdown rule GUID", Required: true, Type: "non-empty-uuid"}
	workorderGuid := ArgDef{Name: "workorderGuid", Description: "Workorder GUID", Required: true, Type: "non-empty-uuid"}

	// Update replaces the whole rule, so it takes the same fields as create; a field left out
	// widens the rule (template or maintenance type) or clears the instructions.
	ruleFlags := func(shutdownTypeRequired bool) []FlagDef {
		return []FlagDef{
			{Name: "shutdown-type", BodyName: "shutdownType", Description: "none, partial, full or lockoutTagout", Required: shutdownTypeRequired, Type: "string", AllowedValues: shutdownTypeValues},
			{Name: "workorder-template-guid", BodyName: "workorderTemplateGuid", Description: "Only work created from this template (default: any template)", Type: "non-empty-uuid"},
			{Name: "maintenance-type", BodyName: "maintenanceType", Description: "Only work of this EN 13306 maintenance type, e.g. preventiveScheduled (default: any type)", Type: "string"},
			{Name: "instructions", Description: "Instructions shown to workers (max 2000 characters)", Type: "string"},
			{Name: "is-active", BodyName: "isActive", Description: "Whether the rule applies", Default: true, Type: "bool"},
		}
	}

	createFlags := append([]FlagDef{
		{Name: "asset-guid", BodyName: "assetGuid", Description: "Asset the rule belongs to", Required: true, Type: "non-empty-uuid"},
	}, ruleFlags(true)...)

	Register(&Domain{
		Name:        "shutdown-rule",
		Aliases:     []string{"shutdown-rules", "asset-shutdown-rule"},
		Description: "Manage which work on an asset needs which shutdown, and override a workorder's required shutdown",
		APIPath:     "/api/assetshutdownrule",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List the shutdown rules on an asset",
				ToolName:    "UteamupAssetShutdownRuleList",
				RESTPath:    "asset/by-guid/{assetGuid}",
				Args:        []ArgDef{assetGuid},
			},
			{
				Name:        "create",
				Description: "Add a shutdown rule to an asset; leave out template and maintenance type to cover all work",
				ToolName:    "UteamupAssetShutdownRuleCreate",
				Flags:       createFlags,
			},
			{
				Name:        "update",
				Description: "Replace a shutdown rule (every field is replaced)",
				ToolName:    "UteamupAssetShutdownRuleUpdate",
				RESTPath:    "by-guid/{ruleGuid}",
				Args:        []ArgDef{ruleGuid},
				Flags:       ruleFlags(true),
			},
			{
				Name:        "delete",
				Description: "Delete a shutdown rule",
				ToolName:    "UteamupAssetShutdownRuleDelete",
				RESTPath:    "by-guid/{ruleGuid}",
				Args:        []ArgDef{ruleGuid},
			},
			{
				Name:         "set-override",
				Description:  "Override a workorder's required shutdown; leave out --shutdown-type to clear the override",
				ToolName:     "UteamupWorkorderSetShutdownOverride",
				RESTBasePath: "/api/workorder",
				RESTPath:     "by-guid/{workorderGuid}/shutdown-override",
				HTTPMethod:   "PUT",
				Args:         []ArgDef{workorderGuid},
				Flags: []FlagDef{
					{Name: "shutdown-type", BodyName: "shutdownType", Description: "none, partial, full or lockoutTagout; leave out to clear the override", Type: "string", AllowedValues: shutdownTypeValues},
					{Name: "reason", Description: "Why the planner overrides the resolved shutdown (max 500 characters)", Type: "string"},
				},
			},
		},
	})
}
