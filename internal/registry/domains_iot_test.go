package registry

import "testing"

import "strings"

func TestIoTDomainMirrorsMCPTools(t *testing.T) {
	domain := findDomain("iot")
	if domain == nil {
		t.Fatal("expected iot domain to be registered")
	}
	expected := map[string]string{
		"overview":                   "UteamupIoTTelemetryOverview",
		"status":                     "UteamupIoTEnvironmentStatus",
		"monitoring":                 "UteamupIoTMonitoringDashboard",
		"telemetry":                  "UteamupIoTTelemetryPoints",
		"rules":                      "UteamupIoTRulesList",
		"device-types":               "UteamupIoTDeviceTypesList",
		"devices":                    "UteamupIoTDevicesList",
		"device":                     "UteamupIoTDeviceGet",
		"device-create":              "UteamupIoTDeviceCreate",
		"device-update":              "UteamupIoTDeviceUpdate",
		"device-keys":                "UteamupIoTDeviceKeysList",
		"device-key-issue":           "UteamupIoTDeviceKeyIssue",
		"device-key-revoke":          "UteamupIoTDeviceKeyRevoke",
		"command-definitions":        "UteamupIoTCommandDefinitionsList",
		"command-definition-history": "UteamupIoTCommandDefinitionHistory",
		"command-definition-create":  "UteamupIoTCommandDefinitionCreate",
		"command-definition-update":  "UteamupIoTCommandDefinitionUpdate",
		"command-control":            "UteamupIoTCommandControlGet",
		"command-control-update":     "UteamupIoTCommandControlUpdate",
		"command-preview":            "UteamupIoTCommandRequestPreview",
		"command-requests":           "UteamupIoTCommandRequestsList",
		"command-request":            "UteamupIoTCommandRequestGet",
		"command-confirm":            "UteamupIoTCommandRequestConfirm",
		"command-approve":            "UteamupIoTCommandRequestApprove",
		"command-reject":             "UteamupIoTCommandRequestReject",
		"command-cancel":             "UteamupIoTCommandRequestCancel",
		"command-monitoring":         "UteamupIoTCommandMonitoringGet",
	}
	for _, action := range domain.Actions {
		if tool, ok := expected[action.Name]; ok {
			if action.ToolName != tool {
				t.Errorf("action %q maps to %q, want %q", action.Name, action.ToolName, tool)
			}
			delete(expected, action.Name)
		}
	}
	for missing := range expected {
		t.Errorf("missing iot action %q", missing)
	}
}

func TestIoTTelemetryUsesGuidFiltersAndBoundedLimit(t *testing.T) {
	domain := findDomain("iot")
	var telemetry Action
	for _, action := range domain.Actions {
		if action.Name == "telemetry" {
			telemetry = action
		}
	}
	flags := map[string]FlagDef{}
	for _, flag := range telemetry.Flags {
		flags[flag.Name] = flag
	}
	for _, name := range []string{"device-guid", "asset-guid", "attribute-definition-guid", "before-point-guid"} {
		if flags[name].Type != "string" {
			t.Errorf("%s must be a GUID string flag", name)
		}
	}
	if flags["limit"].Default != 100 {
		t.Errorf("telemetry limit default = %v, want 100", flags["limit"].Default)
	}
}

func TestIoTMonitoringDescriptionExposesOperationalHealth(t *testing.T) {
	domain := findDomain("iot")
	for _, action := range domain.Actions {
		if action.Name != "monitoring" {
			continue
		}
		for _, term := range []string{"freshness", "backlogs", "credentials"} {
			if !strings.Contains(action.Description, term) {
				t.Errorf("monitoring description must mention %q", term)
			}
		}
		return
	}
	t.Fatal("expected monitoring action")
}

