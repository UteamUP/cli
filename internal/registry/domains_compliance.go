package registry

func init() {
	Register(&Domain{Name: "compliance-code", Description: "Manage compliance codes", Actions: crudActions("ComplianceCode")})
	Register(&Domain{Name: "compliance-standard", Description: "Manage compliance standards", Actions: crudActions("ComplianceStandard")})
	Register(&Domain{Name: "certificate", Aliases: []string{"certificates", "cert"}, Description: "Manage certificates", Actions: crudActions("Certificate")})
	Register(&Domain{Name: "failure-code", Description: "Manage failure codes", Actions: crudActions("FailureCode")})
	Register(&Domain{Name: "root-cause", Aliases: []string{"rca"}, Description: "Manage root cause analyses", Actions: crudActions("RootCauseAnalysis")})
}
