package registry

import (
	"reflect"
	"testing"
)

func TestWorkorderChildrenReadsOnePageByGuid(t *testing.T) {
	t.Parallel()

	domain, action := domainAction(t, "workorder", "children")
	if action.HTTPMethod != "GET" || action.ToolName != "UteamupWorkorderChildren" {
		t.Fatalf("children must GET, got %s %s", action.HTTPMethod, action.ToolName)
	}
	path, consumed := buildRESTPath(domain, action, map[string]any{"workorderGuid": "wo-guid"})
	if path != "/api/workorder/by-guid/wo-guid/children" {
		t.Fatalf("path = %q", path)
	}
	if !reflect.DeepEqual(consumed, []string{"workorderGuid"}) {
		t.Fatalf("consumed = %v", consumed)
	}
	if len(action.Args) != 1 || action.Args[0].Type != "non-empty-uuid" || !action.Args[0].Required {
		t.Fatalf("workorderGuid must be a required non-empty UUID: %+v", action.Args)
	}
	skip, take := findFlag(&action, "skip"), findFlag(&action, "take")
	if skip == nil || skip.Type != "int" || skip.Default != 0 || take == nil || take.Type != "int" || take.Default != 20 {
		t.Fatalf("children must page with int --skip (0) and --take (20): %+v", action.Flags)
	}
}

func TestWorkorderCreateAndUpdateCarryParentAndSitePin(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"create", "update"} {
		_, action := domainAction(t, "workorder", name)
		want := map[string][2]string{
			"parent-workorder-guid": {"parentWorkorderGuid", "uuid"},
			"site-latitude":         {"siteLatitude", "float"},
			"site-longitude":        {"siteLongitude", "float"},
		}
		for flagName, expected := range want {
			flag := findFlag(&action, flagName)
			if flag == nil {
				t.Fatalf("%s must expose --%s", name, flagName)
			}
			if flag.BodyName != expected[0] || flag.Type != expected[1] || flag.Required || flag.Default != nil {
				t.Errorf("%s --%s = %+v, want optional %s sent as %s with no default", name, flagName, *flag, expected[1], expected[0])
			}
		}
	}
}
