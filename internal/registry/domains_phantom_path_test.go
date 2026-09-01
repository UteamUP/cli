package registry

import (
	"strings"
	"testing"
)

// The phantom-path defect: when a domain declares no APIPath and an action declares no
// RESTPath, buildRESTPath falls back to "/api/" + strings.ReplaceAll(domain.Name, "-", "").
// That is CORRECT whenever the derived base happens to match a real controller (asset,
// tenant, plan, user, workorder, iot, partner, bugsandfeatures…), and silently 404s when it
// does not. These eight domains were in the second group; each is now pinned to the real
// controller route below.
//
// Verifying against a live backend: probe UNAUTHENTICATED and compare 401 (endpoint exists)
// with 404 (phantom). Do NOT probe authenticated and read 403 as proof — the tenant
// middleware answers 403 before routing resolves, so a phantom path 403s identically. The
// signal is also controller-dependent (GlobalAdminController 401s, InternalBillingController
// 404s even for real routes), so the authoritative check is the static one these tests do:
// the built path must match a real [Route]/[Http*] template.

func phantomAction(t *testing.T, domainName, actionName string) (*Domain, Action) {
	t.Helper()
	domain := findDomain(domainName)
	if domain == nil {
		t.Fatalf("%s domain is not registered", domainName)
	}
	for _, action := range domain.Actions {
		if action.Name == actionName {
			return domain, action
		}
	}
	t.Fatalf("%s action %q is not registered", domainName, actionName)
	return nil, Action{}
}

func TestPhantomPathDomainsResolveToRealControllerRoutes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		domain, action, method string
		args                   map[string]any
		path                   string
	}{
		// admin-users → GlobalAdminUsersController
		{"admin-users", "list", "GET", nil, "/api/globaladmin/users/list"},
		{"admin-users", "get", "GET", map[string]any{"guid": "u"}, "/api/globaladmin/users/u"},
		{"admin-users", "login-events", "GET", map[string]any{"guid": "u"}, "/api/globaladmin/users/u/login-events"},
		{"admin-users", "disable", "POST", map[string]any{"guid": "u"}, "/api/globaladmin/users/u/disable"},
		{"admin-users", "enable", "POST", map[string]any{"guid": "u"}, "/api/globaladmin/users/u/enable"},
		{"admin-users", "reset-password", "POST", map[string]any{"guid": "u"}, "/api/globaladmin/users/u/reset-password"},

		// condition → AssetConditionAssessmentController
		{"condition", "assess", "POST", nil, "/api/assetconditionassessment"},
		{"condition", "get", "GET", map[string]any{"assetGuid": "a"}, "/api/assetconditionassessment/by-guid/a/latest"},
		{"condition", "history", "GET", map[string]any{"assetGuid": "a"}, "/api/assetconditionassessment/by-guid/a/history"},
		{"condition", "heat-map", "GET", nil, "/api/assetconditionassessment/heat-map"},
		{"condition", "overdue", "GET", nil, "/api/assetconditionassessment/due"},

		// criticality → AssetCriticalityAssessmentController
		{"criticality", "assess", "POST", nil, "/api/assetcriticalityassessment"},
		{"criticality", "get", "GET", map[string]any{"assetGuid": "a"}, "/api/assetcriticalityassessment/by-guid/a"},
		{"criticality", "history", "GET", map[string]any{"assetGuid": "a"}, "/api/assetcriticalityassessment/by-guid/a/history"},
		{"criticality", "matrix", "GET", nil, "/api/assetcriticalityassessment/matrix"},

		// meter-reading → MeterReadingController (/api/assets/{assetGuid}/…)
		{"meter-reading", "current", "GET", map[string]any{"assetGuid": "a"}, "/api/assets/a/meter-readings/current"},
		{"meter-reading", "attributes", "GET", map[string]any{"assetGuid": "a"}, "/api/assets/a/attributes"},
		{"meter-reading", "history", "GET", map[string]any{"assetGuid": "a", "attributeDefinitionGuid": "d"}, "/api/assets/a/meter-readings/d/history"},
		{"meter-reading", "record", "POST", map[string]any{"assetGuid": "a"}, "/api/assets/a/meter-readings"},
		{"meter-reading", "ocr", "POST", map[string]any{"assetGuid": "a", "attributeDefinitionGuid": "d"}, "/api/assets/a/meter-readings/d/ocr"},

		// pdfhotspot → CodePdfLinkController (list-for-drawing is served by /api/document)
		{"pdfhotspot", "list-for-drawing", "GET", map[string]any{"documentGuid": "d"}, "/api/document/d/codehotspots"},
		{"pdfhotspot", "create", "POST", map[string]any{"linkGuid": "l"}, "/api/codepdflink/l/hotspots"},
		{"pdfhotspot", "update", "PUT", map[string]any{"linkGuid": "l", "hotspotGuid": "h"}, "/api/codepdflink/l/hotspots/h"},
		{"pdfhotspot", "delete", "DELETE", map[string]any{"linkGuid": "l", "hotspotGuid": "h"}, "/api/codepdflink/l/hotspots/h"},

		// user-ui-state → UserPreferencesController + UserStateController
		{"user-ui-state", "get-preferences", "GET", nil, "/api/userpreferences"},
		{"user-ui-state", "set-preferences", "PUT", nil, "/api/userpreferences"},
		{"user-ui-state", "get-last-page", "GET", nil, "/api/userstate/last-page"},
		{"user-ui-state", "set-last-page", "PUT", nil, "/api/userstate/last-page"},
		{"user-ui-state", "clear-last-page", "DELETE", nil, "/api/userstate/last-page"},

		// fleet-dashboard → FleetDashboardController
		{"fleet-dashboard", "get", "GET", nil, "/api/fleet/dashboard"},
		{"fleet-dashboard", "utilization", "GET", nil, "/api/fleet/dashboard/utilization"},
		{"fleet-dashboard", "compliance", "GET", nil, "/api/fleet/dashboard/compliance"},
		{"fleet-dashboard", "costs", "GET", nil, "/api/fleet/dashboard/costs"},

		// route → OperationalRoute{,Schedule,Execution}Controller + InspectionAI for optimize
		{"route", "list", "GET", nil, "/api/operationalroutes"},
		{"route", "get", "GET", map[string]any{"routeGuid": "r"}, "/api/operationalroutes/by-guid/r"},
		{"route", "overdue", "GET", nil, "/api/operationalrouteschedules/overdue"},
		{"route", "executions", "GET", nil, "/api/operationalrouteexecutions"},
		{"route", "execution", "GET", map[string]any{"executionGuid": "e"}, "/api/operationalrouteexecutions/e"},
		{"route", "start", "POST", nil, "/api/operationalrouteexecutions/by-guid/start"},
		{"route", "complete-stop", "PUT", map[string]any{"stopGuid": "s"}, "/api/operationalrouteexecutions/by-guid/stops/s/complete"},
		{"route", "flag-issue", "POST", map[string]any{"stopGuid": "s"}, "/api/operationalrouteexecutions/by-guid/stops/s/flag-issue"},
		{"route", "complete", "PUT", map[string]any{"executionGuid": "e"}, "/api/operationalrouteexecutions/by-guid/e/complete"},
		{"route", "abandon", "PUT", map[string]any{"executionGuid": "e"}, "/api/operationalrouteexecutions/by-guid/e/abandon"},
		{"route", "optimize", "GET", map[string]any{"routeGuid": "r"}, "/api/inspectionai/routes/by-guid/r/optimize"},
	}
	for _, test := range tests {
		t.Run(test.domain+"-"+test.action, func(t *testing.T) {
			domain, action := phantomAction(t, test.domain, test.action)
			if action.MCPOnly {
				t.Fatalf("%s %s is MCPOnly — it must not be in the REST route table", test.domain, test.action)
			}
			if action.HTTPMethod != test.method {
				t.Fatalf("method = %q, want %q (names outside the HTTPMethod map silently default to GET)",
					action.HTTPMethod, test.method)
			}
			args := test.args
			if args == nil {
				args = map[string]any{}
			}
			path, _ := buildRESTPath(domain, action, args)
			if path != test.path {
				t.Fatalf("path = %q, want %q", path, test.path)
			}
		})
	}
}

