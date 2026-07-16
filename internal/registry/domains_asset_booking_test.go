package registry

import "testing"

func assetBookingDomain(t *testing.T) *Domain {
	t.Helper()
	for _, domain := range DefaultRegistry.Domains() {
		if domain.Name == "asset-booking" {
			return domain
		}
	}
	t.Fatal("expected asset-booking domain")
	return nil
}

func TestAssetBookingDomainMirrorsGuidFirstMcpTools(t *testing.T) {
	domain := assetBookingDomain(t)
	wantTools := map[string]string{
		"list":      "UteamupAssetCalendarBookingList",
		"conflicts": "UteamupAssetCalendarBookingGetConflicts",
		"create":    "UteamupAssetCalendarBookingCreate",
		"delete":    "UteamupAssetCalendarBookingDelete",
	}

	if len(domain.Actions) != len(wantTools) {
		t.Fatalf("asset-booking actions = %d, want %d", len(domain.Actions), len(wantTools))
	}

	for _, action := range domain.Actions {
		if action.ToolName != wantTools[action.Name] {
			t.Errorf("asset-booking %s tool = %q, want %q", action.Name, action.ToolName, wantTools[action.Name])
		}
		for _, arg := range action.Args {
			if arg.Type != "string" {
				t.Errorf("asset-booking %s arg %s type = %q, want string GUID/date", action.Name, arg.Name, arg.Type)
			}
			if arg.Name == "id" || arg.Name == "assetId" || arg.Name == "bookingId" {
				t.Errorf("asset-booking %s leaked integer-style public arg %q", action.Name, arg.Name)
			}
		}
	}
}
