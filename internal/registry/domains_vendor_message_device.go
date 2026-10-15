package registry

func init() {
	Register(&Domain{
		Name:        "vendor-message-device",
		Description: "Deliver retained message originals through an enrolled device; never enrolls, signs or unlocks",
		APIPath:     "/api/v1/vendor/messages/reviewed",
		Actions: []Action{
			vendorMessageDeviceAction("proof", "UteamupVendorMessageDeviceProof", "Prepare a proof for an exact retained message original", false),
			vendorMessageDeviceAction("register", "UteamupVendorMessageDeviceRegister", "Register an already signed message original without extending its window", false),
			vendorMessageDeviceAction("deliver", "UteamupVendorMessageDeviceDeliver", "Deliver the original message change after confirmation and return its minimal outcome", true),
		},
	})
}

func vendorMessageDeviceAction(name, tool, description string, confirm bool) Action {
	action := vendorDeviceAction(name, tool, description, confirm)
	action.Args[0].AllowedValues = []string{"send", "read"}
	return action
}
