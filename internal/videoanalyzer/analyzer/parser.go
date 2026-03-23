package analyzer

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/uteamup/cli/internal/imageanalyzer/models"
)

// videoResponse is the top-level JSON structure returned by Gemini for video analysis.
type videoResponse struct {
	Entities []videoEntity `json:"entities"`
}

// videoEntity represents a single entity detected in the video.
type videoEntity struct {
	Type            string           `json:"type"`
	Timestamp       string           `json:"timestamp"`
	Confidence      float64          `json:"confidence"`
	Reasoning       string           `json:"reasoning"`
	FlaggedForReview bool            `json:"flagged_for_review"`
	ReviewReason    *string          `json:"review_reason"`
	RelatedTo       *string          `json:"related_to"`
	ExtractedData   *json.RawMessage `json:"extracted_data"`
}

// ParseVideoResponse parses the JSON response from Gemini into ImageAnalysisResult structs.
// It reuses the existing models from the imageanalyzer package for compatibility with
// the CSV exporter, grouper, and other shared components.
func ParseVideoResponse(jsonText string, videoPath string, originalFilename string) ([]models.ImageAnalysisResult, error) {
	// Strip markdown fences if present (belt-and-suspenders even with ResponseMIMEType).
	cleaned := stripMarkdownFences(jsonText)

	var resp videoResponse
	if err := json.Unmarshal([]byte(cleaned), &resp); err != nil {
		return nil, fmt.Errorf("parsing video analysis JSON: %w", err)
	}

	if len(resp.Entities) == 0 {
		return nil, nil
	}

	results := make([]models.ImageAnalysisResult, 0, len(resp.Entities))
	now := time.Now()

	for _, entity := range resp.Entities {
		result := models.ImageAnalysisResult{
			ImagePath:        videoPath,
			OriginalFilename: originalFilename,
			Classification: models.ClassificationResult{
				PrimaryType: mapEntityType(entity.Type),
				Confidence:  entity.Confidence,
				Reasoning:   entity.Reasoning,
			},
			FlaggedForReview: entity.FlaggedForReview,
			ProcessedAt:      now,
		}

		if entity.ReviewReason != nil {
			result.ReviewReason = *entity.ReviewReason
		}
		if entity.RelatedTo != nil {
			result.RelatedTo = *entity.RelatedTo
		}

		// Store timestamp in EXIF metadata map for downstream use.
		result.EXIFMetadata = map[string]interface{}{
			"video_timestamp": entity.Timestamp,
		}

		// Parse extracted data based on entity type.
		if entity.ExtractedData != nil {
			extractedData, err := parseExtractedData(entity.Type, *entity.ExtractedData)
			if err != nil {
				// Non-fatal: flag for review instead of failing.
				result.FlaggedForReview = true
				result.ReviewReason = fmt.Sprintf("failed to parse extracted data: %v", err)
			} else {
				result.ExtractedData = extractedData
			}
		}

		results = append(results, result)
	}

	return results, nil
}

// mapEntityType converts a string entity type to the models.EntityType constant.
func mapEntityType(t string) models.EntityType {
	switch strings.ToLower(strings.TrimSpace(t)) {
	case "asset":
		return models.EntityTypeAsset
	case "tool":
		return models.EntityTypeTool
	case "part":
		return models.EntityTypePart
	case "chemical":
		return models.EntityTypeChemical
	default:
		return models.EntityTypeUnclassified
	}
}

// parseExtractedData unmarshals entity-specific data into the correct ExtractedData wrapper.
func parseExtractedData(entityType string, raw json.RawMessage) (models.ExtractedData, error) {
	var data models.ExtractedData

	switch strings.ToLower(strings.TrimSpace(entityType)) {
	case "asset":
		var asset models.ExtractedAssetData
		if err := json.Unmarshal(raw, &asset); err != nil {
			return data, fmt.Errorf("unmarshaling asset data: %w", err)
		}
		data.Asset = &asset
	case "tool":
		var tool models.ExtractedToolData
		if err := json.Unmarshal(raw, &tool); err != nil {
			return data, fmt.Errorf("unmarshaling tool data: %w", err)
		}
		data.Tool = &tool
	case "part":
		var part models.ExtractedPartData
		if err := json.Unmarshal(raw, &part); err != nil {
			return data, fmt.Errorf("unmarshaling part data: %w", err)
		}
		data.Part = &part
	case "chemical":
		var chem models.ExtractedChemicalData
		if err := json.Unmarshal(raw, &chem); err != nil {
			return data, fmt.Errorf("unmarshaling chemical data: %w", err)
		}
		data.Chemical = &chem
	case "unclassified":
		// No extracted data for unclassified entities.
	default:
		return data, fmt.Errorf("unknown entity type: %s", entityType)
	}

	return data, nil
}

// stripMarkdownFences removes markdown code fences from JSON text.
func stripMarkdownFences(text string) string {
	text = strings.TrimSpace(text)

	// Remove ```json ... ``` wrapper.
	if strings.HasPrefix(text, "```") {
		// Find end of first line (the opening fence).
		idx := strings.Index(text, "\n")
		if idx >= 0 {
			text = text[idx+1:]
		}
		// Remove closing fence.
		if lastIdx := strings.LastIndex(text, "```"); lastIdx >= 0 {
			text = text[:lastIdx]
		}
		text = strings.TrimSpace(text)
	}

	return text
}

// GetTimestamp extracts the video timestamp from an ImageAnalysisResult's EXIF metadata.
func GetTimestamp(result *models.ImageAnalysisResult) string {
	if result.EXIFMetadata == nil {
		return ""
	}
	if ts, ok := result.EXIFMetadata["video_timestamp"]; ok {
		if s, ok := ts.(string); ok {
			return s
		}
	}
	return ""
}
