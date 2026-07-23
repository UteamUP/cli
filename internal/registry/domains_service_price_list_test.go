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
		"list":          "UteamupServicePriceListList",
		"get":           "UteamupServicePriceListGet",
		"create":        "UteamupServicePriceListCreate",
		"update":        "UteamupServicePriceListUpdate",
		"replacement":   "UteamupServicePriceListCreateReplacement",
		"delete":        "UteamupServicePriceListDelete",
		"archive":       "UteamupServicePriceListArchive",
		"restore":       "UteamupServicePriceListRestore",
		"preview-rules": "UteamupServicePriceListPreviewRules",
	}
	for actionName, toolName := range expectedTools {
		_, action := servicePriceListAction(t, actionName)
		if action.ToolName != toolName {
			t.Fatalf("%s tool = %q, want %q", actionName, action.ToolName, toolName)
		}
	}

	for _, actionName := range []string{"create", "update", "replacement"} {
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
	assertServicePriceListFlag(t, list, "include-archived", "includeArchived", false)
}

// The lifecycle actions exist because deactivation was the only way to clear an unused draft.
// Delete must be a real DELETE, and archive/restore must not be reachable by the same verb.
func TestServicePriceListLifecycleUsesDistinctVerbsAndRoutes(t *testing.T) {
	t.Parallel()
	_, del := servicePriceListAction(t, "delete")
	if del.HTTPMethod != "DELETE" {
		t.Fatalf("delete method = %q, want DELETE", del.HTTPMethod)
	}
	if del.RESTPath != "{priceListGuid}" {
		t.Fatalf("delete path = %q, want the version route", del.RESTPath)
	}
	for name, wantPath := range map[string]string{
		"archive": "{priceListGuid}/archive",
		"restore": "{priceListGuid}/restore",
	} {
		_, action := servicePriceListAction(t, name)
		if action.HTTPMethod != "POST" {
			t.Fatalf("%s method = %q, want POST", name, action.HTTPMethod)
		}
		if action.RESTPath != wantPath {
			t.Fatalf("%s path = %q, want %q", name, action.RESTPath, wantPath)
		}
	}
}

// The preview is read-only evidence: it must carry no idempotency key, because a key would imply
// it writes something the server has to de-duplicate.
func TestServicePriceListPreviewIsReadOnlyAndFloatTyped(t *testing.T) {
	t.Parallel()
	_, action := servicePriceListAction(t, "preview-rules")
	if action.RESTPath != "{priceListGuid}/rule-preview" {
		t.Fatalf("preview path = %q, want the version-scoped preview route", action.RESTPath)
	}
	if findServicePriceListFlag(action, "idempotency-key") != nil {
		t.Fatal("a read-only preview must not take a write idempotency key")
	}
	// Go stores an untyped 0 as int, which panics in the registry's float type assertion.
	for _, name := range []string{"labour-hours", "travel-hours", "material-quantity"} {
		flag := findServicePriceListFlag(action, name)
		if flag == nil {
			t.Fatalf("%s flag is missing", name)
		}
		if _, ok := flag.Default.(float64); !ok {
			t.Fatalf("%s default = %T, want a float literal", name, flag.Default)
		}
	}
	assertServicePriceListFlag(t, action, "period-start", "periodStart", true)
	assertServicePriceListFlag(t, action, "period-end", "periodEnd", true)
}

// Replacement must POST to the predecessor-scoped route. Reusing the plain create route leaves
// an unlinked clone overlapping the still-active original, which is the defect the server-side
// replacement transaction exists to prevent.
func TestServicePriceListReplacementPostsToThePredecessorRoute(t *testing.T) {
	t.Parallel()
	domain, action := servicePriceListAction(t, "replacement")
	if action.HTTPMethod != "POST" {
		t.Fatalf("replacement method = %q, want POST", action.HTTPMethod)
	}
	path, consumed := buildRESTPath(
		domain,
		action,
		map[string]any{"priceListGuid": "predecessor-guid"},
	)
	if path != "/api/service-price-lists/predecessor-guid/replacement" {
		t.Fatalf("path = %q, want the predecessor-scoped replacement route", path)
	}
	if len(consumed) != 1 {
		t.Fatalf("replacement consumed %v, want only the predecessor GUID", consumed)
	}
	if len(action.Args) != 1 || action.Args[0].Name != "priceListGuid" || action.Args[0].Type != "uuid" {
		t.Fatalf("predecessor identity is not a UUID argument: %+v", action.Args)
	}
	// effectiveFrom carries the retirement date, so it stays mandatory.
	assertServicePriceListFlag(t, action, "effective-from", "effectiveFrom", true)
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