func TestIoTCommandRoutesUseGuidOnlyPublicIdentities(t *testing.T) {
	domain := findDomain("iot")
	tests := []struct {
		actionName string
		arguments  map[string]any
		path       string
	}{
		{
			"command-definition-update",
			map[string]any{"definitionGuid": "definition-guid"},
			"/api/iot/commands/definitions/definition-guid",
		},
		{
			"command-request",
			map[string]any{"requestGuid": "request-guid"},
			"/api/iot/commands/requests/request-guid",
		},
		{
			"command-confirm",
			map[string]any{"requestGuid": "request-guid"},
			"/api/iot/commands/requests/request-guid/confirm",
		},
		{
			"command-approve",
			map[string]any{"requestGuid": "request-guid"},
			"/api/iot/commands/requests/request-guid/approve",
		},
		{
			"command-reject",
			map[string]any{"requestGuid": "request-guid"},
			"/api/iot/commands/requests/request-guid/reject",
		},
		{
			"command-cancel",
			map[string]any{"requestGuid": "request-guid"},
			"/api/iot/commands/requests/request-guid/cancel",
		},
	}

	for _, test := range tests {
		t.Run(test.actionName, func(t *testing.T) {
			action := findIoTAction(t, test.actionName)
			path, consumed := buildRESTPath(domain, action, test.arguments)
			if path != test.path {
				t.Fatalf("path = %q, want %q", path, test.path)
			}
			if len(consumed) != 1 {
				t.Fatalf("expected one consumed GUID arg, got %v", consumed)
			}
			if action.Args[0].Type != "uuid" ||
				!strings.HasSuffix(action.Args[0].Name, "Guid") {
				t.Fatalf("public route identity is not GUID-only: %+v", action.Args[0])
			}
		})
	}
}

// Every IoT action's URL is pinned to its backend [Route]. The four read actions used to
// declare no path, so each fell through to the bare domain base /api/iot, which no
// controller serves — `ut iot status` answered 404 for every tenant. A new action must be
// added here, or this test fails.
func TestIoTActionsRouteToTheirBackendControllers(t *testing.T) {
	domain := findDomain("iot")
	expected := map[string]string{
		"overview":                   "/api/iot/monitoring/telemetry/overview",
		"status":                     "/api/iot/environment",                  // TenantIoTSelfServiceController.GetStatus
		"monitoring":                 "/api/iot/monitoring",                   // IoTMonitoringController dashboard
		"telemetry":                  "/api/iot/monitoring/telemetry/points",  // IoTMonitoringController.GetTelemetryPoints
		"rules":                      "/api/iot/rules",                        // IoTRuleController.Get
		"device-types":               "/api/iot/device-types",                 // IOTController.GetDeviceTypes
		"devices":                    "/api/iot/devices",                      // IOTController.GetDevices
		"device":                     "/api/iot/devices/device-guid",          // IOTController.GetDeviceByGuid
		"device-create":              "/api/iot/devices",                      // IOTController.CreateDevice
		"device-update":              "/api/iot/devices/device-guid",          // IOTController.UpdateDeviceByGuid
		"device-keys":                "/api/iot/devices/device-guid/api-keys", // IOTController.GetDeviceApiKeysByGuid
		"device-key-issue":           "/api/iot/devices/device-guid/api-keys", // IOTController.CreateDeviceApiKeyByGuid
		"device-key-revoke":          "/api/iot/api-keys/key-guid",            // IOTController.RevokeApiKeyByGuid
		"command-definitions":        "/api/iot/commands/definitions",
		"command-definition-history": "/api/iot/commands/definitions/history",
		"command-definition-create":  "/api/iot/commands/definitions",
		"command-definition-update":  "/api/iot/commands/definitions/definition-guid",
		"command-control":            "/api/iot/commands/control",
		"command-control-update":     "/api/iot/commands/control",
		"command-preview":            "/api/iot/commands/requests/preview",
		"command-requests":           "/api/iot/commands/requests",
		"command-request":            "/api/iot/commands/requests/request-guid",
		"command-confirm":            "/api/iot/commands/requests/request-guid/confirm",
		"command-approve":            "/api/iot/commands/requests/request-guid/approve",
		"command-reject":             "/api/iot/commands/requests/request-guid/reject",
		"command-cancel":             "/api/iot/commands/requests/request-guid/cancel",
		"command-monitoring":         "/api/iot/commands/monitoring",
	}

	for _, action := range domain.Actions {
		want, ok := expected[action.Name]
		if !ok {
			t.Errorf("iot action %q has no pinned route; add it to this table", action.Name)
			continue
		}
		args := map[string]any{"definitionGuid": "definition-guid", "requestGuid": "request-guid",
			"deviceGuid": "device-guid", "apiKeyGuid": "key-guid"}
		if path, _ := buildRESTPath(domain, action, args); path != want {
			t.Errorf("%s routes to %q, want %q", action.Name, path, want)
		}
		delete(expected, action.Name)
	}
	for missing := range expected {
		t.Errorf("missing iot action %q", missing)
	}
}

func TestIoTReadActionsAreGets(t *testing.T) {
	for _, name := range []string{"status", "monitoring", "telemetry", "rules"} {
		if method := findIoTAction(t, name).HTTPMethod; method != "GET" {
			t.Errorf("%s HTTPMethod = %q, want GET", name, method)
		}
	}
}

