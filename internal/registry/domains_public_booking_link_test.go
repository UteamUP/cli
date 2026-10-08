package registry

import "testing"

func TestPublicBookingLinkDomainMirrorsBackendToolsAndGuidRoutes(t *testing.T) {
	domain := findDomain("public-booking-link")
	if domain == nil {
		t.Fatal("public-booking-link domain is not registered")
	}
	if domain.APIPath != "/api/publicbookinglinks" {
		t.Fatalf("APIPath = %q, want /api/publicbookinglinks", domain.APIPath)
	}

	expected := map[string]struct {
		tool   string
		method string
		path   string
	}{
		"list":   {"UteamupPublicBookingLinkList", "GET", ""},
		"create": {"UteamupPublicBookingLinkCreate", "POST", ""},
		"update": {"UteamupPublicBookingLinkUpdate", "PUT", "{linkGuid}"},
	}
	for actionName, want := range expected {
		action := findAction(domain, actionName)
		if action == nil {
			t.Fatalf("action %q is not registered", actionName)
		}
		if action.ToolName != want.tool || action.HTTPMethod != want.method || action.RESTPath != want.path {
			t.Errorf("%s = tool %q method %q path %q, want %q %q %q",
				actionName, action.ToolName, action.HTTPMethod, action.RESTPath, want.tool, want.method, want.path)
		}
	}

	update := findAction(domain, "update")
	if len(update.Args) != 1 || update.Args[0].Name != "linkGuid" || update.Args[0].Type != "uuid" || !update.Args[0].Required {
		t.Fatalf("update must take one required link GUID: %+v", update.Args)
	}
}

func TestPublicBookingLinkSaveFlagsMirrorTheUpsertModel(t *testing.T) {
	domain := findDomain("public-booking-link")
	for _, actionName := range []string{"create", "update"} {
		expected := map[string]struct {
			body     string
			kind     string
			required bool
			def      any
		}{
			"name":                     {"name", "string", true, nil},
			"description":              {"description", "string", false, nil},
			"is-active":                {"isActive", "bool", false, true},
			"allowed-resource-types":   {"allowedResourceTypes", "stringSlice", true, nil},
			"requires-approval":        {"requiresApproval", "bool", false, true},
			"minimum-lead-time-hours":  {"minimumLeadTimeHours", "int", false, 2},
			"booking-horizon-days":     {"bookingHorizonDays", "int", false, 30},
			"max-duration-minutes":     {"maxDurationMinutes", "int", false, 480},
			"slot-minutes":             {"slotMinutes", "int", false, 30},
			"approval-deadline-hours":  {"approvalDeadlineHours", "int", false, 48},
			"max-submissions-per-hour": {"maxSubmissionsPerHour", "int", false, 20},
		}
		for _, flag := range findAction(domain, actionName).Flags {
			want, ok := expected[flag.Name]
			if !ok {
				t.Errorf("%s has unexpected flag %+v", actionName, flag)
				continue
			}
			if flag.BodyName != want.body || flag.Type != want.kind || flag.Required != want.required || flag.Default != want.def {
				t.Errorf("%s --%s = %+v, want body %q type %q required %v default %v",
					actionName, flag.Name, flag, want.body, want.kind, want.required, want.def)
			}
			delete(expected, flag.Name)
		}
		if len(expected) != 0 {
			t.Errorf("%s is missing flags: %v", actionName, expected)
		}
	}
}
