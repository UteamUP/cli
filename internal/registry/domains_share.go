package registry

func init() {
	// Routes mirror EntityShareController (/api/share/*). The management routes are
	// per-kind (api/share/asset, api/share/assetgroup), so --entity-type fills the route
	// segment as well as naming the permission pair the backend requires. Its
	// AllowedValues are the escaping: RESTPath expansion does not escape, so an
	// unconstrained flag there would let a caller reshape the route.
	//
	// The two routes that address a share by its own GUID are kind-agnostic on the wire.
	// They still take --entity-type, because the caller has to state which permission pair
	// they are exercising, exactly as the MCP tools do — but it travels on the query
	// string there so it never lands in a body the update model does not declare.
	shareGuid := ArgDef{Name: "shareGuid", Description: "Share GUID", Required: true, Type: "non-empty-uuid"}
	const entityTypeDescription = "What is being shared: asset or assetgroup"
	routeEntityType := FlagDef{
		Name:          "entity-type",
		BodyName:      "entityType",
		Description:   entityTypeDescription,
		Required:      true,
		Type:          "string",
		AllowedValues: []string{"asset", "assetgroup"},
	}
	queryEntityType := FlagDef{
		Name:          "entity-type",
		QueryName:     "entityType",
		Description:   entityTypeDescription + " — names the permission pair this call exercises",
		Required:      true,
		Type:          "string",
		AllowedValues: []string{"asset", "assetgroup"},
	}

	Register(&Domain{
		Name:        "share",
		Aliases:     []string{"shares", "entity-share"},
		Description: "Share one asset or asset group outside the tenant, scoped to chosen sections",
		APIPath:     "/api/share",
		Actions: []Action{
			{
				Name:        "create",
				Description: "Share an asset or asset group with a UteamUP user by email, or as an anonymous link",
				ToolName:    "UteamupShareCreate",
				HTTPMethod:  "POST",
				RESTPath:    "{entityType}",
				Flags: []FlagDef{
					routeEntityType,
					{Name: "target-guid", BodyName: "targetGuid", Description: "GUID of the asset or asset group being shared", Required: true, Type: "non-empty-uuid"},
					{Name: "email", Description: "Email of an existing UteamUP user (omit for an anonymous link)", Type: "string"},
					{Name: "access-level", BodyName: "accessLevel", Description: "Ceiling: read (default), update or write. No section may exceed it", Default: "read", Type: "string"},
					{Name: "section", BodyName: "sections", Description: "Per-section scope as key=level, repeatable (e.g. --section diagram=update --section meters=none). Omit for the safe default grid", Type: "stringSlice"},
					{Name: "expires-in-days", BodyName: "expiresInDays", Description: "Days until the share expires, 1-365 (default 30)", Type: "int"},
					{Name: "note", Description: "Optional note shown to the recipient", Type: "string"},
				},
			},
			{
				Name:        "list",
				Description: "List the active and revoked shares of one asset or asset group",
				ToolName:    "UteamupShareListByTarget",
				RESTPath:    "{entityType}/by-target/{targetGuid}",
				Flags: []FlagDef{
					routeEntityType,
					{Name: "target-guid", BodyName: "targetGuid", Description: "GUID of the shared asset or asset group", Required: true, Type: "non-empty-uuid"},
				},
			},
			{
				Name:        "update",
				Description: "Change the ceiling, the section grid or the expiry of one existing share",
				ToolName:    "UteamupShareUpdate",
				HTTPMethod:  "PATCH",
				RESTPath:    "{shareGuid}",
				Args:        []ArgDef{shareGuid},
				Flags: []FlagDef{
					queryEntityType,
					{Name: "access-level", BodyName: "accessLevel", Description: "New ceiling: read, update or write", Type: "string"},
					{Name: "section", BodyName: "sections", Description: "Replacement per-section scope as key=level, repeatable. A supplied list replaces the grid wholesale", Type: "stringSlice"},
					{Name: "expires-in-days", BodyName: "expiresInDays", Description: "New expiry counted from now, 1-365", Type: "int"},
				},
			},
			{
				Name:        "revoke",
				Description: "Revoke a share so its link stops working immediately",
				ToolName:    "UteamupShareRevoke",
				HTTPMethod:  "DELETE",
				RESTPath:    "{shareGuid}",
				Args:        []ArgDef{shareGuid},
				Flags:       []FlagDef{queryEntityType},
			},
			{
				// The REST route returns both kinds and takes no filter, so this action
				// deliberately has no --entity-type: offering one would be a silent no-op.
				Name:        "shared-with-me",
				Description: "Assets and asset groups shared with you, across every tenant",
				ToolName:    "UteamupShareSharedWithMe",
				RESTPath:    "shared-with-me",
			},
		},
	})
}
