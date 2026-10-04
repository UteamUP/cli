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

const (
	workPermitLinkPermitGUID       = "77777777-7777-4777-8777-777777777777"
	workPermitLinkTargetGUID       = "88888888-8888-4888-8888-888888888888"
	workPermitLinkGUID             = "99999999-9999-4999-8999-999999999999"
	workPermitPrerequisitePermitID = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
)

func TestWorkPermitLinkAndPrerequisiteActionsAreRegistered(t *testing.T) {
	for _, want := range []struct {
		name, method, tool, path string
		flags                    []string
	}{
		{"links", "GET", "UteamupWorkPermitLinksList", "by-guid/{workPermitGuid}/links", []string{"permit"}},
		{"link-add", "POST", "UteamupWorkPermitLinkAdd", "by-guid/{workPermitGuid}/links", []string{"permit", "type", "target"}},
		{"link-remove", "DELETE", "UteamupWorkPermitLinkRemove", "links/by-guid/{linkGuid}", []string{"link"}},
		{"linked-to", "GET", "UteamupWorkPermitsLinkedTo", "linked/{linkType}/{targetGuid}", []string{"type", "target"}},
		{"prerequisite-add", "POST", "UteamupWorkPermitPrerequisiteAdd", "by-guid/{workPermitGuid}/prerequisites", []string{"permit", "prerequisite"}},
		{"prerequisite-remove", "DELETE", "UteamupWorkPermitPrerequisiteRemove", "by-guid/{workPermitGuid}/prerequisites/{prerequisiteGuid}", []string{"permit", "prerequisite"}},
	} {
		action := findDomainAction(t, "workpermit", want.name)
		if action.HTTPMethod != want.method || action.ToolName != want.tool || action.RESTPath != want.path {
			t.Errorf("%s = %s %q tool %q, want %s %q tool %q",
				want.name, action.HTTPMethod, action.RESTPath, action.ToolName, want.method, want.path, want.tool)
		}
		if len(action.Args) != 0 || len(action.Flags) != len(want.flags) {
			t.Errorf("%s must take exactly the flags %v and no positional args: %+v", want.name, want.flags, action)
		}
		for _, name := range want.flags {
			flag := findFlag(action, name)
			if flag == nil || !flag.Required {
				t.Errorf("%s --%s must be a required flag: %+v", want.name, name, flag)
				continue
			}
			if name != "type" && flag.Type != "non-empty-uuid" {
				t.Errorf("%s --%s type = %q, want non-empty-uuid", want.name, name, flag.Type)
			}
		}
		if err := validateActionDefinition(*action); err != nil {
			t.Errorf("%s definition is invalid: %v", want.name, err)
		}
	}

	if got := findFlag(findDomainAction(t, "workpermit", "link-add"), "type").AllowedValues; !reflect.DeepEqual(got,
		[]string{"WorkOrder", "Asset", "Chemical", "KnowledgeArticle", "Tool", "Certificate"}) {
		t.Errorf("link-add --type allowed values = %v", got)
	}
	if got := findFlag(findDomainAction(t, "workpermit", "linked-to"), "type").AllowedValues; !reflect.DeepEqual(got,
		[]string{"workorder", "asset", "chemical", "knowledgearticle", "tool", "certificate"}) {
		t.Errorf("linked-to --type allowed values = %v", got)
	}
}

