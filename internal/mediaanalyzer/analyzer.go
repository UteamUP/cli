package mediaanalyzer

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/uteamup/cli/internal/client"
	clierrors "github.com/uteamup/cli/internal/errors"
	"github.com/uteamup/cli/internal/imageanalyzer/models"
)

const (
	PhotoEndpoint = "/api/inventoryai/analyze-photo"
	VideoEndpoint = "/api/inventoryai/analyze-video"

	MaxPhotoBytes = 15 * 1024 * 1024
	MaxVideoBytes = 100 * 1024 * 1024
	maxItems      = 100
)

var guidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)
var currencyPattern = regexp.MustCompile(`^[A-Z]{3}$`)

type uploadClient interface {
	CallRESTUploadLimited(
		ctx context.Context,
		method string,
		path string,
		fileField string,
		filePath string,
		uploadFileName string,
		contentType string,
		maxFileBytes int64,
		extraHeaders map[string]string,
	) (json.RawMessage, error)
}

// Analyzer sends media only to authenticated UteamUP backend routes. Provider,
// model, task identity, credentials, and fallback selection remain server-side.
type Analyzer struct {
	client uploadClient
}

func New(client *client.APIClient) *Analyzer {
	return &Analyzer{client: client}
}

func newWithClient(client uploadClient) *Analyzer {
	return &Analyzer{client: client}
}

type Result struct {
	Items   []models.ImageAnalysisResult
	Receipt UsageReceipt
}

type UsageReceipt struct {
	RequestGuid      string   `json:"requestGuid"`
	CredentialSource string   `json:"credentialSource"`
	ProviderAlias    string   `json:"providerAlias"`
	ModelAlias       string   `json:"modelAlias"`
	FallbackUsed     bool     `json:"fallbackUsed"`
	InputTokens      *int64   `json:"inputTokens"`
	OutputTokens     *int64   `json:"outputTokens"`
	CachedTokens     *int64   `json:"cachedTokens"`
	ReasoningTokens  *int64   `json:"reasoningTokens"`
	EstimatedCost    *float64 `json:"estimatedCost"`
	Currency         *string  `json:"currency"`
	CreditsCharged   int      `json:"creditsCharged"`
}

type analysisResponse struct {
	Items              []analysisItem `json:"items"`
	UsageReceipt       UsageReceipt   `json:"usageReceipt"`
	Warnings           []string       `json:"warnings"`
	ErrorCode          string         `json:"errorCode"`
	QuotaBlockedReason string         `json:"quotaBlockedReason"`
}

type analysisItem struct {
	Type             string               `json:"type"`
	Confidence       *float64             `json:"confidence"`
	Reasoning        string               `json:"reasoning"`
	FlaggedForReview bool                 `json:"flaggedForReview"`
	ReviewReason     string               `json:"reviewReason"`
	RelatedTo        string               `json:"relatedTo"`
	Timestamp        string               `json:"timestamp"`
	ExtractedData    models.ExtractedData `json:"extractedData"`
}

