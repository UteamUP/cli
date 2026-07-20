package registry

import "testing"

import "strings"

func TestIoTDomainMirrorsMCPTools(t *testing.T) {
	domain := findDomain("iot")
	if domain == nil {
		t.Fatal("expected iot domain to be registered")
	}
	expected := map[string]string{
		"status":                    "UteamupIoTEnvironmentStatus",
		"monitoring":                "UteamupIoTMonitoringDashboard",
		"telemetry":                 "UteamupIoTTelemetryPoints",
		"rules":                     "UteamupIoTRulesList",
		"command-definitions":       "UteamupIoTCommandDefinitionsList",
		"command-definition-create": "UteamupIoTCommandDefinitionCreate",
		"command-definition-update": "UteamupIoTCommandDefinitionUpdate",
		"command-control":           "UteamupIoTCommandControlGet",
		"command-control-update":    "UteamupIoTCommandControlUpdate",
		"command-preview":           "UteamupIoTCommandRequestPreview",
		"command-requests":          "UteamupIoTCommandRequestsList",
		"command-request":           "UteamupIoTCommandRequestGet",
		"command-confirm":           "UteamupIoTCommandRequestConfirm",
		"command-approve":           "UteamupIoTCommandRequestApprove",
		"command-reject":            "UteamupIoTCommandRequestReject",
		"command-cancel":            "UteamupIoTCommandRequestCancel",
		"command-monitoring":        "UteamupIoTCommandMonitoringGet",
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
