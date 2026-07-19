package registry

func init() {
	requirementFlags := []FlagDef{
		{Name: "name", BodyName: "name", Description: "Plain-language resource requirement name", Type: "string"},
		{Name: "exact-asset-guid", BodyName: "exactAssetGuid", Description: "Exact tenant asset GUID; mutually exclusive with asset-type-guid", Type: "uuid"},
		{Name: "asset-type-guid", BodyName: "assetTypeGuid", Description: "Tenant or system asset-type GUID; mutually exclusive with exact-asset-guid", Type: "uuid"},
		{Name: "quantity", BodyName: "quantity", Description: "Required reusable asset quantity (1-100)", Type: "int", Default: 1},
		{Name: "mandatory", BodyName: "isMandatory", Description: "Block scheduling when evidence is stale or availability is insufficient", Type: "bool", Default: true},
		{Name: "active", BodyName: "isActive", Description: "Include this requirement in schedule readiness", Type: "bool", Default: true},
		{Name: "availability-validated-at-utc", BodyName: "availabilityValidatedAtUtc", Description: "UTC timestamp of the latest availability validation", Type: "string"},
		{Name: "max-evidence-age-hours", BodyName: "maxEvidenceAgeHours", Description: "Maximum evidence age in hours (1-720)", Type: "int", Default: 24},
	}

	Register(&Domain{
		Name:        "workorder-supporting-asset",
		Aliases:     []string{"wo-supporting-asset", "wo-resource"},
		Description: "Review and govern reusable supporting assets required by workorders",
		APIPath:     "/api/workorder",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List supporting resources and readiness evidence for one workorder",
				ToolName:    "UteamupWorkorderSupportingAssetRequirementList",
				HTTPMethod:  "GET",
				RESTPath:    "{workorderGuid}/supporting-assets",
				Args: []ArgDef{
					{Name: "workorderGuid", Description: "Tenant-scoped workorder GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "create",
				Description: "Declare one exact asset or asset type required by a workorder",
				ToolName:    "UteamupWorkorderSupportingAssetRequirementCreate",
				HTTPMethod:  "POST",
				RESTPath:    "{workorderGuid}/supporting-assets",
				Args: []ArgDef{
					{Name: "workorderGuid", Description: "Tenant-scoped workorder GUID", Required: true, Type: "uuid"},
				},
				Flags: requirementFlags,
			},
			{
				Name:        "update",
				Description: "Replace one supporting-resource requirement after review",
				ToolName:    "UteamupWorkorderSupportingAssetRequirementUpdate",
				HTTPMethod:  "PUT",
				RESTPath:    "{workorderGuid}/supporting-assets/{requirementGuid}",
				Args: []ArgDef{
					{Name: "workorderGuid", Description: "Tenant-scoped workorder GUID", Required: true, Type: "uuid"},
					{Name: "requirementGuid", Description: "Supporting-resource requirement GUID", Required: true, Type: "uuid"},
				},
				Flags: requirementFlags,
			},
			{
				Name:        "delete",
				Description: "Delete one supporting-resource requirement",
				ToolName:    "UteamupWorkorderSupportingAssetRequirementDelete",
				HTTPMethod:  "DELETE",
				RESTPath:    "{workorderGuid}/supporting-assets/{requirementGuid}",
				Args: []ArgDef{
					{Name: "workorderGuid", Description: "Tenant-scoped workorder GUID", Required: true, Type: "uuid"},
					{Name: "requirementGuid", Description: "Supporting-resource requirement GUID", Required: true, Type: "uuid"},
				},
			},
		},
	})
}
