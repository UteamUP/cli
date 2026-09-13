package registry

func init() {
	stageArgs := projectResourceArguments("stageGuid", "Project stage GUID")
	Register(&Domain{
		Name: "project-gate", Description: "Review exact scope and evidence before recording gate decisions", APIPath: "/api/projects",
		Actions: []Action{
			{Name: "preview", Description: "Preview scoped criteria, exact sources, conditions and named signers without saving",
				ToolName: "UteamupProjectGateSubmissionPreview", HTTPMethod: "POST", RESTPath: "{projectGuid}/stages/{stageGuid}/decisions/submissions/preview",
				Args: stageArgs, Flags: []FlagDef{jsonFlag()}},
			{Name: "submit", Description: "Retain reviewed requestGuid and expectedReviewFingerprint; sends no messages and releases no work",
				ToolName: "UteamupProjectGateSubmissionSubmit", HTTPMethod: "POST", RESTPath: "{projectGuid}/stages/{stageGuid}/decisions/submissions",
				Args: stageArgs, Flags: []FlagDef{jsonFlag()}},
			{Name: "get", Description: "Read an exact immutable gate review and its controlled signer requirements",
				ToolName: "UteamupProjectGateSubmissionGet", HTTPMethod: "GET", RESTPath: "{projectGuid}/stages/{stageGuid}/decisions/submissions/{submissionGuid}",
				Args: []ArgDef{{Name: "projectGuid", Description: "Project GUID", Required: true}, {Name: "stageGuid", Description: "Stage GUID", Required: true},
					{Name: "submissionGuid", Description: "Exact submission GUID", Required: true}}},
			{Name: "history", Description: "Page submission headers; open a specific submission for evidence and signatures",
				ToolName: "UteamupProjectGateSubmissionList", HTTPMethod: "GET", RESTPath: "{projectGuid}/stages/{stageGuid}/decisions/submissions",
				Args: stageArgs, Flags: []FlagDef{{Name: "page", Description: "One-based page", Type: "int", Default: 1},
					{Name: "page-size", Description: "Rows per page, 1 to 100", Type: "int", Default: 20}}},
			{Name: "decisions", Description: "Read retained decisions for this stage",
				ToolName: "UteamupProjectGateDecisionList", HTTPMethod: "GET", RESTPath: "{projectGuid}/stages/{stageGuid}/decisions", Args: stageArgs},
			{Name: "decide", Description: "Record the exact reviewed outcome; adopted rules require requestGuid, submissionGuid and expectedReviewFingerprint",
				ToolName: "UteamupProjectGateDecisionRecord", HTTPMethod: "POST", RESTPath: "{projectGuid}/stages/{stageGuid}/decisions",
				Args: stageArgs, Flags: []FlagDef{jsonFlag()}},
		},
	})
}
