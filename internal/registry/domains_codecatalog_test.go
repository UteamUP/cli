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

func TestCodecatalogSetResponsibleOwnersActionWired(t *testing.T) {
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

	var action *Action
	for i := range d.Actions {
		if d.Actions[i].Name == "set-responsible-owners" {
			action = &d.Actions[i]
			break
		}
	}
	if action == nil {
		t.Fatal("expected `set-responsible-owners` action on codecatalog domain")
	}

	if action.ToolName != "UteamupCodecatalogSetResponsibleOwners" {
		t.Errorf("set-responsible-owners ToolName = %q, want %q", action.ToolName, "UteamupCodecatalogSetResponsibleOwners")
	}
	if action.HTTPMethod != "PUT" {
		t.Errorf("set-responsible-owners HTTPMethod = %q, want PUT", action.HTTPMethod)
	}
	if action.RESTPath != "entries/by-guid/{codeCatalogEntryGuid}/responsible-owners" {
		t.Errorf("set-responsible-owners RESTPath = %q, want %q", action.RESTPath, "entries/by-guid/{codeCatalogEntryGuid}/responsible-owners")
	}

	// Code catalog identity is a Guid (string) positional arg matching the path placeholder.
	if len(action.Args) != 1 || action.Args[0].Name != "codeCatalogEntryGuid" {
		t.Fatalf("set-responsible-owners expected single positional arg 'codeCatalogEntryGuid', got %+v", action.Args)
	}
	if action.Args[0].Type != "string" {
		t.Errorf("codeCatalogEntryGuid arg type = %q, want string (Guids are strings)", action.Args[0].Type)
	}
	if !action.Args[0].Required {
		t.Error("codeCatalogEntryGuid arg must be Required")
	}

	// Owner ids are user strings, passed via a repeatable/comma-separated stringSlice
	// flag that serializes to the backend `userIds` body field.
	var userIDs *FlagDef
	for i := range action.Flags {
		if action.Flags[i].Name == "user-ids" {
			userIDs = &action.Flags[i]
			break
		}
	}
	if userIDs == nil {
		t.Fatal("set-responsible-owners must expose a `user-ids` flag")
	}
	if userIDs.Type != "stringSlice" {
		t.Errorf("user-ids flag type = %q, want stringSlice", userIDs.Type)
	}
	if userIDs.BodyName != "userIds" {
		t.Errorf("user-ids BodyName = %q, want %q (matches MCP tool arg)", userIDs.BodyName, "userIds")
	}
}
