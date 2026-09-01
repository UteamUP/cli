package registry

import "testing"

func TestVinDecodeDomainTargetsRealController(t *testing.T) {
	t.Parallel()
	domain := findDomain("vindecode")
	if domain == nil {
		t.Fatal("vindecode domain is not registered")
	}
	if domain.APIPath != "/api/vindecode" {
		t.Fatalf("API path = %q, want /api/vindecode", domain.APIPath)
	}
	if len(domain.Actions) != 1 {
		t.Fatalf("actions = %d, want 1", len(domain.Actions))
	}
	action := domain.Actions[0]
	if action.Name != "decode" || action.ToolName != "UteamupFleetVinDecode" || action.HTTPMethod != "GET" {
		t.Fatalf("action contract = %+v", action)
	}
	path, _ := buildRESTPath(domain, action, map[string]any{"vin": "1HGCM82633A004352"})
	if path != "/api/vindecode/1HGCM82633A004352" {
		t.Fatalf("path = %q", path)
	}
}
