package registry

func init() {
	Register(&Domain{
		Name:        "compliance-code",
		Description: "Manage compliance codes",
		Actions: append(crudActions("ComplianceCode"), Action{
			Name:         "list-for-entity",
			Description:  "List compliance codes assigned to an inventory entity public GUID",
			ToolName:     "UteamupComplianceCodeListForEntity",
			RESTBasePath: "/api/compliance",
			RESTPath:     "entity/{entityType}/by-guid/{entityGuid}/codes",
			HTTPMethod:   "GET",
			Args: []ArgDef{
				{Name: "entityType", Description: "asset, tool, part, or chemical", Required: true, Type: "string", AllowedValues: []string{"asset", "tool", "part", "chemical"}},
				{Name: "entityGuid", Description: "Inventory entity public GUID", Required: true, Type: "uuid"},
			},
		}),
	})
	Register(&Domain{Name: "compliance-standard", Description: "Manage compliance standards", Actions: crudActions("ComplianceStandard")})
	Register(&Domain{Name: "certificate", Aliases: []string{"certificates", "cert"}, Description: "Manage certificates", Actions: crudActions("Certificate")})
	Register(&Domain{Name: "failure-code", Description: "Manage failure codes", Actions: crudActions("FailureCode")})
}
