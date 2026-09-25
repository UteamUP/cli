package registry

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/uteamup/cli/internal/auth"
	"github.com/uteamup/cli/internal/client"
	"github.com/uteamup/cli/internal/logging"
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
		"list":          "UteamupWorkReportList",
		"get":           "UteamupWorkReportGet",
		"detail":        "UteamupWorkReportDetail",
		"create":        "UteamupWorkReportCreate",
		"update":        "UteamupWorkReportUpdate",
		"delete":        "UteamupWorkReportDelete",
		"pdf":           "UteamupWorkReportPdf",
		"workorder-pdf": "UteamupWorkReportWorkorderPdf",
		"export":        "UteamupWorkReportExport",
		"send":          "UteamupWorkReportSend",
		"finalize":      "UteamupWorkReportFinalize",
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

const (
	reportCLITestReportGUID    = "6a1d4c2b-8e3f-4a5b-9c6d-7e8f9a0b1c2d"
	reportCLITestWorkorderGUID = "0f9e8d7c-6b5a-4c3d-8e2f-1a0b9c8d7e6f"
	reportCLITestSendKey       = "d4c3b2a1-9f8e-4d7c-8b6a-5f4e3d2c1b0a"
)

type recordedDomainRequest struct {
	method string
	path   string
	query  map[string][]string
	header http.Header
	body   map[string]any
}

