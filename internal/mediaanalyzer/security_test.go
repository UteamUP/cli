package mediaanalyzer

import (
	"github.com/uteamup/cli/internal/imageanalyzer/models"
	"strings"
	"testing"
)

func TestMapItemRejectsUnboundedNestedExtractedNames(t *testing.T) {
	item := analysisItem{Type: "asset", ExtractedData: models.ExtractedData{Asset: &models.ExtractedAssetData{Name: strings.Repeat("a", 513)}}}
	if _, err := mapItem(item, "source", "photo"); err == nil {
		t.Fatal("unbounded nested name accepted")
	}
	item.ExtractedData.Asset.Name = "Pump"
	if _, err := mapItem(item, "source", "photo"); err != nil {
		t.Fatal(err)
	}
}
