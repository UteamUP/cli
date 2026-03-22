package analyzer

import (
	"context"
	"fmt"
	"log"

	"github.com/google/generative-ai-go/genai"
)

// attemptJSONFix sends broken JSON back to Gemini for correction.
func (a *GeminiAnalyzer) attemptJSONFix(ctx context.Context, brokenText string) (map[string]interface{}, error) {
	fixPrompt := fmt.Sprintf(JSONFixPrompt, brokenText)

	a.rateLimiter.Acquire()

	result, err := a.retryHandler.Execute(func() (interface{}, error) {
		return a.callGeminiText(ctx, fixPrompt)
	})
	if err != nil {
		log.Printf("analyzer: json_fix_failed error=%v", err)
		return nil, fmt.Errorf("JSON fix request failed: %w", err)
	}

	fixText := result.(string)
	parsed, parseErr := tryParseJSON(fixText)
	if parseErr != nil {
		log.Printf("analyzer: json_fix_parse_failed error=%v", parseErr)
		return nil, fmt.Errorf("fixed JSON still invalid: %w", parseErr)
	}

	return parsed, nil
}

// callGeminiText sends a text-only prompt to Gemini (no image).
func (a *GeminiAnalyzer) callGeminiText(ctx context.Context, prompt string) (string, error) {
	resp, err := a.model.GenerateContent(ctx,
		genai.Text(prompt),
	)
	if err != nil {
		return "", fmt.Errorf("Gemini GenerateContent: %w", err)
	}

	if resp == nil || len(resp.Candidates) == 0 {
		return "", fmt.Errorf("empty response from Gemini")
	}

	candidate := resp.Candidates[0]
	if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
		return "", fmt.Errorf("no content parts in Gemini response")
	}

	var text string
	for _, part := range candidate.Content.Parts {
		if t, ok := part.(genai.Text); ok {
			text += string(t)
		}
	}

	if text == "" {
		return "", fmt.Errorf("no text content in Gemini response")
	}

	return text, nil
}
