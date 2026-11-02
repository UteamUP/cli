package registry

import "testing"

func catalogAction(t *testing.T, domainName, actionName string) (*Domain, Action) {
	t.Helper()
	for _, domain := range DefaultRegistry.Domains() {
		if domain.Name != domainName {
			continue
		}
		for _, action := range domain.Actions {
			if action.Name == actionName {
				return domain, action
			}
		}
	}
	t.Fatalf("missing %s %s", domainName, actionName)
	return nil, Action{}
}

func TestCatalogDocumentActionsUsePublicGuidsAndExactExistingRoutes(t *testing.T) {
	item := "11111111-1111-4111-8111-111111111111"
	document := "22222222-2222-4222-8222-222222222222"
	for _, kind := range []string{"part", "tool", "chemical"} {
		for _, name := range []string{"add-document", "remove-document", "set-image-from-document"} {
			domain, action := catalogAction(t, kind, name)
			if len(action.Args) != 2 || !action.RejectExtraArgs {
				t.Fatalf("unbounded or missing arguments: %+v", action)
			}
			for _, arg := range action.Args {
				if arg.Type != "non-empty-uuid" || !arg.Required {
					t.Fatalf("public identifier must be a required GUID: %+v", arg)
				}
			}
			path, consumed := buildRESTPath(domain, action, map[string]any{kind + "Guid": item, "documentGuid": document})
			want := "/api/" + kind + "/by-guid/" + item
			if name == "set-image-from-document" {
				want += "/image/from-document/" + document
			} else {
				want += "/documents/"
				if kind != "part" {
					want += "by-guid/"
				}
				want += document
			}
			if path != want || len(consumed) != 2 {
				t.Fatalf("%s %s: path %q consumed %v, want %q", kind, name, path, consumed, want)
			}
		}
	}
}

func TestCatalogImageUploadUsesMcpAndRedactsTheBytes(t *testing.T) {
	for _, kind := range []string{"part", "tool", "chemical"} {
		_, action := catalogAction(t, kind, "upload-image")
		if !action.MCPOnly || len(action.Args) != 1 || action.Args[0].Name != kind+"Guid" {
			t.Fatalf("incorrect upload surface: %+v", action)
		}
		found := false
		for _, flag := range action.Flags {
			if flag.Name == "base64-content" {
				found = flag.BodyName == "base64Content" && flag.Required && flag.Sensitive
			}
		}
		if !found {
			t.Fatal("upload bytes must be required and redacted in logs")
		}
	}
}

func TestCatalogGuidAdditionsRetainLegacyActions(t *testing.T) {
	for _, kind := range []string{"part", "vendor"} {
		_, legacy := catalogAction(t, kind, "get")
		_, public := catalogAction(t, kind, "get-by-guid")
		if legacy.Args[0].Name != "id" || public.Args[0].Name != kind+"Guid" || public.Args[0].Type != "non-empty-uuid" || public.HTTPMethod != "GET" {
			t.Fatalf("GUID migration must be additive: legacy %+v public %+v", legacy, public)
		}
	}
	_, patch := catalogAction(t, "part", "patch-by-guid")
	if !patch.MCPOnly || patch.Args[0].Name != "partGuid" {
		t.Fatalf("partial update must use the MCP patch, not full PUT: %+v", patch)
	}
}

func TestReporterCommentIsAnOwnReportMcpOperation(t *testing.T) {
	_, action := catalogAction(t, "bugsandfeatures", "reporter-comment")
	if !action.MCPOnly || action.ToolName != "uteamup_bug_report_add_comment" || len(action.Args) != 1 ||
		action.Args[0].Name != "reportGuid" || action.Args[0].Type != "non-empty-uuid" || !action.RejectExtraArgs {
		t.Fatalf("incorrect reporter surface: %+v", action)
	}
	if len(action.Flags) != 1 || action.Flags[0].BodyName != "message" || !action.Flags[0].Required {
		t.Fatalf("reporter may submit only the message: %+v", action.Flags)
	}
}

func TestPartCurrencyFlagMatchesEachTransportContract(t *testing.T) {
	for _, name := range []string{"create", "update", "patch-by-guid"} {
		_, action := catalogAction(t, "part", name)
		wantBodyName := "valueCurrencyGuid"
		if name == "patch-by-guid" {
			wantBodyName = "currencyGuid"
		}
		found := false
		for _, flag := range action.Flags {
			if flag.Name == "currency-guid" {
				found = flag.Type == "uuid" && flag.BodyName == wantBodyName && !flag.Required
			}
		}
		if !found {
			t.Fatalf("%s must expose optional --currency-guid as %s", name, wantBodyName)
		}
	}
}

func TestInventoryPartialOperationsRequireGuidAndExactType(t *testing.T) {
	for _, name := range []string{"get-by-guid", "patch-by-guid"} {
		_, action := catalogAction(t, "inventory", name)
		if !action.MCPOnly || !action.RejectExtraArgs || len(action.Args) != 2 ||
			action.Args[0].Name != "itemGuid" || action.Args[0].Type != "non-empty-uuid" ||
			action.Args[1].Name != "itemType" || len(action.Args[1].AllowedValues) != 3 {
			t.Fatalf("incorrect inventory identity contract: %+v", action)
		}
	}
}
