package registry

func init() {
	Register(&Domain{
		Name:        "contact",
		Aliases:     []string{"contacts"},
		Description: "Manage contacts",
		Actions: append(crudActions("Contact"),
			Action{Name: "search", Description: "Search contacts", ToolName: "UteamupContactSearch", Args: queryArg(), Flags: paginationFlags()},
		),
	})

	Register(&Domain{
		Name:        "customer",
		Aliases:     []string{"customers"},
		Description: "Manage customers",
		Actions: append(crudActions("Customer"),
			Action{Name: "search", Description: "Search customers", ToolName: "UteamupCustomerSearch", Args: queryArg(), Flags: paginationFlags()},
		),
	})

}
