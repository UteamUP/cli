package registry

// Workforce time and training: training events, working-time rule sets, travel
// expenses and technician shift availability, routed through the MCP tools that
// mirror the REST controllers (same permissions, GUID-only identifiers).
func init() {
	eventGuid := ArgDef{Name: "event-guid", BodyName: "eventGuid", Description: "Training event public GUID", Required: true, Type: "non-empty-uuid"}
	ruleSetGuid := ArgDef{Name: "rule-set-guid", BodyName: "ruleSetGuid", Description: "Working-time rule set public GUID", Required: true, Type: "non-empty-uuid"}
	userGuid := ArgDef{Name: "user-guid", BodyName: "userGuid", Description: "Technician user public GUID", Required: true, Type: "non-empty-uuid"}

	Register(&Domain{
		Name:        "workforce-time",
		Aliases:     []string{"training-events"},
		Description: "Training events, working-time rules, travel expenses and technician shifts",
		Actions: []Action{
			{Name: "training-list", Description: "List training events in the tenant", ToolName: "UteamupTrainingEventList", MCPOnly: true, RejectExtraArgs: true},
			{Name: "training-mine", Description: "List your own training plan", ToolName: "UteamupTrainingEventMine", MCPOnly: true, RejectExtraArgs: true},
			{Name: "training-get", Description: "Get one training event", ToolName: "UteamupTrainingEventGet", MCPOnly: true, RejectExtraArgs: true,
				Args: []ArgDef{eventGuid}},
			{Name: "training-create", Description: "Create a training event", ToolName: "UteamupTrainingEventCreate", MCPOnly: true, RejectExtraArgs: true,
				Flags: []FlagDef{
					{Name: "topic", Description: "Training topic", Required: true, Type: "string"},
					{Name: "start-date", BodyName: "startDate", Description: "Start date-time (ISO 8601, UTC)", Required: true, Type: "string"},
					{Name: "end-date", BodyName: "endDate", Description: "End date-time (ISO 8601, UTC)", Type: "string"},
					{Name: "description", Description: "Description", Type: "string"},
					{Name: "instructor-user-guid", BodyName: "instructorUserGuid", Description: "Instructor user GUID", Type: "uuid"},
					{Name: "instructor-external-name", BodyName: "instructorExternalName", Description: "External instructor name", Type: "string"},
					{Name: "location-guid", BodyName: "locationGuid", Description: "Location GUID", Type: "uuid"},
					{Name: "max-participants", BodyName: "maxParticipants", Description: "Maximum participants", Type: "int"},
					{Name: "status", Description: "Scheduled, InProgress, Completed or Cancelled", Type: "string"},
					{Name: "required-certificate-guid", BodyName: "requiredCertificateGuid", Description: "Required certificate GUID", Type: "uuid"},
					{Name: "category-guid", BodyName: "categoryGuid", Description: "Training category GUID", Type: "uuid"},
				}},
			{Name: "training-update", Description: "Update a training event", ToolName: "UteamupTrainingEventUpdate", MCPOnly: true, RejectExtraArgs: true,
				Args: []ArgDef{eventGuid},
				Flags: []FlagDef{
					{Name: "topic", Description: "New topic", Type: "string"},
					{Name: "start-date", BodyName: "startDate", Description: "New start date-time (ISO 8601, UTC)", Type: "string"},
					{Name: "end-date", BodyName: "endDate", Description: "New end date-time (ISO 8601, UTC)", Type: "string"},
					{Name: "description", Description: "New description", Type: "string"},
					{Name: "location-guid", BodyName: "locationGuid", Description: "New location GUID", Type: "uuid"},
					{Name: "max-participants", BodyName: "maxParticipants", Description: "New maximum participants", Type: "int"},
					{Name: "status", Description: "Scheduled, InProgress, Completed or Cancelled", Type: "string"},
				}},
			{Name: "training-request-enrollment", Description: "Request enrollment in a training event for yourself", ToolName: "UteamupTrainingEventRequestEnrollment", MCPOnly: true, RejectExtraArgs: true,
				Args: []ArgDef{eventGuid}},
			{Name: "training-enrollments", Description: "List enrollments of a training event", ToolName: "UteamupTrainingEventEnrollmentList", MCPOnly: true, RejectExtraArgs: true,
				Args: []ArgDef{eventGuid}},
			{Name: "training-completions", Description: "List completions of a training event", ToolName: "UteamupTrainingEventCompletionList", MCPOnly: true, RejectExtraArgs: true,
				Args: []ArgDef{eventGuid}},
			{Name: "ruleset-list", Description: "List working-time rule sets", ToolName: "UteamupWorkingTimeRuleSetList", MCPOnly: true, RejectExtraArgs: true},
			{Name: "ruleset-get", Description: "Get one working-time rule set", ToolName: "UteamupWorkingTimeRuleSetGet", MCPOnly: true, RejectExtraArgs: true,
				Args: []ArgDef{ruleSetGuid}},
			{Name: "ruleset-create", Description: "Create a working-time rule set", ToolName: "UteamupWorkingTimeRuleSetCreate", MCPOnly: true, RejectExtraArgs: true,
				Flags: []FlagDef{
					{Name: "name", Description: "Rule set name", Required: true, Type: "string"},
					{Name: "country", BodyName: "countryCode", Description: "ISO country code (e.g. IS, US)", Type: "string"},
				}},
			{Name: "ruleset-update", Description: "Update a working-time rule set", ToolName: "UteamupWorkingTimeRuleSetUpdate", MCPOnly: true, RejectExtraArgs: true,
				Args: []ArgDef{ruleSetGuid},
				Flags: []FlagDef{
					{Name: "name", Description: "Rule set name", Required: true, Type: "string"},
					{Name: "country", BodyName: "countryCode", Description: "ISO country code (e.g. IS, US)", Type: "string"},
					{Name: "is-active", BodyName: "isActive", Description: "Whether the rule set is active", Type: "bool", Default: true},
				}},
			{Name: "policy-assignments", Description: "List working-time policy assignments", ToolName: "UteamupWorkingTimePolicyAssignmentList", MCPOnly: true, RejectExtraArgs: true},
			{Name: "expense-list", Description: "List travel expenses", ToolName: "UteamupTravelExpenseList", MCPOnly: true, RejectExtraArgs: true,
				Flags: []FlagDef{
					{Name: "page", Description: "Page number", Type: "int", Default: 1},
					{Name: "page-size", BodyName: "pageSize", Description: "Page size (max 100)", Type: "int", Default: 25},
				}},
			{Name: "expense-get", Description: "Get one travel expense", ToolName: "UteamupTravelExpenseGet", MCPOnly: true, RejectExtraArgs: true,
				Args: []ArgDef{{Name: "expense-guid", BodyName: "expenseGuid", Description: "Travel expense public GUID", Required: true, Type: "non-empty-uuid"}}},
			{Name: "shift-schedule", Description: "Get a technician's shift schedule for a date range", ToolName: "UteamupTechnicianAvailabilitySchedule", MCPOnly: true, RejectExtraArgs: true,
				Args: []ArgDef{userGuid},
				Flags: []FlagDef{
					{Name: "start-date", BodyName: "startDate", Description: "First date (ISO 8601)", Required: true, Type: "string"},
					{Name: "end-date", BodyName: "endDate", Description: "Last date (ISO 8601)", Required: true, Type: "string"},
				}},
			{Name: "current-shift", Description: "Get a technician's current shift", ToolName: "UteamupTechnicianAvailabilityCurrent", MCPOnly: true, RejectExtraArgs: true,
				Args: []ArgDef{userGuid}},
		},
	})
}
