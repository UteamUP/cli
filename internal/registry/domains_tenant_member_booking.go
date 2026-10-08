package registry

// The registry "tenant" domain is shadowed by the local cmd/tenant.go command,
// so member booking settings live in their own reachable domain.
func init() {
	Register(&Domain{
		Name:        "member-booking",
		Aliases:     []string{"member-bookings"},
		Description: "Manage whether tenant members can be booked as resources and on public booking pages",
		APIPath:     "/api/tenant",
		Actions: []Action{
			{
				Name:        "settings-set",
				Description: "Set whether a member can be booked as a resource and offered on public booking pages; omitted settings are unchanged",
				ToolName:    "UteamupTenantMemberBookingSettingsUpdate",
				HTTPMethod:  "PUT",
				RESTPath:    "users/{userGuid}/booking-settings",
				Args:        []ArgDef{{Name: "userGuid", Description: "Member user public GUID", Required: true, Type: "uuid"}},
				Flags: []FlagDef{
					{Name: "is-bookable-resource", BodyName: "isBookableResource", Description: "Whether the member can be reserved as a technician", Type: "bool"},
					{Name: "allow-public-booking", BodyName: "allowPublicBooking", Description: "Whether public booking pages may offer the member (the member can still opt out)", Type: "bool"},
				},
			},
			{
				Name:        "my-preferences-get",
				Description: "Get your own booking settings in the active tenant",
				ToolName:    "UteamupMyBookingPreferencesGet",
				HTTPMethod:  "GET",
				RESTPath:    "me/booking-preferences",
			},
			{
				Name:        "my-preferences-set",
				Description: "Opt yourself out of, or back into, public booking pages in the active tenant",
				ToolName:    "UteamupMyBookingPreferencesUpdate",
				HTTPMethod:  "PUT",
				RESTPath:    "me/booking-preferences",
				Flags: []FlagDef{
					{Name: "public-booking-opt-out", BodyName: "publicBookingOptOut", Description: "true hides you from public booking pages; false allows it again", Required: true, Type: "bool"},
				},
			},
		},
	})
}
