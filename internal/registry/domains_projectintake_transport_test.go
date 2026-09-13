package registry

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/uteamup/cli/internal/client"
)

func TestProjectScopeMutationCommandsUseExistingRESTOwnersAndPreserveReviewedBodies(t *testing.T) {
	for _, mode := range []string{"login", "apikey"} {
		for _, scenario := range []struct {
			domain, action, method, suffix, tool string
			resource                             bool
		}{
			{"project-intake", "source-preview", "POST", "/intake/source-preview", "UteamupProjectIntakeSourcePreview", false},
			{"project-intake", "create", "POST", "/intake", "UteamupProjectIntakeCreate", false},
			{"project-intake", "update", "PUT", "/intake/{resource}", "UteamupProjectIntakeUpdate", true},
			{"project-intake", "review", "POST", "/intake/{resource}/review", "UteamupProjectIntakeReview", true},
			{"project-intake", "apply-preview", "POST", "/intake/apply-preview", "UteamupProjectScopeApplyPreview", false},
			{"project-intake", "apply", "POST", "/intake/apply", "UteamupProjectScopeApply", false},
			{"project-requirement", "create", "POST", "/requirements", "UteamupProjectRequirementCreate", false},
			{"project-requirement", "update", "PUT", "/requirements/{resource}", "UteamupProjectRequirementUpdate", true},
		} {
			t.Run(mode+"/"+scenario.domain+"/"+scenario.action, func(t *testing.T) {
				action := findDomainAction(t, scenario.domain, scenario.action)
				if action.ToolName != scenario.tool || action.MCPOnly || action.HTTPMethod != scenario.method {
					t.Fatalf("wrong owning route: %+v", action)
				}
				body := `{"requestGuid":"` + projectTransportRequestGUID + `","expectedBaselineGuid":"00000000-0000-0000-0000-000000000000",` +
					`"expectedReviewFingerprint":"` + strings.Repeat("d", 64) + `","expectedUpdatedAt":"2026-09-13T10:11:12.123456Z",` +
					`"evidence":{},"draft":{"sourceReviewChecks":[{"code":"unmatched","resolution":null}],"rows":[{"guid":"` +
					projectTransportResourceGUID + `","quantity":2,"salesValue":null,"allocations":[]}]}}`
				var expected map[string]any
				if err := json.Unmarshal([]byte(body), &expected); err != nil {
					t.Fatal(err)
				}
				apiClient := projectTransportClient(t, mode, client.RetryOptions{}, func(response http.ResponseWriter, request *http.Request) {
					path := "/api/projects/" + projectTransportProjectGUID + strings.ReplaceAll(scenario.suffix, "{resource}", projectTransportResourceGUID)
					if request.Method != scenario.method || request.URL.Path != path || request.URL.RawQuery != "" {
						t.Errorf("request = %s %s, want %s %s", request.Method, request.URL, scenario.method, path)
					}
					var actual map[string]any
					if err := json.NewDecoder(request.Body).Decode(&actual); err != nil {
						t.Errorf("decode body: %v", err)
					}
					if !reflect.DeepEqual(actual, expected) {
						t.Errorf("reviewed body changed: %#v", actual)
					}
					_, _ = response.Write([]byte(`{"retained":true}`))
				})
				args := []string{scenario.action, projectTransportProjectGUID}
				if scenario.resource {
					args = append(args, projectTransportResourceGUID)
				}
				args = append(args, "--from-json", writeRegistryJSONFixture(t, body))
				if err := executeProjectTransport(t, apiClient, scenario.domain, args); err != nil {
					t.Fatalf("scope command failed: %v", err)
				}
			})
		}
	}
}

func TestProjectScopeReadCommandsKeepPaginationAndDeliverableFiltersInTheQuery(t *testing.T) {
	for _, scenario := range []struct {
		domain, action, suffix, tool string
		resource                     bool
	}{
		{"project-intake", "list", "/intake", "UteamupProjectIntakesList", false},
		{"project-intake", "history", "/intake/{resource}/history", "UteamupProjectIntakeHistory", true},
		{"project-requirement", "list", "/requirements", "UteamupProjectRequirementsList", false},
		{"project-requirement", "history", "/requirements/{resource}/history", "UteamupProjectRequirementHistory", true},
	} {
		t.Run(scenario.domain+"/"+scenario.action, func(t *testing.T) {
			action := findDomainAction(t, scenario.domain, scenario.action)
			if action.HTTPMethod != "GET" || action.ToolName != scenario.tool {
				t.Fatalf("read must name its owning GET route: %+v", action)
			}
			filterDeliverable := scenario.domain == "project-requirement" && scenario.action == "list"
			apiClient := projectTransportClient(t, "login", client.RetryOptions{}, func(response http.ResponseWriter, request *http.Request) {
				path := "/api/projects/" + projectTransportProjectGUID + strings.ReplaceAll(scenario.suffix, "{resource}", projectTransportResourceGUID)
				if request.Method != http.MethodGet || request.URL.Path != path {
					t.Errorf("read = %s %s, want GET %s", request.Method, request.URL, path)
				}
				query := request.URL.Query()
				if query.Get("page") != "4" || query.Get("pageSize") != "17" {
					t.Errorf("pagination query = %v", query)
				}
				if filterDeliverable && (query.Get("deliverableGuid") != projectTransportResourceGUID || query.Get("includeRetired") != "true") {
					t.Errorf("owning scope filters = %v", query)
				}
				_, _ = response.Write([]byte(`{"items":[],"totalItems":60,"currentPage":4,"pageSize":17,"totalPages":4}`))
			})
			args := []string{scenario.action, projectTransportProjectGUID}
			if scenario.resource {
				args = append(args, projectTransportResourceGUID)
			}
			args = append(args, "--page", "4", "--page-size", "17")
			if filterDeliverable {
				args = append(args, "--deliverable-guid", projectTransportResourceGUID, "--include-retired")
			}
			if err := executeProjectTransport(t, apiClient, scenario.domain, args); err != nil {
				t.Fatalf("paginated command failed: %v", err)
			}
		})
	}
}
