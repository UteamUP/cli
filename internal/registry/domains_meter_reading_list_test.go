package registry

import (
	"strings"
	"testing"
)

func TestMeterReadingListTargetsTenantWideRoute(t *testing.T) {
	action := findDomainAction(t, "meter-reading", "list")
	if action.ToolName != "UteamupMeterreadingList" || action.HTTPMethod != "GET" {
		t.Errorf("list is miswired: tool=%s method=%s", action.ToolName, action.HTTPMethod)
	}
	if len(action.Args) != 0 {
		t.Errorf("list must take no positional args, got %d", len(action.Args))
	}
	if err := validateActionDefinition(*action); err != nil {
		t.Fatalf("list definition is invalid: %v", err)
	}

	domain := findDomain("meter-reading")
	path, consumed := buildRESTPath(domain, *action, map[string]any{"page": 2, "pageSize": 25, "source": "IoT"})
	if path != "/api/meter-readings" {
		t.Errorf("list path = %q, want /api/meter-readings", path)
	}
	if len(consumed) != 0 {
		t.Errorf("list must keep every filter on the query string, consumed %v", consumed)
	}
}

func TestMeterReadingListFlagsMirrorTheBackendQuery(t *testing.T) {
	action := findDomainAction(t, "meter-reading", "list")

	page := actionFlagByName(t, action, "page")
	if page.Type != "int" || page.Default != 1 || page.Required {
		t.Errorf("page must be an optional int defaulting to 1: %+v", page)
	}
	pageSize := actionFlagByName(t, action, "page-size")
	if pageSize.Type != "int" || pageSize.Default != 25 || pageSize.Required {
		t.Errorf("page-size must be an optional int defaulting to 25: %+v", pageSize)
	}
	for _, name := range []string{"asset-guid", "attribute-definition-guid"} {
		flag := actionFlagByName(t, action, name)
		if flag.Type != "uuid" || flag.Required {
			t.Errorf("%s must be an optional GUID flag: %+v", name, flag)
		}
	}
	for _, name := range []string{"from", "to", "search"} {
		flag := actionFlagByName(t, action, name)
		if flag.Type != "string" || flag.Required {
			t.Errorf("%s must be an optional string flag: %+v", name, flag)
		}
	}

	source := actionFlagByName(t, action, "source")
	if strings.Join(source.AllowedValues, ",") != "Manual,IoT,Import,Inspection,Telematics,RentalHandover" {
		t.Errorf("source must allow exactly the MeterReadingSource names, got %v", source.AllowedValues)
	}
	quality := actionFlagByName(t, action, "quality")
	if strings.Join(quality.AllowedValues, ",") != "Normal,Estimated,Quarantined,Corrected" {
		t.Errorf("quality must allow exactly the MeterReadingQuality names, got %v", quality.AllowedValues)
	}
}

func TestMeterReadingListHasNoPersonFilter(t *testing.T) {
	action := findDomainAction(t, "meter-reading", "list")
	for _, name := range []string{"user-guid", "recorded-by", "recorded-by-guid", "person-guid", "user-id", "tenant-guid"} {
		if actionFlagExists(action, name) {
			t.Errorf("list must not filter by person (privacy): found --%s", name)
		}
	}
}

func TestMeterReadingGetTargetsByGuidRoute(t *testing.T) {
	action := findDomainAction(t, "meter-reading", "get")
	if action.ToolName != "UteamupMeterreadingGet" || action.HTTPMethod != "GET" {
		t.Errorf("get is miswired: tool=%s method=%s", action.ToolName, action.HTTPMethod)
	}
	if len(action.Args) != 1 {
		t.Fatalf("get args = %d, want 1", len(action.Args))
	}
	reading := action.Args[0]
	if reading.Name != "readingGuid" || !reading.Required || reading.Type != "uuid" {
		t.Errorf("readingGuid is miswired: %+v", reading)
	}

	readingGuid := "b2c3d4e5-f6a7-4b5c-8d7e-9f0a1b2c3d4e"
	domain := findDomain("meter-reading")
	path, consumed := buildRESTPath(domain, *action, map[string]any{"readingGuid": readingGuid})
	if path != "/api/meter-readings/by-guid/"+readingGuid {
		t.Errorf("get path = %q, want /api/meter-readings/by-guid/%s", path, readingGuid)
	}
	if len(consumed) != 1 || consumed[0] != "readingGuid" {
		t.Errorf("get must consume readingGuid into the route, consumed %v", consumed)
	}
}
