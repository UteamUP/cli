package registry

// Complete root JSON preserves explicit proof null versus omission. Actor and
// tenant are supplied by the authenticated context, never by request selectors.
func qualificationRequestFile() FlagDef {
	return FlagDef{Name: "from-json", Description: "File containing the complete GUID-only personal request; omit proof to preserve it, explicit null clears only in details edits",
		Type: "string", Required: true, RootJSONObjectFile: true}
}

func personalSkillActions() []Action {
	return []Action{
		{Name: "add-mine", Description: "Add a catalog skillGuid to your profile; verification remains read-only",
			ToolName: "UteamupMySkillAdd", RESTBasePath: "/api/skill", RESTPath: "user-skill/by-guid", HTTPMethod: "POST",
			Flags: []FlagDef{qualificationRequestFile()}},
		{Name: "update-mine", Description: "Update your held skill experience without retargeting or changing verification",
			ToolName: "UteamupMySkillUpdate", RESTBasePath: "/api/skill", RESTPath: "user-skill/by-guid/{userSkillGuid}", HTTPMethod: "PUT",
			Args:  []ArgDef{{Name: "userSkillGuid", Description: "Held skill public GUID", Required: true, Type: "non-empty-uuid"}},
			Flags: []FlagDef{qualificationRequestFile()}},
	}
}

func personalCertificateActions() []Action {
	return []Action{
		{Name: "update-mine", Description: "Edit your certificate details; catalog and skill stay unchanged, approved edits become Pending",
			ToolName: "UteamupMyCertificateUpdate", RESTBasePath: "/api/certificate", RESTPath: "user-certificate/by-guid/{userCertificateGuid}/details", HTTPMethod: "PUT",
			Args:  []ArgDef{{Name: "userCertificateGuid", Description: "Held certificate public GUID", Required: true, Type: "non-empty-uuid"}},
			Flags: []FlagDef{qualificationRequestFile()}},
		{Name: "renew-with-proof", Description: "Renew an Approved held certificate with a public proof GUID; another holder requires Certificate.ViewUserCertificates; do not replay an unknown result",
			ToolName: "UteamupMyCertificateRenew", RESTBasePath: "/api/certificate", RESTPath: "user-certificate/by-guid/{userCertificateGuid}/renew-with-proof", HTTPMethod: "POST",
			Args:  []ArgDef{{Name: "userCertificateGuid", Description: "Held certificate public GUID", Required: true, Type: "non-empty-uuid"}},
			Flags: []FlagDef{qualificationRequestFile()}},
	}
}
