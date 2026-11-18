package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/uteamup/cli/internal/client"
	clierrors "github.com/uteamup/cli/internal/errors"
	"github.com/uteamup/cli/internal/output"
	"github.com/uteamup/cli/internal/security"
)

// Journal image attach limits mirror the backend contract for
// POST /api/journal/{journalGuid}/images (FileUploadCategory.JournalImportImage):
// PNG, JPEG, GIF and WebP only, 25 MB per image, 100 MB per request.
const (
	journalImageMaxBytes        = 25 * 1024 * 1024
	journalImageMaxRequestBytes = 100 * 1024 * 1024
)

var journalImageExtensions = map[string]bool{
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".gif":  true,
	".webp": true,
}

var journalGUIDPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// newJournalAddImagesCmd builds `journal add-images <journalGuid> <files...>`.
// It is hand-built rather than registry-driven because it takes a variable
// number of local files and sends them as repeated `images` multipart parts.
func newJournalAddImagesCmd(clientFactory func() (*client.APIClient, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "add-images <journalGuid> <files...>",
		Short: "Attach one or more images (PNG/JPEG/GIF/WebP, max 25 MB each) to an existing journal entry",
		Long: `Attach one or more local images to an existing journal entry.

Images are sent in one request to POST /api/journal/{journalGuid}/images and
are SHA-256 deduplicated against the tenant's documents. Supported formats are
PNG, JPEG, GIF and WebP; SVG is rejected. Each image may be at most 25 MB and
all images together at most 100 MB. Needs Journal.Import and Document.Upload.

Examples:
  uteamup journal add-images 3f2a...-guid photo.jpg
  uteamup journal add-images 3f2a...-guid before.png after.png -o json`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			journalGUID, files, err := validateJournalAddImagesInput(args)
			if err != nil {
				return err
			}
			apiClient, err := clientFactory()
			if err != nil {
				return err
			}
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			result, err := apiClient.CallRESTUploadFiles(ctx, http.MethodPost, "/api/journal/"+journalGUID+"/images", "images", files, nil, nil, "add-images")
			if err != nil {
				return journalAddImagesError(journalGUID, err)
			}
			return output.Print(output.ParseFormat(outputFormat), result)
		},
	}
}

func validateJournalAddImagesInput(args []string) (string, []string, error) {
	if len(args) < 2 {
		return "", nil, fmt.Errorf("add-images needs a journal GUID and at least one image file")
	}
	journalGUID := strings.TrimSpace(args[0])
	if !journalGUIDPattern.MatchString(journalGUID) || journalGUID == "00000000-0000-0000-0000-000000000000" {
		return "", nil, fmt.Errorf("invalid journalGuid: must be a non-empty GUID")
	}
	var total int64
	files := make([]string, 0, len(args)-1)
	seen := map[string]bool{}
	for _, file := range args[1:] {
		name := filepath.Base(file)
		extension := strings.ToLower(filepath.Ext(file))
		if extension == ".svg" {
			return "", nil, fmt.Errorf("%s: SVG images are not supported; use PNG, JPEG, GIF or WebP", security.SafeText(name))
		}
		if !journalImageExtensions[extension] {
			return "", nil, fmt.Errorf("%s: unsupported image type %q; use PNG, JPEG, GIF or WebP", security.SafeText(name), security.SafeText(extension))
		}
		info, err := os.Stat(file)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return "", nil, fmt.Errorf("%s: file not found", security.SafeText(file))
			}
			return "", nil, fmt.Errorf("%s: %w", security.SafeText(file), err)
		}
		if !info.Mode().IsRegular() {
			return "", nil, fmt.Errorf("%s: not a regular file", security.SafeText(file))
		}
		if info.Size() == 0 {
			return "", nil, fmt.Errorf("%s: file is empty", security.SafeText(name))
		}
		if info.Size() > journalImageMaxBytes {
			return "", nil, fmt.Errorf("%s: image exceeds the 25 MB limit", security.SafeText(name))
		}
		absolute, err := filepath.Abs(file)
		if err == nil {
			if seen[absolute] {
				return "", nil, fmt.Errorf("%s: listed more than once", security.SafeText(file))
			}
			seen[absolute] = true
		}
		total += info.Size()
		if total > journalImageMaxRequestBytes {
			return "", nil, fmt.Errorf("images exceed the 100 MB per-request limit; split them into several add-images calls")
		}
		files = append(files, file)
	}
	return journalGUID, files, nil
}

// journalAddImagesError turns the backend's 404/415 replies into clear messages.
func journalAddImagesError(journalGUID string, err error) error {
	var apiErr *clierrors.APIError
	if !errors.As(err, &apiErr) {
		return err
	}
	switch apiErr.StatusCode {
	case http.StatusNotFound:
		return fmt.Errorf("journal entry %s not found in the active tenant (or it is archived): %w", journalGUID, err)
	case http.StatusUnsupportedMediaType:
		code := ""
		var body struct {
			Error  string `json:"error"`
			Detail string `json:"detail"`
		}
		if json.Unmarshal([]byte(apiErr.Body), &body) == nil {
			code = body.Error
		}
		if code == "svg_not_supported" {
			return fmt.Errorf("image rejected: SVG is not supported; use PNG, JPEG, GIF or WebP: %w", err)
		}
		return fmt.Errorf("image rejected by the server (content does not match a PNG, JPEG, GIF or WebP image, or exceeds 25 MB): %w", err)
	case http.StatusForbidden:
		return fmt.Errorf("permission denied: attaching images needs Journal.Import and Document.Upload: %w", err)
	}
	return err
}
