package registry

import (
	"strings"
	"testing"
)

func TestKnowledgeDomainWired(t *testing.T) {
	d := findDomainByName(t, "knowledge")
	if d.Description == "" {
		t.Error("knowledge domain must have a Description")
	}
	wantAliases := map[string]bool{"kb": true, "knowledge-article": true}
	for _, a := range d.Aliases {
		delete(wantAliases, a)
	}
	if len(wantAliases) != 0 {
		t.Errorf("knowledge missing aliases: %v (got %v)", wantAliases, d.Aliases)
	}

	// Backend controller is KnowledgeArticleController → api/knowledgearticle;
	// the auto-derived /api/knowledge matches no backend route.
	if d.APIPath != "/api/knowledgearticle" {
		t.Errorf("knowledge APIPath = %q, want %q", d.APIPath, "/api/knowledgearticle")
	}

	// Every action maps to the exact backend MCP tool method name.
	expected := map[string]string{
		"move-relative": "UteamupKnowledgeArticleMoveRelative",
		"publish":       "UteamupKnowledgeArticlePublish",
		"submit-review": "UteamupKnowledgeArticleSubmitReview",
		"list":          "UteamupKnowledgeArticleList",
		"get":           "UteamupKnowledgeArticleGet",
		"create":        "UteamupKnowledgeArticleCreate",
		"update":        "UteamupKnowledgeArticleUpdate",
		"delete":        "UteamupKnowledgeArticleDelete",
		"search":        "UteamupKnowledgeArticleSearch",
		"linked":        "UteamupKnowledgeArticleGetLinked",
		"link":          "UteamupKnowledgeArticleLinkEntity",
		"unlink":        "UteamupKnowledgeArticleUnlinkEntity",
		// Reviewed Location relations (domains_location_knowledge.go).
		"location-link":   "UteamupLocationKnowledgeLink",
		"location-unlink": "UteamupLocationKnowledgeUnlink",
	}
	got := map[string]string{}
	for _, a := range d.Actions {
		got[a.Name] = a.ToolName
	}
	if len(got) != len(expected) {
		t.Errorf("knowledge action count = %d, want %d (got %v)", len(got), len(expected), got)
	}
	for name, tool := range expected {
		if got[name] != tool {
			t.Errorf("action %q tool = %q, want %q", name, got[name], tool)
		}
	}
}

func TestKnowledgeArticleCrudGuidKeyed(t *testing.T) {
	d := findDomainByName(t, "knowledge")
	actions := map[string]Action{}
	for _, a := range d.Actions {
		actions[a.Name] = a
	}

	// get/update/delete target the GUID route by-guid/{articleGuid}; the int
	// {id} route is [Obsolete] on the backend and does not survive reseeds.
	for _, name := range []string{"get", "update", "delete"} {
		a, ok := actions[name]
		if !ok {
			t.Errorf("action %q not registered on knowledge domain", name)
			continue
		}
		if a.RESTPath != "by-guid/{articleGuid}" {
			t.Errorf("action %q RESTPath = %q, want %q", name, a.RESTPath, "by-guid/{articleGuid}")
		}
		if len(a.Args) != 1 {
			t.Errorf("action %q arg count = %d, want 1", name, len(a.Args))
			continue
		}
		arg := a.Args[0]
		if arg.Name != "articleGuid" {
			t.Errorf("action %q arg = %q, want %q", name, arg.Name, "articleGuid")
		}
		if !arg.Required {
			t.Errorf("action %q arg %q must be Required", name, arg.Name)
		}
		if arg.Type != "string" {
			t.Errorf("action %q arg %q Type = %q, want %q (GUIDs travel as validated strings)", name, arg.Name, arg.Type, "string")
		}
		// GUID-first: the article CRUD surface never takes an int id.
		for _, ar := range a.Args {
			if ar.Name == "id" || ar.Type == "int" {
				t.Errorf("action %q uses an int id arg; knowledge article CRUD is GUID-first", name)
			}
		}
	}
}

