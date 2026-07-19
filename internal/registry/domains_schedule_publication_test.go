package registry

import "testing"

func TestScheduleDraftDomainMirrorsPublishedRESTSurface(t *testing.T) {
	domain := findDomain("schedule-draft")
	if domain == nil {
		t.Fatal("schedule-draft domain is not registered")
	}
	if domain.APIPath != "/api/scheduledraft" {
		t.Fatalf("APIPath = %q, want /api/scheduledraft", domain.APIPath)
	}

	cases := []struct {
		name   string
		method string
		path   string
		args   map[string]any
	}{
		{"list", "GET", "/api/scheduledraft", map[string]any{}},
		{"get", "GET", "/api/scheduledraft/by-guid/draft-guid", map[string]any{"draftGuid": "draft-guid"}},
		{"create", "POST", "/api/scheduledraft", map[string]any{}},
		{"entry-add", "POST", "/api/scheduledraft/by-guid/draft-guid/entries", map[string]any{"draftGuid": "draft-guid"}},
		{"entry-update", "PUT", "/api/scheduledraft/by-guid/draft-guid/entries/entry-guid", map[string]any{"draftGuid": "draft-guid", "entryGuid": "entry-guid"}},
		{"entry-delete", "DELETE", "/api/scheduledraft/by-guid/draft-guid/entries/entry-guid", map[string]any{"draftGuid": "draft-guid", "entryGuid": "entry-guid"}},
		{"validate", "POST", "/api/scheduledraft/by-guid/draft-guid/validate", map[string]any{"draftGuid": "draft-guid"}},
		{"publish", "POST", "/api/scheduledraft/by-guid/draft-guid/publish", map[string]any{"draftGuid": "draft-guid"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			action := findAction(domain, testCase.name)
			if action == nil {
				t.Fatalf("action %q is not registered", testCase.name)
			}
			if action.HTTPMethod != testCase.method {
				t.Fatalf("method = %q, want %q", action.HTTPMethod, testCase.method)
			}
			path, _ := buildRESTPath(domain, *action, testCase.args)
			if path != testCase.path {
				t.Fatalf("path = %q, want %q", path, testCase.path)
			}
		})
	}
}

func TestScheduleDraftMutationsExposeConcurrencyAndAllocation(t *testing.T) {
	domain := findDomain("schedule-draft")
	update := findAction(domain, "entry-update")
	publish := findAction(domain, "publish")
	if update == nil || publish == nil {
		t.Fatal("expected entry-update and publish actions")
	}

	updateFlags := flagsByName(update.Flags)
	if updateFlags["allocation-percent"] == nil ||
		updateFlags["allocation-percent"].BodyName != "allocationPercent" {
		t.Fatal("entry-update must expose AllocationPercent")
	}
	if updateFlags["expected-row-version"] == nil ||
		!updateFlags["expected-row-version"].Required {
		t.Fatal("entry-update must require the latest row version")
	}
	if updateFlags["workers-json"] == nil ||
		!updateFlags["workers-json"].JSONFile ||
		updateFlags["workers-json"].BodyName != "workers" {
		t.Fatal("entry-update must accept typed worker allocation JSON")
	}

	publishFlags := flagsByName(publish.Flags)
	if publishFlags["expected-row-version"] == nil ||
		!publishFlags["expected-row-version"].Required {
		t.Fatal("publish must require the latest row version")
	}
	if publishFlags["acknowledged-warning-code"] == nil ||
		publishFlags["acknowledged-warning-code"].Type != "stringSlice" {
		t.Fatal("publish must expose explicit warning acknowledgements")
	}
}

func TestSchedulePublicationAndCapacityDomainsUseGUIDBoundaries(t *testing.T) {
	publication := findDomain("schedule-publication")
	capacity := findDomain("workforce-capacity")
	if publication == nil || capacity == nil {
		t.Fatal("schedule publication and workforce capacity domains must be registered")
	}

	diff := findAction(publication, "diff")
	availability := findAction(capacity, "availability")
	reserve := findAction(capacity, "reservation-create")
	if diff == nil || availability == nil || reserve == nil {
		t.Fatal("expected publication diff, availability, and reservation actions")
	}
	if diff.ToolName != "UteamupSchedulePublicationDiff" {
		t.Fatalf("unexpected publication diff tool %q", diff.ToolName)
	}
	for _, argument := range append(diff.Args, availability.Args...) {
		if argument.Type != "uuid" {
			t.Fatalf("public identifier %q must be a UUID, got %q", argument.Name, argument.Type)
		}
	}

	reservationFlags := flagsByName(reserve.Flags)
	if reservationFlags["allocation-percent"] == nil ||
		!reservationFlags["allocation-percent"].Required {
		t.Fatal("capacity reservation must require AllocationPercent")
	}
	if reserve.RESTPath != "reservations" ||
		reserve.ToolName != "UteamupWorkforceReservationCreate" {
		t.Fatalf("unexpected reservation contract: %+v", reserve)
	}
}

func flagsByName(flags []FlagDef) map[string]*FlagDef {
	result := make(map[string]*FlagDef, len(flags))
	for index := range flags {
		result[flags[index].Name] = &flags[index]
	}
	return result
}
