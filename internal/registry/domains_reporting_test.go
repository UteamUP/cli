package registry

import (
	"strings"
	"testing"
)

func TestApprovedReportAnalyticsReadWired(t *testing.T) {
	var domain *Domain
	for _, candidate := range DefaultRegistry.Domains() {
		if candidate.Name == "report-analytics" {
			domain = candidate
			break
		}
	}
	if domain == nil {
		t.Fatal("expected report-analytics domain")
	}
	if domain.APIPath != "/api/report" {
		t.Fatalf("APIPath = %q, want /api/report", domain.APIPath)
	}
	if len(domain.Actions) != 1 {
		t.Fatalf("actions = %d, want one bounded read", len(domain.Actions))
	}

	action := domain.Actions[0]
	if action.Name != "read" || action.ToolName != "UteamupReportAnalytics" {
		t.Errorf("action = %q/%q, want read/UteamupReportAnalytics", action.Name, action.ToolName)
	}
	if action.HTTPMethod != "GET" || action.RESTPath != "analytics" {
		t.Errorf("route = %s %s, want GET analytics", action.HTTPMethod, action.RESTPath)
	}
	if len(action.Args) != 0 {
		t.Errorf("read must not expose positional identifiers, got %+v", action.Args)
	}

	flags := make(map[string]FlagDef, len(action.Flags))
	for _, flag := range action.Flags {
		flags[flag.Name] = flag
	}
	for _, name := range []string{"start-date", "end-date"} {
		if flag, ok := flags[name]; !ok || !flag.Required || flag.Type != "string" {
			t.Errorf("%s = %+v, want required string", name, flag)
		}
	}
	if flag, ok := flags["group-by"]; !ok || flag.Default != "month" || flag.Type != "string" {
		t.Errorf("group-by = %+v, want optional string default month", flag)
	}
}

func TestCostOverviewWorkorderGetUsesGuidContract(t *testing.T) {
	action := findDomainAction(t, "cost-overview", "get")

	if action.ToolName != "UteamupCostByWorkorder" {
		t.Fatalf("ToolName = %q, want UteamupCostByWorkorder", action.ToolName)
	}
	if action.HTTPMethod != "GET" || action.RESTPath != "workorders/by-guid/{workorderGuid}" {
		t.Fatalf("route = %s %s, want GET workorders/by-guid/{workorderGuid}", action.HTTPMethod, action.RESTPath)
	}
	if len(action.Args) != 1 {
		t.Fatalf("args = %+v, want one GUID arg", action.Args)
	}
	arg := action.Args[0]
	if arg.Name != "workorderGuid" || arg.Type != "uuid" || !arg.Required {
		t.Fatalf("arg = %+v, want required workorderGuid uuid", arg)
	}
}

func TestAssetReportsGetUsesGuidContract(t *testing.T) {
	action := findDomainAction(t, "asset-reports", "get")

	if action.ToolName != "UteamupAssetReports" {
		t.Fatalf("ToolName = %q, want UteamupAssetReports", action.ToolName)
	}
	if action.HTTPMethod != "GET" || action.RESTPath != "asset/by-guid/{assetGuid}" {
		t.Fatalf("route = %s %s, want GET asset/by-guid/{assetGuid}", action.HTTPMethod, action.RESTPath)
	}
	if len(action.Args) != 1 {
		t.Fatalf("args = %+v, want one GUID arg", action.Args)
	}
	arg := action.Args[0]
	if arg.Name != "assetGuid" || arg.Type != "uuid" || !arg.Required {
		t.Fatalf("arg = %+v, want required assetGuid uuid", arg)
	}
}

