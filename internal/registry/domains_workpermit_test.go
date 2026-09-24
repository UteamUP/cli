package registry

import (
	"encoding/json"
	"net/http"
	"reflect"
	"sync/atomic"
	"testing"

	"github.com/uteamup/cli/internal/client"
)

const (
	biosecurityWorkorderGUID = "55555555-5555-4555-8555-555555555555"
	biosecurityLocationGUID  = "44444444-4444-4444-8444-444444444444"
)

func TestWorkPermitBiosecurityActionsAreRegistered(t *testing.T) {
	domain := findDomain("workpermit")
	if domain == nil {
		t.Fatal("expected the workpermit domain to be registered")
	}
	if domain.APIPath != "/api/workpermit" {
		t.Fatalf("workpermit APIPath = %q, want /api/workpermit", domain.APIPath)
	}

	status := findDomainAction(t, "workpermit", "biosecurity-status")
	if status.RESTPath != "biosecurity-status" || status.HTTPMethod != "GET" || status.ToolName != "UteamupBiosecurityStatus" {
		t.Fatalf("biosecurity-status = %s %q tool %q", status.HTTPMethod, status.RESTPath, status.ToolName)
	}
	if len(status.Args) != 0 || len(status.Flags) != 1 {
		t.Fatalf("biosecurity-status takes exactly one flag and no args: %+v", status)
	}
	workorder := findFlag(status, "workorder-guid")
	if workorder == nil || !workorder.Required || workorder.Type != "non-empty-uuid" || workorder.QueryName != "workorderGuid" {
		t.Fatalf("--workorder-guid must be a required GUID query parameter: %+v", workorder)
	}

	entry := findDomainAction(t, "workpermit", "biosecurity-entry")
	if entry.RESTPath != "biosecurity-entry" || entry.HTTPMethod != "POST" {
		t.Fatalf("biosecurity-entry = %s %q", entry.HTTPMethod, entry.RESTPath)
	}
	if len(entry.Args) != 0 || len(entry.Flags) != 2 {
		t.Fatalf("biosecurity-entry takes exactly two flags and no args: %+v", entry)
	}
	location := findFlag(entry, "location-guid")
	if location == nil || !location.Required || location.Type != "non-empty-uuid" || location.QueryName != "" {
		t.Fatalf("--location-guid must be a required GUID body field: %+v", location)
	}
	optional := findFlag(entry, "workorder-guid")
	if optional == nil || optional.Required || optional.Type != "uuid" || optional.QueryName != "" {
		t.Fatalf("--workorder-guid must be an optional GUID body field: %+v", optional)
	}
}

func TestWorkPermitBiosecurityStatusSendsTheWorkorderAsAQuery(t *testing.T) {
	var calls atomic.Int32
	apiClient := projectTransportClient(t, "login", client.RetryOptions{}, func(response http.ResponseWriter, request *http.Request) {
		calls.Add(1)
		if request.Method != http.MethodGet ||
			request.URL.Path != "/api/workpermit/biosecurity-status" ||
			request.URL.Query().Get("workorderGuid") != biosecurityWorkorderGUID {
			t.Errorf("request = %s %s, want GET /api/workpermit/biosecurity-status?workorderGuid=%s",
				request.Method, request.URL, biosecurityWorkorderGUID)
		}
		_, _ = response.Write([]byte(`{"workorderGuid":"` + biosecurityWorkorderGUID + `","isZone":true,"policy":"block","cleared":false}`))
	})

	err := executeProjectTransport(t, apiClient, "workpermit", []string{
		"biosecurity-status", "--workorder-guid", biosecurityWorkorderGUID,
	})
	if err != nil {
		t.Fatalf("biosecurity-status failed: %v", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("HTTP calls = %d, want one", calls.Load())
	}
}

func TestWorkPermitBiosecurityEntryPostsTheZoneAndTheWorkorder(t *testing.T) {
	for _, scenario := range []struct {
		name string
		args []string
		want map[string]any
	}{
		{
			name: "with a work order",
			args: []string{"biosecurity-entry", "--location-guid", biosecurityLocationGUID, "--workorder-guid", biosecurityWorkorderGUID},
			want: map[string]any{"locationGuid": biosecurityLocationGUID, "workorderGuid": biosecurityWorkorderGUID},
		},
		{
			name: "zone only",
			args: []string{"biosecurity-entry", "--location-guid", biosecurityLocationGUID},
			want: map[string]any{"locationGuid": biosecurityLocationGUID},
		},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			var calls atomic.Int32
			apiClient := projectTransportClient(t, "apikey", client.RetryOptions{}, func(response http.ResponseWriter, request *http.Request) {
				calls.Add(1)
				if request.Method != http.MethodPost || request.URL.Path != "/api/workpermit/biosecurity-entry" || request.URL.RawQuery != "" {
					t.Errorf("request = %s %s, want POST /api/workpermit/biosecurity-entry", request.Method, request.URL)
				}
				if request.Header.Get("X-Requested-With") != "XMLHttpRequest" {
					t.Error("mutation lost its CSRF request header")
				}
				var body map[string]any
				if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if !reflect.DeepEqual(body, scenario.want) {
					t.Errorf("body = %#v, want %#v", body, scenario.want)
				}
				response.WriteHeader(http.StatusCreated)
				_, _ = response.Write([]byte(`{"guid":"66666666-6666-4666-8666-666666666666","status":"draft","isDisinfectionPermit":true}`))
			})

			if err := executeProjectTransport(t, apiClient, "workpermit", scenario.args); err != nil {
				t.Fatalf("biosecurity-entry failed: %v", err)
			}
			if calls.Load() != 1 {
				t.Fatalf("HTTP calls = %d, want one", calls.Load())
			}
		})
	}
}

func TestWorkPermitBiosecurityCommandsRefuseMissingOrMalformedGuidsBeforeSending(t *testing.T) {
	for _, args := range [][]string{
		{"biosecurity-entry"},
		{"biosecurity-entry", "--location-guid", "not-a-guid"},
		{"biosecurity-entry", "--location-guid", "00000000-0000-0000-0000-000000000000"},
		{"biosecurity-status"},
		{"biosecurity-status", "--workorder-guid", "1234"},
	} {
		var calls atomic.Int32
		apiClient := projectTransportClient(t, "login", client.RetryOptions{}, func(response http.ResponseWriter, request *http.Request) {
			calls.Add(1)
		})
		if err := executeProjectTransport(t, apiClient, "workpermit", args); err == nil {
			t.Errorf("%v must fail before sending", args)
		}
		if calls.Load() != 0 {
			t.Errorf("%v sent %d requests, want none", args, calls.Load())
		}
	}
}
