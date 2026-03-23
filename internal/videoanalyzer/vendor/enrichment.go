// Package vendor provides vendor enrichment via Gemini AI lookup.
package vendor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/google/generative-ai-go/genai"
)

// EnrichedVendor holds enriched vendor information from an AI lookup.
type EnrichedVendor struct {
	FullName         string  `json:"full_name"`
	Website          string  `json:"website"`
	BusinessCategory string  `json:"business_category"`
	Country          string  `json:"country"`
	Confidence       float64 `json:"confidence"`
	Source           string  `json:"source"` // "ai_enriched" or "not_enriched"
}

// Enricher performs vendor name lookups via Gemini and caches results.
type Enricher struct {
	model *genai.GenerativeModel
	cache map[string]*EnrichedVendor
	mu    sync.Mutex
}

// NewEnricher creates a new Enricher using the given Gemini model.
func NewEnricher(model *genai.GenerativeModel) *Enricher {
	return &Enricher{
		model: model,
		cache: make(map[string]*EnrichedVendor),
	}
}

// Enrich looks up additional information about a vendor by name.
// Results are cached by normalized vendor name to avoid duplicate lookups.
// Returns nil if enrichment fails or the vendor name is empty.
func (e *Enricher) Enrich(ctx context.Context, vendorName string) *EnrichedVendor {
	vendorName = strings.TrimSpace(vendorName)
	if vendorName == "" {
		return nil
	}

	key := strings.ToLower(vendorName)

	e.mu.Lock()
	if cached, ok := e.cache[key]; ok {
		e.mu.Unlock()
		return cached
	}
	e.mu.Unlock()

	result := e.lookup(ctx, vendorName)

	e.mu.Lock()
	e.cache[key] = result
	e.mu.Unlock()

	return result
}

// lookup performs the actual Gemini call to enrich a vendor name.
func (e *Enricher) lookup(ctx context.Context, vendorName string) *EnrichedVendor {
	prompt := fmt.Sprintf(vendorEnrichmentPrompt, vendorName)

	resp, err := e.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		log.Printf("vendor enrichment: Gemini call failed for %q: %v", vendorName, err)
		return &EnrichedVendor{Source: "not_enriched"}
	}

	if resp == nil || len(resp.Candidates) == 0 {
		log.Printf("vendor enrichment: empty response for %q", vendorName)
		return &EnrichedVendor{Source: "not_enriched"}
	}

	candidate := resp.Candidates[0]
	if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
		return &EnrichedVendor{Source: "not_enriched"}
	}

	var text string
	for _, part := range candidate.Content.Parts {
		if t, ok := part.(genai.Text); ok {
			text += string(t)
		}
	}

	text = strings.TrimSpace(text)
	if text == "" {
		return &EnrichedVendor{Source: "not_enriched"}
	}

	var enriched EnrichedVendor
	if err := json.Unmarshal([]byte(text), &enriched); err != nil {
		log.Printf("vendor enrichment: failed to parse response for %q: %v", vendorName, err)
		return &EnrichedVendor{Source: "not_enriched"}
	}

	if enriched.Confidence < 0.5 {
		log.Printf("vendor enrichment: low confidence (%.2f) for %q, using basic data", enriched.Confidence, vendorName)
		return &EnrichedVendor{Source: "not_enriched"}
	}

	enriched.Source = "ai_enriched"
	log.Printf("vendor enrichment: enriched %q → %q (%s)", vendorName, enriched.FullName, enriched.Website)
	return &enriched
}

// CacheSize returns the number of cached vendor lookups.
func (e *Enricher) CacheSize() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return len(e.cache)
}

// vendorEnrichmentPrompt is kept in sync with the constant in prompts.go.
// We use a package-level copy to avoid a circular reference.
const vendorEnrichmentPrompt = `Given this manufacturer/vendor name: "%s"

Provide the following information about this company. If you are not confident about any field, set it to null.

Return ONLY a JSON object:
{
  "full_name": "<official full legal company name>",
  "website": "<official website URL>",
  "business_category": "<primary business category: Manufacturing, Industrial Supply, Tools, Electronics, Automotive, Chemical, Safety Equipment, HVAC, Plumbing, Electrical, General Supply, Other>",
  "country": "<country of headquarters>",
  "confidence": 0.0 to 1.0
}

No markdown, no fences, no extra text. Only the JSON object.`