func TestCompletionReportActionsUseGuidContracts(t *testing.T) {
	tests := []struct {
		action string
		method string
		path   string
		arg    string
	}{
		{action: "get", method: "GET", path: "worker/by-guid/{reportGuid}", arg: "reportGuid"},
		{action: "detail", method: "GET", path: "detail/by-guid/{reportGuid}", arg: "reportGuid"},
		{action: "create", method: "POST", path: "workorder/by-guid/{workorderGuid}", arg: "workorderGuid"},
		{action: "update", method: "PUT", path: "by-guid/{reportGuid}", arg: "reportGuid"},
		{action: "delete", method: "DELETE", path: "by-guid/{reportGuid}", arg: "reportGuid"},
	}

	for _, test := range tests {
		action := findDomainAction(t, "report", test.action)
		if action.HTTPMethod != test.method || action.RESTPath != test.path {
			t.Errorf("%s route = %s %s, want %s %s", test.action, action.HTTPMethod, action.RESTPath, test.method, test.path)
		}
		if len(action.Args) != 1 {
			t.Errorf("%s args = %+v, want one GUID arg", test.action, action.Args)
			continue
		}
		arg := action.Args[0]
		if arg.Name != test.arg || arg.Type != "uuid" || !arg.Required {
			t.Errorf("%s arg = %+v, want required %s uuid", test.action, arg, test.arg)
		}
	}

	list := findDomainAction(t, "report", "list")
	for _, arg := range list.Args {
		if arg.Type == "int" || arg.Name == "id" {
			t.Fatalf("report list exposes sequential identity: %+v", arg)
		}
	}
}

func TestReportCreateUsesGuidPeopleFlags(t *testing.T) {
	action := findDomainAction(t, "report", "create")

	for _, removed := range []string{"primary-reporter-id", "additional-worker-ids"} {
		for _, flag := range action.Flags {
			if flag.Name == removed {
				t.Fatalf("report create still exposes identity-key flag --%s", removed)
			}
		}
	}

	reporter := actionFlagByName(t, action, "primary-reporter-guid")
	if reporter.Type != "uuid" || reporter.Required {
		t.Errorf("primary-reporter-guid = %+v, want optional uuid", *reporter)
	}
	if reporter.BodyName != "" && reporter.BodyName != "primaryReporterGuid" {
		t.Errorf("primary-reporter-guid BodyName = %q, want primaryReporterGuid", reporter.BodyName)
	}

	workers := actionFlagByName(t, action, "additional-worker-guids")
	if workers.Type != "stringSlice" || workers.BodyName != "additionalWorkerGuids" || workers.Required {
		t.Errorf("additional-worker-guids = %+v, want optional stringSlice additionalWorkerGuids", *workers)
	}

	notify := actionFlagByName(t, action, "notify-external-workers")
	if notify.Type != "bool" || notify.Required || notify.Default != nil {
		t.Errorf("notify-external-workers = %+v, want optional bool without default", *notify)
	}

	emails := actionFlagByName(t, action, "external-worker-emails")
	if emails.Type != "stringSlice" {
		t.Errorf("external-worker-emails = %+v, want stringSlice", *emails)
	}
}

func TestReportCreateHasIdempotencyHeader(t *testing.T) {
	action := findDomainAction(t, "report", "create")

	key := actionFlagByName(t, action, "idempotency-key")
	if key.HeaderName != "Idempotency-Key" {
		t.Fatalf("idempotency-key HeaderName = %q, want Idempotency-Key", key.HeaderName)
	}
	if key.Type != "non-empty-uuid" || key.Required || key.BodyName != "" || key.MirrorHeaderInBody {
		t.Fatalf("idempotency-key = %+v, want optional header-only non-empty GUID", *key)
	}

	domain := findDomain("report")
	command := buildActionCommand(domain, *action, nil, nil, nil, nil)
	args := []string{"3f2504e0-4f89-11d3-9a0c-0305e82c3301"}
	for name, value := range map[string]string{"description": "Replaced the seal", "report-date": "2026-09-24"} {
		if err := command.Flags().Set(name, value); err != nil {
			t.Fatalf("set %s: %v", name, err)
		}
	}
	if err := command.Flags().Set("idempotency-key", "00000000-0000-0000-0000-000000000000"); err != nil {
		t.Fatalf("set idempotency-key: %v", err)
	}
	if err := validateActionInput(command, args, *action); err == nil || !strings.Contains(err.Error(), "--idempotency-key") {
		t.Fatalf("empty idempotency GUID accepted: %v", err)
	}
	if err := command.Flags().Set("idempotency-key", "8d6f1c2a-4b3e-4f5a-9c7d-1e2f3a4b5c6d"); err != nil {
		t.Fatalf("set idempotency-key: %v", err)
	}
	if err := validateActionInput(command, args, *action); err != nil {
		t.Fatalf("valid idempotency GUID rejected: %v", err)
	}
}