func (a *Analyzer) AnalyzePhoto(ctx context.Context, sourcePath string, imageBytes []byte) (Result, error) {
	if len(imageBytes) == 0 {
		return Result{}, fmt.Errorf("image is empty")
	}
	if len(imageBytes) > MaxPhotoBytes {
		return Result{}, fmt.Errorf("processed image exceeds the %d byte limit", MaxPhotoBytes)
	}

	temporary, err := os.CreateTemp("", "uteamup-media-*.jpg")
	if err != nil {
		return Result{}, fmt.Errorf("preparing image upload: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o600); err != nil {
		temporary.Close()
		return Result{}, fmt.Errorf("securing image upload: %w", err)
	}
	if _, err := temporary.Write(imageBytes); err != nil {
		temporary.Close()
		return Result{}, fmt.Errorf("preparing image upload: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return Result{}, fmt.Errorf("preparing image upload: %w", err)
	}

	return a.analyze(ctx, PhotoEndpoint, temporaryPath, filepath.Base(sourcePath), "image/jpeg", MaxPhotoBytes, sourcePath)
}

func (a *Analyzer) AnalyzeVideo(ctx context.Context, videoPath string, contentType string) (Result, error) {
	if contentType != "video/mp4" && contentType != "video/quicktime" {
		return Result{}, fmt.Errorf("unsupported video content type")
	}
	return a.analyze(ctx, VideoEndpoint, videoPath, filepath.Base(videoPath), contentType, MaxVideoBytes, videoPath)
}

func (a *Analyzer) analyze(
	ctx context.Context,
	endpoint string,
	uploadPath string,
	uploadFileName string,
	contentType string,
	maxFileBytes int64,
	sourcePath string,
) (Result, error) {
	idempotencyKey, err := newRequestGUID()
	if err != nil {
		return Result{}, fmt.Errorf("creating request identity: %w", err)
	}

	raw, err := a.client.CallRESTUploadLimited(
		ctx,
		"POST",
		endpoint,
		"file",
		uploadPath,
		uploadFileName,
		contentType,
		maxFileBytes,
		map[string]string{"Idempotency-Key": idempotencyKey},
	)
	if err != nil {
		return Result{}, safeUploadError(err)
	}

	var response analysisResponse
	if err := json.Unmarshal(raw, &response); err != nil {
		return Result{}, fmt.Errorf("backend returned an invalid media analysis response")
	}
	if response.ErrorCode != "" || response.QuotaBlockedReason != "" {
		return Result{}, safeBusinessError(response.ErrorCode, response.QuotaBlockedReason)
	}
	if !guidPattern.MatchString(response.UsageReceipt.RequestGuid) {
		return Result{}, fmt.Errorf("backend media analysis response is missing a valid request GUID")
	}
	receipt, err := normalizeReceipt(response.UsageReceipt)
	if err != nil {
		return Result{}, err
	}
	if len(response.Items) > maxItems {
		return Result{}, fmt.Errorf("backend media analysis response exceeds the %d item limit", maxItems)
	}

	items := make([]models.ImageAnalysisResult, 0, len(response.Items))
	for _, item := range response.Items {
		mapped, err := mapItem(item, sourcePath, uploadFileName)
		if err != nil {
			return Result{}, err
		}
		items = append(items, mapped)
	}

	return Result{Items: items, Receipt: receipt}, nil
}

func normalizeReceipt(receipt UsageReceipt) (UsageReceipt, error) {
	switch strings.ToLower(cleanText(receipt.CredentialSource, 32)) {
	case "managed":
		receipt.CredentialSource = "Managed"
	case "tenantbyok":
		receipt.CredentialSource = "TenantByok"
	default:
		return UsageReceipt{}, fmt.Errorf("backend media analysis response has an invalid credential source")
	}

	receipt.ProviderAlias = cleanText(receipt.ProviderAlias, 100)
	receipt.ModelAlias = cleanText(receipt.ModelAlias, 100)
	if receipt.ProviderAlias == "" || receipt.ModelAlias == "" {
		return UsageReceipt{}, fmt.Errorf("backend media analysis response is missing its provider or model alias")
	}
	if receipt.CreditsCharged < 0 || hasNegativeValue(
		receipt.InputTokens,
		receipt.OutputTokens,
		receipt.CachedTokens,
		receipt.ReasoningTokens,
	) {
		return UsageReceipt{}, fmt.Errorf("backend media analysis response has invalid usage values")
	}
	if receipt.EstimatedCost != nil && *receipt.EstimatedCost < 0 {
		return UsageReceipt{}, fmt.Errorf("backend media analysis response has an invalid estimated cost")
	}
	if receipt.Currency != nil {
		currency := strings.ToUpper(cleanText(*receipt.Currency, 3))
		if !currencyPattern.MatchString(currency) {
			return UsageReceipt{}, fmt.Errorf("backend media analysis response has an invalid currency")
		}
		receipt.Currency = &currency
	}
	if receipt.EstimatedCost != nil && receipt.Currency == nil {
		return UsageReceipt{}, fmt.Errorf("backend media analysis response is missing the estimated cost currency")
	}
	return receipt, nil
}

func hasNegativeValue(values ...*int64) bool {
	for _, value := range values {
		if value != nil && *value < 0 {
			return true
		}
	}
	return false
}

func mapItem(item analysisItem, sourcePath string, uploadFileName string) (models.ImageAnalysisResult, error) {
	entityType := models.EntityType(strings.ToLower(strings.TrimSpace(item.Type)))
	if !isSupportedEntityType(entityType) {
		return models.ImageAnalysisResult{}, fmt.Errorf("backend returned an unsupported inventory entity type")
	}
	if item.ExtractedData.Type() != entityType && entityType != models.EntityTypeUnclassified {
		return models.ImageAnalysisResult{}, fmt.Errorf("backend returned mismatched inventory analysis data")
	}

	confidence := 0.0
	flaggedForReview := item.FlaggedForReview
	reviewReason := cleanText(item.ReviewReason, 500)
	if item.Confidence == nil {
		flaggedForReview = true
		if reviewReason == "" {
			reviewReason = "Confidence was not reported by the analysis route."
		}
	} else {
		if *item.Confidence < 0 || *item.Confidence > 1 {
			return models.ImageAnalysisResult{}, fmt.Errorf("backend returned an invalid confidence value")
		}
		confidence = *item.Confidence
	}

	metadata := make(map[string]interface{})
	if timestamp := cleanText(item.Timestamp, 64); timestamp != "" {
		metadata["video_timestamp"] = timestamp
	}

	return models.ImageAnalysisResult{
		ImagePath:        sourcePath,
		OriginalFilename: uploadFileName,
		Classification: models.ClassificationResult{
			PrimaryType: entityType,
			Confidence:  confidence,
			Reasoning:   cleanText(item.Reasoning, 2000),
		},
		ExtractedData:    item.ExtractedData,
		EXIFMetadata:     metadata,
		FlaggedForReview: flaggedForReview,
		ReviewReason:     reviewReason,
		ProcessedAt:      time.Now().UTC(),
		RelatedTo:        cleanText(item.RelatedTo, 200),
	}, nil
}

func isSupportedEntityType(entityType models.EntityType) bool {
	switch entityType {
	case models.EntityTypeAsset, models.EntityTypeTool, models.EntityTypePart,
		models.EntityTypeChemical, models.EntityTypeUnclassified:
		return true
	default:
		return false
	}
}

func cleanText(value string, maxLength int) string {
	value = strings.Map(func(r rune) rune {
		if (unicode.IsControl(r) && r != '\n' && r != '\t') || unicode.Is(unicode.Cf, r) {
			return -1
		}
		return r
	}, strings.TrimSpace(value))
	runes := []rune(value)
	if len(runes) > maxLength {
		return string(runes[:maxLength])
	}
	return value
}

func safeUploadError(err error) error {
	var apiError *clierrors.APIError
	if errors.As(err, &apiError) {
		switch apiError.StatusCode {
		case 401:
			return fmt.Errorf("media analysis authentication expired; sign in again")
		case 403:
			return fmt.Errorf("media analysis is not permitted for this tenant or role")
		case 402, 429:
			return fmt.Errorf("media analysis is unavailable because the tenant budget or request limit was reached")
		default:
			return fmt.Errorf("media analysis request failed with HTTP %d", apiError.StatusCode)
		}
	}
	return err
}

func safeBusinessError(errorCode string, quotaBlockedReason string) error {
	if quotaBlockedReason != "" {
		return fmt.Errorf("media analysis is blocked by the tenant AI policy or budget")
	}
	switch strings.ToLower(strings.TrimSpace(errorCode)) {
	case "provider_unavailable", "route_unavailable", "byok_unavailable":
		return fmt.Errorf("no healthy tenant-eligible media analysis route is available")
	case "permission_denied":
		return fmt.Errorf("media analysis is not permitted for this tenant or role")
	default:
		return fmt.Errorf("media analysis could not be completed")
	}
}

func newRequestGUID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	bytes[6] = (bytes[6] & 0x0f) | 0x40
	bytes[8] = (bytes[8] & 0x3f) | 0x80
	encoded := hex.EncodeToString(bytes)
	return fmt.Sprintf("%s-%s-%s-%s-%s", encoded[0:8], encoded[8:12], encoded[12:16], encoded[16:20], encoded[20:32]), nil
}
