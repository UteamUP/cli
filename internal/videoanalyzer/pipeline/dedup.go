// Package pipeline orchestrates the video analysis pipeline.
package pipeline

import (
	"strconv"
	"strings"

	"github.com/uteamup/cli/internal/imageanalyzer/models"
	"github.com/uteamup/cli/internal/videoanalyzer/analyzer"
)

// TemporalDedup merges entities from the same video that have similar names and
// appear within windowSec seconds of each other. This prevents the same physical
// item from appearing multiple times when it's visible at different timestamps.
func TemporalDedup(results []models.ImageAnalysisResult, windowSec int) []models.ImageAnalysisResult {
	if len(results) <= 1 {
		return results
	}

	// Group by video path + entity type.
	type groupKey struct {
		videoPath  string
		entityType models.EntityType
	}
	groups := make(map[groupKey][]models.ImageAnalysisResult)
	for _, r := range results {
		key := groupKey{
			videoPath:  r.ImagePath,
			entityType: r.Classification.PrimaryType,
		}
		groups[key] = append(groups[key], r)
	}

	var deduped []models.ImageAnalysisResult

	for _, items := range groups {
		if len(items) <= 1 {
			deduped = append(deduped, items...)
			continue
		}

		// Mark which items have been merged into another.
		merged := make([]bool, len(items))

		for i := 0; i < len(items); i++ {
			if merged[i] {
				continue
			}

			primary := items[i]
			primaryTS := parseTimestampSec(analyzer.GetTimestamp(&primary))
			primaryName := strings.ToLower(strings.TrimSpace(primary.ExtractedData.GetName()))

			for j := i + 1; j < len(items); j++ {
				if merged[j] {
					continue
				}

				candidate := items[j]
				candidateName := strings.ToLower(strings.TrimSpace(candidate.ExtractedData.GetName()))

				// Check name similarity.
				if primaryName == "" || candidateName == "" {
					continue
				}
				if !namesAreSimilar(primaryName, candidateName) {
					continue
				}

				// Check temporal proximity.
				candidateTS := parseTimestampSec(analyzer.GetTimestamp(&candidate))
				if primaryTS >= 0 && candidateTS >= 0 {
					diff := primaryTS - candidateTS
					if diff < 0 {
						diff = -diff
					}
					if diff > windowSec {
						continue
					}
				}

				// Merge: keep higher confidence, earlier timestamp.
				if candidate.Classification.Confidence > primary.Classification.Confidence {
					// Keep candidate's data but primary's timestamp if earlier.
					if primaryTS >= 0 && (candidateTS < 0 || primaryTS < candidateTS) {
						candidate.EXIFMetadata["video_timestamp"] = analyzer.GetTimestamp(&primary)
					}
					primary = candidate
				}
				merged[j] = true
			}

			deduped = append(deduped, primary)
		}
	}

	return deduped
}

// parseTimestampSec parses a "MM:SS" timestamp string into total seconds.
// Returns -1 if the timestamp is empty or malformed.
func parseTimestampSec(ts string) int {
	if ts == "" {
		return -1
	}
	parts := strings.SplitN(ts, ":", 2)
	if len(parts) != 2 {
		return -1
	}
	minutes, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return -1
	}
	seconds, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return -1
	}
	return minutes*60 + seconds
}

// namesAreSimilar checks if two normalized names are similar enough to be
// considered the same entity. Uses exact match after normalization.
// For fuzzy matching, the cross-video grouper handles it.
func namesAreSimilar(a, b string) bool {
	return a == b
}
