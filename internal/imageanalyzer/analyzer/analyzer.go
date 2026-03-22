package analyzer

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"

	"github.com/uteamup/cli/internal/imageanalyzer/config"
	"github.com/uteamup/cli/internal/imageanalyzer/models"
	"github.com/uteamup/cli/internal/imageanalyzer/ratelimiter"
	"github.com/uteamup/cli/internal/imageanalyzer/retry"
)

// Analyzer defines the interface for image analysis.
type Analyzer interface {
	// AnalyzeImage sends an image to the AI model and returns one result per entity found.
	AnalyzeImage(ctx context.Context, imagePath string, imageBytes []byte) ([]models.ImageAnalysisResult, error)
	// EstimateCost returns a static cost estimate for the given number of images.
	EstimateCost(imageCount int) CostEstimate
	// TotalCost returns the running total cost of images analyzed so far.
	TotalCost() CostEstimate
}

// GeminiAnalyzer implements Analyzer using Google Gemini Vision AI.
type GeminiAnalyzer struct {
	client       *genai.Client
	model        *genai.GenerativeModel
	cfg          config.GeminiConfig
	rateLimiter  *ratelimiter.TokenBucketRateLimiter
	retryHandler *retry.RetryHandler

	mu                sync.Mutex
	totalInputTokens  int
	totalOutputTokens int
	imagesAnalyzed    int
}

// NewGeminiAnalyzer creates a new GeminiAnalyzer from the provided config.
func NewGeminiAnalyzer(cfg config.GeminiConfig) (*GeminiAnalyzer, error) {
	ctx := context.Background()

	client, err := genai.NewClient(ctx, option.WithAPIKey(cfg.APIKey))
	if err != nil {
		return nil, fmt.Errorf("creating Gemini client: %w", err)
	}

	model := client.GenerativeModel(cfg.Model)
	temp := float32(cfg.Temperature)
	model.Temperature = &temp
	maxTokens := int32(cfg.MaxOutputTokens)
	model.MaxOutputTokens = &maxTokens

	a := &GeminiAnalyzer{
		client:       client,
		model:        model,
		cfg:          cfg,
		rateLimiter:  ratelimiter.NewTokenBucket(cfg.RequestsPerMinute),
		retryHandler: retry.NewRetryHandler(cfg.MaxRetries),
	}

	log.Printf("analyzer: initialized model=%s temperature=%.1f rpm=%d",
		cfg.Model, cfg.Temperature, cfg.RequestsPerMinute)

	return a, nil
}

// EstimateCost returns a static cost estimate for the given number of images.
func (a *GeminiAnalyzer) EstimateCost(imageCount int) CostEstimate {
	return EstimateCost(imageCount)
}

// AnalyzeImage sends an image to Gemini for classification and data extraction.
// A single image may contain multiple entities. Returns one result per entity found.
func (a *GeminiAnalyzer) AnalyzeImage(ctx context.Context, imagePath string, imageBytes []byte) ([]models.ImageAnalysisResult, error) {
	originalFilename := filepath.Base(imagePath)

	log.Printf("analyzer: analyzing image=%s", originalFilename)

	// Wait for rate limiter.
	a.rateLimiter.Acquire()

	// Call Gemini with retry handling.
	result, err := a.retryHandler.Execute(func() (interface{}, error) {
		return a.callGemini(ctx, imageBytes)
	})
	if err != nil {
		log.Printf("analyzer: api_error image=%s error=%v", originalFilename, err)
		return []models.ImageAnalysisResult{
			{
				ImagePath:        imagePath,
				OriginalFilename: originalFilename,
				Classification: models.ClassificationResult{
					PrimaryType: models.EntityTypeUnclassified,
					Confidence:  0.0,
					Reasoning:   fmt.Sprintf("API error: %v", err),
				},
				FlaggedForReview: true,
				ReviewReason:     fmt.Sprintf("API error: %v", err),
				ProcessedAt:      time.Now(),
			},
		}, nil
	}

	responseText := result.(string)
	results := parseMultiEntityResponse(responseText, imagePath)

	// Track costs.
	a.mu.Lock()
	a.imagesAnalyzed++
	a.totalInputTokens += estimatedInputTokensPerImage + estimatedPromptTokens
	a.totalOutputTokens += estimatedOutputTokens
	a.mu.Unlock()

	// Flag low confidence results.
	for i := range results {
		if results[i].Classification.Confidence < 0.5 {
			results[i].FlaggedForReview = true
			results[i].ReviewReason = fmt.Sprintf("Low confidence: %.2f", results[i].Classification.Confidence)
		}
		results[i].ProcessedAt = time.Now()

		log.Printf("analyzer: result image=%s entity_type=%s confidence=%.2f flagged=%t related_to=%s",
			originalFilename,
			results[i].Classification.PrimaryType,
			results[i].Classification.Confidence,
			results[i].FlaggedForReview,
			results[i].RelatedTo,
		)
	}

	return results, nil
}

// callGemini makes the actual API call to Gemini and extracts the text response.
func (a *GeminiAnalyzer) callGemini(ctx context.Context, imageBytes []byte) (string, error) {
	resp, err := a.model.GenerateContent(ctx,
		genai.Text(UnifiedAnalysisPrompt),
		genai.ImageData("image/jpeg", imageBytes),
	)
	if err != nil {
		return "", fmt.Errorf("Gemini GenerateContent: %w", err)
	}

	// Extract text from the response.
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
