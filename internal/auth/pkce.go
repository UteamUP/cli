package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

const codeVerifierLength = 64

// GenerateCodeVerifier creates a random PKCE code_verifier (base64url, no padding).
func GenerateCodeVerifier() (string, error) {
	buf := make([]byte, codeVerifierLength)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// CodeChallenge computes the S256 PKCE code_challenge from a code_verifier.
func CodeChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}
