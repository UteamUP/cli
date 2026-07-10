package registry

import "testing"

// findDomainByName and findDomainAction are shared helpers already defined in the
// package's other domain tests (knowledgespace / projectplanning). Reused here.

func actionFlagByName(t *testing.T, action *Action, flagName string) *FlagDef {
	t.Helper()
	for i := range action.Flags {
		if action.Flags[i].Name == flagName {
			return &action.Flags[i]
		}
	}
	t.Fatalf("%q action must expose a %q flag", action.Name, flagName)
	return nil
}

func TestPartCrossReferenceActionsWired(t *testing.T) {
	cases := []struct {
		name, tool, method, rest string
		argCount                 int
	}{
		{"crossrefs-list", "UteamupPartListCrossReferences", "", "by-guid/{guid}/cross-references", 1},
		{"crossrefs-add", "UteamupPartAddCrossReference", "POST", "by-guid/{guid}/cross-references", 1},
		{"crossrefs-delete", "UteamupPartDeleteCrossReference", "DELETE", "by-guid/{guid}/cross-references/{refGuid}", 2},
	}
	for _, tc := range cases {
		a := findDomainAction(t, "part", tc.name)
		if a.ToolName != tc.tool || a.HTTPMethod != tc.method || a.RESTPath != tc.rest {
			t.Errorf("%s miswired: tool=%q method=%q rest=%q", tc.name, a.ToolName, a.HTTPMethod, a.RESTPath)
		}
		if len(a.Args) != tc.argCount {
			t.Errorf("%s expected %d positional args, got %+v", tc.name, tc.argCount, a.Args)
		}
		for _, arg := range a.Args {
			if arg.Type != "string" || !arg.Required {
				t.Errorf("%s arg %q must be a required string GUID, got %+v", tc.name, arg.Name, arg)
			}
		}
	}

	add := findDomainAction(t, "part", "crossrefs-add")
	for _, name := range []string{"reference-type", "reference-number"} {
		if f := actionFlagByName(t, add, name); !f.Required || f.Type != "string" {
			t.Errorf("crossrefs-add flag %q must be a required string, got %+v", name, f)
		}
	}
	for _, name := range []string{"manufacturer", "related-part-guid", "notes"} {
		if f := actionFlagByName(t, add, name); f.Required {
			t.Errorf("crossrefs-add flag %q must be optional", name)
		}
	}
}

func TestPartVendorCatalogActionsWired(t *testing.T) {
	list := findDomainAction(t, "part", "vendor-catalog-list")
	if list.ToolName != "UteamupPartListVendorCatalog" || list.RESTPath != "by-guid/{guid}/vendor-catalog" || list.HTTPMethod != "" {
		t.Errorf("vendor-catalog-list miswired: %+v", list)
	}

	upsert := findDomainAction(t, "part", "vendor-catalog-upsert")
	if upsert.ToolName != "UteamupPartUpsertVendorCatalog" || upsert.HTTPMethod != "POST" || upsert.RESTPath != "by-guid/{guid}/vendor-catalog" {
		t.Errorf("vendor-catalog-upsert miswired: %+v", upsert)
	}
	for _, name := range []string{"vendor-guid", "vendor-part-number"} {
		if f := actionFlagByName(t, upsert, name); !f.Required || f.Type != "string" {
			t.Errorf("vendor-catalog-upsert flag %q must be a required string, got %+v", name, f)
		}
	}
	if f := actionFlagByName(t, upsert, "unit-cost"); f.Type != "float" {
		t.Errorf("vendor-catalog-upsert unit-cost must be a float flag, got %+v", f)
	}
	for _, name := range []string{"minimum-order-quantity", "lead-time-days"} {
		if f := actionFlagByName(t, upsert, name); f.Type != "int" {
			t.Errorf("vendor-catalog-upsert flag %q must be an int flag, got %+v", name, f)
		}
	}
	pref := actionFlagByName(t, upsert, "preferred")
	if pref.Type != "bool" || pref.BodyName != "isPreferred" || pref.Default != false {
		t.Errorf("vendor-catalog-upsert preferred must be a bool→isPreferred flag defaulting to false, got %+v", pref)
	}

	del := findDomainAction(t, "part", "vendor-catalog-delete")
	if del.HTTPMethod != "DELETE" || del.RESTPath != "by-guid/{guid}/vendor-catalog/{entryGuid}" || len(del.Args) != 2 {
		t.Errorf("vendor-catalog-delete miswired: %+v", del)
	}
}

