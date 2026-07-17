package registry

import (
	"strings"
	"testing"
)

func serviceSLAAction(t *testing.T, name string) (*Domain, Action) {
	t.Helper()
	domain := findDomain("service-sla")
	if domain == nil {
		t.Fatal("service-sla domain is not registered")
	}
	for _, action := range domain.Actions {
		if action.Name == name {
			return domain, action
		}
	}
	t.Fatalf("service-sla action %q is not registered", name)
	return nil, Action{}
}

func TestServiceSlaRoutesAndArgumentsAreGuidOnly(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		argName string
		path    string
	}{
		"get":        {argName: "milestoneGuid", path: "/api/service-sla-milestones/milestone-guid"},
		"initialize": {argName: "workorderGuid", path: "/api/service-sla-milestones/workorders/workorder-guid/initialize"},
		"pause":      {argName: "milestoneGuid", path: "/api/service-sla-milestones/milestone-guid/pause"},
		"resume":     {argName: "milestoneGuid", path: "/api/service-sla-milestones/milestone-guid/resume"},
		"complete":   {argName: "milestoneGuid", path: "/api/service-sla-milestones/milestone-guid/complete"},
		"cancel":     {argName: "milestoneGuid", path: "/api/service-sla-milestones/milestone-guid/cancel"},
	}
	for actionName, test := range tests {
		t.Run(actionName, func(t *testing.T) {
			domain, action := serviceSLAAction(t, actionName)
			path, consumed := buildRESTPath(domain, action, map[string]any{test.argName: strings.ReplaceAll(test.argName, "Guid", "-guid")})
			if path != test.path {
				t.Fatalf("path = %q, want %q", path, test.path)
			}
			if len(consumed) != 1 || strings.Contains(strings.ToLower(action.RESTPath), "{id}") {
				t.Fatalf("route is not GUID-only: %q consumed=%v", action.RESTPath, consumed)
			}
			if len(action.Args) != 1 || action.Args[0].Name != test.argName || action.Args[0].Type != "uuid" {
				t.Fatalf("public identity argument is not a UUID: %+v", action.Args)
			}
		})
	}
}

func TestServiceSlaActionsMirrorBackendToolsAndEvidence(t *testing.T) {
	t.Parallel()
	expectedTools := map[string]string{
		"list":       "UteamupServiceSlaMilestoneList",
		"get":        "UteamupServiceSlaMilestoneGet",
		"initialize": "UteamupServiceSlaMilestoneInitialize",
		"pause":      "UteamupServiceSlaMilestonePause",
		"resume":     "UteamupServiceSlaMilestoneResume",
		"complete":   "UteamupServiceSlaMilestoneComplete",
		"cancel":     "UteamupServiceSlaMilestoneCancel",
		"reconcile":  "UteamupServiceSlaMilestoneReconcile",
	}
	for actionName, toolName := range expectedTools {
		_, action := serviceSLAAction(t, actionName)
		if action.ToolName != toolName {
			t.Fatalf("%s tool = %q, want %q", actionName, action.ToolName, toolName)
		}
		for _, flag := range action.Flags {
			lower := strings.ToLower(flag.Name)
			if strings.Contains(lower, "tenant") || strings.Contains(lower, "user") {
				t.Fatalf("%s exposes caller-controlled scope: %+v", actionName, flag)
			}
		}
	}

	for _, actionName := range []string{"initialize", "pause", "resume", "complete", "cancel", "reconcile"} {
		_, action := serviceSLAAction(t, actionName)
		assertServiceSLAFlag(t, action, "idempotency-key", "idempotencyKey", true)
	}
	for _, actionName := range []string{"pause", "resume", "complete", "cancel"} {
		_, action := serviceSLAAction(t, actionName)
		assertServiceSLAFlag(t, action, "expected-updated-at", "expectedUpdatedAt", true)
	}
	for _, actionName := range []string{"pause", "cancel"} {
		_, action := serviceSLAAction(t, actionName)
		assertServiceSLAFlag(t, action, "reason", "reason", true)
	}
	_, list := serviceSLAAction(t, "list")
	assertServiceSLAFlag(t, list, "workorder-guid", "workorderGuid", false)
	assertServiceSLAFlag(t, list, "agreement-guid", "agreementGuid", false)
	assertServiceSLAFlag(t, list, "as-of", "asOf", false)
	_, reconcile := serviceSLAAction(t, "reconcile")
	assertServiceSLAFlag(t, reconcile, "workorder-guid", "workorderGuid", true)
	assertServiceSLAFlag(t, reconcile, "as-of", "asOf", true)
}

func assertServiceSLAFlag(t *testing.T, action Action, name, bodyName string, required bool) {
	t.Helper()
	for _, flag := range action.Flags {
		if flag.Name == name {
			if flag.BodyName != bodyName || flag.Required != required {
				t.Fatalf("%s flag = %+v, want body=%q required=%t", name, flag, bodyName, required)
			}
			return
		}
	}
	t.Fatalf("%s flag is missing from %s", name, action.Name)
}
