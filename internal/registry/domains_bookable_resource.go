package registry

func bookableResourceMutationFlags() []FlagDef {
	return []FlagDef{
		{Name: "name", BodyName: "name", Description: "Resource name", Required: true, Type: "string"},
		{Name: "description", BodyName: "description", Description: "Optional resource description", Type: "string"},
		{Name: "resource-type", BodyName: "resourceType", Description: "0=technician, 1=contractor, 2=crew, 3=equipment, 4=vehicle, 5=facility, 6=pool", Required: true, Type: "int"},
		{Name: "capacity", BodyName: "capacity", Description: "Reservable capacity greater than zero", Default: 1.0, Type: "float"},
		{Name: "capacity-unit", BodyName: "capacityUnit", Description: "Capacity unit label", Default: "unit", Type: "string"},
		{Name: "is-active", BodyName: "isActive", Description: "Whether the resource can be selected", Default: true, Type: "bool"},
		{Name: "atomic-pool-selection", BodyName: "atomicPoolSelection", Description: "Reserve one concrete eligible pool member atomically", Default: true, Type: "bool"},
		{Name: "user-guid", BodyName: "userGuid", Description: "Technician user public GUID", Type: "string"},
		{Name: "contractor-profile-guid", BodyName: "contractorProfileGuid", Description: "Contractor profile public GUID", Type: "string"},
		{Name: "contractor-crew-guid", BodyName: "contractorCrewGuid", Description: "Contractor crew public GUID", Type: "string"},
		{Name: "asset-guid", BodyName: "assetGuid", Description: "Equipment or vehicle asset public GUID", Type: "string"},
		{Name: "location-guid", BodyName: "locationGuid", Description: "Facility location public GUID", Type: "string"},
		{Name: "territory-guid", BodyName: "territoryGuids", Description: "Service territory public GUID (repeatable)", Type: "stringSlice"},
	}
}

func serviceTerritoryMutationFlags() []FlagDef {
	return []FlagDef{
		{Name: "name", BodyName: "name", Description: "Service territory name", Required: true, Type: "string"},
		{Name: "description", BodyName: "description", Description: "Optional territory description", Type: "string"},
		{Name: "center-latitude", BodyName: "centerLatitude", Description: "Optional centre latitude", Type: "float"},
		{Name: "center-longitude", BodyName: "centerLongitude", Description: "Optional centre longitude", Type: "float"},
		{Name: "radius-km", BodyName: "radiusKm", Description: "Optional radius in kilometres", Type: "float"},
		{Name: "boundary-geo-json", BodyName: "boundaryGeoJson", Description: "Optional RFC 7946 Polygon or MultiPolygon", Type: "string"},
		{Name: "time-zone-id", BodyName: "timeZoneId", Description: "Optional IANA or Windows time-zone identifier", Type: "string"},
		{Name: "is-active", BodyName: "isActive", Description: "Whether the territory can constrain scheduling", Default: true, Type: "bool"},
	}
}

func bookableResourceRequirementFlags() []FlagDef {
	return []FlagDef{
		{Name: "role-name", BodyName: "roleName", Description: "Required scheduling role", Default: "Supporting resource", Type: "string"},
		{Name: "resource-type", BodyName: "resourceType", Description: "Required bookable-resource type number", Required: true, Type: "int"},
		{Name: "mode", BodyName: "mode", Description: "0=required hard constraint, 1=preferred ranking input", Default: 0, Type: "int"},
		{Name: "quantity", BodyName: "quantity", Description: "Required resource count", Default: 1, Type: "int"},
		{Name: "capacity-required", BodyName: "capacityRequired", Description: "Capacity required from each allocation", Default: 1.0, Type: "float"},
		{Name: "specific-resource-guid", BodyName: "specificResourceGuid", Description: "Optional exact resource public GUID", Type: "string"},
		{Name: "pool-resource-guid", BodyName: "poolResourceGuid", Description: "Optional pool public GUID", Type: "string"},
		{Name: "service-territory-guid", BodyName: "serviceTerritoryGuid", Description: "Optional service territory public GUID", Type: "string"},
		{Name: "is-active", BodyName: "isActive", Description: "Whether the requirement is enforced", Default: true, Type: "bool"},
	}
}

func expectedUpdatedAtQueryFlag() FlagDef {
	return FlagDef{
		Name:        "expected-updated-at",
		QueryName:   "expectedUpdatedAt",
		Description: "Exact reviewed UpdatedAt timestamp, sent in the query string",
		Required:    true,
		Type:        "string",
	}
}

