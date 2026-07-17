package registry

import "testing"

func TestAssetCertificationStatusActionsUseNoIdentifierBoundary(t *testing.T) {
	domain := findDomain("asset-certification")
	if domain == nil {
		t.Fatal("asset-certification domain is not registered")
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
}
