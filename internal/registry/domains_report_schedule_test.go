package registry

import (
	"reflect"
	"strings"
	"testing"
)

func TestReportScheduleRegistryIsGuidOnlyAndWired(t *testing.T) {
	domain := findDomain("report-schedule")
	if domain == nil {
		t.Fatal("expected report-schedule domain to be registered")
	}
	if domain.APIPath != "/api/reports/schedules" {
		t.Fatalf("APIPath = %q, want /api/reports/schedules", domain.APIPath)
	}
	if !reflect.DeepEqual(domain.Aliases, []string{"rsch", "report-schedules"}) {
		t.Fatalf("Aliases = %v, want [rsch report-schedules] (rs belongs to the partner domain)", domain.Aliases)
	}

	want := []struct {
		name, method, path, tool string
		domainBase               bool
	}{
		{"list", "GET", "", "UteamupReportScheduleList", true},
		{"get", "GET", "by-guid/{scheduleGuid}", "UteamupReportScheduleGet", false},
		{"pause", "POST", "by-guid/{scheduleGuid}/pause", "UteamupReportScheduleSetActive", false},
		{"resume", "POST", "by-guid/{scheduleGuid}/resume", "UteamupReportScheduleSetActive", false},
		{"send-test", "POST", "by-guid/{scheduleGuid}/send-test", "UteamupReportScheduleSendTest", false},
		{"delete", "DELETE", "by-guid/{scheduleGuid}", "UteamupReportScheduleDelete", false},
	}
	if len(domain.Actions) != len(want) {
		t.Fatalf("actions = %d, want %d", len(domain.Actions), len(want))
	}

	for i, expected := range want {
		action := domain.Actions[i]
		if action.Name != expected.name || action.HTTPMethod != expected.method ||
			action.RESTPath != expected.path || action.ToolName != expected.tool ||
			action.UseDomainBasePath != expected.domainBase {
			t.Errorf("action %d = %s %s %q tool %s domainBase %v, want %s %s %q tool %s domainBase %v",
				i, action.Name, action.HTTPMethod, action.RESTPath, action.ToolName, action.UseDomainBasePath,
				expected.name, expected.method, expected.path, expected.tool, expected.domainBase)
		}
		if err := validateActionDefinition(action); err != nil {
			t.Errorf("action %s definition: %v", action.Name, err)
		}

		if expected.name == "list" {
			if len(action.Args) != 0 {
				t.Errorf("list must not take positional identifiers, got %+v", action.Args)
			}
		} else {
			if len(action.Args) != 1 {
				t.Fatalf("%s args = %+v, want exactly scheduleGuid", action.Name, action.Args)
			}
			arg := action.Args[0]
			if arg.Name != "scheduleGuid" || arg.Type != "non-empty-uuid" || !arg.Required {
				t.Errorf("%s arg = %+v, want required non-empty-uuid scheduleGuid", action.Name, arg)
			}
			if len(action.Flags) != 0 {
				t.Errorf("%s must send no body or query flags, got %+v", action.Name, action.Flags)
			}
		}

		if action.DownloadResponseBody || action.DownloadURLField != "" {
			t.Errorf("%s must not download a file", action.Name)
		}

		for _, flag := range action.Flags {
			lowered := strings.ToLower(flag.Name + " " + flag.BodyName + " " + flag.QueryName)
			for _, forbidden := range []string{"tenant", "user", "owner", "sql"} {
				if strings.Contains(lowered, forbidden) {
					t.Errorf("%s flag --%s must not carry %q", action.Name, flag.Name, forbidden)
				}
			}
		}
	}

	// Decision 11: schedules are authored in the web builder, so no create or update verb.
	for _, action := range domain.Actions {
		if action.Name == "create" || action.Name == "update" || action.HTTPMethod == "PUT" || action.HTTPMethod == "PATCH" {
			t.Errorf("unexpected authoring action %s %s", action.HTTPMethod, action.Name)
		}
	}

	list := findDomainAction(t, "report-schedule", "list")
	listQueries := map[string]string{}
	for _, flag := range list.Flags {
		listQueries[flag.Name] = flag.QueryName
	}
	wantQueries := map[string]string{
		"search": "search", "active": "isActive", "mine": "mineOnly", "page": "page", "page-size": "pageSize",
	}
	if !reflect.DeepEqual(listQueries, wantQueries) {
		t.Errorf("list flags = %v, want %v", listQueries, wantQueries)
	}
	for _, name := range []string{"active", "mine"} {
		flag := findFlag(list, name)
		if flag == nil || flag.Type != "bool" || flag.Default != nil {
			t.Errorf("--%s = %+v, want a bool with no default so it is sent only when given", name, flag)
		}
	}
	if pageSize := findFlag(list, "page-size"); pageSize == nil || pageSize.Default != 50 {
		t.Errorf("--page-size = %+v, want default 50 like the REST query model", pageSize)
	}
}
