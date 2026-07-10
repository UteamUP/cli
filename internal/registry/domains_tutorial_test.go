package registry

import "testing"

func TestTutorialDomainWired(t *testing.T) {
	domain := findDomainByName(t, "tutorial")
	if domain.APIPath != "/api/tutorials" {
		t.Fatalf("tutorial APIPath = %q, want /api/tutorials", domain.APIPath)
	}

	actions := map[string]Action{}
	for _, action := range domain.Actions {
		actions[action.Name] = action
	}

	list := actions["list"]
	if list.ToolName != "UteamupTutorialList" || list.HTTPMethod != "GET" {
		t.Errorf("tutorial list action is not wired to the read-only MCP tool: %+v", list)
	}
	if len(list.Args) != 0 {
		t.Errorf("tutorial list must not accept identity arguments: %+v", list.Args)
	}

	get := actions["get"]
	if get.ToolName != "UteamupTutorialGet" || get.HTTPMethod != "GET" {
		t.Errorf("tutorial get action is not wired to the read-only MCP tool: %+v", get)
	}
	if get.RESTPath != "{tutorialId}" {
		t.Errorf("tutorial get RESTPath = %q, want {tutorialId}", get.RESTPath)
	}
	if len(get.Args) != 1 || get.Args[0].Name != "tutorialId" || get.Args[0].Type != "string" {
		t.Errorf("tutorial get must use one semantic string tutorialId: %+v", get.Args)
	}
}

func TestTutorialRESTPathExpansion(t *testing.T) {
	domain := findDomainByName(t, "tutorial")
	var get Action
	for _, action := range domain.Actions {
		if action.Name == "get" {
			get = action
		}
	}

	path, consumed := buildRESTPath(domain, get, map[string]any{"tutorialId": "stock.orientation"})
	if path != "/api/tutorials/stock.orientation" {
		t.Errorf("tutorial get path = %q, want /api/tutorials/stock.orientation", path)
	}
	if len(consumed) != 1 || consumed[0] != "tutorialId" {
		t.Errorf("tutorial get consumed args = %v, want [tutorialId]", consumed)
	}
}
