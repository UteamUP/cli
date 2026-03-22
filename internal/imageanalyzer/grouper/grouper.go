// Package grouper clusters multiple photos of the same physical item.
//
// The algorithm proceeds in six stages:
//  1. Partition results by entity_type (unclassified items are never grouped).
//  2. Pre-merge iPhone IMG_E* edit pairs (using the PairedImages field).
//  3. Exact-match on serial_number (strongest signal).
//  4. Exact-match on name (case-insensitive, stripped).
//  5. Pairwise similarity + single-linkage agglomerative clustering.
//  6. Select a representative (highest confidence) and merge extracted data.
package grouper

import (
	"strings"

	"github.com/uteamup/cli/internal/imageanalyzer/models"
)

// ImageGrouper clusters analysis results that depict the same physical item.
type ImageGrouper struct {
	similarityThreshold float64
}

// NewGrouper creates an ImageGrouper with the given similarity threshold.
func NewGrouper(threshold float64) *ImageGrouper {
	return &ImageGrouper{similarityThreshold: threshold}
}

// GroupImages groups a list of analysis results into ImageGroups.
func (g *ImageGrouper) GroupImages(results []models.ImageAnalysisResult) []models.ImageGroup {
	// Step 1 — partition by entity type.
	partitions := make(map[models.EntityType][]models.ImageAnalysisResult)
	var unclassified []models.ImageGroup

	for _, r := range results {
		etype := r.Classification.PrimaryType
		if etype == models.EntityTypeUnclassified {
			unclassified = append(unclassified, models.ImageGroup{
				Primary:        r,
				Members:        nil,
				GroupConfidence: r.Classification.Confidence,
			})
		} else {
			partitions[etype] = append(partitions[etype], r)
		}
	}

	groups := make([]models.ImageGroup, len(unclassified))
	copy(groups, unclassified)

	for _, items := range partitions {
		groups = append(groups, g.clusterPartition(items)...)
	}

	return groups
}

// clusterPartition clusters a single entity-type partition into groups.
func (g *ImageGrouper) clusterPartition(items []models.ImageAnalysisResult) []models.ImageGroup {
	// Step 2 — pre-merge IMG_E* pairs.
	merged := premergePairs(items)

	// Step 3 — exact serial_number match.
	serialGroups, remaining := groupBySerial(merged)

	// Step 4 — exact name match (case-insensitive, stripped).
	nameGroups, remaining := groupByName(remaining)

	// Step 5 — agglomerative clustering on remaining.
	clustered := g.agglomerativeCluster(remaining)

	allGroups := make([][]models.ImageAnalysisResult, 0, len(serialGroups)+len(nameGroups)+len(clustered))
	allGroups = append(allGroups, serialGroups...)
	allGroups = append(allGroups, nameGroups...)
	allGroups = append(allGroups, clustered...)

	// Step 6 — select representative and merge data.
	final := make([]models.ImageGroup, 0, len(allGroups))
	for _, cluster := range allGroups {
		rep := selectRepresentative(cluster)
		var members []models.ImageAnalysisResult
		for _, r := range cluster {
			if r.ImagePath != rep.ImagePath {
				members = append(members, r)
			}
		}
		mergeExtractedData(&rep, members)
		final = append(final, models.ImageGroup{
			Primary:        rep,
			Members:        members,
			GroupConfidence: rep.Classification.Confidence,
		})
	}

	return final
}

// premergePairs merges items whose PairedImages reference another item in the
// list. The item with paired_images absorbs its partner; the partner is
// removed from the returned list.
func premergePairs(items []models.ImageAnalysisResult) []models.ImageAnalysisResult {
	pathToItem := make(map[string]struct{}, len(items))
	for _, r := range items {
		pathToItem[r.ImagePath] = struct{}{}
	}

	consumed := make(map[string]struct{})
	var merged []models.ImageAnalysisResult

	for _, r := range items {
		if _, ok := consumed[r.ImagePath]; ok {
			continue
		}
		for _, pairedPath := range r.PairedImages {
			if _, exists := pathToItem[pairedPath]; exists {
				consumed[pairedPath] = struct{}{}
			}
		}
		merged = append(merged, r)
	}

	return merged
}

// groupBySerial groups items that share the same non-empty serial number.
// Returns the formed groups and the remaining ungrouped items.
func groupBySerial(items []models.ImageAnalysisResult) (groups [][]models.ImageAnalysisResult, remaining []models.ImageAnalysisResult) {
	serialMap := make(map[string][]models.ImageAnalysisResult)
	var noSerial []models.ImageAnalysisResult

	for _, r := range items {
		sn := r.ExtractedData.GetSerialNumber()
		if sn != "" {
			serialMap[sn] = append(serialMap[sn], r)
		} else {
			noSerial = append(noSerial, r)
		}
	}

	for _, v := range serialMap {
		groups = append(groups, v)
	}
	return groups, noSerial
}

// groupByName groups items that share the exact same name (case-insensitive,
// trimmed). Single-item "groups" are returned as remaining.
func groupByName(items []models.ImageAnalysisResult) (groups [][]models.ImageAnalysisResult, remaining []models.ImageAnalysisResult) {
	nameMap := make(map[string][]models.ImageAnalysisResult)
	var noName []models.ImageAnalysisResult

	for _, r := range items {
		name := r.ExtractedData.GetName()
		if name != "" {
			key := strings.ToLower(strings.TrimSpace(name))
			if key != "" {
				nameMap[key] = append(nameMap[key], r)
			} else {
				noName = append(noName, r)
			}
		} else {
			noName = append(noName, r)
		}
	}

	remaining = append(remaining, noName...)
	for _, members := range nameMap {
		if len(members) > 1 {
			groups = append(groups, members)
		} else {
			remaining = append(remaining, members[0])
		}
	}

	return groups, remaining
}

// agglomerativeCluster performs single-linkage agglomerative clustering based
// on pairwise similarity.
func (g *ImageGrouper) agglomerativeCluster(items []models.ImageAnalysisResult) [][]models.ImageAnalysisResult {
	if len(items) == 0 {
		return nil
	}

	// Start with each item in its own cluster.
	clusters := make([][]models.ImageAnalysisResult, len(items))
	for i, r := range items {
		clusters[i] = []models.ImageAnalysisResult{r}
	}

	changed := true
	for changed {
		changed = false
		i := 0
		for i < len(clusters) {
			j := i + 1
			for j < len(clusters) {
				if g.clustersShouldMerge(clusters[i], clusters[j]) {
					clusters[i] = append(clusters[i], clusters[j]...)
					clusters = append(clusters[:j], clusters[j+1:]...)
					changed = true
				} else {
					j++
				}
			}
			i++
		}
	}

	return clusters
}

// clustersShouldMerge returns true if any pair across the two clusters
// exceeds the similarity threshold (single-linkage criterion).
func (g *ImageGrouper) clustersShouldMerge(a, b []models.ImageAnalysisResult) bool {
	for _, ra := range a {
		for _, rb := range b {
			if computeSimilarity(ra, rb) >= g.similarityThreshold {
				return true
			}
		}
	}
	return false
}
