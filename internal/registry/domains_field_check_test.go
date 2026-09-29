package registry

import "testing"

func TestFieldCheckActionsUseGovernedMcpTools(t *testing.T) {
	for _, name := range []string{"settings", "settings-save", "points", "point-save", "due", "visits", "start", "submit", "review"} {
		action := findDomainAction(t, "field-check", name)
		if !action.MCPOnly || action.ToolName == "" {
			t.Errorf("%s must use a governed MCP tool: %+v", name, action)
		}
	}
	for _, name := range []string{"point-save", "start", "submit", "review"} {
		action := findDomainAction(t, "field-check", name)
		request := actionFlagByName(t, action, "model-file")
		if !request.JSONFile || !request.Required || request.BodyName != "model" {
			t.Errorf("%s must bind a reviewed JSON model: %+v", name, request)
		}
	}
	for _, name := range []string{"submit", "review"} {
		action := findDomainAction(t, "field-check", name)
		if len(action.Args) != 1 || action.Args[0].Type != "uuid" || action.Args[0].BodyName != "visitGuid" {
			t.Errorf("%s must require a public visit GUID: %+v", name, action.Args)
		}
	}
}
