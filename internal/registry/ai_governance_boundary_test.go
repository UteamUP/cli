package registry

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestAIProviderAdministrationIsNotExposedByCLI(t *testing.T) {
	forbiddenFlags := map[string]struct{}{
		"api-key":  {},
		"base-url": {},
		"model":    {},
		"provider": {},
		"secret":   {},
		"task-key": {},
	}

	for _, domain := range DefaultRegistry.Domains() {
		if domain.Name == "ai-provider" ||
			domain.APIPath == "/api/tenant-ai-provider" ||
			domain.APIPath == "/api/tenant-ai-governance" {
			t.Fatalf("AI provider administration must remain web-only: domain=%q path=%q", domain.Name, domain.APIPath)
		}

		for _, action := range domain.Actions {
			toolName := strings.ToLower(action.ToolName)
			if strings.Contains(toolName, "aiproviderconfig") ||
				strings.Contains(toolName, "aicontrolplanesnapshot") {
				t.Fatalf("forbidden AI control-plane tool registered: %s", action.ToolName)
			}

			for _, flag := range action.Flags {
				if _, forbidden := forbiddenFlags[strings.ToLower(flag.Name)]; forbidden {
					t.Fatalf(
						"CLI action %s %s exposes forbidden AI routing flag --%s",
						domain.Name,
						action.Name,
						flag.Name,
					)
				}
			}
		}
	}
}

func TestCLIHasNoDirectAIProviderDependencyOrEndpoint(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not resolve the test source path")
	}
	repositoryRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", ".."))

	moduleBytes, err := os.ReadFile(filepath.Join(repositoryRoot, "go.mod"))
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	moduleText := strings.ToLower(string(moduleBytes))
	for _, forbiddenDependency := range []string{
		"anthropic-sdk-go",
		"generative-ai-go",
		"google.golang.org/genai",
		"mistralai",
		"openai-go",
	} {
		if strings.Contains(moduleText, forbiddenDependency) {
			t.Fatalf("direct AI provider SDK dependency is forbidden: %s", forbiddenDependency)
		}
	}

	forbiddenSourceMarkers := []string{
		"anthropic_api_key",
		"api.anthropic.com",
		"api.mistral.ai",
		"api.minimax.io",
		"api.moonshot.ai",
		"api.openai.com",
		"api.x.ai",
		"gemini_api_key",
		"generativelanguage.googleapis.com",
		"google_api_key",
		"mistral_api_key",
		"openai_api_key",
		"xai_api_key",
	}
	err = filepath.WalkDir(repositoryRoot, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		contents, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		lowerContents := strings.ToLower(string(contents))
		for _, marker := range forbiddenSourceMarkers {
			if strings.Contains(lowerContents, marker) {
				t.Errorf("direct AI provider marker %q found in %s", marker, path)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("scan CLI source: %v", err)
	}
}
