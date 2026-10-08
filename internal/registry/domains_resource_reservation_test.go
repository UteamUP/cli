package registry

import "testing"

func TestResourceReservationDomainMirrorsBackendToolsAndGuidRoutes(t *testing.T) {
	domain := findDomain("resource-reservation")
	if domain == nil {
		t.Fatal("resource-reservation domain is not registered")
	}
	if domain.APIPath != "/api/resourcereservations" {
		t.Fatalf("APIPath = %q, want /api/resourcereservations", domain.APIPath)
	}

	expected := map[string]struct {
		tool   string
		method string
		path   string
	}{
		"list":     {"UteamupResourceReservationList", "GET", ""},
		"get":      {"UteamupResourceReservationGet", "GET", "{reservationGuid}"},
		"create":   {"UteamupResourceReservationCreate", "POST", ""},
		"approve":  {"UteamupResourceReservationDecide", "POST", "{reservationGuid}/approve"},
		"reject":   {"UteamupResourceReservationDecide", "POST", "{reservationGuid}/reject"},
		"cancel":   {"UteamupResourceReservationDecide", "POST", "{reservationGuid}/cancel"},
		"checkout": {"UteamupResourceReservationDecide", "POST", "{reservationGuid}/checkout"},
		"return":   {"UteamupResourceReservationDecide", "POST", "{reservationGuid}/return"},
	}
	if len(domain.Actions) != len(expected) {
		t.Errorf("registered %d actions, want %d", len(domain.Actions), len(expected))
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
		for _, arg := range action.Args {
			if arg.Name != "reservationGuid" || arg.Type != "uuid" || !arg.Required {
				t.Errorf("%s must take only a required reservation GUID: %+v", actionName, arg)
			}
		}
	}
}

func TestResourceReservationDecisionsOnlyCarryANoteWhereTheBackendReadsOne(t *testing.T) {
	domain := findDomain("resource-reservation")
	for _, name := range []string{"approve", "reject", "cancel"} {
		action := findAction(domain, name)
		if len(action.Flags) != 1 || action.Flags[0].Name != "note" || action.Flags[0].BodyName != "note" || action.Flags[0].Type != "string" {
			t.Errorf("%s must expose only an optional --note body field: %+v", name, action.Flags)
		}
	}
	for _, name := range []string{"checkout", "return"} {
		if flags := findAction(domain, name).Flags; len(flags) != 0 {
			t.Errorf("%s takes no body, got flags %+v", name, flags)
		}
	}
}

func TestResourceReservationCreateUsesGuidsAndARequiredRetryHeader(t *testing.T) {
	action := findAction(findDomain("resource-reservation"), "create")
	expected := map[string]struct {
		body     string
		kind     string
		required bool
		def      any
	}{
		"resource-type":      {"resourceType", "int", true, nil},
		"source-guid":        {"sourceGuid", "uuid", false, nil},
		"resource-guid":      {"resourceGuid", "uuid", false, nil},
		"start-utc":          {"startUtc", "string", true, nil},
		"end-utc":            {"endUtc", "string", true, nil},
		"capacity-requested": {"capacityRequested", "float", false, 1.0},
		"purpose":            {"purpose", "string", false, nil},
		"notes":              {"notes", "string", false, nil},
	}
	foundKey := false
	for _, flag := range action.Flags {
		if flag.Name == "idempotency-key" {
			foundKey = true
			if flag.HeaderName != "Idempotency-Key" || flag.BodyName != "" || flag.Type != "uuid" || !flag.Required {
				t.Errorf("idempotency-key must be a required GUID retry header: %+v", flag)
			}
			continue
		}
		want, ok := expected[flag.Name]
		if !ok {
			t.Errorf("unexpected create flag %+v", flag)
			continue
		}
		if flag.BodyName != want.body || flag.Type != want.kind || flag.Required != want.required || flag.Default != want.def {
			t.Errorf("%s = %+v, want body %q type %q required %v default %v", flag.Name, flag, want.body, want.kind, want.required, want.def)
		}
		delete(expected, flag.Name)
	}
	if !foundKey {
		t.Error("create is missing --idempotency-key")
	}
	if len(expected) != 0 {
		t.Fatalf("missing create flags: %v", expected)
	}
	for _, flag := range action.Flags {
		if flag.Name == "capacity-requested" {
			if _, isFloat := flag.Default.(float64); !isFloat {
				t.Fatalf("capacity-requested default must be a float literal, got %T", flag.Default)
			}
		}
	}
}

func TestResourceReservationListFiltersUseQueryParameters(t *testing.T) {
	action := findAction(findDomain("resource-reservation"), "list")
	expected := map[string]struct {
		query string
		kind  string
	}{
		"statuses":      {"statuses", "stringSlice"},
		"from-utc":      {"fromUtc", "string"},
		"to-utc":        {"toUtc", "string"},
		"resource-guid": {"resourceGuid", "uuid"},
		"mine":          {"mine", "bool"},
		"page":          {"page", "int"},
		"page-size":     {"pageSize", "int"},
	}
	for _, flag := range action.Flags {
		want, ok := expected[flag.Name]
		if !ok {
			t.Errorf("unexpected list flag %+v", flag)
			continue
		}
		if flag.QueryName != want.query || flag.Type != want.kind || flag.Required {
			t.Errorf("%s = %+v, want optional query %q type %q", flag.Name, flag, want.query, want.kind)
		}
		delete(expected, flag.Name)
	}
	if len(expected) != 0 {
		t.Fatalf("missing list filters: %v", expected)
	}
}
