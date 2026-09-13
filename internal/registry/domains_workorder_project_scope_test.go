package registry

import "testing"

func TestWorkorderResponsibleScopeUsesGuidOnlyFlags(t *testing.T) {
	d := findDomain("workorder")
	if d == nil {
		t.Fatal("missing workorder domain")
	}
	for _, actionName := range []string{"create", "update"} {
		var action *Action
		for i := range d.Actions {
			if d.Actions[i].Name == actionName {
				action = &d.Actions[i]
			}
		}
		if action == nil {
			t.Fatalf("missing %s", actionName)
		}
		for flag, body := range map[string]string{"project-guid": "projectGuid", "project-stage-guid": "projectStageGuid", "project-output-item-guid": "projectOutputItemGuid"} {
			found := false
			for _, candidate := range action.Flags {
				if candidate.Name == flag {
					found = candidate.Type == "uuid" && candidate.BodyName == body && !candidate.Required
				}
			}
			if !found {
				t.Errorf("%s missing optional GUID flag %s", actionName, flag)
			}
		}
		if actionName == "update" && action.HTTPMethod != "PUT" {
			t.Fatal("update must use the owning PUT route")
		}
	}
}
