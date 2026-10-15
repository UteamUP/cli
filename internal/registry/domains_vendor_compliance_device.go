package registry

func init() {
	Register(&Domain{
		Name:        "vendor-compliance-device",
		Description: "Deliver retained compliance originals through an enrolled device; never enrolls, signs or unlocks",
		APIPath:     "/api/v1/vendor/compliance/reviewed",
		Actions: []Action{
			vendorComplianceDeviceAction("proof", "UteamupVendorComplianceDeviceProof", "Prepare a proof for an exact retained compliance original", false),
			vendorComplianceDeviceAction("register", "UteamupVendorComplianceDeviceRegister", "Register an already signed compliance original without extending its window", false),
			vendorComplianceDeviceAction("deliver", "UteamupVendorComplianceDeviceDeliver", "Deliver the original compliance change after confirmation and return its minimal outcome", true),
		},
	})
}

func vendorComplianceDeviceAction(name, tool, description string, confirm bool) Action {
	action := vendorDeviceAction(name, tool, description, confirm)
	action.Args[0].AllowedValues = []string{"create", "update", "delete", "review"}
	return action
}
