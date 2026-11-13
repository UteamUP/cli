package registry

import (
	"net/http"
	"strings"
	"testing"
)

// recordArchiveDomains are the domains whose records are archived, never deleted, with the REST base path each
// archive and restore route hangs off.
var recordArchiveDomains = map[string]string{
	"contact":         "/api/contact",
	"checklist":       "/api/checklist",
	"tasklist":        "/api/tasklist",
	"tag":             "/api/tag",
	"floorplan":       "/api/floorplan",
	"failure-code":    "/api/failurecode",
	"warranty":        "/api/warranties",
	"upmate-agent":    "/api/upmateagent",
	"automation":      "/api/automation",
	"report":          "/api/report",
	"report-template": "/api/reports/templates",
	"knowledgespace":  "/api/knowledgespace",
	"journal":         "/api/journal",
}

// recordListPaths is where each domain's GET list action is sent, which carries --archived.
var recordListPaths = map[string]string{
	"contact":         "/api/contact/paginated",
	"checklist":       "/api/checklist",
	"tasklist":        "/api/tasklist",
	"tag":             "/api/tag",
	"floorplan":       "/api/floorplan",
	"failure-code":    "/api/failurecode",
	"warranty":        "/api/warranties",
	"upmate-agent":    "/api/upmateagent",
	"report-template": "/api/reports/templates",
	"knowledgespace":  "/api/knowledgespace",
}

func TestRecordArchiveAndRestorePostToTheGuidRoutes(t *testing.T) {
	const guid = "11111111-1111-4111-8111-111111111111"
	for domainName, basePath := range recordArchiveDomains {
		for _, action := range []string{"archive", "restore"} {
			request, err := runReferenceCommand(t, domainName, action, guid)
			if err != nil {
				t.Fatalf("%s %s: %v", domainName, action, err)
			}
			want := basePath + "/by-guid/" + guid + "/" + action
			if request.Method != http.MethodPost || request.URL.Path != want {
				t.Fatalf("%s %s = %s %s, want POST %s", domainName, action, request.Method, request.URL.Path, want)
			}
			if request.URL.RawQuery != "" {
				t.Fatalf("%s %s leaked arguments into the query: %s", domainName, action, request.URL)
			}
		}
	}
}

func TestRecordArchiveDeclaresOneRequiredNonEmptyGuid(t *testing.T) {
	tools := map[string]string{"archive": "ArchiveRecord", "restore": "RestoreRecord"}
	for domainName := range recordArchiveDomains {
		for actionName, toolName := range tools {
			action := findDomainAction(t, domainName, actionName)
			if action.ToolName != toolName || action.HTTPMethod != "POST" || !action.RejectExtraArgs {
				t.Fatalf("%s %s = tool %q method %q rejectExtra %v", domainName, actionName,
					action.ToolName, action.HTTPMethod, action.RejectExtraArgs)
			}
			if len(action.Args) != 1 || !action.Args[0].Required || action.Args[0].Type != "non-empty-uuid" {
				t.Fatalf("%s %s args = %+v, want one required non-empty-uuid", domainName, actionName, action.Args)
			}
		}
	}
}

func TestRecordArchiveRefusesBadGuidsBeforeSending(t *testing.T) {
	cases := [][]string{
		{"archive", "00000000-0000-0000-0000-000000000000"},
		{"archive", "not-a-guid"},
		{"restore", "00000000-0000-0000-0000-000000000000"},
		{"archive", "11111111-1111-4111-8111-111111111111", "extra"},
	}
	for domainName := range recordArchiveDomains {
		for _, args := range cases {
			request, err := runReferenceCommand(t, domainName, args...)
			if err == nil || request != nil {
				t.Fatalf("%s %v must be refused before any request; err=%v request=%v", domainName, args, err, request)
			}
		}
	}
}

