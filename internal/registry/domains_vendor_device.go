package registry

func init() {
	Register(&Domain{
		Name:        "vendor-device",
		Description: "Deliver retained Vendor originals through an existing enrolled device; never enrolls, signs or unlocks",
		APIPath:     "/api/vendor/reviewed",
		Actions: []Action{
			vendorDeviceAction("proof", "UteamupVendorDeviceProof", "Prepare a proof for an exact retained Vendor original", false),
			vendorDeviceAction("register", "UteamupVendorDeviceRegister", "Register an already device-signed original without extending its replay window", false),
			vendorDeviceAction("deliver", "UteamupVendorDeviceDeliver", "Deliver an already registered original and return its minimal durable outcome", true),
		},
	})
}

func vendorDeviceAction(name, tool, description string, confirm bool) Action {
	flags := []FlagDef{{Name: "from-json", Description: "File containing the complete protocol DTO; originalJson must retain exact original bytes", Type: "string", Required: true, RootJSONObjectFile: true}}
	if confirm {
		flags = append(flags, FlagDef{Name: "confirm", Description: "Confirm this exact retained Vendor change", Type: "bool", Required: true, MustBeTrue: true, LocalOnly: true})
	}
	return Action{
		Name: name, Description: description, ToolName: tool,
		HTTPMethod: "POST", RESTPath: "{operationAction}/" + name,
		RejectExtraArgs: true, DisableResponseExport: true,
		SensitiveBodyFields: []string{"originalJson", "signature"},
		Args:                []ArgDef{{Name: "operation-action", BodyName: "operationAction", Description: "Original operation action", Type: "string", Required: true, AllowedValues: []string{"create", "update", "delete"}}},
		Flags:               flags,
	}
}
