package registry

func init() {
	Register(&Domain{
		Name:        "upmate-field-service",
		Aliases:     []string{"field-service-assist"},
		Description: "Create governed, tenant-scoped field-service previews and explanations",
		Actions: []Action{
			{
				Name:        "schedule-preview",
				Description: "Create a durable review-only schedule proposal without applying it",
				ToolName:    "UteamupUpmateScheduleOptimizationPreview",
				MCPOnly:     true,
				Flags: []FlagDef{
					upmateFieldServiceGUIDFlag("idempotency-key", "idempotencyKey", "Retry-stable idempotency GUID", true),
					{Name: "week-start-utc", BodyName: "weekStartUtc", Description: "UTC start of the planning week", Required: true, Type: "string"},
					{Name: "workorder-guids", BodyName: "workorderGuids", Description: "Tenant work-order GUIDs, at most 500", Required: true, Type: "stringSlice"},
					{Name: "technician-guids", BodyName: "technicianGuids", Description: "Eligible tenant-member GUIDs, at most 100", Required: true, Type: "stringSlice"},
					upmateFieldServiceGUIDFlag("team-guid", "teamGuid", "Optional tenant team GUID", false),
					{Name: "horizon-days", BodyName: "horizonDays", Description: "Planning horizon in days, 1-30", Type: "int", Default: 7},
					upmateFieldServiceWeightFlag("competency-weight", "competencyWeight"),
					upmateFieldServiceWeightFlag("travel-weight", "travelWeight"),
					upmateFieldServiceWeightFlag("lateness-weight", "latenessWeight"),
					upmateFieldServiceWeightFlag("overtime-weight", "overtimeWeight"),
					upmateFieldServiceWeightFlag("workload-balance-weight", "workloadBalanceWeight"),
					upmateFieldServiceWeightFlag("continuity-weight", "continuityWeight"),
					upmateFieldServiceWeightFlag("repeat-visit-weight", "repeatVisitWeight"),
					upmateFieldServiceWeightFlag("disruption-weight", "disruptionWeight"),
				},
			},
			{
				Name:        "schedule-explain",
				Description: "Explain stored schedule constraints without changing assignments",
				ToolName:    "UteamupUpmateScheduleOptimizationExplain",
				MCPOnly:     true,
				Args:        upmateFieldServiceGUIDArg("runGuid", "Schedule optimization run GUID"),
			},
			{
				Name:        "maintenance-suggest",
				Description: "Create a maintenance preview without creating work orders",
				ToolName:    "UteamupUpmateMaintenancePlanSuggest",
				MCPOnly:     true,
				Flags: []FlagDef{
					upmateFieldServiceGUIDFlag("idempotency-key", "idempotencyKey", "Retry-stable idempotency GUID", true),
					{Name: "plan-guids", BodyName: "planGuids", Description: "Maintenance-plan GUIDs, at most 100", Required: true, Type: "stringSlice"},
					{Name: "as-of-utc", BodyName: "asOfUtc", Description: "Optional UTC projection point", Type: "string"},
					{Name: "horizon-days", BodyName: "horizonDays", Description: "Suggestion horizon in days, 1-365", Type: "int", Default: 30},
				},
			},
			{
				Name:        "maintenance-due-explain",
				Description: "Explain deterministic maintenance due evidence without creating work",
				ToolName:    "UteamupUpmateMaintenanceDueExplain",
				MCPOnly:     true,
				Args:        upmateFieldServiceGUIDArg("planGuid", "Maintenance-plan GUID"),
				Flags: []FlagDef{
					{Name: "as-of-utc", BodyName: "asOfUtc", Description: "UTC evidence projection point", Required: true, Type: "string"},
				},
			},
			{
				Name:        "fieldnote-transcribe",
				Description: "Read a stored transcript while treating audio as non-executable data",
				ToolName:    "UteamupUpmateFieldnoteTranscribe",
				MCPOnly:     true,
				Args:        upmateFieldServiceGUIDArg("audioFileGuid", "Stored tenant audio-file GUID"),
			},
			{
				Name:        "portal-request-classify",
				Description: "Draft an expiring classification that is never saved or auto-dispatched",
				ToolName:    "UteamupUpmatePortalRequestClassify",
				MCPOnly:     true,
				Args:        upmateFieldServiceGUIDArg("workRequestGuid", "Customer portal work-request GUID"),
				Flags: []FlagDef{
					upmateFieldServiceGUIDFlag("idempotency-key", "idempotencyKey", "Retry-stable idempotency GUID", true),
					{Name: "language", Description: "Response language code", Type: "string", Default: "en"},
				},
			},
			{
				Name:        "service-billing-review",
				Description: "Review exact billing evidence without changing invoices",
				ToolName:    "UteamupUpmateServiceBillingReview",
				MCPOnly:     true,
				Args:        upmateFieldServiceGUIDArg("runGuid", "Service-billing run GUID"),
			},
		},
	})
}

func upmateFieldServiceGUIDArg(name string, description string) []ArgDef {
	return []ArgDef{{
		Name:        name,
		Description: description,
		Required:    true,
		Type:        "uuid",
	}}
}

func upmateFieldServiceGUIDFlag(
	name string,
	bodyName string,
	description string,
	required bool,
) FlagDef {
	return FlagDef{
		Name:        name,
		BodyName:    bodyName,
		Description: description,
		Required:    required,
		Type:        "uuid",
	}
}

func upmateFieldServiceWeightFlag(name string, bodyName string) FlagDef {
	return FlagDef{
		Name:        name,
		BodyName:    bodyName,
		Description: "Objective weight, 0-10",
		Type:        "float",
		Default:     1.0,
	}
}
