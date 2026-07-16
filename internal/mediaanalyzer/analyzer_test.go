package mediaanalyzer

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	clierrors "github.com/uteamup/cli/internal/errors"
	"github.com/uteamup/cli/internal/imageanalyzer/models"
)

type fakeUploadClient struct {
	response   json.RawMessage
	err        error
	endpoint   string
	field      string
	filePath   string
	fileName   string
	mime       string
	maxBytes   int64
	headers    map[string]string
	fileExists bool
}

func (fake *fakeUploadClient) CallRESTUploadLimited(
	_ context.Context,
	_ string,
	path string,
	fileField string,
	filePath string,
	uploadFileName string,
	contentType string,
	maxFileBytes int64,
	extraHeaders map[string]string,
) (json.RawMessage, error) {
	fake.endpoint = path
	fake.field = fileField
	fake.filePath = filePath
	fake.fileName = uploadFileName
	fake.mime = contentType
	fake.maxBytes = maxFileBytes
	fake.headers = extraHeaders
	_, err := os.Stat(filePath)
	fake.fileExists = err == nil
	return fake.response, fake.err
}

func TestAnalyzePhotoUsesServerOwnedRouteAndMapsReceipt(t *testing.T) {
	fake := &fakeUploadClient{response: json.RawMessage(`{
  "items": [{
    "type": "asset",
    "confidence": 0.91,
    "reasoning": "Visible equipment nameplate",
    "flaggedForReview": false,
    "reviewReason": "",
    "relatedTo": "",
    "timestamp": "",
    "extractedData": {
      "asset": {
        "name": "Hydraulic pump",
        "serial_number": "SN-100",
        "suggested_vendor": "Acme"
      }
    }
  }],
  "usageReceipt": {
    "requestGuid": "b966b8c7-04a4-45d4-aa51-519ecf2ef13a",
    "credentialSource": "TenantByok",
    "providerAlias": "Tenant vision",
    "modelAlias": "Vision Pro",
    "fallbackUsed": false,
    "creditsCharged": 0
  },
  "warnings": []
}`)}

	analyzer := newWithClient(fake)
	result, err := analyzer.AnalyzePhoto(context.Background(), "/private/site/pump.png", []byte("jpeg-data"))
	if err != nil {
		t.Fatal(err)
	}
	if fake.endpoint != PhotoEndpoint || fake.field != "file" {
		t.Fatalf("unexpected upload route: %s field=%s", fake.endpoint, fake.field)
	}
	if fake.fileName != "pump.png" || fake.mime != "image/jpeg" || fake.maxBytes != MaxPhotoBytes {
		t.Fatalf("unexpected upload metadata: filename=%q mime=%q max=%d", fake.fileName, fake.mime, fake.maxBytes)
	}
	if !fake.fileExists {
		t.Fatal("temporary upload did not exist during the request")
	}
	if _, err := os.Stat(fake.filePath); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("temporary upload was not removed after the request")
	}
	if !guidPattern.MatchString(fake.headers["Idempotency-Key"]) {
		t.Fatalf("invalid idempotency GUID: %q", fake.headers["Idempotency-Key"])
	}
	if result.Receipt.CredentialSource != "TenantByok" || result.Receipt.ModelAlias != "Vision Pro" {
		t.Fatalf("unexpected receipt: %+v", result.Receipt)
	}
	if len(result.Items) != 1 || result.Items[0].Classification.PrimaryType != models.EntityTypeAsset {
		t.Fatalf("unexpected mapped items: %+v", result.Items)
	}
	if result.Items[0].ExtractedData.Asset == nil || result.Items[0].ExtractedData.Asset.Name != "Hydraulic pump" {
		t.Fatalf("asset data was not mapped: %+v", result.Items[0].ExtractedData)
	}
}

func TestAnalyzePhotoFlagsMissingConfidenceWithoutDroppingClassification(t *testing.T) {
	fake := &fakeUploadClient{response: json.RawMessage(`{
  "items": [{
    "type": "part",
    "extractedData": {"part": {"name": "Bearing"}}
  }],
  "usageReceipt": {
    "requestGuid": "b966b8c7-04a4-45d4-aa51-519ecf2ef13a",
    "credentialSource": "Managed",
    "providerAlias": "UteamUP managed AI",
    "modelAlias": "Inventory vision"
  }
}`)}

	result, err := newWithClient(fake).AnalyzePhoto(context.Background(), "bearing.jpg", []byte("jpeg"))
	if err != nil {
		t.Fatal(err)
	}
	item := result.Items[0]
	if item.Classification.PrimaryType != models.EntityTypePart || !item.FlaggedForReview {
		t.Fatalf("missing confidence should preserve type and require review: %+v", item)
	}
}

func TestAnalyzeVideoRejectsUnsupportedMimeWithoutUpload(t *testing.T) {
	fake := &fakeUploadClient{}
	_, err := newWithClient(fake).AnalyzeVideo(context.Background(), "video.avi", "video/x-msvideo")
	if err == nil || fake.endpoint != "" {
		t.Fatalf("expected pre-upload MIME rejection, got err=%v endpoint=%q", err, fake.endpoint)
	}
}

func TestAnalyzeSanitizesBackendErrorBody(t *testing.T) {
	fake := &fakeUploadClient{err: clierrors.NewAPIError(500, "Internal Server Error", "provider-key=secret raw-output")}
	_, err := newWithClient(fake).AnalyzePhoto(context.Background(), "pump.jpg", []byte("jpeg"))
	if err == nil {
		t.Fatal("expected upload error")
	}
	if strings.Contains(err.Error(), "secret") || strings.Contains(err.Error(), "raw-output") {
		t.Fatalf("backend details leaked: %v", err)
	}
}

func TestAnalyzeSanitizesReceiptAliasesAndRejectsNegativeUsage(t *testing.T) {
	fake := &fakeUploadClient{response: json.RawMessage(`{
  "items": [],
  "usageReceipt": {
    "requestGuid": "b966b8c7-04a4-45d4-aa51-519ecf2ef13a",
    "credentialSource": "managed",
    "providerAlias": "Managed\u001b[31m",
    "modelAlias": "Vision\u202ealias",
    "inputTokens": 12,
    "creditsCharged": 1
  }
}`)}

	result, err := newWithClient(fake).AnalyzePhoto(context.Background(), "pump.jpg", []byte("jpeg"))
	if err != nil {
		t.Fatal(err)
	}
	if result.Receipt.CredentialSource != "Managed" ||
		strings.ContainsRune(result.Receipt.ProviderAlias, '\x1b') ||
		strings.ContainsRune(result.Receipt.ModelAlias, '\u202e') {
		t.Fatalf("receipt aliases were not normalized: %+v", result.Receipt)
	}

	fake.response = json.RawMessage(`{
  "items": [],
  "usageReceipt": {
    "requestGuid": "b966b8c7-04a4-45d4-aa51-519ecf2ef13a",
    "credentialSource": "Managed",
    "providerAlias": "Managed",
    "modelAlias": "Vision",
    "outputTokens": -1
  }
}`)
	if _, err := newWithClient(fake).AnalyzePhoto(context.Background(), "pump.jpg", []byte("jpeg")); err == nil {
		t.Fatal("expected negative usage to be rejected")
	}
}
