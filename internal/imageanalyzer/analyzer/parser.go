package analyzer

import (
	"encoding/json"
	"log"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/uteamup/cli/internal/imageanalyzer/models"
)

// parseMultiEntityResponse parses the multi-entity JSON response from Gemini.
// Handles both the new multi-entity format ({"entities": [...]}) and
// the legacy single-entity format ({"classification": ..., "extracted_data": ...}).
func parseMultiEntityResponse(responseText, imagePath string) []models.ImageAnalysisResult {
	originalFilename := filepath.Base(imagePath)

	parsed, err := tryParseJSON(responseText)
	if err != nil {
		log.Printf("analyzer: invalid_json image=%s preview=%.200s", originalFilename, responseText)
		// Return unclassified — JSON fix is handled at a higher level.
		return []models.ImageAnalysisResult{
			unclassifiedResult(imagePath, originalFilename, "Failed to parse Gemini response as JSON", "JSON parse failure"),
		}
	}

	// Determine format: multi-entity or legacy single-entity.
	var entityDicts []map[string]interface{}

	if entities, ok := parsed["entities"]; ok {
		if arr, ok := entities.([]interface{}); ok {
			for _, item := range arr {
				if m, ok := item.(map[string]interface{}); ok {
					entityDicts = append(entityDicts, m)
				}
			}
		}
	} else if _, ok := parsed["classification"]; ok {
		// Legacy single-entity format — wrap in list.
		entityDicts = []map[string]interface{}{parsed}
	} else {
		log.Printf("analyzer: unknown_response_format image=%s", originalFilename)
		return []models.ImageAnalysisResult{
			unclassifiedResult(imagePath, originalFilename, "Unknown response format from Gemini", "Unknown response format"),
		}
	}

	if len(entityDicts) == 0 {
		return []models.ImageAnalysisResult{
			unclassifiedResult(imagePath, originalFilename, "Empty entities array from Gemini", "No entities detected"),
		}
	}

	var results []models.ImageAnalysisResult
	for _, entityDict := range entityDicts {
		classification, extractedData := parseSingleEntity(entityDict, originalFilename)

		relatedTo := ""
		if rt, ok := entityDict["related_to"]; ok && rt != nil {
			if s, ok := rt.(string); ok {
				relatedTo = s
			}
		}

		results = append(results, models.ImageAnalysisResult{
			ImagePath:        imagePath,
			OriginalFilename: originalFilename,
			Classification:   classification,
			ExtractedData:    extractedData,
			RelatedTo:        relatedTo,
			ProcessedAt:      time.Now(),
		})
	}

	log.Printf("analyzer: entities_found image=%s count=%d", originalFilename, len(results))
	return results
}

// parseSingleEntity parses a single entity dict into classification + extracted data.
func parseSingleEntity(entityMap map[string]interface{}, originalFilename string) (models.ClassificationResult, models.ExtractedData) {
	var classification models.ClassificationResult
	var extracted models.ExtractedData

	// Extract classification.
	classData, ok := entityMap["classification"]
	if !ok {
		log.Printf("analyzer: missing classification image=%s", originalFilename)
		return models.ClassificationResult{
			PrimaryType: models.EntityTypeUnclassified,
			Confidence:  0.0,
			Reasoning:   "Missing classification in entity",
		}, extracted
	}

	classMap, ok := classData.(map[string]interface{})
	if !ok {
		log.Printf("analyzer: invalid classification format image=%s", originalFilename)
		return models.ClassificationResult{
			PrimaryType: models.EntityTypeUnclassified,
			Confidence:  0.0,
			Reasoning:   "Invalid classification format",
		}, extracted
	}

	// Parse classification fields.
	if pt, ok := classMap["primary_type"].(string); ok {
		classification.PrimaryType = models.EntityType(pt)
	} else {
		classification.PrimaryType = models.EntityTypeUnclassified
	}
	if conf, ok := classMap["confidence"].(float64); ok {
		classification.Confidence = conf
	}
	if reason, ok := classMap["reasoning"].(string); ok {
		classification.Reasoning = reason
	}

	// Extract entity-specific data.
	rawData, hasData := entityMap["extracted_data"]
	if !hasData || rawData == nil || classification.PrimaryType == models.EntityTypeUnclassified {
		return classification, extracted
	}

	// Re-marshal the raw data map to JSON, then unmarshal into the typed struct.
	dataBytes, err := json.Marshal(rawData)
	if err != nil {
		log.Printf("analyzer: extraction_marshal_error image=%s error=%v", originalFilename, err)
		return classification, extracted
	}

	switch classification.PrimaryType {
	case models.EntityTypeAsset:
		var asset models.ExtractedAssetData
		if err := json.Unmarshal(dataBytes, &asset); err != nil {
			log.Printf("analyzer: extraction_validation_error image=%s type=asset error=%v", originalFilename, err)
			return classification, extracted
		}
		extracted.Asset = &asset

	case models.EntityTypeTool:
		var tool models.ExtractedToolData
		if err := json.Unmarshal(dataBytes, &tool); err != nil {
			log.Printf("analyzer: extraction_validation_error image=%s type=tool error=%v", originalFilename, err)
			return classification, extracted
		}
		extracted.Tool = &tool

	case models.EntityTypePart:
		var part models.ExtractedPartData
		if err := json.Unmarshal(dataBytes, &part); err != nil {
			log.Printf("analyzer: extraction_validation_error image=%s type=part error=%v", originalFilename, err)
			return classification, extracted
		}
		extracted.Part = &part

	case models.EntityTypeChemical:
		var chemical models.ExtractedChemicalData
		if err := json.Unmarshal(dataBytes, &chemical); err != nil {
			log.Printf("analyzer: extraction_validation_error image=%s type=chemical error=%v", originalFilename, err)
			return classification, extracted
		}
		extracted.Chemical = &chemical
	}

	return classification, extracted
}

// tryParseJSON attempts to parse JSON from text, stripping markdown fences if present.
func tryParseJSON(text string) (map[string]interface{}, error) {
	cleaned := strings.TrimSpace(text)

	// Remove markdown JSON fences if present.
	re1 := regexp.MustCompile(`(?s)^` + "```json\\s*")
	cleaned = re1.ReplaceAllString(cleaned, "")

	re2 := regexp.MustCompile(`(?s)^` + "```\\s*")
	cleaned = re2.ReplaceAllString(cleaned, "")

	re3 := regexp.MustCompile(`(?s)\s*` + "```$")
	cleaned = re3.ReplaceAllString(cleaned, "")

	cleaned = strings.TrimSpace(cleaned)

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
		return nil, err
	}
	return result, nil
}

// unclassifiedResult creates an unclassified ImageAnalysisResult for error cases.
func unclassifiedResult(imagePath, originalFilename, reasoning, reviewReason string) models.ImageAnalysisResult {
	return models.ImageAnalysisResult{
		ImagePath:        imagePath,
		OriginalFilename: originalFilename,
		Classification: models.ClassificationResult{
			PrimaryType: models.EntityTypeUnclassified,
			Confidence:  0.0,
			Reasoning:   reasoning,
		},
		FlaggedForReview: true,
		ReviewReason:     reviewReason,
		ProcessedAt:      time.Now(),
	}
}
