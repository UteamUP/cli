package analyzer

// Gemini 2.0 Flash pricing constants (per 1M tokens, as of 2025).
const (
	inputCostPer1MTokens         = 0.10 // USD
	outputCostPer1MTokens        = 0.40 // USD
	estimatedInputTokensPerImage = 258  // ~258 tokens for image encoding
	estimatedPromptTokens        = 1500 // prompt template overhead
	estimatedOutputTokens        = 500  // average JSON response
)

// CostEstimate holds token counts and estimated costs in USD.
type CostEstimate struct {
	ImageCount             int     `json:"image_count"`
	EstimatedInputTokens   int     `json:"estimated_input_tokens"`
	EstimatedOutputTokens  int     `json:"estimated_output_tokens"`
	EstimatedInputCostUSD  float64 `json:"estimated_input_cost_usd"`
	EstimatedOutputCostUSD float64 `json:"estimated_output_cost_usd"`
	EstimatedTotalCostUSD  float64 `json:"estimated_total_cost_usd"`
}

// EstimateCost returns a static cost estimate for processing imageCount images.
func EstimateCost(imageCount int) CostEstimate {
	totalInput := imageCount * (estimatedInputTokensPerImage + estimatedPromptTokens)
	totalOutput := imageCount * estimatedOutputTokens

	inputCost := float64(totalInput) / 1_000_000 * inputCostPer1MTokens
	outputCost := float64(totalOutput) / 1_000_000 * outputCostPer1MTokens

	return CostEstimate{
		ImageCount:             imageCount,
		EstimatedInputTokens:   totalInput,
		EstimatedOutputTokens:  totalOutput,
		EstimatedInputCostUSD:  inputCost,
		EstimatedOutputCostUSD: outputCost,
		EstimatedTotalCostUSD:  inputCost + outputCost,
	}
}

// TotalCost returns a running cost estimate based on images analyzed so far.
func (a *GeminiAnalyzer) TotalCost() CostEstimate {
	inputCost := float64(a.totalInputTokens) / 1_000_000 * inputCostPer1MTokens
	outputCost := float64(a.totalOutputTokens) / 1_000_000 * outputCostPer1MTokens

	return CostEstimate{
		ImageCount:             a.imagesAnalyzed,
		EstimatedInputTokens:   a.totalInputTokens,
		EstimatedOutputTokens:  a.totalOutputTokens,
		EstimatedInputCostUSD:  inputCost,
		EstimatedOutputCostUSD: outputCost,
		EstimatedTotalCostUSD:  inputCost + outputCost,
	}
}