func TestPhantomPathDomainsNeverDeriveTheStrippedFallback(t *testing.T) {
	t.Parallel()
	// The exact regression: no APIPath + no RESTPath → "/api/" + name with dashes stripped.
	for _, name := range []string{"admin-users", "condition", "criticality", "meter-reading",
		"route", "pdfhotspot", "user-ui-state", "fleet-dashboard"} {
		domain := findDomain(name)
		if domain == nil {
			t.Fatalf("%s domain is not registered", name)
		}
		if domain.APIPath == "" {
			t.Fatalf("%s declares no APIPath — the stripped fallback returns", name)
		}
		phantom := "/api/" + strings.ReplaceAll(name, "-", "")
		for _, action := range domain.Actions {
			if action.MCPOnly {
				continue // never touches buildRESTPath
			}
			path, _ := buildRESTPath(domain, action, map[string]any{
				"guid": "g", "assetGuid": "a", "attributeDefinitionGuid": "d", "documentGuid": "d",
				"linkGuid": "l", "hotspotGuid": "h", "routeGuid": "r", "executionGuid": "e", "stopGuid": "s",
				"readingGuid": "m",
			})
			if strings.HasPrefix(path, phantom) {
				t.Fatalf("%s %s resolves to the phantom path %s", name, action.Name, path)
			}
			if strings.Contains(path, "{") {
				t.Fatalf("%s %s left an unexpanded placeholder: %s (args are camelCased in toolArgs — "+
					"a dashed placeholder like {asset-guid} never resolves)", name, action.Name, path)
			}
		}
	}
}
