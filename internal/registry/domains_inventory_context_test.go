package registry

import (
	"strings"
	"testing"
)

func TestInventoryContextContracts(t *testing.T) {
	var domain *Domain
	for _, candidate := range DefaultRegistry.Domains() {
		if candidate.Name == "inventory" {
			domain = candidate
		}
	}
	if domain == nil {
		t.Fatal("inventory domain missing")
	}
	expected := map[string]struct{ method, path, tool string }{
		"context":         {"POST", "context/search", "UteamupInventoryContextSearch"},
		"handling-get":    {"GET", "{type}/{guid}/handling", "UteamupInventoryHandlingGet"},
		"handling-update": {"PUT", "{type}/{guid}/handling", "UteamupInventoryHandlingUpdate"},
		"context-summary": {"GET", "{type}/{guid}/context", "UteamupInventoryContextGet"},
	}
	for name, want := range expected {
		var action *Action
		for index := range domain.Actions {
			if domain.Actions[index].Name == name {
				action = &domain.Actions[index]
			}
		}
		if action == nil {
			t.Fatalf("%s is missing", name)
		}
		if action.HTTPMethod != want.method || action.RESTPath != want.path || action.ToolName != want.tool {
			t.Fatalf("%s does not mirror the exact Inventory route/tool: %+v", name, action)
		}
		if !action.RejectExtraArgs {
			t.Fatalf("%s accepts extra identifiers", name)
		}
		if len(action.Args) < 2 || action.Args[1].Name != "guid" || action.Args[1].Type != "non-empty-uuid" {
			t.Fatalf("%s must require the public source GUID", name)
		}
		allowed := strings.Join(action.Args[0].AllowedValues, ",")
		if strings.HasPrefix(name, "context") && allowed != "asset,assetgroup,part,tool,chemical" {
			t.Fatal("unified context lost a family")
		}
		if strings.HasPrefix(name, "handling") && strings.Contains(allowed, "assetgroup") {
			t.Fatal("group is not a handling owner")
		}
	}
}

func TestInventoryHandlingRequiresEveryReplacementFlag(t *testing.T) {
	for _, domain := range DefaultRegistry.Domains() {
		if domain.Name != "inventory" {
			continue
		}
		for _, action := range domain.Actions {
			if action.Name != "handling-update" {
				continue
			}
			count := 0
			for _, flag := range action.Flags {
				if strings.HasPrefix(flag.BodyName, "requires") {
					count++
					if !flag.Required || flag.Type != "bool" {
						t.Fatalf("%s must be an explicit reviewed boolean", flag.Name)
					}
				}
			}
			if count != 5 {
				t.Fatalf("replacement requires all five flags, got %d", count)
			}
			return
		}
	}
	t.Fatal("handling-update missing")
}
