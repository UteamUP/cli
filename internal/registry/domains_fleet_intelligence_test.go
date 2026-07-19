package registry

import (
	"strings"
	"testing"
)

func TestFleetIntelligenceMirrorsGuidFirstReviewContracts(t *testing.T) {
	t.Parallel()
	domain := findDomain("fleet-intelligence")
	if domain == nil {
		t.Fatal("fleet-intelligence domain is not registered")
	}
	if domain.APIPath != "/api/fleet/intelligence" {
		t.Fatalf("API path = %q", domain.APIPath)
	}

	expected := map[string]struct {
		tool   string
		method string
		path   string
		args   map[string]any
	}{
		"anomalies": {
			tool:   "UteamupFleetIntelligenceGetAnomalies",
			method: "GET",
			path:   "/api/fleet/intelligence/fuel-idling-anomalies",
			args:   map[string]any{},
		},
		"tire-readiness": {
			tool:   "UteamupFleetIntelligenceGetTireReadiness",
			method: "GET",
			path:   "/api/fleet/intelligence/assets/asset-guid/tire-readiness",
			args:   map[string]any{"assetGuid": "asset-guid"},
		},
		"replacement-tco": {
			tool:   "UteamupFleetIntelligenceGetReplacementTco",
			method: "POST",
			path:   "/api/fleet/intelligence/assets/asset-guid/replacement-tco",
			args:   map[string]any{"assetGuid": "asset-guid"},
		},
	}
	if len(domain.Actions) != len(expected) {
		t.Fatalf("actions = %d, want %d", len(domain.Actions), len(expected))
	}
	for _, action := range domain.Actions {
		want, ok := expected[action.Name]
		if !ok {
			t.Fatalf("unexpected action %q", action.Name)
		}
		if action.ToolName != want.tool || action.HTTPMethod != want.method {
			t.Fatalf("%s contract = %+v", action.Name, action)
		}
		path, consumed := buildRESTPath(domain, action, want.args)
		if path != want.path {
			t.Fatalf("%s path = %q, want %q", action.Name, path, want.path)
		}
		if len(consumed) != len(action.Args) {
			t.Fatalf("%s did not consume all GUID args: %v", action.Name, consumed)
		}
		for _, argument := range action.Args {
			lower := strings.ToLower(argument.Name)
			if argument.Type != "uuid" ||
				(strings.HasSuffix(lower, "id") &&
					!strings.HasSuffix(lower, "guid")) {
				t.Fatalf("%s leaks non-GUID public identity: %+v", action.Name, argument)
			}
		}
	}
}

func TestFleetReplacementTcoRequiresExplicitAssumptions(t *testing.T) {
	t.Parallel()
	domain := findDomain("fleet-intelligence")
	var tco Action
	for _, action := range domain.Actions {
		if action.Name == "replacement-tco" {
			tco = action
			break
		}
	}
	required := map[string]string{
		"currency":                            "currency",
		"residual-value":                      "residualValue",
		"financing-cost":                      "financingCost",
		"downtime-cost-per-hour":              "downtimeCostPerHour",
		"replacement-annual-maintenance-cost": "replacementAnnualMaintenanceCost",
		"replacement-fuel-reduction-percent":  "replacementFuelReductionPercent",
	}
	for _, flag := range tco.Flags {
		bodyName := flag.BodyName
		if bodyName == "" {
			bodyName = toCamelCase(flag.Name)
		}
		if expectedBody, ok := required[flag.Name]; ok {
			if !flag.Required || bodyName != expectedBody {
				t.Fatalf("%s flag = %+v", flag.Name, flag)
			}
			delete(required, flag.Name)
		}
		if strings.Contains(strings.ToLower(flag.Name), "tenant") {
			t.Fatalf("caller-controlled tenant scope: %+v", flag)
		}
	}
	if len(required) != 0 {
		t.Fatalf("missing explicit TCO flags: %v", required)
	}
}
