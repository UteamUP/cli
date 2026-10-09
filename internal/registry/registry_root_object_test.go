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
	"testing"
	"time"

	"github.com/uteamup/cli/internal/auth"
	"github.com/uteamup/cli/internal/client"
	"github.com/uteamup/cli/internal/logging"
)

func TestReadRootJSONObjectFileAcceptsAnUnwrappedEvolvingDTO(t *testing.T) {
	path := writeRegistryJSONFixture(t, `{
		"title":"Pump seal out of tolerance",
		"futureNestedField":{"enabled":true,"values":[1,2]},
		"evidenceGuids":[]
	}`)

	object, err := readRootJSONObjectFile(path)
	if err != nil {
		t.Fatalf("readRootJSONObjectFile() error = %v", err)
	}
	if object["title"] != "Pump seal out of tolerance" {
		t.Errorf("title = %v", object["title"])
	}
	nested, ok := object["futureNestedField"].(map[string]any)
	if !ok || nested["enabled"] != true {
		t.Errorf("future nested field = %#v", object["futureNestedField"])
	}
	if _, ok := object["request"]; ok {
		t.Errorf("root-object primitive introduced a request wrapper: %#v", object)
	}
}

func TestRootObjectKeepsOriginalNumericLexemesAndPrecision(t *testing.T) {
	for _, literal := range []string{"1", "1.0", "1e0", "9007199254740993", "0.000000000000000000123"} {
		t.Run(literal, func(t *testing.T) {
			object, err := readRootJSONObjectFile(writeRegistryJSONFixture(t,
				`{"mutationOutcomeVersion":`+literal+`,"nested":[`+literal+`]}`))
			if err != nil {
				t.Fatal(err)
			}
			encoded, err := json.Marshal(object)
			if err != nil {
				t.Fatal(err)
			}
			var raw map[string]json.RawMessage
			if err := json.Unmarshal(encoded, &raw); err != nil {
				t.Fatal(err)
			}
			if string(raw["mutationOutcomeVersion"]) != literal || string(raw["nested"]) != "["+literal+"]" {
				t.Fatalf("original reviewed numbers changed: %s", encoded)
			}
		})
	}
}

func TestReviewedIceReplacementPreservesRawRequestAndConfirmationBoundary(t *testing.T) {
	for _, literal := range []string{"1", "1.0", "1e0"} {
		t.Run(literal, func(t *testing.T) {
			const contact = "11111111-2222-4333-8444-555555555555"
			const operation = "aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee"
			requestPath := writeRegistryJSONFixture(t, `{"userGuids":[],"idempotencyKey":"`+operation+
				`","expectedAssignmentsHash":"`+strings.Repeat("a", 64)+`","mutationOutcomeVersion":`+literal+`}`)
			t.Setenv("HOME", t.TempDir())
			t.Setenv("USERPROFILE", os.Getenv("HOME"))
			calls := 0
			server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
				calls++
				if request.Method != http.MethodPut || request.URL.Path != "/api/emergencycontact/by-contact/"+contact+"/assignments" {
					t.Errorf("wrong normal route: %s %s", request.Method, request.URL.Path)
				}
				if request.Header.Get("Idempotency-Key") != operation {
					t.Error("original operation header changed")
				}
				var body map[string]json.RawMessage
				if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
					t.Error(err)
				}
				if len(body) != 4 || string(body["mutationOutcomeVersion"]) != literal || string(body["idempotencyKey"]) != `"`+operation+`"` {
					t.Errorf("original reviewed wire changed: %#v", body)
				}
				response.Header().Set("Content-Type", "application/json")
				if literal != "1" {
					response.WriteHeader(http.StatusBadRequest)
					_, _ = response.Write([]byte(`{"title":"Exact version literal1 is required"}`))
					return
				}
				_, _ = response.Write([]byte(`{"schemaVersion":1,"outcome":"applied"}`))
			}))
			defer server.Close()
			if err := auth.SaveToken(&auth.TokenData{APIOrigin: server.URL, AccessToken: "ice-fixture-token",
				ExpiresAt: time.Now().Add(time.Hour), TenantGUID: "55555555-5555-4555-8555-555555555555"}); err != nil {
				t.Fatal(err)
			}
			api := client.NewAPIClient(server.URL, time.Second, true, client.RetryOptions{MaxRetries: 0}, logging.New(logging.LevelError))
			format := "json"
			command := buildDomainCommand(findDomain("contact"), func() (*client.APIClient, error) { return api, nil },
				logging.New(logging.LevelError), &format, &ExportConfig{})
			command.SilenceErrors, command.SilenceUsage = true, true
			command.SetArgs([]string{"ice-assignments-replace-reviewed", contact, "--from-json", requestPath,
				"--idempotency-key", operation, "--confirm"})
			err := command.Execute()
			if (literal == "1" && err != nil) || (literal != "1" && err == nil) || calls != 1 {
				t.Fatalf("calls=%d error=%v version=%s", calls, err, literal)
			}
		})
	}
}