func TestIoTCommandMutationsMirrorIdempotencyHeaderInBody(t *testing.T) {
	for _, actionName := range []string{
		"command-definition-create",
		"command-definition-update",
		"command-control-update",
		"command-preview",
		"command-confirm",
		"command-approve",
		"command-reject",
		"command-cancel",
	} {
		action := findIoTAction(t, actionName)
		var found bool
		for _, flag := range action.Flags {
			if flag.Name != "idempotency-key" {
				continue
			}
			found = true
			if flag.HeaderName != "Idempotency-Key" ||
				flag.BodyName != "idempotencyKey" ||
				!flag.MirrorHeaderInBody ||
				!flag.Required {
				t.Fatalf("%s idempotency flag is not governed: %+v", actionName, flag)
			}
		}
		if !found {
			t.Fatalf("%s is missing the idempotency flag", actionName)
		}
	}
}

func TestIoTCommandPreviewUsesReviewedJSONFileAndConcurrency(t *testing.T) {
	action := findIoTAction(t, "command-preview")
	flags := make(map[string]FlagDef)
	for _, flag := range action.Flags {
		flags[flag.Name] = flag
	}
	parameters := flags["parameters-file"]
	if !parameters.JSONFile || parameters.BodyName != "parameters" || !parameters.Required {
		t.Fatalf("parameters-file must carry a required reviewed JSON object: %+v", parameters)
	}
	expected := flags["expected-updated-at"]
	if expected.BodyName != "expectedUpdatedAt" || !expected.Required {
		t.Fatalf("preview must bind the exact device version: %+v", expected)
	}
}

func TestIoTCommandDefinitionListsSeparateActivePreviewFromConfigureHistory(t *testing.T) {
	active := findIoTAction(t, "command-definitions")
	history := findIoTAction(t, "command-definition-history")
	if active.RESTPath != "definitions" {
		t.Fatalf("active list path = %q, want definitions", active.RESTPath)
	}
	if history.RESTPath != "definitions/history" {
		t.Fatalf("history path = %q, want definitions/history", history.RESTPath)
	}
	for _, flag := range active.Flags {
		if flag.Name == "include-inactive" {
			t.Fatal("active preview list must not expose include-inactive")
		}
	}
}

func TestIoTCommandDefinitionUpdateExposesLifecycleEvidence(t *testing.T) {
	action := findIoTAction(t, "command-definition-update")
	flags := make(map[string]FlagDef)
	for _, flag := range action.Flags {
		flags[flag.Name] = flag
	}
	if flags["is-active"].Default != true || flags["is-active"].BodyName != "isActive" {
		t.Fatalf("is-active must preserve active revisions by default: %+v", flags["is-active"])
	}
	if flags["lifecycle-reason"].BodyName != "lifecycleReason" {
		t.Fatalf("lifecycle-reason must bind durable lifecycle evidence: %+v", flags["lifecycle-reason"])
	}
}

func findIoTAction(t *testing.T, name string) Action {
	t.Helper()
	domain := findDomain("iot")
	if domain == nil {
		t.Fatal("expected iot domain to be registered")
	}
	for _, action := range domain.Actions {
		if action.Name == name {
			return action
		}
	}
	t.Fatalf("expected iot action %q", name)
	return Action{}
}

// The REST device model always applies isActive, so an update without --is-active would
// silently deactivate the device; the flag must be required, never defaulted.
func TestIoTDeviceUpdateRequiresIsActiveAndName(t *testing.T) {
	action := findIoTAction(t, "device-update")
	required := map[string]bool{}
	for _, flag := range action.Flags {
		if flag.Required {
			required[flag.Name] = true
		}
		if flag.Default != nil {
			t.Errorf("device-update flag %q must not have a default", flag.Name)
		}
	}
	for _, name := range []string{"name", "is-active"} {
		if !required[name] {
			t.Errorf("device-update flag %q must be required", name)
		}
	}
	if action.HTTPMethod != "PUT" {
		t.Errorf("device-update HTTPMethod = %q, want PUT", action.HTTPMethod)
	}
}

func TestIoTDeviceKeyActionsUseGuidArgsAndTheRestMethods(t *testing.T) {
	cases := map[string]string{"device-keys": "GET", "device-key-issue": "POST", "device-key-revoke": "DELETE"}
	for name, method := range cases {
		action := findIoTAction(t, name)
		if action.HTTPMethod != method {
			t.Errorf("%s HTTPMethod = %q, want %q", name, action.HTTPMethod, method)
		}
		if len(action.Args) != 1 || action.Args[0].Type != "uuid" || !action.Args[0].Required {
			t.Errorf("%s must take exactly one required GUID argument", name)
		}
	}
}
