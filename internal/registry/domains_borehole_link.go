package registry

var boreholeLinkTargetTypes = []string{"code", "asset", "location"}

func boreholeUidArg() ArgDef {
	return ArgDef{Name: "boreholeUid", Description: "UOS borehole_uid of the borehole (positive number)", Required: true, Type: "int"}
}

func boreholeLinkTargetFlags(queryString bool) []FlagDef {
	targetType := FlagDef{Name: "target-type", Description: "Record type: code, asset or location", Type: "string", Required: true, AllowedValues: boreholeLinkTargetTypes}
	targetGuid := FlagDef{Name: "target-guid", Description: "GUID of the industry code, asset or location", Type: "non-empty-uuid", Required: true}
	if queryString {
		targetType.QueryName = "targetType"
		targetGuid.QueryName = "targetGuid"
	}
	return []FlagDef{targetType, targetGuid}
}

func init() {
	Register(&Domain{
		Name:        "borehole",
		Aliases:     []string{"boreholes"},
		Description: "Link boreholes from the UOS registry (Orkustofnun) to industry codes, assets and locations",
		APIPath:     "/api/boreholes",
		Actions: []Action{
			{Name: "links", Description: "List the codes, assets and locations linked to a borehole", ToolName: "UteamupBoreholeLinkList",
				HTTPMethod: "GET", RESTPath: "{boreholeUid}/links", RejectExtraArgs: true,
				Args: []ArgDef{boreholeUidArg()}},
			{Name: "links-for-target", Description: "List the boreholes linked to a code, asset or location", ToolName: "UteamupBoreholeLinkListForTarget",
				HTTPMethod: "GET", RESTPath: "links", RejectExtraArgs: true,
				Flags: boreholeLinkTargetFlags(true)},
			{Name: "links-add", Description: "Link a borehole to a code, asset or location", ToolName: "UteamupBoreholeLinkCreate",
				HTTPMethod: "POST", RESTPath: "{boreholeUid}/links", RejectExtraArgs: true,
				Args:  []ArgDef{boreholeUidArg()},
				Flags: boreholeLinkTargetFlags(false)},
			{Name: "links-delete", Description: "Remove one borehole link (the borehole and the record are not changed)", ToolName: "UteamupBoreholeLinkDelete",
				HTTPMethod: "DELETE", RESTPath: "{boreholeUid}/links/{linkGuid}", RejectExtraArgs: true,
				Args: []ArgDef{boreholeUidArg(), {Name: "linkGuid", Description: "GUID of the link returned by links", Required: true, Type: "non-empty-uuid"}}},
		},
	})
}
