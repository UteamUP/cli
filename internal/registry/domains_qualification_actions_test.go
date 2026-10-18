package registry

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/uteamup/cli/internal/client"
)

func TestQualificationActionsUseCanonicalGUIDRequestsForBothTokenModes(t *testing.T) {
	const heldGUID = "11111111-1111-4111-8111-111111111111"
	const skillGUID = "22222222-2222-4222-8222-222222222222"
	for _, mode := range []string{"login", "apikey"} {
		for _, scenario := range []struct {
			domain, action, method, path, body string
			held                               bool
		}{
			{"skill", "add-mine", "POST", "/api/skill/user-skill/by-guid", `{"skillGuid":"` + skillGUID + `","proficiencyLevel":2,"yearsOfExperience":0,"notes":null}`, false},
			{"skill", "update-mine", "PUT", "/api/skill/user-skill/by-guid/" + heldGUID, `{"proficiencyLevel":3,"yearsOfExperience":0,"notes":"Observed experience"}`, true},
			{"certificate", "update-mine", "PUT", "/api/certificate/user-certificate/by-guid/" + heldGUID + "/details", `{"certificateNumber":"CERT-2","documentGuid":null}`, true},
			{"certificate", "renew-with-proof", "POST", "/api/certificate/user-certificate/by-guid/" + heldGUID + "/renew-with-proof", `{"newExpirationDate":"2027-10-18T00:00:00Z","newDocumentGuid":"` + skillGUID + `"}`, true},
		} {
			t.Run(mode+"/"+scenario.domain+"/"+scenario.action, func(t *testing.T) {
				action := findDomainAction(t, scenario.domain, scenario.action)
				file := actionFlagByName(t, action, "from-json")
				if !file.Required || !file.RootJSONObjectFile || file.JSONFile || action.MCPOnly {
					t.Fatal("canonical write must send an unwrapped request to its owning REST route")
				}
				for _, arg := range action.Args {
					if arg.Type != "non-empty-uuid" || !arg.Required {
						t.Fatalf("identity must be a nonempty public GUID: %+v", arg)
					}
				}
				var want map[string]any
				if err := json.Unmarshal([]byte(scenario.body), &want); err != nil {
					t.Fatal(err)
				}
				var calls atomic.Int32
				apiClient := projectTransportClient(t, mode, client.RetryOptions{}, func(response http.ResponseWriter, request *http.Request) {
					calls.Add(1)
					if request.Method != scenario.method || request.URL.Path != scenario.path || request.URL.RawQuery != "" {
						t.Errorf("request = %s %s", request.Method, request.URL)
					}
					var got map[string]any
					if err := json.NewDecoder(request.Body).Decode(&got); err != nil {
						t.Error(err)
					}
					if !reflect.DeepEqual(got, want) {
						t.Errorf("request body = %#v, want %#v", got, want)
					}
					for _, key := range []string{"userId", "tenantGuid", "skillId", "certificateId", "documentId", "newDocumentId"} {
						if _, exists := got[key]; exists {
							t.Errorf("public request included internal/actor selector %s", key)
						}
					}
					_, _ = response.Write([]byte(`{"externalGuid":"` + heldGUID + `"}`))
				})
				args := []string{scenario.action}
				if scenario.held {
					args = append(args, heldGUID)
				}
				args = append(args, "--from-json", writeRegistryJSONFixture(t, scenario.body))
				if err := executeProjectTransport(t, apiClient, scenario.domain, args); err != nil {
					t.Fatal(err)
				}
				if calls.Load() != 1 {
					t.Fatalf("calls = %d, want one request", calls.Load())
				}
			})
		}
	}
}

func TestQualificationProofOmissionAndNullRemainDistinct(t *testing.T) {
	for _, body := range []string{`{"notes":"Keep existing proof"}`, `{"documentGuid":null}`} {
		parsed, err := readRootJSONObjectFile(writeRegistryJSONFixture(t, body))
		if err != nil {
			t.Fatal(err)
		}
		_, supplied := parsed["documentGuid"]
		if supplied != strings.Contains(body, "documentGuid") {
			t.Fatalf("proof presence changed: %#v", parsed)
		}
	}
}

func TestQualificationRenewalFailureIsNotReplayed(t *testing.T) {
	var calls atomic.Int32
	apiClient := projectTransportClient(t, "login", client.RetryOptions{MaxRetries: 2}, func(response http.ResponseWriter, request *http.Request) {
		calls.Add(1)
		response.WriteHeader(http.StatusServiceUnavailable)
	})
	err := executeProjectTransport(t, apiClient, "certificate", []string{
		"renew-with-proof", "11111111-1111-4111-8111-111111111111", "--from-json",
		writeRegistryJSONFixture(t, `{"newExpirationDate":"2027-10-18T00:00:00Z"}`),
	})
	if err == nil || calls.Load() != 1 {
		t.Fatalf("uncertain write must surface once, err=%v calls=%d", err, calls.Load())
	}
	if !strings.Contains(findDomainAction(t, "certificate", "renew-with-proof").Description, "another holder requires Certificate.ViewUserCertificates") {
		t.Fatal("REST renewal help must disclose its actual cross-holder boundary")
	}
}
