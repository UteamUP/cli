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

func TestCodecatalogMoveActionWired(t *testing.T) {
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
		if d.Actions[i].Name == "move" {
			action = &d.Actions[i]
			break
		}
	}
	if action == nil {
		t.Fatal("expected `move` action on codecatalog domain")
	}

	if action.ToolName != "UteamupCodingsystemMoveEntryByGuid" {
		t.Errorf("move ToolName = %q, want %q", action.ToolName, "UteamupCodingsystemMoveEntryByGuid")
	}
	if action.HTTPMethod != "POST" {
		t.Errorf("move HTTPMethod = %q, want POST", action.HTTPMethod)
	}
	if action.RESTPath != "entries/by-guid/{guid}/move" {
		t.Errorf("move RESTPath = %q, want %q", action.RESTPath, "entries/by-guid/{guid}/move")
	}

	// Two Guid (string) positional args: the entry to move (consumed by the {guid}
	// path placeholder) and the new parent (camelCase name → `newParentGuid` body field).
	if len(action.Args) != 2 {
		t.Fatalf("move expected 2 positional args, got %d (%+v)", len(action.Args), action.Args)
	}
	if action.Args[0].Name != "guid" {
		t.Errorf("move arg[0] = %q, want %q (matches {guid} path placeholder)", action.Args[0].Name, "guid")
	}
	if action.Args[1].Name != "newParentGuid" {
		t.Errorf("move arg[1] = %q, want %q (camelCase → body field newParentGuid)", action.Args[1].Name, "newParentGuid")
	}
	for i, arg := range action.Args {
		if arg.Type != "string" {
			t.Errorf("move arg[%d] type = %q, want string (Guids are strings)", i, arg.Type)
		}
		if !arg.Required {
			t.Errorf("move arg[%d] (%s) must be Required", i, arg.Name)
		}
	}
}

func TestCodecatalogDuplicateActionWired(t *testing.T) {
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
		if d.Actions[i].Name == "duplicate" {
			action = &d.Actions[i]
			break
		}
	}
	if action == nil {
		t.Fatal("expected `duplicate` action on codecatalog domain")
	}

	if action.ToolName != "UteamupCodeCatalogDuplicate" {
		t.Errorf("duplicate ToolName = %q, want %q", action.ToolName, "UteamupCodeCatalogDuplicate")
	}
	if action.HTTPMethod != "POST" {
		t.Errorf("duplicate HTTPMethod = %q, want POST", action.HTTPMethod)
	}
	if action.RESTPath != "entries/by-guid/{guid}/duplicate" {
		t.Errorf("duplicate RESTPath = %q, want %q", action.RESTPath, "entries/by-guid/{guid}/duplicate")
	}

	// Single Guid (string) positional arg feeding the {guid} path placeholder.
	if len(action.Args) != 1 || action.Args[0].Name != "guid" {
		t.Fatalf("duplicate expected single positional arg 'guid', got %+v", action.Args)
	}
	if action.Args[0].Type != "string" {
		t.Errorf("guid arg type = %q, want string (Guids are strings)", action.Args[0].Type)
	}
	if !action.Args[0].Required {
		t.Error("guid arg must be Required")
	}

	flags := make(map[string]FlagDef)
	for _, f := range action.Flags {
		flags[f.Name] = f
	}

	if f, ok := flags["copies"]; !ok {
		t.Error("duplicate must expose a `copies` flag")
	} else {
		if f.Type != "int" {
			t.Errorf("copies flag type = %q, want int", f.Type)
		}
		if f.BodyName != "copies" {
			t.Errorf("copies BodyName = %q, want copies", f.BodyName)
		}
		if f.Default != 1 {
			t.Errorf("copies Default = %v, want 1", f.Default)
		}
	}

	if f, ok := flags["target-parent-guid"]; !ok {
		t.Error("duplicate must expose a `target-parent-guid` flag")
	} else if f.Type != "string" || f.BodyName != "targetParentGuid" {
		t.Errorf("target-parent-guid flag = %+v, want string Guid → body targetParentGuid", f)
	}

	if f, ok := flags["include-descendant-guids"]; !ok {
		t.Error("duplicate must expose an `include-descendant-guids` flag")
	} else if f.Type != "stringSlice" || f.BodyName != "includeDescendantGuids" {
		t.Errorf("include-descendant-guids flag = %+v, want stringSlice → body includeDescendantGuids", f)
	}

	if f, ok := flags["include-tagged-assets"]; !ok {
		t.Error("duplicate must expose an `include-tagged-assets` flag")
	} else {
		if f.Type != "bool" || f.BodyName != "includeTaggedAssets" {
			t.Errorf("include-tagged-assets flag = %+v, want bool → body includeTaggedAssets", f)
		}
		if f.Default != true {
			t.Errorf("include-tagged-assets Default = %v, want true", f.Default)
		}
	}

	if f, ok := flags["new-root-code"]; !ok {
		t.Error("duplicate must expose a `new-root-code` flag")
	} else if f.Type != "string" || f.BodyName != "newRootCode" {
		t.Errorf("new-root-code flag = %+v, want string → body newRootCode", f)
	}
}
