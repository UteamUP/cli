package errors

import (
	"fmt"
	"testing"
)

func TestAuthError(t *testing.T) {
	err := NewAuthError("token expired", nil)
	if err.Error() != "auth error: token expired" {
		t.Errorf("unexpected message: %s", err.Error())
	}

	cause := fmt.Errorf("network down")
	errWithCause := NewAuthError("login failed", cause)
	if errWithCause.Unwrap() != cause {
		t.Error("Unwrap should return cause")
	}
}

func TestAPIError(t *testing.T) {
	err := NewAPIError(404, "Not Found", "resource missing")
	if err.StatusCode != 404 {
		t.Errorf("expected status 404, got %d", err.StatusCode)
	}
	expected := "API error 404 Not Found: resource missing"
	if err.Error() != expected {
		t.Errorf("unexpected message: %s", err.Error())
	}
}

func TestConfigError(t *testing.T) {
	err := NewConfigError("baseUrl", "must not be empty")
	if err.Error() != "config error [baseUrl]: must not be empty" {
		t.Errorf("unexpected message: %s", err.Error())
	}

	errNoField := NewConfigError("", "file not found")
	if errNoField.Error() != "config error: file not found" {
		t.Errorf("unexpected message: %s", errNoField.Error())
	}
}

func TestValidationError(t *testing.T) {
	err := NewValidationError("apiKey", "must be 32 chars")
	if err.Error() != "validation error [apiKey]: must be 32 chars" {
		t.Errorf("unexpected message: %s", err.Error())
	}
}

func TestNotAuthenticatedError(t *testing.T) {
	err := &NotAuthenticatedError{}
	expected := "not authenticated — run \"uteamup login\" or \"ut login\" first"
	if err.Error() != expected {
		t.Errorf("unexpected message: %s", err.Error())
	}
}
