package registry

import (
	"bytes"
	"encoding/json"
	"errors"
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

func TestBuildCommandsCreatesAPIClientWhenActionExecutes(t *testing.T) {
	wantErr := errors.New("client factory reached")
	called := 0
	factory := func() (*client.APIClient, error) {
		called++
		return nil, wantErr
	}

	registry := &Registry{domains: []*Domain{{
		Name: "sample",
		Actions: []Action{{
			Name:     "list",
			ToolName: "SampleList",
		}},
	}}}
	format := "json"
	commands := registry.BuildCommands(
		factory,
		logging.New(logging.LevelError),
		&format,
		&ExportConfig{},
	)

	if called != 0 {
		t.Fatalf("client factory called during command registration: %d", called)
	}

	commands[0].SetArgs([]string{"list"})
	err := commands[0].Execute()
	if !errors.Is(err, wantErr) {
		t.Fatalf("Execute() error = %v, want %v", err, wantErr)
	}
	if called != 1 {
		t.Fatalf("client factory calls = %d, want 1", called)
	}
}

const responseBodyReportGUID = "3f2b8c1e-5d4a-4b6f-9e7d-1a2b3c4d5e6f"

var responseBodyPDF = []byte("%PDF-1.7\n\x00\x01binary-report-body\n%%EOF")

type recordedResponseBodyRequest struct {
	method string
	path   string
	query  map[string][]string
	body   map[string]any
}

func responseBodyPDFAction() Action {
	return Action{
		Name:       "pdf",
		ToolName:   "UteamupWorkReportPdf",
		HTTPMethod: http.MethodGet,
		RESTPath:   "by-guid/{reportGuid}/pdf",
		Args:       []ArgDef{{Name: "reportGuid", Type: "non-empty-uuid", Required: true}},
		Flags: []FlagDef{
			{Name: "include-costs", QueryName: "includeCosts", Type: "bool", Default: true},
			{Name: "out", Type: "string", Description: "Output file"},
		},
		DownloadResponseBody: true,
		DownloadOutputFlag:   "out",
		DownloadDefaultName:  "work-report-{reportGuid}.pdf",
	}
}

func responseBodyExportAction() Action {
	return Action{
		Name:       "export",
		ToolName:   "UteamupWorkReportExport",
		HTTPMethod: http.MethodPost,
		RESTPath:   "export",
		Flags: []FlagDef{
			{Name: "format", Type: "string", Default: "csv", AllowedValues: []string{"csv", "xlsx"}},
			{Name: "out", Type: "string", Description: "Output file"},
		},
		DownloadResponseBody: true,
		DownloadOutputFlag:   "out",
		DownloadDefaultName:  "reports.{format}",
	}
}

// runResponseBodyAction executes one action against a TLS test API with
// the profile JSON export enabled into exportDir.
func runResponseBodyAction(
	t *testing.T,
	action Action,
	exportDir string,
	args ...string,
) ([]recordedResponseBodyRequest, string, error) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	if err := auth.SaveToken(&auth.TokenData{
		AccessToken: "response-body-download-token",
		ExpiresAt:   time.Now().Add(time.Hour),
		TenantGUID:  "55555555-5555-4555-8555-555555555555",
	}); err != nil {
		t.Fatalf("save test token: %v", err)
	}

	var requestsMu sync.Mutex
	var requests []recordedResponseBodyRequest
	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		recorded := recordedResponseBodyRequest{
			method: request.Method,
			path:   request.URL.Path,
			query:  request.URL.Query(),
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
		response.Header().Set("Content-Type", "application/pdf")
		_, _ = response.Write(responseBodyPDF)
	}))
	t.Cleanup(server.Close)

	domain := &Domain{Name: "report", APIPath: "/api/report", Actions: []Action{action}}
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
		&ExportConfig{Enabled: true, Dir: exportDir},
	)
	command.SilenceErrors = true
	command.SilenceUsage = true
	command.SetArgs(append([]string{action.Name}, args...))

	var runErr error
	stdout := captureRegistryStdout(t, func() { runErr = command.Execute() })
	requestsMu.Lock()
	defer requestsMu.Unlock()
	return append([]recordedResponseBodyRequest(nil), requests...), stdout, runErr
}

