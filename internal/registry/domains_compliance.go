package registry

func init() {
	Register(&Domain{
		Name:        "compliance-code",
		Description: "Manage compliance codes",
		Actions: append(append(crudActions("ComplianceCode"), locationComplianceActions()...), Action{
			Name:         "list-for-entity",
			Description:  "List compliance codes assigned to an inventory entity public GUID",
			ToolName:     "UteamupComplianceCodeListForEntity",
			RESTBasePath: "/api/compliance",
			RESTPath:     "entity/{entityType}/by-guid/{entityGuid}/codes",
			HTTPMethod:   "GET",
			Args: []ArgDef{
				{Name: "entityType", Description: "asset, tool, part, chemical, or location", Required: true, Type: "string", AllowedValues: []string{"asset", "tool", "part", "chemical", "location"}},
				{Name: "entityGuid", Description: "Inventory entity public GUID", Required: true, Type: "uuid"},
			},
		}),
	})
	Register(&Domain{Name: "compliance-standard", Description: "Manage compliance standards", Actions: crudActions("ComplianceStandard")})
	Register(&Domain{Name: "certificate", Aliases: []string{"certificates", "cert"}, Description: "Manage certificates", Actions: append(crudActions("Certificate"), personalCertificateActions()...)})
	Register(&Domain{Name: "failure-code", Description: "Manage failure codes", Actions: crudActions("FailureCode")})
}
