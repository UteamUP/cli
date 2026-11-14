package registry

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const filterPresetGuid = "55555555-5555-4555-8555-555555555555"

func writeFilterFile(t *testing.T, json string) string {
	t.Helper()
	file := filepath.Join(t.TempDir(), "filter.json")
	if err := os.WriteFile(file, []byte(json), 0600); err != nil {
		t.Fatal(err)
	}
	return file
}

func TestFilterPresetListReadsTheHubRoute(t *testing.T) {
	requests, err := runRegisteredDomainCommand(t, "filter-preset", "list", "asset-types")
	if err != nil || len(requests) != 1 {
		t.Fatalf("filter-preset list: %v, %+v", err, requests)
	}
	if requests[0].method != http.MethodGet || requests[0].path != "/api/user/filter-presets/asset-types" {
		t.Fatalf("list = %s %s", requests[0].method, requests[0].path)
	}
}

func TestFilterPresetCreateSendsNameFilterAndDefault(t *testing.T) {
	file := writeFilterFile(t, `{"sort":"dueDate","showArchived":false}`)
	requests, err := runRegisteredDomainCommand(t, "filter-preset", "create", "workorders",
		"--name", "Overdue", "--filter-file", file, "--default")
	if err != nil || len(requests) != 1 {
		t.Fatalf("filter-preset create: %v, %+v", err, requests)
	}
	request := requests[0]
	if request.method != http.MethodPost || request.path != "/api/user/filter-presets/workorders" {
		t.Fatalf("create = %s %s", request.method, request.path)
	}
	if request.body["name"] != "Overdue" || request.body["isDefault"] != true {
		t.Fatalf("create body = %+v", request.body)
	}
	filter, ok := request.body["filter"].(map[string]any)
	if !ok || filter["sort"] != "dueDate" {
		t.Fatalf("filter not sent as an object: %+v", request.body)
	}
	if _, leaked := request.body["hubKey"]; leaked {
		t.Fatal("hubKey belongs in the route, not the body")
	}
}

func TestFilterPresetUpdateSetDefaultAndRemoveUseTheGuidRoutes(t *testing.T) {
	file := writeFilterFile(t, `{"sort":"name"}`)
	cases := []struct {
		args   []string
		method string
		path   string
	}{
		{[]string{"update", "workorders", filterPresetGuid, "--name", "Renamed", "--filter-file", file},
			http.MethodPut, "/api/user/filter-presets/workorders/" + filterPresetGuid},
		{[]string{"set-default", "workorders", filterPresetGuid},
			http.MethodPut, "/api/user/filter-presets/workorders/" + filterPresetGuid + "/set-default"},
		{[]string{"remove", "workorders", filterPresetGuid},
			http.MethodDelete, "/api/user/filter-presets/workorders/" + filterPresetGuid},
	}
	for _, item := range cases {
		requests, err := runRegisteredDomainCommand(t, "filter-preset", item.args...)
		if err != nil || len(requests) != 1 {
			t.Fatalf("filter-preset %v: %v, %+v", item.args, err, requests)
		}
		if requests[0].method != item.method || requests[0].path != item.path {
			t.Fatalf("filter-preset %v = %s %s, want %s %s", item.args, requests[0].method, requests[0].path, item.method, item.path)
		}
	}
}

func TestFilterPresetRefusesIntegerOrEmptyGuidsBeforeSending(t *testing.T) {
	for _, guid := range []string{"12", "00000000-0000-0000-0000-000000000000"} {
		requests, err := runRegisteredDomainCommand(t, "filter-preset", "remove", "workorders", guid)
		if err == nil || len(requests) != 0 {
			t.Fatalf("remove %s: want a refusal before sending, got err=%v requests=%+v", guid, err, requests)
		}
	}
}

func TestFilterPresetActionsTakeNoUserOrTenantInput(t *testing.T) {
	domain := findDomain("filter-preset")
	if domain == nil {
		t.Fatal("filter-preset domain not registered")
	}
	for _, action := range domain.Actions {
		for _, flag := range action.Flags {
			name := strings.ToLower(flag.Name + flag.BodyName)
			if strings.Contains(name, "user") || strings.Contains(name, "tenant") {
				t.Fatalf("%s exposes --%s; saved filters act only for the signed-in user", action.Name, flag.Name)
			}
		}
		for _, arg := range action.Args {
			name := strings.ToLower(arg.Name)
			if strings.Contains(name, "user") || strings.Contains(name, "tenant") {
				t.Fatalf("%s takes %s; saved filters act only for the signed-in user", action.Name, arg.Name)
			}
		}
	}
}