func TestRecordListsHideArchivedUnlessAsked(t *testing.T) {
	for domainName, listPath := range recordListPaths {
		list := findDomainAction(t, domainName, "list")
		flag := findFlag(list, "archived")
		if flag == nil || flag.QueryName != "archived" || flag.Type != "bool" || flag.Default != nil || flag.Required {
			t.Fatalf("%s list --archived = %+v, want optional bool sent as archived with no default", domainName, flag)
		}

		request, err := runReferenceCommand(t, domainName, "list", "--archived")
		if err != nil {
			t.Fatalf("%s list --archived: %v", domainName, err)
		}
		if request.Method != http.MethodGet || request.URL.Path != listPath || request.URL.Query().Get("archived") != "true" {
			t.Fatalf("%s list --archived = %s %s, want GET %s?archived=true", domainName, request.Method, request.URL, listPath)
		}

		request, err = runReferenceCommand(t, domainName, "list")
		if err != nil {
			t.Fatalf("%s list: %v", domainName, err)
		}
		if request.URL.Query().Has("archived") {
			t.Fatalf("%s list must leave archived out so the server hides archived records: %s", domainName, request.URL)
		}
	}
}

func TestPostListsSendArchivedInTheBody(t *testing.T) {
	bodyNames := map[string]string{"automation": "archived", "journal": "isArchived"}
	for domainName, bodyName := range bodyNames {
		flag := findFlag(findDomainAction(t, domainName, "list"), "archived")
		if flag == nil || flag.BodyName != bodyName || flag.QueryName != "" || flag.Type != "bool" || flag.Default != nil {
			t.Fatalf("%s list --archived = %+v, want an optional bool sent in the body as %s", domainName, flag, bodyName)
		}
	}
}

func TestDriverRestorePostsToTheGuidRoute(t *testing.T) {
	const guid = "11111111-1111-4111-8111-111111111111"
	request, err := runReferenceCommand(t, "driver", "restore", guid)
	if err != nil {
		t.Fatal(err)
	}
	if request.Method != http.MethodPost || request.URL.Path != "/api/driver/by-guid/"+guid+"/restore" {
		t.Fatalf("driver restore = %s %s", request.Method, request.URL.Path)
	}
	if request, err := runReferenceCommand(t, "driver", "restore", "00000000-0000-0000-0000-000000000000"); err == nil || request != nil {
		t.Fatalf("driver restore of the empty GUID must be refused before sending; err=%v", err)
	}
}

func TestRecordDeleteIsDescribedAsArchive(t *testing.T) {
	for _, domainName := range []string{"contact", "checklist", "tasklist", "tag", "failure-code", "warranty",
		"report", "report-template", "knowledgespace", "journal", "driver"} {
		action := findDomainAction(t, domainName, "delete")
		if !strings.Contains(action.Description, "never deleted") || !strings.Contains(action.Description, "archive") {
			t.Fatalf("%s delete must say it archives: %q", domainName, action.Description)
		}
	}
}

func TestWorkforceGroupAndHolidayArchiveThroughTheirActiveSwitch(t *testing.T) {
	const guid = "11111111-1111-4111-8111-111111111111"
	routes := map[string]string{"workforce-group": "/api/workforcegroups", "tenant-holiday": "/api/tenantholiday"}
	for domainName, basePath := range routes {
		for _, action := range []string{"archive", "restore"} {
			definition := findDomainAction(t, domainName, action)
			if definition.ToolName != map[string]string{"archive": "ArchiveReferenceRecord", "restore": "RestoreReferenceRecord"}[action] {
				t.Fatalf("%s %s tool = %q", domainName, action, definition.ToolName)
			}
			request, err := runReferenceCommand(t, domainName, action, guid)
			if err != nil {
				t.Fatalf("%s %s: %v", domainName, action, err)
			}
			want := basePath + "/by-guid/" + guid + "/" + action
			if request.Method != http.MethodPost || request.URL.Path != want {
				t.Fatalf("%s %s = %s %s, want POST %s", domainName, action, request.Method, request.URL.Path, want)
			}
		}
	}

	request, err := runReferenceCommand(t, "workforce-group", "list", "--is-active=false")
	if err != nil {
		t.Fatal(err)
	}
	if request.URL.Path != "/api/workforcegroups" || request.URL.Query().Get("isActive") != "false" {
		t.Fatalf("workforce-group list --is-active=false = %s", request.URL)
	}

	request, err = runReferenceCommand(t, "tenant-holiday", "year", "2026", "--is-active=false")
	if err != nil {
		t.Fatal(err)
	}
	if request.URL.Path != "/api/tenantholiday/year/2026" || request.URL.Query().Get("isActive") != "false" {
		t.Fatalf("tenant-holiday year --is-active=false = %s", request.URL)
	}
}