func captureRegistryStdout(t *testing.T, run func()) string {
	t.Helper()
	original := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = writer
	defer func() { os.Stdout = original }()
	run()
	_ = writer.Close()
	captured, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	return string(captured)
}

func assertOnlyFiles(t *testing.T, dir string, want ...string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	var got []string
	for _, entry := range entries {
		got = append(got, entry.Name())
	}
	if len(got) != len(want) || (len(want) > 0 && !reflect.DeepEqual(got, want)) {
		t.Fatalf("files in %s = %v, want %v", dir, got, want)
	}
}

func TestDownloadResponseBodyWritesFile(t *testing.T) {
	outputDir := t.TempDir()
	exportDir := t.TempDir()
	outputPath := filepath.Join(outputDir, "service-report.pdf")

	requests, stdout, err := runResponseBodyAction(
		t, responseBodyPDFAction(), exportDir, responseBodyReportGUID, "--out", outputPath,
	)
	if err != nil {
		t.Fatalf("pdf command error = %v", err)
	}

	written, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if !bytes.Equal(written, responseBodyPDF) {
		t.Fatalf("output bytes = %q, want the raw response body %q", written, responseBodyPDF)
	}
	assertOnlyFiles(t, outputDir, "service-report.pdf")

	var summary map[string]any
	if err := json.Unmarshal([]byte(stdout), &summary); err != nil {
		t.Fatalf("stdout %q is not the JSON summary: %v", stdout, err)
	}
	if summary["path"] != outputPath || summary["bytes"] != float64(len(responseBodyPDF)) {
		t.Fatalf("summary = %#v, want path %q and %d bytes", summary, outputPath, len(responseBodyPDF))
	}
	if len(summary) != 2 {
		t.Fatalf("summary carries extra fields: %#v", summary)
	}

	if len(requests) != 1 {
		t.Fatalf("requests = %d, want 1", len(requests))
	}
	wantPath := "/api/report/by-guid/" + responseBodyReportGUID + "/pdf"
	if requests[0].method != http.MethodGet || requests[0].path != wantPath {
		t.Fatalf("request = %s %s, want GET %s", requests[0].method, requests[0].path, wantPath)
	}
	if got := requests[0].query["includeCosts"]; !reflect.DeepEqual(got, []string{"true"}) {
		t.Fatalf("includeCosts query = %v, want [true]", got)
	}
	assertOnlyFiles(t, exportDir)
}

func TestDownloadResponseBodyRefusesExistingFile(t *testing.T) {
	outputDir := t.TempDir()
	outputPath := filepath.Join(outputDir, "existing.pdf")
	if err := os.WriteFile(outputPath, []byte("keep me"), 0o600); err != nil {
		t.Fatalf("seed existing file: %v", err)
	}

	requests, _, err := runResponseBodyAction(
		t, responseBodyPDFAction(), t.TempDir(), responseBodyReportGUID, "--out", outputPath,
	)
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("error = %v, want an already-exists refusal", err)
	}
	kept, readErr := os.ReadFile(outputPath)
	if readErr != nil || string(kept) != "keep me" {
		t.Fatalf("existing file changed: %q, %v", kept, readErr)
	}
	if len(requests) != 0 {
		t.Fatalf("requests = %d, want 0: the refusal must come before the API call", len(requests))
	}
	assertOnlyFiles(t, outputDir, "existing.pdf")
}

func TestDownloadResponseBodyDefaultNameFromArg(t *testing.T) {
	workingDir := t.TempDir()
	t.Chdir(workingDir)

	_, stdout, err := runResponseBodyAction(t, responseBodyPDFAction(), t.TempDir(), responseBodyReportGUID)
	if err != nil {
		t.Fatalf("pdf command error = %v", err)
	}

	wantName := "work-report-" + responseBodyReportGUID + ".pdf"
	written, err := os.ReadFile(filepath.Join(workingDir, wantName))
	if err != nil {
		t.Fatalf("default file %s was not written: %v", wantName, err)
	}
	if !bytes.Equal(written, responseBodyPDF) {
		t.Fatalf("default file bytes = %q", written)
	}
	assertOnlyFiles(t, workingDir, wantName)
	if !strings.Contains(stdout, wantName) {
		t.Fatalf("stdout %q does not name %s", stdout, wantName)
	}
}

