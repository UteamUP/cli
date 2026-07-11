package registry

func init() {
	Register(&Domain{
		Name:        "handoverattestation",
		Aliases:     []string{"handover-attestation", "attestation"},
		Description: "Issue and verify handover attestation tokens (co-presence proof)",
		APIPath:     "/api/handoverattestation",
		Actions: []Action{
			{
				Name:        "issue",
				Description: "Issue a short-TTL attestation token for a handover (you are the subject)",
				ToolName:    "UteamupHandoverAttestationIssue",
				HTTPMethod:  "POST",
				RESTPath:    "{handover-guid}/issue",
				Args: []ArgDef{
					{Name: "handover-guid", Description: "Handover external Guid", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "verify",
				Description: "Verify a handover attestation token (signature + TTL + single-use)",
				ToolName:    "UteamupHandoverAttestationVerify",
				HTTPMethod:  "POST",
				RESTPath:    "verify",
				Flags: []FlagDef{
					{Name: "token", Description: "The attestation token to verify", Type: "string", Required: true},
				},
			},
		},
	})
}
