package registry

import (
	"testing"
)

// --- Project copilot domain ---

func findProjectCopilotAction(t *testing.T, name string) *Action {
	t.Helper()
	d := findDomain("project-copilot")
	if d == nil {
		t.Fatal("expected project-copilot domain to be registered")
	}
	for i := range d.Actions {
		if d.Actions[i].Name == name {
			return &d.Actions[i]
		}
	}
	t.Fatalf("expected `%s` action on project-copilot domain", name)
	return nil
}

func TestProjectCopilotDomainRegistered(t *testing.T) {
	d := findDomain("project-copilot")
	if d == nil {
		t.Fatal("expected project-copilot domain to be registered")
	}
	// ProjectCopilotController routes live under /api/projects (plural) —
	// NOT the /api/project base the `project` domain auto-derives. The
	// explicit APIPath is what makes every copilot action route correctly.
	if d.APIPath != "/api/projects" {
		t.Errorf("project-copilot APIPath = %q, want /api/projects", d.APIPath)
	}
	expected := map[string]bool{"projectcopilot": true, "copilot": true}
	for _, alias := range d.Aliases {
		delete(expected, alias)
	}
	if len(expected) > 0 {
		t.Errorf("missing aliases: %v", expected)
	}
}

func TestProjectCopilotHealthActionsWired(t *testing.T) {
	compute := findProjectCopilotAction(t, "health-compute")
	if compute.ToolName != "UteamupProjectComputeHealth" || compute.HTTPMethod != "POST" || compute.RESTPath != "{projectGuid}/health/compute" {
		t.Errorf("health-compute must be POST {projectGuid}/health/compute, got %+v", compute)
	}

	health := findProjectCopilotAction(t, "health")
	if health.ToolName != "UteamupProjectGetHealth" || health.HTTPMethod != "" || health.RESTPath != "{projectGuid}/health" {
		t.Errorf("health must be GET {projectGuid}/health, got %+v", health)
	}

	for _, name := range []string{"health-compute", "health"} {
		a := findProjectCopilotAction(t, name)
		if len(a.Args) != 1 || a.Args[0].Name != "projectGuid" || !a.Args[0].Required || a.Args[0].Type != "string" {
			t.Errorf("%s expected single required string positional arg 'projectGuid', got %+v", name, a.Args)
		}
	}
}

func TestProjectCopilotSummaryActionWired(t *testing.T) {
	action := findProjectCopilotAction(t, "summary")
	if action.ToolName != "UteamupProjectGenerateSummary" || action.HTTPMethod != "POST" || action.RESTPath != "{projectGuid}/summary" {
		t.Errorf("summary must be POST {projectGuid}/summary, got %+v", action)
	}
	if len(action.Args) != 1 || action.Args[0].Name != "projectGuid" || !action.Args[0].Required {
		t.Fatalf("summary expected single required positional arg 'projectGuid', got %+v", action.Args)
	}
}

func TestProjectCopilotBomSuggestActionWired(t *testing.T) {
	action := findProjectCopilotAction(t, "bom-suggest")
	if action.ToolName != "UteamupProjectSuggestBom" || action.HTTPMethod != "POST" || action.RESTPath != "{projectGuid}/bom/suggest" {
		t.Errorf("bom-suggest must be POST {projectGuid}/bom/suggest, got %+v", action)
	}
	if len(action.Args) != 1 || action.Args[0].Name != "projectGuid" || !action.Args[0].Required {
		t.Fatalf("bom-suggest expected single required positional arg 'projectGuid', got %+v", action.Args)
	}

	var description *FlagDef
	for i := range action.Flags {
		if action.Flags[i].Name == "description" {
			description = &action.Flags[i]
		}
	}
	if description == nil {
		t.Fatal("bom-suggest must expose a `description` flag")
	}
	// Default "" keeps the POST body present — the backend binds
	// [FromBody] SuggestProjectBomRequestModel and 400s on a missing body.
	if description.Required || description.Type != "string" || description.Default != "" {
		t.Errorf("bom-suggest description must be an optional string defaulting to \"\", got %+v", description)
	}
}

func TestProjectCopilotBomApplyActionWired(t *testing.T) {
	action := findProjectCopilotAction(t, "bom-apply")
	if action.ToolName != "UteamupProjectApplyBom" || action.HTTPMethod != "POST" || action.RESTPath != "{projectGuid}/bom/apply" {
		t.Errorf("bom-apply must be POST {projectGuid}/bom/apply, got %+v", action)
	}
	if len(action.Args) != 1 || action.Args[0].Name != "projectGuid" || !action.Args[0].Required {
		t.Fatalf("bom-apply expected single required positional arg 'projectGuid', got %+v", action.Args)
	}

	var file *FlagDef
	for i := range action.Flags {
		if action.Flags[i].Name == "file" {
			file = &action.Flags[i]
		}
	}
	if file == nil {
		t.Fatal("bom-apply must expose a `file` flag")
	}
	if !file.Required || !file.JSONFile || file.Short != "f" {
		t.Errorf("bom-apply file flag must be Required JSONFile with -f short, got %+v", file)
	}
	if file.BodyName != "lines" {
		t.Errorf("bom-apply file BodyName = %q, want lines (backend binds ApplyProjectBomRequestModel.Lines)", file.BodyName)
	}
}
