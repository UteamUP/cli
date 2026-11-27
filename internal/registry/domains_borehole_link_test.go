package registry

import (
	"net/http"
	"testing"
)

const (
	boreholeLinkTestTargetGUID = "3b0e6f7a-1c2d-4e5f-8a9b-0c1d2e3f4a5b"
	boreholeLinkTestLinkGUID   = "9c8b7a6d-5e4f-4a3b-8c2d-1e0f9a8b7c6d"
)

// Every action maps to the exact MCP tool method in MCP/Tools/BoreholeLinkTools.cs (UTE-2116).
func TestBoreholeLinkActionsMapToMCPTools(t *testing.T) {
	expected := map[string]string{
		"links":            "UteamupBoreholeLinkList",
		"links-for-target": "UteamupBoreholeLinkListForTarget",
		"links-add":        "UteamupBoreholeLinkCreate",
		"links-delete":     "UteamupBoreholeLinkDelete",
	}
	for name, tool := range expected {
		if got := findDomainAction(t, "borehole", name).ToolName; got != tool {
			t.Errorf("borehole %s ToolName = %q, want %q", name, got, tool)
		}
	}
	if got := len(findDomainByName(t, "borehole").Actions); got != len(expected) {
		t.Errorf("borehole has %d actions, want %d", got, len(expected))
	}
}

func singleBoreholeRequest(t *testing.T, args ...string) recordedDomainRequest {
	t.Helper()
	requests, err := runRegisteredDomainCommand(t, "borehole", args...)
	if err != nil {
		t.Fatalf("borehole %v error = %v", args, err)
	}
	if len(requests) != 1 {
		t.Fatalf("borehole %v sent %d requests, want 1", args, len(requests))
	}
	return requests[0]
}

func TestBoreholeLinksListsLinksOfOneBorehole(t *testing.T) {
	request := singleBoreholeRequest(t, "links", "12345")
	if request.method != http.MethodGet || request.path != "/api/boreholes/12345/links" {
		t.Errorf("request = %s %s, want GET /api/boreholes/12345/links", request.method, request.path)
	}
}

func TestBoreholeLinksForTargetSendsTargetAsQuery(t *testing.T) {
	request := singleBoreholeRequest(t, "links-for-target", "--target-type", "asset", "--target-guid", boreholeLinkTestTargetGUID)
	if request.method != http.MethodGet || request.path != "/api/boreholes/links" {
		t.Errorf("request = %s %s, want GET /api/boreholes/links", request.method, request.path)
	}
	if got := request.query["targetType"]; len(got) != 1 || got[0] != "asset" {
		t.Errorf("targetType query = %v, want [asset]", got)
	}
	if got := request.query["targetGuid"]; len(got) != 1 || got[0] != boreholeLinkTestTargetGUID {
		t.Errorf("targetGuid query = %v, want [%s]", got, boreholeLinkTestTargetGUID)
	}
}

func TestBoreholeLinksAddPostsTargetInBody(t *testing.T) {
	request := singleBoreholeRequest(t, "links-add", "12345", "--target-type", "code", "--target-guid", boreholeLinkTestTargetGUID)
	if request.method != http.MethodPost || request.path != "/api/boreholes/12345/links" {
		t.Errorf("request = %s %s, want POST /api/boreholes/12345/links", request.method, request.path)
	}
	if request.body["targetType"] != "code" || request.body["targetGuid"] != boreholeLinkTestTargetGUID {
		t.Errorf("body = %v, want targetType=code and targetGuid=%s", request.body, boreholeLinkTestTargetGUID)
	}
	if _, leaked := request.body["boreholeUid"]; leaked {
		t.Errorf("route argument boreholeUid leaked into body: %v", request.body)
	}
}

func TestBoreholeLinksDeleteTargetsOneLink(t *testing.T) {
	request := singleBoreholeRequest(t, "links-delete", "12345", boreholeLinkTestLinkGUID)
	want := "/api/boreholes/12345/links/" + boreholeLinkTestLinkGUID
	if request.method != http.MethodDelete || request.path != want {
		t.Errorf("request = %s %s, want DELETE %s", request.method, request.path, want)
	}
}

func TestBoreholeLinksRejectBadInputBeforeAnyRequest(t *testing.T) {
	cases := map[string][]string{
		"unknown target type":    {"links-for-target", "--target-type", "vendor", "--target-guid", boreholeLinkTestTargetGUID},
		"empty target GUID":      {"links-add", "12345", "--target-type", "asset", "--target-guid", "00000000-0000-0000-0000-000000000000"},
		"malformed target GUID":  {"links-add", "12345", "--target-type", "asset", "--target-guid", "not-a-guid"},
		"missing target type":    {"links-add", "12345", "--target-guid", boreholeLinkTestTargetGUID},
		"malformed link GUID":    {"links-delete", "12345", "../../asset"},
		"route delimiter in uid": {"links", "12345/../../asset"},
		"extra positional":       {"links", "12345", "extra"},
	}
	for name, args := range cases {
		t.Run(name, func(t *testing.T) {
			requests, err := runRegisteredDomainCommand(t, "borehole", args...)
			if err == nil {
				t.Errorf("borehole %v succeeded, want a validation error", args)
			}
			if len(requests) != 0 {
				t.Errorf("borehole %v sent %d requests, want none", args, len(requests))
			}
		})
	}
}
