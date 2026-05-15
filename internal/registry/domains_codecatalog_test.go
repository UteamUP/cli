package registry

import (
	"testing"
)

func TestCodecatalogDomainRegistered(t *testing.T) {
	var d *Domain
	for _, dom := range DefaultRegistry.Domains() {
		if dom.Name == "codecatalog" {
			d = dom
			break
		}
	}
	if d == nil {
		t.Fatal("expected codecatalog domain to be registered")
	}
}

func TestCodecatalogHistoryActionWired(t *testing.T) {
	var d *Domain
	for _, dom := range DefaultRegistry.Domains() {
		if dom.Name == "codecatalog" {
			d = dom
			break
		}
	}
	if d == nil {
		t.Fatal("expected codecatalog domain to be registered")
	}

	var historyAction *Action
	for i := range d.Actions {
		if d.Actions[i].Name == "history" {
			historyAction = &d.Actions[i]
			break
		}
	}
	if historyAction == nil {
		t.Fatal("expected `history` action on codecatalog domain")
	}

	if historyAction.ToolName != "UteamupCodecatalogHistory" {
		t.Errorf("history action ToolName = %q, want %q",
			historyAction.ToolName, "UteamupCodecatalogHistory")
	}

	if len(historyAction.Args) != 1 || historyAction.Args[0].Name != "code-guid" {
		t.Errorf("history action expected single positional arg 'code-guid', got %+v", historyAction.Args)
	}
	if !historyAction.Args[0].Required {
		t.Error("code-guid arg must be Required")
	}

	expectedFlags := map[string]string{
		"types":      "string",
		"actor-guid": "string",
		"from-utc":   "string",
		"to-utc":     "string",
		"q":          "string",
		"cursor":     "string",
		"page-size":  "int",
	}
	gotFlags := make(map[string]string)
	for _, f := range historyAction.Flags {
		gotFlags[f.Name] = f.Type
	}
	for name, ty := range expectedFlags {
		got, ok := gotFlags[name]
		if !ok {
			t.Errorf("history action missing expected flag %q", name)
			continue
		}
		if got != ty {
			t.Errorf("history action flag %q type = %q, want %q", name, got, ty)
		}
	}

	// page-size must default to 25 (server clamps but the CLI default matches REST).
	var pageSize *FlagDef
	for i := range historyAction.Flags {
		if historyAction.Flags[i].Name == "page-size" {
			pageSize = &historyAction.Flags[i]
			break
		}
	}
	if pageSize == nil {
		t.Fatal("page-size flag missing")
	}
	if pageSize.Default != 25 {
		t.Errorf("page-size Default = %v, want 25", pageSize.Default)
	}
}