func TestPartKitActionsWired(t *testing.T) {
	cases := []struct {
		name, tool, method, rest string
	}{
		{"kit-list", "UteamupPartListKitComponents", "", "by-guid/{guid}/kit-components"},
		{"kit-add", "UteamupPartAddKitComponent", "POST", "by-guid/{guid}/kit-components"},
		{"kit-delete", "UteamupPartDeleteKitComponent", "DELETE", "by-guid/{guid}/kit-components/{componentGuid}"},
		{"kit-availability", "UteamupPartGetKitAvailability", "", "by-guid/{guid}/kit-availability"},
	}
	for _, tc := range cases {
		a := findDomainAction(t, "part", tc.name)
		if a.ToolName != tc.tool || a.HTTPMethod != tc.method || a.RESTPath != tc.rest {
			t.Errorf("%s miswired: tool=%q method=%q rest=%q", tc.name, a.ToolName, a.HTTPMethod, a.RESTPath)
		}
	}

	add := findDomainAction(t, "part", "kit-add")
	if f := actionFlagByName(t, add, "component-part-guid"); !f.Required || f.Type != "string" {
		t.Errorf("kit-add component-part-guid must be a required string, got %+v", f)
	}
	// Float default rule: a float flag default (if any) must be a float literal; here
	// quantity is required with no default, so no panic risk.
	if f := actionFlagByName(t, add, "quantity"); !f.Required || f.Type != "float" || f.Default != nil {
		t.Errorf("kit-add quantity must be a required float flag with no default, got %+v", f)
	}
}

func TestPartCompatibilityActionsWired(t *testing.T) {
	cases := []struct {
		name, tool, method, rest string
	}{
		{"compat-list", "UteamupPartListCompatibility", "", "by-guid/{guid}/compatibility"},
		{"compat-add", "UteamupPartAddCompatibility", "POST", "by-guid/{guid}/compatibility"},
		{"compat-delete", "UteamupPartDeleteCompatibility", "DELETE", "by-guid/{guid}/compatibility/{compatibilityGuid}"},
	}
	for _, tc := range cases {
		a := findDomainAction(t, "part", tc.name)
		if a.ToolName != tc.tool || a.HTTPMethod != tc.method || a.RESTPath != tc.rest {
			t.Errorf("%s miswired: tool=%q method=%q rest=%q", tc.name, a.ToolName, a.HTTPMethod, a.RESTPath)
		}
	}
	add := findDomainAction(t, "part", "compat-add")
	if f := actionFlagByName(t, add, "asset-type-guid"); !f.Required || f.Type != "string" {
		t.Errorf("compat-add asset-type-guid must be a required string, got %+v", f)
	}
}

func TestVendorCatalogActionWired(t *testing.T) {
	a := findDomainAction(t, "vendor", "catalog")
	if a.ToolName != "UteamupVendorGetCatalog" || a.RESTPath != "by-guid/{guid}/catalog" || a.HTTPMethod != "" {
		t.Errorf("vendor catalog miswired: %+v", a)
	}
	if len(a.Args) != 1 || a.Args[0].Name != "guid" || !a.Args[0].Required || a.Args[0].Type != "string" {
		t.Errorf("vendor catalog expected single required string arg 'guid', got %+v", a.Args)
	}
}

func TestAssetTypeCompatiblePartsWired(t *testing.T) {
	d := findDomainByName(t, "asset-type")
	// The controller routes at api/asset-type (hyphenated); the derived path would strip
	// the hyphen to /api/assettype, so the domain must pin the base path explicitly.
	if d.APIPath != "/api/asset-type" {
		t.Errorf("asset-type APIPath = %q, want /api/asset-type (backend route is hyphenated)", d.APIPath)
	}

	a := findDomainAction(t, "asset-type", "compatible-parts")
	if a.ToolName != "UteamupAssetTypeListCompatibleParts" || a.RESTPath != "by-guid/{guid}/compatible-parts" {
		t.Errorf("compatible-parts miswired: %+v", a)
	}

	// The full URL must resolve to the hyphenated controller route.
	url, consumed := buildRESTPath(d, *a, map[string]any{"guid": "11111111-1111-1111-1111-111111111111"})
	if url != "/api/asset-type/by-guid/11111111-1111-1111-1111-111111111111/compatible-parts" {
		t.Errorf("compatible-parts URL = %q, want the hyphenated api/asset-type route", url)
	}
	if len(consumed) != 1 || consumed[0] != "guid" {
		t.Errorf("compatible-parts should consume the guid placeholder, consumed=%+v", consumed)
	}
}