func TestReviewedIceFalseConfirmationNeverCreatesClient(t *testing.T) {
	created := false
	format := "json"
	command := buildDomainCommand(findDomain("contact"), func() (*client.APIClient, error) {
		created = true
		return nil, nil
	}, logging.New(logging.LevelError), &format, &ExportConfig{})
	command.SilenceErrors, command.SilenceUsage = true, true
	command.SetArgs([]string{"ice-assignments-replace-reviewed", "11111111-2222-4333-8444-555555555555",
		"--from-json", filepath.Join(t.TempDir(), "unused.json"),
		"--idempotency-key", "aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee", "--confirm=false"})
	if err := command.Execute(); err == nil || !strings.Contains(err.Error(), "--confirm must be explicitly enabled") {
		t.Fatalf("confirmation error = %v", err)
	}
	if created {
		t.Fatal("client created before explicit confirmation")
	}
}

func TestReadRootJSONObjectFileRejectsAmbiguousOrNonObjectJSON(t *testing.T) {
	tests := []struct {
		name string
		json string
		want string
	}{
		{name: "root array", json: `[]`, want: "one JSON object at the root"},
		{name: "root string", json: `"draft"`, want: "one JSON object at the root"},
		{name: "root number", json: `42`, want: "one JSON object at the root"},
		{name: "root null", json: `null`, want: "one JSON object at the root"},
		{name: "duplicate root field", json: `{"reason":"one","reason":"two"}`, want: `duplicate JSON object field "reason"`},
		{name: "duplicate nested field", json: `{"evidence":{"guid":"one","guid":"two"}}`, want: `duplicate JSON object field "guid"`},
		{name: "multiple roots", json: `{} {}`, want: "multiple root values"},
		{name: "known lowercase wrapper", json: `{"request":{"reason":"reviewed"}}`, want: `without a "request" wrapper`},
		{name: "known mixed-case wrapper", json: `{"Payload":[]}`, want: `without a "Payload" wrapper`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := readRootJSONObjectFile(writeRegistryJSONFixture(t, test.json))
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want containing %q", err, test.want)
			}
		})
	}
}

func TestMergeRootJSONObjectIsAtomicAndReportsTheFirstSortedCollision(t *testing.T) {
	destination := map[string]any{"b": "original-b", "a": "original-a"}
	original := map[string]any{"b": "original-b", "a": "original-a"}
	source := map[string]any{"c": "new-c", "b": "new-b", "a": "new-a"}

	err := mergeRootJSONObject(destination, source)
	if err == nil || !strings.Contains(err.Error(), `field "a"`) {
		t.Fatalf("mergeRootJSONObject() error = %v, want deterministic first collision a", err)
	}
	if !reflect.DeepEqual(destination, original) {
		t.Fatalf("failed merge mutated destination: got %#v, want %#v", destination, original)
	}

	if err := mergeRootJSONObject(destination, map[string]any{"c": "new-c"}); err != nil {
		t.Fatalf("non-colliding merge error = %v", err)
	}
	if destination["c"] != "new-c" {
		t.Errorf("merged c = %v", destination["c"])
	}
}

