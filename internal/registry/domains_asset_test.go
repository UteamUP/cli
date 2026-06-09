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
