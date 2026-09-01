package registry

import "testing"

// Locks the ifta domain to the REAL controller route. The previous generic
// crudActions registration derived a phantom /api/ifta base and 404'd on every
// action (same bug class as commit 6acd7c8); this test prevents a regression.
func TestIftaDomainTargetsRealFleetIftaRoute(t *testing.T) {
	t.Parallel()
	domain := findDomain("ifta")
	if domain == nil {
		t.Fatal("ifta domain is not registered")
	}
	if domain.APIPath != "/api/fleet/ifta" {
		t.Fatalf("API path = %q, want /api/fleet/ifta", domain.APIPath)
	}
	if len(domain.Actions) != 1 {
		t.Fatalf("actions = %d, want exactly the quarterly report (CSV export is [NonAction] server-side)", len(domain.Actions))
	}
	action := domain.Actions[0]
	if action.Name != "quarterly-report" || action.ToolName != "UteamupIftaGetQuarterlyReport" || action.HTTPMethod != "GET" {
		t.Fatalf("action contract = %+v", action)
	}
	path, _ := buildRESTPath(domain, action, map[string]any{})
	if path != "/api/fleet/ifta/report" {
		t.Fatalf("path = %q, want /api/fleet/ifta/report", path)
	}
}

// The GUID-first meter-reading domain in domains_meter_reading.go must be the ONLY
// registration — a duplicate generic one used to shadow it in help output.
func TestMeterReadingDomainIsRegisteredExactlyOnce(t *testing.T) {
	t.Parallel()
	count := 0
	for _, domain := range DefaultRegistry.Domains() {
		if domain.Name == "meter-reading" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("meter-reading registrations = %d, want 1", count)
	}
}