func TestUomDomainWired(t *testing.T) {
	d := findDomainByName(t, "uom")
	if len(d.Aliases) != 1 || d.Aliases[0] != "units-of-measure" {
		t.Errorf("uom aliases = %+v, want [units-of-measure]", d.Aliases)
	}

	list := findDomainAction(t, "uom", "list")
	if list.ToolName != "UteamupUomList" || list.HTTPMethod != "" {
		t.Errorf("uom list miswired: %+v", list)
	}
	if url, _ := buildRESTPath(d, *list, map[string]any{}); url != "/api/uom" {
		t.Errorf("uom list URL = %q, want /api/uom", url)
	}

	create := findDomainAction(t, "uom", "create")
	if create.ToolName != "UteamupUomCreate" {
		t.Errorf("uom create ToolName = %q, want UteamupUomCreate", create.ToolName)
	}
	for _, name := range []string{"code", "name"} {
		if f := actionFlagByName(t, create, name); !f.Required || f.Type != "string" {
			t.Errorf("uom create flag %q must be a required string, got %+v", name, f)
		}
	}
	if url, _ := buildRESTPath(d, *create, map[string]any{}); url != "/api/uom" {
		t.Errorf("uom create URL = %q, want /api/uom (POST)", url)
	}

	del := findDomainAction(t, "uom", "delete")
	if del.ToolName != "UteamupUomDelete" || del.RESTPath != "{guid}" {
		t.Errorf("uom delete miswired: %+v", del)
	}
	if url, consumed := buildRESTPath(d, *del, map[string]any{"guid": "22222222-2222-2222-2222-222222222222"}); url != "/api/uom/22222222-2222-2222-2222-222222222222" || len(consumed) != 1 {
		t.Errorf("uom delete URL = %q consumed=%+v, want /api/uom/{guid}", url, consumed)
	}
}

func TestStockSuggestedOrdersActionsWired(t *testing.T) {
	get := findDomainAction(t, "stock", "suggested-orders")
	if get.ToolName != "UteamupStockGetSuggestedOrders" || get.RESTPath != "suggested-orders" || get.HTTPMethod != "" {
		t.Errorf("suggested-orders miswired: %+v", get)
	}

	createPo := findDomainAction(t, "stock", "suggested-orders-create-po")
	if createPo.ToolName != "UteamupStockCreateSuggestedOrderPo" || createPo.HTTPMethod != "POST" || createPo.RESTPath != "suggested-orders/create-po" {
		t.Errorf("suggested-orders-create-po miswired: %+v", createPo)
	}
	// vendor-guid is optional (omitting it confirms the no-resolvable-vendor group) and
	// maps to the body field vendorGuid.
	vg := actionFlagByName(t, createPo, "vendor-guid")
	if vg.Required || vg.Type != "string" || vg.BodyName != "vendorGuid" {
		t.Errorf("suggested-orders-create-po vendor-guid must be an optional string → vendorGuid, got %+v", vg)
	}
}

func TestPartCrossRefDeleteConsumesBothGuids(t *testing.T) {
	d := findDomainByName(t, "part")
	a := findDomainAction(t, "part", "crossrefs-delete")
	url, consumed := buildRESTPath(d, *a, map[string]any{
		"guid":    "11111111-1111-1111-1111-111111111111",
		"refGuid": "22222222-2222-2222-2222-222222222222",
	})
	want := "/api/part/by-guid/11111111-1111-1111-1111-111111111111/cross-references/22222222-2222-2222-2222-222222222222"
	if url != want {
		t.Errorf("crossrefs-delete URL = %q, want %q", url, want)
	}
	if len(consumed) != 2 {
		t.Errorf("crossrefs-delete should consume both path placeholders, consumed=%+v", consumed)
	}
}
