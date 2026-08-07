package registry

import "testing"

// findDocumentDomain returns the registered `document` domain or fails the test.
func findDocumentDomain(t *testing.T) *Domain {
	t.Helper()
	for _, d := range DefaultRegistry.Domains() {
		if d.Name == "document" {
			return d
		}
	}
	t.Fatal("expected document domain to be registered")
	return nil
}

// findDocumentAction returns the named action on the document domain or fails.
func findDocumentAction(t *testing.T, name string) *Action {
	t.Helper()
	doc := findDocumentDomain(t)
	for i := range doc.Actions {
		if doc.Actions[i].Name == name {
			return &doc.Actions[i]
		}
	}
	t.Fatalf("expected %q action on document domain", name)
	return nil
}

// TestDocumentLifecycleActionsAreGuidKeyed guards the GUID-first migration of
// the document lifecycle verbs. update/delete/list-versions/upload-version/
// Each lifecycle action exposes the canonical `documentGuid` positional arg,
// never a legacy int `id`/`documentId`.
func TestDocumentLifecycleActionsAreGuidKeyed(t *testing.T) {
	for _, name := range []string{"get", "update", "delete", "archive", "unarchive", "list-versions", "upload-version"} {
		action := findDocumentAction(t, name)
		if len(action.Args) != 1 {
			t.Fatalf("%s expected exactly 1 positional arg, got %+v", name, action.Args)
		}
		arg := action.Args[0]
		if arg.Name != "documentGuid" {
			t.Errorf("%s positional arg = %q, want documentGuid", name, arg.Name)
		}
		if arg.Type != "string" {
			t.Errorf("%s arg type = %q, want string (GUIDs are strings)", name, arg.Type)
		}
		if !arg.Required {
			t.Errorf("%s documentGuid arg must be Required", name)
		}
	}
}

// TestDocumentRestoreVersionGuidKeyed guards restore-version specifically: it
// takes the document GUID plus the version ordinal, routes POST to
// /api/document/{documentGuid}/versions/{versionNumber}/restore.
func TestDocumentRestoreVersionGuidKeyed(t *testing.T) {
	action := findDocumentAction(t, "restore-version")
	if action.HTTPMethod != "POST" {
		t.Errorf("restore-version HTTPMethod = %q, want POST", action.HTTPMethod)
	}
	if action.RESTPath != "{documentGuid}/versions/{versionNumber}/restore" {
		t.Errorf("restore-version RESTPath = %q, want %q", action.RESTPath, "{documentGuid}/versions/{versionNumber}/restore")
	}
	if len(action.Args) != 2 {
		t.Fatalf("restore-version expected 2 positional args, got %+v", action.Args)
	}
	if action.Args[0].Name != "documentGuid" || action.Args[0].Type != "string" {
		t.Errorf("restore-version arg[0] = %+v, want documentGuid/string", action.Args[0])
	}
	if action.Args[1].Name != "versionNumber" || action.Args[1].Type != "int" {
		t.Errorf("restore-version arg[1] = %+v, want versionNumber/int", action.Args[1])
	}
}

// TestDocumentVersionRoutesResolve proves buildRESTPath produces the contract
// GUID URLs for the version actions and strips the path-consumed args from the
// JSON body (so the GUID never double-leaks into the payload).
func TestDocumentVersionRoutesResolve(t *testing.T) {
	doc := findDocumentDomain(t)
	guid := "11111111-2222-3333-4444-555555555555"

	cases := []struct {
		action   string
		args     map[string]any
		wantPath string
		wantBody []string // arg names that must remain in body (not consumed)
	}{
		{"list-versions", map[string]any{"documentGuid": guid}, "/api/document/" + guid + "/versions", nil},
		{"upload-version", map[string]any{"documentGuid": guid}, "/api/document/" + guid + "/versions", nil},
		{"restore-version", map[string]any{"documentGuid": guid, "versionNumber": 3}, "/api/document/" + guid + "/versions/3/restore", nil},
	}

	for _, c := range cases {
		var action *Action
		for i := range doc.Actions {
			if doc.Actions[i].Name == c.action {
				action = &doc.Actions[i]
				break
			}
		}
		if action == nil {
			t.Fatalf("missing action %q", c.action)
		}
		// Clone args so buildRESTPath's consumed-list semantics don't mutate the case data.
		args := make(map[string]any, len(c.args))
		for k, v := range c.args {
			args[k] = v
		}
		path, consumed := buildRESTPath(doc, *action, args)
		if path != c.wantPath {
			t.Errorf("%s path = %q, want %q", c.action, path, c.wantPath)
		}
		for _, name := range consumed {
			delete(args, name)
		}
		for _, want := range c.wantBody {
			if _, ok := args[want]; !ok {
				t.Errorf("%s expected %q to remain in body, was consumed", c.action, want)
			}
		}
		if _, leaked := args["documentGuid"]; leaked {
			t.Errorf("%s leaked documentGuid into body", c.action)
		}
	}
}

// TestDocumentUpdateDeleteRoutesResolve proves the GUID update/delete verbs
// route to /api/document/{documentGuid} (no /status, no int collision).
func TestDocumentUpdateDeleteRoutesResolve(t *testing.T) {
	doc := findDocumentDomain(t)
	guid := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	for _, name := range []string{"update", "delete"} {
		var action *Action
		for i := range doc.Actions {
			if doc.Actions[i].Name == name {
				action = &doc.Actions[i]
				break
			}
		}
		if action == nil {
			t.Fatalf("missing action %q", name)
		}
		args := map[string]any{"documentGuid": guid}
		path, consumed := buildRESTPath(doc, *action, args)
		want := "/api/document/" + guid
		if path != want {
			t.Errorf("%s path = %q, want %q", name, path, want)
		}
		if len(consumed) != 1 || consumed[0] != "documentGuid" {
			t.Errorf("%s consumed = %v, want [documentGuid]", name, consumed)
		}
	}
}

// TestDocumentGetUsesCanonicalGuidRoute guards the GUID-only read contract.
func TestDocumentGetUsesCanonicalGuidRoute(t *testing.T) {
	get := findDocumentAction(t, "get")
	if len(get.Args) != 1 || get.Args[0].Name != "documentGuid" {
		t.Errorf("get must use documentGuid, got %+v", get.Args)
	}
	if get.Args[0].Type != "string" || get.RESTPath != "by-guid/{documentGuid}" {
		t.Errorf("get contract = %+v, want string GUID at by-guid/{documentGuid}", get)
	}
}

func TestDocumentVersionUploadUsesMultipartFile(t *testing.T) {
	upload := findDocumentAction(t, "upload-version")
	if upload.ToolName != "UteamupDocumentUploadVersionByGuid" {
		t.Errorf("upload-version ToolName = %q", upload.ToolName)
	}
	if len(upload.Flags) != 1 {
		t.Fatalf("upload-version flags = %+v, want one file flag", upload.Flags)
	}
	file := upload.Flags[0]
	if file.Name != "file" || file.BodyName != "file" || !file.Required || !file.UploadFile {
		t.Errorf("upload-version file flag = %+v, want required multipart field named file", file)
	}
}

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
