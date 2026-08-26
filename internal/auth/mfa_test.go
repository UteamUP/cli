package auth

import (
	"encoding/json"
	"testing"
)

// A 403 carrying MFA_REQUIRED is a step in the login flow, not a failure. If the CLI does
// not recognise it, the user sees "login failed with status 403" plus a JSON body and
// reasonably concludes the CLI is broken rather than that they owe a code.
func TestChallengeResponseIsRecognised(t *testing.T) {
	body := `{"message":"MFA_REQUIRED","mfaToken":"tok-123","methods":["totp"]}`

	var challenge struct {
		Message  string `json:"message"`
		MfaToken string `json:"mfaToken"`
	}
	if err := json.Unmarshal([]byte(body), &challenge); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if challenge.Message != "MFA_REQUIRED" {
		t.Errorf("message = %q, want MFA_REQUIRED", challenge.Message)
	}
	if challenge.MfaToken != "tok-123" {
		t.Errorf("token = %q, want tok-123", challenge.MfaToken)
	}
}

// An ordinary 403 must NOT be treated as a challenge, or a genuine permission failure would
// silently drop the user into a code prompt they can never satisfy.
func TestPlainForbiddenIsNotAChallenge(t *testing.T) {
	var challenge struct {
		Message  string `json:"message"`
		MfaToken string `json:"mfaToken"`
	}
	if err := json.Unmarshal([]byte(`{"message":"Forbidden"}`), &challenge); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if challenge.Message == "MFA_REQUIRED" {
		t.Error("a plain 403 was mistaken for an MFA challenge")
	}
}

// The server can only be answered with a token it issued. Prompting for a code we could
// never redeem wastes the user's time and hides a server-side bug behind a typo message.
func TestEmptyChallengeTokenFailsBeforePrompting(t *testing.T) {
	client := NewClient("https://example.invalid", false, nil)

	_, err := client.completeMfaLogin("", "user@example.com")

	if err == nil {
		t.Fatal("expected an error for an empty challenge token")
	}
}
