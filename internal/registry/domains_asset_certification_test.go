package registry

import (
	"testing"
)

func TestAssetCertificationStatusActionsUseNoIdentifierBoundary(t *testing.T) {
	domain := findDomain("asset-certification")
	if domain == nil {
		t.Fatal("asset-certification domain is not registered")
	}
	if domain.APIPath != "/api/assetcertification" {
		t.Fatalf("APIPath = %q, want /api/assetcertification", domain.APIPath)
	}

	actions := map[string]*Action{}
	for index := range domain.Actions {
		actions[domain.Actions[index].Name] = &domain.Actions[index]
	}
	expired := actions["expired"]
	if expired == nil || expired.ToolName != "UteamupAssetcertificationGetExpired" ||
		expired.HTTPMethod != "GET" || expired.RESTPath != "expired" ||
		len(expired.Args) != 0 || len(expired.Flags) != 0 {
		t.Fatalf("unexpected expired-certification action: %+v", expired)
	}
	expiring := actions["expiring"]
	if expiring == nil || expiring.ToolName != "UteamupAssetcertificationGetExpiringSoon" ||
		expiring.HTTPMethod != "GET" || expiring.RESTPath != "expiring-soon" ||
		len(expiring.Args) != 0 || len(expiring.Flags) != 1 ||
		expiring.Flags[0].Name != "days" {
		t.Fatalf("unexpected expiring-certification action: %+v", expiring)
	}
	list := actions["list"]
	if list == nil || list.ToolName != "UteamupAssetcertificationGetByAssetGuid" ||
		list.RESTPath != "asset/by-guid/{assetGuid}" {
		t.Fatalf("unexpected list-certification action: %+v", list)
	}
	get := actions["get"]
	if get == nil || get.ToolName != "UteamupAssetcertificationGetByGuid" ||
		get.RESTPath != "by-guid/{certificationGuid}" {
		t.Fatalf("unexpected get-certification action: %+v", get)
	}
}

func findAssetCertificationAction(t *testing.T, name string) (*Domain, *Action) {
	t.Helper()
	domain := findDomain("asset-certification")
	if domain == nil {
		t.Fatal("asset-certification domain is not registered")
	}
	for index := range domain.Actions {
		if domain.Actions[index].Name == name {
			return domain, &domain.Actions[index]
		}
	}
	t.Fatalf("asset-certification %s action is not registered", name)
	return nil, nil
}

func TestAssetCertificationListForToolUsesToolGuidRoute(t *testing.T) {
	domain, action := findAssetCertificationAction(t, "list-for-tool")
	assertListForToolAction(t, domain, action, "/api/assetcertification", "UteamupAssetcertificationGetByToolGuid")
}

func TestAssetCertificationCreateAcceptsAssetOrToolGuid(t *testing.T) {
	_, action := findAssetCertificationAction(t, "create")
	if action.ToolName != "UteamupAssetcertificationCreate" || action.HTTPMethod != "POST" || len(action.Args) != 0 {
		t.Fatalf("unexpected create action: %+v", action)
	}
	assertSubjectGuidFlags(t, action)
}

// assertSubjectGuidFlags checks the create action takes the reviewed JSON plus exactly the
// two GUID subject flags, sent as assetGuid / toolGuid; the backend enforces exactly one.
func assertSubjectGuidFlags(t *testing.T, action *Action) {
	t.Helper()
	want := map[string]string{"from-json": "", "asset-guid": "assetGuid", "tool-guid": "toolGuid"}
	if len(action.Flags) != len(want) {
		t.Fatalf("create flags = %+v, want from-json, asset-guid and tool-guid", action.Flags)
	}
	for _, flag := range action.Flags {
		body, ok := want[flag.Name]
		if !ok {
			t.Fatalf("unexpected create flag %+v", flag)
		}
		if flag.Name == "from-json" {
			if !flag.RootJSONObjectFile {
				t.Fatalf("from-json must be the root JSON object file: %+v", flag)
			}
			continue
		}
		if flag.Type != "uuid" || flag.BodyName != body || flag.Required {
			t.Fatalf("%s = %+v, want optional uuid sent as %s", flag.Name, flag, body)
		}
	}
}

func assertListForToolAction(t *testing.T, domain *Domain, action *Action, apiPath, toolName string) {
	t.Helper()
	if domain.APIPath != apiPath || action.ToolName != toolName || action.HTTPMethod != "GET" ||
		action.RESTPath != "tool/by-guid/{toolGuid}" {
		t.Fatalf("unexpected list-for-tool action: %+v", action)
	}
	if len(action.Args) != 1 || action.Args[0].Name != "toolGuid" || action.Args[0].Type != "uuid" || !action.Args[0].Required {
		t.Fatalf("list-for-tool args = %+v, want one required uuid toolGuid", action.Args)
	}
	guid := "11111111-2222-3333-4444-555555555555"
	path, consumed := buildRESTPath(domain, *action, map[string]any{"toolGuid": guid})
	if path != apiPath+"/tool/by-guid/"+guid {
		t.Fatalf("list-for-tool path = %q", path)
	}
	if len(consumed) != 1 || consumed[0] != "toolGuid" {
		t.Fatalf("list-for-tool consumed args = %v", consumed)
	}
}
