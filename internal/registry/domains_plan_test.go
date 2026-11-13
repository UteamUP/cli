package registry

import (
	"net/url"
	"testing"
)

// The `plan` domain exposes AI-credit packages as a read-only action mirroring the
// backend UteamupAiCreditPackagesList MCP tool (GET /api/plan/ai-credit-packages).
func TestPlanDomainAiCreditPackagesAction(t *testing.T) {
	a := findDomainAction(t, "plan", "ai-credit-packages")

	// Method "" derives GET; the static path resolves to /api/plan/ai-credit-packages
	// under the plan domain base. No args, no flags.
	if a.HTTPMethod != "" || a.RESTPath != "ai-credit-packages" {
		t.Errorf("plan ai-credit-packages: want derived-GET static path ai-credit-packages, got method=%q path=%s", a.HTTPMethod, a.RESTPath)
	}
	if len(a.Args) != 0 || len(a.Flags) != 0 {
		t.Errorf("plan ai-credit-packages must take no args/flags, got args=%+v flags=%+v", a.Args, a.Flags)
	}
	if a.ToolName != "UteamupAiCreditPackagesList" {
		t.Errorf("plan ai-credit-packages ToolName = %q, want UteamupAiCreditPackagesList", a.ToolName)
	}
}

// Archived plans are hidden from plan lists by default; --include-archived asks the backend for them.
func TestPlanListIncludesArchivedPlansOnlyWhenAsked(t *testing.T) {
	requests, err := runRegisteredDomainCommand(t, "plan", "list")
	if err != nil {
		t.Fatal(err)
	}
	if len(requests) != 1 || url.Values(requests[0].query).Get("includeArchived") != "" {
		t.Fatalf("plan list must leave archived plans out by default: %+v", requests)
	}
	requests, err = runRegisteredDomainCommand(t, "plan", "list", "--include-archived")
	if err != nil {
		t.Fatal(err)
	}
	if len(requests) != 1 || requests[0].path != "/api/plan" || url.Values(requests[0].query).Get("includeArchived") != "true" {
		t.Fatalf("plan list --include-archived must ask for archived plans: %+v", requests)
	}
}