func init() {
	Register(&Domain{
		Name:        "bookable-resource",
		Aliases:     []string{"bookable-resources", "schedule-resources"},
		Description: "Manage Preview scheduling resources, pools, territories, requirements, and route evidence",
		APIPath:     "/api/bookableresources",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List tenant-scoped bookable resources and capacity",
				ToolName:    "UteamupBookableResourceList",
				HTTPMethod:  "GET",
				Flags: []FlagDef{
					{Name: "search", BodyName: "search", Description: "Name or description search", Type: "string"},
					{Name: "resource-type", BodyName: "resourceType", Description: "Optional resource type number", Type: "int"},
					{Name: "territory-guid", BodyName: "territoryGuid", Description: "Optional service territory public GUID", Type: "string"},
					{Name: "is-active", BodyName: "isActive", Description: "Optional active-state filter", Type: "bool"},
					{Name: "page", BodyName: "page", Description: "One-based page number", Default: 1, Type: "int"},
					{Name: "page-size", BodyName: "pageSize", Description: "Results per page, maximum 100", Default: 25, Type: "int"},
				},
			},
			{
				Name:        "get",
				Description: "Get one resource with pool and territory evidence",
				ToolName:    "UteamupBookableResourceGet",
				HTTPMethod:  "GET",
				RESTPath:    "{resourceGuid}",
				Args: []ArgDef{
					{Name: "resourceGuid", Description: "Bookable-resource public GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "create",
				Description: "Create a reviewed bookable resource",
				ToolName:    "UteamupBookableResourceCreate",
				HTTPMethod:  "POST",
				Flags:       bookableResourceMutationFlags(),
			},
			{
				Name:        "update",
				Description: "Update an exact bookable-resource version",
				ToolName:    "UteamupBookableResourceUpdate",
				HTTPMethod:  "PUT",
				RESTPath:    "{resourceGuid}",
				Args: []ArgDef{
					{Name: "resourceGuid", Description: "Bookable-resource public GUID", Required: true, Type: "uuid"},
				},
				Flags: append(
					bookableResourceMutationFlags(),
					expectedUpdatedAtQueryFlag(),
				),
			},
			{
				Name:        "pool-members-set",
				Description: "Replace a pool's eligible member set atomically from reviewed JSON",
				ToolName:    "UteamupBookableResourcePoolMembersSet",
				HTTPMethod:  "PUT",
				RESTPath:    "{poolGuid}/members",
				Args: []ArgDef{
					{Name: "poolGuid", Description: "Bookable-resource pool public GUID", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{
					{Name: "members-file", BodyName: "members", Description: "JSON file containing reviewed pool-member rows", Required: true, Type: "string", JSONFile: true},
				},
			},
			{
				Name:        "territory-list",
				Description: "List service territories used by scheduling",
				ToolName:    "UteamupServiceTerritoryList",
				HTTPMethod:  "GET",
				RESTPath:    "territories",
			},
			{
				Name:        "territory-create",
				Description: "Create a reviewed service territory",
				ToolName:    "UteamupServiceTerritoryCreate",
				HTTPMethod:  "POST",
				RESTPath:    "territories",
				Flags:       serviceTerritoryMutationFlags(),
			},
			{
				Name:        "territory-update",
				Description: "Update an exact service-territory version",
				ToolName:    "UteamupServiceTerritoryUpdate",
				HTTPMethod:  "PUT",
				RESTPath:    "territories/{territoryGuid}",
				Args: []ArgDef{
					{Name: "territoryGuid", Description: "Service territory public GUID", Required: true, Type: "uuid"},
				},
				Flags: append(
					serviceTerritoryMutationFlags(),
					expectedUpdatedAtQueryFlag(),
				),
			},
			{
				Name:        "requirement-list",
				Description: "List hard and preferred resource requirements for a work order",
				ToolName:    "UteamupBookableResourceRequirementList",
				HTTPMethod:  "GET",
				RESTPath:    "workorders/{workorderGuid}/requirements",
				Args: []ArgDef{
					{Name: "workorderGuid", Description: "Work-order public GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "requirement-create",
				Description: "Add a hard or preferred resource requirement",
				ToolName:    "UteamupBookableResourceRequirementCreate",
				HTTPMethod:  "POST",
				RESTPath:    "workorders/{workorderGuid}/requirements",
				Args: []ArgDef{
					{Name: "workorderGuid", Description: "Work-order public GUID", Required: true, Type: "uuid"},
				},
				Flags: bookableResourceRequirementFlags(),
			},
			{
				Name:        "requirement-update",
				Description: "Update an exact resource-requirement version",
				ToolName:    "UteamupBookableResourceRequirementUpdate",
				HTTPMethod:  "PUT",
				RESTPath:    "requirements/{requirementGuid}",
				Args: []ArgDef{
					{Name: "requirementGuid", Description: "Resource-requirement public GUID", Required: true, Type: "uuid"},
				},
				Flags: append(
					bookableResourceRequirementFlags(),
					expectedUpdatedAtQueryFlag(),
				),
			},
			{
				Name:        "route-estimate",
				Description: "Get directional, departure-aware route evidence with honest degradation labels",
				ToolName:    "UteamupBookableResourceRouteEstimate",
				HTTPMethod:  "POST",
				RESTPath:    "route-estimate",
				Flags: []FlagDef{
					{Name: "from-latitude", BodyName: "fromLatitude", Description: "Origin latitude", Required: true, Type: "float"},
					{Name: "from-longitude", BodyName: "fromLongitude", Description: "Origin longitude", Required: true, Type: "float"},
					{Name: "to-latitude", BodyName: "toLatitude", Description: "Destination latitude", Required: true, Type: "float"},
					{Name: "to-longitude", BodyName: "toLongitude", Description: "Destination longitude", Required: true, Type: "float"},
					{Name: "departure-time-utc", BodyName: "departureTimeUtc", Description: "UTC route departure time", Required: true, Type: "string"},
				},
			},
		},
	})
}