func TestKnowledgeArticleGuidRESTPathExpansion(t *testing.T) {
	d := findDomainByName(t, "knowledge")
	var get Action
	for _, a := range d.Actions {
		if a.Name == "get" {
			get = a
		}
	}
	args := map[string]any{"articleGuid": "11111111-2222-3333-4444-555555555555"}
	path, consumed := buildRESTPath(d, get, args)
	want := "/api/knowledgearticle/by-guid/11111111-2222-3333-4444-555555555555"
	if path != want {
		t.Errorf("get REST path = %q, want %q", path, want)
	}
	if len(consumed) != 1 || consumed[0] != "articleGuid" {
		t.Errorf("get path expansion consumed %v, want [articleGuid]", consumed)
	}
}

func TestKnowledgeEntityLinkActionsWired(t *testing.T) {
	d := findDomainByName(t, "knowledge")
	actions := map[string]Action{}
	for _, a := range d.Actions {
		actions[a.Name] = a
	}

	entityArgs := []string{"entityType", "entityGuid"}
	linkArgs := []string{"entityType", "entityGuid", "articleGuid"}
	cases := []struct {
		name       string
		httpMethod string
		restPath   string
		args       []string
	}{
		{"linked", "GET", "linked/{entityType}/{entityGuid}", entityArgs},
		{"link", "POST", "linked/{entityType}/{entityGuid}/{articleGuid}", linkArgs},
		{"unlink", "DELETE", "linked/{entityType}/{entityGuid}/{articleGuid}", linkArgs},
	}

	for _, tc := range cases {
		a, ok := actions[tc.name]
		if !ok {
			t.Errorf("action %q not registered on knowledge domain", tc.name)
			continue
		}
		if a.HTTPMethod != tc.httpMethod {
			t.Errorf("action %q HTTPMethod = %q, want %q", tc.name, a.HTTPMethod, tc.httpMethod)
		}
		if a.RESTPath != tc.restPath {
			t.Errorf("action %q RESTPath = %q, want %q", tc.name, a.RESTPath, tc.restPath)
		}
		if len(a.Args) != len(tc.args) {
			t.Errorf("action %q arg count = %d, want %d", tc.name, len(a.Args), len(tc.args))
			continue
		}
		for i, want := range tc.args {
			arg := a.Args[i]
			if arg.Name != want {
				t.Errorf("action %q arg[%d] = %q, want %q", tc.name, i, arg.Name, want)
			}
			if !arg.Required {
				t.Errorf("action %q arg %q must be Required", tc.name, arg.Name)
			}
			if arg.Type != "string" {
				t.Errorf("action %q arg %q Type = %q, want %q (GUIDs travel as validated strings)", tc.name, arg.Name, arg.Type, "string")
			}
		}
		// GUID-first: the entity-link surface never takes an int id.
		for _, arg := range a.Args {
			if arg.Name == "id" {
				t.Errorf("action %q uses int id arg; entity links are GUID-first", tc.name)
			}
		}
		// Help text must enumerate the valid entity types.
		for _, entityType := range []string{"asset", "part", "tool", "chemical", "location", "workorder", "workordertemplate", "industrycode"} {
			if !strings.Contains(a.Args[0].Description, entityType) {
				t.Errorf("action %q entityType help text missing %q (got %q)", tc.name, entityType, a.Args[0].Description)
			}
		}
	}
}

func TestKnowledgeEntityLinkRESTPathExpansion(t *testing.T) {
	d := findDomainByName(t, "knowledge")
	var link Action
	for _, a := range d.Actions {
		if a.Name == "link" {
			link = a
		}
	}
	args := map[string]any{
		"entityType":  "asset",
		"entityGuid":  "11111111-2222-3333-4444-555555555555",
		"articleGuid": "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
	}
	path, consumed := buildRESTPath(d, link, args)
	wantPath := "/api/knowledgearticle/linked/asset/11111111-2222-3333-4444-555555555555/aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	if path != wantPath {
		t.Errorf("link REST path = %q, want %q", path, wantPath)
	}
	if len(consumed) != 3 {
		t.Errorf("link path expansion consumed %d args, want 3 (%v)", len(consumed), consumed)
	}
}

