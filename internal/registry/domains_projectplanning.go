package registry

// Project planning domains: stages, output items, and the budget summary.
//
// All three controllers route under /api/projects (plural — like
// ProjectCopilotController), NOT the /api/project base the `project` domain
// auto-derives, hence dedicated domains with an explicit APIPath. The
// runtime in registry.go calls apiClient.CallREST(...), so action Name +
// RESTPath build the URL; positional arg names must literally match the
// `{placeholder}` tokens.
//
// REST surface (Guid-keyed):
//
//	GET    /api/projects/{projectGuid}/stages                       — list stages
//	GET    /api/projects/{projectGuid}/stages/{stageGuid}           — fetch one stage
//	POST   /api/projects/{projectGuid}/stages                       — create stage
//	PUT    /api/projects/{projectGuid}/stages/{stageGuid}           — full update (incl. status)
//	DELETE /api/projects/{projectGuid}/stages/{stageGuid}           — delete stage
//	POST   /api/projects/{projectGuid}/stages/{stageGuid}/advance   — advance past the stage gate
//	PUT    /api/projects/{projectGuid}/stages/reorder               — reorder (body: stageGuids)
//	PUT    /api/projects/{projectGuid}/stages/{stageGuid}/status    — status-only transition
//
//	GET    /api/projects/{projectGuid}/outputitems                    — list output items
//	GET    /api/projects/{projectGuid}/outputitems/{itemGuid}         — fetch one item
//	POST   /api/projects/{projectGuid}/outputitems                    — create item
//	PUT    /api/projects/{projectGuid}/outputitems/{itemGuid}         — full update
//	DELETE /api/projects/{projectGuid}/outputitems/{itemGuid}         — delete item
//	POST   /api/projects/{projectGuid}/outputitems/{itemGuid}/deliver — mark delivered
//
//	GET    /api/projects/{projectGuid}/budget                       — budget summary
func init() {
	Register(&Domain{
		Name:        "project-stage",
		Aliases:     []string{"project-stages", "stages"},
		Description: "Manage project stages: gated phases with ordering, milestones, and status transitions",
		APIPath:     "/api/projects",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List the stages of a project in order",
				ToolName:    "UteamupProjectStageList",
				RESTPath:    "{projectGuid}/stages",
				Args:        []ArgDef{{Name: "projectGuid", Description: "Project GUID", Required: true, Type: "string"}},
			},
			{
				Name:        "get",
				Description: "Get a single project stage by GUID",
				ToolName:    "UteamupProjectStageGet",
				RESTPath:    "{projectGuid}/stages/{stageGuid}",
				Args: []ArgDef{
					{Name: "projectGuid", Description: "Project GUID", Required: true, Type: "string"},
					{Name: "stageGuid", Description: "Stage GUID", Required: true, Type: "string"},
				},
			},
			{
				Name:        "create",
				Description: "Create a stage on a project",
				ToolName:    "UteamupProjectStageCreate",
				RESTPath:    "{projectGuid}/stages",
				Args:        []ArgDef{{Name: "projectGuid", Description: "Project GUID", Required: true, Type: "string"}},
				Flags: []FlagDef{
					{Name: "name", Description: "Stage display name", Required: true, Type: "string"},
					{Name: "order", Description: "Ordering position (1 = first)", Required: true, Type: "int"},
					{Name: "gate-criteria-json", Description: "Optional JSON-encoded gate criteria that must be met before advancing past this stage", Type: "string"},
					{Name: "start-date", Description: "Optional planned start date (ISO 8601, e.g. 2026-07-01)", Type: "string"},
					{Name: "due-date", Description: "Optional planned due date / milestone (ISO 8601)", Type: "string"},
				},
			},
			{
				Name:        "update",
				Description: "Full update of a stage (PUT — supply every field you want kept)",
				ToolName:    "UteamupProjectStageUpdate",
				RESTPath:    "{projectGuid}/stages/{stageGuid}",
				Args: []ArgDef{
					{Name: "projectGuid", Description: "Project GUID", Required: true, Type: "string"},
					{Name: "stageGuid", Description: "Stage GUID", Required: true, Type: "string"},
				},
				Flags: []FlagDef{
					{Name: "name", Description: "Stage display name", Required: true, Type: "string"},
					{Name: "order", Description: "Ordering position (1 = first)", Required: true, Type: "int"},
					{Name: "status", Description: "Stage status: NotStarted, InProgress, Completed, Blocked, or Skipped", Required: true, Type: "string"},
					{Name: "gate-criteria-json", Description: "Optional JSON-encoded gate criteria", Type: "string"},
					{Name: "start-date", Description: "Optional planned start date (ISO 8601)", Type: "string"},
					{Name: "due-date", Description: "Optional planned due date / milestone (ISO 8601)", Type: "string"},
				},
			},
			{
				Name:        "delete",
				Description: "Delete a stage from a project",
				ToolName:    "UteamupProjectStageDelete",
				RESTPath:    "{projectGuid}/stages/{stageGuid}",
				Args: []ArgDef{
					{Name: "projectGuid", Description: "Project GUID", Required: true, Type: "string"},
					{Name: "stageGuid", Description: "Stage GUID", Required: true, Type: "string"},
				},
			},
			{
				Name:        "advance",
				Description: "Advance the project past this stage's gate (marks it complete and moves on)",
				ToolName:    "UteamupProjectStageAdvance",
				HTTPMethod:  "POST",
				RESTPath:    "{projectGuid}/stages/{stageGuid}/advance",
				Args: []ArgDef{
					{Name: "projectGuid", Description: "Project GUID", Required: true, Type: "string"},
					{Name: "stageGuid", Description: "Stage GUID", Required: true, Type: "string"},
				},
			},
			{
				Name:        "reorder",
				Description: "Reorder a project's stages — pass exactly the project's stage GUIDs in the desired order",
				ToolName:    "UteamupProjectStageReorder",
				HTTPMethod:  "PUT",
				RESTPath:    "{projectGuid}/stages/reorder",
				Args:        []ArgDef{{Name: "projectGuid", Description: "Project GUID", Required: true, Type: "string"}},
				Flags: []FlagDef{
					{Name: "stage-guids", Description: "Stage GUIDs in the desired order (comma-separated or repeated)", Required: true, Type: "stringSlice"},
				},
			},
			{
				Name:        "set-status",
				Description: "Status-only transition for a stage (does not touch name/order/dates)",
				ToolName:    "UteamupProjectStageSetStatus",
				HTTPMethod:  "PUT",
				RESTPath:    "{projectGuid}/stages/{stageGuid}/status",
				Args: []ArgDef{
					{Name: "projectGuid", Description: "Project GUID", Required: true, Type: "string"},
					{Name: "stageGuid", Description: "Stage GUID", Required: true, Type: "string"},
				},
				Flags: []FlagDef{
					{Name: "status", Description: "New status: NotStarted, InProgress, Completed, Blocked, or Skipped", Required: true, Type: "string"},
				},
			},
		},
	})

	Register(&Domain{
		Name:        "project-output",
		Aliases:     []string{"project-outputs", "output-items"},
		Description: "Manage project output items (deliverables) with expected/actual quantities and delivery state",
		APIPath:     "/api/projects",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List the output items of a project",
				ToolName:    "UteamupProjectOutputItemList",
				RESTPath:    "{projectGuid}/outputitems",
				Args:        []ArgDef{{Name: "projectGuid", Description: "Project GUID", Required: true, Type: "string"}},
			},
			{
				Name:        "get",
				Description: "Get a single output item by GUID",
				ToolName:    "UteamupProjectOutputItemGet",
				RESTPath:    "{projectGuid}/outputitems/{itemGuid}",
				Args: []ArgDef{
					{Name: "projectGuid", Description: "Project GUID", Required: true, Type: "string"},
					{Name: "itemGuid", Description: "Output item GUID", Required: true, Type: "string"},
				},
			},
			{
				Name:        "create",
				Description: "Create an output item on a project",
				ToolName:    "UteamupProjectOutputItemCreate",
				RESTPath:    "{projectGuid}/outputitems",
				Args:        []ArgDef{{Name: "projectGuid", Description: "Project GUID", Required: true, Type: "string"}},
				Flags: []FlagDef{
					{Name: "description", Description: "Description of this output/deliverable", Required: true, Type: "string"},
					{Name: "expected-quantity", Description: "Expected quantity of this output", Required: true, Type: "float"},
					{Name: "customer-guid", Description: "Optional customer GUID to whom this output is delivered", Type: "string"},
				},
			},
			{
				Name:        "update",
				Description: "Full update of an output item (PUT — supply every field you want kept)",
				ToolName:    "UteamupProjectOutputItemUpdate",
				RESTPath:    "{projectGuid}/outputitems/{itemGuid}",
				Args: []ArgDef{
					{Name: "projectGuid", Description: "Project GUID", Required: true, Type: "string"},
					{Name: "itemGuid", Description: "Output item GUID", Required: true, Type: "string"},
				},
				Flags: []FlagDef{
					{Name: "description", Description: "Description of this output/deliverable", Required: true, Type: "string"},
					{Name: "expected-quantity", Description: "Expected quantity of this output", Required: true, Type: "float"},
					// Defaults keep the PUT body deterministic: an omitted flag still
					// sends 0.0 / false, matching the backend's non-nullable DTO fields.
					{Name: "actual-quantity", Description: "Actual quantity produced/delivered so far", Default: 0.0, Type: "float"},
					{Name: "is-delivered", Description: "Whether this output has been fully delivered", Default: false, Type: "bool"},
					{Name: "customer-guid", Description: "Optional customer GUID to whom this output is delivered", Type: "string"},
				},
			},
			{
				Name:        "delete",
				Description: "Delete an output item from a project",
				ToolName:    "UteamupProjectOutputItemDelete",
				RESTPath:    "{projectGuid}/outputitems/{itemGuid}",
				Args: []ArgDef{
					{Name: "projectGuid", Description: "Project GUID", Required: true, Type: "string"},
					{Name: "itemGuid", Description: "Output item GUID", Required: true, Type: "string"},
				},
			},
			{
				Name:        "deliver",
				Description: "Mark an output item as delivered (optionally recording the final quantity)",
				ToolName:    "UteamupProjectOutputItemDeliver",
				HTTPMethod:  "POST",
				RESTPath:    "{projectGuid}/outputitems/{itemGuid}/deliver",
				Args: []ArgDef{
					{Name: "projectGuid", Description: "Project GUID", Required: true, Type: "string"},
					{Name: "itemGuid", Description: "Output item GUID", Required: true, Type: "string"},
				},
				Flags: []FlagDef{
					// No Default on purpose: when omitted the field stays out of the
					// body and the backend keeps the current actual quantity.
					{Name: "actual-quantity", Description: "Optional final delivered quantity; when omitted the current actual quantity is kept", Type: "float"},
				},
			},
		},
	})

	Register(&Domain{
		Name:        "project-budget",
		Aliases:     []string{"budget"},
		Description: "Project budget summary: budget, estimated cost, computed actual cost, variance, utilisation",
		APIPath:     "/api/projects",
		Actions: []Action{
			{
				Name:        "get",
				Description: "Get the budget summary for a project",
				ToolName:    "UteamupProjectGetBudget",
				RESTPath:    "{projectGuid}/budget",
				Args:        []ArgDef{{Name: "projectGuid", Description: "Project GUID", Required: true, Type: "string"}},
			},
		},
	})
}
