package registry

func commentActions() []Action {
	comment := []ArgDef{{Name: "commentGuid", Description: "Comment public GUID", Type: "non-empty-uuid", Required: true}}
	part := ArgDef{Name: "partGuid", Description: "Part public GUID", Type: "non-empty-uuid", Required: true}
	fields := jsonFlag()
	fields.Required = true
	return []Action{
		{Name: "list", Description: "List comments on one workorder by its public GUID", ToolName: "UteamupCommentListByWorkorder", HTTPMethod: "GET", RESTPath: "workorder/by-guid/{workorderGuid}", Args: []ArgDef{{Name: "workorderGuid", Description: "Workorder public GUID", Type: "non-empty-uuid", Required: true}}},
		{Name: "get", Description: "Read a comment by its public GUID", ToolName: "UteamupCommentGet", HTTPMethod: "GET", RESTPath: "by-guid/{commentGuid}", Args: comment},
		{Name: "create", Description: "Create a comment with text and exactly one workorderGuid or partGuid", ToolName: "UteamupCommentCreate", HTTPMethod: "POST", Flags: []FlagDef{fields}},
		{Name: "delete", Description: "Remove a comment; part comments require current authorship", ToolName: "UteamupCommentDelete", HTTPMethod: "DELETE", RESTPath: "by-guid/{commentGuid}", Args: comment},
		{Name: "list-by-part", Description: "Read a bounded page of part comments (Part.View and Comment.View)", ToolName: "UteamupCommentListByPartPaginated", HTTPMethod: "GET", RESTPath: "part/{partGuid}/paginated", Args: []ArgDef{part}, Flags: []FlagDef{
			{Name: "page", Description: "Page number, starting at 1", Type: "int", Default: 1},
			{Name: "page-size", Description: "Page size, 1 to 100", QueryName: "pageSize", Type: "int", Default: 20},
		}},
		{Name: "create-for-part", Description: "Add a text comment without mentions/media to one part", ToolName: "UteamupCommentCreate", HTTPMethod: "POST", RESTPath: "part/{partGuid}", Args: []ArgDef{part}, Flags: []FlagDef{fields}},
		{Name: "update-for-part", Description: "Replace your own part comment text (Part.View and Comment.Update)", ToolName: "UteamupCommentUpdateForPart", HTTPMethod: "PUT", RESTPath: "part/{partGuid}/{commentGuid}", Args: append([]ArgDef{part}, comment...), Flags: []FlagDef{fields}},
	}
}
