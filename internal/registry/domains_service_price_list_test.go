package registry

import (
	"strings"
	"testing"
)

func servicePriceListAction(t *testing.T, name string) (*Domain, Action) {
	t.Helper()
	domain := findDomain("service-price-list")
	if domain == nil {
		t.Fatal("service-price-list domain is not registered")
	}
	for _, action := range domain.Actions {
		if action.Name == name {
			return domain, action
		}
	}
	t.Fatalf("service-price-list action %q is not registered", name)
	return nil, Action{}
}

func TestServicePriceListRoutesAndArgumentsAreGuidOnly(t *testing.T) {
	t.Parallel()
	for _, actionName := range []string{"get", "update"} {
		t.Run(actionName, func(t *testing.T) {
			domain, action := servicePriceListAction(t, actionName)
			path, consumed := buildRESTPath(
				domain,
				action,
				map[string]any{"priceListGuid": "price-list-guid"},
			)
			if path != "/api/service-price-lists/price-list-guid" {
				t.Fatalf("path = %q, want GUID route", path)
			}
			if len(consumed) != 1 || strings.Contains(strings.ToLower(action.RESTPath), "{id}") {
				t.Fatalf("route is not GUID-only: %q consumed=%v", action.RESTPath, consumed)
			}
			if len(action.Args) != 1 || action.Args[0].Name != "priceListGuid" || action.Args[0].Type != "uuid" {
				t.Fatalf("public identity argument is not a price-list UUID: %+v", action.Args)
			}
		})
	}
}

func TestServicePriceListActionsMirrorBackendToolsAndEvidence(t *testing.T) {
	t.Parallel()
	expectedTools := map[string]string{
		"list":   "UteamupServicePriceListList",
		"get":    "UteamupServicePriceListGet",
		"create": "UteamupServicePriceListCreate",
		"update": "UteamupServicePriceListUpdate",
	}
	for actionName, toolName := range expectedTools {
		_, action := servicePriceListAction(t, actionName)
		if action.ToolName != toolName {
			t.Fatalf("%s tool = %q, want %q", actionName, action.ToolName, toolName)
		}
	}

	for _, actionName := range []string{"create", "update"} {
		_, action := servicePriceListAction(t, actionName)
		assertServicePriceListFlag(t, action, "idempotency-key", "idempotencyKey", true)
		assertServicePriceListFlag(t, action, "items-json", "items", true)
		items := findServicePriceListFlag(action, "items-json")
		if items == nil || !items.JSONFile {
			t.Fatalf("%s items must be supplied as reviewed JSON: %+v", actionName, items)
		}
		for _, flag := range action.Flags {
			if strings.Contains(strings.ToLower(flag.Name), "tenant") || strings.Contains(strings.ToLower(flag.Name), "user") {
				t.Fatalf("%s exposes caller-controlled scope: %+v", actionName, flag)
			}
		}
	}

	_, update := servicePriceListAction(t, "update")
	assertServicePriceListFlag(t, update, "expected-updated-at", "expectedUpdatedAt", true)
	_, list := servicePriceListAction(t, "list")
	assertServicePriceListFlag(t, list, "active-only", "activeOnly", false)
	assertServicePriceListFlag(t, list, "as-of", "asOf", false)
}

func assertServicePriceListFlag(t *testing.T, action Action, name, bodyName string, required bool) {
	t.Helper()
	flag := findServicePriceListFlag(action, name)
	if flag == nil || flag.BodyName != bodyName || flag.Required != required {
		t.Fatalf("%s flag = %+v, want body=%q required=%t", name, flag, bodyName, required)
	}
}

func findServicePriceListFlag(action Action, name string) *FlagDef {
	for index := range action.Flags {
		if action.Flags[index].Name == name {
			return &action.Flags[index]
		}
	}
	return nil
}
