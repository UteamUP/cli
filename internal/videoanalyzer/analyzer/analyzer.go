package analyzer

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"

	"github.com/uteamup/cli/internal/imageanalyzer/models"
	"github.com/uteamup/cli/internal/imageanalyzer/ratelimiter"
	"github.com/uteamup/cli/internal/imageanalyzer/retry"
	"github.com/uteamup/cli/internal/videoanalyzer/spinner"
)

// Analyzer defines the interface for video analysis.
type Analyzer interface {
	// AnalyzeVideo uploads a video to Gemini, waits for processing, and returns analysis results.
	AnalyzeVideo(ctx context.Context, videoPath string, mimeType string) ([]models.ImageAnalysisResult, error)
	// CostTracker returns the cost tracker for cumulative cost reporting.
	CostTracker() *CostTracker
	// Close releases resources (Gemini client).
	Close() error
}

// VideoAnalyzer implements Analyzer using Google Gemini Vision AI via the File API.
type VideoAnalyzer struct {
	client      *genai.Client
	model       *genai.GenerativeModel
	rateLimiter *ratelimiter.TokenBucketRateLimiter
	retryHandler *retry.RetryHandler
	costTracker *CostTracker
	spinner     *spinner.Spinner
	pollInterval time.Duration
	pollTimeout  time.Duration
}

// Config holds configuration for creating a VideoAnalyzer.
type Config struct {
	APIKey            string
	Model             string
	MaxOutputTokens   int
	Temperature       float64
	RequestsPerMinute int
	MaxRetries        int
	PollIntervalSec   int
	PollTimeoutSec    int
}

// NewVideoAnalyzer creates a new VideoAnalyzer with the given configuration.
func NewVideoAnalyzer(ctx context.Context, cfg Config) (*VideoAnalyzer, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(cfg.APIKey))
	if err != nil {
		return nil, fmt.Errorf("creating Gemini client: %w", err)
	}

	model := client.GenerativeModel(cfg.Model)
	model.ResponseMIMEType = "application/json"
	if cfg.MaxOutputTokens > 0 {
		model.MaxOutputTokens = toInt32Ptr(int32(cfg.MaxOutputTokens))
	}
	if cfg.Temperature > 0 {
		model.Temperature = toFloat32Ptr(float32(cfg.Temperature))
	}

	pollInterval := time.Duration(cfg.PollIntervalSec) * time.Second
	if pollInterval == 0 {
		pollInterval = 5 * time.Second
	}
	pollTimeout := time.Duration(cfg.PollTimeoutSec) * time.Second
	if pollTimeout == 0 {
		pollTimeout = 10 * time.Minute
	}

	rpm := cfg.RequestsPerMinute
	if rpm <= 0 {
		rpm = 10
	}

	maxRetries := cfg.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 3
	}

	return &VideoAnalyzer{
		client:       client,
		model:        model,
		rateLimiter:  ratelimiter.NewTokenBucket(rpm),
		retryHandler: retry.NewRetryHandler(maxRetries),
		costTracker:  NewCostTracker(),
		spinner:      spinner.New(),
		pollInterval: pollInterval,
		pollTimeout:  pollTimeout,
	}, nil
}

