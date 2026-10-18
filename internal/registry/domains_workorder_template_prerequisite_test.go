package registry

import "testing"

func TestWorkorderTemplatePrerequisiteRoutes(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"list", "set"} {
		domain, action := domainAction(t, "workorder-template-prerequisite", name)
		path, _ := buildRESTPath(domain, action, map[string]any{"templateGuid": "tpl-guid"})
		if path != "/api/workordertemplate/by-guid/tpl-guid/prerequisites" {
			t.Fatalf("%s path = %q", name, path)
		}
	}
	_, set := domainAction(t, "workorder-template-prerequisite", "set")
	if set.HTTPMethod != "PUT" || set.ToolName != "UteamupWorkorderTemplatePrerequisiteSet" {
		t.Fatalf("unexpected set action: %+v", set)
	}
	if len(set.Flags) != 1 || set.Flags[0].BodyName != "prerequisiteTemplateGuids" || set.Flags[0].Type != "stringSlice" {
		t.Fatalf("unexpected set flags: %+v", set.Flags)
	}
}

func TestWorkorderTemplateCreateWorkorderAcceptsCreatePrerequisites(t *testing.T) {
	t.Parallel()

	_, create := domainAction(t, "workorder-template", "create-workorder")
	for _, flag := range create.Flags {
		if flag.Name == "create-prerequisites" {
			if flag.BodyName != "createPrerequisites" || flag.Type != "bool" {
				t.Fatalf("unexpected flag: %+v", flag)
			}
			return
		}
	}
	t.Fatal("create-workorder has no create-prerequisites flag")
}