func TestReportAnalyticsGroupByAllowedValues(t *testing.T) {
	domain := findDomain("report-analytics")
	action := findDomainAction(t, "report-analytics", "read")

	groupBy := actionFlagByName(t, action, "group-by")
	if got := strings.Join(groupBy.AllowedValues, ","); got != "day,week,month,quarter,year" {
		t.Fatalf("group-by AllowedValues = %q, want day,week,month,quarter,year", got)
	}
	if !strings.Contains(actionFlagByName(t, action, "end-date").Description, "whole day") {
		t.Errorf("end-date help must state the inclusive whole end day")
	}

	command := buildActionCommand(domain, *action, nil, nil, nil, nil)
	for name, value := range map[string]string{"start-date": "2026-07-01", "end-date": "2026-09-24"} {
		if err := command.Flags().Set(name, value); err != nil {
			t.Fatalf("set %s: %v", name, err)
		}
	}
	if err := validateActionInput(command, nil, *action); err != nil {
		t.Fatalf("default group-by rejected: %v", err)
	}
	if err := command.Flags().Set("group-by", "fortnight"); err != nil {
		t.Fatalf("set group-by: %v", err)
	}
	if err := validateActionInput(command, nil, *action); err == nil || !strings.Contains(err.Error(), "--group-by") {
		t.Fatalf("group-by fortnight accepted locally: %v", err)
	}
	for _, value := range []string{"day", "week", "month", "quarter", "year"} {
		if err := command.Flags().Set("group-by", value); err != nil {
			t.Fatalf("set group-by: %v", err)
		}
		if err := validateActionInput(command, nil, *action); err != nil {
			t.Fatalf("group-by %s rejected: %v", value, err)
		}
	}

	pageSize := actionFlagByName(t, findDomainAction(t, "asset-reports", "get"), "page-size")
	if !strings.Contains(pageSize.Description, "1-100") {
		t.Errorf("asset-reports page-size help = %q, want the 1-100 bound", pageSize.Description)
	}
}

func TestWorkReportToolNamesMatchMcp(t *testing.T) {
	expected := map[string]string{
		"list":   "UteamupWorkReportList",
		"get":    "UteamupWorkReportGet",
		"detail": "UteamupWorkReportDetail",
		"create": "UteamupWorkReportCreate",
		"update": "UteamupWorkReportUpdate",
		"delete": "UteamupWorkReportDelete",
	}
	for actionName, toolName := range expected {
		if got := findDomainAction(t, "report", actionName).ToolName; got != toolName {
			t.Errorf("report %s ToolName = %q, want %q", actionName, got, toolName)
		}
	}

	for _, action := range findDomain("report").Actions {
		if strings.HasPrefix(action.ToolName, "UteamupReport") {
			t.Errorf("report %s uses the registered-catalog prefix: %q", action.Name, action.ToolName)
		}
	}
	if got := findDomainAction(t, "registered-report", "list").ToolName; got != "UteamupReportList" {
		t.Errorf("registered-report list ToolName = %q, want UteamupReportList", got)
	}
}

func TestReportListUsesWorkerRoute(t *testing.T) {
	action := findDomainAction(t, "report", "list")

	if action.HTTPMethod != "GET" || action.RESTPath != "worker" {
		t.Fatalf("list route = %s %s, want GET worker", action.HTTPMethod, action.RESTPath)
	}
	if len(action.Args) != 0 {
		t.Fatalf("list args = %+v, want none", action.Args)
	}

	workorder := actionFlagByName(t, action, "workorder-guid")
	if workorder.Type != "uuid" || workorder.Required {
		t.Errorf("workorder-guid = %+v, want optional uuid", *workorder)
	}
	if name := actionFlagByName(t, action, "name-filter"); name.Type != "string" || name.Required {
		t.Errorf("name-filter = %+v, want optional string", *name)
	}
	for _, name := range []string{"page", "page-size"} {
		if flag := actionFlagByName(t, action, name); flag.Type != "int" {
			t.Errorf("%s = %+v, want int pagination flag", name, *flag)
		}
	}
}

