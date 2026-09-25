package registry

import (
	"bytes"
	"encoding/json"
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

const reportTemplateTestGUID = "7c1d2e3f-4a5b-4c6d-8e9f-0a1b2c3d4e5f"

var reportTemplateTestXLSX = []byte("PK\x03\x04\x00\x01report-template-xlsx-body")

func TestReportTemplateRegistryIsGuidOnlyAndWired(t *testing.T) {
	domain := findDomain("report-template")
	if domain == nil {
		t.Fatal("expected report-template domain to be registered")
	}
	if domain.APIPath != "/api/reports/templates" {
		t.Fatalf("APIPath = %q, want /api/reports/templates", domain.APIPath)
	}
	if !reflect.DeepEqual(domain.Aliases, []string{"rt", "report-templates"}) {
		t.Fatalf("Aliases = %v, want [rt report-templates]", domain.Aliases)
	}

	want := []struct {
		name, method, path, tool string
		domainBase               bool
	}{
		{"list", "GET", "", "UteamupReportTemplateList", true},
		{"get", "GET", "by-guid/{templateGuid}", "UteamupReportTemplateGet", false},
		{"preview", "POST", "by-guid/{templateGuid}/preview", "UteamupReportTemplatePreview", false},
		{"export", "GET", "by-guid/{templateGuid}/export", "UteamupReportTemplateExport", false},
		{"delete", "DELETE", "by-guid/{templateGuid}", "UteamupReportTemplateDelete", false},
	}
	if len(domain.Actions) != len(want) {
		t.Fatalf("actions = %d, want %d (no create or update: authoring is web-only)", len(domain.Actions), len(want))
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
				t.Fatalf("%s args = %+v, want exactly templateGuid", action.Name, action.Args)
			}
			arg := action.Args[0]
			if arg.Name != "templateGuid" || arg.Type != "non-empty-uuid" || !arg.Required {
				t.Errorf("%s arg = %+v, want required non-empty-uuid templateGuid", action.Name, arg)
			}
		}

		for _, flag := range action.Flags {
			lowered := strings.ToLower(flag.Name + " " + flag.BodyName + " " + flag.QueryName)
			for _, forbidden := range []string{"tenant", "user", "sql"} {
				if strings.Contains(lowered, forbidden) {
					t.Errorf("%s flag --%s must not carry %q", action.Name, flag.Name, forbidden)
				}
			}
		}
	}

	list := findDomainAction(t, "report-template", "list")
	listQueries := map[string]string{}
	for _, flag := range list.Flags {
		listQueries[flag.Name] = flag.QueryName
	}
	wantQueries := map[string]string{"scope": "scope", "search": "search", "page": "page", "page-size": "pageSize"}
	if !reflect.DeepEqual(listQueries, wantQueries) {
		t.Errorf("list flags = %v, want %v", listQueries, wantQueries)
	}
	if scope := findFlag(list, "scope"); scope == nil ||
		!reflect.DeepEqual(scope.AllowedValues, []string{"all", "mine", "shared"}) {
		t.Errorf("--scope = %+v, want allowed values all, mine, shared", scope)
	}
	if pageSize := findFlag(list, "page-size"); pageSize == nil || pageSize.Default != 50 {
		t.Errorf("--page-size = %+v, want default 50 like the REST query model", pageSize)
	}

	export := findDomainAction(t, "report-template", "export")
	format := findFlag(export, "format")
	if format == nil || format.Default != "csv" ||
		!reflect.DeepEqual(format.AllowedValues, []string{"csv", "xlsx", "pdf"}) {
		t.Errorf("--format = %+v, want default csv and allowed csv, xlsx, pdf", format)
	}
	if !export.DownloadResponseBody || export.DownloadOutputFlag != "out" ||
		export.DownloadDefaultName != "report-template-{templateGuid}.{format}" || findFlag(export, "out") == nil {
		t.Errorf("export download = body %v flag %q default %q, want the --out response-body download",
			export.DownloadResponseBody, export.DownloadOutputFlag, export.DownloadDefaultName)
	}
}