// AnalyzeVideo uploads a video to Gemini File API, polls until processing completes,
// sends the analysis prompt, and returns parsed results.
func (a *VideoAnalyzer) AnalyzeVideo(ctx context.Context, videoPath string, mimeType string) ([]models.ImageAnalysisResult, error) {
	filename := filepath.Base(videoPath)
	log.Printf("analyzer: starting video analysis file=%s mime=%s", filename, mimeType)

	// Rate limit before upload.
	a.rateLimiter.Acquire()

	// Upload the video file.
	a.spinner.Start(fmt.Sprintf("Uploading %s", filename))
	file, err := a.uploadVideo(ctx, videoPath, mimeType)
	if err != nil {
		a.spinner.StopWithMessage(fmt.Sprintf("Upload failed: %s", filename))
		return nil, fmt.Errorf("uploading video %s: %w", filename, err)
	}

	// Ensure cleanup of uploaded file.
	defer func() {
		if delErr := a.client.DeleteFile(ctx, file.Name); delErr != nil {
			log.Printf("analyzer: warning: failed to delete uploaded file %s: %v", file.Name, delErr)
		}
	}()

	// Poll until file is processed.
	a.spinner.UpdateText(fmt.Sprintf("Processing %s", filename))
	file, err = a.waitForProcessing(ctx, file)
	if err != nil {
		a.spinner.StopWithMessage(fmt.Sprintf("Processing failed: %s", filename))
		return nil, fmt.Errorf("processing video %s: %w", filename, err)
	}
	a.spinner.StopWithMessage(fmt.Sprintf("Video ready: %s", filename))

	// Send analysis prompt.
	log.Printf("analyzer: sending analysis prompt file=%s uri=%s", filename, file.URI)
	text, err := a.callGemini(ctx, file)
	if err != nil {
		return nil, fmt.Errorf("analyzing video %s: %w", filename, err)
	}

	// Parse the response.
	results, err := ParseVideoResponse(text, videoPath, filename)
	if err != nil {
		return nil, fmt.Errorf("parsing response for video %s: %w", filename, err)
	}

	// Track cost (estimate based on defaults since exact token counts require usage metadata).
	inputTokens := int(estimatedInputTokensPerMinute*estimatedAvgVideoDurationMin) + estimatedPromptTokens
	a.costTracker.AddVideo(inputTokens, estimatedOutputTokens)

	log.Printf("analyzer: video analysis complete file=%s entities=%d", filename, len(results))
	return results, nil
}

// CostTracker returns the cost tracker.
func (a *VideoAnalyzer) CostTracker() *CostTracker {
	return a.costTracker
}

// Close releases the Gemini client resources.
func (a *VideoAnalyzer) Close() error {
	return a.client.Close()
}

// uploadVideo uploads a video file to Gemini's file storage.
func (a *VideoAnalyzer) uploadVideo(ctx context.Context, videoPath string, mimeType string) (*genai.File, error) {
	opts := &genai.UploadFileOptions{
		MIMEType: mimeType,
	}
	result, err := a.retryHandler.Execute(func() (interface{}, error) {
		return a.client.UploadFileFromPath(ctx, videoPath, opts)
	})
	if err != nil {
		return nil, err
	}
	return result.(*genai.File), nil
}

// waitForProcessing polls the file status until it becomes active or fails.
func (a *VideoAnalyzer) waitForProcessing(ctx context.Context, file *genai.File) (*genai.File, error) {
	deadline := time.Now().Add(a.pollTimeout)

	for file.State == genai.FileStateProcessing {
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout waiting for video processing after %v", a.pollTimeout)
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(a.pollInterval):
		}

		var err error
		file, err = a.client.GetFile(ctx, file.Name)
		if err != nil {
			return nil, fmt.Errorf("checking file status: %w", err)
		}
	}

	if file.State != genai.FileStateActive {
		return nil, fmt.Errorf("video processing failed with state: %s", file.State)
	}

	return file, nil
}

// callGemini sends the analysis prompt with the video file reference and returns the text response.
func (a *VideoAnalyzer) callGemini(ctx context.Context, file *genai.File) (string, error) {
	result, err := a.retryHandler.Execute(func() (interface{}, error) {
		resp, genErr := a.model.GenerateContent(ctx,
			genai.Text(VideoAnalysisPrompt),
			genai.FileData{URI: file.URI},
		)
		if genErr != nil {
			return nil, fmt.Errorf("Gemini GenerateContent: %w", genErr)
		}

		if resp == nil || len(resp.Candidates) == 0 {
			return nil, fmt.Errorf("empty response from Gemini")
		}

		candidate := resp.Candidates[0]
		if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
			return nil, fmt.Errorf("no content parts in Gemini response")
		}

		var sb []byte
		for _, part := range candidate.Content.Parts {
			if t, ok := part.(genai.Text); ok {
				sb = append(sb, []byte(t)...)
			}
		}

		if len(sb) == 0 {
			return nil, fmt.Errorf("no text content in Gemini response")
		}

		return string(sb), nil
	})
	if err != nil {
		return "", err
	}
	return result.(string), nil
}

func toInt32Ptr(v int32) *int32       { return &v }
func toFloat32Ptr(v float32) *float32 { return &v }