func TestReportGetUsesWorkerDetailRoute(t *testing.T) {
	action := findDomainAction(t, "report", "get")

	if action.HTTPMethod != "GET" || action.RESTPath != "worker/by-guid/{reportGuid}" {
		t.Fatalf("get route = %s %s, want GET worker/by-guid/{reportGuid}", action.HTTPMethod, action.RESTPath)
	}
	if len(action.Args) != 1 || action.Args[0].Name != "reportGuid" || action.Args[0].Type != "uuid" {
		t.Fatalf("get args = %+v, want one reportGuid uuid", action.Args)
	}
	if len(action.Flags) != 0 {
		t.Errorf("get flags = %+v, want none", action.Flags)
	}
}

func TestReportUpdateRequiresDescriptionAndReportDate(t *testing.T) {
	domain := findDomain("report")
	action := findDomainAction(t, "report", "update")

	if action.HTTPMethod != "PUT" || action.RESTPath != "by-guid/{reportGuid}" {
		t.Fatalf("update route = %s %s, want PUT by-guid/{reportGuid}", action.HTTPMethod, action.RESTPath)
	}
	for _, name := range []string{"description", "report-date"} {
		if flag := actionFlagByName(t, action, name); !flag.Required || flag.Type != "string" {
			t.Errorf("%s = %+v, want required string", name, *flag)
		}
	}
	optional := map[string]string{
		"close-out-notes":         "string",
		"time-spent":              "float",
		"cost-incurred":           "float",
		"primary-reporter-guid":   "uuid",
		"additional-worker-guids": "stringSlice",
		"external-worker-emails":  "stringSlice",
	}
	for name, flagType := range optional {
		if flag := actionFlagByName(t, action, name); flag.Required || flag.Type != flagType {
			t.Errorf("%s = %+v, want optional %s", name, *flag, flagType)
		}
	}
	if workers := actionFlagByName(t, action, "additional-worker-guids"); workers.BodyName != "additionalWorkerGuids" {
		t.Errorf("additional-worker-guids BodyName = %q, want additionalWorkerGuids", workers.BodyName)
	}
	for _, removed := range []string{"primary-reporter-id", "additional-worker-ids"} {
		for _, flag := range action.Flags {
			if flag.Name == removed {
				t.Fatalf("report update exposes identity-key flag --%s", removed)
			}
		}
	}
	if !strings.Contains(action.Description, "cleared") || !strings.Contains(action.Description, "kept") {
		t.Errorf("update help = %q, want it to state what is cleared and what is kept", action.Description)
	}

	command := buildActionCommand(domain, *action, nil, nil, nil, nil)
	if err := command.Flags().Set("time-spent", "3.5"); err != nil {
		t.Fatalf("set time-spent: %v", err)
	}
	err := command.ValidateRequiredFlags()
	if err == nil {
		t.Fatal("update without --description and --report-date passed required-flag validation")
	}
	for _, name := range []string{"description", "report-date"} {
		if !strings.Contains(err.Error(), name) {
			t.Errorf("required-flag error %q does not name --%s", err, name)
		}
	}

	for name, value := range map[string]string{"description": "Replaced seal", "report-date": "2026-09-20"} {
		if err := command.Flags().Set(name, value); err != nil {
			t.Fatalf("set %s: %v", name, err)
		}
	}
	if err := command.ValidateRequiredFlags(); err != nil {
		t.Fatalf("update with the required flags failed validation: %v", err)
	}
	if err := validateActionInput(command, []string{"7b7f0c2a-4b3e-4f5a-9c7d-1e2f3a4b5c6d"}, *action); err != nil {
		t.Fatalf("valid update input rejected: %v", err)
	}
}

func TestReportUpdateFloatFlagsHaveNoIntDefaults(t *testing.T) {
	action := findDomainAction(t, "report", "update")

	for _, name := range []string{"time-spent", "cost-incurred"} {
		flag := actionFlagByName(t, action, name)
		if flag.Type != "float" {
			t.Errorf("%s type = %q, want float", name, flag.Type)
		}
		if flag.Default == nil {
			continue
		}
		if _, ok := flag.Default.(float64); !ok {
			t.Errorf("%s default = %#v (%T), want a float literal", name, flag.Default, flag.Default)
		}
	}
}
