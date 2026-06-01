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
// restore-version must expose a single `externalGuid` positional arg (string),
// NOT the legacy int `id`/`documentId`. The corresponding backend routes are
// the new /api/document/{externalGuid}/... siblings of the [Obsolete] int ones.
func TestDocumentLifecycleActionsAreGuidKeyed(t *testing.T) {
	for _, name := range []string{"update", "delete", "list-versions", "upload-version"} {
		action := findDocumentAction(t, name)
		if len(action.Args) != 1 {
			t.Fatalf("%s expected exactly 1 positional arg, got %+v", name, action.Args)
		}
		arg := action.Args[0]
		if arg.Name != "externalGuid" {
			t.Errorf("%s positional arg = %q, want externalGuid", name, arg.Name)
		}
		if arg.Type != "string" {
			t.Errorf("%s arg type = %q, want string (GUIDs are strings)", name, arg.Type)
		}
		if !arg.Required {
			t.Errorf("%s externalGuid arg must be Required", name)
		}
	}
}

// TestDocumentRestoreVersionGuidKeyed guards restore-version specifically: it
// takes the document GUID plus the version ordinal, routes POST to
// /api/document/{externalGuid}/versions/{versionNumber}/restore.
func TestDocumentRestoreVersionGuidKeyed(t *testing.T) {
	action := findDocumentAction(t, "restore-version")
	if action.HTTPMethod != "POST" {
		t.Errorf("restore-version HTTPMethod = %q, want POST", action.HTTPMethod)
	}
	if action.RESTPath != "{externalGuid}/versions/{versionNumber}/restore" {
		t.Errorf("restore-version RESTPath = %q, want %q", action.RESTPath, "{externalGuid}/versions/{versionNumber}/restore")
	}
	if len(action.Args) != 2 {
		t.Fatalf("restore-version expected 2 positional args, got %+v", action.Args)
	}
	if action.Args[0].Name != "externalGuid" || action.Args[0].Type != "string" {
		t.Errorf("restore-version arg[0] = %+v, want externalGuid/string", action.Args[0])
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
		{"list-versions", map[string]any{"externalGuid": guid}, "/api/document/" + guid + "/versions", nil},
		{"upload-version", map[string]any{"externalGuid": guid, "notes": "x"}, "/api/document/" + guid + "/versions", []string{"notes"}},
		{"restore-version", map[string]any{"externalGuid": guid, "versionNumber": 3}, "/api/document/" + guid + "/versions/3/restore", nil},
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
		if _, leaked := args["externalGuid"]; leaked {
			t.Errorf("%s leaked externalGuid into body", c.action)
		}
	}
}

// TestDocumentUpdateDeleteRoutesResolve proves the GUID update/delete verbs
// route to /api/document/{externalGuid} (no /status, no int collision).
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
		args := map[string]any{"externalGuid": guid}
		path, consumed := buildRESTPath(doc, *action, args)
		want := "/api/document/" + guid
		if path != want {
			t.Errorf("%s path = %q, want %q", name, path, want)
		}
		if len(consumed) != 1 || consumed[0] != "externalGuid" {
			t.Errorf("%s consumed = %v, want [externalGuid]", name, consumed)
		}
	}
}

// TestDocumentReadActionsStayIntKeyed guards that list/get/create are left
// untouched by the lifecycle migration — get keeps the legacy {id:int} route
// per the document GUID-first contract ("Do NOT touch GetDocument").
func TestDocumentReadActionsStayIntKeyed(t *testing.T) {
	get := findDocumentAction(t, "get")
	if len(get.Args) != 1 || get.Args[0].Name != "id" {
		t.Errorf("get must keep legacy int `id` arg, got %+v", get.Args)
	}
	if get.Args[0].Type != "int" {
		t.Errorf("get id arg type = %q, want int", get.Args[0].Type)
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
