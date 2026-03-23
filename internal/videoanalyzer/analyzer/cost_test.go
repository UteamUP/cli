package analyzer

import (
	"strings"
	"testing"
)

func TestEstimateCost(t *testing.T) {
	est := EstimateCost(5)

	if est.InputTokens <= 0 {
		t.Errorf("expected InputTokens > 0, got %d", est.InputTokens)
	}
	if est.OutputTokens <= 0 {
		t.Errorf("expected OutputTokens > 0, got %d", est.OutputTokens)
	}
	if est.EstimatedCostUSD <= 0 {
		t.Errorf("expected EstimatedCostUSD > 0, got %f", est.EstimatedCostUSD)
	}
	if est.VideosProcessed != 5 {
		t.Errorf("expected VideosProcessed=5, got %d", est.VideosProcessed)
	}
	if est.TotalTokens != est.InputTokens+est.OutputTokens {
		t.Errorf("expected TotalTokens=%d, got %d", est.InputTokens+est.OutputTokens, est.TotalTokens)
	}
}

func TestEstimateCost_Zero(t *testing.T) {
	est := EstimateCost(0)

	if est.InputTokens != 0 {
		t.Errorf("expected InputTokens=0, got %d", est.InputTokens)
	}
	if est.OutputTokens != 0 {
		t.Errorf("expected OutputTokens=0, got %d", est.OutputTokens)
	}
	if est.EstimatedCostUSD != 0 {
		t.Errorf("expected EstimatedCostUSD=0, got %f", est.EstimatedCostUSD)
	}
	if est.VideosProcessed != 0 {
		t.Errorf("expected VideosProcessed=0, got %d", est.VideosProcessed)
	}
	if est.TotalTokens != 0 {
		t.Errorf("expected TotalTokens=0, got %d", est.TotalTokens)
	}
}

func TestCostTracker_AddVideo(t *testing.T) {
	ct := NewCostTracker()
	ct.AddVideo(1000, 500)
	ct.AddVideo(2000, 800)
	ct.AddVideo(1500, 600)

	total := ct.TotalCost()
	if total.VideosProcessed != 3 {
		t.Errorf("expected 3 videos processed, got %d", total.VideosProcessed)
	}
	if total.InputTokens != 4500 {
		t.Errorf("expected InputTokens=4500, got %d", total.InputTokens)
	}
	if total.OutputTokens != 1900 {
		t.Errorf("expected OutputTokens=1900, got %d", total.OutputTokens)
	}
	if total.TotalTokens != 6400 {
		t.Errorf("expected TotalTokens=6400, got %d", total.TotalTokens)
	}
	if total.EstimatedCostUSD <= 0 {
		t.Errorf("expected EstimatedCostUSD > 0, got %f", total.EstimatedCostUSD)
	}
}

func TestCostTracker_AddVendorLookup(t *testing.T) {
	ct := NewCostTracker()
	ct.AddVendorLookup(200, 150)
	ct.AddVendorLookup(200, 150)

	total := ct.TotalCost()
	if total.VendorLookups != 2 {
		t.Errorf("expected 2 vendor lookups, got %d", total.VendorLookups)
	}
	if total.InputTokens != 400 {
		t.Errorf("expected InputTokens=400, got %d", total.InputTokens)
	}
	if total.OutputTokens != 300 {
		t.Errorf("expected OutputTokens=300, got %d", total.OutputTokens)
	}
}

func TestCostEstimate_String(t *testing.T) {
	est := CostEstimate{
		InputTokens:      10000,
		OutputTokens:     5000,
		TotalTokens:      15000,
		EstimatedCostUSD: 0.0030,
		VideosProcessed:  2,
		VendorLookups:    1,
	}

	s := est.String()
	if !strings.Contains(s, "tokens:") {
		t.Errorf("expected String() to contain 'tokens:', got %q", s)
	}
	if !strings.Contains(s, "$") {
		t.Errorf("expected String() to contain '$', got %q", s)
	}
}
