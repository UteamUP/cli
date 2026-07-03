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
		"list":   "UteamupKnowledgeArticleList",
		"get":    "UteamupKnowledgeArticleGet",
		"create": "UteamupKnowledgeArticleCreate",
		"update": "UteamupKnowledgeArticleUpdate",
		"delete": "UteamupKnowledgeArticleDelete",
		"search": "UteamupKnowledgeArticleSearch",
		"linked": "UteamupKnowledgeArticleGetLinked",
		"link":   "UteamupKnowledgeArticleLinkEntity",
		"unlink": "UteamupKnowledgeArticleUnlinkEntity",
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
