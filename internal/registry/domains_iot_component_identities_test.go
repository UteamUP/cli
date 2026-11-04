package registry

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/uteamup/cli/internal/auth"
	"github.com/uteamup/cli/internal/client"
	"github.com/uteamup/cli/internal/logging"
)

func TestIoTIdentityReviewTransportPreservesExactAliasChannelsAndOriginalRetry(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	const source = "11111111-1111-4111-8111-111111111111"
	const generation = "22222222-2222-4222-8222-222222222222"
	const revision = "33333333-3333-4333-8333-333333333333"
	const operation = "44444444-4444-4444-8444-444444444444"
	const unit = "55555555-5555-4555-8555-555555555555"
	const effective = "2026-11-02T10:00:00Z"
	identity := filepath.Join(t.TempDir(), "identity.json")
	channels := filepath.Join(t.TempDir(), "channels.json")
	for path, body := range map[string]string{
		identity: `{"namespace":"provider.hardware","stableHardwareId":"Thread-Device-CaseSensitive"}`,
		channels: `[{"channelNamespace":"homeassistant.entity","stableChannelId":"sensor.bulb_power","attribute":"power","attributeDefinitionGuid":"66666666-6666-4666-8666-666666666666","unit":"W","valueKind":"gauge","quality":"measured","expectsTelemetry":true}]`,
	} {
		if err := os.WriteFile(path, []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	var bodies []map[string]any
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/iot/devices/"+source+"/identities/review" {
			t.Errorf("wrong reviewed-identity route: %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("X-Requested-With") != "XMLHttpRequest" {
			t.Error("mutation CSRF header missing")
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		bodies = append(bodies, body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"identityGuid":"` + operation + `","bindings":[]}`))
	}))
	defer server.Close()
	if err := auth.SaveToken(&auth.TokenData{APIOrigin: server.URL, AccessToken: "isolated-identity-review-token", ExpiresAt: time.Now().Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}
	logger := logging.New(logging.LevelError)
	format := "json"
	factory := func() (*client.APIClient, error) {
		return client.NewAPIClient(server.URL, time.Second, true, client.RetryOptions{MaxRetries: 0}, logger), nil
	}
	for attempt := 0; attempt < 2; attempt++ {
		command := buildActionCommand(findDomain("iot"), findIoTAction(t, "physical-identity-review"), factory, logger, &format, nil)
		command.SetArgs([]string{source, "--operation-id", operation, "--expected-source-revision", revision,
			"--source-generation-guid", generation, "--identity-file", identity, "--channels-file", channels,
			"--unit-guid", unit, "--expected-unit-sequence", "0", "--effective-from", effective, "--reason", "Reviewed Thread identity"})
		if err := command.Execute(); err != nil {
			t.Fatal(err)
		}
	}
	if len(bodies) != 2 || !reflect.DeepEqual(bodies[0], bodies[1]) {
		t.Fatalf("retry changed reviewed terms: %+v", bodies)
	}
	for field, expected := range map[string]any{"operationId": operation, "expectedSourceRevision": revision,
		"sourceGenerationGuid": generation, "unitGuid": unit, "expectedUnitSequence": float64(0), "effectiveFrom": effective} {
		if actual, present := bodies[0][field]; !present || actual != expected {
			t.Errorf("body[%s] = %v (%t), want %v", field, actual, present, expected)
		}
	}
	alias, ok := bodies[0]["identity"].(map[string]any)
	if !ok || alias["stableHardwareId"] != "Thread-Device-CaseSensitive" || alias["namespace"] != "provider.hardware" {
		t.Fatalf("exact identity changed: %+v", bodies[0]["identity"])
	}
	rows, ok := bodies[0]["channels"].([]any)
	if !ok || len(rows) != 1 || rows[0].(map[string]any)["quality"] != "measured" {
		t.Fatalf("reviewed channels changed: %+v", bodies[0]["channels"])
	}
	for _, field := range []string{"sourceGuid", "tenantGuid", "tenantId", "userId", "sourceSecret", "expectedIdentityRevision", "effectiveTo"} {
		if _, present := bodies[0][field]; present {
			t.Errorf("unexpected route, authority or omitted field %s", field)
		}
	}
}

func TestIoTSourceAndIdentityActionsUseExactRoutesAndCredentialPurpose(t *testing.T) {
	const source = "11111111-1111-4111-8111-111111111111"
	const generation = "22222222-2222-4222-8222-222222222222"
	const initialRevision = "00000000-0000-0000-0000-000000000000"
	review := []string{"--operation-id", source, "--expected-source-revision", generation, "--reason", "Reviewed source"}
	initialReview := []string{"--operation-id", source, "--expected-source-revision", initialRevision, "--reason", "Reviewed legacy source"}
	for _, test := range []struct {
		name, method, suffix string
		args                 []string
		wantSourceRevision   string
	}{
		{"source-generations", "GET", "source-generations", nil, ""},
		{"source-enroll", "POST", "source-generations", append(append([]string{}, review...), "--source-kind", "homeassistant"), generation},
		{"source-enroll", "POST", "source-generations", append(append([]string{}, initialReview...), "--source-kind", "homeassistant"), initialRevision},
		{"source-retire", "POST", "source-generations/retire", append(append([]string{}, review...), "--source-generation-guid", generation), generation},
		{"source-key-issue", "POST", "api-keys", []string{"--name", "Collector", "--credential-purpose", "componentSourceV2", "--source-generation-guid", generation}, ""},
		{"physical-identities", "GET", "identities", []string{"--generation-guid", generation, "--page", "2", "--page-size", "10"}, ""},
		{"physical-identity-end", "POST", "identities/" + generation + "/end", append(append([]string{generation}, review...), "--expected-identity-revision", generation, "--expected-unit-sequence", "0", "--effective-at", "2026-11-02T10:00:00Z", "--block-identity"), generation},
	} {
		name := test.name
		if test.wantSourceRevision == initialRevision {
			name += "/initial-zero-revision"
		}
		t.Run(name, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			requests := 0
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests++
				if r.Method != test.method || r.URL.Path != "/api/iot/devices/"+source+"/"+test.suffix {
					t.Errorf("wrong source route: %s %s", r.Method, r.URL.Path)
				}
				if test.name == "physical-identities" && (r.URL.Query().Get("generationGuid") != generation || r.URL.Query().Get("page") != "2" || r.URL.Query().Get("pageSize") != "10") {
					t.Errorf("generation/pagination query changed: %s", r.URL.RawQuery)
				}
				if test.method == "POST" {
					var body map[string]any
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Error(err)
					}
					if r.Header.Get("X-Requested-With") != "XMLHttpRequest" {
						t.Error("missing mutation CSRF header")
					}
					if test.wantSourceRevision != "" && body["expectedSourceRevision"] != test.wantSourceRevision {
						t.Errorf("reviewed source revision changed: got %v, want %s", body["expectedSourceRevision"], test.wantSourceRevision)
					}
					if test.name == "source-key-issue" && (body["credentialPurpose"] != "componentSourceV2" || body["sourceGenerationGuid"] != generation) {
						t.Errorf("key enrollment purpose/generation changed: %+v", body)
					}
					if test.name == "physical-identity-end" && (body["blockIdentity"] != true || body["expectedUnitSequence"] != float64(0)) {
						t.Errorf("explicit ending terms changed: %+v", body)
					}
					for _, field := range []string{"sourceGuid", "identityGuid", "tenantGuid", "actorId", "sourceSecret"} {
						if _, present := body[field]; present {
							t.Errorf("unexpected identity/authority field: %s", field)
						}
					}
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"accepted":true}`))
			}))
			defer server.Close()
			if err := auth.SaveToken(&auth.TokenData{APIOrigin: server.URL, AccessToken: "isolated-source-test-token", ExpiresAt: time.Now().Add(time.Hour)}); err != nil {
				t.Fatal(err)
			}
			logger := logging.New(logging.LevelError)
			factory := func() (*client.APIClient, error) {
				return client.NewAPIClient(server.URL, time.Second, true, client.RetryOptions{MaxRetries: 0}, logger), nil
			}
			format := "json"
			command := buildActionCommand(findDomain("iot"), findIoTAction(t, test.name), factory, logger, &format, nil)
			command.SetArgs(append([]string{source}, test.args...))
			if err := command.Execute(); err != nil {
				t.Fatal(err)
			}
			if requests != 1 {
				t.Errorf("requests = %d, want1", requests)
			}
		})
	}
}

