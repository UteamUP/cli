package registry

func init() {
	Register(&Domain{Name: "contract", Aliases: []string{"contracts"}, Description: "Manage contracts", Actions: crudActions("Contract")})
	Register(&Domain{Name: "contractor", Aliases: []string{"contractors"}, Description: "Manage contractor profiles", Actions: crudActions("ContractorProfile")})
	Register(&Domain{Name: "contractor-workorder", Description: "Manage contractor work orders", Actions: crudActions("ContractorWorkOrder")})
	Register(&Domain{Name: "labour-rate", Description: "Manage labour rates", Actions: crudActions("LabourRate")})
	Register(&Domain{Name: "rental-rate", Description: "Manage rental rates", Actions: crudActions("RentalRate")})
	Register(&Domain{Name: "warranty", Aliases: []string{"warranties"}, Description: "Manage warranties", Actions: crudActions("Warranty")})
	Register(&Domain{Name: "commission", Aliases: []string{"commissions"}, Description: "Manage commissions", Actions: crudActions("Commission")})
}
