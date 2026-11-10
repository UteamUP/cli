package registry

import "testing"

func TestWorkforcePreviewTimeActionsRetainVersionsAndGUIDs(t *testing.T) {
	for _, name := range []string{"approve-entry", "reject-entry", "request-info", "review-context"} {
		a := findDomainAction(t, "time-entry", name)
		if len(a.Args) != 1 || a.Args[0].Name != "externalGuid" || a.Args[0].Type != "uuid" {
			t.Fatalf("%s must identify an entry by public GUID", name)
		}
		if findFlag(a, "expected-updated-at") == nil || findFlag(a, "idempotency-key") == nil {
			t.Errorf("%s lost observed-version or retry identity", name)
		}
	}
}

func TestWorkforcePreviewPresenceAndAbsenceMatchRESTOwners(t *testing.T) {
	for _, item := range []struct{ domain, action, path string }{
		{"worker-presence", "declare", "me/presence"}, {"worker-presence", "list", "presence"},
		{"worker-presence", "manager-position", "assignments/{assignmentGuid}/manager-position"},
		{"workforce-absence", "approve", "by-guid/{overrideGuid}/approve"}, {"workforce-absence", "report", "report"},
		{"payroll-preparation", "create", "by-guid"}, {"payroll-preparation", "timesheet-choices", "timesheet-choices"},
	} {
		a := findDomainAction(t, item.domain, item.action)
		if a.RESTPath != item.path {
			t.Errorf("%s %s path = %q, want %q", item.domain, item.action, a.RESTPath, item.path)
		}
	}
}
