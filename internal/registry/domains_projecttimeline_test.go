package registry

import "testing"

func TestProjectTimelineUsesGuardedDatesOnlyRoutes(t *testing.T) {
	for _, kind := range []string{"workorders", "stages"} {
		a := findDomainAction(t, "project-timeline-"+kind, "update")
		if a.HTTPMethod != "PUT" || len(a.Args) != 2 || a.Args[0].Type != "non-empty-uuid" || a.Args[1].Type != "non-empty-uuid" {
			t.Fatalf("invalid guarded GUID schedule action: %+v", a)
		}
		for _, name := range []string{"due-date", "expected-updated-at"} {
			if flag := findFlag(a, name); flag == nil || !flag.Required {
				t.Fatalf("missing required %s", name)
			}
		}
		if len(a.Flags) != 3 {
			t.Fatal("timeline must not accept status or gate fields")
		}
		if findFlag(a, "start-date").Required != (kind == "workorders") {
			t.Fatal("only work orders require a start")
		}
	}
	if findFlag(findDomainAction(t, "automation", "list"), "project-guid") == nil {
		t.Fatal("missing workflow project filter")
	}
}
