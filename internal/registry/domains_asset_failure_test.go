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
