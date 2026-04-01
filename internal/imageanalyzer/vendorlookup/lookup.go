// Package vendorlookup provides online vendor enrichment using Gemini AI
// to look up company details (description, email, website, phone) for
// detected vendors.
package vendorlookup

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"

	"github.com/uteamup/cli/internal/imageanalyzer/models"
)

// VendorLookup enriches vendor records with online information via Gemini.
type VendorLookup struct {
	model  *genai.GenerativeModel
	client *genai.Client
}

// NewVendorLookup creates a new VendorLookup using the given Gemini API key and model name.
func NewVendorLookup(apiKey, modelName string) (*VendorLookup, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("creating genai client for vendor lookup: %w", err)
	}

	model := client.GenerativeModel(modelName)
	temp := float32(0.2)
	model.Temperature = &temp
	maxTokens := int32(512)
	model.MaxOutputTokens = &maxTokens

	return &VendorLookup{
		model:  model,
		client: client,
	}, nil
}

// vendorLookupResponse is the expected JSON response from Gemini.
type vendorLookupResponse struct {
	Description *string `json:"description"`
	Email       *string `json:"email"`
	Website     *string `json:"website"`
	PhoneNumber *string `json:"phone_number"`
}

// EnrichVendor sends a prompt to Gemini to look up vendor details and fills
// in empty fields on the vendor struct.
func (v *VendorLookup) EnrichVendor(ctx context.Context, vendor *models.DetectedVendor) error {
	prompt := fmt.Sprintf(`You are looking up information about a company or manufacturer.
Company name: "%s"

Return ONLY a JSON object with these fields (set to null if unknown):
{
  "description": "Brief description of what the company does/manufactures",
  "email": "Contact email if known",
  "website": "Company website URL",
  "phone_number": "Contact phone number if known"
}
Return ONLY valid JSON, no markdown fences or extra text.`, vendor.Name)

	resp, err := v.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return fmt.Errorf("gemini vendor lookup for %q: %w", vendor.Name, err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return fmt.Errorf("empty response for vendor %q", vendor.Name)
	}

	// Extract text from response.
	part := resp.Candidates[0].Content.Parts[0]
	text, ok := part.(genai.Text)
	if !ok {
		return fmt.Errorf("unexpected response type for vendor %q", vendor.Name)
	}

	// Clean markdown fences if present despite instructions.
	jsonStr := strings.TrimSpace(string(text))
	jsonStr = strings.TrimPrefix(jsonStr, "```json")
	jsonStr = strings.TrimPrefix(jsonStr, "```")
	jsonStr = strings.TrimSuffix(jsonStr, "```")
	jsonStr = strings.TrimSpace(jsonStr)

	var result vendorLookupResponse
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return fmt.Errorf("parsing vendor lookup response for %q: %w", vendor.Name, err)
	}

	// Fill in empty fields only.
	if vendor.Description == "" && result.Description != nil {
		vendor.Description = *result.Description
	}
	if vendor.Email == "" && result.Email != nil {
		vendor.Email = *result.Email
	}
	if vendor.Website == "" && result.Website != nil {
		vendor.Website = *result.Website
	}
	if vendor.PhoneNumber == "" && result.PhoneNumber != nil {
		vendor.PhoneNumber = *result.PhoneNumber
	}

	return nil
}

// EnrichBatch enriches each vendor in the slice, skipping vendors that
// already have both a website and description. Rate-limits to 1 request
// per second to avoid API throttling.
func (v *VendorLookup) EnrichBatch(ctx context.Context, vendors []models.DetectedVendor) {
	for i := range vendors {
		// Skip if already enriched.
		if vendors[i].Website != "" && vendors[i].Description != "" {
			continue
		}

		if err := v.EnrichVendor(ctx, &vendors[i]); err != nil {
			log.Printf("Vendor lookup failed for %q: %v", vendors[i].Name, err)
		}

		// Rate limit: 1 per second.
		time.Sleep(time.Second)
	}
}

// Close releases the underlying Gemini client resources.
func (v *VendorLookup) Close() {
	if v.client != nil {
		v.client.Close()
	}
}
