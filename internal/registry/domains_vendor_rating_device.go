package registry

func init() {
	Register(&Domain{
		Name:        "vendor-rating-device",
		Description: "Deliver retained rating originals through an enrolled device; never enrolls, signs or unlocks",
		APIPath:     "/api/v1/vendorratings/reviewed",
		Actions: []Action{
			vendorRatingDeviceAction("proof", "UteamupVendorRatingDeviceProof", "Prepare a proof for an exact retained rating original", false),
			vendorRatingDeviceAction("register", "UteamupVendorRatingDeviceRegister", "Register an already signed rating original without extending its window", false),
			vendorRatingDeviceAction("deliver", "UteamupVendorRatingDeviceDeliver", "Deliver the original rating after confirmation and return its minimal outcome", true),
		},
	})
}

func vendorRatingDeviceAction(name, tool, description string, confirm bool) Action {
	action := vendorDeviceAction(name, tool, description, confirm)
	action.Args[0].AllowedValues = []string{"create", "update", "delete", "flag", "moderate"}
	return action
}