func TestIoTComponentCommandsRejectMissingReviewAndLegacyPurposeBeforeTransport(t *testing.T) {
	const guid = "11111111-1111-4111-8111-111111111111"
	const empty = "00000000-0000-0000-0000-000000000000"
	for _, test := range []struct {
		action string
		args   []string
	}{
		{"source-generations", []string{"not-a-guid"}},
		{"source-enroll", []string{guid, "--source-kind", "homeassistant"}},
		{"source-enroll", []string{empty, "--operation-id", guid, "--expected-source-revision", empty, "--reason", "Reviewed source", "--source-kind", "homeassistant"}},
		{"source-enroll", []string{guid, "--operation-id", empty, "--expected-source-revision", empty, "--reason", "Reviewed source", "--source-kind", "homeassistant"}},
		{"source-enroll", []string{guid, "--operation-id", guid, "--expected-source-revision", "not-a-guid", "--reason", "Reviewed source", "--source-kind", "homeassistant"}},
		{"source-retire", []string{guid}},
		{"source-retire", []string{guid, "--operation-id", guid, "--expected-source-revision", empty, "--source-generation-guid", guid, "--reason", "Reviewed source"}},
		{"physical-identities", []string{guid}},
		{"physical-identity-review", []string{guid, "--unit-guid", guid}},
		{"source-key-issue", []string{guid, "--name", "Collector", "--credential-purpose", "legacyDeviceTelemetryV1", "--source-generation-guid", guid}},
		{"source-key-issue", []string{guid, "--name", "Collector", "--credential-purpose", "componentSourceV2", "--source-generation-guid", "00000000-0000-0000-0000-000000000000"}},
		{"source-generations", []string{guid, "--tenant-guid", guid}},
	} {
		created := false
		factory := func() (*client.APIClient, error) { created = true; return nil, nil }
		format := "json"
		command := buildActionCommand(findDomain("iot"), findIoTAction(t, test.action), factory, logging.New(logging.LevelError), &format, nil)
		command.SetArgs(test.args)
		if err := command.Execute(); err == nil || created {
			t.Errorf("%s accepted invalid review or created transport: %v", test.action, test.args)
		}
	}
}
