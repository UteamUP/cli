package registry

import (
	"reflect"
	"testing"
)

func TestWorkorderSimilarReadsByGuidWithoutAICredits(t *testing.T) {
	t.Parallel()

	domain, action := domainAction(t, "workorder", "similar")
	if action.HTTPMethod != "GET" || action.ToolName != "UteamupWorkorderSimilarJobs" {
		t.Fatalf("similar must GET via UteamupWorkorderSimilarJobs, got %s %s", action.HTTPMethod, action.ToolName)
	}
	path, consumed := buildRESTPath(domain, action, map[string]any{"workorderGuid": "wo-guid"})
	if path != "/api/workorder/by-guid/wo-guid/similar" {
		t.Fatalf("path = %q", path)
	}
	if !reflect.DeepEqual(consumed, []string{"workorderGuid"}) {
		t.Fatalf("consumed = %v", consumed)
	}
	if len(action.Args) != 1 || action.Args[0].Type != "non-empty-uuid" || !action.Args[0].Required {
		t.Fatalf("workorderGuid must be a required non-empty UUID: %+v", action.Args)
	}
	if len(action.Flags) != 1 || action.Flags[0].Name != "take" || action.Flags[0].Type != "int" || action.Flags[0].Default != 5 {
		t.Fatalf("similar must take one optional int --take defaulting to 5: %+v", action.Flags)
	}
}
