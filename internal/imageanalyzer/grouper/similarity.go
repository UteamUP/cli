package grouper

import (
	"math/bits"
	"strconv"
	"strings"

	"github.com/uteamup/cli/internal/imageanalyzer/models"
)

// computeSimilarity returns a weighted multi-signal similarity score between
// two analysis results. Different entity types always return 0.0.
//
// Weights:
//
//	serial_number exact match       0.40
//	model_number exact match        0.20
//	name fuzzy match                0.20
//	description fuzzy match         0.10
//	perceptual hash similarity      0.05
//	manufacturer_brand match        0.05
func computeSimilarity(a, b models.ImageAnalysisResult) float64 {
	// Different entity types -> zero similarity.
	if a.Classification.PrimaryType != b.Classification.PrimaryType {
		return 0.0
	}

	score := 0.0

	// --- serial_number (0.40) ---
	snA := a.ExtractedData.GetSerialNumber()
	snB := b.ExtractedData.GetSerialNumber()
	if snA != "" && snB != "" && snA == snB {
		score += 0.40
	}

	// --- model_number (0.20) ---
	mnA := a.ExtractedData.GetModelNumber()
	mnB := b.ExtractedData.GetModelNumber()
	if mnA != "" && mnB != "" && mnA == mnB {
		score += 0.20
	}

	// --- name fuzzy match (0.20) ---
	nameA := a.ExtractedData.GetName()
	nameB := b.ExtractedData.GetName()
	if nameA != "" && nameB != "" {
		ratio := levenshteinRatio(strings.ToLower(nameA), strings.ToLower(nameB))
		score += 0.20 * ratio
	}

	// --- description fuzzy match (0.10) ---
	descA := a.ExtractedData.GetDescription()
	descB := b.ExtractedData.GetDescription()
	if descA != "" && descB != "" {
		ratio := levenshteinRatio(strings.ToLower(descA), strings.ToLower(descB))
		score += 0.10 * ratio
	}

	// --- perceptual hash similarity (0.05) ---
	if a.PerceptualHash != "" && b.PerceptualHash != "" {
		score += 0.05 * phashSimilarity(a.PerceptualHash, b.PerceptualHash)
	}

	// --- manufacturer_brand (0.05) ---
	brandA := a.ExtractedData.GetBrand()
	brandB := b.ExtractedData.GetBrand()
	if brandA != "" && brandB != "" && strings.EqualFold(brandA, brandB) {
		score += 0.05
	}

	return score
}

// levenshteinRatio returns the normalised similarity between two strings
// using Levenshtein edit distance: 1.0 - distance/max(len(a), len(b)).
// Empty strings: if both empty returns 1.0, if one empty returns 0.0.
func levenshteinRatio(a, b string) float64 {
	if a == b {
		return 1.0
	}
	la, lb := len(a), len(b)
	if la == 0 || lb == 0 {
		return 0.0
	}

	// Wagner-Fischer algorithm with single-row optimisation.
	if la < lb {
		a, b = b, a
		la, lb = lb, la
	}

	prev := make([]int, lb+1)
	for j := 0; j <= lb; j++ {
		prev[j] = j
	}

	for i := 1; i <= la; i++ {
		curr := make([]int, lb+1)
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			ins := curr[j-1] + 1
			del := prev[j] + 1
			sub := prev[j-1] + cost
			m := ins
			if del < m {
				m = del
			}
			if sub < m {
				m = sub
			}
			curr[j] = m
		}
		prev = curr
	}

	dist := prev[lb]
	maxLen := la // la >= lb after swap
	return 1.0 - float64(dist)/float64(maxLen)
}

// phashSimilarity compares two hex-encoded perceptual hashes via normalised
// Hamming distance. Returns 1.0 for identical hashes, approaching 0.0 as
// they diverge. Returns 0.0 on parse errors.
func phashSimilarity(hashA, hashB string) float64 {
	intA, errA := strconv.ParseUint(hashA, 16, 64)
	intB, errB := strconv.ParseUint(hashB, 16, 64)
	if errA != nil || errB != nil {
		return 0.0
	}

	xor := intA ^ intB
	hamming := bits.OnesCount64(xor)

	// maxBits: use 64 as the hash width (uint64).
	const maxBits = 64
	return 1.0 - float64(hamming)/float64(maxBits)
}
