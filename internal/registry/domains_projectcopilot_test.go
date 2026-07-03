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

func TestProjectCopilotImageReportActionWired(t *testing.T) {
	action := findProjectCopilotAction(t, "image-report")
	if action.ToolName != "UteamupProjectGenerateImageReport" || action.HTTPMethod != "POST" || action.RESTPath != "{projectGuid}/image-report" {
		t.Errorf("image-report must be POST {projectGuid}/image-report, got %+v", action)
	}
	if len(action.Args) != 1 || action.Args[0].Name != "projectGuid" || !action.Args[0].Required || action.Args[0].Type != "string" {
		t.Fatalf("image-report expected single required positional arg 'projectGuid', got %+v", action.Args)
	}
}

// --- AI Planning Suite (phase B4) ---

func TestProjectCopilotAiPlanningActionsWired(t *testing.T) {
	// Every AI planning action is a POST with a single required string
	// positional arg 'projectGuid' — the name must literally match the
	// RESTPath placeholder or expandPathTemplate leaves the raw token.
	cases := []struct {
		action   string
		tool     string
		restPath string
		// bodyName is the JSONFile flag's body field for apply actions
		// (mirrors the backend Apply*Request list property); "" = no flags.
		bodyName string
	}{
		{"wbs-suggest", "UteamupProjectSuggestWbs", "{projectGuid}/wbs/suggest", ""},
		{"wbs-apply", "UteamupProjectApplyWbs", "{projectGuid}/wbs/apply", "stages"},
		{"prioritize-suggest", "UteamupProjectSuggestPrioritization", "{projectGuid}/prioritize/suggest", ""},
		{"prioritize-apply", "UteamupProjectApplyPrioritization", "{projectGuid}/prioritize/apply", "items"},
		{"risks-suggest", "UteamupProjectSuggestRisks", "{projectGuid}/risks/suggest", ""},
		{"risks-apply", "UteamupProjectApplyRisks", "{projectGuid}/risks/apply", "risks"},
		{"lessons-learned", "UteamupProjectGenerateLessonsLearned", "{projectGuid}/lessons-learned", ""},
	}
	for _, c := range cases {
		action := findProjectCopilotAction(t, c.action)
		if action.ToolName != c.tool || action.HTTPMethod != "POST" || action.RESTPath != c.restPath {
			t.Errorf("%s: want tool=%s POST %s, got tool=%s method=%q path=%s",
				c.action, c.tool, c.restPath, action.ToolName, action.HTTPMethod, action.RESTPath)
		}
		if len(action.Args) != 1 || action.Args[0].Name != "projectGuid" || !action.Args[0].Required || action.Args[0].Type != "string" {
			t.Errorf("%s expected single required string positional arg 'projectGuid', got %+v", c.action, action.Args)
		}

		if c.bodyName == "" {
			// Suggest + lessons-learned endpoints bind no request body — a flag
			// here would leak an unexpected JSON field into the POST.
			if len(action.Flags) != 0 {
				t.Errorf("%s should take no flags, got %+v", c.action, action.Flags)
			}
			continue
		}

		var file *FlagDef
		for i := range action.Flags {
			if action.Flags[i].Name == "file" {
				file = &action.Flags[i]
			}
		}
		if file == nil {
			t.Errorf("%s must expose a `file` flag", c.action)
			continue
		}
		if !file.Required || !file.JSONFile || file.Short != "f" || file.Type != "string" {
			t.Errorf("%s file flag must be a Required string JSONFile with -f short, got %+v", c.action, file)
		}
		if file.BodyName != c.bodyName {
			t.Errorf("%s file BodyName = %q, want %q (backend Apply*Request property)", c.action, file.BodyName, c.bodyName)
		}
	}
}

func TestProjectCopilotEstimateActionWired(t *testing.T) {
	action := findProjectCopilotAction(t, "estimate")
	if action.ToolName != "UteamupProjectEstimate" || action.HTTPMethod != "POST" || action.RESTPath != "{projectGuid}/estimate" {
		t.Errorf("estimate must be POST {projectGuid}/estimate, got %+v", action)
	}
	if len(action.Args) != 1 || action.Args[0].Name != "projectGuid" || !action.Args[0].Required || action.Args[0].Type != "string" {
		t.Fatalf("estimate expected single required positional arg 'projectGuid', got %+v", action.Args)
	}
}

func TestProjectCopilotEstimateApplyActionWired(t *testing.T) {
	action := findProjectCopilotAction(t, "estimate-apply")
	if action.ToolName != "UteamupProjectApplyEstimate" || action.HTTPMethod != "POST" || action.RESTPath != "{projectGuid}/estimate/apply" {
		t.Errorf("estimate-apply must be POST {projectGuid}/estimate/apply, got %+v", action)
	}
	if len(action.Args) != 1 || action.Args[0].Name != "projectGuid" || !action.Args[0].Required {
		t.Fatalf("estimate-apply expected single required positional arg 'projectGuid', got %+v", action.Args)
	}
	// The four scalar flags must map to the flat ApplyProjectEstimateRequest body.
	want := map[string]string{
		"duration-days":  "estimatedDurationDays",
		"cost":           "estimatedCost",
		"apply-duration": "applyDuration",
		"apply-cost":     "applyCost",
	}
	got := map[string]string{}
	for i := range action.Flags {
		got[action.Flags[i].Name] = action.Flags[i].BodyName
	}
	for name, bodyName := range want {
		if got[name] != bodyName {
			t.Errorf("estimate-apply flag %q BodyName = %q, want %q", name, got[name], bodyName)
		}
	}
}