func TestRootObjectAndLocalOnlyDefinitionsFailClosed(t *testing.T) {
	valid := Action{Flags: []FlagDef{
		{Name: "request-file", Type: "string", RootJSONObjectFile: true},
		{Name: "confirm", Type: "bool", Required: true, MustBeTrue: true, LocalOnly: true},
	}}
	if err := validateActionDefinition(valid); err != nil {
		t.Fatalf("valid declaration error = %v", err)
	}

	tests := []struct {
		name   string
		action Action
		want   string
	}{
		{
			name: "root file cannot add a wrapper",
			action: Action{Flags: []FlagDef{
				{Name: "request-file", Type: "string", RootJSONObjectFile: true, BodyName: "request"},
			}},
			want: "unwrapped root-object JSON file",
		},
		{
			name: "only one root file",
			action: Action{Flags: []FlagDef{
				{Name: "first", Type: "string", RootJSONObjectFile: true},
				{Name: "second", Type: "string", RootJSONObjectFile: true},
			}},
			want: "only one root-object JSON file",
		},
		{
			name: "local flag cannot enter body",
			action: Action{Flags: []FlagDef{
				{Name: "confirm", Type: "bool", LocalOnly: true, BodyName: "confirm"},
			}},
			want: "local-only boolean",
		},
		{
			name: "local flag must be boolean",
			action: Action{Flags: []FlagDef{
				{Name: "confirm", Type: "string", LocalOnly: true},
			}},
			want: "local-only boolean",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateActionDefinition(test.action)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want containing %q", err, test.want)
			}
		})
	}
}

