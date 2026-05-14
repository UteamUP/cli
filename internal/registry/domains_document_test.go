package registry

import "testing"

// TestDocumentDomainHasMetadataAndTimelineActions guards the two CLI actions
// added by the document-metadata-and-timeline change. Failure means the
// registry stopped exposing one of the actions or its ToolName drifted from
// the backend MCP tool method name.
func TestDocumentDomainHasMetadataAndTimelineActions(t *testing.T) {
	var doc *Domain
	for _, d := range DefaultRegistry.Domains() {
		if d.Name == "document" {
			doc = d
			break
		}
	}
	if doc == nil {
		t.Fatal("expected document domain to be registered")
	}

	expected := map[string]string{
		"get-metadata": "UteamupDocumentGetMetadata",
		"get-timeline": "UteamupDocumentGetTimeline",
	}

	found := map[string]bool{}
	for _, a := range doc.Actions {
		if want, ok := expected[a.Name]; ok {
			if a.ToolName != want {
				t.Errorf("action %q ToolName = %q, want %q", a.Name, a.ToolName, want)
			}
			found[a.Name] = true
		}
	}

	for name := range expected {
		if !found[name] {
			t.Errorf("missing action %q on document domain", name)
		}
	}
}

// TestDocumentGetTimelineFlags guards the timeline action's flag surface so
// the CLI keeps the from/to/types/q/limit contract aligned with the backend.
func TestDocumentGetTimelineFlags(t *testing.T) {
	var doc *Domain
	for _, d := range DefaultRegistry.Domains() {
		if d.Name == "document" {
			doc = d
			break
		}
	}
	if doc == nil {
		t.Fatal("expected document domain to be registered")
	}

	var timeline *Action
	for i := range doc.Actions {
		if doc.Actions[i].Name == "get-timeline" {
			timeline = &doc.Actions[i]
			break
		}
	}
	if timeline == nil {
		t.Fatal("expected get-timeline action on document domain")
	}

	expectedFlags := map[string]bool{"from": false, "to": false, "types": false, "q": false, "limit": false}
	for _, f := range timeline.Flags {
		if _, ok := expectedFlags[f.Name]; ok {
			expectedFlags[f.Name] = true
		}
	}
	for name, seen := range expectedFlags {
		if !seen {
			t.Errorf("get-timeline missing flag %q", name)
		}
	}
}
