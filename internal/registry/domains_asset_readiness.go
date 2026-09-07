package registry

func init() {
	actions := []Action{}
	for _, item := range []struct{ name, tool, method, path, key string }{
		{"holds", "GetHolds", "GET", "asset/{assetGuid}/holds", "assetGuid"},
		{"hold", "CreateHold", "POST", "asset/{assetGuid}/holds", "assetGuid"},
		{"release-hold", "ReleaseHold", "POST", "holds/{holdGuid}/release", "holdGuid"},
		{"damage", "GetDamage", "GET", "asset/{assetGuid}/damage", "assetGuid"},
		{"record-damage", "CreateDamage", "POST", "asset/{assetGuid}/damage", "assetGuid"},
		{"resolve-damage", "ResolveDamage", "POST", "damage/{damageGuid}/resolve", "damageGuid"},
	} {
		action := Action{Name: item.name, Description: "Review and manage vehicle readiness evidence", ToolName: "UteamupAssetreadiness" + item.tool,
			HTTPMethod: item.method, RESTPath: item.path, Args: []ArgDef{{Name: item.key, Description: "Public source GUID", Required: true, Type: "uuid"}}}
		if item.method == "POST" {
			action.Flags = []FlagDef{{Name: "request-file", Short: "f", BodyName: "model", Description: "JSON evidence with GUID links and expectedUpdatedAtUtc for reviewed updates", Required: true, Type: "string", JSONFile: true}}
		}
		actions = append(actions, action)
	}
	Register(&Domain{Name: "asset-readiness", Description: "Manage operational holds and damage cases", Actions: actions})
}
