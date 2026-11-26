package logging

import (
	"strings"
	"testing"
)

func TestCredentialRedactionDoesNotStopAtHostnameOrFirstOccurrence(t *testing.T) {
	input := `GET https://token-api.example.invalid/api/salesbooking?token=first-canary&token=second-canary tenant=tenant-a Bearer third-canary token=fourth-canary password="fifth canary"`
	got := redact(input)
	for _, canary := range []string{"first-canary", "second-canary", "third-canary", "fourth-canary", "fifth canary"} {
		if strings.Contains(got, canary) {
			t.Fatalf("credential survived: %s", got)
		}
	}
	if !strings.Contains(got, "token-api.example.invalid/api/salesbooking") || !strings.Contains(got, "tenant=tenant-a") {
		t.Fatal("ordinary diagnostic context lost")
	}
}
func TestTransportErrorRedactsEncodedQueryAndFragmentCapabilities(t *testing.T) {
	got := SafeDiagnostic(`request failed: Get "https://example.invalid/path?%74oken=canary#second-canary": refused`)
	if strings.Contains(got, "canary") || !strings.Contains(got, "refused") {
		t.Fatal(got)
	}
}

func TestDeviceTokenRouteRedactsTheTokenFromRequestLogs(t *testing.T) {
	got := SafeDiagnostic(`DELETE https://api.uteamup.com/api/devicetokens/fcm-token-canary tenant=tenant-a`)
	if strings.Contains(got, "fcm-token-canary") {
		t.Fatalf("device token survived route redaction: %s", got)
	}
	if !strings.Contains(got, "/api/devicetokens/[REDACTED]") || !strings.Contains(got, "tenant=tenant-a") {
		t.Fatalf("ordinary request context was not preserved: %s", got)
	}
}

func TestMcpCapabilityRoutesRedactTokensFromDiagnostics(t *testing.T) {
	input := `PUT https://api.uteamup.com/api/mcp/upload-tickets/secret-upload-canary GET https://api.uteamup.com/api/mcp/download-links/secret-download-canary`
	got := SafeDiagnostic(input)
	for _, canary := range []string{"secret-upload-canary", "secret-download-canary"} {
		if strings.Contains(got, canary) {
			t.Fatalf("MCP capability token survived redaction: %s", got)
		}
	}
	if !strings.Contains(got, "/api/mcp/upload-tickets/[REDACTED]") || !strings.Contains(got, "/api/mcp/download-links/[REDACTED]") {
		t.Fatalf("redacted MCP capability route markers missing: %s", got)
	}
}
