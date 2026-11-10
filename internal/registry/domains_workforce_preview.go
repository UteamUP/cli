package registry

func workforcePreviewTimeActions() []Action {
	entry := []ArgDef{{Name: "externalGuid", Description: "Time-entry public GUID", Required: true, Type: "uuid"}}
	version := FlagDef{Name: "expected-updated-at", BodyName: "expectedUpdatedAt", Description: "Exact reviewed entry timestamp", Required: true, Type: "string"}
	result := []Action{
		{Name: "review-list", Description: "Preview: page individual time entries for review", ToolName: "UteamupTimeEntryReviewList", HTTPMethod: "GET", RESTBasePath: "/api/timeentry", RESTPath: "review-queue", Flags: []FlagDef{
			{Name: "from", Type: "string"}, {Name: "to", Type: "string"}, {Name: "search", Type: "string"},
			{Name: "status", Type: "string"}, {Name: "user-guid", BodyName: "userGuid", Type: "string"},
			{Name: "page", Default: 1, Type: "int"}, {Name: "page-size", BodyName: "pageSize", Default: 25, Type: "int"},
		}},
		{Name: "review-context", Description: "Preview: review work context and 100% cost attribution without charging", ToolName: "UteamupTimeEntryReviewContext", HTTPMethod: "PUT", RESTBasePath: "/api/timeentry", RESTPath: "by-guid/{externalGuid}/review-context", Args: entry, Flags: []FlagDef{jsonFlag(), timeEntryIdempotencyFlag(), version}},
		{Name: "cost-report", Description: "Preview: approved time by cost target; external systems own invoicing", ToolName: "UteamupTimeEntryCostReport", HTTPMethod: "GET", RESTBasePath: "/api/timeentry", RESTPath: "cost-report", Flags: []FlagDef{
			{Name: "from", Required: true, Type: "string"}, {Name: "to", Required: true, Type: "string"}, {Name: "dimension", Default: "target", Type: "string"},
		}},
	}
	for _, decision := range []struct{ name, path string }{{"approve-entry", "approve"}, {"reject-entry", "reject"}, {"request-info", "request-info"}} {
		flags := []FlagDef{timeEntryIdempotencyFlag(), version}
		if decision.path != "approve" {
			flags = append(flags, FlagDef{Name: "reason", Description: "Useful message visible to the worker", Required: true, Type: "string"})
		}
		result = append(result, Action{Name: decision.name, Description: "Preview: decide only the reviewed time entry", ToolName: "UteamupTimeEntryDecide", HTTPMethod: "POST", RESTBasePath: "/api/timeentry", RESTPath: "by-guid/{externalGuid}/" + decision.path, Args: entry, Flags: flags, SensitiveBodyFields: []string{"reason", "message", "reply"}})
	}
	return result
}

func init() {
	Register(&Domain{Name: "workforce-absence", Description: "Preview: roster-based absence requests and period reports", APIPath: "/api/availabilityoverride", Actions: []Action{
		{Name: "report", Description: "Filter inclusive dates; totals retain unknown rosters and planned time", ToolName: "UteamupWorkforceAbsenceReport", HTTPMethod: "GET", RESTPath: "report", Flags: []FlagDef{
			{Name: "from", Required: true, Type: "string"}, {Name: "to", Required: true, Type: "string"},
			{Name: "user-guid", BodyName: "userGuid", Type: "string"}, {Name: "group-guid", BodyName: "groupGuid", Type: "string"},
			{Name: "location-guid", BodyName: "locationGuid", Type: "string"}, {Name: "category", Type: "string"},
			{Name: "page", Default: 1, Type: "int"}, {Name: "page-size", BodyName: "pageSize", Default: 50, Type: "int"},
		}},
		{Name: "create", Description: "Request absence using [startDate,endDate) and optional minutesPerDay", ToolName: "UteamupWorkforceAbsenceCreate", HTTPMethod: "POST", UseDomainBasePath: true, Flags: []FlagDef{jsonFlag()}, SensitiveBodyFields: []string{"reason"}},
		{Name: "approve", Description: "Approve the entire observed absence record", ToolName: "UteamupWorkforceAbsenceDecide", HTTPMethod: "POST", RESTPath: "by-guid/{overrideGuid}/approve", Args: []ArgDef{{Name: "overrideGuid", Required: true, Type: "uuid"}}, Flags: []FlagDef{{Name: "expected-updated-at", BodyName: "expectedUpdatedAt", Required: true, Type: "string"}}},
		{Name: "reject", Description: "Reject the entire observed absence record", ToolName: "UteamupWorkforceAbsenceDecide", HTTPMethod: "POST", RESTPath: "by-guid/{overrideGuid}/reject", Args: []ArgDef{{Name: "overrideGuid", Required: true, Type: "uuid"}}, Flags: []FlagDef{{Name: "expected-updated-at", BodyName: "expectedUpdatedAt", Required: true, Type: "string"}}},
	}})
	Register(&Domain{Name: "worker-presence", Description: "Preview: voluntary work modes, independent of GPS or absence", APIPath: "/api/workerprofiles", Actions: []Action{
		{Name: "list", ToolName: "UteamupWorkerPresenceList", HTTPMethod: "GET", RESTPath: "presence", Flags: []FlagDef{{Name: "page", Default: 1, Type: "int"}, {Name: "page-size", BodyName: "pageSize", Default: 50, Type: "int"}, {Name: "search", Type: "string"}, {Name: "mode", Type: "string"}}},
		{Name: "mine", ToolName: "UteamupWorkerPresenceMine", HTTPMethod: "GET", RESTPath: "me/presence"},
		{Name: "declare", ToolName: "UteamupWorkerPresenceDeclare", HTTPMethod: "PUT", RESTPath: "me/presence", Flags: []FlagDef{jsonFlag()}},
		{Name: "manager-position", Description: "Current consented assignment point only; recipient access is audited", ToolName: "UteamupWorkforceManagerPosition", HTTPMethod: "GET", RESTBasePath: "/api/assignmenttracking", RESTPath: "assignments/{assignmentGuid}/manager-position", Args: []ArgDef{{Name: "assignmentGuid", Required: true, Type: "uuid"}}},
	}})
	Register(&Domain{Name: "payroll-preparation", Description: "Preview: approved timesheets prepared for manager review", APIPath: "/api/payrollexports", Actions: []Action{
		{Name: "list", ToolName: "UteamupPayrollPreparationList", HTTPMethod: "GET", RESTPath: "by-guid", Flags: []FlagDef{{Name: "page", Default: 1, Type: "int"}, {Name: "page-size", BodyName: "pageSize", Default: 25, Type: "int"}}},
		{Name: "timesheet-choices", ToolName: "UteamupPayrollTimesheetChoices", HTTPMethod: "GET", RESTPath: "timesheet-choices", Flags: []FlagDef{{Name: "page", Default: 1, Type: "int"}, {Name: "page-size", BodyName: "pageSize", Default: 25, Type: "int"}, {Name: "search", Type: "string"}}},
		{Name: "create", Description: "Create PendingReview preparation from approved public timesheet GUIDs; no payment or submission", ToolName: "UteamupPayrollPreparationCreate", HTTPMethod: "POST", RESTPath: "by-guid", Flags: []FlagDef{jsonFlag()}, SensitiveBodyFields: []string{"fieldMappingJson"}},
	}})
}
