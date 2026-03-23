package analyzer

import (
	"fmt"
	"sync"
)

// Video analysis token estimates.
// Videos use significantly more tokens than images due to frame-by-frame processing.
const (
	// Estimated input tokens per minute of video (Gemini processes ~1 frame/sec).
	estimatedInputTokensPerMinute = 4500

	// Average video duration in minutes for estimation (when actual duration is unknown).
	estimatedAvgVideoDurationMin = 2.0

	// Prompt overhead tokens.
	estimatedPromptTokens = 2000

	// Average output tokens per video.
	estimatedOutputTokens = 1500

	// Gemini 2.5 Flash pricing (per 1M tokens, as of 2025).
	inputCostPer1MTokens  = 0.10 // USD
	outputCostPer1MTokens = 0.40 // USD

	// Vendor enrichment tokens per lookup.
	vendorEnrichmentInputTokens  = 200
	vendorEnrichmentOutputTokens = 150
)

// CostEstimate holds estimated or actual token usage and cost.
type CostEstimate struct {
	InputTokens       int
	OutputTokens      int
	TotalTokens       int
	EstimatedCostUSD  float64
	VideosProcessed   int
	VendorLookups     int
}

// String returns a human-readable cost summary.
func (c CostEstimate) String() string {
	return fmt.Sprintf("tokens: %d input + %d output = %d total, est. cost: $%.4f (%d videos, %d vendor lookups)",
		c.InputTokens, c.OutputTokens, c.TotalTokens, c.EstimatedCostUSD, c.VideosProcessed, c.VendorLookups)
}

// CostTracker tracks running token usage and cost across multiple video analyses.
type CostTracker struct {
	mu            sync.Mutex
	inputTokens   int
	outputTokens  int
	videosProcessed int
	vendorLookups int
}

// NewCostTracker creates a new CostTracker.
func NewCostTracker() *CostTracker {
	return &CostTracker{}
}

// AddVideo records token usage for a video analysis.
func (ct *CostTracker) AddVideo(inputTokens, outputTokens int) {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	ct.inputTokens += inputTokens
	ct.outputTokens += outputTokens
	ct.videosProcessed++
}

// AddVendorLookup records token usage for a vendor enrichment call.
func (ct *CostTracker) AddVendorLookup(inputTokens, outputTokens int) {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	ct.inputTokens += inputTokens
	ct.outputTokens += outputTokens
	ct.vendorLookups++
}

// TotalCost returns the running total cost estimate.
func (ct *CostTracker) TotalCost() CostEstimate {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	total := ct.inputTokens + ct.outputTokens
	cost := float64(ct.inputTokens)*inputCostPer1MTokens/1_000_000 +
		float64(ct.outputTokens)*outputCostPer1MTokens/1_000_000
	return CostEstimate{
		InputTokens:      ct.inputTokens,
		OutputTokens:     ct.outputTokens,
		TotalTokens:      total,
		EstimatedCostUSD: cost,
		VideosProcessed:  ct.videosProcessed,
		VendorLookups:    ct.vendorLookups,
	}
}

// EstimateCost returns a static cost estimate for the given number of videos.
// Uses average video duration and token estimates.
func EstimateCost(videoCount int) CostEstimate {
	inputPerVideo := int(estimatedInputTokensPerMinute*estimatedAvgVideoDurationMin) + estimatedPromptTokens
	totalInput := inputPerVideo * videoCount
	totalOutput := estimatedOutputTokens * videoCount
	total := totalInput + totalOutput
	cost := float64(totalInput)*inputCostPer1MTokens/1_000_000 +
		float64(totalOutput)*outputCostPer1MTokens/1_000_000
	return CostEstimate{
		InputTokens:      totalInput,
		OutputTokens:     totalOutput,
		TotalTokens:      total,
		EstimatedCostUSD: cost,
		VideosProcessed:  videoCount,
	}
}
