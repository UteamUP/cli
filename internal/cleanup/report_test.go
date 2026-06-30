package cleanup

import (
	"testing"
	"time"
)

func TestBuildSummaries_DiffsUsedAndUnused(t *testing.T) {
	cat := Catalog{Entries: []CatalogEntry{
		{Type: TypeBackendEndpoint, Key: "Asset.GetSummary", File: "a.cs", Instrumented: true},
		{Type: TypeBackendEndpoint, Key: "Asset.Unused", File: "a.cs", Instrumented: true},
		{Type: TypeFrontendComponent, Key: "Tracked", File: "t.vue", Instrumented: true},
		{Type: TypeFrontendComponent, Key: "NotInstrumented", File: "n.vue", Instrumented: false},
	}}
	usage := []UsageRow{
		{Type: TypeBackendEndpoint, Key: "Asset.GetSummary", HitCount: 5},
		{Type: TypeBackendEndpoint, Key: "Asset.Unused", HitCount: 0},
		{Type: TypeBackendEndpoint, Key: "Asset.GhostRoute", HitCount: 9}, // runtime-only → reverse diff
	}

	summaries := BuildSummaries(ReportInput{
		Catalog: cat, Usage: usage, Env: "test", Now: time.Now(), MinDays: 14,
	})

	var be, fc *TypeSummary
	for i := range summaries {
		switch summaries[i].Type {
		case TypeBackendEndpoint:
			be = &summaries[i]
		case TypeFrontendComponent:
			fc = &summaries[i]
		}
	}

	if be == nil || fc == nil {
		t.Fatal("expected backend-endpoint and frontend-component summaries")
	}
	if be.Used != 1 || be.Unused != 1 {
		t.Errorf("backend endpoint used/unused = %d/%d, want 1/1", be.Used, be.Unused)
	}
	if len(be.UnusedKeys) != 1 || be.UnusedKeys[0].Key != "Asset.Unused" {
		t.Errorf("unused key wrong: %v", be.UnusedKeys)
	}
	if len(be.ReverseDiff) != 1 || be.ReverseDiff[0] != "Asset.GhostRoute" {
		t.Errorf("reverse diff wrong: %v", be.ReverseDiff)
	}

	// Opt-in component: only the instrumented one counts; the other is "not tracked".
	if fc.Eligible != 1 || fc.NotTracked != 1 {
		t.Errorf("component eligible/notTracked = %d/%d, want 1/1", fc.Eligible, fc.NotTracked)
	}
	if fc.Unused != 1 {
		t.Errorf("instrumented-but-unused component count = %d, want 1", fc.Unused)
	}
}