func TestKnowledgeRelativeMoveRequiresExplicitPlacementConfirmation(t *testing.T) {
	domain := findDomainByName(t, "knowledge")
	for _, action := range domain.Actions {
		if action.Name != "move-relative" {
			continue
		}
		if action.HTTPMethod != "POST" || action.RESTPath != "by-guid/{articleGuid}/move/relative" {
			t.Fatal("relative placement must use its dedicated GUID route")
		}
		flags := map[string]FlagDef{}
		for _, flag := range action.Flags {
			flags[flag.Name] = flag
		}
		if !flags["confirm"].MustBeTrue || !flags["confirm"].LocalOnly || !flags["confirm"].Required {
			t.Fatal("relative move requires explicit local confirmation")
		}
		if !action.RejectExtraArgs || action.Args[0].Type != "non-empty-uuid" || flags["target-article-guid"].Type != "non-empty-uuid" ||
			!flags["target-article-guid"].Required || flags["target-article-guid"].BodyName != "targetArticleGuid" ||
			!flags["place-after"].Required || flags["place-after"].Type != "bool" {
			t.Fatal("relative move must carry target GUID and explicit direction")
		}
		return
	}
	t.Fatal("relative move command is missing")
}

func TestKnowledgeStatusActionsPostToGuidRoutes(t *testing.T) {
	d := findDomainByName(t, "knowledge")
	want := map[string]string{
		"publish":       "/api/knowledgearticle/by-guid/11111111-2222-3333-4444-555555555555/publish",
		"submit-review": "/api/knowledgearticle/by-guid/11111111-2222-3333-4444-555555555555/submit-review",
	}
	for _, a := range d.Actions {
		wantPath, ok := want[a.Name]
		if !ok {
			continue
		}
		delete(want, a.Name)
		if a.HTTPMethod != "POST" || !a.RejectExtraArgs {
			t.Errorf("%s must be a POST that rejects extra args, got method %q reject=%v", a.Name, a.HTTPMethod, a.RejectExtraArgs)
		}
		if len(a.Args) != 1 || a.Args[0].Name != "articleGuid" || !a.Args[0].Required || a.Args[0].Type != "non-empty-uuid" {
			t.Errorf("%s must take exactly one required non-empty articleGuid, got %+v", a.Name, a.Args)
		}
		if len(a.Flags) != 0 {
			t.Errorf("%s must not accept flags (no actor or status override), got %+v", a.Name, a.Flags)
		}
		path, consumed := buildRESTPath(d, a, map[string]any{"articleGuid": "11111111-2222-3333-4444-555555555555"})
		if path != wantPath {
			t.Errorf("%s REST path = %q, want %q", a.Name, path, wantPath)
		}
		if len(consumed) != 1 || consumed[0] != "articleGuid" {
			t.Errorf("%s path expansion consumed %v, want [articleGuid]", a.Name, consumed)
		}
	}
	if len(want) != 0 {
		t.Errorf("missing knowledge status actions: %v", want)
	}
}

// The backend answers GET /api/knowledgearticle and GET .../search with 405; listing and
// searching go through POST /api/knowledgearticle/search with pageNumber/pageSize in the body.
func TestKnowledgeListAndSearchPostToSearch(t *testing.T) {
	d := findDomainByName(t, "knowledge")
	for _, name := range []string{"list", "search"} {
		var action Action
		for _, a := range d.Actions {
			if a.Name == name {
				action = a
			}
		}
		if action.HTTPMethod != "POST" {
			t.Errorf("%s HTTPMethod = %q, want POST", name, action.HTTPMethod)
		}
		path, _ := buildRESTPath(d, action, map[string]any{"query": "pump"})
		if path != "/api/knowledgearticle/search" {
			t.Errorf("%s REST path = %q, want /api/knowledgearticle/search", name, path)
		}
		bodies := map[string]bool{}
		for _, f := range action.Flags {
			bodies[f.BodyName] = true
		}
		if !bodies["pageNumber"] || !bodies["pageSize"] {
			t.Errorf("%s flags must send pageNumber and pageSize, got %v", name, bodies)
		}
	}
}
