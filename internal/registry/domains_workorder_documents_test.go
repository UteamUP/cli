package registry

import "testing"

func TestWorkorderDocumentActionsUseGuidMcpTools(t *testing.T) {
	t.Parallel()
	domain := findDomain("workorder")
	if domain == nil {
		t.Fatal("workorder domain is not registered")
	}
	want := map[string]string{
		"add-documents":   "UteamupWorkorderAddDocuments",
		"remove-document": "UteamupWorkorderRemoveDocument",
	}
	found := 0
	for _, action := range domain.Actions {
		tool, ok := want[action.Name]
		if !ok {
			continue
		}
		found++
		if action.ToolName != tool || !action.MCPOnly {
			t.Fatalf("%s contract = %+v", action.Name, action)
		}
		for _, arg := range action.Args {
			if arg.Type != "uuid" || !arg.Required {
				t.Fatalf("%s arg %s must be a required uuid", action.Name, arg.Name)
			}
		}
	}
	if found != len(want) {
		t.Fatalf("found %d document actions, want %d", found, len(want))
	}
}
