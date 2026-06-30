package registry

import "testing"

// The analysis-preview / analyze actions on the workorder-template domain and the
// translate action on the language domain are the CLI surface for the "Analyze
// Workorders to enhance a template" + AI-translate features. The ToolName values
// are the backend contract — a typo ships a command that 404s server-side.

func TestWorkorderTemplateAnalysisActions(t *testing.T) {
	d := findDomain("workorder-template")
	if d == nil {
		t.Fatal("expected workorder-template domain to be registered")
	}

	cases := []struct{ action, tool string }{
		{"analysis-preview", "UteamupWorkorderTemplateAnalyzePreview"},
		{"analyze", "UteamupWorkorderTemplateAnalyze"},
	}
	for _, c := range cases {
		a := findAction(d, c.action)
		if a == nil {
			t.Fatalf("expected %q action on the workorder-template domain", c.action)
		}
		if a.ToolName != c.tool {
			t.Errorf("%s: expected tool %q, got %q", c.action, c.tool, a.ToolName)
		}
		var hasTemplate bool
		for _, f := range a.Flags {
			if f.Name == "template" {
				hasTemplate = true
				if !f.Required {
					t.Errorf("%s: flag --template must be Required", c.action)
				}
				if f.BodyName != "templateGuid" {
					t.Errorf("%s: flag --template must map to templateGuid, got %q", c.action, f.BodyName)
				}
			}
		}
		if !hasTemplate {
			t.Errorf("%s: missing required flag --template", c.action)
		}
	}
}

func TestLanguageTranslateAction(t *testing.T) {
	d := findDomain("language")
	if d == nil {
		t.Fatal("expected language domain to be registered")
	}
	a := findAction(d, "translate")
	if a == nil {
		t.Fatal("expected translate action on the language domain")
	}
	if a.ToolName != "UteamupLanguageTranslate" {
		t.Errorf("translate: expected tool UteamupLanguageTranslate, got %q", a.ToolName)
	}
	required := map[string]string{
		"source-text":  "sourceText",
		"source-lang":  "sourceLanguage",
		"target-langs": "targetLanguages",
	}
	seen := map[string]bool{}
	for _, f := range a.Flags {
		if body, ok := required[f.Name]; ok {
			seen[f.Name] = true
			if !f.Required {
				t.Errorf("translate: flag --%s must be Required", f.Name)
			}
			if f.BodyName != body {
				t.Errorf("translate: flag --%s must map to %q, got %q", f.Name, body, f.BodyName)
			}
		}
	}
	for name := range required {
		if !seen[name] {
			t.Errorf("translate: missing required flag --%s", name)
		}
	}
}
