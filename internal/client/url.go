package client

import (
	"fmt"
	"net/url"
	"strings"
)

// ValidateBaseURL rejects ambiguous or insecure backend origins. Media files
// may contain sensitive tenant data, so they are only sent to a clean HTTPS
// origin selected in the active CLI profile.
func ValidateBaseURL(rawURL string) error {
	parsed, err := url.ParseRequestURI(strings.TrimSpace(rawURL))
	if err != nil {
		return fmt.Errorf("invalid API base URL")
	}
	if !strings.EqualFold(parsed.Scheme, "https") {
		return fmt.Errorf("API base URL must use HTTPS")
	}
	if parsed.Hostname() == "" {
		return fmt.Errorf("API base URL must include a hostname")
	}
	if parsed.User != nil {
		return fmt.Errorf("API base URL must not include credentials")
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return fmt.Errorf("API base URL must not include a query or fragment")
	}
	if parsed.EscapedPath() != "" && parsed.EscapedPath() != "/" {
		return fmt.Errorf("API base URL must be an origin without a path")
	}
	return nil
}
