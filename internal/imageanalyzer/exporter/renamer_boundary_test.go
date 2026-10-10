package exporter

import (
	"github.com/uteamup/cli/internal/imageanalyzer/models"
	"github.com/uteamup/cli/internal/security"
	"os"
	"path/filepath"
	"testing"
)

func TestRenamedSourceCannotChangeAfterAnalysisAndExistingOutputSurvives(t *testing.T) {
	for _, mutation := range []string{"replacement", "symlink", "oversized"} {
		t.Run(mutation, func(t *testing.T) {
			dir := t.TempDir()
			source := filepath.Join(dir, "image.jpg")
			dest := filepath.Join(dir, "copy.jpg")
			if err := os.WriteFile(source, []byte("original"), 0600); err != nil {
				t.Fatal(err)
			}
			info, _ := os.Stat(source)
			security.BindSource(source, info)
			if err := os.Remove(source); err != nil {
				t.Fatal(err)
			}
			if mutation == "symlink" {
				secret := filepath.Join(dir, "private")
				os.WriteFile(secret, []byte("private"), 0600)
				if err := os.Symlink(secret, source); err != nil {
					t.Skip(err)
				}
			} else {
				os.WriteFile(source, []byte("replacement"), 0600)
				if mutation == "oversized" {
					if err := os.Truncate(source, 16*1024*1024); err != nil {
						t.Fatal(err)
					}
				}
			}
			if copyFile(source, dest) == nil {
				t.Fatal("changed source copied")
			}
			if _, err := os.Stat(dest); !os.IsNotExist(err) {
				t.Fatal("invalid source produced output")
			}
		})
	}
	dir := t.TempDir()
	source := filepath.Join(dir, "source")
	dest := filepath.Join(dir, "dest")
	os.WriteFile(source, []byte("normal"), 0600)
	os.WriteFile(dest, []byte("keep"), 0600)
	info, _ := os.Stat(source)
	security.BindSource(source, info)
	if copyFile(source, dest) == nil {
		t.Fatal("existing destination replaced")
	}
	data, _ := os.ReadFile(dest)
	if string(data) != "keep" {
		t.Fatal("existing destination changed")
	}
	os.Remove(dest)
	if err := copyFile(source, dest); err != nil {
		t.Fatal(err)
	}
	data, _ = os.ReadFile(dest)
	if string(data) != "normal" {
		t.Fatal("normal source changed")
	}
}

func TestCheckpointCannotIntroduceUnscannedCopySource(t *testing.T) {
	source := filepath.Join(t.TempDir(), "private.jpg")
	dest := filepath.Join(t.TempDir(), "export.jpg")
	if err := os.WriteFile(source, []byte("private"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := copyFile(source, dest); err == nil {
		t.Fatal("unscanned checkpoint source copied")
	}
	if _, err := os.Stat(dest); !os.IsNotExist(err) {
		t.Fatal("unscanned input created output")
	}
}

func TestCheckpointClassificationCannotEscapeExportDirectory(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "renamed")
	os.Mkdir(output, 0700)
	source := filepath.Join(dir, "input.jpg")
	os.WriteFile(source, []byte("image"), 0600)
	info, _ := os.Stat(source)
	security.BindSource(source, info)
	group := makeAssetGroup("safe", 1, false)
	group.Primary.ImagePath = source
	group.Primary.Classification.PrimaryType = models.EntityType("../escaped")
	exporter := NewExporter(dir, output, true, "")
	if _, err := exporter.RenameImages([]models.ImageGroup{group}); err == nil {
		t.Fatal("checkpoint escaped generated output folder")
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 2 {
		t.Fatal("invalid classification created an output")
	}
	group.Primary.Classification.PrimaryType = models.EntityTypeAsset
	if _, err := exporter.RenameImages([]models.ImageGroup{group}); err != nil {
		t.Fatal(err)
	}
}