// runRegisteredDomainCommand runs one action of a registered domain against a
// TLS test API that answers every request with an empty JSON object.
func runRegisteredDomainCommand(t *testing.T, domainName string, args ...string) ([]recordedDomainRequest, error) {
	t.Helper()
	domain := findDomain(domainName)
	if domain == nil {
		t.Fatalf("expected %s domain to be registered", domainName)
	}
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	if err := auth.SaveToken(&auth.TokenData{
		AccessToken: "registered-domain-token",
		ExpiresAt:   time.Now().Add(time.Hour),
		TenantGUID:  "55555555-5555-4555-8555-555555555555",
	}); err != nil {
		t.Fatalf("save test token: %v", err)
	}

	var requestsMu sync.Mutex
	var requests []recordedDomainRequest
	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		recorded := recordedDomainRequest{
			method: request.Method,
			path:   request.URL.Path,
			query:  request.URL.Query(),
			header: request.Header.Clone(),
		}
		bodyBytes, err := io.ReadAll(request.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
		}
		if len(bodyBytes) > 0 {
			if err := json.Unmarshal(bodyBytes, &recorded.body); err != nil {
				t.Errorf("decode body %q: %v", bodyBytes, err)
			}
		}
		requestsMu.Lock()
		requests = append(requests, recorded)
		requestsMu.Unlock()
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{}`))
	}))
	t.Cleanup(server.Close)

	apiClient := client.NewAPIClient(
		server.URL,
		time.Second,
		true,
		client.RetryOptions{MaxRetries: 0},
		logging.New(logging.LevelError),
	)
	format := "json"
	command := buildDomainCommand(
		domain,
		func() (*client.APIClient, error) { return apiClient, nil },
		logging.New(logging.LevelError),
		&format,
		&ExportConfig{},
	)
	command.SilenceErrors = true
	command.SilenceUsage = true
	command.SetArgs(args)

	var runErr error
	captureRegistryStdout(t, func() { runErr = command.Execute() })
	requestsMu.Lock()
	defer requestsMu.Unlock()
	return append([]recordedDomainRequest(nil), requests...), runErr
}

func TestReportPdfUsesGuidRouteAndOutFile(t *testing.T) {
	cases := []struct {
		action   string
		guid     string
		wantPath string
		wantName string
	}{
		{"pdf", reportCLITestReportGUID, "/api/report/by-guid/" + reportCLITestReportGUID + "/pdf",
			"work-report-{reportGuid}.pdf"},
		{"workorder-pdf", reportCLITestWorkorderGUID, "/api/report/workorder/by-guid/" + reportCLITestWorkorderGUID + "/pdf",
			"workorder-reports-{workorderGuid}.pdf"},
	}
	for _, test := range cases {
		t.Run(test.action, func(t *testing.T) {
			action := findDomainAction(t, "report", test.action)
			if !action.DownloadResponseBody || action.DownloadOutputFlag != "out" || action.DownloadURLField != "" {
				t.Fatalf("%s download = %+v, want a response-body download through --out", test.action, *action)
			}
			if action.DownloadDefaultName != test.wantName {
				t.Errorf("%s DownloadDefaultName = %q, want %q", test.action, action.DownloadDefaultName, test.wantName)
			}
			if len(action.Args) != 1 || action.Args[0].Type != "uuid" || !action.Args[0].Required {
				t.Fatalf("%s args = %+v, want one required GUID", test.action, action.Args)
			}
			costs := actionFlagByName(t, action, "include-costs")
			if costs.QueryName != "includeCosts" || costs.Type != "bool" || costs.Default != true {
				t.Errorf("%s include-costs = %+v, want a bool query flag defaulting to true", test.action, *costs)
			}

			outputPath := filepath.Join(t.TempDir(), "report.pdf")
			exportDir := t.TempDir()
			requests, stdout, err := runResponseBodyAction(t, *action, exportDir, test.guid, "--out", outputPath)
			if err != nil {
				t.Fatalf("%s error = %v", test.action, err)
			}
			if len(requests) != 1 {
				t.Fatalf("requests = %d, want 1", len(requests))
			}
			request := requests[0]
			if request.method != http.MethodGet || request.path != test.wantPath {
				t.Errorf("route = %s %s, want GET %s", request.method, request.path, test.wantPath)
			}
			if !reflect.DeepEqual(request.query, map[string][]string{"includeCosts": {"true"}}) {
				t.Errorf("query = %v, want only includeCosts=true (the --out path stays local)", request.query)
			}
			written, err := os.ReadFile(outputPath)
			if err != nil {
				t.Fatalf("read output: %v", err)
			}
			if !reflect.DeepEqual(written, responseBodyPDF) {
				t.Errorf("output = %q, want the raw response body", written)
			}
			if !strings.Contains(stdout, outputPath) {
				t.Errorf("stdout = %q, want the written path", stdout)
			}
			assertOnlyFiles(t, exportDir)

			noCostsPath := filepath.Join(t.TempDir(), "no-costs.pdf")
			requests, _, err = runResponseBodyAction(
				t, *action, t.TempDir(), test.guid, "--include-costs=false", "--out", noCostsPath,
			)
			if err != nil {
				t.Fatalf("%s --include-costs=false error = %v", test.action, err)
			}
			if len(requests) != 1 || !reflect.DeepEqual(requests[0].query, map[string][]string{"includeCosts": {"false"}}) {
				t.Errorf("requests = %+v, want one request with includeCosts=false", requests)
			}
		})
	}
}

func TestReportExportPostsRootJsonWithFormat(t *testing.T) {
	action := findDomainAction(t, "report", "export")
	if !action.DownloadResponseBody || action.DownloadDefaultName != "reports.{format}" {
		t.Fatalf("export download = %+v, want a response-body download named reports.{format}", *action)
	}
	format := actionFlagByName(t, action, "format")
	if format.Default != "csv" || strings.Join(format.AllowedValues, ",") != "csv,xlsx" {
		t.Errorf("format = %+v, want default csv limited to csv and xlsx", *format)
	}
	if fromJSON := actionFlagByName(t, action, "from-json"); !fromJSON.RootJSONObjectFile {
		t.Errorf("from-json = %+v, want a root JSON object file", *fromJSON)
	}

	filters := writeRegistryJSONFixture(t, `{
		"nameFilter": "pump",
		"workorderGuidFilter": "`+reportCLITestWorkorderGUID+`",
		"reportDateFrom": "2026-09-01"
	}`)
	outputPath := filepath.Join(t.TempDir(), "reports.xlsx")
	requests, _, err := runResponseBodyAction(
		t, *action, t.TempDir(), "--from-json", filters, "--format", "xlsx", "--out", outputPath,
	)
	if err != nil {
		t.Fatalf("export error = %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("requests = %d, want 1", len(requests))
	}
	request := requests[0]
	if request.method != http.MethodPost || request.path != "/api/report/export" {
		t.Errorf("route = %s %s, want POST /api/report/export", request.method, request.path)
	}
	wantBody := map[string]any{
		"nameFilter":          "pump",
		"workorderGuidFilter": reportCLITestWorkorderGUID,
		"reportDateFrom":      "2026-09-01",
		"format":              "xlsx",
	}
	if !reflect.DeepEqual(request.body, wantBody) {
		t.Errorf("body = %#v, want the unwrapped filters plus format %#v", request.body, wantBody)
	}
	if len(request.query) != 0 {
		t.Errorf("query = %v, want none", request.query)
	}
	if _, err := os.Stat(outputPath); err != nil {
		t.Errorf("export file not written: %v", err)
	}

	requests, _, err = runResponseBodyAction(
		t, *action, t.TempDir(), "--format", "pdf", "--out", filepath.Join(t.TempDir(), "reports.pdf"),
	)
	if err == nil || !strings.Contains(err.Error(), "--format") || len(requests) != 0 {
		t.Errorf("--format pdf: requests = %d, error = %v; want a local rejection", len(requests), err)
	}

	collision := writeRegistryJSONFixture(t, `{"format":"csv"}`)
	requests, _, err = runResponseBodyAction(
		t, *action, t.TempDir(), "--from-json", collision, "--out", filepath.Join(t.TempDir(), "reports.csv"),
	)
	if err == nil || !strings.Contains(err.Error(), "format") || len(requests) != 0 {
		t.Errorf("format in both the file and the flag: requests = %d, error = %v; want a local rejection",
			len(requests), err)
	}
}

func TestReportSendRequiresConfirmAndIdempotencyHeader(t *testing.T) {
	confirm := actionFlagByName(t, findDomainAction(t, "report", "send"), "confirm")
	if confirm.Type != "bool" || !confirm.Required || !confirm.MustBeTrue || confirm.LocalOnly {
		t.Fatalf("confirm = %+v, want a required must-be-true bool sent in the body", *confirm)
	}

	sendArgs := func(extra ...string) []string {
		return append([]string{
			"send", reportCLITestReportGUID,
			"--recipients", "owner@customer.example",
			"--recipients", "site@customer.example",
			"--message", "Work finished on the feed pump",
			"--publish-to-portal",
			"--idempotency-key", reportCLITestSendKey,
		}, extra...)
	}

	requests, err := runRegisteredDomainCommand(t, "report", sendArgs()...)
	if err == nil || !strings.Contains(err.Error(), "confirm") || len(requests) != 0 {
		t.Fatalf("send without --confirm: requests = %d, error = %v; want a local rejection", len(requests), err)
	}
	requests, err = runRegisteredDomainCommand(t, "report", sendArgs("--confirm=false")...)
	if err == nil || !strings.Contains(err.Error(), "--confirm must be explicitly enabled") || len(requests) != 0 {
		t.Fatalf("send with --confirm=false: requests = %d, error = %v; want a local rejection", len(requests), err)
	}

	requests, err = runRegisteredDomainCommand(t, "report", sendArgs("--confirm")...)
	if err != nil {
		t.Fatalf("send error = %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("requests = %d, want 1", len(requests))
	}
	request := requests[0]
	wantPath := "/api/report/by-guid/" + reportCLITestReportGUID + "/send"
	if request.method != http.MethodPost || request.path != wantPath {
		t.Errorf("route = %s %s, want POST %s", request.method, request.path, wantPath)
	}
	if got := request.header.Get("Idempotency-Key"); got != reportCLITestSendKey {
		t.Errorf("Idempotency-Key = %q, want %q", got, reportCLITestSendKey)
	}
	wantBody := map[string]any{
		"recipients":      []any{"owner@customer.example", "site@customer.example"},
		"message":         "Work finished on the feed pump",
		"publishToPortal": true,
		"confirm":         true,
	}
	if !reflect.DeepEqual(request.body, wantBody) {
		t.Errorf("body = %#v, want %#v (no key, no GUID in the body)", request.body, wantBody)
	}
}

func TestReportSendIdempotencyKeyIsRequired(t *testing.T) {
	action := findDomainAction(t, "report", "send")
	key := actionFlagByName(t, action, "idempotency-key")
	if key.HeaderName != "Idempotency-Key" || !key.Required || key.Type != "non-empty-uuid" ||
		key.MirrorHeaderInBody || key.BodyName != "" {
		t.Fatalf("idempotency-key = %+v, want a required header-only non-empty GUID", *key)
	}

	base := []string{"send", reportCLITestReportGUID, "--recipients", "owner@customer.example", "--confirm"}
	requests, err := runRegisteredDomainCommand(t, "report", base...)
	if err == nil || !strings.Contains(err.Error(), "idempotency-key") || len(requests) != 0 {
		t.Fatalf("send without a key: requests = %d, error = %v; want a local rejection", len(requests), err)
	}
	requests, err = runRegisteredDomainCommand(t, "report",
		append(base, "--idempotency-key", "00000000-0000-0000-0000-000000000000")...)
	if err == nil || !strings.Contains(err.Error(), "--idempotency-key") || len(requests) != 0 {
		t.Fatalf("send with an empty GUID key: requests = %d, error = %v; want a local rejection", len(requests), err)
	}
	requests, err = runRegisteredDomainCommand(t, "report", append(base, "--idempotency-key", "not-a-guid")...)
	if err == nil || !strings.Contains(err.Error(), "--idempotency-key") || len(requests) != 0 {
		t.Fatalf("send with a malformed key: requests = %d, error = %v; want a local rejection", len(requests), err)
	}
}

func TestReportFinalizePostsGuidRoute(t *testing.T) {
	action := findDomainAction(t, "report", "finalize")
	if len(action.Flags) != 0 {
		t.Errorf("finalize flags = %+v, want none", action.Flags)
	}

	requests, err := runRegisteredDomainCommand(t, "report", "finalize", reportCLITestReportGUID)
	if err != nil {
		t.Fatalf("finalize error = %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("requests = %d, want 1", len(requests))
	}
	request := requests[0]
	wantPath := "/api/report/by-guid/" + reportCLITestReportGUID + "/finalize"
	if request.method != http.MethodPost || request.path != wantPath {
		t.Errorf("route = %s %s, want POST %s", request.method, request.path, wantPath)
	}
	if len(request.body) != 0 || len(request.query) != 0 {
		t.Errorf("body = %#v, query = %v; want neither (the GUID is in the route)", request.body, request.query)
	}

	requests, err = runRegisteredDomainCommand(t, "report", "finalize", "12")
	if err == nil || !strings.Contains(err.Error(), "must be a GUID") || len(requests) != 0 {
		t.Errorf("finalize with an integer id: requests = %d, error = %v; want a local rejection", len(requests), err)
	}
}

func TestReportUpdateSendsOverrideReasonInBody(t *testing.T) {
	reason := actionFlagByName(t, findDomainAction(t, "report", "update"), "override-reason")
	if reason.BodyName != "overrideReason" || reason.QueryName != "" || reason.Required || reason.Type != "string" {
		t.Fatalf("override-reason = %+v, want an optional string sent as body field overrideReason", *reason)
	}

	requests, err := runRegisteredDomainCommand(t, "report",
		"update", reportCLITestReportGUID,
		"--description", "Replaced the seal",
		"--report-date", "2026-09-24",
		"--override-reason", "Customer asked for the corrected hours",
	)
	if err != nil {
		t.Fatalf("update error = %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("requests = %d, want 1", len(requests))
	}
	request := requests[0]
	if request.method != http.MethodPut || request.path != "/api/report/by-guid/"+reportCLITestReportGUID {
		t.Errorf("route = %s %s, want PUT /api/report/by-guid/%s", request.method, request.path, reportCLITestReportGUID)
	}
	wantBody := map[string]any{
		"description":    "Replaced the seal",
		"reportDate":     "2026-09-24",
		"overrideReason": "Customer asked for the corrected hours",
	}
	if !reflect.DeepEqual(request.body, wantBody) {
		t.Errorf("body = %#v, want %#v", request.body, wantBody)
	}
	if len(request.query) != 0 {
		t.Errorf("query = %v, want none", request.query)
	}
}

func TestReportDeleteSendsOverrideReasonAsQuery(t *testing.T) {
	reason := actionFlagByName(t, findDomainAction(t, "report", "delete"), "override-reason")
	if reason.QueryName != "overrideReason" || reason.BodyName != "" || reason.Required || reason.Type != "string" {
		t.Fatalf("override-reason = %+v, want an optional string sent as query overrideReason", *reason)
	}

	const overrideReason = "Duplicate report & filed twice"
	requests, err := runRegisteredDomainCommand(t, "report",
		"delete", reportCLITestReportGUID, "--override-reason", overrideReason,
	)
	if err != nil {
		t.Fatalf("delete error = %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("requests = %d, want 1", len(requests))
	}
	request := requests[0]
	if request.method != http.MethodDelete || request.path != "/api/report/by-guid/"+reportCLITestReportGUID {
		t.Errorf("route = %s %s, want DELETE /api/report/by-guid/%s", request.method, request.path, reportCLITestReportGUID)
	}
	if !reflect.DeepEqual(request.query, map[string][]string{"overrideReason": {overrideReason}}) {
		t.Errorf("query = %v, want the escaped overrideReason only", request.query)
	}
	if len(request.body) != 0 {
		t.Errorf("body = %#v, want none", request.body)
	}

	requests, err = runRegisteredDomainCommand(t, "report", "delete", reportCLITestReportGUID)
	if err != nil {
		t.Fatalf("delete without a reason error = %v", err)
	}
	if len(requests) != 1 || len(requests[0].query) != 0 {
		t.Errorf("requests = %+v, want one request without a query", requests)
	}
}
