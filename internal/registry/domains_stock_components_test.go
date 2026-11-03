package registry

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/uteamup/cli/internal/auth"
	"github.com/uteamup/cli/internal/client"
	"github.com/uteamup/cli/internal/logging"
)

func TestComponentOperationTransportPreservesRetryTermsAndExplicitZeroSequences(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	const unit = "11111111-1111-4111-8111-111111111111"
	const operation = "22222222-2222-4222-8222-222222222222"
	const position = "33333333-3333-4333-8333-333333333333"
	const revision = "44444444-4444-4444-8444-444444444444"
	const replacement = "55555555-5555-4555-8555-555555555555"
	const effective = "2026-11-03T10:00:00Z"
	var bodies []map[string]any
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/stock/units/"+unit+"/component-operation" {
			t.Errorf("wrong operation route: %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("X-Requested-With") != "XMLHttpRequest" {
			t.Error("missing mutation CSRF header")
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		for _, key := range []string{"unitGuid", "unitId", "tenantGuid", "tenantId", "userId", "actorGuid"} {
			if _, exists := body[key]; exists {
				t.Errorf("route or authority field leaked: %s", key)
			}
		}
		bodies = append(bodies, body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"operationId":"` + operation + `","unit":{"guid":"` + unit + `","status":"Scrapped"}}`))
	}))
	defer server.Close()
	if err := auth.SaveToken(&auth.TokenData{APIOrigin: server.URL, AccessToken: "component-test-token", ExpiresAt: time.Now().Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}
	logger := logging.New(logging.LevelError)
	format := "json"
	factory := func() (*client.APIClient, error) {
		return client.NewAPIClient(server.URL, time.Second, true, client.RetryOptions{MaxRetries: 0}, logger), nil
	}
	for attempt := 0; attempt < 2; attempt++ {
		command := buildActionCommand(findStockDomain(t), *findStockAction(t, "component-apply"), factory, logger, &format, nil)
		command.SetArgs([]string{unit, "--operation-id", operation, "--kind", "Replace",
			"--expected-unit-sequence", "0", "--effective-at", effective, "--position-guid", position,
			"--expected-position-revision", revision, "--replacement-unit-guid", replacement,
			"--expected-replacement-unit-sequence", "0", "--removed-disposition", "Scrapped", "--reason", "Reviewed replacement"})
		if err := command.Execute(); err != nil {
			t.Fatal(err)
		}
	}
	if len(bodies) != 2 || !reflect.DeepEqual(bodies[0], bodies[1]) {
		t.Fatalf("retry changed original terms: %+v", bodies)
	}
	for key, want := range map[string]any{"operationId": operation, "kind": "Replace", "effectiveAt": effective,
		"expectedUnitSequence": float64(0), "expectedReplacementUnitSequence": float64(0),
		"positionGuid": position, "expectedPositionRevision": revision, "replacementUnitGuid": replacement,
		"removedDisposition": "Scrapped", "reason": "Reviewed replacement"} {
		if got, exists := bodies[0][key]; !exists || got != want {
			t.Errorf("body[%s] = %v (%t), want %v", key, got, exists, want)
		}
	}
}

func TestComponentReadQueriesAndPositionCreateUseTheirGuidRoutes(t *testing.T) {
	const guid = "11111111-1111-4111-8111-111111111111"
	for _, test := range []struct {
		name, method, suffix string
		flags                []string
	}{
		{"component-positions", "GET", "assets/" + guid + "/component-positions", nil},
		{"component-position-create", "POST", "assets/" + guid + "/component-positions", []string{"--operation-id", guid, "--position-key", "bulb-1", "--name", "Bulb1"}},
		{"unit-lifecycle-facts", "GET", "units/" + guid + "/lifecycle-facts", []string{"--page", "2", "--page-size", "10"}},
		{"unit-installation-at", "GET", "units/" + guid + "/installation-at", []string{"--observed-at", "2026-11-02T10:00:00Z"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			requests := 0
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests++
				if r.Method != test.method || r.URL.Path != "/api/stock/"+test.suffix {
					t.Errorf("wrong route: %s %s", r.Method, r.URL.Path)
				}
				if test.name == "unit-installation-at" && r.URL.Query().Get("observedAt") != "2026-11-02T10:00:00Z" {
					t.Error("observation instant changed or missing")
				}
				if test.name == "unit-lifecycle-facts" && (r.URL.Query().Get("page") != "2" || r.URL.Query().Get("pageSize") != "10") {
					t.Error("fact pagination changed")
				}
				if test.method == "POST" {
					var body map[string]any
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Error(err)
					}
					if body["operationId"] != guid || body["positionKey"] != "bulb-1" || body["name"] != "Bulb1" {
						t.Errorf("wrong position body: %+v", body)
					}
					if _, leaked := body["assetGuid"]; leaked {
						t.Error("route identity leaked into create body")
					}
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"state":"UnknownBeforeRecordedHistory"}`))
			}))
			defer server.Close()
			if err := auth.SaveToken(&auth.TokenData{APIOrigin: server.URL, AccessToken: "component-read-test-token", ExpiresAt: time.Now().Add(time.Hour)}); err != nil {
				t.Fatal(err)
			}
			logger := logging.New(logging.LevelError)
			format := "json"
			factory := func() (*client.APIClient, error) {
				return client.NewAPIClient(server.URL, time.Second, true, client.RetryOptions{MaxRetries: 0}, logger), nil
			}
			command := buildActionCommand(findStockDomain(t), *findStockAction(t, test.name), factory, logger, &format, nil)
			command.SetArgs(append([]string{guid}, test.flags...))
			if err := command.Execute(); err != nil {
				t.Fatal(err)
			}
			if requests != 1 {
				t.Fatalf("requests = %d, want1", requests)
			}
		})
	}
}

func TestComponentCommandsRejectMalformedIdentityOrMissingReviewBeforeTransport(t *testing.T) {
	for _, args := range [][]string{
		{"not-a-guid"},
		{"00000000-0000-0000-0000-000000000000"},
		{"11111111-1111-4111-8111-111111111111", "--kind", "Replace"},
		{"11111111-1111-4111-8111-111111111111", "unexpected-extra-identity"},
	} {
		created := false
		factory := func() (*client.APIClient, error) { created = true; return nil, nil }
		format := "json"
		command := buildActionCommand(findStockDomain(t), *findStockAction(t, "component-apply"), factory, logging.New(logging.LevelError), &format, nil)
		command.SetArgs(args)
		if err := command.Execute(); err == nil || created {
			t.Fatalf("unreviewed request reached transport: %+v", args)
		}
	}
}
