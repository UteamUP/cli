package registry

import "testing"

func TestMaintenanceAnalyticsUsesOwningReadRoutesWithoutIdentitySelectors(t *testing.T) {
	domain := findDomain("maintenance-analytics")
	for name, route := range map[string]string{
		"published-window-compliance": "pm-schedule-compliance",
		"backlog-weeks":               "maintenance-backlog-weeks",
		"emergency-percentage":        "emergency-work-percentage",
		"mttf":                        "mean-time-to-first-failure",
		"stockout-time-rate":          "stockout-time-rate",
	} {
		action := findAction(domain, name)
		if action == nil || action.HTTPMethod != "GET" || action.MCPOnly || len(action.Args) != 0 {
			t.Fatalf("invalid read contract for %s: %+v", name, action)
		}
		path, consumed := buildRESTPath(domain, *action, map[string]any{})
		if path != "/api/analytics/"+route || len(consumed) != 0 {
			t.Fatalf("invalid route %s: %s, %v", name, path, consumed)
		}
		for _, flag := range action.Flags {
			if flag.Name != "start-utc" && flag.Name != "end-utc" && flag.Name != "timeframe" {
				t.Fatalf("unexpected identity or metric selector: %+v", flag)
			}
		}
	}
}

func TestMaintenanceWindowFiltersStayOnReadRequests(t *testing.T) {
	for _, action := range []string{"emergency-percentage", "stockout-time-rate"} {
		requests, err := runRegisteredDomainCommand(t, "maintenance-analytics", action,
			"--start-utc", "2026-11-01T00:00:00Z", "--end-utc", "2026-11-04T00:00:00Z")
		if err != nil || len(requests) != 1 {
			t.Fatalf("unexpected read: %v, %+v", err, requests)
		}
		if firstQueryValue(requests[0].query["startUtc"]) != "2026-11-01T00:00:00Z" ||
			firstQueryValue(requests[0].query["endUtc"]) != "2026-11-04T00:00:00Z" || len(requests[0].body) != 0 {
			t.Fatalf("window filters changed: %+v", requests[0])
		}
	}
}

func TestWorkorderEmergencyFalseIsExplicitAndOmissionStaysOmitted(t *testing.T) {
	for _, actionName := range []string{"create", "update"} {
		action := findAction(findDomain("workorder"), actionName)
		found := false
		for _, flag := range action.Flags {
			if flag.Name == "emergency" {
				found = flag.Type == "bool" && flag.BodyName == "isEmergency" && flag.Default == nil
			}
		}
		if !found {
			t.Fatalf("%s has no explicit nullable emergency field", actionName)
		}
	}
	for _, flags := range [][]string{nil, {"--emergency=false"}, {"--emergency=true"}} {
		args := append([]string{"update", "8d2e914b-8445-47b6-b1be-041a6b29c840"}, flags...)
		requests, err := runRegisteredDomainCommand(t, "workorder", args...)
		if err != nil || len(requests) != 1 {
			t.Fatalf("update failed: %v, %+v", err, requests)
		}
		value, present := requests[0].body["isEmergency"]
		if present != (len(flags) > 0) || present && value != (flags[0] == "--emergency=true") {
			t.Fatalf("classification presence changed: %+v", requests[0].body)
		}
	}
}

func firstQueryValue(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}
