package registry

import "testing"

func findDomainByName(t *testing.T, name string) *Domain {
	t.Helper()
	for _, d := range DefaultRegistry.Domains() {
		if d.Name == name {
			return d
		}
	}
	t.Fatalf("expected %q domain to be registered", name)
	return nil
}

func TestKnowledgeSpaceDomainWired(t *testing.T) {
	d := findDomainByName(t, "knowledgespace")
	if d.Description == "" {
		t.Error("knowledgespace domain must have a Description")
	}
	wantAliases := map[string]bool{"kbspace": true, "knowledge-space": true, "space": true}
	for _, a := range d.Aliases {
		delete(wantAliases, a)
	}
	if len(wantAliases) != 0 {
		t.Errorf("knowledgespace missing aliases: %v (got %v)", wantAliases, d.Aliases)
	}

	// Every action maps to the exact MCP tool method name in MCP/Tools/KnowledgeSpaceTools.cs.
	expected := map[string]string{
		"list":         "UteamupKnowledgeSpaceList",
		"get":          "UteamupKnowledgeSpaceGet",
		"create":       "UteamupKnowledgeSpaceCreate",
		"update":       "UteamupKnowledgeSpaceUpdate",
		"delete":       "UteamupKnowledgeSpaceDelete",
		"usage":        "UteamupKnowledgeSpaceUsage",
		"list-members": "UteamupKnowledgeSpaceListMembers",
		"add-member":   "UteamupKnowledgeSpaceAddMember",
	}
	got := map[string]string{}
	for _, a := range d.Actions {
		got[a.Name] = a.ToolName
	}
	if len(got) != len(expected) {
		t.Errorf("knowledgespace action count = %d, want %d (got %v)", len(got), len(expected), got)
	}
	for name, tool := range expected {
		if got[name] != tool {
			t.Errorf("action %q tool = %q, want %q", name, got[name], tool)
		}
	}

	// GUID-first: identifier args are named spaceGuid (never an int id).
	for _, a := range d.Actions {
		for _, arg := range a.Args {
			if arg.Name == "id" {
				t.Errorf("action %q uses int id arg; knowledgespace must be GUID-first (spaceGuid)", a.Name)
			}
		}
	}
}

func TestKnowledgeAiDomainWired(t *testing.T) {
	d := findDomainByName(t, "knowledgeai")
	if d.Description == "" {
		t.Error("knowledgeai domain must have a Description")
	}
	wantAliases := map[string]bool{"upmate": true, "kbai": true, "knowledge-ai": true}
	for _, a := range d.Aliases {
		delete(wantAliases, a)
	}
	if len(wantAliases) != 0 {
		t.Errorf("knowledgeai missing aliases: %v (got %v)", wantAliases, d.Aliases)
	}

	expected := map[string]string{
		"translate":          "UteamupKnowledgeAiTranslate",
		"generate-from-text": "UteamupKnowledgeAiGenerateFromText",
	}
	got := map[string]string{}
	for _, a := range d.Actions {
		got[a.Name] = a.ToolName
	}
	if len(got) != len(expected) {
		t.Errorf("knowledgeai action count = %d, want %d (got %v)", len(got), len(expected), got)
	}
	for name, tool := range expected {
		if got[name] != tool {
			t.Errorf("action %q tool = %q, want %q", name, got[name], tool)
		}
	}
}
