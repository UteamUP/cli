package registry

func init() {
	Register(&Domain{
		Name:        "handoverattestation",
		Aliases:     []string{"handover-attestation", "handover-transfer", "attestation"},
		Description: "Issue and atomically redeem durable handover transfer challenges",
		APIPath:     "/api/handoverattestation",
		Actions: []Action{
			{
				Name:                  "issue",
				Description:           "Issue a rotating challenge as the designated outgoing operator (0 AI credits); automatic JSON export is disabled because the response contains secrets",
				ToolName:              "UteamupHandoverAttestationIssue",
				HTTPMethod:            "POST",
				RESTPath:              "{handover-guid}/issue",
				DisableResponseExport: true,
				Args: []ArgDef{
					{Name: "handover-guid", Description: "Handover external Guid", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "redeem",
				Description: "Atomically redeem a challenge as the designated incoming operator (0 AI credits)",
				ToolName:    "UteamupHandoverAttestationRedeem",
				HTTPMethod:  "POST",
				RESTPath:    "redeem",
				Flags: []FlagDef{
					{Name: "token", Description: "The signed transfer token to redeem once", Type: "string", Required: true, Sensitive: true},
				},
			},
			{
				Name:        "redeem-code",
				Description: "Redeem an 8-digit fallback code (5 attempts/minute, 0 AI credits); opens review only and never completes the handover",
				ToolName:    "UteamupHandoverAttestationRedeemCode",
				HTTPMethod:  "POST",
				RESTPath:    "redeem-code",
				Flags: []FlagDef{
					{Name: "code", BodyName: "code", Description: "The 8-digit fallback transfer code", Type: "string", Required: true, Sensitive: true},
				},
			},
			{
				Name:        "verify",
				Description: "Deprecated alias: atomically redeems the token; it is not a read-only check",
				ToolName:    "UteamupHandoverAttestationRedeem",
				HTTPMethod:  "POST",
				RESTPath:    "redeem",
				Flags: []FlagDef{
					{Name: "token", Description: "The signed transfer token to redeem once", Type: "string", Required: true, Sensitive: true},
				},
			},
		},
	})
}
