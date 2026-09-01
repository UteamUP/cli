package registry

import "testing"

func TestDistanceTaxDomainTargetsRealController(t *testing.T) {
	t.Parallel()
	domain := findDomain("distance-tax")
	if domain == nil {
		t.Fatal("distance-tax domain is not registered")
	}
	if domain.APIPath != "/api/fleet/compliance" {
		t.Fatalf("API path = %q, want /api/fleet/compliance", domain.APIPath)
	}
	if len(domain.Actions) != 3 {
		t.Fatalf("actions = %d, want 3", len(domain.Actions))
	}

	wantPaths := map[string]string{
		"schemes":       "/api/fleet/compliance/distance-tax/schemes",
		"report":        "/api/fleet/compliance/distance-tax/report",
		"statutory-due": "/api/fleet/compliance/statutory-readings/due",
	}
	for _, action := range domain.Actions {
		want, ok := wantPaths[action.Name]
		if !ok {
			t.Fatalf("unexpected action %q", action.Name)
		}
		if action.HTTPMethod != "GET" {
			t.Fatalf("%s method = %q, want GET", action.Name, action.HTTPMethod)
		}
		path, _ := buildRESTPath(domain, action, map[string]any{})
		if path != want {
			t.Fatalf("%s path = %q, want %q", action.Name, path, want)
		}
	}
}

func TestMeterReadingStatutoryActionsTargetLedgerRoutes(t *testing.T) {
	t.Parallel()
	domain := findDomain("meter-reading")
	if domain == nil {
		t.Fatal("meter-reading domain is not registered")
	}

	assetGuid := "3e1cd135-07d1-46ff-8f5f-549da5390e78"
	attrGuid := "a1b2c3d4-e5f6-4a5b-8c7d-9e0f1a2b3c4d"
	readingGuid := "b2c3d4e5-f6a7-4b5c-8d7e-9f0a1b2c3d4e"

	cases := map[string]string{
		"record-statutory": "/api/assets/" + assetGuid + "/meter-readings/" + attrGuid + "/statutory",
		"correct":          "/api/assets/" + assetGuid + "/meter-readings/by-guid/" + readingGuid + "/correct",
		"replace-meter":    "/api/assets/" + assetGuid + "/meter-readings/" + attrGuid + "/replace-meter",
	}
	seen := map[string]bool{}
	for _, action := range domain.Actions {
		want, ok := cases[action.Name]
		if !ok {
			continue
		}
		seen[action.Name] = true
		if action.HTTPMethod != "POST" {
			t.Fatalf("%s method = %q, want POST", action.Name, action.HTTPMethod)
		}
		path, _ := buildRESTPath(domain, action, map[string]any{
			"asset-guid":                assetGuid,
			"attribute-definition-guid": attrGuid,
			"reading-guid":              readingGuid,
		})
		if path != want {
			t.Fatalf("%s path = %q, want %q", action.Name, path, want)
		}
	}
	for name := range cases {
		if !seen[name] {
			t.Fatalf("action %q missing from meter-reading domain", name)
		}
	}
}

// The camelCase default of the value/timestamp flags silently posted
// {value, timestamp}, which the backend does not bind — recording 0. The
// explicit BodyName mapping is load-bearing; lock it.
func TestMeterReadingRecordFlagsCarryBackendBodyNames(t *testing.T) {
	t.Parallel()
	domain := findDomain("meter-reading")
	if domain == nil {
		t.Fatal("meter-reading domain is not registered")
	}
	wantBodyNames := map[string]map[string]string{
		"record":           {"value": "readingValue", "timestamp": "readingTimestamp"},
		"record-statutory": {"value": "readingValue", "observed-at": "observedAt", "evidence-url": "evidenceDocumentUrl"},
		"correct":          {"value": "correctedValue"},
		"replace-meter":    {"initial-value": "initialValue"},
	}
	for _, action := range domain.Actions {
		want, ok := wantBodyNames[action.Name]
		if !ok {
			continue
		}
		for _, flag := range action.Flags {
			if expected, tracked := want[flag.Name]; tracked && flag.BodyName != expected {
				t.Fatalf("%s flag %q BodyName = %q, want %q", action.Name, flag.Name, flag.BodyName, expected)
			}
		}
	}
}