func TestWorkPermitLinkAndPrerequisiteActionsSendTheBackendRoutes(t *testing.T) {
	permitBase := "/api/workpermit/by-guid/" + workPermitLinkPermitGUID
	for _, scenario := range []struct {
		name   string
		args   []string
		method string
		path   string
		body   map[string]any
	}{
		{
			name:   "links",
			args:   []string{"links", "--permit", workPermitLinkPermitGUID},
			method: http.MethodGet, path: permitBase + "/links",
		},
		{
			name:   "link-add",
			args:   []string{"link-add", "--permit", workPermitLinkPermitGUID, "--type", "KnowledgeArticle", "--target", workPermitLinkTargetGUID},
			method: http.MethodPost, path: permitBase + "/links",
			body: map[string]any{"linkType": "KnowledgeArticle", "targetGuid": workPermitLinkTargetGUID},
		},
		{
			name:   "link-remove",
			args:   []string{"link-remove", "--link", workPermitLinkGUID},
			method: http.MethodDelete, path: "/api/workpermit/links/by-guid/" + workPermitLinkGUID,
		},
		{
			name:   "linked-to",
			args:   []string{"linked-to", "--type", "workorder", "--target", workPermitLinkTargetGUID},
			method: http.MethodGet, path: "/api/workpermit/linked/workorder/" + workPermitLinkTargetGUID,
		},
		{
			name:   "prerequisite-add",
			args:   []string{"prerequisite-add", "--permit", workPermitLinkPermitGUID, "--prerequisite", workPermitPrerequisitePermitID},
			method: http.MethodPost, path: permitBase + "/prerequisites",
			body: map[string]any{"prerequisiteGuid": workPermitPrerequisitePermitID},
		},
		{
			name:   "prerequisite-remove",
			args:   []string{"prerequisite-remove", "--permit", workPermitLinkPermitGUID, "--prerequisite", workPermitPrerequisitePermitID},
			method: http.MethodDelete, path: permitBase + "/prerequisites/" + workPermitPrerequisitePermitID,
		},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			var calls atomic.Int32
			apiClient := projectTransportClient(t, "login", client.RetryOptions{}, func(response http.ResponseWriter, request *http.Request) {
				calls.Add(1)
				if request.Method != scenario.method || request.URL.Path != scenario.path || request.URL.RawQuery != "" {
					t.Errorf("request = %s %s, want %s %s", request.Method, request.URL, scenario.method, scenario.path)
				}
				if scenario.method != http.MethodGet && request.Header.Get("X-Requested-With") != "XMLHttpRequest" {
					t.Error("mutation lost its CSRF request header")
				}
				if scenario.body != nil {
					var body map[string]any
					if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
						t.Errorf("decode body: %v", err)
					}
					if !reflect.DeepEqual(body, scenario.body) {
						t.Errorf("body = %#v, want %#v (route GUIDs must not leak into the body)", body, scenario.body)
					}
				}
				_, _ = response.Write([]byte(`[]`))
			})

			if err := executeProjectTransport(t, apiClient, "workpermit", scenario.args); err != nil {
				t.Fatalf("%s failed: %v", scenario.name, err)
			}
			if calls.Load() != 1 {
				t.Fatalf("HTTP calls = %d, want one", calls.Load())
			}
		})
	}
}

func TestWorkPermitLinkCommandsRefuseBadInputBeforeSending(t *testing.T) {
	for _, args := range [][]string{
		{"links"},
		{"links", "--permit", "not-a-guid"},
		{"links", "--permit", "00000000-0000-0000-0000-000000000000"},
		{"link-add", "--permit", workPermitLinkPermitGUID, "--target", workPermitLinkTargetGUID},
		{"link-add", "--permit", workPermitLinkPermitGUID, "--type", "Vendor", "--target", workPermitLinkTargetGUID},
		{"link-add", "--permit", workPermitLinkPermitGUID, "--type", "workorder", "--target", workPermitLinkTargetGUID},
		{"link-add", "--permit", workPermitLinkPermitGUID, "--type", "Asset", "--target", "../../users"},
		{"link-remove"},
		{"link-remove", "--link", "1234"},
		{"linked-to", "--type", "WorkOrder", "--target", workPermitLinkTargetGUID},
		{"linked-to", "--type", "../admin", "--target", workPermitLinkTargetGUID},
		{"linked-to", "--type", "asset"},
		{"prerequisite-add", "--permit", workPermitLinkPermitGUID},
		{"prerequisite-remove", "--permit", workPermitLinkPermitGUID, "--prerequisite", "not-a-guid"},
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
