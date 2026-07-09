package registry

import (
	"testing"
)

func findAssetDomain(t *testing.T) *Domain {
	t.Helper()
	for _, dom := range DefaultRegistry.Domains() {
		if dom.Name == "asset" {
			return dom
		}
	}
	t.Fatal("expected asset domain to be registered")
	return nil
}

func findAssetAction(t *testing.T, name string) *Action {
	t.Helper()
	d := findAssetDomain(t)
	for i := range d.Actions {
		if d.Actions[i].Name == name {
			return &d.Actions[i]
		}
	}
	t.Fatalf("expected %q action on asset domain", name)
	return nil
}

func TestAssetDuplicateAction(t *testing.T) {
	action := findAssetAction(t, "duplicate")

	if action.ToolName != "UteamupAssetDuplicate" {
		t.Errorf("duplicate ToolName = %q, want %q", action.ToolName, "UteamupAssetDuplicate")
	}
	if action.HTTPMethod != "POST" {
		t.Errorf("duplicate HTTPMethod = %q, want POST", action.HTTPMethod)
	}
	if action.RESTPath != "by-guid/{assetGuid}/duplicate" {
		t.Errorf("duplicate RESTPath = %q, want %q", action.RESTPath, "by-guid/{assetGuid}/duplicate")
	}

	// Asset identity is a Guid (string) positional arg matching the path placeholder.
	if len(action.Args) != 1 || action.Args[0].Name != "assetGuid" {
		t.Fatalf("duplicate expected single positional arg 'assetGuid', got %+v", action.Args)
	}
	if action.Args[0].Type != "string" {
		t.Errorf("assetGuid arg type = %q, want string (Guids are strings)", action.Args[0].Type)
	}
	if !action.Args[0].Required {
		t.Error("assetGuid arg must be Required")
	}
}

func TestAssetAskAction(t *testing.T) {
	action := findAssetAction(t, "ask")
	if action.ToolName != "UteamupAssetAsk" {
		t.Errorf("ask ToolName = %q", action.ToolName)
	}
	if action.HTTPMethod != "POST" || action.RESTPath != "{assetGuid}/ask" {
		t.Fatalf("ask route = %s %s", action.HTTPMethod, action.RESTPath)
	}
	if len(action.Args) != 1 || action.Args[0].Name != "assetGuid" || action.Args[0].Type != "string" {
		t.Fatalf("ask args = %+v", action.Args)
	}
	if len(action.Flags) != 1 || action.Flags[0].Name != "question" || !action.Flags[0].Required {
		t.Fatalf("ask flags = %+v", action.Flags)
	}
}

func TestAssetSetResponsibleOwnersAction(t *testing.T) {
	action := findAssetAction(t, "set-responsible-owners")

	if action.ToolName != "UteamupAssetSetResponsibleOwners" {
		t.Errorf("set-responsible-owners ToolName = %q, want %q", action.ToolName, "UteamupAssetSetResponsibleOwners")
	}
	if action.HTTPMethod != "PUT" {
		t.Errorf("set-responsible-owners HTTPMethod = %q, want PUT", action.HTTPMethod)
	}
	if action.RESTPath != "by-guid/{assetGuid}/responsible-owners" {
		t.Errorf("set-responsible-owners RESTPath = %q, want %q", action.RESTPath, "by-guid/{assetGuid}/responsible-owners")
	}

	// Asset identity is a Guid (string) positional arg matching the path placeholder.
	if len(action.Args) != 1 || action.Args[0].Name != "assetGuid" {
		t.Fatalf("set-responsible-owners expected single positional arg 'assetGuid', got %+v", action.Args)
	}
	if action.Args[0].Type != "string" {
		t.Errorf("assetGuid arg type = %q, want string (Guids are strings)", action.Args[0].Type)
	}
	if !action.Args[0].Required {
		t.Error("assetGuid arg must be Required")
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

func TestAssetEditCodeAssignmentAction(t *testing.T) {
	action := findAssetAction(t, "edit-code-assignment")

	if action.ToolName != "UteamupAssetEditCodeAssignment" {
		t.Errorf("edit-code-assignment ToolName = %q, want %q", action.ToolName, "UteamupAssetEditCodeAssignment")
	}
	if action.HTTPMethod != "PATCH" {
		t.Errorf("edit-code-assignment HTTPMethod = %q, want PATCH", action.HTTPMethod)
	}
	if action.RESTPath != "by-guid/{assetGuid}/codeassignment" {
		t.Errorf("edit-code-assignment RESTPath = %q, want %q", action.RESTPath, "by-guid/{assetGuid}/codeassignment")
	}

	// Asset identity is a Guid (string) positional arg matching the path placeholder.
	if len(action.Args) != 1 || action.Args[0].Name != "assetGuid" {
		t.Fatalf("edit-code-assignment expected single positional arg 'assetGuid', got %+v", action.Args)
	}
	if action.Args[0].Type != "string" {
		t.Errorf("assetGuid arg type = %q, want string (Guids are strings)", action.Args[0].Type)
	}
	if !action.Args[0].Required {
		t.Error("assetGuid arg must be Required")
	}

	// The rename flags drive the new editable Name + Code surface. Each is
	// optional; BodyName defaults to camelCase(Name) → `name`, `desiredCode`,
	// matching the UteamupAssetEditCodeAssignment MCP tool args.
	flagByName := func(name string) *FlagDef {
		for i := range action.Flags {
			if action.Flags[i].Name == name {
				return &action.Flags[i]
			}
		}
		return nil
	}
	for _, want := range []struct {
		name string
		typ  string
	}{
		{"name", "string"},
		{"desired-code", "string"},
		{"code-catalog-entry-guid", "string"},
		{"parent-asset-guid", "string"},
		{"demote-to-root", "bool"},
	} {
		f := flagByName(want.name)
		if f == nil {
			t.Fatalf("edit-code-assignment must expose a %q flag", want.name)
		}
		if f.Type != want.typ {
			t.Errorf("%s flag type = %q, want %q", want.name, f.Type, want.typ)
		}
		if f.Required {
			t.Errorf("%s flag must be optional (leave-unchanged semantics)", want.name)
		}
	}
}
