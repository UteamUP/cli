package vendor

import (
	"testing"
)

func TestNewEnricher(t *testing.T) {
	// NewEnricher requires a genai.GenerativeModel which needs a real client.
	// Test the cache behavior with a nil model (won't make API calls).
	e := &Enricher{
		model: nil,
		cache: make(map[string]*EnrichedVendor),
	}

	if e.CacheSize() != 0 {
		t.Errorf("expected empty cache, got %d", e.CacheSize())
	}
}

func TestEnricher_EmptyVendorName(t *testing.T) {
	e := &Enricher{
		model: nil,
		cache: make(map[string]*EnrichedVendor),
	}

	// Empty vendor name should return nil without making any call.
	result := e.Enrich(nil, "")
	if result != nil {
		t.Error("expected nil for empty vendor name")
	}

	result = e.Enrich(nil, "   ")
	if result != nil {
		t.Error("expected nil for whitespace vendor name")
	}
}

func TestEnricher_CacheDedup(t *testing.T) {
	e := &Enricher{
		model: nil,
		cache: make(map[string]*EnrichedVendor),
	}

	// Pre-populate cache.
	e.cache["atlas copco"] = &EnrichedVendor{
		FullName:         "Atlas Copco AB",
		Website:          "https://www.atlascopco.com",
		BusinessCategory: "Manufacturing",
		Country:          "Sweden",
		Confidence:       0.95,
		Source:           "ai_enriched",
	}

	// Lookup should return cached result.
	result := e.Enrich(nil, "Atlas Copco")
	if result == nil {
		t.Fatal("expected cached result, got nil")
	}
	if result.FullName != "Atlas Copco AB" {
		t.Errorf("expected 'Atlas Copco AB', got %q", result.FullName)
	}
	if result.Source != "ai_enriched" {
		t.Errorf("expected source 'ai_enriched', got %q", result.Source)
	}

	// Case-insensitive lookup should also hit cache.
	result2 := e.Enrich(nil, "atlas copco")
	if result2 == nil {
		t.Fatal("expected cached result for lowercase, got nil")
	}
	if result2.FullName != "Atlas Copco AB" {
		t.Errorf("expected 'Atlas Copco AB', got %q", result2.FullName)
	}

	// Cache size should still be 1.
	if e.CacheSize() != 1 {
		t.Errorf("expected cache size 1, got %d", e.CacheSize())
	}
}

func TestEnrichedVendor_Fields(t *testing.T) {
	v := EnrichedVendor{
		FullName:         "DeWalt Industrial Tool Co.",
		Website:          "https://www.dewalt.com",
		BusinessCategory: "Tools",
		Country:          "United States",
		Confidence:       0.88,
		Source:           "ai_enriched",
	}

	if v.FullName != "DeWalt Industrial Tool Co." {
		t.Errorf("unexpected FullName: %s", v.FullName)
	}
	if v.Source != "ai_enriched" {
		t.Errorf("unexpected Source: %s", v.Source)
	}
}