func TestJSONFlagReadsRootObjectFileInsteadOfSendingItsPath(t *testing.T) {
	requestPath := writeRegistryJSONFixture(t, `{
		"title":"Pump seal investigation",
		"problemStatement":"Seal temperature exceeded the reviewed limit"
	}`)
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", request.Method)
		}
		if request.URL.Path != "/api/json-helper" {
			t.Errorf("path = %q, want /api/json-helper", request.URL.Path)
		}
		bodyBytes, err := io.ReadAll(request.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
		}
		var body map[string]any
		if err := json.Unmarshal(bodyBytes, &body); err != nil {
			t.Errorf("decode body: %v", err)
		}
		want := map[string]any{
			"title":            "Pump seal investigation",
			"problemStatement": "Seal temperature exceeded the reviewed limit",
		}
		if !reflect.DeepEqual(body, want) {
			t.Errorf("body = %#v, want %#v", body, want)
		}
		if body["fromJson"] == requestPath {
			t.Errorf("local file path leaked into request body: %#v", body)
		}
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{}`))
	}))
	t.Cleanup(server.Close)
	if err := auth.SaveToken(&auth.TokenData{
		APIOrigin:   server.URL,
		AccessToken: "json-helper-contract-token",
		ExpiresAt:   time.Now().Add(time.Hour),
		TenantGUID:  "55555555-5555-4555-8555-555555555555",
	}); err != nil {
		t.Fatalf("save test token: %v", err)
	}

	domain := &Domain{
		Name:    "json-helper",
		APIPath: "/api/json-helper",
		Actions: []Action{{
			Name:       "create",
			HTTPMethod: http.MethodPost,
			Flags:      []FlagDef{jsonFlag()},
		}},
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
	command.SetArgs([]string{"create", "--from-json", requestPath})

	if err := command.Execute(); err != nil {
		t.Fatalf("create command error = %v", err)
	}
}

func TestExplicitDomainBaseRoutingIsNarrowAndCannotEscape(t *testing.T) {
	domain := &Domain{Name: "sample-record", APIPath: "/api/quality/sample-records"}

	defaultSearch, _ := buildRESTPath(domain, Action{Name: "search"}, nil)
	if defaultSearch != "/api/quality/sample-records/search" {
		t.Errorf("default search path = %q", defaultSearch)
	}
	explicitSearch, consumed := buildRESTPath(domain, Action{
		Name:              "search",
		UseDomainBasePath: true,
	}, nil)
	if explicitSearch != "/api/quality/sample-records" || len(consumed) != 0 {
		t.Errorf("explicit base path = %q, consumed = %v", explicitSearch, consumed)
	}

	invalid := Action{
		Name:              "search",
		UseDomainBasePath: true,
		RESTBasePath:      "/api/admin/escape",
		RESTPath:          "escape",
	}
	path, _ := buildRESTPath(domain, invalid, nil)
	if path != domain.APIPath {
		t.Errorf("domain-base routing escaped to %q", path)
	}
	if err := validateActionDefinition(invalid); err == nil {
		t.Fatal("conflicting domain-base route declaration was accepted")
	}
}

func TestNCRTransitionConfirmationIsValidatedLocallyAndNotSerialized(t *testing.T) {
	domain := findNCRDomain(t)
	requestPath := writeRegistryJSONFixture(t, `{"reason":"Reviewed and approved"}`)
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", request.Method)
		}
		wantPath := "/api/quality/non-conformances/" + testNonConformanceGUID + "/transitions/ncr.submit"
		if request.URL.Path != wantPath {
			t.Errorf("path = %q, want %q", request.URL.Path, wantPath)
		}
		if request.Header.Get("Idempotency-Key") != testEvidenceGUID {
			t.Errorf("Idempotency-Key = %q", request.Header.Get("Idempotency-Key"))
		}
		if request.Header.Get("If-Match") != `"ncr-version-7"` {
			t.Errorf("If-Match = %q", request.Header.Get("If-Match"))
		}
		bodyBytes, err := io.ReadAll(request.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
		}
		var body map[string]any
		if err := json.Unmarshal(bodyBytes, &body); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if !reflect.DeepEqual(body, map[string]any{"reason": "Reviewed and approved"}) {
			t.Errorf("body = %#v; confirmation or route identity leaked", body)
		}
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{"transitionAvailability":{"options":[]}}`))
	}))
	t.Cleanup(server.Close)
	if err := auth.SaveToken(&auth.TokenData{
		APIOrigin:   server.URL,
		AccessToken: "ncr-contract-token",
		ExpiresAt:   time.Now().Add(time.Hour),
		TenantGUID:  "55555555-5555-4555-8555-555555555555",
	}); err != nil {
		t.Fatalf("save test token: %v", err)
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
	command.SetArgs([]string{
		"transition",
		testNonConformanceGUID,
		"ncr.submit",
		"--request-file", requestPath,
		"--idempotency-key", testEvidenceGUID,
		"--concurrency-token-file", protectedRegistrySecret(t, "ncr-version-7"),
		"--confirm",
	})
	if err := command.Execute(); err != nil {
		t.Fatalf("transition command error = %v", err)
	}
}

func TestNCRTransitionRejectsExplicitFalseConfirmationBeforeCreatingAClient(t *testing.T) {
	domain := findNCRDomain(t)
	clientCreated := false
	format := "json"
	command := buildDomainCommand(
		domain,
		func() (*client.APIClient, error) {
			clientCreated = true
			return nil, nil
		},
		logging.New(logging.LevelError),
		&format,
		&ExportConfig{},
	)
	command.SilenceErrors = true
	command.SilenceUsage = true
	command.SetArgs([]string{
		"transition",
		testNonConformanceGUID,
		"ncr.submit",
		"--request-file", filepath.Join(t.TempDir(), "unused.json"),
		"--idempotency-key", testEvidenceGUID,
		"--concurrency-token-file", protectedRegistrySecret(t, "ncr-version-7"),
		"--confirm=false",
	})

	err := command.Execute()
	if err == nil || !strings.Contains(err.Error(), "--confirm must be explicitly enabled") {
		t.Fatalf("error = %v", err)
	}
	if clientCreated {
		t.Fatal("API client was created before confirmation validation")
	}
}

func writeRegistryJSONFixture(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "request.json")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write JSON fixture: %v", err)
	}
	return path
}