func TestDownloadResponseBodyDefaultNameFromFormat(t *testing.T) {
	workingDir := t.TempDir()
	t.Chdir(workingDir)

	requests, _, err := runResponseBodyAction(t, responseBodyExportAction(), t.TempDir(), "--format", "xlsx")
	if err != nil {
		t.Fatalf("export command error = %v", err)
	}

	assertOnlyFiles(t, workingDir, "reports.xlsx")
	if len(requests) != 1 || requests[0].method != http.MethodPost || requests[0].path != "/api/report/export" {
		t.Fatalf("requests = %#v, want one POST /api/report/export", requests)
	}
	if !reflect.DeepEqual(requests[0].body, map[string]any{"format": "xlsx"}) {
		t.Fatalf("body = %#v, want the format to stay in the request", requests[0].body)
	}
}

func TestDownloadResponseBodyOutFlagNotSentInBody(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "filtered.csv")

	requests, _, err := runResponseBodyAction(
		t, responseBodyExportAction(), t.TempDir(), "--format", "csv", "--out", outputPath,
	)
	if err != nil {
		t.Fatalf("export command error = %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("requests = %d, want 1", len(requests))
	}
	if !reflect.DeepEqual(requests[0].body, map[string]any{"format": "csv"}) {
		t.Fatalf("body = %#v; the local --out path leaked into the request", requests[0].body)
	}
	if len(requests[0].query) != 0 {
		t.Fatalf("query = %v; the local --out path leaked into the query", requests[0].query)
	}

	pdfPath := filepath.Join(t.TempDir(), "report.pdf")
	pdfRequests, _, err := runResponseBodyAction(
		t, responseBodyPDFAction(), t.TempDir(), responseBodyReportGUID, "--out", pdfPath,
	)
	if err != nil {
		t.Fatalf("pdf command error = %v", err)
	}
	if _, leaked := pdfRequests[0].query["out"]; leaked {
		t.Fatalf("GET query = %v; the local --out path leaked", pdfRequests[0].query)
	}
}

func TestDownloadResponseBodyRejectsURLFieldCombo(t *testing.T) {
	combo := responseBodyPDFAction()
	combo.DownloadURLField = "sasUrl"
	if err := validateActionDefinition(combo); err == nil ||
		!strings.Contains(err.Error(), "download URL field") {
		t.Fatalf("URL-field combo error = %v", err)
	}

	missingFlag := responseBodyPDFAction()
	missingFlag.DownloadOutputFlag = "destination"
	if err := validateActionDefinition(missingFlag); err == nil ||
		!strings.Contains(err.Error(), "string flag") {
		t.Fatalf("undeclared output flag error = %v", err)
	}

	boolFlag := responseBodyPDFAction()
	boolFlag.Flags = []FlagDef{{Name: "out", Type: "bool"}}
	if err := validateActionDefinition(boolFlag); err == nil ||
		!strings.Contains(err.Error(), "string flag") {
		t.Fatalf("bool output flag error = %v", err)
	}

	if err := validateActionDefinition(responseBodyPDFAction()); err != nil {
		t.Fatalf("valid response-body action rejected: %v", err)
	}

	outputPath := filepath.Join(t.TempDir(), "combo.pdf")
	requests, _, err := runResponseBodyAction(t, combo, t.TempDir(), responseBodyReportGUID, "--out", outputPath)
	if err == nil || !strings.Contains(err.Error(), "download URL field") {
		t.Fatalf("executing the combo error = %v", err)
	}
	if len(requests) != 0 {
		t.Fatalf("requests = %d, want 0 for an invalid definition", len(requests))
	}
	if _, statErr := os.Stat(outputPath); !os.IsNotExist(statErr) {
		t.Fatalf("invalid definition wrote %s", outputPath)
	}
}
