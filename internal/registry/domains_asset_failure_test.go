package registry

import "testing"

func TestAssetFailureOpenActionMirrorsAssistantSafeMCPRead(t *testing.T) {
	domain := findDomain("asset-failure")
	if domain == nil {
		t.Fatal("asset-failure domain is not registered")
	}

	var action *Action
	for index := range domain.Actions {
		if domain.Actions[index].Name == "open" {
			action = &domain.Actions[index]
			break
		}
	}
	if action == nil {
		t.Fatal("asset-failure open action is not registered")
	}
	if action.ToolName != "UteamupAssetfailureGetOpen" ||
		action.HTTPMethod != "GET" || action.RESTPath != "open" {
		t.Fatalf(
			"asset-failure open action = tool %q, method %q, path %q",
			action.ToolName,
			action.HTTPMethod,
			action.RESTPath,
		)
	}
	if len(action.Args) != 0 || len(action.Flags) != 0 {
		t.Fatalf("asset-failure open action unexpectedly accepts identifiers: %+v", action)
	}
}

func TestAssetFailureSeverityActionAcceptsOnlySeverityFilter(t *testing.T) {
	domain := findDomain("asset-failure")
	if domain == nil {
		t.Fatal("asset-failure domain is not registered")
	}
	var action *Action
	for index := range domain.Actions {
		if domain.Actions[index].Name == "by-severity" {
			action = &domain.Actions[index]
			break
		}
	}
	if action == nil || action.ToolName != "UteamupAssetfailureGetBySeverity" ||
		len(action.Args) != 0 || len(action.Flags) != 1 ||
		action.Flags[0].Name != "severity" {
		t.Fatalf("unexpected asset-failure severity action: %+v", action)
	}
}

func TestAssetFailureEvidencePackUsesTheGuidRouteAndWindowQueryFlags(t *testing.T) {
	domain := findDomain("asset-failure")
	if domain == nil {
		t.Fatal("asset-failure domain is not registered")
	}
	action := findAction(domain, "evidence-pack")
	if action == nil {
		t.Fatal("asset-failure evidence-pack action is not registered")
	}
	if action.ToolName != "UteamupAssetFailureEvidencePack" || action.HTTPMethod != "GET" ||
		action.RESTPath != "by-guid/{failureGuid}/evidence-pack" || len(action.Args) != 0 {
		t.Fatalf("unexpected asset-failure evidence-pack action: %+v", action)
	}
	if err := validateActionDefinition(*action); err != nil {
		t.Fatalf("evidence-pack action definition is invalid: %v", err)
	}

	flags := flagsToMap(action.Flags)
	if len(flags) != 3 {
		t.Fatalf("evidence-pack flags = %+v, want guid, hours-before and hours-after", action.Flags)
	}
	guid, ok := flags["guid"]
	if !ok || guid.BodyName != "failureGuid" || guid.Type != "non-empty-uuid" || !guid.Required ||
		guid.QueryName != "" {
		t.Fatalf("--guid must be a required non-empty GUID that fills {failureGuid}: %+v", guid)
	}
	for name, query := range map[string]string{"hours-before": "hoursBefore", "hours-after": "hoursAfter"} {
		flag, ok := flags[name]
		if !ok || flag.QueryName != query || flag.Type != "int" || flag.Required || flag.Default != nil {
			t.Fatalf("--%s must be an optional int query flag named %q: %+v", name, query, flag)
		}
	}

	const failureGuid = "3f2b9c1d-7e4a-4b5c-9d6e-1a2b3c4d5e6f"
	path, consumed := buildRESTPath(domain, *action, map[string]any{"failureGuid": failureGuid})
	if path != "/api/assetfailure/by-guid/"+failureGuid+"/evidence-pack" {
		t.Fatalf("evidence-pack path = %q", path)
	}
	if len(consumed) != 1 || consumed[0] != "failureGuid" {
		t.Fatalf("consumed = %v, want [failureGuid]", consumed)
	}
	withWindow := appendQueryParameters(path, map[string]any{"hoursBefore": 12, "hoursAfter": 3})
	if withWindow != path+"?hoursAfter=3&hoursBefore=12" {
		t.Fatalf("evidence-pack window query = %q", withWindow)
	}
}

func TestAssetFailureIdentityActionsUsePublicGUIDs(t *testing.T) {
	domain := findDomain("asset-failure")
	if domain == nil {
		t.Fatal("asset-failure domain is not registered")
	}

	expected := map[string]string{
		"by-asset":   "assetGuid",
		"statistics": "assetGuid",
		"get":        "failureGuid",
		"update":     "failureGuid",
		"delete":     "failureGuid",
		"classify":   "failureGuid",
	}
	for actionName, argumentName := range expected {
		var action *Action
		for index := range domain.Actions {
			if domain.Actions[index].Name == actionName {
				action = &domain.Actions[index]
				break
			}
		}
		if action == nil {
			t.Fatalf("asset-failure %s action is not registered", actionName)
		}
		if len(action.Args) != 1 || action.Args[0].Name != argumentName ||
			action.Args[0].Type != "uuid" || !action.Args[0].Required {
			t.Fatalf("asset-failure %s must use required %s UUID: %+v", actionName, argumentName, action.Args)
		}
	}
}
