package registry

// Domain: identity
//
// Read-only identity verification status (electronic ID, rafræn skilríki) for the signed-in
// user. Identity comes from the bearer token; the command takes no user argument. The CLI never
// starts an electronic-ID flow: the person verifies in the UteamUP app.
//
// Paired MCP tool:
//
//	UteamupIdentityVerificationStatus
func init() {
	Register(&Domain{
		Name:        "identity",
		Aliases:     []string{"eid", "electronic-id"},
		Description: "Show the signed-in user's identity verification (electronic ID)",
		APIPath:     "/api/auth/eid",
		Actions: []Action{
			{
				Name:        "status",
				Description: "Whether you are verified with an electronic ID, when, provider, assurance level and the masked national ID",
				ToolName:    "UteamupIdentityVerificationStatus",
				HTTPMethod:  "GET",
				RESTPath:    "status",
			},
		},
	})
}
