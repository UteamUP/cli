package registry

import "testing"

func TestWorkforceTimeDomainRoutesThroughMcpTools(t *testing.T) {
	t.Parallel()
	domain := findDomain("workforce-time")
	if domain == nil {
		t.Fatal("workforce-time domain is not registered")
	}
	want := map[string]string{
		"training-list":               "UteamupTrainingEventList",
		"training-mine":               "UteamupTrainingEventMine",
		"training-get":                "UteamupTrainingEventGet",
		"training-create":             "UteamupTrainingEventCreate",
		"training-update":             "UteamupTrainingEventUpdate",
		"training-request-enrollment": "UteamupTrainingEventRequestEnrollment",
		"training-enrollments":        "UteamupTrainingEventEnrollmentList",
		"training-completions":        "UteamupTrainingEventCompletionList",
		"ruleset-list":                "UteamupWorkingTimeRuleSetList",
		"ruleset-get":                 "UteamupWorkingTimeRuleSetGet",
		"ruleset-create":              "UteamupWorkingTimeRuleSetCreate",
		"ruleset-update":              "UteamupWorkingTimeRuleSetUpdate",
		"policy-assignments":          "UteamupWorkingTimePolicyAssignmentList",
		"expense-list":                "UteamupTravelExpenseList",
		"expense-get":                 "UteamupTravelExpenseGet",
		"shift-schedule":              "UteamupTechnicianAvailabilitySchedule",
		"current-shift":               "UteamupTechnicianAvailabilityCurrent",
	}
	if len(domain.Actions) != len(want) {
		t.Fatalf("actions = %d, want %d", len(domain.Actions), len(want))
	}
	for _, action := range domain.Actions {
		if want[action.Name] != action.ToolName || !action.MCPOnly {
			t.Fatalf("action contract = %+v", action)
		}
		for _, arg := range action.Args {
			if !arg.Required || arg.Type != "non-empty-uuid" {
				t.Fatalf("%s: positional identifiers must be required GUIDs: %+v", action.Name, arg)
			}
		}
	}
}

func TestWorkforceTimeDomainTakesNoTenantOrUserIds(t *testing.T) {
	t.Parallel()
	domain := findDomain("workforce-time")
	if domain == nil {
		t.Fatal("workforce-time domain is not registered")
	}
	for _, action := range domain.Actions {
		for _, flag := range action.Flags {
			switch flag.Name {
			case "tenant-id", "tenant-guid", "user-id", "id":
				t.Fatalf("%s must not accept %s; the tenant and caller come from the session", action.Name, flag.Name)
			}
		}
	}
}
