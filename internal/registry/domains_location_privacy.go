package registry

// Mirrors the MCP UteamupLocationPrivacy* tools backed by
// LocationPrivacyController on the backend. GUID-first per the
// GUIDs-In/Integer-Ids-Out rule.
//
// There is deliberately no approve, decline or grant action: consent to share
// a location comes only from the person, in the app. The self-service actions
// (grants, access-log, revoke) act for the signed-in user and take no user or
// tenant argument.
//
// REST surface:
//
//	GET  /api/locationprivacy/my/grants                   — my requests and shares
//	GET  /api/locationprivacy/my/access-log               — who viewed my location (?page&pageSize)
//	POST /api/locationprivacy/grants/{grantGuid}/revoke   — end a share I gave or received
//	POST /api/locationprivacy/requests                    — ask a colleague to share (Geofence.RequestLocation)
//	GET  /api/locationprivacy/policy                      — tenant policy (Privacy.Read)
//	PUT  /api/locationprivacy/policy                      — save tenant policy (Privacy.Manage)
func init() {
	Register(&Domain{
		Name:        "location-privacy",
		Description: "Your location shares and access log, share requests, and the tenant location privacy policy",
		APIPath:     "/api/locationprivacy",
		Actions: []Action{
			{
				Name:        "grants",
				Description: "List requests waiting for your answer and the location shares you have given",
				ToolName:    "UteamupLocationPrivacyMyGrants",
				HTTPMethod:  "GET",
				RESTPath:    "my/grants",
			},
			{
				Name:        "access-log",
				Description: "See who viewed your location, newest first",
				ToolName:    "UteamupLocationPrivacyMyAccessLog",
				HTTPMethod:  "GET",
				RESTPath:    "my/access-log",
				Flags: []FlagDef{
					{Name: "page", Description: "Page number, from 1", Default: 1, Type: "int"},
					{Name: "page-size", Description: "Rows per page, 1-100", Default: 50, Type: "int"},
				},
			},
			{
				Name:        "revoke",
				Description: "End a location share you gave or received",
				ToolName:    "UteamupLocationPrivacyRevokeGrant",
				HTTPMethod:  "POST",
				RESTPath:    "grants/{grantGuid}/revoke",
				Args: []ArgDef{
					{Name: "grantGuid", Description: "The share's public GUID", Required: true, Type: "non-empty-uuid"},
				},
			},
			{
				Name:        "request",
				Description: "Ask a colleague to share their location with you; nothing is shared until they approve in the app",
				ToolName:    "UteamupLocationPrivacyRequestShare",
				HTTPMethod:  "POST",
				RESTPath:    "requests",
				Flags: []FlagDef{
					{Name: "subject-user-guid", Description: "The colleague's user GUID", Required: true, Type: "non-empty-uuid"},
					{Name: "scopes", Description: "Sum of 1 = current presence, 2 = history, 4 = arrival/departure alerts", Required: true, Type: "int"},
					{Name: "duration-days", Description: "How long the share should last, 1-365 (capped by the tenant policy)", Default: 30, Type: "int"},
					{Name: "reason", Description: "Why you need it, 5-500 characters; the colleague sees this", Required: true, Type: "string"},
				},
			},
			{
				Name:        "policy-get",
				Description: "Read the tenant's location privacy policy",
				ToolName:    "UteamupLocationPrivacyGetPolicy",
				HTTPMethod:  "GET",
				RESTPath:    "policy",
			},
			{
				Name:        "policy-update",
				Description: "Save the tenant's location privacy policy from a JSON file; every change is recorded",
				ToolName:    "UteamupLocationPrivacyUpdatePolicy",
				HTTPMethod:  "PUT",
				RESTPath:    "policy",
				Flags: []FlagDef{
					{Name: "policy-file", Short: "f", Description: "The full policy as a JSON object", Required: true, Type: "string", RootJSONObjectFile: true},
				},
			},
		},
	})
}
