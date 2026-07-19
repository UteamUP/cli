package registry

import (
	"strings"
	"testing"
)

func TestRegisteredReportRegistryIsAllowlistedAndScopeSafe(t *testing.T) {
	t.Parallel()
	domain := findDomain("registered-report")
	if domain == nil {
		t.Fatal("registered-report domain is not registered")
	}
	if domain.APIPath != "/api/reports/registered" {
		t.Fatalf("APIPath = %q", domain.APIPath)
	}

	var run *Action
	for index := range domain.Actions {
		action := &domain.Actions[index]
		if action.Name == "run" {
			run = action
		}
		if strings.Contains(strings.ToLower(action.Name), "sql") {
			t.Fatalf("registered reports must not expose arbitrary SQL: %+v", action)
		}
	}
	if run == nil || run.ToolName != "UteamupReportGenerate" ||
		run.HTTPMethod != "POST" || run.RESTPath != "query" {
		t.Fatalf("registered report run is miswired: %+v", run)
	}

	expectedFlags := map[string]string{
		"report-key": "reportKey",
		"start-date": "startDate",
		"end-date":   "endDate",
	}
	for name, bodyName := range expectedFlags {
		found := false
		for _, flag := range run.Flags {
			if flag.Name != name {
				continue
			}
			found = true
			if flag.BodyName != bodyName ||
				strings.Contains(strings.ToLower(flag.Name), "tenant") ||
				strings.Contains(strings.ToLower(flag.Name), "sql") {
				t.Fatalf("unsafe registered report flag: %+v", flag)
			}
		}
		if !found {
			t.Fatalf("missing registered report flag %q", name)
		}
	}
}