func TestReportTemplateExportWritesResponseBodyToOutFile(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "costs.xlsx")

	requests, stdout, err := runReportTemplateExport(t, reportTemplateTestGUID, "--format", "xlsx", "--out", outputPath)
	if err != nil {
		t.Fatalf("export command error = %v", err)
	}

	written, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if !bytes.Equal(written, reportTemplateTestXLSX) {
		t.Fatalf("output bytes = %q, want the raw response body", written)
	}

	var summary map[string]any
	if err := json.Unmarshal([]byte(stdout), &summary); err != nil {
		t.Fatalf("stdout %q is not the JSON summary: %v", stdout, err)
	}
	if summary["path"] != outputPath || summary["bytes"] != float64(len(reportTemplateTestXLSX)) {
		t.Fatalf("summary = %#v, want path %q and %d bytes", summary, outputPath, len(reportTemplateTestXLSX))
	}

	if len(requests) != 1 {
		t.Fatalf("requests = %d, want 1", len(requests))
	}
	wantPath := "/api/reports/templates/by-guid/" + reportTemplateTestGUID + "/export"
	if requests[0].method != http.MethodGet || requests[0].path != wantPath {
		t.Fatalf("request = %s %s, want GET %s", requests[0].method, requests[0].path, wantPath)
	}
	if !reflect.DeepEqual(requests[0].query, map[string][]string{"format": {"xlsx"}}) {
		t.Fatalf("query = %v, want only format=xlsx (the --out path stays local)", requests[0].query)
	}

	again, _, err := runReportTemplateExport(t, reportTemplateTestGUID, "--format", "xlsx", "--out", outputPath)
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("second export error = %v, want an already-exists refusal", err)
	}
	if len(again) != 0 {
		t.Fatalf("second export made %d requests; the refusal must come before the API call", len(again))
	}
	if kept, _ := os.ReadFile(outputPath); !bytes.Equal(kept, reportTemplateTestXLSX) {
		t.Fatalf("existing file changed to %q", kept)
	}

	workingDir := t.TempDir()
	t.Chdir(workingDir)
	if _, _, err := runReportTemplateExport(t, reportTemplateTestGUID); err != nil {
		t.Fatalf("default-name export error = %v", err)
	}
	assertOnlyFiles(t, workingDir, "report-template-"+reportTemplateTestGUID+".csv")
}

// runReportTemplateExport runs the registered report-template export action
// against a TLS test API that answers with an XLSX body.
func runReportTemplateExport(t *testing.T, args ...string) ([]recordedResponseBodyRequest, string, error) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	if err := auth.SaveToken(&auth.TokenData{
		AccessToken: "report-template-export-token",
		ExpiresAt:   time.Now().Add(time.Hour),
		TenantGUID:  "55555555-5555-4555-8555-555555555555",
	}); err != nil {
		t.Fatalf("save test token: %v", err)
	}

	var requestsMu sync.Mutex
	var requests []recordedResponseBodyRequest
	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		requestsMu.Lock()
		requests = append(requests, recordedResponseBodyRequest{
			method: request.Method,
			path:   request.URL.Path,
			query:  request.URL.Query(),
		})
		requestsMu.Unlock()
		response.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		_, _ = response.Write(reportTemplateTestXLSX)
	}))
	t.Cleanup(server.Close)

	domain := findDomain("report-template")
	if domain == nil {
		t.Fatal("expected report-template domain to be registered")
	}
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
	command.SetArgs(append([]string{"export"}, args...))

	var runErr error
	stdout := captureRegistryStdout(t, func() { runErr = command.Execute() })
	requestsMu.Lock()
	defer requestsMu.Unlock()
	return append([]recordedResponseBodyRequest(nil), requests...), stdout, runErr
}
