package registry

import "testing"

func TestWorkorderListPreviewPostsGuidOnlyResources(t *testing.T) {
	d := findDomain("workorder")
	if d == nil {
		t.Fatal("missing workorder domain")
	}
	var action *Action
	for i := range d.Actions {
		if d.Actions[i].Name == "list-preview" {
			action = &d.Actions[i]
		}
	}
	if action == nil {
		t.Fatal("missing list-preview action")
	}
	if action.HTTPMethod != "POST" || action.RESTPath != "list-preview" || action.ToolName != "UteamupWorkorderListPreview" {
		t.Fatalf("list-preview must POST to workorder/list-preview via UteamupWorkorderListPreview, got %s %s %s",
			action.HTTPMethod, action.RESTPath, action.ToolName)
	}
	if len(action.Args) != 0 {
		t.Fatal("list-preview takes no positional arguments")
	}
	want := map[string]string{
		"asset-guids": "assetGuids", "asset-group-guids": "assetGroupGuids", "part-guids": "partGuids",
		"tool-guids": "toolGuids", "chemical-guids": "chemicalGuids", "maintenance-type": "maintenanceType",
	}
	for _, flag := range action.Flags {
		body, ok := want[flag.Name]
		if !ok {
			t.Errorf("unexpected flag %s", flag.Name)
			continue
		}
		if flag.BodyName != body || flag.Required {
			t.Errorf("flag %s must be optional and bound to %s", flag.Name, body)
		}
		if flag.Name != "maintenance-type" && flag.Type != "stringSlice" {
			t.Errorf("flag %s must accept several GUIDs", flag.Name)
		}
		delete(want, flag.Name)
	}
	for missing := range want {
		t.Errorf("missing flag %s", missing)
	}
}
