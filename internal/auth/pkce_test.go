package auth

import (
	"crypto/sha256"
	"encoding/base64"
	"testing"
)

func TestGenerateCodeVerifier(t *testing.T) {
	verifier, err := GenerateCodeVerifier()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Base64url encoded 64 bytes = 86 characters
	if len(verifier) == 0 {
		t.Fatal("verifier should not be empty")
	}
	if len(verifier) < 43 || len(verifier) > 128 {
		t.Errorf("verifier length %d outside PKCE range [43,128]", len(verifier))
	}
}

func TestGenerateCodeVerifierUniqueness(t *testing.T) {
	v1, _ := GenerateCodeVerifier()
	v2, _ := GenerateCodeVerifier()
	if v1 == v2 {
		t.Error("two generated verifiers should not be identical")
	}
}

func TestCodeChallenge(t *testing.T) {
	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"

	challenge := CodeChallenge(verifier)
	if challenge == "" {
		t.Fatal("challenge should not be empty")
	}

	// Verify it's a valid S256 challenge
	hash := sha256.Sum256([]byte(verifier))
	expected := base64.RawURLEncoding.EncodeToString(hash[:])
	if challenge != expected {
		t.Errorf("challenge mismatch: got %s, want %s", challenge, expected)
	}
}

func TestCodeChallengeDeterministic(t *testing.T) {
	verifier := "test-verifier-value"
	c1 := CodeChallenge(verifier)
	c2 := CodeChallenge(verifier)
	if c1 != c2 {
		t.Error("same verifier should produce same challenge")
	}
}
